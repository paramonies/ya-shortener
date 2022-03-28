package store

import (
	"context"
	"errors"
	"fmt"
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
	ctx, cancel := context.WithTimeout(context.Background(), DBConnectTimeout)
	defer cancel()

	query := `
INSERT INTO urls 
(
    short,
    original,
    user_id
)
VALUES ($1, $2, $3)
RETURNING id
`
	var id string
	row := p.Conn.QueryRow(ctx, query, key, val, userID)
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("failed to insert new row")
		}
		return err
	}
	return nil
}

func (p *PostgresDB) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DBConnectTimeout)
	defer cancel()

	query := `
SELECT original
FROM urls WHERE short=$1
`
	fmt.Println(key)
	var original string
	row := p.Conn.QueryRow(ctx, query, key)
	if err := row.Scan(&original); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("failed to get original url")
		}
		return "", err
	}
	return original, nil
}

func (p *PostgresDB) GetAllByID(id string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DBConnectTimeout)
	defer cancel()

	query := `
SELECT short,original
FROM urls WHERE user_id=$1
`
	rows, err := p.Conn.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make(map[string]string)
	var short, original string
	for rows.Next() {
		err = rows.Scan(&short, &original)
		if err != nil {
			return nil, err
		}
		data[short] = original
	}
	return data, nil
}
func (p *PostgresDB) Close() error {
	return p.Conn.Close(context.Background())
}
