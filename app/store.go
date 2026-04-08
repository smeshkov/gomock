package app

import (
	"sync"
)

type store struct {
	table map[string]any
	lock  sync.RWMutex
}

func newStore() *store {
	return &store{
		table: map[string]any{},
	}
}

func (s *store) Write(entity, key string, value any) {
	s.lock.Lock()
	defer s.lock.Unlock()

	var table map[string]any

	val, exists := s.table[entity]
	if !exists {
		table = map[string]any{}
	} else {
		table, exists = val.(map[string]any)
		if !exists {
			table = map[string]any{}
		}
	}

	table[key] = value
	s.table[entity] = table
}

func (s *store) Read(entity, key string) (any, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	val, exists := s.table[entity]
	if !exists {
		return nil, false
	}

	table, isMap := val.(map[string]any)
	if !isMap {
		return nil, false
	}

	obj, found := table[key]

	return obj, found
}

func (s *store) ReadAll(entity string) (map[string]any, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	val, exists := s.table[entity]
	if !exists {
		return nil, false
	}

	table, isMap := val.(map[string]any)

	return table, isMap
}
