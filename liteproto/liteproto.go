package liteproto

import (
	"context"
	"time"
)

type ServerClient interface {
	Server
	Client
}

// Server allows objects that implement Execer interface to be registered to execute tasks.
type Server interface {
	// Register registers an Execer to run tasks for the provided task type.
	Register(t string, execer Execer)

	// RegisterWithResponder registers an ExecerWithResponder to run tasks for the provided task type.
	// The ExecerWithResponder implementation has ability to send messages back to the caller.
	RegisterWithResponder(t string, execer ExecerWithResponder)

	// RegisterCatchAll registers an ExecerWithResponder to run all tasks that are not already registered.
	RegisterCatchAll(execer ExecerWithResponder)
}

// Client allows calls to a remote server.
type Client interface {
	// Call executes a task on a remote server.
	// An implementation of Execer interface over there will execute the task.
	Call(ctx context.Context, request TaskRequest) (err error)

	// CallWithResponse executes a task on a remote server.
	// An implementation of ExecerWithResponder can send responses back to the caller.
	CallWithResponse(ctx context.Context, request TaskRequest) (response <-chan TaskResponse, stop chan<- struct{}, err error)

	// CallWithDeadline executes a task on a remote server.
	// An implementation of ExecerWithResponder can will receive a Context with the deadline specified as a parameter.
	// The deadline refers to the deadline until the caller accepts responses.
	CallWithDeadline(ctx context.Context, request TaskRequest, deadline time.Time) (response <-chan TaskResponse, stop chan<- struct{}, err error)
}

// Responder is a simple interface to send a response. Value of the parameter status must not be an empty string.
type Responder interface {
	Respond(ctx context.Context, status string, data []byte) error
	RespondWithType(ctx context.Context, newType, status string, data []byte) error
}

type ResponderClient interface {
	Client
	Responder
}

// Execer is an interface that runs tasks. It should be passed as a parameter to Register method.
type Execer interface {
	Exec(ctx context.Context, request TaskRequest, client Client)
}

// ExecerWithResponder is an interface that runs tasks. It should be passed as a parameter to RegisterWithResponder method.
// The ResponderClient parameter can be used to send a response back to the caller.
type ExecerWithResponder interface {
	Exec(ctx context.Context, request TaskRequest, client ResponderClient)
}
