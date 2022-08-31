package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Record struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	UserID string `json:"user_id"`
}

type RecordsCache struct {
	Records []Record `json:"records"`
}

type FileDB struct {
	DB    *os.File
	Cache RecordsCache
}

func NewFileDB(path string) (*FileDB, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return nil, err
	}

	var records RecordsCache

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if fileInfo.Size() != 0 {
		data, err := io.ReadAll(file)
		defer file.Close()
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &records)
		if err != nil {
			return nil, err
		}
	}

	return &FileDB{DB: file, Cache: RecordsCache{records.Records}}, nil
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

func (f *FileDB) GetStats() (Stats, error) {
	users := make(map[string]struct{})
	for _, r := range f.Cache.Records {
		if _, ok := users[r.UserID]; !ok {
			users[r.UserID] = struct{}{}
		}
	}

	return Stats{len(f.Cache.Records), len(f.Cache.Records)}, nil
}

func (f *FileDB) Delete(urlID, userID string) error {
	return nil
}

func (f *FileDB) Ping() error {
	return nil
}

func (f *FileDB) Close() error {
	return f.DB.Close()
}
