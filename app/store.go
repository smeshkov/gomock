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
	v, ok := s.table[entity]
	if !ok {
		table = map[string]any{}
	} else {
		table = v.(map[string]any)
	}
	table[key] = value
	s.table[entity] = table
}

func (s *store) Read(entity, key string) (any, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	v, ok := s.table[entity]
	if !ok {
		return nil, false
	}
	table := v.(map[string]any)
	obj, ok := table[key]
	return obj, ok
}

func (s *store) ReadAll(entity string) (map[string]any, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	v, ok := s.table[entity]
	if !ok {
		return nil, false
	}
	table, ok := v.(map[string]any)
	return table, ok
}
