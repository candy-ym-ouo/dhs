package store

import "errors"

var ErrNotFound = errors.New("store: node not found")
var ErrConflict = errors.New("store: concurrent update conflict")
var ErrClosed = errors.New("store: database is closed")

func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }
func IsConflict(err error) bool { return errors.Is(err, ErrConflict) }
