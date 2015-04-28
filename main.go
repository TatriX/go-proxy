// simple caching proxy
package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// see parseFlags() for default values
var config = struct {
	NetAddr string
	Log     bool
	Timeout time.Duration
}{}

var cache *Cache

func main() {
	parseFlags()
	cache = NewCache(config.Timeout)
	panic(http.ListenAndServe(config.NetAddr, http.HandlerFunc(proxyHandler)))
}

func parseFlags() {
	flag.StringVar(&config.NetAddr, "listen", "localhost:8080", "listen to")
	flag.BoolVar(&config.Log, "log", false, "print debug info")
	flag.DurationVar(&config.Timeout, "timeout", 60*time.Second, "cache timeout")
	flag.Parse()
}

// proxyHandler func will handle request using global cache for GET method.
// All other methods (and cache misses) will be be served via http.Client
func proxyHandler(w http.ResponseWriter, r *http.Request) {
	if !r.URL.IsAbs() {
		http.Error(w, "Non abs url", http.StatusInternalServerError)
		return
	}
	var data CacheData
	var err error
	if r.Method == "GET" {
		key := CacheKey{
			URI:   r.RequestURI,
			Range: r.Header.Get("Range"),
		}
		var ok bool
		data, ok = cache.Get(key)
		if ok {
			logf("Cache hit for %+v", key)
		} else {
			logf("Cache miss for %+v", key)
			data, err = doRequest(w, r)
			cache.Set(key, data) // also cache errors
		}
	} else {
		data, err = doRequest(w, r)
	}

	if err != nil {
		logf("Request: %s", err)
	}

	if data.StatusCode >= http.StatusBadRequest {
		writeError(w, data.StatusCode)
	} else {
		copyHeaders(w.Header(), data.Header)
		w.WriteHeader(data.StatusCode)
		w.Write(data.Body)
	}
}

// doRequest makes actual request and creates CacheData for proxyHandler
func doRequest(w http.ResponseWriter, r *http.Request) (data CacheData, err error) {
	prepareRequest(r)
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		data.StatusCode = http.StatusBadRequest // really?
		return
	}

	defer resp.Body.Close()

	data.StatusCode = resp.StatusCode
	data.Header = resp.Header

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		data.StatusCode = http.StatusInternalServerError
		return
	}
	data.Body = buf.Bytes()
	return
}

// prepareRequest makes http.Client happy
func prepareRequest(r *http.Request) {
	r.RequestURI = ""
	if r.URL.Scheme == "" {
		r.URL.Scheme = "http"
	}
	r.URL.Scheme = strings.ToLower(r.URL.Scheme)
}

func copyHeaders(dst, src http.Header) {
	for k, _ := range dst {
		dst.Del(k)
	}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

func writeError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

func logf(fmt string, args ...interface{}) {
	if config.Log {
		log.Printf(fmt, args...)
	}
}
