package liteprotohttp

import (
	"net/http"

	"github.com/drone/liteproto/liteproto"
	"github.com/drone/liteproto/liteproto/internal"
)

func handlerResponse(respPub internal.ResponsePub) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		compressed := r.Header.Get("Content-Encoding") == "gzip"

		request, err := unmarshalMessage(r.Body, compressed)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		responseID := request.ID
		responseType := request.Type
		data := request.Data

		if responseID == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		if responseType == "" {
			http.Error(w, "missing type", http.StatusBadRequest)
			return
		}

		err = respPub.Publish(liteproto.Message{
			MessageID:   responseID,
			MessageType: responseType,
			MessageData: data,
		})
		if err != nil {
			// TODO: Log the error
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}
