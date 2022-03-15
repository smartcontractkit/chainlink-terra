package monitoring

import (
	"fmt"
	"net/url"
	"os"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
	"github.com/smartcontractkit/terra.go/msg"
)

// TerraConfig contains configuration for connecting to a terra client.
type TerraConfig struct {
	TendermintURL    string
	FCDURL           string
	GRPCAddr         string
	GRPCAPIKey       string
	NetworkName      string
	NetworkID        string
	ChainID          string
	ReadTimeout      time.Duration
	PollInterval     time.Duration
	LinkTokenAddress sdk.AccAddress
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
	if value, isPresent := os.LookupEnv("TERRA_FCD_URL"); isPresent {
		cfg.FCDURL = value
	}
	if value, isPresent := os.LookupEnv("TERRA_GRPC_ADDR"); isPresent {
		cfg.GRPCAddr = value
	}
	if value, isPresent := os.LookupEnv("TERRA_GRPC_API_KEY"); isPresent {
		cfg.GRPCAPIKey = value
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
		"TERRA_FCD_URL":      cfg.FCDURL,
		"TERRA_NETWORK_NAME": cfg.NetworkName,
		"TERRA_NETWORK_ID":   cfg.NetworkID,
		"TERRA_CHAIN_ID":     cfg.ChainID,
	} {
		if currentValue == "" {
			return fmt.Errorf("'%s' env var is required", envVarName)
		}
	}
	if cfg.TendermintURL == "" && cfg.GRPCAddr == "" {
		return fmt.Errorf("either TERRA_TENDERMINT_URL or TERRA_GRPC_ADDR need to be set")
	}
	if cfg.GRPCAddr != "" && cfg.GRPCAPIKey == "" {
		return fmt.Errorf("TERRA_GRPC_API_KEY needs to be set if TERRA_GRPC_ADDR is used")
	}
	// Validate URLs.
	for envVarName, currentValue := range map[string]string{
		"TERRA_TENDERMINT_URL": cfg.TendermintURL,
		"TERRA_GRPC_ADDR":      cfg.GRPCAddr,
		"TERRA_FCD_URL":        cfg.FCDURL,
	} {
		if _, err := url.ParseRequestURI(currentValue); currentValue != "" && err != nil {
			return fmt.Errorf("%s='%s' is not a valid URL: %w", envVarName, currentValue, err)
		}
	}
	return nil
}

func applyDefaults(cfg *TerraConfig) {
	if cfg.FCDURL == "" {
		cfg.FCDURL = "https://fcd.terra.dev/"
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 2 * time.Second
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 5 * time.Second
	}
}
