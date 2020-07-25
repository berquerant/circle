/*
Package circle provides a stream API.
*/
package circle

import "fmt"

type (
	// StreamBuilder provides a convenient interface for streaming.
	StreamBuilder interface {
		// Map maps stream.
		// Convert each element by f, func(A) (B, error).
		// if f returns error, the element is filtered from this stream.
		Map(f interface{}, opt ...StreamOption) StreamBuilder
		// MaybeMap maps stream with Maybe.
		// If an element is Just (has value), converts the value of it by f, func(A) (B, error),
		// If f returns error, yield Nothing (has no value).
		// If an element is not Maybe, it is filtered from this stream.
		MaybeMap(f interface{}, opt ...StreamOption) StreamBuilder
		// EitherMap maps stream with Either.
		// If an element is Right, converts the value of it by f, func(A) (B, error).
		// If f returns error, yield Left with the error.
		// If an element is not Either, it is filtered from this stream.
		EitherMap(f interface{}, opt ...StreamOption) StreamBuilder
		// TupleMap maps stream with Tuple.
		// Converts each element by f, func(A1, A2, ..., An) (B, error).
		// If an element is not Tuple or size of Tuple is not equal to n or type of each element do not match to A1, A2, ...., An,
		// it is filtered from this stream.
		TupleMap(f interface{}, opt ...StreamOption) StreamBuilder
		// Filter filters stream.
		// Select elements by f, func(A) (bool, error).
		// If f returns false, the element is filtered from this stream.
		// If f returns error, stops streaming.
		Filter(f interface{}, opt ...StreamOption) StreamBuilder
		// TupleFilter filters stream with Tuple.
		// Select elements by f, func(A1, A2, ..., An) (bool, error).
		// If f returns false, the element is filtered from this stream.
		// If f returns error,
		// or an element is not Tuple or size of Tuple is not equal to n or type of each element do not match to A1, A2, ...., An,
		// stops streaming.
		TupleFilter(f interface{}, opt ...StreamOption) StreamBuilder
		// Aggregate aggregates stream.
		// Aggregate elements by f, func(A, B) (A, error) or func(A, B) (B, error) with initial value iv.
		Aggregate(f, iv interface{}, opt ...StreamOption) StreamBuilder
		// Sort sorts stream.
		// Sort elements by f, func(A, A) (bool, error).
		// If f returns error, the element is regarded as bigger.
		Sort(f interface{}, opt ...StreamOption) StreamBuilder
		// Flat flattens stream.
		// See NewFlatExecutor().
		Flat(opt ...StreamOption) StreamBuilder
		// Consume consumes stream.
		// If f returns error, stops consuming.
		Consume(f interface{}, opt ...StreamOption) error
		// MaybeConsume consumes stream with Maybe.
		// If an element is Just, consumes the value of it by f, func(A) error,
		// else calls g.
		MaybeConsume(f interface{}, g func() error, opt ...StreamOption) error
		// EitherConsume consumes stream with Either.
		// If an element is Right, consumers the value of it by g, func(A) error,
		// else by f, func(B) error.
		EitherConsume(f, g interface{}, opt ...StreamOption) error
		// TupleConsume consumes stream with Tuple.
		// If an element is not Tuple or size of Tuple is not equal to n or type of each element do not match to A1, A2, ...., An
		// or f returns error, stops consuming.
		TupleConsume(f interface{}, opt ...StreamOption) error
		Executor
	}

	StreamFactory func(Stream) (Stream, error)

	streamBuilder struct {
		stream Stream
		nodes  []StreamFactory
	}
)

// NewStreamBuilder returns a new StreamBuilder.
func NewStreamBuilder(it Iterator) StreamBuilder {
	return &streamBuilder{
		stream: NewStream(it),
		nodes:  []StreamFactory{},
	}
}

func (s *streamBuilder) add(f StreamFactory) StreamBuilder {
	s.nodes = append(s.nodes, f)
	return s
}

func (s *streamBuilder) Map(f interface{}, opt ...StreamOption) StreamBuilder {
	x, err := NewMapper(f)
	return s.add(func(a Stream) (Stream, error) {
		if err != nil {
			return nil, err
		}
		return a.Map(x, opt...), nil
	})
}
func (s *streamBuilder) Filter(f interface{}, opt ...StreamOption) StreamBuilder {
	x, err := NewFilter(f)
	return s.add(func(a Stream) (Stream, error) {
		if err != nil {
			return nil, err
		}
		return a.Filter(x, opt...), nil
	})
}
func (s *streamBuilder) Aggregate(f, iv interface{}, opt ...StreamOption) StreamBuilder {
	x, err := NewAggregator(f)
	return s.add(func(a Stream) (Stream, error) {
		if err != nil {
			return nil, err
		}
		return a.Aggregate(x, iv, opt...), nil
	})
}
func (s *streamBuilder) Sort(f interface{}, opt ...StreamOption) StreamBuilder {
	x, err := NewComparator(f)
	return s.add(func(a Stream) (Stream, error) {
		if err != nil {
			return nil, err
		}
		return a.Sort(x, opt...), nil
	})
}
func (s *streamBuilder) Flat(opt ...StreamOption) StreamBuilder {
	return s.add(func(a Stream) (Stream, error) {
		return a.Flat(opt...), nil
	})
}
func (s *streamBuilder) MaybeMap(f interface{}, opt ...StreamOption) StreamBuilder {
	x, err := NewMaybeMapper(f)
	return s.add(func(a Stream) (Stream, error) {
		if err != nil {
			return nil, err
		}
		return a.Map(x, opt...), nil
	})
}
func (s *streamBuilder) EitherMap(f interface{}, opt ...StreamOption) StreamBuilder {
	x, err := NewEitherMapper(f)
	return s.add(func(a Stream) (Stream, error) {
		if err != nil {
			return nil, err
		}
		return a.Map(x, opt...), nil
	})
}
func (s *streamBuilder) TupleMap(f interface{}, opt ...StreamOption) StreamBuilder {
	x, err := NewTupleMapper(f)
	return s.add(func(a Stream) (Stream, error) {
		if err != nil {
			return nil, err
		}
		return a.Map(x, opt...), nil
	})
}
func (s *streamBuilder) TupleFilter(f interface{}, opt ...StreamOption) StreamBuilder {
	x, err := NewTupleFilter(f)
	return s.add(func(a Stream) (Stream, error) {
		if err != nil {
			return nil, err
		}
		return a.Filter(x, opt...), nil
	})
}
func (s *streamBuilder) connect() (Stream, error) {
	var st Stream = s.stream
	for _, f := range s.nodes {
		n, err := f(st)
		if err != nil {
			return nil, fmt.Errorf("%w %v", ErrCannotCreateStream, err)
		}
		st = n
	}
	return st, nil
}
func (s *streamBuilder) Execute() (Iterator, error) {
	st, err := s.connect()
	if err != nil {
		return nil, err
	}
	return st.Execute()
}
func (s *streamBuilder) consume(f func() (Consumer, error), opt ...StreamOption) error {
	x, err := f()
	if err != nil {
		return fmt.Errorf("%w %v", ErrCannotCreateStream, err)
	}
	st, err := s.connect()
	if err != nil {
		return err
	}
	return st.Consume(x, opt...)
}
func (s *streamBuilder) Consume(f interface{}, opt ...StreamOption) error {
	return s.consume(func() (Consumer, error) { return NewConsumer(f) }, opt...)
}
func (s *streamBuilder) MaybeConsume(f interface{}, g func() error, opt ...StreamOption) error {
	return s.consume(func() (Consumer, error) { return NewMaybeConsumer(f, g) }, opt...)
}
func (s *streamBuilder) EitherConsume(f, g interface{}, opt ...StreamOption) error {
	return s.consume(func() (Consumer, error) { return NewEitherConsumer(f, g) }, opt...)
}
func (s *streamBuilder) TupleConsume(f interface{}, opt ...StreamOption) error {
	return s.consume(func() (Consumer, error) { return NewTupleConsumer(f) }, opt...)
}
