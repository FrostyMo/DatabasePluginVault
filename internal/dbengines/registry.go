package dbengines

import (
	"DatabasePluginVault/internal/dbengines/mysql"
	"fmt"

	dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
)

var registry = map[string]func(map[string]interface{}) (dbplugin.Database, error){
	"mysql": func(cfgMap map[string]interface{}) (dbplugin.Database, error) {
		cfg := &mysql.Config{}
		if err := cfg.Load(cfgMap); err != nil {
			return nil, fmt.Errorf("failed to load mysql config: %w", err)
		}
		return mysql.NewEngine(cfg)
	},
}

// Get returns the engine constructor for the given db type.
func Get(dbType string) (func(map[string]interface{}) (dbplugin.Database, error), error) {
	engineFn, ok := registry[dbType]
	if !ok {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
	return engineFn, nil
}

// Types returns a list of registered engine types.
func Types() []string {
	var types []string
	for k := range registry {
		types = append(types, k)
	}
	return types
}
