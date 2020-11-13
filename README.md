# BigFile

![](https://github.com/twharmon/bigfile/workflows/Test/badge.svg) [![](https://goreportcard.com/badge/github.com/twharmon/bigfile)](https://goreportcard.com/report/github.com/twharmon/bigfile) [![](https://gocover.io/_badge/github.com/twharmon/bigfile)](https://gocover.io/github.com/twharmon/bigfile)

Use BigFile to treat partitioned files as one.

## Documentation

For full documentation see [pkg.go.dev](https://pkg.go.dev/github.com/twharmon/bigfile).

## Example

```go
package main

import (
	"fmt"

	"github.com/twharmon/bigfile"
)

func main() {
	f := bigfile.Open("foo.txt", 10) // max file size of 10 bytes
	content := []byte("foo bar baz")
	f.Write(content) // notice two files were created
	b := make([]byte, len(content))
	f.Read(b)
	fmt.Println(string(b)) // read "foo bar baz" from two files
}
```

## Contribute

Make a pull request.
