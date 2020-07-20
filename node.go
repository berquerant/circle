package circle

import (
	"circle/internal/util"
	"context"
	"fmt"
)

type (
	nodeID struct {
		s string
		u util.UUID
	}

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
		nid      *nodeID
	}
	errStreamNode struct {
		nid *nodeID
		err error
	}
)

func newNodeID(s string) *nodeID {
	return &nodeID{
		u: util.NewUUID(),
		s: s,
	}
}

func (s *nodeID) String() string {
	if s.s != "" {
		return s.s
	}
	return s.u.String()
}

// NewStreamNode returns a new StreamNode.
// If nid is empty, the id of this node is uuid.
func NewStreamNode(executor Executor, nid string) StreamNode {
	return &streamNode{
		executor: executor,
		nid:      newNodeID(nid),
	}
}

// NewErrStreamNode returns a new failed StreamNode.
// If nid is empty, the id of this node is uuid.
func NewErrStreamNode(err error, nid string) StreamNode {
	return &errStreamNode{
		nid: newNodeID(nid),
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
func (s *streamNode) ID() string { return s.nid.String() }
func (*streamNode) Err() error   { return nil }

func (*errStreamNode) Execute() (Iterator, error) { return nil, ErrCannotCreateIterator }
func (s *errStreamNode) ID() string               { return s.nid.String() }
func (s *errStreamNode) Err() error               { return s.err }

type (
	// StreamNodeIterator is an Iterator that appends node id to iterator errors.
	StreamNodeIterator struct {
		it  Iterator
		nid *nodeID
	}
)

func (s *StreamNodeIterator) ID() string { return s.nid.String() }
func (s *StreamNodeIterator) Next() (interface{}, error) {
	r, err := s.it.Next()
	if err == ErrEOI {
		return nil, ErrEOI
	}
	if err != nil {
		return nil, fmt.Errorf("%s %w", s.nid.String(), err)
	}
	return r, nil
}
func (s *StreamNodeIterator) Channel() IteratorChannel { return s.it.Channel() }
func (s *StreamNodeIterator) ChannelWithContext(ctx context.Context) IteratorChannel {
	return s.it.ChannelWithContext(ctx)
}
