package cmdutil

import (
	"errors"
	"fmt"
)

// FlagErrorf returns a new FlagError that wraps an error produced by
// fmt.Errorf(format, args...).
func FlagErrorf(format string, args ...interface{}) error {
	return FlagErrorWrap(fmt.Errorf(format, args...))
}

// FlagErrorWrap returns a new FlagError that wraps the specified error.
func FlagErrorWrap(err error) error { return &FlagError{err} }

// A *FlagError indicates an error processing command-line flags or other arguments.
// Such errors cause the application to display the usage message.
type FlagError struct {
	err error
}

func (fe *FlagError) Error() string {
	return fe.err.Error()
}

func (fe *FlagError) Unwrap() error {
	return fe.err
}

// SilentError is an error that triggers exit code 1 without any error messaging
var SilentError = errors.New("SilentError")

// CancelError signals user-initiated cancellation
var CancelError = errors.New("CancelError")

// SafeModeError indicates a command was blocked due to safe mode
type SafeModeError struct {
	Command string
}

func (e *SafeModeError) Error() string {
	return fmt.Sprintf("'fm %s' is disabled in safe mode.\n\n"+
		"Safe mode is active because stdin is not a terminal.\n"+
		"This prevents accidental destructive actions when run by AI or scripts.\n\n"+
		"To override, use one of:\n"+
		"  fm %s ... --unsafe     # Allow this command\n"+
		"  FM_UNSAFE=1 fm %s ...  # Allow via environment",
		e.Command, e.Command, e.Command)
}

// AuthError indicates an authentication problem
type AuthError struct {
	err error
}

func (ae *AuthError) Error() string {
	return ae.err.Error()
}

func NewAuthError(msg string) *AuthError {
	return &AuthError{err: errors.New(msg)}
}

// NotFoundError indicates a resource was not found
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID '%s' not found", e.Resource, e.ID)
}

// MutuallyExclusive returns an error if more than one condition is true
func MutuallyExclusive(message string, conditions ...bool) error {
	numTrue := 0
	for _, ok := range conditions {
		if ok {
			numTrue++
		}
	}
	if numTrue > 1 {
		return FlagErrorf("%s", message)
	}
	return nil
}
