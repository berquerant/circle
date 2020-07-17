package atomic

import "sync"

type (
	Bool struct {
		v   bool
		mux *sync.RWMutex
	}
)

func NewBool(v bool) *Bool {
	return &Bool{
		v:   v,
		mux: &sync.RWMutex{},
	}
}

func (s *Bool) Get() bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.v
}

func (s *Bool) Set(v bool) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.v = v
	return v
}
