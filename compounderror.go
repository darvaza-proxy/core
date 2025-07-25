package core

import (
	"errors"
	"fmt"
	"strings"
)

// Errors in an error that contains a slice of errors
type Errors interface {
	Error() string
	Errors() []error
}

var (
	_ Errors = (*CompoundError)(nil)
)

// A CompoundError can contain more that one error
type CompoundError struct {
	Errs []error
}

func (w *CompoundError) Error() string {
	s := make([]string, 0, len(w.Errs))
	for _, err := range w.Errs {
		if err != nil {
			s = append(s, err.Error())
		}
	}

	return strings.Join(s, "\n")
}

// Errors returns the contained slice of errors
func (w *CompoundError) Errors() []error {
	return w.Errs
}

// Unwrap returns the contained slice of errors
func (w *CompoundError) Unwrap() []error {
	return w.Errs
}

// Ok tells when there are no errors stored
func (w *CompoundError) Ok() bool {
	return len(w.Errs) == 0
}

// AsError returns itself as an `error` when
// there are errors stored, and nil when there aren't
func (w *CompoundError) AsError() error {
	if len(w.Errs) > 0 {
		return w
	}
	return nil
}

// AppendError adds an error to the collection, unwrapping other implementers of the [Errors]
// interface when possible. It returns the updated CompoundError for method chaining.
func (w *CompoundError) AppendError(errs ...error) *CompoundError {
	for _, err := range errs {
		if err != nil {
			w.doAppendUnwrapped(err)
		}
	}
	return w
}

func (w *CompoundError) doAppendUnwrapped(err error) {
	switch v := err.(type) {
	case Errors:
		w.doAppend(v.Errors()...)
	case interface{ Unwrap() []error }:
		w.doAppend(v.Unwrap()...)
	default:
		w.doAppend(v)
	}
}

func (w *CompoundError) doAppend(errs ...error) {
	for _, err := range errs {
		if err != nil {
			w.Errs = append(w.Errs, err)
		}
	}
}

// Append adds an error to the collection, optionally annotating it with a formatted note.
// If err is nil and a note is provided, a new error is created from the note.
// If both err and note are non-empty, the error is wrapped with the note.
// Returns the updated CompoundError for method chaining.
func (w *CompoundError) Append(err error, note string, args ...any) *CompoundError {
	if len(args) > 0 {
		note = fmt.Sprintf(note, args...)
	}

	switch {
	case err == nil && note == "":
		// nothing
		return w
	case err == nil:
		// note-only
		err = errors.New(note)
	case note != "":
		// wrap
		err = Wrap(err, note)
	}

	w.Errs = append(w.Errs, err)
	return w
}
