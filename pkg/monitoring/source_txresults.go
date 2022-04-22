package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
	"go.uber.org/ratelimit"
)

// NewTxResultsSourceFactory builds sources of TxResults objects expected by the relay monitoring.
func NewTxResultsSourceFactory(log relayMonitoring.Logger) relayMonitoring.SourceFactory {
	return &txResultsSourceFactory{
		log,
		&http.Client{},
		ratelimit.New(1,
			ratelimit.Per(1*time.Second), // one request every 1 second
			ratelimit.WithoutSlack,       // don't accumulate previously "unspent" requests for future bursts
		),
	}
}

type txResultsSourceFactory struct {
	log           relayMonitoring.Logger
	httpClient    *http.Client
	globalLimiter ratelimit.Limiter
}

func (t *txResultsSourceFactory) NewSource(
	chainConfig relayMonitoring.ChainConfig,
	feedConfig relayMonitoring.FeedConfig,
) (relayMonitoring.Source, error) {
	terraConfig, ok := chainConfig.(TerraConfig)
	if !ok {
		return nil, fmt.Errorf("expected chainConfig to be of type TerraConfig not %T", chainConfig)
	}
	terraFeedConfig, ok := feedConfig.(TerraFeedConfig)
	if !ok {
		return nil, fmt.Errorf("expected feedConfig to be of type TerraFeedConfig not %T", feedConfig)
	}
	return &txResultsSource{
		t.log,
		terraConfig,
		terraFeedConfig,
		t.httpClient,
		t.globalLimiter,
		0,
		sync.Mutex{},
	}, nil
}

func (t *txResultsSourceFactory) GetType() string {
	return "txresults"
}

type txResultsSource struct {
	log             relayMonitoring.Logger
	terraConfig     TerraConfig
	terraFeedConfig TerraFeedConfig
	httpClient      *http.Client
	globalLimiter   ratelimit.Limiter

	latestTxID   uint64
	latestTxIDMu sync.Mutex
}

type fcdTxsResponse struct {
	Txs []fcdTx `json:"txs"`
}

type fcdTx struct {
	ID uint64 `json:"id"`
	// Error code if present
	Code int `json:"code,omitempty"`
}

func (t *txResultsSource) Fetch(ctx context.Context) (interface{}, error) {
	t.globalLimiter.Take()
	// Query the FCD endpoint.
	query := url.Values{}
	query.Set("account", t.terraFeedConfig.ContractAddressBech32)
	query.Set("limit", "10")
	query.Set("offset", "0")
	getTxsURL, err := url.Parse(t.terraConfig.FCDURL)
	if err != nil {
		return nil, err
	}
	getTxsURL.Path = "/v1/txs"
	getTxsURL.RawQuery = query.Encode()
	readTxsReq, err := http.NewRequestWithContext(ctx, http.MethodGet, getTxsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to build a request to the terra FCD: %w", err)
	}
	res, err := t.httpClient.Do(readTxsReq)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch transactions from terra FCD: %w", err)
	}
	defer res.Body.Close()
	// Decode the response
	txsResponse := fcdTxsResponse{}
	resBody, _ := io.ReadAll(res.Body)
	if err := json.Unmarshal(resBody, &txsResponse); err != nil {
		return nil, fmt.Errorf("unable to decode transactions from response '%s': %w", resBody, err)
	}
	// Filter recent transactions
	// TODO (dru) keep latest processed tx in the state.
	recentTxs := []fcdTx{}
	func() {
		t.latestTxIDMu.Lock()
		defer t.latestTxIDMu.Unlock()
		maxTxID := t.latestTxID
		for _, tx := range txsResponse.Txs {
			if tx.ID > t.latestTxID {
				recentTxs = append(recentTxs, tx)
			}
			if tx.ID > maxTxID {
				maxTxID = tx.ID
			}
		}
		t.latestTxID = maxTxID
	}()
	// Count failed and succeeded recent transactions
	output := relayMonitoring.TxResults{}
	for _, tx := range recentTxs {
		if isFailedTransaction(tx) {
			output.NumFailed++
		} else {
			output.NumSucceeded++
		}
	}
	return output, nil
}

// Helpers

func isFailedTransaction(tx fcdTx) bool {
	// See https://docs.cosmos.network/master/building-modules/errors.html
	return tx.Code != 0
}
