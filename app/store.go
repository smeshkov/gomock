package app

import (
	"sync"
)

type store struct {
	table map[string]interface{}
	lock  sync.RWMutex
}

func newStore() *store {
	return &store{
		table: map[string]interface{}{},
	}
}

func (s *store) Write(entity, key string, value interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	var table map[string]interface{}
	v, ok := s.table[entity]
	if !ok {
		table = map[string]interface{}{}
	} else {
		table = v.(map[string]interface{})
	}
	table[key] = value
	s.table[entity] = table
}

func (s *store) Read(entity, key string) (interface{}, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	v, ok := s.table[entity]
	if !ok {
		return nil, false
	}
	table := v.(map[string]interface{})
	obj, ok := table[key]
	return obj, ok
}

func (s *store) ReadAll(entity string) (map[string]interface{}, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	v, ok := s.table[entity]
	if !ok {
		return nil, false
	}
	table, ok := v.(map[string]interface{})
	return table, ok
}
