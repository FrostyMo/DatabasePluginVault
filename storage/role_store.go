package storage

import (
	"DatabasePluginVault/role"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/vault/sdk/logical"
)

func rolePath(dbType, name string) string {
	return fmt.Sprintf("roles/%s/%s", dbType, name)
}

func SaveRole(ctx context.Context, s logical.Storage, dbType, name string, r *role.RoleEntry) error {
	entry, err := logical.StorageEntryJSON(rolePath(dbType, name), r)
	if err != nil {
		return err
	}
	return s.Put(ctx, entry)
}

func LoadRole(ctx context.Context, s logical.Storage, dbType, name string) (*role.RoleEntry, error) {
	entry, err := s.Get(ctx, rolePath(dbType, name))
	if err != nil || entry == nil {
		return nil, err
	}
	var r role.RoleEntry
	if err := json.Unmarshal(entry.Value, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func DeleteRole(ctx context.Context, s logical.Storage, dbType, name string) error {
	return s.Delete(ctx, rolePath(dbType, name))
}
