package Engine

import (
	"context"
	"database/sql"
)

// Engine is the minimal interface your path handlers need.
type Engine interface {
	// Connect returns a live *sql.DB.
	Connect(ctx context.Context) (*sql.DB, error)
	// Close tears down the driver.
	Close() error

	// Static‐role credential ops:
	NewUser(ctx context.Context, username, password string) error
	UpdateUser(ctx context.Context, username, password string) error
	DeleteUser(ctx context.Context, username string) error
}
