package liteprotohttp

import (
	"encoding/json"
	"io"
	"time"
)

// message is used to form request body for all HTTP requests.
type message struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Status   string          `json:"status,omitempty"` // status is used only for response messages
	Data     json.RawMessage `json:"data"`
	Deadline *time.Time      `json:"deadline,omitempty"` // deadline is used only for request messages
}

type nopCloseWriter struct {
	io.Writer
}

func (w nopCloseWriter) Write(data []byte) (int, error) {
	return w.Writer.Write(data)
}

func (nopCloseWriter) Close() error {
	return nil
}

func wrapWriter(wc io.WriteCloser, f func(writer io.Writer) error) error {
	if err := f(wc); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return nil
}

func wrapReader(rc io.ReadCloser, f func(reader io.Reader) (*message, error)) (m *message, err error) {
	if m, err = f(rc); err != nil {
		return
	}
	if err = rc.Close(); err != nil {
		return nil, err
	}
	return
}

type messageMarshaller interface {
	messageMarshal(writer io.Writer, m *message) error
	messageUnmarshal(reader io.Reader) (*message, error)
}

type jsoner struct{}

func (jsoner) messageMarshal(w io.Writer, m *message) error {
	return json.NewEncoder(w).Encode(m)
}

func (jsoner) messageUnmarshal(r io.Reader) (*message, error) {
	var m message

	err := json.NewDecoder(r).Decode(&m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}
