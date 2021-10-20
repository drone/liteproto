package internal

import (
	"context"
	"time"

	"github.com/drone/liteproto/liteproto"
)

// Caller makes a call to a remote server to run a job. It is used by Runner to make the call.
// This enables abstraction of remote server calls.
type Caller interface {
	Call(ctx context.Context, jobID, jobType string, data []byte, deadline time.Time) (err error)
}

// Feeder is an interface
type Feeder interface {
	Feed(ctx context.Context, jobID, jobType string, data []byte, deadline time.Time) error
}

// ResponsePub is publisher part of response publisher/subscriber interface.
type ResponsePub interface {
	Publish(response liteproto.Message) (err error)
}

// ResponseSub is subscriber part of response publisher/subscriber interface.
type ResponseSub interface {
	Subscribe(jobID string) (responseData <-chan liteproto.Message, err error)
	Unsubscribe(jobID string) (err error)
}

// ResponsePubSub is response publisher/subscriber interface.
type ResponsePubSub interface {
	ResponsePub
	ResponseSub
}

// ResponderFactory is a generator of ResponderClient objects.
type ResponderFactory interface {
	Client() liteproto.Client
	MakeResponder(jobID, jobType string) liteproto.ResponderClient
}
