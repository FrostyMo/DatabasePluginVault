package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/vault/sdk/logical"
)

type DatabaseConfig struct {
	PluginName    string `json:"plugin_name" structs:"plugin_name" mapstructure:"plugin_name"`
	PluginVersion string `json:"plugin_version" structs:"plugin_version" mapstructure:"plugin_version"`

	ConnectionDetails map[string]interface{} `json:"connection_details" structs:"connection_details" mapstructure:"connection_details"`
	AllowedRoles      []string               `json:"allowed_roles" structs:"allowed_roles" mapstructure:"allowed_roles"`

	RootCredentialsRotateStatements []string `json:"root_credentials_rotate_statements" structs:"root_credentials_rotate_statements" mapstructure:"root_credentials_rotate_statements"`
	PasswordPolicy                  string   `json:"password_policy" structs:"password_policy" mapstructure:"password_policy"`
	CiName                          string   `json:"ci_name" structs:"ci_name" mapstructure:"ci_name"`
	Emails                          []string `json:"emails" structs:"emails" mapstructure:"emails"`
}

func ConfigPath(dbType, name string) string {
	return fmt.Sprintf("config/%s/%s", dbType, name)
}

func SaveDBConfig(ctx context.Context, s logical.Storage, dbType, name string, data interface{}) error {
	entry, err := logical.StorageEntryJSON(ConfigPath(dbType, name), data)
	if err != nil {
		return err
	}
	return s.Put(ctx, entry)
}

func LoadDBConfig[T any](ctx context.Context, s logical.Storage, dbType, name string) (*T, error) {
	entry, err := s.Get(ctx, ConfigPath(dbType, name))
	if err != nil || entry == nil {
		return nil, err
	}
	var conf T
	if err := json.Unmarshal(entry.Value, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

func DeleteDBConfig(ctx context.Context, s logical.Storage, dbType, name string) error {
	return s.Delete(ctx, ConfigPath(dbType, name))
}
