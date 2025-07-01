package dbsecretengine

import (
	"DatabasePluginVault/internal/dbengines/Engine"
	"context"
	"fmt"
	"strings"
	"sync"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	// version is your backend’s semver; bump when you cut a release
	version = "1.0.0"

	// configPathPrefix is the Vault storage prefix for configs
	configPathPrefix = "config/"
)

// connectionManager caches and tears down Engine instances.
type connectionManager struct {
	mu      sync.RWMutex
	engines map[string]Engine.Engine
}

func newConnectionManager() *connectionManager {
	return &connectionManager{
		engines: make(map[string]Engine.Engine),
	}
}

// Put stores a new Engine under name, returning any old one.
func (m *connectionManager) Put(name string, eng Engine.Engine) (old Engine.Engine) {
	m.mu.Lock()
	defer m.mu.Unlock()
	old = m.engines[name]
	m.engines[name] = eng
	return old
}

// ClearConnection closes and removes the Engine for name.
func (m *connectionManager) ClearConnection(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if eng, ok := m.engines[name]; ok {
		eng.Close()
		delete(m.engines, name)
	}
}

// ClearAll closes and removes *all* Engines (used on unmount).
func (m *connectionManager) ClearAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, eng := range m.engines {
		eng.Close()
	}
	m.engines = make(map[string]Engine.Engine)
}

// databaseBackend is the Vault logical backend.
type databaseBackend struct {
	*framework.Backend
	conn *connectionManager

	logger log.Logger
}

var _ logical.Factory = Factory

// Factory is the entrypoint called by Vault to mount this backend.
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {

	if conf == nil {
		return nil, fmt.Errorf("configuration is required")
	}
	b := Backend(conf)

	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

func Backend(conf *logical.BackendConfig) *databaseBackend {
	b := &databaseBackend{
		conn: newConnectionManager(),
	}
	b.Backend = &framework.Backend{
		Help:        backendHelp,
		BackendType: logical.TypeLogical,
		PathsSpecial: &logical.Paths{
			LocalStorage: []string{
				framework.WALPrefix,
			},
			SealWrapStorage: []string{
				"config",
			},
		},
		Secrets: []*framework.Secret{},
		Paths: framework.PathAppend(
			pathConfigurePluginConnection(b),
		),
		// Clean is called when the backend is unmounted; shut everything down.
		Clean: b.clean,
		// Invalidate is called when any storage key changes; used to clear a single entry.
		Invalidate: b.invalidate,
	}
	b.logger = conf.Logger
	return b
}

const backendHelp = `
MySQL-only Database Secrets Engine (clean refactor).

Configure connection info via the config/<name> endpoint.
`

// clean tears down every Engine instance (called on unmount).
func (b *databaseBackend) clean(context.Context) {
	b.conn.ClearAll()
}

// invalidate is called when any storage key changes.
// We only care about "config/<name>" keys, to clear that single cache.
func (b *databaseBackend) invalidate(ctx context.Context, key string) {
	if strings.HasPrefix(key, configPathPrefix) {
		name := strings.TrimPrefix(key, configPathPrefix)
		b.conn.ClearConnection(name)
	}
}
