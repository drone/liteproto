package liteprotohttp

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/drone/liteproto/liteproto"
)

// newCaller creates a new Caller that calls a remote server with using HTTP/HTTPS protocol.
func newCaller(client *http.Client, marshaller messageMarshaller, url string, compress bool) *caller {
	return &caller{
		client:     client,
		marshaller: marshaller,
		url:        url,
		compress:   compress,
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(nil)
			},
		},
	}
}

type caller struct {
	client     *http.Client
	marshaller messageMarshaller
	url        string
	compress   bool
	bufferPool *sync.Pool
}

func (c *caller) Call(ctx context.Context, r liteproto.TaskRequest, deadline time.Time) error {
	return c.do(ctx, r.ID, r.Type, "", r.Data, &deadline)
}

func (c *caller) do(ctx context.Context, id, t, status string, data []byte, deadline *time.Time) (err error) {

	buf := c.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.Grow(512)
	defer c.bufferPool.Put(buf)

	var wc io.WriteCloser
	if c.compress {
		wc = gzip.NewWriter(buf)
	} else {
		wc = nopCloseWriter{buf}
	}

	if deadline != nil && deadline.IsZero() {
		deadline = nil
	}
	m := &message{
		ID:       id,
		Type:     t,
		Status:   status,
		Data:     data,
		Deadline: deadline,
	}

	err = wrapWriter(wc, func(writer io.Writer) error {
		return c.marshaller.messageMarshal(writer, m)
	})
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, buf)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Content-Length", strconv.Itoa(buf.Len()))
	if c.compress {
		req.Header.Set("Content-Encoding", "gzip")
	}

	resp, err := c.client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {

		var body []byte

		if resp.StatusCode >= http.StatusBadRequest {
			body, _ = io.ReadAll(resp.Body)
		}

		err = CallFailedError{
			StatusCode: resp.StatusCode,
			Body:       body,
		}
		return
	}

	return
}

// CallFailedError is returned by Caller when a remote server returns an error.
type CallFailedError struct {
	StatusCode int
	Body       []byte
}

func (e CallFailedError) Error() string {
	if e.Body == nil {
		return fmt.Sprintf("call failed, status=%d", e.StatusCode)
	}

	return fmt.Sprintf("call failed, status=%d body=%s", e.StatusCode, e.Body)
}
