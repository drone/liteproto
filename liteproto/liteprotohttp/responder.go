package liteprotohttp

import (
	"context"
	"net/http"
	"time"

	"github.com/drone/liteproto/liteproto"
	"github.com/drone/liteproto/liteproto/internal"
)

type responderFactory struct {
	client     liteproto.Client
	httpClient *http.Client
	url        string
	compress   bool
}

func NewResponderFactory(l liteproto.Client, client *http.Client, url string) internal.ResponderFactory {
	return &responderFactory{
		client:     l,
		httpClient: client,
		url:        url,
		compress:   false,
	}
}

func (r *responderFactory) MakeResponder(jobID, jobType string) liteproto.ResponderClient {
	return &responder{
		Client:     r.client,
		jobID:      jobID,
		jobType:    jobType,
		httpClient: r.httpClient,
		url:        r.url,
		compress:   r.compress,
	}
}

func (r *responderFactory) Client() liteproto.Client {
	return r.client
}

type responder struct {
	liteproto.Client
	jobID      string
	jobType    string
	httpClient *http.Client
	url        string
	compress   bool
}

func (r *responder) Respond(ctx context.Context, data []byte) (err error) {
	return call(ctx, r.httpClient, r.url, r.compress, r.jobID, r.jobType, data, time.Time{})
}

func (r *responder) RespondWithType(ctx context.Context, responseType string, data []byte) (err error) {
	return call(ctx, r.httpClient, r.url, r.compress, r.jobID, responseType, data, time.Time{})
}
