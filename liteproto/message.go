package liteproto

const (
	StatusSuccess = "success"
	StatusOK      = "ok"
	StatusError   = "error"
)

// TaskRequest contains information about a task that needs to executed.
type TaskRequest struct {
	// ID is an id. Response message uses use the same ID as the request to which it's a response to.
	ID string

	// Type describes type of a task.
	Type string

	// Data holds arbitrary byte data payload.
	Data []byte
}

// TaskResponse contains response of a task.
type TaskResponse struct {
	// ID is an id.
	ID string

	// Type describes type of a task.
	Type string

	// Status holds a status of a task ("success" or "error")
	Status string

	// Data holds arbitrary byte data payload.
	Data []byte
}
