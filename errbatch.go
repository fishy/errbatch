package errbatch

import (
	"errors"
	"fmt"
	"strings"
)

// Make sure *ErrBatch satisfies error interface.
// (ErrBatch satisfies error interface as well.)
var _ error = (*ErrBatch)(nil)

// ErrBatch is an error that can contain multiple errors.
//
// The zero value of ErrBatch is valid (with no errors) and ready to use.
type ErrBatch struct {
	errors []error
}

// Error satisfies the error interface.
func (eb ErrBatch) Error() string {
	var builder strings.Builder
	fmt.Fprintf(
		&builder,
		"errbatch: total %d error(s) in this batch",
		len(eb.errors),
	)
	for i, err := range eb.errors {
		if i == 0 {
			builder.WriteString(": ")
		} else {
			builder.WriteString("; ")
		}
		fmt.Fprintf(&builder, "%+v", err)
	}
	return builder.String()
}

// As implements helper interface for errors.As.
func (eb ErrBatch) As(v interface{}) bool {
	if target, ok := v.(*ErrBatch); ok {
		target.errors = eb.GetErrors()
		return true
	}
	return false
}

// Unwrap implements the hidden errors interface.
//
// When the batch contains exactly one error, that error is returned.
// It returns nil otherwise.
func (eb ErrBatch) Unwrap() error {
	if len(eb.errors) == 1 {
		return eb.errors[0]
	}
	return nil
}

func (eb *ErrBatch) addBatch(batch *ErrBatch) {
	eb.errors = append(eb.errors, batch.errors...)
}

// Add adds an error into the batch.
//
// If the error is also an ErrBatch,
// its underlying error(s) will be added instead of the ErrBatch itself.
//
// Nil error will be skipped.
func (eb *ErrBatch) Add(err error) {
	if err == nil {
		return
	}

	var batch ErrBatch
	if errors.As(err, &batch) {
		eb.addBatch(&batch)
	} else {
		eb.errors = append(eb.errors, err)
	}
}

// Compile compiles the batch.
//
// If the batch contains zero errors, it will return nil.
//
// If the batch contains exactly one error,
// that underlying error will be returned.
//
// Otherwise, the batch itself will be returned.
func (eb *ErrBatch) Compile() error {
	switch len(eb.errors) {
	case 0:
		return nil
	case 1:
		return eb.errors[0]
	default:
		return eb
	}
}

// Clear clears the batch.
func (eb *ErrBatch) Clear() {
	eb.errors = make([]error, 0)
}

// GetErrors returns a copy of the underlying error(s).
func (eb *ErrBatch) GetErrors() []error {
	errors := make([]error, len(eb.errors))
	copy(errors, eb.errors)
	return errors
}
