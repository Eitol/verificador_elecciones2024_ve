package results

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func NewJsonFileResultStore(dirPath string) (ResultStore, error) {
	if !dirExist(dirPath) {
		err := os.MkdirAll(dirPath, 0777)
		if err != nil {
			return nil, fmt.Errorf("error creating directory: %w", err)
		}
	}
	store := &jsonFileResultStore{
		dirPath: dirPath,
		cache:   make(map[string]Result),
	}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			data, err := os.ReadFile(filepath.Join(dirPath, file.Name()))
			if err != nil {
				return nil, err
			}

			var result Result
			err = json.Unmarshal(data, &result)
			if err != nil {
				os.Remove(filepath.Join(dirPath, file.Name()))
				return nil, fmt.Errorf("error unmarshalling json cache file %s: %w", file.Name(), err)
			}

			store.cache[result.Url] = result
		}
	}

	return store, nil
}

type jsonFileResultStore struct {
	dirPath string
	cache   map[string]Result
	mutex   sync.RWMutex
}

func (s *jsonFileResultStore) StoreResult(result Result) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if len(result.Bytes) > 0 {
		// Write the bytes to a separate file
		fileName := filepath.Join(s.dirPath, filepath.Base(result.Url))
		err := os.WriteFile(fileName, result.Bytes, 0777)
		if err != nil {
			return err
		}
	}

	// Don't write bytes to JSON
	result.Bytes = nil

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	// Write the json to a file
	filename := filepath.Base(result.Url) + ".json"
	err = os.WriteFile(filepath.Join(s.dirPath, filename), data, 0777)
	if err != nil {
		return err
	}

	// Cache the result
	s.cache[result.Url] = result

	return nil
}

func (s *jsonFileResultStore) GetResultByURL(url string) (*Result, error) {
	s.mutex.RLock()
	result, ok := s.cache[url]
	s.mutex.RUnlock()
	if ok {
		return &result, nil
	}
	return nil, ErrResultNotFound
}

func dirExist(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}
