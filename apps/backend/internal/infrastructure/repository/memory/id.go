package memory

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

// newID returns a UUIDv7: time-ordered like the IDs PostgreSQL 18 generates
// with uuidv7(), so listings sort the same way on both adapters.
func newID() string {
	var id [16]byte
	var timestamp [8]byte
	binary.BigEndian.PutUint64(timestamp[:], uint64(time.Now().UnixMilli()))
	copy(id[0:6], timestamp[2:8])
	_, _ = rand.Read(id[6:])
	id[6] = (id[6] & 0x0f) | 0x70
	id[8] = (id[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:16])
}
