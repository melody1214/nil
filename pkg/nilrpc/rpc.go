package nilrpc

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/s3"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/chanyoung/nil/pkg/swim"
)

// MDSRPCPrefix is the prefix for calling mds rpc methods.
const MDSRPCPrefix = "MDS"

// DSRPCPrefix is the prefix for calling ds rpc methods.
const DSRPCPrefix = "DS"

// MethodName indicates what procedure will be called.
type MethodName int

const (
	// MDS methods.
	Join MethodName = iota
	AddUser
	GetCredential
	AddBucket
	GetClusterMap
	RegisterVolume

	// DS methods.
	AddVolume
)

func (m MethodName) String() string {
	switch m {
	case Join:
		return MDSRPCPrefix + "." + "Join"
	case AddUser:
		return MDSRPCPrefix + "." + "AddUser"
	case GetCredential:
		return MDSRPCPrefix + "." + "GetCredential"
	case AddBucket:
		return MDSRPCPrefix + "." + "AddBucket"
	case GetClusterMap:
		return MDSRPCPrefix + "." + "GetClusterMap"
	case RegisterVolume:
		return MDSRPCPrefix + "." + "RegisterVolume"
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
}

// AddBucketResponse responses the result of addBucket.
type AddBucketResponse struct {
	S3ErrCode s3.ErrorCode
}

// GetClusterMapRequest requests to get current local cluster map.
type GetClusterMapRequest struct{}

// GetClusterMapResponse contains a current local cluster members.
type GetClusterMapResponse struct {
	Members []swim.Member
}

// RegisterVolumeRequest contains a new volume information.
type RegisterVolumeRequest struct {
	ID string
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
