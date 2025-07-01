package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

type MySQLDriver struct {
	config        *Config
	db            *sql.DB
	connLock      sync.Mutex
	tlsConfigName string
}

// NewConnectionProducer creates a driver instance from a raw config map.
func NewConnectionProducer(raw map[string]interface{}) (*MySQLDriver, error) {
	cfg, err := Load(raw)
	if err != nil {
		return nil, err
	}

	return &MySQLDriver{
		config:   cfg,
		connLock: sync.Mutex{},
	}, nil
}

// Connect returns a pooled DB connection (cached if already connected).
func (m *MySQLDriver) Connect(ctx context.Context) (*sql.DB, error) {
	m.connLock.Lock()
	defer m.connLock.Unlock()

	// Reuse connection if healthy
	if m.db != nil {
		if err := m.db.PingContext(ctx); err == nil {
			return m.db, nil
		}
		_ = m.db.Close()
		m.db = nil
	}

	dsn := m.config.ConnectionURL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql open error: %w", err)
	}

	db.SetMaxOpenConns(m.config.MaxOpenConnections)
	db.SetMaxIdleConns(m.config.MaxIdleConnections)
	db.SetConnMaxLifetime(m.config.MaxConnectionLifetime)

	m.db = db
	return db, nil
}

// Close closes any open DB connection.
func (m *MySQLDriver) Close() error {
	m.connLock.Lock()
	defer m.connLock.Unlock()

	if m.db != nil {
		err := m.db.Close()
		m.db = nil
		return err
	}
	return nil
}

// NewUser creates a new MySQL user with the provided credentials.
func (m *MySQLDriver) NewUser(username, password string) error {
	query := fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s'", username, password)
	_, err := m.db.Exec(query)
	return err
}

// UpdateUser updates the password of an existing MySQL user.
func (m *MySQLDriver) UpdateUser(username, password string) error {
	query := fmt.Sprintf("ALTER USER '%s'@'%%' IDENTIFIED BY '%s'", username, password)
	_, err := m.db.Exec(query)
	return err
}

// DeleteUser removes a user from the MySQL instance.
func (m *MySQLDriver) DeleteUser(username string) error {
	query := fmt.Sprintf("DROP USER IF EXISTS '%s'@'%%'", username)
	_, err := m.db.Exec(query)
	return err
}

// Type returns the database type identifier.
func (m *MySQLDriver) Type() string {
	return "mysql"
}
