package render

import (
	"fmt"
	"io"
)

func String(body []byte, writer io.Writer) {
	fmt.Fprintln(writer, string(body))
}
