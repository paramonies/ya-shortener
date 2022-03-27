package store

import (
	"context"
	"github.com/jackc/pgx/v4"
	"time"
)

var (
	DBConnectTimeout = 1 * time.Second
)

type PostgresDB struct {
	Conn *pgx.Conn
}

func NewPostgresDB(dns string) (*PostgresDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DBConnectTimeout)
	defer cancel()
	conn, err := pgx.Connect(ctx, dns)
	if err != nil {
		return nil, err
	}

	return &PostgresDB{Conn: conn}, nil
}

func (p *PostgresDB) Set(key, val, userID string) error {
	return nil
}

func (p *PostgresDB) Get(key string) (string, error) {
	return "", nil
}
func (p *PostgresDB) GetAllByID(id string) map[string]string {
	return nil
}
func (p *PostgresDB) Close() error {
	return nil
}
