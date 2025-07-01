package mysql

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-secure-stdlib/parseutil"
	"github.com/mitchellh/mapstructure"
)

// Config contains the full decoded, validated config for a MySQL connection.
type Config struct {
	ConnectionURL      string `mapstructure:"connection_url"`
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
	AuthType           string `mapstructure:"auth_type"`
	ServiceAccountJSON string `mapstructure:"service_account_json"`

	MaxOpenConnections       int         `mapstructure:"max_open_connections"`
	MaxIdleConnections       int         `mapstructure:"max_idle_connections"`
	MaxConnectionLifetimeRaw interface{} `mapstructure:"max_connection_lifetime"`
	MaxConnectionLifetime    time.Duration

	// TLS settings
	TLSCACert     []byte `mapstructure:"tls_ca_cert"`
	TLSClientKey  []byte `mapstructure:"tls_client_key"`
	TLSClientCert []byte `mapstructure:"tls_client_cert"`
	TLSSkipVerify bool   `mapstructure:"tls_skip_verify"`
	TLSServerName string `mapstructure:"tls_server_name"`

	// Preserve original config input
	RawConfig map[string]interface{}
}

// Load creates and validates a Config from raw config input (Vault passes this as map[string]interface{}).
func Load(raw map[string]interface{}) (*Config, error) {
	var cfg Config
	cfg.RawConfig = raw

	if err := mapstructure.WeakDecode(raw, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode mysql config: %w", err)
	}

	if cfg.ConnectionURL == "" {
		return nil, fmt.Errorf("connection_url is required")
	}

	if cfg.MaxOpenConnections <= 0 {
		cfg.MaxOpenConnections = 4
	}
	if cfg.MaxIdleConnections < 0 || cfg.MaxIdleConnections > cfg.MaxOpenConnections {
		cfg.MaxIdleConnections = cfg.MaxOpenConnections
	}

	if cfg.MaxConnectionLifetimeRaw == nil {
		cfg.MaxConnectionLifetimeRaw = "0s"
	}
	dur, err := parseutil.ParseDurationSecond(cfg.MaxConnectionLifetimeRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid max_connection_lifetime: %w", err)
	}
	cfg.MaxConnectionLifetime = dur

	return &cfg, nil
}

// Validate can be optionally reused to double-check if needed externally.
func (c *Config) Validate() error {
	if c.ConnectionURL == "" {
		return fmt.Errorf("connection_url is required")
	}
	return nil
}
