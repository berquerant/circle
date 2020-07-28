package circle_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/berquerant/circle"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func ExampleStreamBuilder_map() {
	it, _ := circle.NewIterator([]int{1, 2, 3, -1, 4})
	err := circle.NewStreamBuilder(it).
		Map(func(x int) (int, error) {
			if x < 0 {
				return 0, errors.New("negative")
			}
			return x + 10, nil
		}).
		Consume(func(x int) error {
			fmt.Println(x)
			return nil
		})
	fmt.Println(err)
	// Output:
	// 11
	// 12
	// 13
	// 14
	// <nil>
}

func ExampleStreamBuilder_maybeMap() {
	it, _ := circle.NewIterator([]circle.Maybe{
		circle.NewJust(1),
		circle.NewNothing(),
		circle.NewJust(3),
		circle.NewJust(-1),
	})
	err := circle.NewStreamBuilder(it).
		MaybeMap(func(x int) (int, error) {
			if x < 0 {
				return 0, errors.New("negative")
			}
			return x + 10, nil
		}).
		MaybeConsume(func(x int) error {
			fmt.Println(x)
			return nil
		}, func() error {
			fmt.Println("nothing")
			return nil
		})
	fmt.Println(err)
	// Output:
	// 11
	// nothing
	// 13
	// nothing
	// <nil>
}

func ExampleStreamBuilder_eitherMap() {
	it, _ := circle.NewIterator([]circle.Either{
		circle.NewRight(1),
		circle.NewLeft(errors.New("e1")),
		circle.NewRight(3),
		circle.NewRight(-1),
	})
	err := circle.NewStreamBuilder(it).
		EitherMap(func(x int) (int, error) {
			if x < 0 {
				return 0, fmt.Errorf("negative: %d", x)
			}
			return x + 10, nil
		}).
		EitherConsume(func(err error) error {
			fmt.Println(err)
			return nil
		}, func(x int) error {
			fmt.Println(x)
			return nil
		})
	fmt.Println(err)
	// Output:
	// 11
	// e1
	// 13
	// negative: -1
	// <nil>
}

func ExampleStreamBuilder_tupleMap() {
	it, _ := circle.NewIterator([]int{1, 2, 3, 4, -1})
	err := circle.NewStreamBuilder(it).
		Map(func(x int) (circle.Tuple, error) {
			return circle.NewTuple(x, x*x), nil
		}).
		TupleFilter(func(x, xx int) (bool, error) {
			return x > 0, nil
		}).
		TupleMap(func(x, xx int) (circle.Tuple, error) {
			return circle.NewTuple(x, xx, x*xx), nil
		}).
		TupleConsume(func(x, xx, xxx int) error {
			fmt.Printf("%d %d %d\n", x, xx, xxx)
			return nil
		})
	fmt.Println(err)
	// Output:
	// 1 1 1
	// 2 4 8
	// 3 9 27
	// 4 16 64
	// <nil>
}

func ExampleStreamBuilder_filter() {
	it, _ := circle.NewIterator([]int{1, 2, 3, -1, 4, 5, 6})
	err := circle.NewStreamBuilder(it).
		Filter(func(x int) (bool, error) {
			if x < 0 {
				return false, fmt.Errorf("negative: %d", x)
			}
			return x&1 == 1, nil
		}, circle.WithNodeID("f1")).
		Consume(func(x int) error {
			fmt.Println(x)
			return nil
		})
	fmt.Println(err)
	// Output:
	// 1
	// 3
	// f1 negative: -1
}

func ExampleStreamBuilder_aggregate() {
	src := func() circle.Iterator {
		it, _ := circle.NewIterator([]int{1, 2, 3})
		return it
	}
	_ = circle.NewStreamBuilder(src()).
		Aggregate(func(acc string, x int) (string, error) {
			return fmt.Sprintf("(%s+%d)", acc, x), nil
		}, "iv").
		Consume(func(x string) error {
			fmt.Printf("left: %s\n", x)
			return nil
		})
	_ = circle.NewStreamBuilder(src()).
		Aggregate(func(x int, acc string) (string, error) {
			return fmt.Sprintf("(%d+%s)", x, acc), nil
		}, "iv").
		Consume(func(x string) error {
			fmt.Printf("right: %s\n", x)
			return nil
		})
	// Output:
	// left: (((iv+1)+2)+3)
	// right: (1+(2+(3+iv)))
}

func ExampleStreamBuilder_sort() {
	it, _ := circle.NewIterator([]int{4, 1, 3, 2})
	err := circle.NewStreamBuilder(it).
		Sort(func(x, y int) (bool, error) {
			return x < y, nil
		}).
		Consume(func(x int) error {
			fmt.Println(x)
			return nil
		})
	fmt.Println(err)
	// Output:
	// 1
	// 2
	// 3
	// 4
	// <nil>
}

func ExampleStreamBuilder_flat() {
	it, _ := circle.NewIterator([][]int{[]int{1}, []int{2, 3}})
	_ = circle.NewStreamBuilder(it).
		Flat().
		Consume(func(x int) error {
			fmt.Println(x)
			return nil
		})
	// Output:
	// 1
	// 2
	// 3
}

func ExampleStreamBuilder_flatMap() {
	var isEOI bool
	it, _ := circle.NewIterator(func() (interface{}, error) {
		if isEOI {
			return nil, circle.ErrEOI
		}
		isEOI = true
		return map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
		}, nil
	})
	_ = circle.NewStreamBuilder(it).
		Flat().
		TupleMap(func(k string, v int) (string, error) {
			return fmt.Sprintf("%s-%d", k, v), nil
		}).
		Sort(func(x, y string) (bool, error) {
			return x < y, nil
		}).
		Consume(func(x string) error {
			fmt.Println(x)
			return nil
		})
	// Output:
	// a-1
	// b-2
	// c-3
}

func ExampleStreamBuilder_consume() {
	it, _ := circle.NewIterator([]int{1, 2, 3, 4, -1, 5, 6, 7})
	err := circle.NewStreamBuilder(it).
		Consume(func(x int) error {
			if x < 0 {
				return fmt.Errorf("negative: %d", x)
			}
			fmt.Println(x)
			return nil
		})
	fmt.Println(err)
	// Output:
	// 1
	// 2
	// 3
	// 4
	// negative: -1
}

func ExampleStreamBuilder_failedToCreateStream1() {
	it, _ := circle.NewIterator([]int{1, 2, 3})
	err := circle.NewStreamBuilder(it).
		Map(func(x int) (int, error) { return x + 1, nil }).    // index 0, valid mapper
		Filter(func(x int) (int, error) { return x * 2, nil }). // index 1, invalid filter!
		Consume(func(x int) error {
			fmt.Println(x)
			return nil
		})
	fmt.Println(err)
	// Output:
	// [1] cannot create stream invalid filter
}

func ExampleStreamBuilder_failedToCreateStream2() {
	it, _ := circle.NewIterator([]int{1, 2, 3})
	err := circle.NewStreamBuilder(it).
		Map(func(x int) (int, error) { return x + 1, nil }).
		Filter(func(x int) (bool, error) { return x&1 == 1, nil }).
		Consume(func(x int) { // invalid consumer!
			fmt.Println(x)
		})
	fmt.Println(err)
	// Output:
	// cannot create stream invalid consumer
}

func ExampleStreamBuilder_failedToYield1() {
	it, _ := circle.NewIterator([]int{1, 2, 3})
	err := circle.NewStreamBuilder(it).
		Map(func(x int) (int, error) { return x, nil }). // index 0
		Filter(func(x int) (bool, error) {               // index 1 returns an error
			if x > 2 {
				return false, fmt.Errorf("ERROR %d", x)
			}
			return x&1 == 0, nil
		}).
		Consume(func(x int) error {
			fmt.Println(x)
			return nil
		})
	fmt.Println(err)
	// Output:
	// 2
	// 1 ERROR 3
}

func ExampleStreamBuilder_failedToYield2() {
	it, _ := circle.NewIterator([]int{1, 2, 3})
	err := circle.NewStreamBuilder(it).
		Map(func(x int) (int, error) { return x + 1, nil }). // index 0
		Filter(func(x int) (bool, error) {                   // index 1 returns an error
			if x > 2 {
				return false, fmt.Errorf("ERROR %d", x)
			}
			return x&1 == 0, nil
		}).
		Map(func(x int) (int, error) { return x * 2, nil }). // index 2 itself returns no error
		Consume(func(x int) error {                          // but receives an error from index 1
			fmt.Println(x)
			return nil
		})
	fmt.Println(err)
	// Output:
	// 4
	// 2 1 ERROR 3
}

func ExampleStreamBuilder_failedToYield3() {
	it, _ := circle.NewIterator([]int{1, 2, 3})
	err := circle.NewStreamBuilder(it).
		Map(func(x int) (int, error) { return x + 1, nil }). // index 0
		Filter(func(x int) (bool, error) {                   // index 1 returns an error
			if x > 2 {
				return false, fmt.Errorf("ERROR %d", x)
			}
			return x&1 == 0, nil
		}, circle.WithNodeID("NID")).                        // override index by node id
		Map(func(x int) (int, error) { return x * 2, nil }). // index 2
		Consume(func(x int) error {
			fmt.Println(x)
			return nil
		})
	fmt.Println(err)
	// Output:
	// 4
	// 2 NID ERROR 3
}

type (
	testcaseStreamBuilderConsume struct {
		title   string
		src     interface{}
		consume func(circle.Iterator, chan<- interface{}) error
		isError bool
		want    []interface{}
	}
)

func (s *testcaseStreamBuilderConsume) test(t *testing.T) {
	it, err := circle.NewIterator(s.src)
	assert.Nil(t, err)
	ch := make(chan interface{})
	go func() {
		assert.Equal(t, s.isError, s.consume(it, ch) != nil)
		close(ch)
	}()
	got := []interface{}{}
	for v := range ch {
		got = append(got, v)
	}
	assert.Equal(t, "", cmp.Diff(s.want, got))
}

func TestStreamBuilderConsume(t *testing.T) {
	for _, tc := range []*testcaseStreamBuilderConsume{
		{
			title: "invalid consumer",
			src:   []int{1, 2, 3},
			consume: func(it circle.Iterator, ch chan<- interface{}) error {
				return circle.NewStreamBuilder(it).
					Consume(func() {
						ch <- 1
					})
			},
			isError: true,
			want:    []interface{}{},
		},
		{
			title: "map consume",
			src:   []int{1, 2, 3},
			consume: func(it circle.Iterator, ch chan<- interface{}) error {
				return circle.NewStreamBuilder(it).
					Map(func(x int) (int, error) {
						return x + 1, nil
					}).
					Consume(func(x int) error {
						ch <- x
						return nil
					})
			},
			want: []interface{}{2, 3, 4},
		},
		{
			title: "maybe consume",
			src:   []circle.Maybe{circle.NewJust(1), circle.NewNothing(), circle.NewJust(10)},
			consume: func(it circle.Iterator, ch chan<- interface{}) error {
				return circle.NewStreamBuilder(it).
					MaybeConsume(func(x int) error {
						ch <- x
						return nil
					}, func() error {
						ch <- "nothing"
						return nil
					})
			},
			want: []interface{}{1, "nothing", 10},
		},
		{
			title: "either consume",
			src:   []circle.Either{circle.NewLeft(1), circle.NewRight(2), circle.NewRight(3)},
			consume: func(it circle.Iterator, ch chan<- interface{}) error {
				return circle.NewStreamBuilder(it).
					EitherConsume(func(x int) error {
						ch <- x
						return nil
					}, func(x int) error {
						ch <- x + 10
						return nil
					})
			},
			want: []interface{}{1, 12, 13},
		},
		{
			title: "tuple consume",
			src:   []circle.Tuple{circle.NewTuple(1, "one"), circle.NewTuple(2, "two"), circle.NewTuple(3, "three")},
			consume: func(it circle.Iterator, ch chan<- interface{}) error {
				return circle.NewStreamBuilder(it).
					TupleConsume(func(x int, y string) error {
						ch <- fmt.Sprintf("%d - %s", x, y)
						return nil
					})
			},
			want: []interface{}{"1 - one", "2 - two", "3 - three"},
		},
	} {
		t.Run(tc.title, tc.test)
	}
}

type (
	testcaseStreamBuilder struct {
		title        string
		src          interface{}
		builder      func(circle.Iterator) circle.StreamBuilder
		wantNewErr   error
		wantYieldErr error
		wantVal      []interface{}
	}
)

func (s *testcaseStreamBuilder) test(t *testing.T) {
	it, err := circle.NewIterator(s.src)
	assert.Nil(t, err)
	x, err := s.builder(it).Execute()
	assert.Equal(t, fmt.Sprint(s.wantNewErr), fmt.Sprint(err))
	if err != nil {
		return
	}
	got := []interface{}{}
	c := x.Channel()
	for v := range c.C() {
		got = append(got, v)
	}
	assert.Equal(t, fmt.Sprint(s.wantYieldErr), fmt.Sprint(c.Err()))
	assert.Equal(t, "", cmp.Diff(s.wantVal, got))
}

func TestStreamBuilder(t *testing.T) {
	for _, tc := range []*testcaseStreamBuilder{
		{
			title:   "just iterator",
			src:     []int{1, 2, 3},
			builder: circle.NewStreamBuilder,
			wantVal: []interface{}{1, 2, 3},
		},
		{
			title: "invalid map",
			src:   []int{1, 2, 3},
			builder: func(it circle.Iterator) circle.StreamBuilder {
				return circle.NewStreamBuilder(it).
					Map(func() {})
			},
			wantNewErr: errors.New("[0] cannot create stream invalid mapper"),
		},
		{
			title: "invalid second filter",
			src:   []int{1, 2, 3},
			builder: func(it circle.Iterator) circle.StreamBuilder {
				return circle.NewStreamBuilder(it).
					Map(func(int) (int, error) { return 0, nil }).
					Filter(func() {}).
					Map(func(int) (int, error) { return 0, nil })
			},
			wantNewErr: errors.New("[1] cannot create stream invalid filter"),
		},
		{
			title: "default first yield error",
			src:   []int{1, 2, 3},
			builder: func(it circle.Iterator) circle.StreamBuilder {
				return circle.NewStreamBuilder(it).
					Filter(func(int) (bool, error) {
						return false, errors.New("ERROR")
					})
			},
			wantYieldErr: errors.New("0 ERROR"),
			wantVal:      []interface{}{},
		},
		{
			title: "specified first yield error",
			src:   []int{1, 2, 3},
			builder: func(it circle.Iterator) circle.StreamBuilder {
				return circle.NewStreamBuilder(it).
					Filter(func(int) (bool, error) {
						return false, errors.New("ERROR")
					}, circle.WithNodeID("NID"))
			},
			wantYieldErr: errors.New("NID ERROR"),
			wantVal:      []interface{}{},
		},
		{
			title: "default second yield error",
			src:   []int{1, 2, 3},
			builder: func(it circle.Iterator) circle.StreamBuilder {
				return circle.NewStreamBuilder(it).
					Map(func(int) (int, error) { return 1, nil }).
					Filter(func(int) (bool, error) {
						return false, errors.New("ERROR")
					}).
					Map(func(int) (int, error) { return 1, nil })
			},
			wantYieldErr: errors.New("2 1 ERROR"),
			wantVal:      []interface{}{},
		},
		{
			title: "specified second yield error",
			src:   []int{1, 2, 3},
			builder: func(it circle.Iterator) circle.StreamBuilder {
				return circle.NewStreamBuilder(it).
					Map(func(int) (int, error) { return 1, nil }, circle.WithNodeID("NID1")).
					Filter(func(int) (bool, error) {
						return false, errors.New("ERROR")
					}, circle.WithNodeID("NID2")).
					Map(func(int) (int, error) { return 1, nil }, circle.WithNodeID("NID3"))
			},
			wantYieldErr: errors.New("NID3 NID2 ERROR"),
			wantVal:      []interface{}{},
		},
		{
			title: "default and specified second yield error",
			src:   []int{1, 2, 3},
			builder: func(it circle.Iterator) circle.StreamBuilder {
				return circle.NewStreamBuilder(it).
					Map(func(int) (int, error) { return 1, nil }, circle.WithNodeID("NID1")).
					Filter(func(int) (bool, error) {
						return false, errors.New("ERROR")
					}, circle.WithNodeID("NID2")).
					Map(func(int) (int, error) { return 1, nil })
			},
			wantYieldErr: errors.New("2 NID2 ERROR"),
			wantVal:      []interface{}{},
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
			wantVal: []interface{}{2, 3, 4},
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
			wantVal: []interface{}{11, 13},
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
			wantVal: []interface{}{7, 23},
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
			wantVal: []interface{}{"right(2)", "left(left)", "right(3)", "left(negative)"},
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
			wantVal: []interface{}{"(((((iv+1)+2)+3)+4)+5)"},
		},
	} {
		t.Run(tc.title, tc.test)
	}
}
