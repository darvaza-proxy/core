package core

import "strings"

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

// AppendError adds an error to the collection,
// unwrapping other implementers of the [Errors]
// interface when possible
func (w *CompoundError) AppendError(err error) {
	switch v := err.(type) {
	case *CompoundError:
		// one of us
		w.Errs = append(w.Errs, v.Errs...)
	case Errors:
		// cosin, I can't trust you don't have nil entries
		// there
		for _, e := range v.Errors() {
			w.AppendError(e)
		}
	case nil:
		// skip
	default:
		// just a normal error
		w.Errs = append(w.Errs, err)
	}
}
