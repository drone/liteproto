package liteprotohttp

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/drone/liteproto/liteproto"
	"github.com/drone/liteproto/liteproto/internal"
)

type CallProto struct {
	caller internal.Caller
	pubsub internal.ResponsePubSub
	runner *internal.Runner
	sf     *internal.ServerFeeder
}

func New(urlRequest, urlResponse string, compress bool, httpClient *http.Client, logger *log.Logger) *CallProto {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	if logger == nil {
		logger = log.Default()
	}

	h := &CallProto{}

	c := NewCaller(httpClient, urlRequest, compress)
	f := NewResponderFactory(h, httpClient, urlResponse)

	sf := internal.NewServerFeeder(f, logger)
	pubsub := &internal.PubSub{}
	runner := internal.NewRunner(c, pubsub)

	h.caller = c
	h.pubsub = pubsub
	h.runner = runner
	h.sf = sf

	// make sure it implements ServerClient interface
	var e liteproto.ServerClient = h
	_ = e

	return h
}

func (h *CallProto) Register(jobType string, execer liteproto.Execer) {
	h.sf.Register(jobType, execer)
}

func (h *CallProto) RegisterWithResponder(jobType string, execer liteproto.ExecerWithResponder) {
	h.sf.RegisterWithResponder(jobType, execer)
}

func (h *CallProto) RegisterCatchAll(execer liteproto.ExecerWithResponder) {
	h.sf.RegisterCatchAll(execer)
}

func (h *CallProto) CallWithResponse(ctx context.Context, job liteproto.Message) (response <-chan liteproto.Message, stop chan<- struct{}, err error) {
	return h.runner.Run(ctx, job.MessageID, job.MessageType, job.MessageData, time.Time{})
}

func (h *CallProto) CallWithDeadline(ctx context.Context, job liteproto.Message, deadline time.Time) (response <-chan liteproto.Message, stop chan<- struct{}, err error) {
	return h.runner.Run(ctx, job.MessageID, job.MessageType, job.MessageData, deadline)
}

func (h *CallProto) Call(ctx context.Context, job liteproto.Message) (err error) {
	return h.caller.Call(ctx, job.MessageID, job.MessageType, job.MessageData, time.Time{})
}

func (h *CallProto) HandlerRequest() http.Handler {
	return handlerRequest(h.sf)
}

func (h *CallProto) HandlerResponse() http.Handler {
	return handlerResponse(h.pubsub)
}
