package dbengines

import "context"

// Engine is the minimal interface your backend and path handlers need.
type Engine interface {
	// Connect is used to verify credentials and/or cache the connection.
	Connect(ctx context.Context) (interface{}, error)
	// Close tears down any pooled connections.
	Close() error

	// Credential operations (used by static-role / creds paths):
	NewUser(ctx context.Context, username, password string) error
	UpdateUser(ctx context.Context, username, password string) error
	DeleteUser(ctx context.Context, username string) error
}
