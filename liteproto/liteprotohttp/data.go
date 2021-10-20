package liteprotohttp

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"time"
)

// message is used to form request body for all HTTP requests.
type message struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Data     json.RawMessage `json:"data"`
	Deadline *time.Time      `json:"deadline,omitempty"`
}

func marshalMessage(writer io.Writer, compress bool, messageID, messageType string, messageData []byte, messageDeadline *time.Time) error {
	if messageDeadline != nil && messageDeadline.IsZero() {
		messageDeadline = nil
	}

	var w io.Writer

	if compress {
		gz := gzip.NewWriter(writer)
		defer gz.Close()

		w = gz
	} else {
		w = writer
	}

	return json.NewEncoder(w).Encode(&message{
		ID:       messageID,
		Type:     messageType,
		Data:     messageData,
		Deadline: messageDeadline,
	})
}

func unmarshalMessage(reader io.Reader, compress bool) (*message, error) {
	var r io.Reader

	if compress {
		gz, err := gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
		defer gz.Close()

		r = gz
	} else {
		r = reader
	}

	value := &message{}

	err := json.NewDecoder(r).Decode(value)
	if err != nil {
		return nil, err
	}

	return value, nil
}
