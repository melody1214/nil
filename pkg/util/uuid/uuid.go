package uuid

import (
	"crypto/rand"
	"fmt"
)

// Gen generates UUID.
func Gen() string {
	buf := make([]byte, 8)
	rand.Read(buf)

	return fmt.Sprintf("%x-%x", buf[0:4], buf[4:])
}
