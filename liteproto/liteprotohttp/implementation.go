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

func New(url string, compress bool, httpClient *http.Client, logger *log.Logger) *CallProto {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	if logger == nil {
		logger = log.Default()
	}

	h := &CallProto{}

	marshaller := jsoner{}

	c := newCaller(httpClient, marshaller, url, compress)
	f := newResponderFactory(h, c)

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

func (h *CallProto) Register(t string, execer liteproto.Execer) {
	h.sf.Register(t, execer)
}

func (h *CallProto) RegisterWithResponder(t string, execer liteproto.ExecerWithResponder) {
	h.sf.RegisterWithResponder(t, execer)
}

func (h *CallProto) RegisterCatchAll(execer liteproto.ExecerWithResponder) {
	h.sf.RegisterCatchAll(execer)
}

func (h *CallProto) CallWithResponse(ctx context.Context, r liteproto.TaskRequest) (response <-chan liteproto.TaskResponse, stop chan<- struct{}, err error) {
	return h.runner.Run(ctx, r, time.Time{})
}

func (h *CallProto) CallWithDeadline(ctx context.Context, r liteproto.TaskRequest, deadline time.Time) (response <-chan liteproto.TaskResponse, stop chan<- struct{}, err error) {
	return h.runner.Run(ctx, r, deadline)
}

func (h *CallProto) Call(ctx context.Context, r liteproto.TaskRequest) (err error) {
	return h.caller.Call(ctx, r, time.Time{})
}

func (h *CallProto) Handler() http.Handler {
	return handler(h.sf, h.pubsub)
}
