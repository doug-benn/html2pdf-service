package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	mu  sync.Mutex
	dsn string
	db  *sql.DB
}

func NewDB() *DB {
	return &DB{}
}

func (p *DB) Get(dsn string) (*sql.DB, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.db != nil && p.dsn == dsn {
		return p.db, nil
	}
	if p.db != nil {
		_ = p.db.Close()
		p.db = nil
		p.dsn = ""
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	p.db = db
	p.dsn = dsn
	return db, nil
}

func VerifySchema(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, `SELECT fn_verify_tokens_schema();`); err != nil {
		return fmt.Errorf("auth-service schema check failed: %w", err)
	}
	return nil
}
