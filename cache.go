package main

import (
	"net/http"
	"sync"
	"time"
)

type Cache struct {
	sync.RWMutex
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

func (c *Cache) Get(key CacheKey) (data CacheData, ok bool) {
	c.RLock()
	data, ok = c.data[key]
	if ok && time.Now().After(data.expires) {
		ok = false
	}
	c.RUnlock() // defer is a bit slower then explicit call
	return
}

func (c *Cache) Set(key CacheKey, data CacheData) {
	data.expires = time.Now().Add(c.timeout)
	c.Lock()
	c.data[key] = data
	c.Unlock()
}
