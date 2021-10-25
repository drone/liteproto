package liteprotohttp

import (
	"context"

	"github.com/drone/liteproto/liteproto"
	"github.com/drone/liteproto/liteproto/internal"
)

type responderFactory struct {
	client liteproto.Client
	caller *caller
}

func newResponderFactory(cl liteproto.Client, c *caller) internal.ResponderFactory {
	return &responderFactory{
		client: cl,
		caller: c,
	}
}

func (r *responderFactory) MakeResponder(id, t string) liteproto.ResponderClient {
	return &responder{
		Client:  r.client,
		caller:  r.caller,
		id:      id,
		defType: t,
	}
}

func (r *responderFactory) Client() liteproto.Client {
	return r.client
}

type responder struct {
	liteproto.Client
	caller  *caller
	id      string
	defType string
}

func (r *responder) Respond(ctx context.Context, status string, data []byte) (err error) {
	if status == "" {
		panic("status can't be empty")
	}
	return r.caller.do(ctx, r.id, r.defType, status, data, nil)
}

func (r *responder) RespondWithType(ctx context.Context, responseType, status string, data []byte) (err error) {
	if status == "" {
		panic("status can't be empty")
	}
	return r.caller.do(ctx, r.id, responseType, status, data, nil)
}
