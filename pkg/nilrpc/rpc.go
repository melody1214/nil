package nilrpc

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/s3"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/chanyoung/nil/pkg/swim"
)

// MDSADMINPrefix is the prefix for calling mds rpc methods.
const (
	MdsAdminPrefix      = "MDS_ADMIN"
	MdsAuthPrefix       = "MDS_AUTH"
	MdsBucketPrefix     = "MDS_BUCKET"
	MdsClustermapPrefix = "MDS_CLUSTERMAP"
	MdsMembershipPrefix = "MDS_MEMBERSHIP"
	MdsRecoveryPrefix   = "MDS_RECOVERY"

	DSRPCPrefix = "DS"
)

// DSRPCPrefix is the prefix for calling ds rpc methods.

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

	// MDS auth domain methods.
	MdsAuthGetCredential

	// MDS bucket domain methods.
	MdsBucketAddBucket

	// MDS clustermap domain methods.
	MdsClustermapGetClusterMap
	MdsClustermapIsUpdated

	// MDS membership domain methods.
	MdsMembershipGetMembershipList

	// MDS recovery domain methods
	MdsRecoveryRecover
	MdsRecoveryRebalance

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

	case MdsAuthGetCredential:
		return MdsAuthPrefix + "." + "GetCredential"

	case MdsBucketAddBucket:
		return MdsBucketPrefix + "." + "AddBucket"

	case MdsClustermapGetClusterMap:
		return MdsClustermapPrefix + "." + "GetClusterMap"
	case MdsClustermapIsUpdated:
		return MdsClustermapPrefix + "." + "IsUpdated"

	case MdsMembershipGetMembershipList:
		return MdsMembershipPrefix + "." + "GetMembershipList"

	case MdsRecoveryRecover:
		return MdsRecoveryPrefix + "." + "Recover"
	case MdsRecoveryRebalance:
		return MdsRecoveryPrefix + "." + "Rebalance"

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

// GetCredentialRequest requests a credential for the given access key.
type GetCredentialRequest struct {
	AccessKey string
}

// GetCredentialResponse response the credential.
type GetCredentialResponse struct {
	Exist     bool
	AccessKey string
	SecretKey string
}

// AddBucketRequest requests to create bucket for given name and user.
type AddBucketRequest struct {
	BucketName string
	AccessKey  string
	Region     string
}

// AddBucketResponse responses the result of addBucket.
type AddBucketResponse struct {
	S3ErrCode s3.ErrorCode
}

// GetClusterMapRequest requests to get local cluster map.
// Version == 0; requests the latest version.
// Version > 0; requests higher version than given version.
type GetClusterMapRequest struct {
	Version int64
}

// GetClusterMapResponse contains a current local cluster members.
type GetClusterMapResponse struct {
	Version int64
	Nodes   []ClusterNode
}

// ClusterMapIsUpdatedRequest requests to receive notification
// when the cluster map is updated. Gives some notification if
// has higher than given version of cluster map.
type ClusterMapIsUpdatedRequest struct {
	Version int64
}

// ClusterMapIsUpdatedResponse will response the cluster map is updated.
type ClusterMapIsUpdatedResponse struct{}

// ClusterNode represents the nodes.
type ClusterNode struct {
	ID   int64
	Name string
	Addr string
	Type string
	Stat string
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

type Chain struct {
	ID             int64
	FirstVolumeID  int64
	SecondVolumeID int64
	ThirdVolumeID  int64
	ParityVolumeID int64
}
type GetAllChainRequest struct{}
type GetAllChainResponse struct {
	Chains []Chain
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

type RecoverRequest struct {
	Pe swim.PingError
}
type RecoverResponse struct{}

type RebalanceRequest struct{}
type RebalanceResponse struct{}

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
