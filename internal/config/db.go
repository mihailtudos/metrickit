package config

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func (sc *ServerConfig) InitPostgresDB(dsn string) (*sql.DB, error) {
	DB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open a DB connection: %w", err)
	}

	return DB, nil
}
