package main

import (
	"container/list"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// CacheItem represents an item stored in the cache
type CacheItem struct {
	Key   string
	Value string
	Exp   time.Time // Expiration time for the cache item
}

// LRUCache represents the LRU cache
type LRUCache struct {
	capacity int
	items    map[string]*list.Element
	ll       *list.List
	mu       sync.Mutex
}

var cache *LRUCache // Declare cache as a global variable

// NewLRUCache creates a new LRUCache with the given capacity
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		ll:       list.New(),
	}
}

// Get retrieves the value associated with the key from the cache
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

// Set adds or updates a value in the cache with the specified expiration time
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

// removeOldest removes the oldest item from the cache
func (c *LRUCache) removeOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

// removeElement removes the specified element from the cache
func (c *LRUCache) removeElement(ele *list.Element) {
	c.ll.Remove(ele)
	item := ele.Value.(*CacheItem)
	delete(c.items, item.Key)
}

// handleSet handles the HTTP POST request to set a value in the cache
func handleSet(w http.ResponseWriter, r *http.Request) {
	type SetRequest struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		Exp   int    `json:"exp"`
	}

	var req SetRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	expiration := time.Duration(req.Exp) * time.Second
	cache.Set(req.Key, req.Value, expiration)

	w.WriteHeader(http.StatusOK)
}

// handleGet handles the HTTP GET request to retrieve a value from the cache
func handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	value, ok := cache.Get(key)
	if !ok {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"value": value})
}

func main() {
	cache = NewLRUCache(1024)

	r := mux.NewRouter()
	r.HandleFunc("/set", handleSet).Methods("POST")
	r.HandleFunc("/get", handleGet).Methods("GET")

    //cors middleware
	c := cors.Default().Handler(r)

	http.ListenAndServe(":8080", c)
}
