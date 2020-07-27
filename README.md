## circle [![GoDoc](https://godoc.org/github.com/berquerant/circle?status.svg)](https://godoc.org/github.com/berquerant/circle)

circle provides sequences which support aggregation operations.

### Example

``` go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/berquerant/circle"
)

// tr -d ' ' | tr '[:upper:]' '[:lower:]' | grep -o . | sort | uniq -c | sort -rnk 1 | awk '{print $2, $1}'
func main() {
	sc := bufio.NewScanner(os.Stdin)
	it, _ := circle.NewIterator(func() (interface{}, error) {
		if sc.Scan() {
			return sc.Text(), nil
		}
		if err := sc.Err(); err != nil {
			return nil, err
		}
		return nil, circle.ErrEOI
	})

	st, _ := circle.NewStreamBuilder(it).
		Filter(func(x string) (bool, error) {
			return x != "", nil
		}).
		Map(func(x string) (string, error) {
			return strings.ReplaceAll(x, " ", ""), nil
		}).
		Map(func(x string) (string, error) {
			return strings.ToLower(x), nil
		}).
		Map(func(x string) ([]string, error) {
			return strings.Split(x, ""), nil
		}).
		Flat().
		Aggregate(func(d map[string]int, x string) (map[string]int, error) {
			d[x]++
			return d, nil
		}, map[string]int{}).
		Flat().
		Sort(func(x, y circle.Tuple) (bool, error) {
			nx, _ := x.Get(1)
			ny, _ := y.Get(1)
			return nx.(int) > ny.(int), nil
		}).
		TupleMap(func(x string, y int) (string, error) {
			return fmt.Sprintf("%s %d", x, y), nil
		}).
		Execute()

	for x := range st.Channel().C() {
		fmt.Println(x)
	}
}
```

### Test

``` shell
make test
```
