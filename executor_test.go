package circle_test

import (
	"circle"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapExecutor(t *testing.T) {
	it, err := circle.NewIterator([]int{-1, 0, 1})
	assert.Nil(t, err)
	f, err := circle.NewMapper(func(x int) (int, error) {
		if x < 0 {
			return 0, errors.New("negative")
		}
		return x + 10, nil
	})
	assert.Nil(t, err)
	exit, err := circle.NewMapExecutor(f, it).Execute()
	assert.Nil(t, err)
	{
		v, err := exit.Next()
		assert.Nil(t, err)
		assert.Equal(t, 10, v)
	}
	{
		v, err := exit.Next()
		assert.Nil(t, err)
		assert.Equal(t, 11, v)
	}
	{
		_, err := exit.Next()
		assert.Equal(t, circle.ErrEOI, err)
	}
}
