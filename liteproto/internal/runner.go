package internal

import (
	"context"
	"time"

	"github.com/drone/liteproto/liteproto"
)

// NewRunner creates a new Runner.
func NewRunner(caller Caller, respSub ResponseSub) *Runner {
	return &Runner{
		caller:  caller,
		respSub: respSub,
	}
}

// Runner is used to make a call to a remote server to run a job.
// The call is made by an implementation of Caller interface.
// Responses are handled by an implementation of ResponseSub interface.
type Runner struct {
	caller  Caller
	respSub ResponseSub
}

// Run method makes a call to a remote server. The response includes two channels,
// one for receiving a response (or potentially several responses) and a stop channel.
// The stop channel should be closed by the caller when no further responses are expected.
// If the function returns an error both channels will be nil.
func (rq *Runner) Run(ctx context.Context, jobID, jobType string, data []byte, deadline time.Time) (<-chan liteproto.Message, chan<- struct{}, error) {
	if jobID == "" {
		panic("ID must not be empty")
	}

	if jobType == "" {
		panic("Type must not be empty")
	}

	if !deadline.IsZero() && deadline.Before(time.Now()) {
		return nil, nil, context.DeadlineExceeded
	}

	var err error
	err = rq.caller.Call(ctx, jobID, jobType, data, deadline)
	if err != nil {
		return nil, nil, err
	}

	outChan, err := rq.respSub.Subscribe(jobID)
	if err != nil {
		return nil, nil, err
	}

	var ctxJob context.Context
	var cancelFunc func()

	if deadline.IsZero() {
		ctxJob = ctx
		cancelFunc = func() {}
	} else {
		ctxJob, cancelFunc = context.WithDeadline(ctx, deadline)
	}

	responseChan := make(chan liteproto.Message)
	stopChan := make(chan struct{})

	go func(ctx context.Context) {
		defer func() {
			_ = rq.respSub.Unsubscribe(jobID)
			close(responseChan)
			cancelFunc()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-stopChan: // the caller closes stop channel to signal that it no longer awaits responses
				return
			case responseData, ok := <-outChan:
				if !ok {
					return
				}
				responseChan <- responseData
			}
		}
	}(ctxJob)

	return responseChan, stopChan, nil
}
