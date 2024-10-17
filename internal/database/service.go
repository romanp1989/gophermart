package database

import (
	"database/sql"
	"sync"
)

type RDB struct {
	DB *sql.DB
	MU sync.RWMutex
}

func NewDB(cfg *Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.Dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.MaxIdleConn)
	db.SetMaxOpenConns(cfg.MaxOpenConn)
	db.SetConnMaxIdleTime(cfg.MaxLifetimeConn)

	return db, nil
}
