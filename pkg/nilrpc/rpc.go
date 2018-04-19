package nilrpc

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/chanyoung/nil/pkg/swim"
)

// Prefixes for domains.
const (
	MdsAdminPrefix      = "MDS_ADMIN"
	MdsAuthPrefix       = "MDS_AUTH"
	MdsBucketPrefix     = "MDS_BUCKET"
	MdsClustermapPrefix = "MDS_CLUSTERMAP"
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
	MdsAdminGetLocalChain
	MdsAdminGetAllChain
	MdsAdminGetAllVolume
	MdsAdminAddUser
	MdsAdminRegisterVolume
	MdsAdminGetClusterConfig

	// MDS auth domain methods.
	MdsAuthGetCredential

	// MDS bucket domain methods.
	MdsBucketMakeBucket

	// MDS clustermap domain methods.
	MdsClustermapGetClusterMap
	MdsClustermapGetUpdateNoti
	MdsClustermapUpdateClusterMap

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
	case MdsAdminGetLocalChain:
		return MdsAdminPrefix + "." + "GetLocalChain"
	case MdsAdminGetAllChain:
		return MdsAdminPrefix + "." + "GetAllChain"
	case MdsAdminGetAllVolume:
		return MdsAdminPrefix + "." + "GetAllVolume"
	case MdsAdminAddUser:
		return MdsAdminPrefix + "." + "AddUser"
	case MdsAdminRegisterVolume:
		return MdsAdminPrefix + "." + "RegisterVolume"
	case MdsAdminGetClusterConfig:
		return MdsAdminPrefix + "." + "GetClusterConfig"

	case MdsAuthGetCredential:
		return MdsAuthPrefix + "." + "GetCredential"

	case MdsBucketMakeBucket:
		return MdsBucketPrefix + "." + "MakeBucket"

	case MdsClustermapGetClusterMap:
		return MdsClustermapPrefix + "." + "GetClusterMap"
	case MdsClustermapGetUpdateNoti:
		return MdsClustermapPrefix + "." + "GetUpdateNoti"
	case MdsClustermapUpdateClusterMap:
		return MdsClustermapPrefix + "." + "UpdateClusterMap"

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

// JoinRequest includes an information for joining a new node into the raft clsuter.
// RaftAddr: address of the requested node.
// NodeID: ID of the requested node.
type JoinRequest struct {
	RaftAddr string
	NodeID   string
}

// JoinResponse is a NilRPC response message to join an existing cluster.
type JoinResponse struct{}

// AddUserRequest requests to create a new user with the given name.
type AddUserRequest struct {
	Name string
}

// AddUserResponse response AddUserRequest with the AccessKey and SecretKey.
type AddUserResponse struct {
	AccessKey string
	SecretKey string
}

// RegisterVolumeRequest contains a new volume information.
type RegisterVolumeRequest struct {
	ID     string
	Ds     string
	Size   uint64
	Free   uint64
	Used   uint64
	Speed  string
	Status string
}

// RegisterVolumeResponse contains a registered volume id and the results.
type RegisterVolumeResponse struct {
	ID string
}

// AddVolumeRequest requests to add new volume with the given device path.
type AddVolumeRequest struct {
	DevicePath string
}

// AddVolumeResponse is a response message to add volume request.
type AddVolumeResponse struct{}

type GetLocalChainRequest struct{}
type GetLocalChainResponse struct {
	LocalChainID   int64
	ParityVolumeID int64
	ParityNodeID   int64
}

type GetAllChainRequest struct{}
type GetAllChainResponse struct {
	EncGrps []cmap.EncodingGroup
}

type Volume struct {
	ID     int64
	NodeID int64
}
type GetAllVolumeRequest struct{}
type GetAllVolumeResponse struct {
	Volumes []Volume
}

type GetMembershipListRequest struct{}
type GetMembershipListResponse struct {
	Nodes []swim.Member
}

type ObjectPutRequest struct {
	Name          string
	Bucket        string
	EncodingGroup string
	Volume        string
}
type ObjectPutResponse struct{}

type ObjectGetRequest struct {
	Name   string
	Bucket string
}
type ObjectGetResponse struct {
	EncodingGroupID int64
	VolumeID        int64
	DsID            int64
}

type GetClusterConfigRequest struct{}
type GetClusterConfigResponse struct {
	LocalParityShards int
}

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
