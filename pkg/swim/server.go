package swim

import "github.com/chanyoung/nil/pkg/util/uuid"

// Server has functions
type Server struct {
	id   uuid.UUID
	meml *memList
}

// NewServer creates swim server object.
func NewServer(id uuid.UUID, addr, port string) *Server {
	me := newMember(id, addr, port, ALIVE, 0)
	memList := newMemList()

	memList.set(me)

	return &Server{
		id:   id,
		meml: memList,
	}
}
