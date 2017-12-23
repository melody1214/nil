package server

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/hashicorp/raft"
)

type raftTransportLayer struct {
	*nilmux.Layer
}

func newRaftTransportLayer(l *nilmux.Layer) *raftTransportLayer {
	return &raftTransportLayer{Layer: l}
}

func (l *raftTransportLayer) Dial(addr raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}

	config := security.DefaultTLSConfig()

	conn, err := tls.DialWithDialer(dialer, "tcp", string(addr), config)
	if err != nil {
		return nil, err
	}

	// Write RPC header.
	_, err = conn.Write([]byte{
		0x01, // rpcRaft
	})
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, err
}
