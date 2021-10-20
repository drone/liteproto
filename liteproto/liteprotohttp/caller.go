package liteprotohttp

import (
	"context"
	"net/http"
	"time"

	"github.com/drone/liteproto/liteproto/internal"
)

// NewCaller creates a new Caller that calls a remote server with using HTTP/HTTPS protocol.
func NewCaller(client *http.Client, url string, compressed bool) internal.Caller {
	return &caller{
		client:     client,
		url:        url,
		compressed: compressed,
	}
}

type caller struct {
	client     *http.Client
	url        string
	compressed bool
}

func (c *caller) Call(ctx context.Context, jobID, jobType string, data []byte, deadline time.Time) error {
	return call(ctx, c.client, c.url, c.compressed, jobID, jobType, data, deadline)
}
