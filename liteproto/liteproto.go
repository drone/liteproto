package liteproto

import (
	"context"
	"time"
)

type ServerClient interface {
	Server
	Client
}

// Server allows jobs to be registered.
type Server interface {
	// Register registers an Execer to run jobs for the provided job type.
	Register(jobType string, execer Execer)

	// RegisterWithResponder registers an ExecerWithResponder to run jobs for the provided job type.
	// The ExecerWithResponder implementation has ability to send messages back to the caller.
	RegisterWithResponder(jobType string, execer ExecerWithResponder)

	// RegisterCatchAll registers an ExecerWithResponder to all jobs that are not already registered.
	RegisterCatchAll(execer ExecerWithResponder)
}

// Client allows calls to a remote server.
type Client interface {
	// Call executes a job on a remote server.
	// An implementation of Execer over there will execute the job.
	Call(ctx context.Context, job Message) (err error)

	// CallWithResponse executes a job on a remote server.
	// An implementation of ExecerWithResponder can send responses back to the caller.
	CallWithResponse(ctx context.Context, job Message) (response <-chan Message, stop chan<- struct{}, err error)

	// CallWithDeadline executes a job on a remote server.
	// An implementation of ExecerWithResponder can will receive a Context with the deadline specified as a parameter.
	// The deadline refers to the deadline until the caller accepts responses.
	CallWithDeadline(ctx context.Context, job Message, deadline time.Time) (response <-chan Message, stop chan<- struct{}, err error)
}

// Responder is a simple interface to send a response.
type Responder interface {
	Respond(ctx context.Context, data []byte) error
	RespondWithType(ctx context.Context, jobType string, data []byte) error
}

type ResponderClient interface {
	Client
	Responder
}

// Execer is an interface that runs jobs. It should be passed as a parameter to Register method.
type Execer interface {
	Exec(ctx context.Context, job Message, client Client)
}

// ExecerWithResponder is an interface that runs jobs. It should be passed as a parameter to RegisterWithResponder method.
// The ResponderClient parameter can be used to send a response back to the caller.
type ExecerWithResponder interface {
	Exec(ctx context.Context, job Message, client ResponderClient)
}
