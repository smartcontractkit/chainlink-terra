package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/pkg/errors"

	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	tmtypes "github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/terra-money/core/app"
	"github.com/terra-money/core/app/params"
	wasmtypes "github.com/terra-money/core/x/wasm/types"

	"github.com/smartcontractkit/terra.go/key"
	"github.com/smartcontractkit/terra.go/msg"
	"github.com/smartcontractkit/terra.go/tx"
)

const httpResponseLimit = 10_000_000 // 10MB

var encodingConfig = params.MakeEncodingConfig()

func init() {
	// Extracted from app.MakeEncodingConfig() to ensure that we only call them once, since they race and can panic.
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	app.ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	app.ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	// authz module use this codec to get signbytes.
	// authz MsgExec can execute all message types,
	// so legacy.Cdc need to register all amino messages to get proper signature
	app.ModuleBasics.RegisterLegacyAminoCodec(legacy.Cdc)
}

//go:generate mockery --name ReaderWriter --output ./mocks/
type ReaderWriter interface {
	Writer
	Reader
}

// Reader provides methods for reading from a terra chain.
type Reader interface {
	Account(address sdk.AccAddress) (uint64, uint64, error)
	ContractStore(contractAddress sdk.AccAddress, queryMsg []byte) ([]byte, error)
	TxsEvents(events []string, paginationParams *query.PageRequest) (*txtypes.GetTxsEventResponse, error)
	Tx(hash string) (*txtypes.GetTxResponse, error)
	LatestBlock() (*tmtypes.GetLatestBlockResponse, error)
	BlockByHeight(height int64) (*tmtypes.GetBlockByHeightResponse, error)
	Balance(addr sdk.AccAddress, denom string) (*sdk.Coin, error)
}

// Writer provides methods for writing to a terra chain.
// Assumes all msgs are for the same from address.
// We may want to support multiple from addresses + signers if a use case arises.
type Writer interface {
	SignAndBroadcast(msgs []sdk.Msg, accountNum uint64, sequence uint64, gasPrice sdk.DecCoin, signer key.PrivKey, mode txtypes.BroadcastMode) (*txtypes.BroadcastTxResponse, error)
	Broadcast(txBytes []byte, mode txtypes.BroadcastMode) (*txtypes.BroadcastTxResponse, error)
	Simulate(txBytes []byte) (*txtypes.SimulateResponse, error)
	BatchSimulateUnsigned(msgs SimMsgs, sequence uint64) (*BatchSimResults, error)
	SimulateUnsigned(msgs []sdk.Msg, sequence uint64) (*txtypes.SimulateResponse, error)
	CreateAndSign(msgs []sdk.Msg, account uint64, sequence uint64, gasLimit uint64, gasLimitMultiplier float64, gasPrice sdk.DecCoin, signer key.PrivKey, timeoutHeight uint64) ([]byte, error)
}

type Subscriber interface {
	Start() error
	Stop() error
	Subscribe(ctx context.Context, query string, outCapacity int) (<-chan ctypes.ResultEvent, error)
	Unsubscribe(ctx context.Context, query string) error
	UnsubscribeAll(ctx context.Context) error
}

var _ ReaderWriter = (*Client)(nil)
var _ Subscriber = (*Client)(nil)

const (
	// DefaultTimeout is the default Terra client timeout.
	// Note that while the terra node is processing a heavy block,
	// requests can be delayed significantly (https://github.com/tendermint/tendermint/issues/6899),
	// however there's nothing we can do but wait until the block is processed.
	// So we set a fairly high timeout here.
	DefaultTimeout = 30 * time.Second
	// DefaultGasLimitMultiplier is the default gas limit multiplier.
	// It scales up the gas limit for 3 reasons:
	// 1. We simulate without a fee present (since we're simulating in order to determine the fee)
	// since we simulate unsigned. The fee is included in the signing data:
	// https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/tx/direct.go#L40)
	// 2. Potential state changes between estimation and execution.
	// 3. The simulation doesn't include db writes in the tendermint node
	// (https://github.com/cosmos/cosmos-sdk/issues/4938)
	DefaultGasLimitMultiplier = 1.5
)

//go:generate mockery --name Logger --output ./mocks/
// Logger is for logging in the client
type Logger interface {
	Infof(format string, values ...interface{})
	Warnf(format string, values ...interface{})
	Errorf(format string, values ...interface{})
}

// Client is a terra client
type Client struct {
	chainID                 string
	clientCtx               cosmosclient.Context
	cosmosServiceClient     txtypes.ServiceClient
	authClient              authtypes.QueryClient
	wasmClient              wasmtypes.QueryClient
	bankClient              banktypes.QueryClient
	tendermintServiceClient tmtypes.ServiceClient
	tmClient                *rpchttp.HTTP
	log                     Logger
}

// responseRoundTripper is a http.RoundTripper which calls respFn with each response body.
type responseRoundTripper struct {
	original http.RoundTripper
	respFn   func([]byte)
}

func (rt *responseRoundTripper) RoundTrip(r *http.Request) (resp *http.Response, err error) {
	resp, err = rt.original.RoundTrip(r)
	if err != nil {
		return
	}
	source := http.MaxBytesReader(nil, resp.Body, httpResponseLimit)
	b, err := io.ReadAll(resp.Body)
	source.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response")
	}
	go rt.respFn(b)
	resp.Body = ioutil.NopCloser(bytes.NewReader(b))
	return
}

// NewClient creates a new terra client
func NewClient(chainID string,
	tendermintURL string,
	requestTimeout time.Duration,
	lggr Logger,
) (*Client, error) {
	if requestTimeout <= 0 {
		requestTimeout = DefaultTimeout
	}
	// Note rpchttp.New or rpchttp.NewWithTimeout use a (buggy) custom transport
	// which results in new connections being created per request.
	// Pass our own client here which uses a default transport and caches connections properly.
	tmClient, err := rpchttp.NewWithClient(tendermintURL, "/websocket", &http.Client{
		Timeout: requestTimeout,
		Transport: &responseRoundTripper{original: http.DefaultTransport,
			// Log any response that is missing the JSONRPC 'id' field, because the tendermint/rpc/jsonrpc/client rejects them.
			respFn: func(b []byte) {
				jsonRPC := struct {
					ID json.RawMessage `json:"id"`
				}{}
				if err := json.Unmarshal(b, &jsonRPC); err != nil {
					lggr.Warnf("Response is not a JSON object: %s: %v", string(b), err)
					return
				}
				if len(jsonRPC.ID) == 0 || string(jsonRPC.ID) == "null" {
					lggr.Warnf("Response is missing JSONRPC ID: %s", string(b))
					return
				}
			},
		},
	})
	if err != nil {
		return nil, err
	}
	ec := encodingConfig
	// Note should terra nodes start exposing grpc, its preferable
	// to connect directly with grpc.Dial to avoid using clientCtx (according to tendermint team).
	// If so then we would start putting timeouts on the ctx we pass in to the generate grpc client calls.
	clientCtx := cosmosclient.Context{}.
		WithClient(tmClient).
		WithChainID(chainID).
		WithCodec(ec.Marshaler).
		WithLegacyAmino(ec.Amino).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithInterfaceRegistry(ec.InterfaceRegistry).
		WithTxConfig(ec.TxConfig)

	cosmosServiceClient := txtypes.NewServiceClient(clientCtx)
	authClient := authtypes.NewQueryClient(clientCtx)
	wasmClient := wasmtypes.NewQueryClient(clientCtx)
	tendermintServiceClient := tmtypes.NewServiceClient(clientCtx)
	bankClient := banktypes.NewQueryClient(clientCtx)

	return &Client{
		chainID:                 chainID,
		cosmosServiceClient:     cosmosServiceClient,
		authClient:              authClient,
		wasmClient:              wasmClient,
		tendermintServiceClient: tendermintServiceClient,
		bankClient:              bankClient,
		clientCtx:               clientCtx,
		tmClient:                tmClient,
		log:                     lggr,
	}, nil
}

// Account read the account address for the account number and sequence number.
// !!Note only one sequence number can be used per account per block!!
func (c *Client) Account(addr sdk.AccAddress) (uint64, uint64, error) {
	r, err := c.authClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: addr.String()})
	if err != nil {
		return 0, 0, err
	}
	var a authtypes.AccountI
	err = c.clientCtx.InterfaceRegistry.UnpackAny(r.Account, &a)
	if err != nil {
		return 0, 0, err
	}
	return a.GetAccountNumber(), a.GetSequence(), nil
}

// ContractStore reads from a WASM contract store
func (c *Client) ContractStore(contractAddress sdk.AccAddress, queryMsg []byte) ([]byte, error) {
	s, err := c.wasmClient.ContractStore(context.Background(), &wasmtypes.QueryContractStoreRequest{
		ContractAddress: contractAddress.String(),
		QueryMsg:        queryMsg,
	})
	if err != nil {
		return nil, err
	}
	//  Note s will be nil on err
	return s.QueryResult, err
}

// TxsEvents returns in tx events in descending order (latest txes first).
// Each event is ANDed together and follows the query language defined
// https://docs.cosmos.network/master/core/events.html
// Note one current issue https://github.com/cosmos/cosmos-sdk/issues/10448
func (c *Client) TxsEvents(events []string, paginationParams *query.PageRequest) (*txtypes.GetTxsEventResponse, error) {
	e, err := c.cosmosServiceClient.GetTxsEvent(context.Background(), &txtypes.GetTxsEventRequest{
		Events:     events,
		Pagination: paginationParams,
		OrderBy:    txtypes.OrderBy_ORDER_BY_DESC,
	})
	return e, err
}

// Tx gets a tx by hash
func (c *Client) Tx(hash string) (*txtypes.GetTxResponse, error) {
	e, err := c.cosmosServiceClient.GetTx(context.Background(), &txtypes.GetTxRequest{
		Hash: hash,
	})
	return e, err
}

// LatestBlock returns the latest block
func (c *Client) LatestBlock() (*tmtypes.GetLatestBlockResponse, error) {
	return c.tendermintServiceClient.GetLatestBlock(context.Background(), &tmtypes.GetLatestBlockRequest{})
}

// BlockByHeight gets a block by height
func (c *Client) BlockByHeight(height int64) (*tmtypes.GetBlockByHeightResponse, error) {
	return c.tendermintServiceClient.GetBlockByHeight(context.Background(), &tmtypes.GetBlockByHeightRequest{Height: height})
}

// CreateAndSign creates and signs a transaction
func (c *Client) CreateAndSign(msgs []sdk.Msg, account uint64, sequence uint64, gasLimit uint64, gasLimitMultiplier float64, gasPrice sdk.DecCoin, signer key.PrivKey, timeoutHeight uint64) ([]byte, error) {
	txbuilder := tx.NewTxBuilder(encodingConfig.TxConfig)
	err := txbuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	gasLimitBuffered := uint64(math.Ceil(float64(gasLimit) * float64(gasLimitMultiplier)))
	txbuilder.SetGasLimit(gasLimitBuffered)
	gasFee := msg.NewCoin(gasPrice.Denom, gasPrice.Amount.MulInt64(int64(gasLimitBuffered)).Ceil().RoundInt())
	txbuilder.SetFeeAmount(sdk.NewCoins(gasFee))
	// 0 timeout height means unset.
	txbuilder.SetTimeoutHeight(timeoutHeight)
	err = txbuilder.Sign(tx.SignModeDirect, tx.SignerData{
		AccountNumber: account,
		ChainID:       c.chainID,
		Sequence:      sequence,
	}, signer, true)
	if err != nil {
		return nil, err
	}
	signedTx, err := txbuilder.GetTxBytes()
	if err != nil {
		return nil, err
	}
	return signedTx, nil
}

// SimMsg binds an ID to a msg
type SimMsg struct {
	ID  int64
	Msg sdk.Msg
}

// SimMsgs is a slice of SimMsg
type SimMsgs []SimMsg

// GetMsgs extracts all msgs from SimMsgs
func (simMsgs SimMsgs) GetMsgs() []sdk.Msg {
	msgs := make([]sdk.Msg, len(simMsgs))
	for i := range simMsgs {
		msgs[i] = simMsgs[i].Msg
	}
	return msgs
}

// GetSimMsgsIDs extracts all IDs from SimMsgs
func (simMsgs SimMsgs) GetSimMsgsIDs() []int64 {
	ids := make([]int64, len(simMsgs))
	for i := range simMsgs {
		ids[i] = simMsgs[i].ID
	}
	return ids
}

// BatchSimResults indicates which msgs failed and which succeeded
type BatchSimResults struct {
	Failed    SimMsgs
	Succeeded SimMsgs
}

var failedMsgIndexRe = regexp.MustCompile(`^.*failed to execute message; message index: (?P<Index>\d+):.*$`)

func (c *Client) failedMsgIndex(err error) (bool, int) {
	if err == nil {
		return false, 0
	}

	m := failedMsgIndexRe.FindStringSubmatch(err.Error())
	if len(m) != 2 {
		return false, 0
	}
	index, err := strconv.ParseInt(m[1], 10, 32)
	if err != nil {
		return false, 0
	}
	return true, int(index)
}

// BatchSimulateUnsigned simulates a group of msgs.
// Assumes at least one msg is present.
// If we fail to simulate the batch, remove the offending tx
// and try again. Repeat until we have a successful batch.
// Keep track of failures so we can mark them as errored.
// Note that the error from simulating indicates the first
// msg in the slice which failed (it simply loops over the msgs
// and simulates them one by one, breaking at the first failure).
func (c *Client) BatchSimulateUnsigned(msgs SimMsgs, sequence uint64) (*BatchSimResults, error) {
	var succeeded []SimMsg
	var failed []SimMsg
	toSim := msgs
	for {
		_, err := c.SimulateUnsigned(toSim.GetMsgs(), sequence)
		containsFailure, failureIndex := c.failedMsgIndex(err)
		if err != nil && !containsFailure {
			return nil, err
		}
		if containsFailure {
			failed = append(failed, toSim[failureIndex])
			succeeded = append(succeeded, toSim[:failureIndex]...)
			// remove offending msg and retry
			if failureIndex == len(toSim)-1 {
				// we're done, last one failed
				c.log.Warnf("simulation error found in last msg, failure %v, index %v, err %v", toSim[failureIndex], failureIndex, err)
				break
			}
			// otherwise there may be more to sim
			c.log.Warnf("simulation error found in a msg, retrying with %v, failure %v, index %v, err %v", toSim[failureIndex+1:], toSim[failureIndex], failureIndex, err)
			toSim = toSim[failureIndex+1:]
		} else {
			// we're done they all succeeded
			succeeded = append(succeeded, toSim...)
			break
		}
	}
	return &BatchSimResults{
		Failed:    failed,
		Succeeded: succeeded,
	}, nil
}

// SimulateUnsigned simulates an unsigned msg
func (c *Client) SimulateUnsigned(msgs []sdk.Msg, sequence uint64) (*txtypes.SimulateResponse, error) {
	txbuilder := tx.NewTxBuilder(encodingConfig.TxConfig)
	if err := txbuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}
	// Create an empty signature literal as the ante handler will populate with a
	// sentinel pubkey.
	// Note the simulation actually won't work without this
	sig := signing.SignatureV2{
		PubKey: &secp256k1.PubKey{},
		Data: &signing.SingleSignatureData{
			SignMode: tx.SignModeDirect,
		},
		Sequence: sequence,
	}
	if err := txbuilder.SetSignatures(sig); err != nil {
		return nil, err
	}
	txBytes, err := txbuilder.GetTxBytes()
	if err != nil {
		return nil, err
	}
	s, err := c.cosmosServiceClient.Simulate(context.Background(), &txtypes.SimulateRequest{
		TxBytes: txBytes,
	})
	return s, err
}

// Simulate simulates a signed transaction
func (c *Client) Simulate(txBytes []byte) (*txtypes.SimulateResponse, error) {
	s, err := c.cosmosServiceClient.Simulate(context.Background(), &txtypes.SimulateRequest{
		TxBytes: txBytes,
	})
	return s, err
}

// Broadcast broadcasts a tx
func (c *Client) Broadcast(txBytes []byte, mode txtypes.BroadcastMode) (*txtypes.BroadcastTxResponse, error) {
	res, err := c.cosmosServiceClient.BroadcastTx(context.Background(), &txtypes.BroadcastTxRequest{
		Mode:    mode,
		TxBytes: txBytes,
	})
	if err != nil {
		return nil, err
	}
	if res.TxResponse == nil {
		return nil, errors.Errorf("got nil tx response")
	}
	if res.TxResponse.Code != 0 {
		return res, errors.Errorf("tx failed with error code: %d, resp %v", res.TxResponse.Code, res.TxResponse)
	}
	return res, err
}

// SignAndBroadcast signs and broadcasts a group of msgs.
func (c *Client) SignAndBroadcast(msgs []sdk.Msg, account uint64, sequence uint64, gasPrice sdk.DecCoin, signer key.PrivKey, mode txtypes.BroadcastMode) (*txtypes.BroadcastTxResponse, error) {
	sim, err := c.SimulateUnsigned(msgs, sequence)
	if err != nil {
		return nil, err
	}
	txBytes, err := c.CreateAndSign(msgs, account, sequence, sim.GasInfo.GasUsed, DefaultGasLimitMultiplier, gasPrice, signer, 0)
	if err != nil {
		return nil, err
	}
	return c.Broadcast(txBytes, mode)
}

// Balance returns the balance of an address
func (c *Client) Balance(addr sdk.AccAddress, denom string) (*sdk.Coin, error) {
	b, err := c.bankClient.Balance(context.Background(), &banktypes.QueryBalanceRequest{Address: addr.String(), Denom: denom})
	if err != nil {
		return nil, err
	}
	return b.Balance, nil
}

// Subscriber implementation

func (c *Client) Start() error {
	return c.tmClient.WSEvents.Start()
}

func (c *Client) Stop() error {
	return c.tmClient.WSEvents.Stop()
}

func (c *Client) Subscribe(ctx context.Context, query string, outCapacity int) (<-chan ctypes.ResultEvent, error) {
	return c.tmClient.WSEvents.Subscribe(ctx, "subscriber-is-ignored", query, outCapacity)
}

func (c *Client) Unsubscribe(ctx context.Context, query string) error {
	return c.tmClient.WSEvents.Unsubscribe(ctx, "subscriber-is-ignored", query)
}

func (c *Client) UnsubscribeAll(ctx context.Context) error {
	return c.tmClient.WSEvents.UnsubscribeAll(ctx, "subscriber-is-ignored")
}
