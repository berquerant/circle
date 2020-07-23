package circle_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/berquerant/circle"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

type (
	testcaseStreamBuilder struct {
		title   string
		src     interface{}
		builder func(circle.Iterator) circle.StreamBuilder
		isError bool
		want    []interface{}
	}
)

func (s *testcaseStreamBuilder) test(t *testing.T) {
	it, err := circle.NewIterator(s.src)
	assert.Nil(t, err)
	x, err := s.builder(it).Execute()
	assert.Equal(t, s.isError, err != nil)
	if s.isError {
		return
	}
	got := []interface{}{}
	c := x.Channel()
	for v := range c.C() {
		got = append(got, v)
	}
	assert.Equal(t, "", cmp.Diff(s.want, got))
	assert.Nil(t, c.Err())
}

func TestStreamBuilder(t *testing.T) {
	for _, tc := range []*testcaseStreamBuilder{
		{
			title:   "just iterator",
			src:     []int{1, 2, 3},
			builder: circle.NewStreamBuilder,
			want:    []interface{}{1, 2, 3},
		},
		{
			title: "invalid map",
			src:   []int{1, 2, 3},
			builder: func(it circle.Iterator) circle.StreamBuilder {
				return circle.NewStreamBuilder(it).
					Map(func() {})
			},
			isError: true,
		},
		{
			title: "just map",
			src:   []int{1, 2, 3},
			builder: func(it circle.Iterator) circle.StreamBuilder {
				return circle.NewStreamBuilder(it).
					Map(func(x int) (int, error) {
						return x + 1, nil
					})
			},
			want: []interface{}{2, 3, 4},
		},
		{
			title: "maybe map filter",
			src:   []interface{}{circle.NewJust(1), circle.NewNothing(), circle.NewJust(3), circle.NewJust(-1)},
			builder: func(it circle.Iterator) circle.StreamBuilder {
				return circle.NewStreamBuilder(it).
					MaybeMap(func(x int) (int, error) {
						if x < 0 {
							return 0, errors.New("negative")
						}
						return x + 10, nil
					}).
					Filter(func(x circle.Maybe) (bool, error) {
						return !x.IsNothing(), nil
					}).
					Map(func(x circle.Maybe) (interface{}, error) {
						v, _ := x.Get()
						return v, nil
					})
			},
			want: []interface{}{11, 13},
		},
		{
			title: "tuple filter tuple map",
			src:   []interface{}{circle.NewTuple(1, 2), circle.NewTuple(3, 4), circle.NewTuple(5, 6)},
			builder: func(it circle.Iterator) circle.StreamBuilder {
				return circle.NewStreamBuilder(it).
					TupleFilter(func(x, y int) (bool, error) {
						return x+y > 5, nil
					}).
					TupleMap(func(x, y int) (circle.Tuple, error) {
						return circle.NewTuple(x+y, x*y, 2*x-y), nil
					}).
					TupleMap(func(x, y, z int) (int, error) {
						return y - x + z, nil
					})
			},
			want: []interface{}{7, 23},
		},
		{
			title: "either map",
			src: []interface{}{
				circle.NewRight(1),
				circle.NewLeft(errors.New("left")),
				circle.NewRight(2),
				circle.NewRight(-1),
			},
			builder: func(it circle.Iterator) circle.StreamBuilder {
				return circle.NewStreamBuilder(it).
					EitherMap(func(x int) (int, error) {
						if x < 0 {
							return 0, errors.New("negative")
						}
						return x + 1, nil
					}).
					Map(func(x circle.Either) (string, error) {
						if v, ok := x.Left(); ok {
							return fmt.Sprintf("left(%v)", v), nil
						}
						if v, ok := x.Right(); ok {
							return fmt.Sprintf("right(%v)", v), nil
						}
						return "", errors.New("unreachable")
					})
			},
			want: []interface{}{"right(2)", "left(left)", "right(3)", "left(negative)"},
		},
		{
			title: "flat sort aggregate",
			src:   []interface{}{[]int{5}, []int{3, 4}, []int{1, 2}},
			builder: func(it circle.Iterator) circle.StreamBuilder {
				return circle.NewStreamBuilder(it).
					Flat().
					Sort(func(x, y int) (bool, error) {
						return x < y, nil
					}).
					Aggregate(func(x string, y int) (string, error) {
						return fmt.Sprintf("(%s+%d)", x, y), nil
					}, "iv")
			},
			want: []interface{}{"(((((iv+1)+2)+3)+4)+5)"},
		},
	} {
		t.Run(tc.title, tc.test)
	}
}
