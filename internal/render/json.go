package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// PrettyPrintJSON pretty prints JSON
func PrettyPrintJSON(body []byte, writer io.Writer) error {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, body, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(writer, prettyJSON.String())
	return nil
}

// BytesToBinarySI converts bytes to human readable string using binary SI units
func BytesToBinarySI(bytes int64) (float64, string) {
	const (
		kibi float64 = 1024
		mebi float64 = 1048576
		gibi float64 = 1073741824
		tebi float64 = 1099511627776
		pebi float64 = 1125899906842624
	)

	b := float64(bytes)
	if b >= pebi {
		return b / pebi, "PiB"
	} else if b >= tebi {
		return b / tebi, "TiB"
	} else if b >= gibi {
		return b / gibi, "GiB"
	} else if b >= mebi {
		return b / mebi, "MiB"
	} else if b >= kibi {
		return b / kibi, "KiB"
	}
	return b, "B"
}
