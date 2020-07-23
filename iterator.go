package circle

import (
	"context"
	"errors"
	"reflect"

	"github.com/berquerant/circle/internal/atomic"
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
	Iterator interface {
		// Next yields the next element.
		//
		// This returns an error if the source of this iterator yields an error
		// or the iteration ends.
		//
		// Once this returns some error, returns ErrEOI forever.
		Next() (interface{}, error)
		// Channel converts the iterator to IteratorChannel.
		Channel() IteratorChannel
		// ChannelWithContext converts the iterator to IteratorChannel.
		// If context canceled, the channel closes.
		ChannelWithContext(ctx context.Context) IteratorChannel
	}
	iterator struct {
		isEOI bool
		f     IteratorFunc
	}
	// IteratorFunc is an iterator as a function.
	IteratorFunc func() (interface{}, error)
)

// NewIterator returns a new Iterator.
//
// If v is nil, returns an iterator that yields nothing.
//
// If v is a slice, an array or a receivable channel, returns an iterator that iterates on them.
//
// If v is a map, returns an iterator that iterates on it, an element is Tuple, (Key, Value).
//
// If v is an IteratorFunc, returns an iterator that yields a value from v calls.
//
// If v is an Iterator, returns v.
//
// Otherwise, returns an iterator that yields v itself.
func NewIterator(v interface{}) (Iterator, error) {
	f, err := newIteratorFunc(v)
	if err != nil {
		return nil, err
	}
	return newIterator(f), nil
}

func newIterator(f IteratorFunc) Iterator {
	return &iterator{
		f: f,
	}
}

func (s *iterator) Next() (interface{}, error) {
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

func (s *iterator) Channel() IteratorChannel                               { return s.channel(context.Background()) }
func (s *iterator) ChannelWithContext(ctx context.Context) IteratorChannel { return s.channel(ctx) }
func (s *iterator) channel(ctx context.Context) IteratorChannel            { return newIteratorChannel(ctx, s) }

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
		iter     Iterator
		c        chan interface{}
		err      error
		isClosed *atomic.Bool
	}
)

func newIteratorChannel(ctx context.Context, iter Iterator) IteratorChannel {
	s := &iteratorChannel{
		iter:     iter,
		c:        make(chan interface{}),
		isClosed: atomic.NewBool(false),
	}
	go s.iterate(ctx)
	return s
}

func (s *iteratorChannel) iterate(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-ctx.Done()
		s.isClosed.Set(true)
		for range s.c {
		}
	}()

	defer func() {
		cancel()
		close(s.c)
	}()

	for {
		if s.isClosed.Get() {
			return
		}
		v, err := s.iter.Next()
		if err != nil {
			if err != ErrEOI {
				s.err = err
			}
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
	case Iterator:
		return v.Next, nil
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.Array, reflect.Slice:
		return newArrayOrSliceIteratorFunc(v)
	case reflect.Chan:
		return newChanIteratorFunc(v)
	case reflect.Map:
		return newMapIteratorFunc(v)
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

func newMapIteratorFunc(v interface{}) (IteratorFunc, error) {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Map {
		return nil, ErrCannotCreateIterator
	}
	iter := reflect.ValueOf(v).MapRange()
	return func() (interface{}, error) {
		if iter.Next() {
			return NewTuple(iter.Key().Interface(), iter.Value().Interface()), nil
		}
		return nil, ErrEOI
	}, nil
}
