package internal

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"github.com/drone/liteproto/liteproto"
)

type ServerFeeder struct {
	execerMap        map[string]interface{}
	execerDefault    liteproto.ExecerWithResponder
	responderFactory ResponderFactory
	logger           *log.Logger
}

func NewServerFeeder(factory ResponderFactory, logger *log.Logger) (sf *ServerFeeder) {
	return &ServerFeeder{
		execerMap:        map[string]interface{}{},
		responderFactory: factory,
		logger:           logger,
	}
}

func (sf *ServerFeeder) Register(jobType string, execer liteproto.Execer) {
	sf.execerMap[jobType] = execer
}

func (sf *ServerFeeder) RegisterWithResponder(jobType string, execer liteproto.ExecerWithResponder) {
	sf.execerMap[jobType] = execer
}

func (sf *ServerFeeder) RegisterCatchAll(execer liteproto.ExecerWithResponder) {
	sf.execerDefault = execer
}

func (sf *ServerFeeder) Feed(ctx context.Context, jobID, jobType string, data []byte, deadline time.Time) error {
	execer, ok := sf.execerMap[jobType]
	if !ok {
		if sf.execerDefault == nil {
			return liteproto.ErrUnknownType
		} else {
			execer = sf.execerDefault
		}
	}

	var (
		ctxJob     context.Context
		cancelFunc func()
	)

	if !deadline.IsZero() {
		now := time.Now()
		if deadline.Before(now) {
			return context.DeadlineExceeded
		}

		ctxJob, cancelFunc = context.WithDeadline(ctx, deadline)
	} else {
		ctxJob = ctx
		cancelFunc = func() {}
	}

	job := liteproto.Message{
		MessageID:   jobID,
		MessageType: jobType,
		MessageData: data,
	}

	switch execer := execer.(type) {
	case liteproto.Execer:
		go func(ctx context.Context, job *liteproto.Message) {
			defer sf.panicRecovery(cancelFunc)

			client := sf.responderFactory.Client()
			execer.Exec(ctx, *job, client)
		}(ctxJob, &job)

	case liteproto.ExecerWithResponder:
		go func(ctx context.Context, job *liteproto.Message) {
			defer sf.panicRecovery(cancelFunc)

			responder := sf.responderFactory.MakeResponder(jobID, jobType)
			execer.Exec(ctx, *job, responder)
		}(ctxJob, &job)
	default:
		cancelFunc()
	}

	return nil
}

func (sf *ServerFeeder) panicRecovery(cancel func()) {
	if r := recover(); r != nil && sf.logger != nil {
		sf.logger.Printf("PANIC: %v\n%s", r, debug.Stack())
	}

	cancel()
}
