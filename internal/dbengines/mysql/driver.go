package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	// register the mysql driver
	_ "github.com/go-sql-driver/mysql"
)

// MySQLDriver manages the *sql.DB pool.
type MySQLDriver struct {
	cfg *Config
	mu  sync.Mutex
	db  *sql.DB
}

// NewConnectionProducer builds a driver from a typed Config.
func NewConnectionProducer(cfg *Config) (*MySQLDriver, error) {
	return &MySQLDriver{cfg: cfg}, nil
}

// Connect returns a cached *sql.DB or opens a new one.
func (d *MySQLDriver) Connect(ctx context.Context) (*sql.DB, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	log.Print("Inside connect sql")
	if d.db != nil {
		log.Print("DB NIl pinging")
		if err := d.db.PingContext(ctx); err == nil {
			return d.db, nil
		}
		d.db.Close()
		d.db = nil
	}
	log.Print("Opening new connection")
	db, err := sql.Open("mysql", d.cfg.ConnectionURL)
	if err != nil {
		log.Print("Failed to open")
		return nil, fmt.Errorf("mysql open: %w", err)
	}
	// enforce an actual connect + auth step
	if err := db.PingContext(ctx); err != nil {
		log.Print("Ping failed:", err)
		db.Close()
		return nil, fmt.Errorf("mysql ping: %w", err)
	}
	db.SetMaxOpenConns(d.cfg.MaxOpenConnections)
	db.SetMaxIdleConns(d.cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(d.cfg.MaxConnectionLifetime)
	log.Print(fmt.Sprintf("Received db %v", db))
	d.db = db
	return db, nil
}

// Close tears down the DB pool.
func (d *MySQLDriver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.db != nil {
		d.db.Close()
		d.db = nil
	}
	return nil
}
