// Package store define repository interface.
package store

type Stats struct {
	URLNumber  int
	UserNumber int
}

type Repository interface {
	Set(key, val, userID string) error
	Get(key string) (string, error)
	GetAllByID(id string) (map[string]string, error)
	Delete(urlID, userID string) error
	Ping() error
	GetStats() (Stats, error)
	Close() error
}
