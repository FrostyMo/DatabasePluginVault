package backend

import (
	path "DatabasePluginVault/path/config"
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	operationPrefixDatabase = "database"
)

// connectionManager handles caching and lifecycle of engine instances.
type connectionManager struct {
	mu      sync.RWMutex
	engines map[string]dbengines.Engine
}

func newConnectionManager() *connectionManager {
	return &connectionManager{
		engines: make(map[string]dbengines.Engine),
	}
}

// Put installs a new engine under name, returning any old one.
func (m *connectionManager) Put(name string, eng dbengines.Engine) (old dbengines.Engine) {
	m.mu.Lock()
	defer m.mu.Unlock()
	old = m.engines[name]
	m.engines[name] = eng
	return old
}

// ClearConnection closes & removes the engine for name.
func (m *connectionManager) ClearConnection(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if eng, ok := m.engines[name]; ok {
		_ = eng.Close()
		delete(m.engines, name)
	}
	return nil
}

// databaseBackend wires Vault paths to your engine.
type databaseBackend struct {
	*framework.Backend
	conn *connectionManager
}

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := &databaseBackend{
		conn: newConnectionManager(),
	}
	b.Backend = &framework.Backend{
		Help:           "MySQL-only database secrets engine (clean refactor)",
		RunningVersion: "v1.0.0", // bump as you like
		Paths: []*framework.Path{
			path.PathConfigurePluginConnection(b),
			// (later you can append /static-roles, /rotate, etc.)
		},
		BackendType: logical.TypeLogical,
	}
	return b, nil
}

// storeConfig persists the DatabaseConfig at config/<name>.
func (b *databaseBackend) storeConfig(ctx context.Context, s logical.Storage, name string, config *storage.DatabaseConfig) error {
	entry, err := logical.StorageEntryJSON(fmt.Sprintf("config/%s", name), config)
	if err != nil {
		return fmt.Errorf("failed to marshal config/%s: %w", name, err)
	}
	return s.Put(ctx, entry)
}

// connectionExistenceCheck implements framework.ExistenceFunc for config/<name>.
func (b *databaseBackend) connectionExistenceCheck() framework.ExistenceFunc {
	return func(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
		name := data.Get("name").(string)
		entry, err := req.Storage.Get(ctx, fmt.Sprintf("config/%s", name))
		if err != nil {
			return false, err
		}
		return entry != nil, nil
	}
}
