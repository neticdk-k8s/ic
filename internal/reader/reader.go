package reader

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Reader interface {
	// ReadString reads a string from stdin
	ReadString(prompt string) (string, error)
}

type reader struct {
	Stdin io.Reader
}

// NewReader creates a new Reader
func NewReader() *reader {
	return &reader{
		Stdin: os.Stdin,
	}
}

// ReadString reads a string from stdin
func (r *reader) ReadString(prompt string) (string, error) {
	if _, err := fmt.Fprint(os.Stderr, prompt); err != nil {
		return "", fmt.Errorf("write error: %w", err)
	}
	br := bufio.NewReader(r.Stdin)
	s, err := br.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read error: %w", err)
	}
	s = strings.TrimRight(s, "\r\n")
	return s, nil
}
