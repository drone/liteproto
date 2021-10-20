package liteproto

// Message contains information about a job that needs to executed or a response to such a job.
type Message struct {
	// MessageID is an ID. Response message uses use the same ID as the job to which it's a response to.
	MessageID string

	// MessageType describes type of the MessageData.
	MessageType string

	// MessageData hold arbitrary byte data payload.
	MessageData []byte
}
