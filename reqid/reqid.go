package reqid

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	mathrand "math/rand"
	"time"
)

type ReqID string

const RequestIDKey ReqID = "request_id"

// NewRequestID creates and returns a 26-character base32 (RFC4648, no padding)
// string representing 128 bits composed of:
// - 48-bit big-endian timestamp (milliseconds since epoch) -> 6 bytes
// - 80-bit cryptographically secure random bytes -> 10 bytes
//
// The resulting 16 bytes are base32-encoded without padding, producing 26 chars.
// This layout keeps IDs roughly sortable by creation time while maintaining
// strong randomness. If the crypto RNG fails, a time-seeded math/rand fallback
// is used to ensure an ID is still produced.
func NewRequestID() ReqID {
	return ReqID(text())
}

func RequestID(c context.Context) (string, bool) {
	id, ok := c.Value(RequestIDKey).(ReqID)
	return string(id), ok
}

const crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

var base32NoPad = base32.NewEncoding(crockfordAlphabet).WithPadding(base32.NoPadding)

func text() string {
	// 16 bytes total: 6 bytes timestamp + 10 bytes randomness
	var buf [16]byte

	// 48-bit timestamp in milliseconds (big-endian)
	ts := uint64(time.Now().UnixMilli())
	var tmp [8]byte
	binary.BigEndian.PutUint64(tmp[:], ts)
	// copy the lower 6 bytes of the 8-byte big-endian representation so that
	// we store a 48-bit timestamp in the first 6 bytes.
	copy(buf[0:6], tmp[2:8])

	// fill remaining 10 bytes with crypto random
	if _, err := rand.Read(buf[6:]); err != nil {
		// fallback to math/rand seeded with current time if crypto fails
		seeded := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
		for i := 6; i < len(buf); i++ {
			buf[i] = byte(seeded.Intn(256))
		}
	}

	// base32 encode without padding -> 26 characters for 16 bytes
	// Use a Crockford-like alphabet with digits first so lexical order of the
	// encoded string matches the byte order (timestamp first) and thus is
	// sortable by creation time.
	return base32NoPad.EncodeToString(buf[:])
}
