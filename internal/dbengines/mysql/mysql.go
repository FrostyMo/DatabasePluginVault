package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/database/dbplugin/v5"
)

type Engine struct {
	producer *MySQLDriver
}

func NewEngine(cfg *Config) (*Engine, error) {
	producer, err := NewConnectionProducer(cfg)
	if err != nil {
		return nil, err
	}
	return &Engine{producer: producer}, nil
}

func (e *Engine) Type() (string, error) {
	return "mysql", nil
}

func (e *Engine) Initialize(ctx context.Context, req dbplugin.InitializeRequest) (dbplugin.InitializeResponse, error) {
	cfg := &Config{}
	for k, v := range req.Config {
		cfg.RawConfig[k] = v
	}

	engine, err := NewEngine(cfg)
	if err != nil {
		return dbplugin.InitializeResponse{}, err
	}
	e.producer = engine.producer

	return dbplugin.InitializeResponse{
		Config: req.Config,
	}, nil
}

func (e *Engine) NewUser(ctx context.Context, req dbplugin.NewUserRequest) (dbplugin.NewUserResponse, error) {
	username := req.UsernameConfig.RoleName + "-" + fmt.Sprintf("%d", time.Now().Unix())
	password := req.Password

	query := fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s';", username, password)
	db, err := e.producer.Connect(ctx)
	if err != nil {
		return dbplugin.NewUserResponse{}, err
	}

	if _, err := db.ExecContext(ctx, query); err != nil {
		return dbplugin.NewUserResponse{}, err
	}

	return dbplugin.NewUserResponse{Username: username}, nil
}

func (e *Engine) UpdateUser(ctx context.Context, req dbplugin.UpdateUserRequest) (dbplugin.UpdateUserResponse, error) {
	if req.Password == nil {
		return dbplugin.UpdateUserResponse{}, fmt.Errorf("password update required")
	}

	db, err := e.producer.Connect(ctx)
	if err != nil {
		return dbplugin.UpdateUserResponse{}, err
	}

	query := fmt.Sprintf("ALTER USER '%s'@'%%' IDENTIFIED BY '%s';", req.Username, *req.Password)
	if _, err := db.ExecContext(ctx, query); err != nil {
		return dbplugin.UpdateUserResponse{}, err
	}

	return dbplugin.UpdateUserResponse{}, nil
}

func (e *Engine) DeleteUser(ctx context.Context, req dbplugin.DeleteUserRequest) (dbplugin.DeleteUserResponse, error) {
	db, err := e.producer.Connect(ctx)
	if err != nil {
		return dbplugin.DeleteUserResponse{}, err
	}

	query := fmt.Sprintf("DROP USER IF EXISTS '%s'@'%%';", req.Username)
	if _, err := db.ExecContext(ctx, query); err != nil {
		return dbplugin.DeleteUserResponse{}, err
	}

	return dbplugin.DeleteUserResponse{}, nil
}

func (e *Engine) Close() error {
	return e.producer.Close()
}
