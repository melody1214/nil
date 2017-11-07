package swim

import "github.com/chanyoung/nil/pkg/swim/swimpb"

// Server has functions
type Server struct {
	id   string
	meml *memList

	q []*swimpb.Member
}

// NewServer creates swim server object.
func NewServer(id string, addr, port string) *Server {
	me := newMember(id, addr, port, swimpb.Status_ALIVE, 0)
	memList := newMemList()

	memList.set(me)

	return &Server{
		id:   id,
		meml: memList,
		q:    make([]*swimpb.Member, 0),
	}
}

func (s *Server) makeQ() {
	s.q = s.q[:0]

	s.q = append(s.q, s.meml.getAll()...)
}
