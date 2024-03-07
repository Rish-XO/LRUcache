package main

import (
	"container/list"
	"sync"
	"time"
)

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
