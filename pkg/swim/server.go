package swim

// Server has functions
type Server struct {
	id   string
	meml *memList

	q []*Member
}

// NewServer creates swim server object.
func NewServer(id string, addr, port string) *Server {
	me := newMember(id, addr, port, Status_ALIVE, 0)
	memList := newMemList()

	// memList.set(me)
	memList.list[me.Uuid] = me

	return &Server{
		id:   id,
		meml: memList,
		q:    make([]*Member, 0),
	}
}
