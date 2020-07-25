package circle_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/berquerant/circle"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func ExampleMapper() {
	f, err := circle.NewMapper(func(x int) (bool, error) {
		if x > 0 {
			return true, nil
		}
		if x == 0 {
			return false, nil
		}
		return false, errors.New("out of range")
	})
	if err != nil {
		panic(err)
	}
	for _, x := range []int{1, 0, -1} {
		fmt.Println(f.Apply(x))
	}
	// Output:
	// true <nil>
	// false <nil>
	// false out of range
}

func TestMapper(t *testing.T) {
	for name, tc := range map[string]func(t *testing.T){
		"invalid": testInvalidMapper,
		"apply":   testMapperApply,
	} {
		t.Run(name, tc)
	}
}

func testInvalidMapper(t *testing.T) {
	_, err := circle.NewMapper(func() {})
	assert.Equal(t, circle.ErrInvalidMapper, err)
}

func testMapperApply(t *testing.T) {
	for name, tc := range map[string]func(t *testing.T){
		"invalid argument": func(t *testing.T) {
			f, err := circle.NewMapper(strconv.Atoi)
			assert.Nil(t, err)
			_, err = f.Apply(100)
			assert.NotNil(t, err)
		},
		"atoi": func(t *testing.T) {
			f, err := circle.NewMapper(strconv.Atoi)
			assert.Nil(t, err)
			v, err := f.Apply("100")
			assert.Nil(t, err)
			assert.Equal(t, v, 100)
		},
		"sum": func(t *testing.T) {
			f, err := circle.NewMapper(func(xs []int) (int, error) {
				var s int
				for _, x := range xs {
					s += x
				}
				return s, nil
			})
			assert.Nil(t, err)
			v, err := f.Apply([]int{1, 2, 3, 4})
			assert.Nil(t, err)
			assert.Equal(t, v, 10)
		},
		"sumslice": func(t *testing.T) {
			f, err := circle.NewMapper(func(xss [][]int) ([]int, error) {
				ss := make([]int, len(xss))
				for i, xs := range xss {
					for _, x := range xs {
						ss[i] += x
					}
				}
				return ss, nil
			})
			assert.Nil(t, err)
			v, err := f.Apply([][]int{[]int{1}, []int{2, 3}, []int{4, 5, 6}})
			assert.Nil(t, err)
			assert.Equal(t, "", cmp.Diff([]int{1, 5, 15}, v))
		},
	} {
		t.Run(name, tc)
	}
}

func TestFilter(t *testing.T) {
	for name, tc := range map[string]func(t *testing.T){
		"invalid": testInvalidFilter,
		"apply":   testFilterApply,
	} {
		t.Run(name, tc)
	}
}

func testInvalidFilter(t *testing.T) {
	_, err := circle.NewFilter(func() {})
	assert.Equal(t, circle.ErrInvalidFilter, err)
}

func testFilterApply(t *testing.T) {
	t.Run("invalid argument", func(t *testing.T) {
		f, err := circle.NewFilter(func(string) (bool, error) { return true, nil })
		assert.Nil(t, err)
		_, err = f.Apply(1)
		assert.NotNil(t, err)
	})

	t.Run("do", func(t *testing.T) {
		f, err := circle.NewFilter(func(x string) (bool, error) {
			if x == "" {
				return false, errors.New("empty")
			}
			return len(x) < 5, nil
		})
		assert.Nil(t, err)
		{
			_, err := f.Apply("")
			assert.Equal(t, errors.New("empty"), err)
		}
		{
			v, err := f.Apply("cat")
			assert.Nil(t, err)
			assert.True(t, v)
		}
		{
			v, err := f.Apply("timeline")
			assert.Nil(t, err)
			assert.False(t, v)
		}
	})
}

func TestAggregator(t *testing.T) {
	for name, tc := range map[string]func(t *testing.T){
		"type":  testAggregatorType,
		"apply": testAggregatorApply,
	} {
		t.Run(name, tc)
	}
}

func testAggregatorType(t *testing.T) {
	for name, tc := range map[string]func(t *testing.T){
		"invalid": func(t *testing.T) {
			_, err := circle.NewAggregator(func() {})
			assert.Equal(t, circle.ErrInvalidAggregator, err)
		},
		"invalid argument": func(t *testing.T) {
			f, err := circle.NewAggregator(func(int, int) (int, error) { return 0, nil })
			assert.Nil(t, err)
			_, err = f.Apply("1", "2")
			assert.NotNil(t, err)
		},
		"right": func(t *testing.T) {
			f, err := circle.NewAggregator(func(_ int, _ string) (string, error) { return "", nil })
			assert.Nil(t, err)
			assert.Equal(t, circle.RightAggregatorType, f.Type())
		},
		"left": func(t *testing.T) {
			f, err := circle.NewAggregator(func(_ string, _ int) (string, error) { return "", nil })
			assert.Nil(t, err)
			assert.Equal(t, circle.LeftAggregatorType, f.Type())
		},
		"perfect": func(t *testing.T) {
			f, err := circle.NewAggregator(func(_ string, _ string) (string, error) { return "", nil })
			assert.Nil(t, err)
			assert.Equal(t, circle.PerfectAggregatorType, f.Type())
		},
	} {
		t.Run(name, tc)
	}
}

func testAggregatorApply(t *testing.T) {
	f, err := circle.NewAggregator(func(x string, y int) (int, error) {
		i, err := strconv.Atoi(x)
		if err != nil {
			return 0, err
		}
		return i + y, nil
	})
	assert.Nil(t, err)
	{
		_, err := f.Apply("", 10)
		assert.NotNil(t, err)
	}
	{
		v, err := f.Apply("10", 1)
		assert.Nil(t, err)
		assert.Equal(t, 11, v)
	}
}

func TestComparator(t *testing.T) {
	for name, tc := range map[string]func(t *testing.T){
		"invalid": testInvalidComparator,
		"apply":   testComparatorApply,
	} {
		t.Run(name, tc)
	}
}

func testInvalidComparator(t *testing.T) {
	_, err := circle.NewComparator(func() {})
	assert.Equal(t, circle.ErrInvalidComparator, err)
}

func testComparatorApply(t *testing.T) {
	t.Run("invalid argument", func(t *testing.T) {
		f, err := circle.NewComparator(func(int, int) (bool, error) {
			return true, nil
		})
		assert.Nil(t, err)
		_, err = f.Apply("1", "2")
		assert.NotNil(t, err)
	})

	t.Run("do", func(t *testing.T) {
		type T struct {
			i int
		}
		f, err := circle.NewComparator(func(x, y *T) (bool, error) {
			if x == nil || y == nil {
				return false, errors.New("white")
			}
			return x.i < y.i, nil
		})
		assert.Nil(t, err)
		{
			var x *T
			_, err := f.Apply(x, &T{10})
			assert.Equal(t, errors.New("white"), err)
		}
		{
			v, err := f.Apply(&T{5}, &T{10})
			assert.Nil(t, err)
			assert.True(t, v)
		}
		{
			v, err := f.Apply(&T{10}, &T{5})
			assert.Nil(t, err)
			assert.False(t, v)
		}
	})
}

func TestConsumer(t *testing.T) {
	for name, tc := range map[string]func(*testing.T){
		"invalid": testInvalidConsumer,
		"apply":   testConsumerApply,
	} {
		t.Run(name, tc)
	}
}

func testInvalidConsumer(t *testing.T) {
	_, err := circle.NewConsumer(func() {})
	assert.Equal(t, circle.ErrInvalidConsumer, err)
}

func testConsumerApply(t *testing.T) {
	t.Run("invalid argument", func(t *testing.T) {
		f, err := circle.NewConsumer(func(int) error {
			return nil
		})
		assert.Nil(t, err)
		assert.NotNil(t, f.Apply("1"))
	})

	t.Run("do", func(t *testing.T) {
		type T struct {
			s string
		}
		var ret string
		f, err := circle.NewConsumer(func(x T) error {
			if x.s == "bug" {
				return errors.New("bug found")
			}
			ret = x.s
			return nil
		})
		assert.Nil(t, err)
		{
			assert.Equal(t, errors.New("bug found"), f.Apply(T{"bug"}))
		}
		{
			ret = ""
			assert.Nil(t, f.Apply(T{"green"}))
			assert.Equal(t, "green", ret)
		}
	})
}
