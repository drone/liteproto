package liteprotohttp

import (
	"bytes"
	"compress/gzip"
	"testing"
	"time"
)

func TestMarshalMessage(t *testing.T) {
	tests := []struct {
		messageID       string
		messageType     string
		messageData     []byte
		messageDeadline time.Time
		compress        bool
		expectedJSON    []byte
	}{
		{
			messageID:    "abc",
			messageType:  "simple-test",
			messageData:  []byte("42"),
			compress:     false,
			expectedJSON: []byte(`{"id":"abc","type":"simple-test","data":42}`),
		},
		{
			messageID:    "xyz",
			messageType:  "compress-test",
			messageData:  []byte("66"),
			compress:     true,
			expectedJSON: []byte(`{"id":"xyz","type":"compress-test","data":66}`),
		},
		{
			messageID:       "123",
			messageType:     "deadline-test",
			messageData:     []byte("7"),
			messageDeadline: time.Date(2030, 6, 15, 0, 0, 0, 0, time.UTC),
			compress:        true,
			expectedJSON:    []byte(`{"id":"123","type":"deadline-test","data":7,"deadline":"2030-06-15T00:00:00Z"}`),
		},
	}

	for i := 0; i < len(tests); i++ {
		tests[i].expectedJSON = append(tests[i].expectedJSON, '\n')

		if !tests[i].compress {
			continue
		}

		buf := &bytes.Buffer{}
		gzw := gzip.NewWriter(buf)
		_, _ = gzw.Write(tests[i].expectedJSON)
		_ = gzw.Close()

		tests[i].expectedJSON = buf.Bytes()
	}

	for _, test := range tests {
		buf := &bytes.Buffer{}
		_ = marshalMessage(buf, test.compress, test.messageID, test.messageType, test.messageData, &test.messageDeadline)

		if !bytes.Equal(buf.Bytes(), test.expectedJSON) {
			t.Errorf("test %q failed. expected=`%s`, but got=`%s`", test.messageType, test.expectedJSON, buf.Bytes())
			continue
		}
	}
}

func TestUnmarshalMessage(t *testing.T) {
	tests := []struct {
		data        []byte
		compress    bool
		expectID    string
		expectType  string
		expectData  []byte
		expectError bool
	}{
		{
			data:       []byte(`{"id":"abc","type":"simple-test","data":42}`),
			compress:   false,
			expectID:   "abc",
			expectType: "simple-test",
			expectData: []byte("42"),
		},
		{
			data:       []byte(`{"id":"xyz","type":"compress-test","data":66}`),
			compress:   true,
			expectID:   "xyz",
			expectType: "compress-test",
			expectData: []byte("66"),
		},
		{
			data:        []byte(`{invalid_json}`),
			compress:    false,
			expectType:  "error-test",
			expectError: true,
		},
	}

	for i := 0; i < len(tests); i++ {
		if !tests[i].compress {
			continue
		}

		buf := &bytes.Buffer{}
		gzw := gzip.NewWriter(buf)
		_, _ = gzw.Write(tests[i].data)
		_ = gzw.Close()

		tests[i].data = buf.Bytes()
	}

	for _, test := range tests {
		m, err := unmarshalMessage(bytes.NewReader(test.data), test.compress)
		if err != nil && !test.expectError || err == nil && test.expectError {
			t.Errorf("test %q failed. expected error=%t, but got %v", test.expectType, test.expectError, err)
			continue
		}

		if test.expectError {
			continue
		}

		if m.ID != test.expectID {
			t.Errorf("test %q failed. expected ID=%s, but got ID=%s", test.expectType, test.expectID, m.ID)
			continue
		}
		if m.Type != test.expectType {
			t.Errorf("test %q failed. expected Type=%s, but got Type=%s", test.expectType, test.expectType, m.Type)
			continue
		}
		if !bytes.Equal(m.Data, test.expectData) {
			t.Errorf("test %q failed. expected ID=%s, but got ID=%s", test.expectType, test.expectData, m.Data)
			continue
		}
	}
}
