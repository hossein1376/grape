package reqid

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"time"
)

type ReqID string

const RequestIDKey ReqID = "request_id"

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
	rand.Read(src[8:])
	for i := range src {
		src[i] = base32alphabet[src[i]%32]
	}
	return string(src)
}
