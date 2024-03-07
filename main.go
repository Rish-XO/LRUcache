package main

import (
	"container/list"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// LRU Cache implementation
type CacheItem struct {
	Key   string
	Value string
	Exp   time.Time
}

type LRUCache struct {
	capacity int
	items    map[string]*list.Element
	ll       *list.List
	mu       sync.Mutex
}

var cache *LRUCache // Declare cache as a global variable

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		ll:       list.New(),
	}
}

func (c *LRUCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.items[key]; ok {
		c.ll.MoveToFront(ele)
		item := ele.Value.(*CacheItem)
		if time.Now().After(item.Exp) {
			c.removeElement(ele)
			return "", false
		}
		return item.Value, true
	}
	return "", false
}

func (c *LRUCache) Set(key string, value string, exp time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.items[key]; ok {
		c.ll.MoveToFront(ele)
		item := ele.Value.(*CacheItem)
		item.Value = value
		item.Exp = time.Now().Add(exp)
	} else {
		ele := c.ll.PushFront(&CacheItem{Key: key, Value: value, Exp: time.Now().Add(exp)})
		c.items[key] = ele
		if c.ll.Len() > c.capacity {
			c.removeOldest()
		}
	}
}

func (c *LRUCache) removeOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *LRUCache) removeElement(ele *list.Element) {
	c.ll.Remove(ele)
	item := ele.Value.(*CacheItem)
	delete(c.items, item.Key)
}

func handleSet(w http.ResponseWriter, r *http.Request) {
	// Parse the key, value, and expiration from the request
	// This is a simplified example; you might want to use a more robust method for parsing
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")
	expStr := r.URL.Query().Get("exp")
	exp, err := strconv.Atoi(expStr)
	if err != nil {
		http.Error(w, "Invalid expiration", http.StatusBadRequest)
		return
	}

	// Call the Set method on the cache
	cache.Set(key, value, time.Duration(exp)*time.Second)

	// Send a success response
	w.WriteHeader(http.StatusOK)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	// Parse the key from the request
	key := r.URL.Query().Get("key")

	// Call the Get method on the cache
	value, ok := cache.Get(key)
	if !ok {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	// Send the value in the response
	json.NewEncoder(w).Encode(map[string]string{"value": value})
}



func main() {
	// Instantiate the LRU cache with a capacity of 1024
	cache := NewLRUCache(1024)

	// Set up the HTTP server
	r := mux.NewRouter()
	r.HandleFunc("/set", handleSet).Methods("POST")
	r.HandleFunc("/get", handleGet).Methods("GET")

	// Start the server
	http.ListenAndServe(":8080", r)
}
