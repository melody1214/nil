package nilmux

import (
	"net"
	"sync"
	"time"
)

type nilConn struct {
	conn     net.Conn
	once     sync.Once
	signByte byte
}

func newNilConn(conn net.Conn, signByte byte) *nilConn {
	return &nilConn{
		conn:     conn,
		signByte: signByte,
	}
}

func (nc *nilConn) Read(b []byte) (n int, err error) {
	nc.once.Do(func() {
		if len(b) < 1 {
			return
		}

		b[0] = nc.signByte
		b = b[1:]
		n++
	})
	read, err := nc.conn.Read(b)
	return read + n, err
}

func (nc *nilConn) Write(b []byte) (n int, err error) {
	return nc.conn.Write(b)
}

func (nc *nilConn) Close() error {
	return nc.conn.Close()
}

func (nc *nilConn) LocalAddr() net.Addr {
	return nc.conn.LocalAddr()
}

func (nc *nilConn) RemoteAddr() net.Addr {
	return nc.conn.RemoteAddr()
}

func (nc *nilConn) SetDeadline(t time.Time) error {
	return nc.conn.SetDeadline(t)
}

func (nc *nilConn) SetReadDeadline(t time.Time) error {
	return nc.conn.SetReadDeadline(t)
}

func (nc *nilConn) SetWriteDeadline(t time.Time) error {
	return nc.conn.SetWriteDeadline(t)
}
