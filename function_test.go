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
