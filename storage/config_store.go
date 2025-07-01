package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/vault/sdk/logical"
)

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
