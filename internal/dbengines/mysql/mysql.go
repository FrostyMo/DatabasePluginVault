package mysql

import (
	engine "DatabasePluginVault/internal/dbengines/Engine"
	"context"
	"database/sql"
	"fmt"
)

// Engine implements dbengines.Engine for MySQL.
type Engine struct {
	driver *MySQLDriver
}

// NewEngine is used by the registry.
func NewEngine(raw map[string]interface{}) (engine.Engine, error) {
	cfg, err := Load(raw) // Config decode+validate
	if err != nil {
		return nil, err
	}
	driver, err := NewConnectionProducer(cfg)
	if err != nil {
		return nil, err
	}
	return &Engine{driver: driver}, nil
}

// Connect returns *sql.DB directly—no interface assertion needed.
func (e *Engine) Connect(ctx context.Context) (*sql.DB, error) {
	return e.driver.Connect(ctx)
}

func (e *Engine) Close() error {
	return e.driver.Close()
}

func (e *Engine) NewUser(ctx context.Context, username, password string) error {
	db, err := e.driver.Connect(ctx)
	if err != nil {
		return err
	}
	query := fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s'", username, password)
	_, err = db.ExecContext(ctx, query)
	return err
}

func (e *Engine) UpdateUser(ctx context.Context, username, password string) error {
	db, err := e.driver.Connect(ctx)
	if err != nil {
		return err
	}
	query := fmt.Sprintf("ALTER USER '%s'@'%%' IDENTIFIED BY '%s'", username, password)
	_, err = db.ExecContext(ctx, query)
	return err
}

func (e *Engine) DeleteUser(ctx context.Context, username string) error {
	db, err := e.driver.Connect(ctx)
	if err != nil {
		return err
	}
	query := fmt.Sprintf("DROP USER IF EXISTS '%s'@'%%'", username)
	_, err = db.ExecContext(ctx, query)
	return err
}
