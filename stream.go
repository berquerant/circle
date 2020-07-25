package circle

import (
	"errors"
	"fmt"
)

type (
	// Stream provides an interface for streaming.
	Stream interface {
		// Map maps Stream.
		// Convert each element by f.
		// If f returns error, the element is filtered from this stream.
		Map(f Mapper, opt ...StreamOption) Stream
		// Filter filters Stream.
		// Select elements by f.
		// If f returns error, stops streaming.
		Filter(f Filter, opt ...StreamOption) Stream
		// Aggregate aggregates Stream.
		// Aggregate elements by f and iv as initial value.
		Aggregate(f Aggregator, iv interface{}, opt ...StreamOption) Stream
		// Sort sorts Stream.
		// Sort elements by f.
		// If f returns error, the element is regarded as bigger.
		Sort(f Comparator, opt ...StreamOption) Stream
		// Flat flattens Stream.
		// See NewFlatExecutor().
		Flat(opt ...StreamOption) Stream
		// Consume consumes Stream.
		// If f returns error, stops consuming.
		Consume(f Consumer, opt ...StreamOption) error
		Executor
	}

	StreamNodeFactory func(Iterator) StreamNode

	stream struct {
		it    Iterator
		nodes []StreamNodeFactory
	}
)

var (
	ErrCannotCreateStream = errors.New("cannot create stream")
)

// NewStream returns a new Stream.
func NewStream(it Iterator) Stream {
	return &stream{
		it:    it,
		nodes: []StreamNodeFactory{},
	}
}

func (s *stream) Execute() (Iterator, error) { return s.connect() }

func (s *stream) connect() (Iterator, error) {
	var it Iterator = s.it
	for _, f := range s.nodes {
		n := f(it)
		if err := n.Err(); err != nil {
			return nil, fmt.Errorf("%w %s %v", ErrCannotCreateStream, n.ID(), err)
		}
		nit, err := n.Execute()
		if err != nil {
			return nil, fmt.Errorf("%w %s %v", ErrCannotCreateStream, n.ID(), err)
		}
		it = nit
	}
	return it, nil
}

func (s *stream) add(f StreamNodeFactory) Stream {
	s.nodes = append(s.nodes, f)
	return s
}
func (s *stream) Map(f Mapper, opt ...StreamOption) Stream {
	c := newStreamConfig(opt...)
	return s.add(func(it Iterator) StreamNode {
		return NewStreamNode(NewMapExecutor(f, it), c.NodeID)
	})
}
func (s *stream) Filter(f Filter, opt ...StreamOption) Stream {
	c := newStreamConfig(opt...)
	return s.add(func(it Iterator) StreamNode {
		return NewStreamNode(NewFilterExecutor(f, it), c.NodeID)
	})
}
func (s *stream) Aggregate(f Aggregator, iv interface{}, opt ...StreamOption) Stream {
	c := newStreamConfig(opt...)
	aopts := []ExecutorOption{}
	if c.Aggregate.Type != UnknownAggregateExecutorType {
		aopts = append(aopts, WithAggregateExecutorType(c.Aggregate.Type))
	}
	return s.add(func(it Iterator) StreamNode {
		f, err := NewAggregateExecutor(f, it, iv, aopts...)
		if err != nil {
			return NewErrStreamNode(err, c.NodeID)
		}
		return NewStreamNode(f, c.NodeID)
	})
}
func (s *stream) Sort(f Comparator, opt ...StreamOption) Stream {
	c := newStreamConfig(opt...)
	return s.add(func(it Iterator) StreamNode {
		return NewStreamNode(NewCompareExecutor(f, it), c.NodeID)
	})
}
func (s *stream) Flat(opt ...StreamOption) Stream {
	c := newStreamConfig(opt...)
	return s.add(func(it Iterator) StreamNode {
		return NewStreamNode(NewFlatExecutor(it), c.NodeID)
	})
}

func (s *stream) Consume(f Consumer, opt ...StreamOption) error {
	it, err := s.connect()
	if err != nil {
		return err
	}
	return NewConsumeExecutor(f, it).ConsumeExecute()
}

type (
	// StreamOption is an option of Stream.
	StreamOption func(*StreamConfig)

	StreamConfig struct {
		NodeID    string
		Aggregate StreamConfigAggregate
	}
	// StreamConfigAggregate is a config for Aggregate.
	StreamConfigAggregate struct {
		Type AggregateExecutorType
	}

	// AggregateType is a type of aggregation.
	AggregateType int
)

const (
	UnknownAggregateType AggregateType = iota
	// RAggregateType is foldr.
	RAggregateType
	// LAggregateType is foldl.
	LAggregateType
)

func newStreamConfig(opt ...StreamOption) *StreamConfig {
	x := &StreamConfig{}
	x.Apply(opt...)
	return x
}

// Apply applies opt to this.
func (s *StreamConfig) Apply(opt ...StreamOption) {
	for _, o := range opt {
		o(s)
	}
}

// WithAggregateType returns a new StreamOption that sets a type of the aggregation.
// Stream.Aggregate selects an aggregate type automatically using the function signature,
// but you can also select the aggregate type.
func WithAggregateType(t AggregateType) StreamOption {
	return func(c *StreamConfig) {
		switch t {
		case RAggregateType:
			c.Aggregate.Type = RAggregateExecutorType
		case LAggregateType:
			c.Aggregate.Type = LAggregateExecutorType
		}
	}
}

// WithNodeID returns a new StreamOption that sets an id of the node.
// The node id is useful for debugging stream.
// The errors yielded from the iteration of the stream contains the node id.
func WithNodeID(nid string) StreamOption {
	return func(c *StreamConfig) {
		c.NodeID = nid
	}
}
