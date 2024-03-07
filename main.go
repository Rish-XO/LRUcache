package main

import (
	"container/list"
	"fmt"
	"sync"
	"time"
	"encoding/json"
	"net/http"
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



// Test code
func main() {
	// Instantiate the LRU cache with a capacity of 3
	cache := NewLRUCache(3)

	// Set key-value pairs with different expiration times
	cache.Set("key1", "value1", 5*time.Second)
	cache.Set("key2", "value2", 10*time.Second)
	cache.Set("key3", "value3", 15*time.Second)

	// Attempt to get a value
	value, ok := cache.Get("key1")
	fmt.Printf("Get key1: %s, ok: %v\n", value, ok)

	// Wait for a while to let some items expire
	time.Sleep(10 * time.Second)

	// Attempt to get a value that should have expired
	value, ok = cache.Get("key1")
	fmt.Printf("Get key1 after expiration: %s, ok: %v\n", value, ok)

	// Set a new item to test eviction
	cache.Set("key4", "value4", 20*time.Second)

	// Attempt to get the new item and an old item
	value, ok = cache.Get("key4")
	fmt.Printf("Get key4: %s, ok: %v\n", value, ok)
	value, ok = cache.Get("key2")
	fmt.Printf("Get key2: %s, ok: %v\n", value, ok)
}
