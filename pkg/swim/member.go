package swim

import (
	"github.com/chanyoung/nil/pkg/util/uuid"
)

const (
	// ALIVE is the status of healthy member.
	ALIVE int32 = 0
	// SUSPECT is the status of faulty suspected member.
	SUSPECT int32 = 1
	// FAULTY is the status of real faulty member.
	FAULTY int32 = 2
)

type member struct {
	id          uuid.UUID
	ip          string
	port        string
	status      int32
	incarnation uint32
}

func newMember(id uuid.UUID, ip string, port string,
	status int32, incarnation uint32) *member {
	return &member{
		id:          id,
		ip:          ip,
		port:        port,
		status:      status,
		incarnation: incarnation,
	}
}
