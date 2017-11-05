package uuid

import (
	"crypto/rand"
	"fmt"
)

// UUID is using for ID of each node
type UUID string

// Gen generates UUID.
func Gen() UUID {
	buf := make([]byte, 8)
	rand.Read(buf)

	return UUID(fmt.Sprintf("%x-%x", buf[0:4], buf[4:]))
}

func (u *UUID) String() string {
	return string(*u)
}
