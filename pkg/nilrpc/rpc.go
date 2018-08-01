package nilrpc

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/security"
)

// Prefixes for domains.
const (
	MdsAccountPrefix    = "MDS_ACCOUNT"
	MdsMembershipPrefix = "MDS_MEMBERSHIP"
	MdsObjectPrefix     = "MDS_OBJECT"
	MdsGencodingPrefix  = "MDS_GENCODING"

	DsClusterPrefix   = "DS_CLUSTER"
	DsGencodingPrefix = "DS_GENCODING"
	DsObjectPrefix    = "DS_OBJECT"
)

// MethodName indicates what procedure will be called.
type MethodName int

const (
	// MDS user domain methods.
	MdsAccountAddUser MethodName = iota
	MdsAccountMakeBucket
	MdsAccountGetCredential

	// MDS cluster domain methods.
	MdsMembershipGetClusterMap
	MdsMembershipGetUpdateNoti
	MdsMembershipLocalJoin
	MdsMembershipGlobalJoin
	MdsMembershipRegisterVolume

	// MDS object domain methods.
	MdsObjectPut
	MdsObjectGet
	MdsObjectGetChunk
	MdsObjectSetChunk

	// MDS global encoding domain methods
	MdsGencodingGGG
	MdsGencodingUpdateUnencodedChunk
	MdsGencodingSelectEncodingGroup
	MdsGencodingHandleToken
	MdsGencodingGetEncodingJob
	MdsGencodingSetJobStatus
	MdsGencodingJobFinished
	MdsGencodingSetPrimaryChunk

	// DS cluster domain methods.
	DsClusterAddVolume
	DsClusterRecoveryChunk

	// Ds gencoding domain methods.
	DsGencodingRenameChunk
	DsGencodingTruncateChunk
	DsGencodingEncode
	DsGencodingGetCandidateChunk

	DsObjectSetChunkPool
)

func (m MethodName) String() string {
	switch m {
	case MdsAccountAddUser:
		return MdsAccountPrefix + "." + "AddUser"
	case MdsAccountMakeBucket:
		return MdsAccountPrefix + "." + "MakeBucket"
	case MdsAccountGetCredential:
		return MdsAccountPrefix + "." + "GetCredential"

	case MdsMembershipGetClusterMap:
		return MdsMembershipPrefix + "." + "GetClusterMap"
	case MdsMembershipGetUpdateNoti:
		return MdsMembershipPrefix + "." + "GetUpdateNoti"
	case MdsMembershipLocalJoin:
		return MdsMembershipPrefix + "." + "LocalJoin"
	case MdsMembershipGlobalJoin:
		return MdsMembershipPrefix + "." + "GlobalJoin"
	case MdsMembershipRegisterVolume:
		return MdsMembershipPrefix + "." + "RegisterVolume"

	case MdsObjectPut:
		return MdsObjectPrefix + "." + "Put"
	case MdsObjectGet:
		return MdsObjectPrefix + "." + "Get"
	case MdsObjectGetChunk:
		return MdsObjectPrefix + "." + "GetChunk"
	case MdsObjectSetChunk:
		return MdsObjectPrefix + "." + "SetChunk"

	case MdsGencodingGGG:
		return MdsGencodingPrefix + "." + "GGG"
	case MdsGencodingUpdateUnencodedChunk:
		return MdsGencodingPrefix + "." + "UpdateUnencodedChunk"
	case MdsGencodingSelectEncodingGroup:
		return MdsGencodingPrefix + "." + "SelectEncodingGroup"
	case MdsGencodingHandleToken:
		return MdsGencodingPrefix + "." + "HandleToken"
	case MdsGencodingGetEncodingJob:
		return MdsGencodingPrefix + "." + "GetEncodingJob"
	case MdsGencodingSetJobStatus:
		return MdsGencodingPrefix + "." + "SetJobStatus"
	case MdsGencodingJobFinished:
		return MdsGencodingPrefix + "." + "JobFinished"
	case MdsGencodingSetPrimaryChunk:
		return MdsGencodingPrefix + "." + "SetPrimaryChunk"

	case DsClusterAddVolume:
		return DsClusterPrefix + "." + "AddVolume"
	case DsClusterRecoveryChunk:
		return DsClusterPrefix + "." + "RecoveryChunk"

	case DsGencodingRenameChunk:
		return DsGencodingPrefix + "." + "RenameChunk"
	case DsGencodingTruncateChunk:
		return DsGencodingPrefix + "." + "TruncateChunk"
	case DsGencodingEncode:
		return DsGencodingPrefix + "." + "Encode"
	case DsGencodingGetCandidateChunk:
		return DsGencodingPrefix + "." + "GetCandidateChunk"

	case DsObjectSetChunkPool:
		return DsObjectPrefix + "." + "SetChunkPool"

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
