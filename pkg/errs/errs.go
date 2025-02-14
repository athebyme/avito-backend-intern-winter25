package errs

import "errors"

var (
	ErrTransactionNotFound = errors.New("transaction not found")
)
