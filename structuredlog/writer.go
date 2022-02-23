package structuredlog

import (
	"io"
	"os"
)

var out io.Writer = os.Stdout

func SetWriter( o io.Writer) {
	out = o
}
