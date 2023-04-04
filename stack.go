package golog

import "github.com/pkg/errors"

// stackTracer interface.
type stackTracer interface {
	StackTrace() errors.StackTrace
}
