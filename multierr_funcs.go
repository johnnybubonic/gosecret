package gosecret

import (
	"fmt"
)

/*
	NewErrors returns a new MultiError based on a slice of error.Error (errs).
	Any nil errors are trimmed. If there are no actual errors after trimming, err will be nil.
*/
func NewErrors(errs ...error) (err error) {

	if errs == nil || len(errs) == 0 {
		return
	}

	var realErrs []error = make([]error, 0)

	for _, e := range errs {
		if e == nil {
			continue
		}
		realErrs = append(realErrs, e)
	}

	if len(realErrs) == 0 {
		return
	}

	err = &MultiError{
		Errors:   realErrs,
		ErrorSep: "\n",
	}

	return
}

// Error makes a MultiError conform to the error interface.
func (e *MultiError) Error() (errStr string) {

	var numErrs int

	if e == nil || len(e.Errors) == 0 {
		return
	} else {
		numErrs = len(e.Errors)
	}

	for idx, err := range e.Errors {
		if (idx + 1) < numErrs {
			errStr += fmt.Sprintf("%v%v", err.Error(), e.ErrorSep)
		} else {
			errStr += err.Error()
		}
	}

	return
}
