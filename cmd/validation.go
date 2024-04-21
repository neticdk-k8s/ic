package cmd

import (
	"fmt"
	"strings"
)

type InvalidArgumentError struct {
	Flag     string
	Val      string
	OneOf    []string
	SeeOther string
	Context  string
}

func (e *InvalidArgumentError) Error() string {
	ret := fmt.Sprintf(`invalid argument "%s" for "%s" flag`, e.Val, e.Flag)
	if len(e.OneOf) > 0 {
		ret = fmt.Sprintf("%s: must be one of: %s", ret, strings.Join(e.OneOf, "|"))
	}
	if e.SeeOther != "" {
		ret = fmt.Sprintf(`%s: see "%s"`, ret, e.SeeOther)
	}
	if e.Context != "" {
		ret = fmt.Sprintf(`%s: %s`, ret, e.Context)
	}
	return ret
}
