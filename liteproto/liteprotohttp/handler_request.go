package liteprotohttp

import (
	"context"
	"net/http"
	"time"

	"github.com/drone/liteproto/liteproto"
	"github.com/drone/liteproto/liteproto/internal"
)

func handlerRequest(f internal.Feeder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		compressed := r.Header.Get("Content-Encoding") == "gzip"

		request, err := unmarshalMessage(r.Body, compressed)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		jobID := request.ID
		jobType := request.Type
		data := request.Data

		if jobID == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		if jobType == "" {
			http.Error(w, "missing type", http.StatusBadRequest)
			return
		}

		var deadline time.Time
		if request.Deadline != nil {
			deadline = *request.Deadline
		}

		err = f.Feed(context.Background(), jobID, jobType, data, deadline)
		if err == liteproto.ErrUnknownType {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}
