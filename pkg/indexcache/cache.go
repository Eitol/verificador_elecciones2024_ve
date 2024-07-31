package indexcache

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

const contention = 100

type IndexCache interface {
	GetLatest() (*CacheItem, error)
	SetLatest(value string) error
}

type indexCache struct {
	path            string
	mutex           sync.Mutex
	contentionCount int
}

type CacheItem struct {
	Latest string    `json:"latest"`
	Date   time.Time `json:"date"`
}

func NewIndexCache(cacheFilePath string) IndexCache {
	return &indexCache{path: cacheFilePath, contentionCount: 0}
}

func (c *indexCache) GetLatest() (*CacheItem, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	f, err := os.ReadFile(c.path)
	if err != nil {
		return nil, err
	}
	var fc CacheItem
	err = json.Unmarshal(f, &fc)
	if err != nil {
		return nil, err
	}
	return &fc, nil
}

func (c *indexCache) SetLatest(value string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.contentionCount++
	if c.contentionCount < contention {
		return nil
	}
	ci := CacheItem{
		Latest: value,
		Date:   time.Now(),
	}
	b, err := json.Marshal(ci)
	if err != nil {
		return err
	}
	err = os.WriteFile(c.path, b, 0777)
	if err != nil {
		return err
	}
	return nil
}
