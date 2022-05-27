package terrad

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"

	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	opsChainlink "github.com/smartcontractkit/chainlink-relay/ops/chainlink"
	"github.com/smartcontractkit/chainlink-relay/ops/utils"
	common "github.com/smartcontractkit/chainlink-terra/ops/deployer/common"
	"github.com/smartcontractkit/libocr/offchainreporting2/confighelper"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

const (
	// uploaded code IDs
	CW20_ID = iota
	OCR2_ID

	// deployed contract addresses
	LINK
	OCR2
)

type Terrad struct {
	url      string
	chainID  string
	keyID    string
	addr     string
	args     []string
	Uploaded map[int]string
	Deployed map[int]string
}

type key struct {
	Name    string
	Address string
}

func New(ctx *pulumi.Context) (Terrad, error) {
	// check if terrad is installed
	_, err := exec.LookPath("terrad")
	if err != nil {
		return Terrad{}, errors.New("'terrad' is not installed")
	}

	// check the deployer key exists in terrad
	msg := utils.LogStatus("'terrad' is configured correctly")
	keyID := config.Require(ctx, "TERRA-DEPLOYER")
	out, err := exec.Command("terrad", "keys", "show", keyID, "--output", "json").Output()
	if msg.Check(err) != nil {
		return Terrad{}, fmt.Errorf("'%s' key not found in terrad - please set", keyID)
	}
	var deployerAddr key
	if err := json.Unmarshal(out, &deployerAddr); err != nil {
		return Terrad{}, err
	}
	fmt.Printf("Deployer address: %s\n", deployerAddr.Address)

	chainID := config.Require(ctx, "CL-RELAY_CHAINID")
	return Terrad{
		url:      config.Require(ctx, "CL-TENDERMINT_URL"),
		chainID:  chainID,
		keyID:    keyID,
		addr:     deployerAddr.Address,
		args:     []string{"--from", keyID, "--chain-id", chainID, "--gas=auto", "--gas-adjustment=1.25", "--fees=100000uluna", "--broadcast-mode=block", "-y", "-o=json"},
		Uploaded: map[int]string{},
		Deployed: map[int]string{},
	}, err
}

type TxResponse struct {
	Code      int32
	Codespace string
	Logs      cosmostypes.ABCIMessageLogs
}

func (t *Terrad) Load() error {
	msg := utils.LogStatus("Uploading contract artifacts")
	for _, contract := range []string{"ocr2", "cw20_base"} {
		args := append([]string{"tx", "wasm", "store", fmt.Sprintf("./terrad/artifacts/%s.wasm", contract)}, t.args...)
		out, err := exec.Command("terrad", args...).Output()
		if err != nil {
			return msg.Check(err)
		}

		var res TxResponse
		if err := json.Unmarshal(out, &res); err != nil {
			return msg.Check(err)
		}
		for _, event := range res.Logs[0].Events {
			if event.Type == "store_code" {
				for _, attr := range event.Attributes {
					if attr.Key == "code_id" {
						var key int
						switch contract {
						case "ocr2":
							key = OCR2_ID
						case "cw20_base":
							key = CW20_ID
						default:
							return errors.New("unknown contract type does not have assigned key")
						}
						t.Uploaded[key] = attr.Value
					}
				}
			}
		}
	}
	return msg.Check(nil)
}

func (t *Terrad) DeployLINK() error {
	msg := utils.LogStatus("Deployed LINK token")
	initBal := common.balance{Address: t.addr, Amount: "1000000000000000000000000000"}
	initMsg := common.LINKinit{
		Name:            "ChainLink Token",
		Symbol:          "LINK",
		Decimals:        18,
		InitialBalances: []common.balance{initBal},
		Mint:            nil,
		Marketing:       nil,
	}

	msgBytes, err := json.Marshal(initMsg)
	if err != nil {
		return msg.Check(err)
	}

	args := append([]string{"tx", "wasm", "instantiate", t.Uploaded[CW20_ID], string(msgBytes)}, t.args...)
	out, err := exec.Command("terrad", args...).Output()
	if err != nil {
		return msg.Check(err)
	}

	var res TxResponse
	if err := json.Unmarshal(out, &res); err != nil {
		return msg.Check(err)
	}
	for _, event := range res.Logs[0].Events {
		if event.Type == "instantiate_contract" {
			for _, attr := range event.Attributes {
				if attr.Key == "contract_address" {
					t.Deployed[LINK] = attr.Value
					fmt.Printf(" - %s", attr.Value)
				}
			}
		}
	}

	return msg.Check(nil)
}

func (t *Terrad) DeployOCR() error {
	msg := utils.LogStatus("Deployed OCR contract")
	initMsg := common.OCRinit{
		LinkToken:                 t.Deployed[LINK],
		MinAnswer:                 "0",
		MaxAnswer:                 "999999999999999999",
		Decimals:                  8,
		BillingAccessController:   "terra1dcegyrekltswvyy0xy69ydgxn9x8x32zdtapd8", // placeholder
		RequesterAccessController: "terra1dcegyrekltswvyy0xy69ydgxn9x8x32zdtapd8", // placeholder
		Description:               "LINK/USD - OCR2",
	}

	msgBytes, err := json.Marshal(initMsg)
	if err != nil {
		return msg.Check(err)
	}

	args := append([]string{"tx", "wasm", "instantiate", t.Uploaded[OCR2_ID], string(msgBytes)}, t.args...)
	out, err := exec.Command("terrad", args...).Output()
	if err != nil {
		return msg.Check(err)
	}

	var res TxResponse
	if err := json.Unmarshal(out, &res); err != nil {
		return msg.Check(err)
	}
	for _, event := range res.Logs[0].Events {
		if event.Type == "instantiate_contract" {
			for _, attr := range event.Attributes {
				if attr.Key == "contract_address" {
					t.Deployed[OCR2] = attr.Value
					fmt.Printf(" - %s", attr.Value)
				}
			}
		}
	}

	return msg.Check(nil)
}

func (t Terrad) TransferLINK() error {
	msg := utils.LogStatus("Sending LINK to OCR contract")
	tx := common.Send{
		Send: common.SendDetails{
			Contract: t.Deployed[OCR2],
			Amount:   "100000000000000000000",
			Msg:      "",
		},
	}

	txBytes, err := json.Marshal(tx)
	if err != nil {
		return msg.Check(err)
	}

	args := append([]string{"tx", "wasm", "execute", t.Deployed[LINK], string(txBytes)}, t.args...)
	_, err = exec.Command("terrad", args...).Output()
	return msg.Check(err)
}

func (t Terrad) InitOCR(keys []opsChainlink.NodeKeys) (rerr error) {
	S := []int{}
	helperOracles := []confighelper.OracleIdentityExtra{}
	for _, k := range keys {
		S = append(S, 1)
		offchainPKByte, err := hex.DecodeString(k.OCR2OffchainPublicKey)
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
	alphaPPB := uint64(1000000)
	signers, transmitters, f, onchainConfig, offchainConfigVersion, offchainConfig, err := confighelper.ContractSetConfigArgsForTests(
		2*time.Second,        // deltaProgress time.Duration,
		5*time.Second,        // deltaResend time.Duration,
		1*time.Second,        // deltaRound time.Duration,
		500*time.Millisecond, // deltaGrace time.Duration,
		10*time.Second,       // deltaStage time.Duration,
		3,                    // rMax uint8,
		S,                    // s []int,
		helperOracles,        // oracles []OracleIdentityExtra,
		median.OffchainConfig{
			false,
			alphaPPB,
			false,
			alphaPPB,
			15 * time.Second,
		}.Encode(), //reportingPluginConfig []byte,
		500*time.Millisecond, // maxDurationQuery time.Duration,
		500*time.Millisecond, // maxDurationObservation time.Duration,
		500*time.Millisecond, // maxDurationReport time.Duration,
		2*time.Second,        // maxDurationShouldAcceptFinalizedReport time.Duration,
		2*time.Second,        // maxDurationShouldTransmitAcceptedReport time.Duration,
		1,                    // f int,
		[]byte{},             // onchainConfig []byte (calculated by the contract)
	)
	if status.Check(err) != nil {
		return err
	}

	// convert type for marshalling
	signerArray := [][]byte{}
	transmitterArray := []string{}
	for i := 0; i < len(signers); i++ {
		signerArray = append(signerArray, signers[i])
		transmitterArray = append(transmitterArray, string(transmitters[i]))
	}

	status = utils.LogStatus("InitOCR: begin proposal")
	resp, err := t.ExecuteOCR2(common.BeginProposal)
	if err == nil && len(resp.Logs) == 0 {
		err = errors.New("begin proposal produced no logs")
	}

	if status.Check(err) != nil {
		return err
	}

	var id string
	for _, e := range resp.Logs[0].Events {
		if e.Type == "wasm" {
			for _, a := range e.Attributes {
				if a.Key == "proposal_id" {
					if a.Value == "" {
						return errors.New("empty proposal id")
					}
					id = a.Value
				}
			}
		}
	}
	if id == "" {
		return errors.New("failed to find event with attribute: wasm.proposal-id")
	}

	// Be prepared to clear the proposal if incomplete.
	defer func() {
		if rerr == nil {
			return // Success
		}
		// Failure: Try to clean up incomplete proposal.
		status = utils.LogStatus("InitOCR: clear proposal: " + id)
		resp, err = t.ExecuteOCR2(common.ClearProposal{
			ClearProposal: common.ClearProposalDetails{ID: id},
		})
		if status.Check(err) != nil {
			fmt.Println(err)
			return
		}
	}()

	payees := make([]string, 0)
	for i := 0; i < len(transmitterArray); i++ {
		payees = append(payees, t.addr)
	}
	status = utils.LogStatus("InitOCR: propose config " + id)
	resp, err = t.ExecuteOCR2(common.ProposeConfig{
		ProposeConfig: common.ProposeConfigDetails{
			ID:            id,
			Payees:        payees,
			Signers:       signerArray,
			Transmitters:  transmitterArray,
			F:             f,
			OnchainConfig: onchainConfig,
		},
	})
	if status.Check(err) != nil {
		return err
	}

	status = utils.LogStatus("InitOCR: propose offchain config")
	resp, err = t.ExecuteOCR2(common.ProposeOffchainConfig{
		ProposeOffchainConfig: common.ProposeOffchainConfigDetails{
			ID:                    id,
			OffchainConfigVersion: offchainConfigVersion,
			OffchainConfig:        offchainConfig,
		},
	})
	if status.Check(err) != nil {
		return err
	}

	status = utils.LogStatus("InitOCR: finalize proposal")
	resp, err = t.ExecuteOCR2(common.FinalizeProposal{
		FinalizeProposal: common.FinalizeProposalDetails{ID: id},
	})
	if status.Check(err) != nil {
		return err
	}

	var digest []byte
	for _, e := range resp.Logs[0].Events {
		if e.Type == "wasm" {
			for _, a := range e.Attributes {
				if a.Key == "digest" {
					h, err := hex.DecodeString(a.Value)
					if err != nil {
						return fmt.Errorf("failed to parse digest: %v", err)
					}
					digest = h
				}
			}
		}
	}
	switch len(digest) {
	case 0:
		return errors.New("failed to find event with attribute: wasm.digest")
	case 32:
		// expected
	default:
		return fmt.Errorf("wrong length for: wasm.digest: %d", len(digest))
	}

	status = utils.LogStatus("InitOCR: accept proposal")
	resp, err = t.ExecuteOCR2(common.AcceptProposal{
		AcceptProposal: common.AcceptProposalDetails{
			ID:     id,
			Digest: digest,
		},
	})
	if status.Check(err) != nil {
		return err
	}

	return nil
}

func (t Terrad) ExecuteOCR2(msg interface{}) (resp TxResponse, err error) {
	return t.Execute(t.OCR2Address(), msg)
}

func (t Terrad) Execute(addr string, msg interface{}) (resp TxResponse, err error) {
	var b []byte
	b, err = json.Marshal(msg)
	if err != nil {
		return
	}
	args := []string{"tx", "wasm", "execute", addr, string(b)}
	cmd := exec.Command("terrad", append(args, t.args...)...)
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	var out []byte
	out, err = cmd.Output()
	if err != nil {
		err = fmt.Errorf("%s: %s", err, stdErr.String())
		return
	}
	err = json.Unmarshal(out, &resp)
	if err == nil && resp.Code != 0 {
		err = fmt.Errorf("tx response contains error: %s %d", resp.Codespace, resp.Code)
	}
	return
}

func (t Terrad) OCR2Address() string {
	return t.Deployed[OCR2]
}

func (t Terrad) Addresses() map[int]string {
	return t.Deployed
}

func (t Terrad) Fund(addresses []string) error {
	for _, a := range addresses {
		msg := utils.LogStatus(fmt.Sprintf("Funded %s", a))
		args := append([]string{"tx", "bank", "send", t.keyID, a, "1000000000uluna"}, t.args...)
		if _, err := exec.Command("terrad", args...).Output(); msg.Check(err) != nil {
			return err
		}
	}
	return nil
}
