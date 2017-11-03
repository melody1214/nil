package swim

import (
	"net"

	"github.com/chanyoung/nil/pkg/util/uuid"
)

type member struct {
	id uuid.UUID
	ip net.IP
}

func newMember() *member {
	return &member{}
}
