package store

import (
	"fmt"
	"sync"
)

type InMemoryKVStore struct {
	capacity int
	store    map[interface{}]interface{}
	lock     *sync.RWMutex
}

func NewInMemoryKVStore(capacity int) AdvancedKVStore {
	return InMemoryKVStore{
		capacity: capacity,
		store:    make(map[interface{}]interface{}),
		lock:     new(sync.RWMutex),
	}
}

func (s InMemoryKVStore) withRead(cb func()) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	cb()
}

func (s InMemoryKVStore) withWrite(cb func()) {
	s.lock.Lock()
	defer s.lock.Unlock()
	cb()
}

func (s InMemoryKVStore) Get(key interface{}) (res interface{}, err error) {
	s.withRead(func() {
		res = s.store[key]
	})
	return
}

func (s InMemoryKVStore) Query(filter func(record interface{}) bool) (res []interface{}, err error) {
	res = []interface{}{}
	s.withRead(func() {
		for _, v := range s.store {
			if filter(v) {
				res = append(res, v)
			}
		}
	})
	return
}

func (s InMemoryKVStore) BulkGet(keys []interface{}) (res []interface{}, err error) {
	res = make([]interface{}, len(keys), len(keys))
	s.withRead(func() {
		for i, k := range keys {
			res[i] = s.store[k]
		}
	})
	return
}

func (s InMemoryKVStore) Has(key interface{}) (exists bool, err error) {
	s.withRead(func() {
		_, exists = s.store[key]
	})
	return
}

func (s InMemoryKVStore) Put(key interface{}, value interface{}) (success bool, err error) {
	s.withWrite(func() {
		if len(s.store) < s.capacity {
			s.setWithoutLock(key, value)
			success = true
			return
		}
		success = false
		err = fmt.Errorf("exceeded maximum capacity %d", s.capacity)
	})
	return
}

func (s InMemoryKVStore) BulkPut(bulk map[interface{}]interface{}) (success bool, err error) {
	s.withWrite(func() {
		for k, v := range bulk {
			if len(s.store) < s.capacity {
				success = false
				err = fmt.Errorf("exceeded maximum capacity %d", s.capacity)
				return
			}
			s.setWithoutLock(k, v)
			success = true
		}
	})
	return
}

func (s InMemoryKVStore) Update(key interface{}, value interface{}) (success bool, err error) {
	s.withWrite(func() {
		if _, exists := s.store[key]; !exists {
			success = false
			err = fmt.Errorf("record %s does not exist", key)
			return
		}
		s.setWithoutLock(key, value)
		success = true
	})
	return
}

func (s InMemoryKVStore) Delete(key interface{}) (success bool, err error) {
	s.withWrite(func() {
		if _, exists := s.store[key]; !exists {
			success = false
			err = fmt.Errorf("record %s does not exist", key)
			return
		}
		delete(s.store, key)
		success = true
	})
	return
}

func (s InMemoryKVStore) setWithoutLock(key interface{}, value interface{}) {
	s.store[key] = value
}
