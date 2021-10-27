package liteprotohttp

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/drone/liteproto/liteproto"
	"github.com/drone/liteproto/liteproto/internal"
)

// ServerClient can function as a server and as a client. Requests and responses to a remote server
// are sent with HTTP protocol. Message payload 'Data' must be JSON encoded (json.RawMessage).
type ServerClient struct {
	caller internal.Caller
	pubsub internal.ResponsePubSub
	runner *internal.Runner
	sf     *internal.ServerFeeder
}

// New creates a new ServerClient. Parameter 'url' is a full URL to which client calls and server responses
// will be directed. If bool parameter 'compress' is true, all HTTP request bodies will be gzipped and
// "Content-Encoding: gzip" header will be added. The library automatically handles gzipped HTTP requests.
// Parameters 'httpClient' and 'logger' can be nil. Default implementations will be used for those in that case.
// Logger is used only for logging panics that occur during execution of tasks.
func New(url string, compress bool, httpClient *http.Client, logger *log.Logger) *ServerClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	if logger == nil {
		logger = log.Default()
	}

	h := &ServerClient{}

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

func (h *ServerClient) Register(t string, execer liteproto.Execer) {
	h.sf.Register(t, execer)
}

func (h *ServerClient) RegisterWithResponder(t string, execer liteproto.ExecerWithResponder) {
	h.sf.RegisterWithResponder(t, execer)
}

func (h *ServerClient) RegisterCatchAll(execer liteproto.ExecerWithResponder) {
	h.sf.RegisterCatchAll(execer)
}

func (h *ServerClient) CallWithResponse(ctx context.Context, r liteproto.TaskRequest) (response <-chan liteproto.TaskResponse, stop chan<- struct{}, err error) {
	return h.runner.Run(ctx, r, time.Time{})
}

func (h *ServerClient) CallWithDeadline(ctx context.Context, r liteproto.TaskRequest, deadline time.Time) (response <-chan liteproto.TaskResponse, stop chan<- struct{}, err error) {
	return h.runner.Run(ctx, r, deadline)
}

func (h *ServerClient) Call(ctx context.Context, r liteproto.TaskRequest) (err error) {
	return h.caller.Call(ctx, r, time.Time{})
}

func (h *ServerClient) Handler() http.Handler {
	return handler(h.sf, h.pubsub)
}
