package nilrpc

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/security"
)

// Prefixes for domains.
const (
	MdsUserPrefix      = "MDS_USER"
	MdsClusterPrefix   = "MDS_CLUSTER"
	MdsObjectPrefix    = "MDS_OBJECT"
	MdsGencodingPrefix = "MDS_GENCODING"

	DsClusterPrefix = "DS_CLUSTER"
)

// MethodName indicates what procedure will be called.
type MethodName int

const (
	// MDS user domain methods.
	MdsUserAddUser MethodName = iota
	MdsUserMakeBucket
	MdsUserGetCredential

	// MDS cluster domain methods.
	MdsClusterGetClusterMap
	MdsClusterGetUpdateNoti
	MdsClusterLocalJoin
	MdsClusterGlobalJoin
	MdsClusterRegisterVolume
	MdsClusterListJob

	// MDS object domain methods.
	MdsObjectPut
	MdsObjectGet

	// MDS global encoding domain methods
	MdsGencodingGGG

	// DS cluster domain methods.
	DsClusterAddVolume
)

func (m MethodName) String() string {
	switch m {
	case MdsUserAddUser:
		return MdsUserPrefix + "." + "AddUser"
	case MdsUserMakeBucket:
		return MdsUserPrefix + "." + "MakeBucket"
	case MdsUserGetCredential:
		return MdsUserPrefix + "." + "GetCredential"

	case MdsClusterGetClusterMap:
		return MdsClusterPrefix + "." + "GetClusterMap"
	case MdsClusterGetUpdateNoti:
		return MdsClusterPrefix + "." + "GetUpdateNoti"
	case MdsClusterLocalJoin:
		return MdsClusterPrefix + "." + "LocalJoin"
	case MdsClusterGlobalJoin:
		return MdsClusterPrefix + "." + "GlobalJoin"
	case MdsClusterRegisterVolume:
		return MdsClusterPrefix + "." + "RegisterVolume"
	case MdsClusterListJob:
		return MdsClusterPrefix + "." + "ListJob"

	case MdsObjectPut:
		return MdsObjectPrefix + "." + "Put"
	case MdsObjectGet:
		return MdsObjectPrefix + "." + "Get"

	case MdsGencodingGGG:
		return MdsGencodingPrefix + "." + "GGG"

	case DsClusterAddVolume:
		return DsClusterPrefix + "." + "AddVolume"
	default:
		return "unknown"
	}
}

// RPCType is the first byte of connection and it implies the type of the RPC.
type RPCType byte

const (
	// RPCRaft used when raft connection.
	RPCRaft RPCType = 0x01
	// RPCNil used when nil admin connection.
	RPCNil = 0x02
	// RPCSwim used when swim membership connection.
	RPCSwim = 0x03
)

// Dial dials with the given rpc type connection to the address.
func Dial(addr string, rpcType RPCType, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}

	config := security.DefaultTLSConfig()

	conn, err := tls.DialWithDialer(dialer, "tcp", addr, config)
	if err != nil {
		return nil, err
	}

	// Write RPC header.
	_, err = conn.Write([]byte{
		byte(rpcType),
	})
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, err
}
