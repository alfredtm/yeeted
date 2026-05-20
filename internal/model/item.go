package model

import (
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Item struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Value     int       `json:"value"`
	CreatedAt time.Time `json:"createdAt"`
}

type Store struct {
	mu     sync.RWMutex
	items  map[string]Item
	nextID atomic.Uint64
}

func NewStore() *Store {
	return &Store{items: make(map[string]Item)}
}

func (s *Store) List() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Item, 0, len(s.items))
	for _, it := range s.items {
		out = append(out, it)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (s *Store) Get(id string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	it, ok := s.items[id]
	return it, ok
}

func (s *Store) Create(name string, value int) Item {
	id := strconv.FormatUint(s.nextID.Add(1), 10)
	it := Item{
		ID:        id,
		Name:      name,
		Value:     value,
		CreatedAt: time.Now().UTC(),
	}
	s.mu.Lock()
	s.items[id] = it
	s.mu.Unlock()
	return it
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}
