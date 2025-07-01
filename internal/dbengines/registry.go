package dbengines

import (
	"DatabasePluginVault/internal/dbengines/Engine"
	"DatabasePluginVault/internal/dbengines/mysql"
	"fmt"
)

var registry = map[string]func(map[string]interface{}) (Engine.Engine, error){
	"mysql": mysql.NewEngine,
	// add other DBs here later...
}

// New picks the right Engine constructor by name.
func New(name string, raw map[string]interface{}) (Engine.Engine, error) {
	f, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unsupported database engine %q", name)
	}
	return f(raw)
}
