package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	_ "github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
)

var (
	DBConnectTimeout       = 3 * time.Second
	ErrConstraintViolation = errors.New("original url conflict")
	ErrGone                = errors.New("gone")
	MigDirName             = "migrations"
)

type PostgresDB struct {
	Conn *pgxpool.Pool
}

func NewPostgresDB(dsn string) (*PostgresDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DBConnectTimeout)
	defer cancel()
	conn, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		log.Printf("failed to establishe a connection with a PostgreSQL server: %v", err)
		return nil, err
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("failed to open DB: %v", err)
		return nil, err
	}

	rootDir, err := os.Getwd()
	if err != nil {
		log.Println(rootDir)
		return nil, err
	}

	MigDirPath := fmt.Sprintf("%s/%s", rootDir, MigDirName)

	migrations := &migrate.FileMigrationSource{
		Dir: MigDirPath,
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		log.Printf("failed to apply migrations: %v", err)
		return nil, err
	}
	log.Printf("Applied %d migrations!\n", n)

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
VALUES ($1, $2, $3, false)
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
WHERE short = $1 and user_id = $2 and deleted= false
`
	rows, err := p.Conn.Query(ctx, query, urlID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("record not found or already deleted")
		}
		return err
	}

	defer rows.Close()
	return nil
}

func (p *PostgresDB) Ping() error {
	return p.Conn.Ping(context.Background())
}

func (p *PostgresDB) Close() error {
	p.Conn.Close()
	return nil
}
