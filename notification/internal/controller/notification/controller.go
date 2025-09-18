package notification

import "errors"

var ErrTooManyMessages = errors.New("too many messages sent to given user")
