package circle

import (
	"context"
	"fmt"
)

type (
	// StreamNode is an intermediate part of Stream.
	StreamNode interface {
		Executor
		// ID returns the id of this.
		ID() string
		// Err returns error if this is a failed node.
		Err() error
	}

	streamNode struct {
		executor Executor
		nid      string
	}
	errStreamNode struct {
		nid string
		err error
	}
)

// NewStreamNode returns a new StreamNode.
func NewStreamNode(executor Executor, nid string) StreamNode {
	return &streamNode{
		executor: executor,
		nid:      nid,
	}
}

// NewErrStreamNode returns a new failed StreamNode.
func NewErrStreamNode(err error, nid string) StreamNode {
	return &errStreamNode{
		nid: nid,
		err: err,
	}
}

func (s *streamNode) Execute() (Iterator, error) {
	it, err := s.executor.Execute()
	if err != nil {
		return nil, err
	}
	return &StreamNodeIterator{
		it:  it,
		nid: s.nid,
	}, nil
}
func (s *streamNode) ID() string { return s.nid }
func (*streamNode) Err() error   { return nil }

func (*errStreamNode) Execute() (Iterator, error) { return nil, ErrCannotCreateIterator }
func (s *errStreamNode) ID() string               { return s.nid }
func (s *errStreamNode) Err() error               { return s.err }

type (
	// StreamNodeIterator is an Iterator that appends node id to iterator errors.
	StreamNodeIterator struct {
		it  Iterator
		nid string
	}
)

func (s *StreamNodeIterator) ID() string { return s.nid }
func (s *StreamNodeIterator) Next() (interface{}, error) {
	r, err := s.it.Next()
	if err == ErrEOI {
		return nil, ErrEOI
	}
	if err != nil {
		return nil, fmt.Errorf("%s %w", s.nid, err)
	}
	return r, nil
}
func (s *StreamNodeIterator) channel(ctx context.Context) IteratorChannel {
	it, _ := NewIterator(s.Next)
	return it.ChannelWithContext(ctx)
}
func (s *StreamNodeIterator) Channel() IteratorChannel { return s.channel(context.Background()) }
func (s *StreamNodeIterator) ChannelWithContext(ctx context.Context) IteratorChannel {
	return s.channel(ctx)
}
