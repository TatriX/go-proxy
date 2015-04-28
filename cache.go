package main

import (
	"net/http"
	"sync"
	"time"
)

type Cache struct {
	mu      sync.RWMutex
	data    map[CacheKey]CacheData
	timeout time.Duration
}

type CacheKey struct {
	URI   string
	Range string
}

type CacheData struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	expires    time.Time
}

func NewCache(timeout time.Duration) *Cache {
	return &Cache{
		data:    make(map[CacheKey]CacheData),
		timeout: timeout,
	}
}

func (c *Cache) Set(key CacheKey, data CacheData) {
	data.expires = time.Now().Add(c.timeout)
	c.mu.Lock()
	c.data[key] = data
	c.mu.Unlock()
}

func (c *Cache) Get(key CacheKey) (data CacheData, ok bool) {
	c.mu.RLock()
	data, ok = c.data[key]
	c.mu.RUnlock()
	if ok && time.Now().After(data.expires) {
		ok = false
		c.Remove(key)
	}
	return
}

func (c *Cache) Remove(key CacheKey) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock() // defer is a bit slower then explicit call
}
