package circle_test

import (
	"circle"
	"fmt"
	"testing"

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

type (
	testcaseStream struct {
		title   string
		src     interface{}
		stream  func(circle.Iterator) circle.Stream
		isError bool
		want    interface{}
	}
)

func (s *testcaseStream) test(t *testing.T) {
	it, err := circle.NewIterator(s.src)
	assert.Nil(t, err)
	stream := s.stream(it)
	git, gotErr := stream.Execute()
	assert.Equal(t, s.isError, gotErr != nil)
	if s.isError {
		return
	}
	got := []interface{}{}
	c := git.Channel()
	for v := range c.C() {
		got = append(got, v)
	}
	assert.Nil(t, c.Err())
	assert.Equal(t, "", cmp.Diff(got, s.want))
}

func TestStream(t *testing.T) {
	for _, tc := range []*testcaseStream{
		{
			title:  "just iterator",
			src:    []int{1, 2, 3},
			stream: circle.NewStream,
			want:   []interface{}{1, 2, 3},
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
			want: []interface{}{2, 3, 4},
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
			want: []interface{}{1, 3},
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
			want: []interface{}{6},
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
			want: []interface{}{1, 2, 3},
		},
		{
			title: "just flat",
			src:   [][]int{[]int{1}, []int{2, 3}, []int{4}},
			stream: func(it circle.Iterator) circle.Stream {
				return circle.NewStream(it).Flat()
			},
			want: []interface{}{1, 2, 3, 4},
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
			want: []interface{}{"A1234"},
		},
	} {
		t.Run(tc.title, tc.test)
	}
}
