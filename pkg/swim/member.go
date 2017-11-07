package swim

import (
	"github.com/chanyoung/nil/pkg/swim/swimpb"
)

func newMember(id, ip, port string, status swimpb.Status, incarnation uint32) *swimpb.Member {
	return &swimpb.Member{
		Uuid:        id,
		Addr:        ip,
		Port:        port,
		Status:      status,
		Incarnation: incarnation,
	}
}
