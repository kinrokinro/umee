package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
)

const (
	denomUSD = "USD"

	defaultListenAddr      = "0.0.0.0:7171"
	defaultSrvWriteTimeout = 15 * time.Second
	defaultSrvReadTimeout  = 15 * time.Second

	ProviderKraken  = "kraken"
	ProviderBinance = "binance"
	ProviderOsmosis = "osmosis"
	ProviderHuobi   = "huobi"
	ProviderMock    = "mock"
)

var (
	validate *validator.Validate = validator.New()

	// ErrEmptyConfigPath defines a sentinel error for an empty config path.
	ErrEmptyConfigPath = errors.New("empty configuration file path")

	// SupportedProviders defines a lookup table of all the supported currency API
	// providers.
	SupportedProviders = map[string]struct{}{
		ProviderKraken:  {},
		ProviderBinance: {},
		ProviderOsmosis: {},
		ProviderHuobi:   {},
		ProviderMock:    {},
	}
)

type (
	// Config defines all necessary price-feeder configuration parameters.
	Config struct {
		Server        Server         `toml:"server"`
		CurrencyPairs []CurrencyPair `toml:"currency_pairs" validate:"required,gt=0,dive,required"`
		Account       Account        `toml:"account" validate:"required,gt=0,dive,required"`
		Keyring       Keyring        `toml:"keyring" validate:"required,gt=0,dive,required"`
		RPC           RPC            `toml:"rpc" validate:"required,gt=0,dive,required"`
		GasAdjustment float64        `toml:"gas_adjustment" validate:"required"`
	}

	// Server defines the API server configuration.
	Server struct {
		ListenAddr     string   `toml:"listen_addr"`
		WriteTimeout   string   `toml:"write_timeout"`
		ReadTimeout    string   `toml:"read_timeout"`
		VerboseCORS    bool     `toml:"verbose_cors"`
		AllowedOrigins []string `toml:"allowed_origins"`
	}

	// CurrencyPair defines a price quote of the exchange rate for two different
	// currencies and the supported providers for getting the exchange rate.
	CurrencyPair struct {
		Base      string   `toml:"base" validate:"required"`
		Quote     string   `toml:"quote" validate:"required"`
		Providers []string `toml:"providers" validate:"required,gt=0,dive,required"`
	}

	// Account defines account related configuration that is related to the Umee
	// network and transaction signing functionality.
	Account struct {
		ChainID   string `toml:"chain_id" validate:"required"`
		Address   string `toml:"address" validate:"required"`
		Validator string `toml:"validator" validate:"required"`
	}

	// Keyring defines the required Umee keyring configuration.
	Keyring struct {
		Backend string `toml:"backend" validate:"required"`
		Pass    string `toml:"pass"`
		Dir     string `toml:"dir" validate:"required"`
	}

	// RPC defines RPC configuration of both the Umee gRPC and Tendermint nodes.
	RPC struct {
		TMRPCEndpoint string `toml:"tmrpc_endpoint" validate:"required"`
		GRPCEndpoint  string `toml:"grpc_endpoint" validate:"required"`
		RPCTimeout    string `toml:"rpc_timeout" validate:"required"`
	}
)

// Validate returns an error if the Config object is invalid.
func (c Config) Validate() error {
	return validate.Struct(c)
}

// ParseConfig attempts to read and parse configuration from the given file path.
// An error is returned if reading or parsing the config fails.
func ParseConfig(configPath string) (Config, error) {
	var cfg Config

	if configPath == "" {
		return cfg, ErrEmptyConfigPath
	}

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config: %w", err)
	}

	if _, err := toml.Decode(string(configData), &cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode config: %w", err)
	}

	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = defaultListenAddr
	}
	if len(cfg.Server.WriteTimeout) == 0 {
		cfg.Server.WriteTimeout = defaultSrvWriteTimeout.String()
	}
	if len(cfg.Server.ReadTimeout) == 0 {
		cfg.Server.ReadTimeout = defaultSrvReadTimeout.String()
	}

	for _, cp := range cfg.CurrencyPairs {
		if !strings.Contains(strings.ToUpper(cp.Quote), denomUSD) {
			return cfg, fmt.Errorf("unsupported pair quote: %s", cp.Quote)
		}

		for _, provider := range cp.Providers {
			if _, ok := SupportedProviders[provider]; !ok {
				return cfg, fmt.Errorf("unsupported provider: %s", provider)
			}
		}
	}

	return cfg, cfg.Validate()
}
