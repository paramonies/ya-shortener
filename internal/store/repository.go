// Package store define repository interface.
package store

type Repository interface {
	Set(key, val, userID string) error
	Get(key string) (string, error)
	GetAllByID(id string) (map[string]string, error)
	Delete(urlID, userID string) error
	Ping() error
	Close() error
}
