package reqid

import (
	"context"
	"uuid"
)

type ReqID string

const RequestIDKey ReqID = "request_id"

// NewRequestID creates and returns a UUID v7.
func NewRequestID() ReqID {
	return ReqID(uuid.NewV7().String())
}

func RequestID(c context.Context) (string, bool) {
	id, ok := c.Value(RequestIDKey).(ReqID)
	return string(id), ok
}
