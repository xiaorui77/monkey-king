package engine

import (
	"hash/fnv"
)

type Store interface {
	Visit(url string)
	IsVisited(url string) bool
}

type defaultStore struct {
	set map[uint64]bool
}

func NewStore() Store {
	return &defaultStore{set: map[uint64]bool{}}
}

func (s *defaultStore) Visit(url string) {
	h := fnv.New64a()
	_, _ = h.Write([]byte(url))
	urlHash := h.Sum64()

	s.set[urlHash] = true
}

func (s *defaultStore) IsVisited(url string) bool {
	h := fnv.New64a()
	_, _ = h.Write([]byte(url))
	urlHash := h.Sum64()

	if v, ok := s.set[urlHash]; ok {
		return v
	}
	return false
}
