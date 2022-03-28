package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Repository interface {
	Set(key, val, userID string) error
	Get(key string) (string, error)
	GetAllByID(id string) (map[string]string, error)
	Close() error
}

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

func (db *MapDB) Close() error {
	return nil
}

type Record struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	UserID string `json:"user_id"`
}

type Records struct {
	Records []Record `json:"records"`
}

type FileDB struct {
	DB    *os.File
	Cache Records
}

func NewFileDB(path string) (*FileDB, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(file)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	var records Records
	err = json.Unmarshal(data, &records)
	if err != nil {
		return nil, err
	}

	return &FileDB{DB: file, Cache: Records{records.Records}}, nil
}

func (f *FileDB) Set(key, value, userID string) error {
	for _, r := range f.Cache.Records {
		if r.ID == key {
			return nil
		}
	}

	r := Record{ID: key, URL: value, UserID: userID}
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	f.Cache.Records = append(f.Cache.Records, r)
	_, err = f.DB.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (f *FileDB) Get(key string) (string, error) {
	for _, r := range f.Cache.Records {
		if r.ID == key {
			return r.URL, nil
		}
	}

	return "", fmt.Errorf("key %s not found in database", key)
}

func (f *FileDB) GetAllByID(id string) (map[string]string, error) {
	data := make(map[string]string)
	for _, record := range f.Cache.Records {
		if record.UserID == id {
			data[record.ID] = record.URL
		}
	}
	return data, nil
}

func (f *FileDB) Close() error {
	return f.DB.Close()
}
