package cache

import "sync"

type SimpleCache struct {
	visited *sync.Map
}

func (s SimpleCache) ShouldVisit(url string) bool {
	_, loaded := s.visited.LoadOrStore(url, true)
	return !loaded
}

func NewSimpleCache() SimpleCache {
	return SimpleCache{&sync.Map{}}
}
