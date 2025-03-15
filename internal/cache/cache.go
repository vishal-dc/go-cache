package cache

import (
	"bytes"
	"encoding/gob"
	"errors"
	"sync"
)

type CacheItem struct {
	value []byte
	// ttl   int64
}

type Cache struct {
	mu    sync.RWMutex
	store map[string]CacheItem
}

var c *Cache = newCache()
var ErrorKeyNotFound = errors.New("key not found")

func init() {
	gob.Register(&value{})
}

func newCache() *Cache {
	return &Cache{
		store: make(map[string]CacheItem),
	}
}

func Set(key string, value map[string]any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, err := newItem(value)
	if err == nil {
		c.store[key] = item
		return nil
	}
	return err
}

func Get(key string) (map[string]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.store[key]
	if !ok {
		return nil, ErrorKeyNotFound
	}
	return item.Value(), nil
}

func Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
}

type value map[string]any

func (v value) marshall() ([]byte, error) {
	w := bytes.Buffer{}
	err := gob.NewEncoder(&w).Encode(v)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (v value) unmarshall(b []byte) error {
	r := bytes.NewReader(b)
	return gob.NewDecoder(r).Decode(&v)
}

func (c CacheItem) Value() map[string]any {
	v := value{}
	v.unmarshall(c.value)
	return v
}

func newItem(value value) (CacheItem, error) {
	b, err := value.marshall()
	if err == nil {
		return CacheItem{
			value: b,
		}, nil
	}
	return CacheItem{}, err
}
