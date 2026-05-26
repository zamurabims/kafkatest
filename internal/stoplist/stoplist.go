package stoplist

import (
	"strings"
	"sync"
)

type StopList struct {
	mu    sync.RWMutex
	words map[string]struct{}
}

func New() *StopList {
	return &StopList{words: make(map[string]struct{})}
}

func (s *StopList) Add(word string) {
	word = normalize(word)
	s.mu.Lock()
	s.words[word] = struct{}{}
	s.mu.Unlock()
}

func (s *StopList) Remove(word string) {
	word = normalize(word)
	s.mu.Lock()
	delete(s.words, word)
	s.mu.Unlock()
}

func (s *StopList) Contains(word string) bool {
	s.mu.RLock()
	_, ok := s.words[normalize(word)]
	s.mu.RUnlock()
	return ok
}

func (s *StopList) List() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.words))
	for w := range s.words {
		out = append(out, w)
	}
	return out
}

func normalize(w string) string {
	return strings.ToLower(strings.TrimSpace(w))
}
