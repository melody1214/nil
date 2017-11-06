package swim

func newMember(id, ip, port string, status Status, incarnation uint32) *Member {
	return &Member{
		Uuid:        id,
		Addr:        ip,
		Port:        port,
		Status:      status,
		Incarnation: incarnation,
	}
}
