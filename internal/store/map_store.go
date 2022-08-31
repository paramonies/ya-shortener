package store

import "fmt"

type MapDB struct {
	DB map[string]map[string]string
}

func NewMapDB() *MapDB {
	return &MapDB{
		DB: make(map[string]map[string]string),
	}
}

func (db *MapDB) Set(key, val, userID string) error {
	db.DB[key] = map[string]string{
		"url":    val,
		"userID": userID,
	}
	return nil
}

func (db *MapDB) Get(key string) (string, error) {
	val, ok := db.DB[key]
	if !ok {
		return "", fmt.Errorf("key %s not found in database", key)
	}
	return val["url"], nil
}

func (db *MapDB) GetAllByID(id string) (map[string]string, error) {
	data := make(map[string]string)
	for key, row := range db.DB {
		if row["userID"] == id {
			data[key] = row["url"]
		}
	}
	return data, nil
}

func (db *MapDB) GetStats() (Stats, error) {
	users := make(map[string]struct{})
	for _, row := range db.DB {
		if _, ok := users[row["userID"]]; !ok {
			users[row["userID"]] = struct{}{}
		}
	}
	return Stats{len(db.DB), len(users)}, nil
}

func (db *MapDB) Delete(urlID, userID string) error {
	return nil
}

func (db *MapDB) Ping() error {
	return nil
}

func (db *MapDB) Close() error {
	return nil
}
