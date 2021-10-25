package internal

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"github.com/drone/liteproto/liteproto"
)

// ServerFeeder is a helper object that handles requests for task execution
// and relays them to one of the registered task executors.
type ServerFeeder struct {
	execerMap        map[string]interface{}
	execerDefault    liteproto.ExecerWithResponder
	responderFactory ResponderFactory
	logger           *log.Logger
}

// NewServerFeeder creates new ServerFeeder objects.
func NewServerFeeder(factory ResponderFactory, logger *log.Logger) (sf *ServerFeeder) {
	return &ServerFeeder{
		execerMap:        map[string]interface{}{},
		responderFactory: factory,
		logger:           logger,
	}
}

// Register assigns an liteproto.Execer to run tasks of a given type.
// This method is a part of liteproto.Server interface implementation.
func (sf *ServerFeeder) Register(typ string, execer liteproto.Execer) {
	sf.execerMap[typ] = execer
}

// RegisterWithResponder assigns an liteproto.ExecerWithResponder to run tasks of a given type.
// This method is a part of liteproto.Server interface implementation.
func (sf *ServerFeeder) RegisterWithResponder(typ string, execer liteproto.ExecerWithResponder) {
	sf.execerMap[typ] = execer
}

// RegisterCatchAll assigns an liteproto.ExecerWithResponder to run all tasks
// that are not already assigned to some other Execer.
// This method is a part of liteproto.Server interface implementation.
func (sf *ServerFeeder) RegisterCatchAll(execer liteproto.ExecerWithResponder) {
	sf.execerDefault = execer
}

// Feed accepts requests for task execution. Parameter deadline should be zero time if it's not needed.
// This method implements Feeder interface.
func (sf *ServerFeeder) Feed(ctx context.Context, r liteproto.TaskRequest, deadline time.Time) error {
	execer, ok := sf.execerMap[r.Type]
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

	switch execer := execer.(type) {
	case liteproto.Execer:
		go func(ctx context.Context, r *liteproto.TaskRequest) {
			defer sf.panicRecovery(cancelFunc)

			client := sf.responderFactory.Client()
			execer.Exec(ctx, *r, client)
		}(ctxJob, &r)

	case liteproto.ExecerWithResponder:
		go func(ctx context.Context, r *liteproto.TaskRequest) {
			defer sf.panicRecovery(cancelFunc)

			responder := sf.responderFactory.MakeResponder(r.ID, r.Type)
			execer.Exec(ctx, *r, responder)
		}(ctxJob, &r)
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
