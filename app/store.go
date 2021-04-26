package app

import (
	"fmt"
	"sync"
)

type store struct {
	sm sync.Map
}

func (s *store) Write(entity, key string, value interface{}) {
	s.sm.Store(toStoreKey(entity, key), value)
}

func (s *store) Read(entity, key string) (interface{}, bool) {
	return s.sm.Load(toStoreKey(entity, key))
}

func toStoreKey(entity, key string) string {
	return fmt.Sprintf("%s|%s", entity, key)
}
