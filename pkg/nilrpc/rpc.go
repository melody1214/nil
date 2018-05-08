package nilrpc

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/security"
)

// Prefixes for domains.
const (
	MdsAdminPrefix      = "MDS_ADMIN"
	MdsAuthPrefix       = "MDS_AUTH"
	MdsBucketPrefix     = "MDS_BUCKET"
	MdsClusterPrefix    = "MDS_CLUSTER"
	MdsMembershipPrefix = "MDS_MEMBERSHIP"
	MdsObjectPrefix     = "MDS_OBJECT"
	MdsRecoveryPrefix   = "MDS_RECOVERY"

	DSRPCPrefix = "DS"
)

// MethodName indicates what procedure will be called.
type MethodName int

const (
	// MDS admin domain methods.
	MdsAdminJoin MethodName = iota
	MdsAdminAddUser
	MdsAdminRegisterVolume

	// MDS auth domain methods.
	MdsAuthGetCredential

	// MDS bucket domain methods.
	MdsBucketMakeBucket

	// MDS clustermap domain methods.
	MdsClustermapGetClusterMap
	MdsClustermapGetUpdateNoti
	MdsClustermapUpdateClusterMap
	MdsClustermapJoin

	// MDS membership domain methods.
	MdsMembershipGetMembershipList

	// MDS object domain methods.
	MdsObjectPut
	MdsObjectGet

	// MDS recovery domain methods
	MdsRecoveryRecovery

	// DS methods.
	AddVolume
)

func (m MethodName) String() string {
	switch m {
	case MdsAdminJoin:
		return MdsAdminPrefix + "." + "Join"
	case MdsAdminAddUser:
		return MdsAdminPrefix + "." + "AddUser"
	case MdsAdminRegisterVolume:
		return MdsAdminPrefix + "." + "RegisterVolume"

	case MdsAuthGetCredential:
		return MdsAuthPrefix + "." + "GetCredential"

	case MdsBucketMakeBucket:
		return MdsBucketPrefix + "." + "MakeBucket"

	case MdsClustermapGetClusterMap:
		return MdsClusterPrefix + "." + "GetClusterMap"
	case MdsClustermapGetUpdateNoti:
		return MdsClusterPrefix + "." + "GetUpdateNoti"
	case MdsClustermapUpdateClusterMap:
		return MdsClusterPrefix + "." + "UpdateClusterMap"
	case MdsClustermapJoin:
		return MdsClusterPrefix + "." + "Join"

	case MdsMembershipGetMembershipList:
		return MdsMembershipPrefix + "." + "GetMembershipList"

	case MdsObjectPut:
		return MdsObjectPrefix + "." + "Put"
	case MdsObjectGet:
		return MdsObjectPrefix + "." + "Get"

	case MdsRecoveryRecovery:
		return MdsRecoveryPrefix + "." + "Recovery"

	case AddVolume:
		return DSRPCPrefix + "." + "AddVolume"
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
