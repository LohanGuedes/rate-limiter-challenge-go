package gateway

import "errors"

// ErrNotFound is returned when the data is not found.
var ErrNotFound = errors.New("not found")

// ErrTooManyMessages is returned when too many messags of given data are send for a config limit.
var ErrTooManyMessages = errors.New("too many messages sent to given user")
