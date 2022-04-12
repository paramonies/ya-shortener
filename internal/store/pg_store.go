package store

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"time"
)

var (
	DBConnectTimeout = 1 * time.Second

	ErrConstraintViolation = errors.New("original url conflict")
	ErrGone                = errors.New("gone")
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
    user_id,
    deleted
)
VALUES ($1, $2, $3, true)
RETURNING id
`
	var id string
	row := p.Conn.QueryRow(ctx, query, key, val, userID)
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("failed to insert new row")
		}

		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerrcode.IsIntegrityConstraintViolation(pgerr.SQLState()) {
				return ErrConstraintViolation
			}
		}
		return err
	}
	return nil
}

func (p *PostgresDB) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DBConnectTimeout)
	defer cancel()

	query := `
SELECT original, deleted
FROM urls WHERE short=$1
`
	var original string
	var deleted bool
	row := p.Conn.QueryRow(ctx, query, key)
	if err := row.Scan(&original, &deleted); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("failed to get original url")
		}
		return "", err
	}

	if deleted {
		return "", ErrGone
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

func (p *PostgresDB) Delete(urlID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), DBConnectTimeout)
	defer cancel()

	query := `
UPDATE urls 
SET deleted = true
WHERE short = $1 and user_id = $2
`
	rows, err := p.Conn.Query(ctx, query, urlID, userID)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func (p *PostgresDB) Ping() error {
	return p.Conn.Ping(context.Background())
}

func (p *PostgresDB) Close() error {
	return p.Conn.Close(context.Background())
}
