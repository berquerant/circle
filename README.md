## circle

circle provides sequences which support aggregation operations.

[![GoDoc](https://godoc.org/github.com/berquerant/circle?status.svg)](https://godoc.org/github.com/berquerant/circle)

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
	_ = circle.NewStreamBuilder(circle.MustNewIterator(func() (interface{}, error) {
		if sc.Scan() {
			return sc.Text(), nil
		}
		if err := sc.Err(); err != nil {
			return nil, err
		}
		return nil, circle.ErrEOI
	})).
		Filter(func(x string) bool { return x != "" }).
		Map(func(x string) string { return strings.ReplaceAll(x, " ", "") }).
		Map(strings.ToLower).
		Map(func(x string) []string { return strings.Split(x, "") }).
		Flat().
		Aggregate(func(d map[string]int, x string) map[string]int {
			d[x]++
			return d
		}, map[string]int{}).
		Flat().
		Sort(func(x, y circle.Tuple) bool { return x.MustGet(1).(int) > y.MustGet(1).(int) }).
		TupleMap(func(x string, y int) string { return fmt.Sprintf("%s %d", x, y) }).
		Consume(func(x string) { fmt.Println(x) })
}
```

### Test

``` shell
make test
```
