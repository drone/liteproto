package liteprotohttp

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/drone/liteproto/liteproto"
	"github.com/drone/liteproto/liteproto/internal"
)

func handler(f internal.Feeder, respPub internal.ResponsePub) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		// select unmarshaler based on the Content-Type header

		var unmarshaler messageMarshaller
		contentType := r.Header.Get("Content-Type")

		if contentType == "" || strings.HasPrefix(contentType, "application/json") {
			unmarshaler = jsoner{}
			// TODO Implement support for other content types.
			//} else if strings.HasPrefix(contentType, "application/octet-stream") {
			//	unmarshaler = nil
		} else {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}

		// handle compressed request body based on the Content-Encoding header

		var rc io.ReadCloser
		encoding := r.Header.Get("Content-Encoding")

		compressed := encoding == "gzip"
		if compressed {
			rc, err = gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip header", http.StatusBadRequest)
				return
			}
		} else {
			rc = io.NopCloser(r.Body)
		}

		// read message

		m, err := wrapReader(rc, func(reader io.Reader) (*message, error) {
			return unmarshaler.messageUnmarshal(reader)
		})
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if m.ID == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		if m.Type == "" {
			http.Error(w, "missing type", http.StatusBadRequest)
			return
		}

		// process either task request or task response...
		// field Status tells us if it's one or the other.

		if isRequest := m.Status == ""; isRequest {
			var deadline time.Time
			if m.Deadline != nil {
				deadline = *m.Deadline
			}

			err = f.Feed(context.Background(), liteproto.TaskRequest{ID: m.ID, Type: m.Type, Data: m.Data}, deadline)
			if err == liteproto.ErrUnknownType {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			err = respPub.Publish(liteproto.TaskResponse{ID: m.ID, Type: m.Type, Status: m.Status, Data: m.Data})
		}
		if err != nil {
			// TODO: Log the error
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}
