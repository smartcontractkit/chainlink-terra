package gauntlet

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	common "github.com/smartcontractkit/chainlink-terra/ops/deployer/common"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	opsChainlink "github.com/smartcontractkit/chainlink-relay/ops/chainlink"
	"github.com/smartcontractkit/chainlink-relay/ops/utils"
	relayUtils "github.com/smartcontractkit/chainlink-relay/ops/utils"
	"github.com/smartcontractkit/libocr/offchainreporting2/confighelper"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

const (
	CONTRACTS_VERSION = "v0.1.5"
	// uploaded code IDs
	CW20_ID = iota
	OCR2_ID

	// deployed contract addresses
	LINK
	OCR2
	RequesterAccessController
	BillingAccessController
)

type GauntlerDeployer struct {
	gauntlet relayUtils.Gauntlet
	network  string
	Account  map[int]string
	chainID  string
	keyID    string
	args     []string
}

func New(ctx *pulumi.Context) (GauntlerDeployer, error) {
	os.Setenv("SKIP_PROMPTS", "true")

	// Check gauntlet works
	gauntlet, err := relayUtils.NewGauntlet("../")

	if err != nil {
		return GauntlerDeployer{}, err
	}

	keyID := config.Require(ctx, "TERRA-DEPLOYER-KEY-ID")
	chainID := config.Require(ctx, "CL-RELAY_CHAINID")

	return GauntlerDeployer{
		gauntlet: gauntlet,
		network:  "testnet-bombay",
		Account:  make(map[int]string),
		chainID:  chainID,
		keyID:    keyID,
		args:     []string{"--from", keyID, "--chain-id", chainID, "--gas=auto", "--gas-adjustment=1.25", "--fees=100000uluna", "--broadcast-mode=block", "-y", "-o=json"},
	}, nil
}

func (t *GauntlerDeployer) Load() error {
	msg := utils.LogStatus("Uploading contract artifacts")
	err := t.gauntlet.ExecCommand(
		"upload",
		t.gauntlet.Flag("network", t.network),
		t.gauntlet.Flag("version", CONTRACTS_VERSION),
		"link",
		"ocr2",
		"access_controller",
	)

	if err != nil {
		return fmt.Errorf("Uploading contracts failed %s", err.Error())
	}
	return msg.Check(nil)
}

func (t *GauntlerDeployer) DeployLINK() error {
	fmt.Println("Deploying LINK Token...")
	err := t.gauntlet.ExecCommand(
		"token:deploy",
		t.gauntlet.Flag("network", t.network),
	)
	if err != nil {
		return fmt.Errorf("LINK contract deployment failed: %s", err.Error())
	}

	report, err := t.gauntlet.ReadCommandReport()
	if err != nil {
		return fmt.Errorf("No command report available: %s", err.Error())
	}

	if err == nil && len(report.Data) == 0 {
		err = fmt.Errorf("deploy link produced no logs: %s", err.Error())
	}

	linkAddress := report.Responses[0].Contract
	t.Account[LINK] = linkAddress

	msg := utils.LogStatus("Deployed LINK token")

	return msg.Check(nil)
}

func (t *GauntlerDeployer) DeployOCR() error {
	fmt.Println("Deploying OCR Feed:")
	fmt.Println("Step 1: Init Requester Access Controller")
	err := t.gauntlet.ExecCommand(
		"access_controller:deploy",
		t.gauntlet.Flag("network", t.network),
	)
	if err != nil {
		return fmt.Errorf("Request AC initialization failed: %s", err.Error())
	}
	report, err := t.gauntlet.ReadCommandReport()
	if err != nil {
		return err
	}
	t.Account[RequesterAccessController] = report.Responses[0].Contract

	fmt.Println("Step 2: Init Billing Access Controller")
	err = t.gauntlet.ExecCommand(
		"access_controller:deploy",
		t.gauntlet.Flag("network", t.network),
	)
	if err != nil {
		return fmt.Errorf("Billing AC initialization failed: %s", err.Error())
	}
	report, err = t.gauntlet.ReadCommandReport()
	if err != nil {
		return err
	}
	t.Account[BillingAccessController] = report.Responses[0].Contract

	fmt.Println("Step 6: Init OCR 2 Feed")
	input := common.OCRinit{
		MinAnswer:                 "0",
		MaxAnswer:                 "10000000000",
		Decimals:                  2,
		Description:               "Hello",
		RequesterAccessController: t.Account[RequesterAccessController],
		BillingAccessController:   t.Account[BillingAccessController],
		LinkToken:                 t.Account[LINK],
	}

	jsonInput, err := json.Marshal(input)
	if err != nil {
		return err
	}

	err = t.gauntlet.ExecCommand(
		"ocr2:deploy",
		t.gauntlet.Flag("network", t.network),
		t.gauntlet.Flag("input", string(jsonInput)),
	)
	if err != nil {
		return fmt.Errorf("feed initialization failed: %s", err.Error())
	}

	report, err = t.gauntlet.ReadCommandReport()
	if err != nil {
		return err
	}

	t.Account[OCR2] = report.Responses[0].Contract
	msg := utils.LogStatus("Deployed OCR contract")
	fmt.Printf(" - %s", report.Data["state"])
	return msg.Check(nil)
}

func (t GauntlerDeployer) TransferLINK() error {
	msg := utils.LogStatus("Sending LINK to OCR contract")

	err := t.gauntlet.ExecCommand(
		"cw20_base:transfer",
		t.gauntlet.Flag("network", t.network),
		t.gauntlet.Flag("to", t.Account[OCR2]),
		t.gauntlet.Flag("amount", "1000000000"),
		t.Account[LINK],
	)
	if err != nil {
		return fmt.Errorf("LINK transfer failed: %s", err.Error())
	}

	return msg.Check(err)
}

func (t GauntlerDeployer) InitOCR(keys []opsChainlink.NodeKeys) (rerr error) {
	S := []int{}
	signersArray := []string{}
	transmitterArray := []string{}
	helperOracles := []confighelper.OracleIdentityExtra{}
	offChainPublicKeys := []string{}
	configPublicKeys := []string{}
	peerIDs := []string{}
	for _, k := range keys {
		S = append(S, 1)
		signersArray = append(signersArray, k.OCR2OnchainPublicKey)
		transmitterArray = append(transmitterArray, k.OCR2Transmitter)
		offchainPKByte, err := hex.DecodeString(k.OCR2OffchainPublicKey)
		offChainPublicKeys = append(offChainPublicKeys, k.OCR2OffchainPublicKey)
		configPublicKeys = append(configPublicKeys, k.OCR2ConfigPublicKey)
		peerIDs = append(peerIDs, k.P2PID)
		if err != nil {
			return err
		}
		onchainPKByte, err := hex.DecodeString(k.OCR2OnchainPublicKey)
		if err != nil {
			return err
		}
		configPKByteTemp, err := hex.DecodeString(k.OCR2ConfigPublicKey)
		if err != nil {
			return err
		}
		configPKByte := [32]byte{}
		copy(configPKByte[:], configPKByteTemp)
		helperOracles = append(helperOracles, confighelper.OracleIdentityExtra{
			OracleIdentity: confighelper.OracleIdentity{
				OffchainPublicKey: types.OffchainPublicKey(offchainPKByte),
				OnchainPublicKey:  types.OnchainPublicKey(onchainPKByte),
				PeerID:            k.P2PID,
				TransmitAccount:   types.Account(k.OCR2Transmitter),
			},
			ConfigEncryptionPublicKey: types.ConfigEncryptionPublicKey(configPKByte),
		})
	}

	status := utils.LogStatus("InitOCR: set config test args")
	var f uint8 = 1
	var offchainConfigVersion uint64 = 2
	onchainConfig := []byte{}

	offchainConfig := common.OffchainConfigDetails{
		DeltaProgressNanoseconds: 2 * time.Second,        // pacemaker (timeout rotating leaders, can't be too short)
		DeltaResendNanoseconds:   5 * time.Second,        // resending epoch (help nodes rejoin system)
		DeltaRoundNanoseconds:    1 * time.Second,        // round time (polling data source)
		DeltaGraceNanoseconds:    400 * time.Millisecond, // timeout for waiting observations beyond minimum
		DeltaStageNanoseconds:    5 * time.Second,        // transmission schedule (just for calling transmit)
		RMax:                     3,                      // max rounds prior to rotating leader (longer could be more reliable with good leader)
		S:                        S,
		OffchainPublicKeys:       offChainPublicKeys,
		PeerIDs:                  peerIDs,
		ReportingPluginConfig: common.ReportingPluginConfig{
			AlphaReportInfinite: false,
			AlphaReportPpb:      uint64(0), // always send report
			AlphaAcceptInfinite: false,
			AlphaAcceptPpb:      uint64(0),       // accept all reports (if deviation matches number)
			DeltaCNanoseconds:   0 * time.Second, // heartbeat
		},
		MaxDurationQueryNanoseconds:                        0 * time.Millisecond,
		MaxDurationObservationNanoseconds:                  300 * time.Millisecond,
		MaxDurationReportNanoseconds:                       300 * time.Millisecond,
		MaxDurationShouldAcceptFinalizedReportNanoseconds:  1 * time.Second,
		MaxDurationShouldTransmitAcceptedReportNanoseconds: 1 * time.Second,
		ConfigPublicKeys:                                   configPublicKeys,
	}

	status = utils.LogStatus("InitOCR: begin proposal")
	err := t.gauntlet.ExecCommand(
		"ocr2:begin_proposal",
		t.gauntlet.Flag("network", t.network),
		t.Account[OCR2],
	)

	report, err := t.gauntlet.ReadCommandReport()
	if err != nil {
		return err
	}
	if err == nil && len(report.Data) == 0 {
		err = fmt.Errorf("begin proposal produced no logs: %s", err.Error())
	}

	if status.Check(err) != nil {
		return err
	}

	fmt.Printf(" - %s", report.Data["proposalId"])

	var id string = report.Data["proposalId"]

	// Be prepared to clear the proposal if incomplete.
	defer func() {
		if rerr == nil {
			return // Success
		}
		// Failure: Try to clean up incomplete proposal.
		status = utils.LogStatus("InitOCR: clear proposal: " + id)
		err = t.gauntlet.ExecCommand(
			"ocr2:clear_proposal",
			t.gauntlet.Flag("network", t.network),
			t.gauntlet.Flag("proposalId", id),
			t.Account[OCR2],
		)
		if status.Check(err) != nil {
			fmt.Println(err)
			return
		}
	}()

	jsonInput, err := json.Marshal(common.ProposeConfigDetails{
		ID:            id,
		Payees:        transmitterArray,
		Transmitters:  transmitterArray,
		F:             f,
		OnchainConfig: onchainConfig,
		Signers:       signersArray,
	})
	if err != nil {
		return err
	}

	status = utils.LogStatus("InitOCR: propose config " + id)
	err = t.gauntlet.ExecCommand(
		"ocr2:propose_config",
		t.gauntlet.Flag("network", t.network),
		t.gauntlet.Flag("input", string(jsonInput)),
		t.Account[OCR2],
	)
	if status.Check(err) != nil {
		return err
	}

	jsonInput, err = json.Marshal(
		common.ProposeOffchainConfigDetails{
			ID:                    id,
			OffchainConfigVersion: offchainConfigVersion,
			OffchainConfig:        offchainConfig,
		},
	)
	if err != nil {
		return err
	}
	status = utils.LogStatus("InitOCR: propose offchain config")
	err = t.gauntlet.ExecCommand(
		"ocr2:propose_offchain_config",
		t.gauntlet.Flag("network", t.network),
		t.gauntlet.Flag("input", string(jsonInput)),
		t.Account[OCR2],
	)
	if status.Check(err) != nil {
		return err
	}

	report, err = t.gauntlet.ReadCommandReport()
	if err != nil {
		return err
	}
	if err == nil && len(report.Data) == 0 {
		err = fmt.Errorf("propose offchain config produced no logs: %s", err.Error())
	}

	if status.Check(err) != nil {
		return err
	}

	fmt.Println(report.Data)
	var secret string = report.Data["secret"]

	status = utils.LogStatus("InitOCR: finalize proposal")
	err = t.gauntlet.ExecCommand(
		"ocr2:finalize_proposal",
		t.gauntlet.Flag("network", t.network),
		t.gauntlet.Flag("proposalId", id),
		t.Account[OCR2],
	)
	if status.Check(err) != nil {
		fmt.Println(err)
		return
	}

	report, err = t.gauntlet.ReadCommandReport()
	if err != nil {
		return err
	}
	if err == nil && len(report.Data) == 0 {
		err = fmt.Errorf("finalize proposal produced no logs: %s", err.Error())
	}

	if status.Check(err) != nil {
		return err
	}

	fmt.Println(report.Data)
	var digest string = report.Data["digest"]

	input := common.AcceptProposalDetails{
		ID:     id,
		Digest: digest,
		Secret: secret,
	}
	jsonInput, err = json.Marshal(input)
	if err != nil {
		return err
	}
	status = utils.LogStatus("InitOCR: accept proposal")
	err = t.gauntlet.ExecCommand(
		"ocr2:accept_proposal",
		t.gauntlet.Flag("network", t.network),
		t.gauntlet.Flag("input", string(jsonInput)),
		t.Account[OCR2],
	)
	if status.Check(err) != nil {
		fmt.Println(err)
		return
	}

	return nil
}

func (t GauntlerDeployer) OCR2Address() string {
	return t.Account[OCR2]
}

func (t GauntlerDeployer) Addresses() map[int]string {
	return t.Account
}

func (t GauntlerDeployer) Fund(addresses []string) error {
	for _, a := range addresses {
		msg := utils.LogStatus(fmt.Sprintf("Funded %s", a))
		args := append([]string{"tx", "bank", "send", t.keyID, a, "1000000000uluna"}, t.args...)
		if _, err := exec.Command("terrad", args...).Output(); msg.Check(err) != nil {
			return err
		}
	}
	return nil
}
