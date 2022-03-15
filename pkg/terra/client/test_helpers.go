package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/pelletier/go-toml"
	"github.com/smartcontractkit/terra.go/key"
	"github.com/smartcontractkit/terra.go/msg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createKeyFromMnemonic(t *testing.T, mnemonic string) (key.PrivKey, sdk.AccAddress) {
	// Derive Raw Private Key
	privKeyBz, err := key.DerivePrivKeyBz(mnemonic, key.CreateHDPath(0, 0))
	assert.NoError(t, err)
	// Generate StdPrivKey
	privKey, err := key.PrivKeyGen(privKeyBz)
	assert.NoError(t, err)
	addr := msg.AccAddress(privKey.PubKey().Address())
	return privKey, addr
}

type Account struct {
	Name       string
	PrivateKey key.PrivKey
	Address    sdk.AccAddress
}

// 0.001
var minGasPrice = msg.NewDecCoinFromDec("uluna", msg.NewDecWithPrec(1, 3))

// SetupLocalTerraNode sets up a local terra node via terrad, and returns pre-funded accounts, the test directory, and the url.
func SetupLocalTerraNode(t *testing.T, chainID string) ([]Account, string, string) {
	testdir, err := ioutil.TempDir("", "integration-test")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(testdir))
	})
	t.Log(testdir)
	out, err := exec.Command("terrad", "init", "integration-test", "-o", "--chain-id", chainID, "--home", testdir).Output()
	require.NoError(t, err, string(out))

	p := path.Join(testdir, "config", "app.toml")
	f, err := os.ReadFile(p)
	require.NoError(t, err)
	config, err := toml.Load(string(f))
	require.NoError(t, err)
	// Enable if desired to use lcd endpoints config.Set("api.enable", "true")
	config.Set("minimum-gas-prices", minGasPrice.String())
	require.NoError(t, os.WriteFile(p, []byte(config.String()), 0600))
	// TODO: could also speed up the block mining config

	// Create 2 test accounts
	var accounts []Account
	for i := 0; i < 2; i++ {
		account := fmt.Sprintf("test%d", i)
		key, err2 := exec.Command("terrad", "keys", "add", account, "--output", "json", "--keyring-backend", "test", "--keyring-dir", testdir).CombinedOutput()
		require.NoError(t, err2, string(key))
		var k struct {
			Address  string `json:"address"`
			Mnemonic string `json:"mnemonic"`
		}
		require.NoError(t, json.Unmarshal(key, &k))
		expAcctAddr, err3 := sdk.AccAddressFromBech32(k.Address)
		require.NoError(t, err3)
		privateKey, address := createKeyFromMnemonic(t, k.Mnemonic)
		require.Equal(t, expAcctAddr, address)
		// Give it 100 luna
		out2, err2 := exec.Command("terrad", "add-genesis-account", k.Address, "100000000uluna", "--home", testdir).Output() //nolint:gosec
		require.NoError(t, err2, string(out2))
		accounts = append(accounts, Account{
			Name:       account,
			Address:    address,
			PrivateKey: privateKey,
		})
	}
	// Stake 10 luna in first acct
	out, err = exec.Command("terrad", "gentx", accounts[0].Name, "10000000uluna", fmt.Sprintf("--chain-id=%s", chainID), "--keyring-backend", "test", "--keyring-dir", testdir, "--home", testdir).CombinedOutput() //nolint:gosec
	require.NoError(t, err, string(out))
	out, err = exec.Command("terrad", "collect-gentxs", "--home", testdir).CombinedOutput()
	require.NoError(t, err, string(out))

	port := mustRandomPort()
	tendermintURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	t.Log(tendermintURL)
	cmd := exec.Command("terrad", "start", "--home", testdir,
		"--rpc.laddr", fmt.Sprintf("tcp://127.0.0.1:%d", port),
		"--rpc.pprof_laddr", "0.0.0.0:0",
		"--grpc.address", "0.0.0.0:0",
		"--grpc-web.address", "0.0.0.0:0",
		"--p2p.laddr", "0.0.0.0:0")
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		assert.NoError(t, cmd.Process.Kill())
		if err2 := cmd.Wait(); assert.Error(t, err2) {
			if !assert.Contains(t, err2.Error(), "signal: killed", cmd.ProcessState.String()) {
				t.Log("terrad stderr:", stdErr.String())
			}
		}
	})

	// Wait for api server to boot
	var ready bool
	for i := 0; i < 30; i++ {
		time.Sleep(time.Second)
		out, err = exec.Command("curl", tendermintURL+"/abci_info").Output() //nolint:gosec
		if err != nil {
			t.Logf("API server not ready yet (attempt %d): %v\n", i+1, err)
			continue
		}
		var a struct {
			Result struct {
				Response struct {
					LastBlockHeight string `json:"last_block_height"`
				} `json:"response"`
			} `json:"result"`
		}
		require.NoError(t, json.Unmarshal(out, &a), string(out))
		if a.Result.Response.LastBlockHeight == "" {
			t.Logf("API server not ready yet (attempt %d)\n", i+1)
			continue
		}
		ready = true
		break
	}
	require.True(t, ready)
	return accounts, testdir, tendermintURL
}

// DeployTestContract deploys a test contract.
func DeployTestContract(ctx context.Context, t *testing.T, tendermintURL string, deployAccount, ownerAccount Account, tc *Client, testdir, wasmTestContractPath string) sdk.AccAddress {
	//nolint:gosec
	out, err := exec.Command("terrad", "tx", "wasm", "store", wasmTestContractPath, "--node", tendermintURL,
		"--from", deployAccount.Name, "--gas", "auto", "--fees", "100000uluna", "--chain-id", "42", "--broadcast-mode", "block", "--home", testdir, "--keyring-backend", "test", "--keyring-dir", testdir, "--yes").CombinedOutput()
	require.NoError(t, err, string(out))
	an, sn, err2 := tc.Account(ctx, ownerAccount.Address)
	require.NoError(t, err2)
	r, err3 := tc.SignAndBroadcast(ctx, []msg.Msg{
		msg.NewMsgInstantiateContract(ownerAccount.Address, nil, 1, []byte(`{"count":0}`), nil)}, an, sn, minGasPrice, ownerAccount.PrivateKey, txtypes.BroadcastMode_BROADCAST_MODE_BLOCK)
	require.NoError(t, err3)
	return GetContractAddr(ctx, t, tc, r.TxResponse.TxHash)
}

func GetContractAddr(ctx context.Context, t *testing.T, tc *Client, deploymentHash string) sdk.AccAddress {
	var deploymentTx *txtypes.GetTxResponse
	var err error
	for try := 0; try < 5; try++ {
		deploymentTx, err = tc.Tx(ctx, deploymentHash)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
	}
	require.NoError(t, err)
	var contractAddr string
	for _, etype := range deploymentTx.TxResponse.Events {
		if etype.Type == "wasm" {
			for _, attr := range etype.Attributes {
				if string(attr.Key) == "contract_address" {
					contractAddr = string(attr.Value)
				}
			}
		}
	}
	require.NotEqual(t, "", contractAddr)
	contract, err := sdk.AccAddressFromBech32(contractAddr)
	require.NoError(t, err)
	return contract
}

func mustRandomPort() int {
	r, err := rand.Int(rand.Reader, big.NewInt(65535-1023))
	if err != nil {
		panic(fmt.Errorf("unexpected error generating random port: %w", err))
	}
	return int(r.Int64() + 1024)
}
