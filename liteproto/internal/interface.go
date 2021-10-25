package internal

import (
	"context"
	"time"

	"github.com/drone/liteproto/liteproto"
)

// Caller makes a call to a remote server to execute a task. It is used by a Runner to make the call.
// This enables abstraction of remote server calls.
type Caller interface {
	Call(ctx context.Context, request liteproto.TaskRequest, deadline time.Time) (err error)
}

// Feeder accepts new task execution requests. Deadline parameter
// should be zero time if it's not needed (equal to time.Time{}).
type Feeder interface {
	Feed(ctx context.Context, request liteproto.TaskRequest, deadline time.Time) error
}

// ResponsePub is publisher part of response publisher/subscriber interface.
type ResponsePub interface {
	Publish(response liteproto.TaskResponse) (err error)
}

// ResponseSub is subscriber part of response publisher/subscriber interface.
type ResponseSub interface {
	Subscribe(id string) (responseData <-chan liteproto.TaskResponse, err error)
	Unsubscribe(id string) (err error)
}

// ResponsePubSub is response publisher/subscriber interface.
type ResponsePubSub interface {
	ResponsePub
	ResponseSub
}

// ResponderFactory is a generator of ResponderClient objects.
type ResponderFactory interface {
	Client() liteproto.Client
	MakeResponder(id, t string) liteproto.ResponderClient
}
