# BigFile

![](https://github.com/twharmon/bigfile/workflows/Test/badge.svg) [![](https://goreportcard.com/badge/github.com/twharmon/bigfile)](https://goreportcard.com/report/github.com/twharmon/bigfile) [![](https://gocover.io/_badge/github.com/twharmon/bigfile)](https://gocover.io/github.com/twharmon/bigfile)

Use BigFile to treat partitioned files as one.

## Documentation

For full documentation see [pkg.go.dev](https://pkg.go.dev/github.com/twharmon/bigfile).

## Example

```go
package main

import (
	"log"

	"github.com/twharmon/bigfile"
)

func main() {
	f := bigfile.Open("foo.txt", 10) // max file size of 10 bytes
	f.Write([]byte("foo bar baz"))
}
```

## Contribute

Make a pull request.
