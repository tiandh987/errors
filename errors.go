package errors

import "fmt"

type withCode struct {
	err   error
	code  int
	cause error
	*stack
}

func (w *withCode) Error() string {
	return fmt.Sprintf("%v", w)
}