package reqid

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"time"
)

type ReqID string

const RequestIDKey ReqID = "request_id"

// NewRequestID creates and return a 24 bits random string, encoded as base32.
// The first 8 bits are filled based on the current UNIX time, while the other
// 16 bits are randomly generated.
func NewRequestID() ReqID {
	return ReqID(text())
}

func RequestID(c context.Context) (string, bool) {
	id, ok := c.Value(RequestIDKey).(ReqID)
	return string(id), ok
}

const base32alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

func text() string {
	src := make([]byte, 24)
	binary.BigEndian.PutUint64(src[:8], uint64(time.Now().UnixNano()))
	_, _ = rand.Read(src[8:])
	for i := range src {
		src[i] = base32alphabet[src[i]%32]
	}
	return string(src)
}
