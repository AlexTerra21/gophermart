package errs

import "errors"

// ErrConflict указывает на конфликт данных в хранилище.
var ErrConflict = errors.New("data conflict")

var ErrNoContent = errors.New("no content")
