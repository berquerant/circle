package circle

import "errors"

type (
	// Executor provides an interface for applying function to iterator.
	Executor interface {
		Execute() (*Iterator, error)
	}

	// ExecutorOption sets an option for Executor.
	ExecutorOption func(Executor)

	executorOption struct {
		aggregateExecutorOption
	}
)

type (
	mapExecutor struct {
		f  Mapper
		it *Iterator
	}
)

// NewMapExecutor returns a new Executor for map.
//
// If f returns error, the argument of f is ignored, this does not yield it.
func NewMapExecutor(f Mapper, it *Iterator) Executor {
	return &mapExecutor{
		f:  f,
		it: it,
	}
}

func (s *mapExecutor) Execute() (*Iterator, error) {
	var f func() (interface{}, error)
	f = func() (interface{}, error) {
		x, err := s.it.Next()
		if err != nil {
			return nil, err
		}
		v, err := s.f.Apply(x)
		if err != nil {
			// ignore this value
			return f()
		}
		return v, nil
	}
	return NewIterator(f)
}

type (
	filterExecutor struct {
		f  Filter
		it *Iterator
	}
)

// NewFilterExecutor returns a new Executor for filter.
//
// If f returns error, the iterator ends here.
func NewFilterExecutor(f Filter, it *Iterator) Executor {
	return &filterExecutor{
		f:  f,
		it: it,
	}
}

func (s *filterExecutor) Execute() (*Iterator, error) {
	var f func() (interface{}, error)
	f = func() (interface{}, error) {
		x, err := s.it.Next()
		if err != nil {
			return nil, err
		}
		v, err := s.f.Apply(x)
		if err != nil {
			// ends iterator
			return nil, err
		}
		if !v {
			// skip
			return f()
		}
		return x, nil
	}
	return NewIterator(f)
}

var (
	ErrInvalidAggregateExecutor = errors.New("invalid aggregate executor")
)

type (
	aggregateExecutor struct {
		f   Aggregator
		it  *Iterator
		iv  interface{}
		opt *executorOption
	}

	aggregateExecutorOption struct {
		aggregateExecutorType AggregateExecutorType
	}

	// AgregateExecutorType is a type of aggregation.
	AggregateExecutorType int
)

const (
	UnknownAggregateExecutorType AggregateExecutorType = iota
	// RAggregateExecutorType is foldr.
	RAggregateExecutorType
	// LAggregateExecutorType is foldl.
	LAggregateExecutorType
)

// NewAggregateExecutor returns a new Executor for aggregate.
//
// If f is not appropriate for aggregate, returns ErrInvalidAggregateExecutor.
func NewAggregateExecutor(f Aggregator, it *Iterator, iv interface{}, opt ...ExecutorOption) (Executor, error) {
	ex := &aggregateExecutor{
		f:   f,
		it:  it,
		iv:  iv,
		opt: &executorOption{},
	}
	for _, o := range opt {
		o(ex)
	}

	if ex.opt.aggregateExecutorType != UnknownAggregateExecutorType &&
		!isValidAggregateExecutorType(ex.opt.aggregateExecutorType, ex.f.Type()) {
		return nil, ErrInvalidAggregateExecutor
	}
	return ex, nil
}

// WithAggregateExecutorType sets AggregateExecutorType of Executor for aggregate.
func WithAggregateExecutorType(t AggregateExecutorType) ExecutorOption {
	return func(ex Executor) {
		if ax, ok := ex.(*aggregateExecutor); ok {
			ax.opt.aggregateExecutorType = t
		}
	}
}

func isValidAggregateExecutorType(aet AggregateExecutorType, at AggregatorType) bool {
	switch aet {
	case RAggregateExecutorType:
		return at == RightAggregatorType || at == PerfectAggregatorType
	case LAggregateExecutorType:
		return at == LeftAggregatorType || at == PerfectAggregatorType
	default:
		return false
	}
}

func (s *aggregateExecutor) executorType() AggregateExecutorType {
	if s.opt.aggregateExecutorType != UnknownAggregateExecutorType {
		return s.opt.aggregateExecutorType
	}
	switch s.f.Type() {
	case RightAggregatorType:
		return RAggregateExecutorType
	case LeftAggregatorType, PerfectAggregatorType:
		return LAggregateExecutorType
	default:
		return UnknownAggregateExecutorType
	}
}

func (s *aggregateExecutor) Execute() (*Iterator, error) {
	switch s.executorType() {
	case RAggregateExecutorType:
		var isEOI bool
		return NewIterator(func() (interface{}, error) {
			if isEOI {
				return nil, ErrEOI
			}
			isEOI = true
			return s.foldr(s.iv)
		})
	case LAggregateExecutorType:
		var isEOI bool
		return NewIterator(func() (interface{}, error) {
			if isEOI {
				return nil, ErrEOI
			}
			isEOI = true
			return s.foldl(s.iv)
		})
	default:
		return nil, ErrInvalidAggregateExecutor
	}
}

// foldr requires a -> b -> b
func (s *aggregateExecutor) foldr(acc interface{}) (interface{}, error) {
	x, err := s.it.Next()
	if err == ErrEOI {
		return acc, nil
	}
	if err != nil {
		return nil, err
	}
	r, err := s.foldr(acc)
	if err != nil {
		return nil, err
	}
	return s.f.Apply(x, r)
}

// foldl requies b -> a -> b
func (s *aggregateExecutor) foldl(acc interface{}) (interface{}, error) {
	x, err := s.it.Next()
	if err == ErrEOI {
		return acc, nil
	}
	if err != nil {
		return nil, err
	}
	r, err := s.f.Apply(acc, x)
	if err != nil {
		return nil, err
	}
	return s.foldl(r)
}
