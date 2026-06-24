package playwright

import (
	"errors"
	"fmt"
)

var (
	// ErrPlaywright wraps all Playwright errors.
	//   - Use errors.Is to check if the error is a Playwright error.
	//   - Use errors.As to cast an error to [Error] if you want to access "Stack".
	ErrPlaywright = errors.New("playwright")
	// ErrTargetClosed usually wraps a reason.
	ErrTargetClosed = errors.New("target closed")
	// ErrTimeout wraps timeout errors. It can be either Playwright TimeoutError or client timeout.
	ErrTimeout = errors.New("timeout")
)

// Error represents a Playwright error
type Error struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Stack   string `json:"stack"`
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Is(target error) bool {
	err, ok := target.(*Error)
	if !ok {
		return false
	}
	if err.Name != e.Name {
		return false
	}
	if e.Name != "Error" {
		return true // same name and not normal error
	}
	return e.Message == err.Message
}

func parseError(err Error) error {
	switch err.Name {
	case "TimeoutError":
		return fmt.Errorf("%w: %w: %w", ErrPlaywright, ErrTimeout, &err)
	case "TargetClosedError":
		return fmt.Errorf("%w: %w: %w", ErrPlaywright, ErrTargetClosed, &err)
	}
	return fmt.Errorf("%w: %w", ErrPlaywright, &err)
}

func targetClosedError(reason *string) error {
	if reason == nil {
		return ErrTargetClosed
	}
	return fmt.Errorf("%w: %s", ErrTargetClosed, *reason)
}
