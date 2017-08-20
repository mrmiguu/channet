package safedb

import "sync"

type Type struct {
	lock   sync.RWMutex
	lookup map[string]chan interface{}
}

func New() *Type {
	return &Type{lookup: make(map[string]chan interface{})}
}

func (db *Type) Put(key string, val chan interface{}) {
	db.lock.Lock()
	db.lookup[key] = val
	db.lock.Unlock()
}

func (db *Type) Get(key string) chan interface{} {
	db.lock.RLock()
	defer db.lock.RUnlock()
	return db.lookup[key]
}

func (db *Type) Lookup(key string) (chan interface{}, bool) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	c, exists := db.lookup[key]
	return c, exists
}
