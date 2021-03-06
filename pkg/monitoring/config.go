package monitoring

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
	"github.com/smartcontractkit/terra.go/msg"
)

// TerraConfig contains configuration for connecting to a terra client.
type TerraConfig struct {
	TendermintURL        string
	TendermintReqsPerSec int
	FCDURL               string
	FCDReqsPerSec        int
	NetworkName          string
	NetworkID            string
	ChainID              string
	ReadTimeout          time.Duration
	PollInterval         time.Duration
	LinkTokenAddress     sdk.AccAddress
}

var _ relayMonitoring.ChainConfig = TerraConfig{}

// GetRPCEndpoint return the tendermint url of a terra client.
func (t TerraConfig) GetRPCEndpoint() string { return t.TendermintURL }

// GetNetworkName returns the network name.
func (t TerraConfig) GetNetworkName() string { return t.NetworkName }

// GetNetworkID returns the network id.
func (t TerraConfig) GetNetworkID() string { return t.NetworkID }

// GetChainID returns the chain id.
func (t TerraConfig) GetChainID() string { return t.ChainID }

// GetReadTimeout returns the max allowed duration of a request to a Terra client.
func (t TerraConfig) GetReadTimeout() time.Duration { return t.ReadTimeout }

// GetPollInterval returns the interval at which data from the chain is read.
func (t TerraConfig) GetPollInterval() time.Duration { return t.PollInterval }

// ToMapping returns a data structure expected by the Avro schema encoders.
func (t TerraConfig) ToMapping() map[string]interface{} {
	return map[string]interface{}{
		"network_name": t.NetworkName,
		"network_id":   t.NetworkID,
		"chain_id":     t.ChainID,
	}
}

// ParseTerraConfig extracts chain specific configuration from env vars.
func ParseTerraConfig() (TerraConfig, error) {
	cfg := TerraConfig{}

	if err := parseEnvVars(&cfg); err != nil {
		return cfg, err
	}

	applyDefaults(&cfg)

	err := validateConfig(cfg)
	return cfg, err
}

func parseEnvVars(cfg *TerraConfig) error {
	if value, isPresent := os.LookupEnv("TERRA_TENDERMINT_URL"); isPresent {
		cfg.TendermintURL = value
	}
	if value, isPresent := os.LookupEnv("TERRA_TENDERMINT_REQS_PER_SEC"); isPresent {
		rps, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("failed to parse env var TERRA_TENDERMINT_REQS_PER_SEC '%s' as int: %w", value, err)
		}
		cfg.TendermintReqsPerSec = rps
	}
	if value, isPresent := os.LookupEnv("TERRA_FCD_URL"); isPresent {
		cfg.FCDURL = value
	}
	if value, isPresent := os.LookupEnv("TERRA_FCD_REQS_PER_SEC"); isPresent {
		rps, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("failed to parse env var TERRA_FCD_REQS_PER_SEC '%s' as int: %w", value, err)
		}
		cfg.FCDReqsPerSec = rps
	}
	if value, isPresent := os.LookupEnv("TERRA_NETWORK_NAME"); isPresent {
		cfg.NetworkName = value
	}
	if value, isPresent := os.LookupEnv("TERRA_NETWORK_ID"); isPresent {
		cfg.NetworkID = value
	}
	if value, isPresent := os.LookupEnv("TERRA_CHAIN_ID"); isPresent {
		cfg.ChainID = value
	}
	if value, isPresent := os.LookupEnv("TERRA_READ_TIMEOUT"); isPresent {
		readTimeout, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("failed to parse env var TERRA_READ_TIMEOUT, see https://pkg.go.dev/time#ParseDuration: %w", err)
		}
		cfg.ReadTimeout = readTimeout
	}
	if value, isPresent := os.LookupEnv("TERRA_POLL_INTERVAL"); isPresent {
		pollInterval, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("failed to parse env var TERRA_POLL_INTERVAL, see https://pkg.go.dev/time#ParseDuration: %w", err)
		}
		cfg.PollInterval = pollInterval
	}
	if value, isPresent := os.LookupEnv("TERRA_LINK_TOKEN_ADDRESS"); isPresent {
		address, err := msg.AccAddressFromBech32(value)
		if err != nil {
			return fmt.Errorf("failed to parse the bech32-encoded link token address from '%s': %w", value, err)
		}
		cfg.LinkTokenAddress = address
	}
	return nil
}

func validateConfig(cfg TerraConfig) error {
	// Required config
	for envVarName, currentValue := range map[string]string{
		"TERRA_TENDERMINT_URL": cfg.TendermintURL,
		"TERRA_FCD_URL":        cfg.FCDURL,
		"TERRA_NETWORK_NAME":   cfg.NetworkName,
		"TERRA_NETWORK_ID":     cfg.NetworkID,
		"TERRA_CHAIN_ID":       cfg.ChainID,
	} {
		if currentValue == "" {
			return fmt.Errorf("'%s' env var is required", envVarName)
		}
	}
	// Validate URLs.
	for envVarName, currentValue := range map[string]string{
		"TERRA_TENDERMINT_URL": cfg.TendermintURL,
		"TERRA_FCD_URL":        cfg.FCDURL,
	} {
		if _, err := url.ParseRequestURI(currentValue); currentValue != "" && err != nil {
			return fmt.Errorf("%s='%s' is not a valid URL: %w", envVarName, currentValue, err)
		}
	}
	return nil
}

func applyDefaults(cfg *TerraConfig) {
	if cfg.TendermintReqsPerSec == 0 {
		cfg.TendermintReqsPerSec = 1
	}
	if cfg.FCDURL == "" {
		cfg.FCDURL = "https://fcd.terra.dev/"
	}
	if cfg.FCDReqsPerSec == 0 {
		cfg.FCDReqsPerSec = 1
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 2 * time.Second
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 5 * time.Second
	}
}
