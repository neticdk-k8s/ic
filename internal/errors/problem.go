package errors

import (
	"fmt"

	"github.com/neticdk-k8s/ic/internal/apiclient"
)

type ProblemError struct {
	Title   string
	Problem *apiclient.Problem
}

func (e *ProblemError) Error() string {
	var title, detail string
	if e.Problem.Title != nil {
		title = *e.Problem.Title
	}
	if e.Problem.Detail != nil {
		detail = fmt.Sprintf(" (%s)", *e.Problem.Detail)
	}
	return fmt.Sprintf("%s: %s%s", e.Title, title, detail)
}
