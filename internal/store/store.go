package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Repository interface {
	Set(key, val string) error
	Get(key string) (string, error)
	Close() error
}

type MapDB struct {
	DB map[string]string
}

func NewMapDB() *MapDB {
	return &MapDB{
		DB: make(map[string]string),
	}
}

func (db *MapDB) Set(key, val string) error {
	db.DB[key] = val
	return nil
}

func (db *MapDB) Get(key string) (string, error) {
	val, ok := db.DB[key]
	if !ok {
		return "", fmt.Errorf("key %s not found in database", key)
	}
	return val, nil
}

func (db *MapDB) Close() error {
	return nil
}

type Record struct {
	ID  string `json:"id"`
	URL string `json:"url"`
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
	//defer file.Close()
	if err != nil {
		return nil, err
	}

	var records []Record
	err = json.Unmarshal(data, &records)
	if err != nil {
		return nil, err
	}

	return &FileDB{DB: file, Cache: Records{records}}, nil
}

func (f *FileDB) Set(key, value string) error {
	for _, r := range f.Cache.Records {
		if r.ID == key {
			return nil
		}
	}

	r := Record{ID: key, URL: value}
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

func (f *FileDB) Close() error {
	return f.DB.Close()
}
