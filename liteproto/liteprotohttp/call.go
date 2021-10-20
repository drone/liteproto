package liteprotohttp

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/drone/liteproto/liteproto"
)

var callBuffPool = &sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(nil)
	},
}

func call(ctx context.Context, client *http.Client, url string, compress bool,
	messageID, messageType string, messageData []byte, messageDeadline time.Time) (err error) {

	buf := callBuffPool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.Grow(512)
	defer callBuffPool.Put(buf)

	err = marshalMessage(buf, compress, messageID, messageType, messageData, &messageDeadline)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Content-Length", strconv.Itoa(buf.Len()))
	if compress {
		req.Header.Set("Content-Encoding", "gzip")
	}

	resp, err := client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		err = liteproto.ErrCallFailed
		return
	}

	return
}
