package circle

import (
	"errors"
	"reflect"
)

var (
	// ErrEOI is returned by Iterator iterates or IteratorFunc calls
	// when the iteration ends.
	ErrEOI = errors.New("EOI")
	// ErrCannotCreateIterator is returned by NewIterator calls
	// when fails to create a new iterator.
	ErrCannotCreateIterator = errors.New("cannot create iterator")
)

type (
	// Iterator provides an interface for iterate on some iterables.
	Iterator struct {
		isEOI bool
		f     IteratorFunc
	}
	// IteratorFunc is an iterator as function.
	IteratorFunc func() (interface{}, error)
)

// NewIterator returns a new Iterator.
//
// If v is nil, returns an iterator that yields nothing.
// If v is a slice, an array or a receivable channel, returns an iterator that iterates on them.
// If v is an IteratorFunc, returns an iterator that yields a value from v calls.
// If v is an Iterator, returns v.
// Otherwise, returns an iterator that yields v.
func NewIterator(v interface{}) (*Iterator, error) {
	f, err := newIteratorFunc(v)
	if err != nil {
		return nil, err
	}
	return newIterator(f), nil
}

func newIterator(f IteratorFunc) *Iterator {
	return &Iterator{
		f: f,
	}
}

// Next yields the next element.
//
// This returns an error if the source of this iterator yields an error
// or the iteration ends.
//
// Once this returns some error, returns ErrEOI forever.
func (s *Iterator) Next() (interface{}, error) {
	if s.isEOI {
		return nil, ErrEOI
	}
	v, err := s.f()
	if err != nil {
		s.isEOI = true
		return nil, err
	}
	return v, nil
}

// Channel converts the iterator to IteratorChannel.
func (s *Iterator) Channel() IteratorChannel { return newIteratorChannel(s) }

type (
	// IteratorChannel is an iterator like a channel.
	IteratorChannel interface {
		// C returns the channel of the iterator.
		// The channel closes if the source yields some error.
		C() <-chan interface{}
		// Err returns the first non-EOI error that was encountered by the iteration.
		Err() error
	}
	iteratorChannel struct {
		iter *Iterator
		c    chan interface{}
		err  error
	}
)

func newIteratorChannel(iter *Iterator) IteratorChannel {
	s := &iteratorChannel{
		iter: iter,
		c:    make(chan interface{}),
	}
	go s.iterate()
	return s
}

func (s *iteratorChannel) iterate() {
	for {
		v, err := s.iter.Next()
		if err != nil {
			if err != ErrEOI {
				s.err = err
			}
			close(s.c)
			return
		}
		s.c <- v
	}
}
func (s *iteratorChannel) C() <-chan interface{} { return s.c }
func (s *iteratorChannel) Err() error            { return s.err }

/* IteratorFunc constructors */

func newIteratorFunc(v interface{}) (IteratorFunc, error) {
	if v == nil {
		return newNilIteratorFunc()
	}
	switch v := v.(type) {
	case func() (interface{}, error):
		return IteratorFunc(v), nil
	case IteratorFunc:
		return v, nil
	case *Iterator:
		return v.Next, nil
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.Array, reflect.Slice:
		return newArrayOrSliceIteratorFunc(v)
	case reflect.Chan:
		return newChanIteratorFunc(v)
	default:
		return newOrphanIteratorFunc(v)
	}
}

func newNilIteratorFunc() (IteratorFunc, error) {
	return func() (interface{}, error) { return nil, ErrEOI }, nil
}

func newOrphanIteratorFunc(v interface{}) (IteratorFunc, error) {
	var isEOI bool
	return func() (interface{}, error) {
		if isEOI {
			return nil, ErrEOI
		}
		isEOI = true
		return v, nil
	}, nil
}

func newArrayOrSliceIteratorFunc(v interface{}) (IteratorFunc, error) {
	t := reflect.TypeOf(v).Kind()
	if !(t == reflect.Array || t == reflect.Slice) {
		return nil, ErrCannotCreateIterator
	}
	var (
		xs = reflect.ValueOf(v)
		i  int
	)
	return func() (interface{}, error) {
		if i >= xs.Len() {
			return nil, ErrEOI
		}
		defer func() { i++ }()
		return xs.Index(i).Interface(), nil
	}, nil
}

func newChanIteratorFunc(v interface{}) (IteratorFunc, error) {
	t := reflect.TypeOf(v)
	if !(t.Kind() == reflect.Chan && t.ChanDir() != reflect.SendDir) {
		return nil, ErrCannotCreateIterator
	}
	c := reflect.ValueOf(v)
	return func() (interface{}, error) {
		x, ok := c.Recv()
		if ok {
			return x.Interface(), nil
		}
		return nil, ErrEOI
	}, nil
}
