package circle_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/berquerant/circle"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func mustNewMapper(t *testing.T, f interface{}) circle.Mapper {
	x, err := circle.NewMapper(f)
	assert.Nil(t, err, "must new mapper")
	return x
}

func mustNewFilter(t *testing.T, f interface{}) circle.Filter {
	x, err := circle.NewFilter(f)
	assert.Nil(t, err, "must new filter")
	return x
}

func mustNewAggregator(t *testing.T, f interface{}) circle.Aggregator {
	x, err := circle.NewAggregator(f)
	assert.Nil(t, err, "must new aggregator")
	return x
}

func mustNewComparator(t *testing.T, f interface{}) circle.Comparator {
	x, err := circle.NewComparator(f)
	assert.Nil(t, err, "must new comparator")
	return x
}

func mustNewConsumer(t *testing.T, f interface{}) circle.Consumer {
	x, err := circle.NewConsumer(f)
	assert.Nil(t, err, "must new consumer")
	return x
}

func TestStreamConsume(t *testing.T) {
	it, err := circle.NewIterator([]int{1, 2, 3})
	assert.Nil(t, err)
	ch := make(chan interface{})
	st := circle.NewStream(it)
	go func() {
		assert.Nil(t, st.Consume(mustNewConsumer(t, func(x int) error {
			ch <- x
			return nil
		})))
		close(ch)
	}()
	got := []interface{}{}
	for x := range ch {
		got = append(got, x)
	}
	assert.Equal(t, "", cmp.Diff([]interface{}{1, 2, 3}, got))
}

type (
	testcaseStream struct {
		title        string
		src          interface{}
		stream       func(circle.Iterator) circle.Stream
		wantNewErr   error
		wantYieldErr error
		wantVal      interface{}
	}
)

func (s *testcaseStream) test(t *testing.T) {
	it, err := circle.NewIterator(s.src)
	assert.Nil(t, err)
	stream := s.stream(it)
	git, gotErr := stream.Execute()
	assert.Equal(t, fmt.Sprint(s.wantNewErr), fmt.Sprint(gotErr))
	if gotErr != nil {
		return
	}
	got := []interface{}{}
	c := git.Channel()
	for v := range c.C() {
		got = append(got, v)
	}
	assert.Equal(t, fmt.Sprint(s.wantYieldErr), fmt.Sprint(c.Err()))
	assert.Equal(t, "", cmp.Diff(got, s.wantVal))
}

func TestStream(t *testing.T) {
	for _, tc := range []*testcaseStream{
		{
			title: "yield error default 0 of 0",
			src:   []int{1, 2, 3},
			stream: func(it circle.Iterator) circle.Stream {
				return circle.NewStream(it).
					Filter(mustNewFilter(t, func(int) (bool, error) {
						return false, errors.New("ERROR")
					}))
			},
			wantYieldErr: errors.New("0 ERROR"),
			wantVal:      []interface{}{},
		},
		{
			title: "yield error specified node id",
			src:   []int{1, 2, 3},
			stream: func(it circle.Iterator) circle.Stream {
				return circle.NewStream(it).
					Filter(mustNewFilter(t, func(int) (bool, error) {
						return false, errors.New("ERROR")
					}), circle.WithNodeID("NID"))
			},
			wantYieldErr: errors.New("NID ERROR"),
			wantVal:      []interface{}{},
		},
		{
			title: "yield error default 1 of 2",
			src:   []int{1, 2, 3},
			stream: func(it circle.Iterator) circle.Stream {
				return circle.NewStream(it).
					Map(mustNewMapper(t, func(int) (int, error) {
						return 0, nil
					})).
					Filter(mustNewFilter(t, func(int) (bool, error) {
						return false, errors.New("ERROR")
					})).
					Map(mustNewMapper(t, func(int) (int, error) {
						return 0, nil
					}))
			},
			wantYieldErr: errors.New("2 1 ERROR"),
			wantVal:      []interface{}{},
		},
		{
			title: "yield error specified 1 of 2",
			src:   []int{1, 2, 3},
			stream: func(it circle.Iterator) circle.Stream {
				return circle.NewStream(it).
					Map(mustNewMapper(t, func(int) (int, error) {
						return 0, nil
					}), circle.WithNodeID("N1")).
					Filter(mustNewFilter(t, func(int) (bool, error) {
						return false, errors.New("ERROR")
					}), circle.WithNodeID("N2")).
					Map(mustNewMapper(t, func(int) (int, error) {
						return 0, nil
					}), circle.WithNodeID("N3"))
			},
			wantYieldErr: errors.New("N3 N2 ERROR"),
			wantVal:      []interface{}{},
		},
		{
			title: "yield error specified and default 1 of 2",
			src:   []int{1, 2, 3},
			stream: func(it circle.Iterator) circle.Stream {
				return circle.NewStream(it).
					Map(mustNewMapper(t, func(int) (int, error) {
						return 0, nil
					}), circle.WithNodeID("N1")).
					Filter(mustNewFilter(t, func(int) (bool, error) {
						return false, errors.New("ERROR")
					})).
					Map(mustNewMapper(t, func(int) (int, error) {
						return 0, nil
					}), circle.WithNodeID("N3"))
			},
			wantYieldErr: errors.New("N3 1 ERROR"),
			wantVal:      []interface{}{},
		},
		{
			title:   "just iterator",
			src:     []int{1, 2, 3},
			stream:  circle.NewStream,
			wantVal: []interface{}{1, 2, 3},
		},
		{
			title: "just map",
			src:   []int{1, 2, 3},
			stream: func(it circle.Iterator) circle.Stream {
				return circle.NewStream(it).
					Map(mustNewMapper(t, func(x int) (int, error) {
						return x + 1, nil
					}))
			},
			wantVal: []interface{}{2, 3, 4},
		},
		{
			title: "just filter",
			src:   []int{1, 2, 3},
			stream: func(it circle.Iterator) circle.Stream {
				return circle.NewStream(it).
					Filter(mustNewFilter(t, func(x int) (bool, error) {
						return x&1 == 1, nil
					}))
			},
			wantVal: []interface{}{1, 3},
		},
		{
			title: "just aggregate",
			src:   []int{1, 2, 3},
			stream: func(it circle.Iterator) circle.Stream {
				return circle.NewStream(it).
					Aggregate(mustNewAggregator(t, func(x, y int) (int, error) {
						return x + y, nil
					}), 0)
			},
			wantVal: []interface{}{6},
		},
		{
			title: "just sort",
			src:   []int{3, 1, 2},
			stream: func(it circle.Iterator) circle.Stream {
				return circle.NewStream(it).
					Sort(mustNewComparator(t, func(x, y int) (bool, error) {
						return x < y, nil
					}))
			},
			wantVal: []interface{}{1, 2, 3},
		},
		{
			title: "just flat",
			src:   [][]int{{1}, {2, 3}, {4}},
			stream: func(it circle.Iterator) circle.Stream {
				return circle.NewStream(it).Flat()
			},
			wantVal: []interface{}{1, 2, 3, 4},
		},
		{
			title: "filter map aggregate",
			src:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			stream: func(it circle.Iterator) circle.Stream {
				return circle.NewStream(it).
					Filter(mustNewFilter(t, func(x int) (bool, error) {
						return x < 5, nil
					})).
					Map(mustNewMapper(t, func(x int) (string, error) {
						return fmt.Sprint(x), nil
					})).
					Aggregate(mustNewAggregator(t, func(x, y string) (string, error) {
						return fmt.Sprintf("%s%s", x, y), nil
					}), "A", circle.WithAggregateType(circle.LAggregateType))
			},
			wantVal: []interface{}{"A1234"},
		},
	} {
		t.Run(tc.title, tc.test)
	}
}
