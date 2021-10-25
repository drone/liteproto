package liteproto

import "errors"

// ErrUnknownType is returned when the system encounters an unregistered type.
var ErrUnknownType = errors.New("unrecognized type")

// ErrCallFailed is returned by Caller when a remote server returns an error.
var ErrCallFailed = errors.New("call failed")

// ErrAlreadySubscribed is returned by ResponseHandler when there is already a subscription to an ID.
var ErrAlreadySubscribed = errors.New("subscription already exists for ID")

// ErrNotSubscribed is returned by ResponseHandler when there is no subscription to an ID.
var ErrNotSubscribed = errors.New("no subscription for ID")
