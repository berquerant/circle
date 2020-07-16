package circle_test

import (
	"circle"
	"errors"
	"fmt"
	"strconv"
	"testing"

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
