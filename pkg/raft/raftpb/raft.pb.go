// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pkg/raft/raftpb/raft.proto

/*
Package raftpb is a generated protocol buffer package.

It is generated from these files:
	pkg/raft/raftpb/raft.proto

It has these top-level messages:
	JoinRequest
	JoinResponse
*/
package raftpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type JoinResponse_Type int32

const (
	JoinResponse_ACK           JoinResponse_Type = 0
	JoinResponse_DB_MIGRATION  JoinResponse_Type = 1
	JoinResponse_LOG_MIGRATION JoinResponse_Type = 2
)

var JoinResponse_Type_name = map[int32]string{
	0: "ACK",
	1: "DB_MIGRATION",
	2: "LOG_MIGRATION",
}
var JoinResponse_Type_value = map[string]int32{
	"ACK":           0,
	"DB_MIGRATION":  1,
	"LOG_MIGRATION": 2,
}

func (x JoinResponse_Type) String() string {
	return proto.EnumName(JoinResponse_Type_name, int32(x))
}
func (JoinResponse_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1, 0} }

type JoinRequest struct {
}

func (m *JoinRequest) Reset()                    { *m = JoinRequest{} }
func (m *JoinRequest) String() string            { return proto.CompactTextString(m) }
func (*JoinRequest) ProtoMessage()               {}
func (*JoinRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type JoinResponse struct {
	MessageType JoinResponse_Type `protobuf:"varint,1,opt,name=messageType,enum=github.com.chanyoung.nil.pkg.raft.raftpb.JoinResponse_Type" json:"messageType,omitempty"`
	Query       string            `protobuf:"bytes,2,opt,name=query" json:"query,omitempty"`
}

func (m *JoinResponse) Reset()                    { *m = JoinResponse{} }
func (m *JoinResponse) String() string            { return proto.CompactTextString(m) }
func (*JoinResponse) ProtoMessage()               {}
func (*JoinResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *JoinResponse) GetMessageType() JoinResponse_Type {
	if m != nil {
		return m.MessageType
	}
	return JoinResponse_ACK
}

func (m *JoinResponse) GetQuery() string {
	if m != nil {
		return m.Query
	}
	return ""
}

func init() {
	proto.RegisterType((*JoinRequest)(nil), "github.com.chanyoung.nil.pkg.raft.raftpb.JoinRequest")
	proto.RegisterType((*JoinResponse)(nil), "github.com.chanyoung.nil.pkg.raft.raftpb.JoinResponse")
	proto.RegisterEnum("github.com.chanyoung.nil.pkg.raft.raftpb.JoinResponse_Type", JoinResponse_Type_name, JoinResponse_Type_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Raft service

type RaftClient interface {
	Join(ctx context.Context, in *JoinRequest, opts ...grpc.CallOption) (Raft_JoinClient, error)
}

type raftClient struct {
	cc *grpc.ClientConn
}

func NewRaftClient(cc *grpc.ClientConn) RaftClient {
	return &raftClient{cc}
}

func (c *raftClient) Join(ctx context.Context, in *JoinRequest, opts ...grpc.CallOption) (Raft_JoinClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Raft_serviceDesc.Streams[0], c.cc, "/github.com.chanyoung.nil.pkg.raft.raftpb.Raft/Join", opts...)
	if err != nil {
		return nil, err
	}
	x := &raftJoinClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Raft_JoinClient interface {
	Recv() (*JoinResponse, error)
	grpc.ClientStream
}

type raftJoinClient struct {
	grpc.ClientStream
}

func (x *raftJoinClient) Recv() (*JoinResponse, error) {
	m := new(JoinResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Raft service

type RaftServer interface {
	Join(*JoinRequest, Raft_JoinServer) error
}

func RegisterRaftServer(s *grpc.Server, srv RaftServer) {
	s.RegisterService(&_Raft_serviceDesc, srv)
}

func _Raft_Join_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(JoinRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(RaftServer).Join(m, &raftJoinServer{stream})
}

type Raft_JoinServer interface {
	Send(*JoinResponse) error
	grpc.ServerStream
}

type raftJoinServer struct {
	grpc.ServerStream
}

func (x *raftJoinServer) Send(m *JoinResponse) error {
	return x.ServerStream.SendMsg(m)
}

var _Raft_serviceDesc = grpc.ServiceDesc{
	ServiceName: "github.com.chanyoung.nil.pkg.raft.raftpb.Raft",
	HandlerType: (*RaftServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Join",
			Handler:       _Raft_Join_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pkg/raft/raftpb/raft.proto",
}

func init() { proto.RegisterFile("pkg/raft/raftpb/raft.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 246 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2a, 0xc8, 0x4e, 0xd7,
	0x2f, 0x4a, 0x4c, 0x2b, 0x01, 0x13, 0x05, 0x49, 0x60, 0x4a, 0xaf, 0xa0, 0x28, 0xbf, 0x24, 0x5f,
	0x48, 0x23, 0x3d, 0xb3, 0x24, 0xa3, 0x34, 0x49, 0x2f, 0x39, 0x3f, 0x57, 0x2f, 0x39, 0x23, 0x31,
	0xaf, 0x32, 0xbf, 0x34, 0x2f, 0x5d, 0x2f, 0x2f, 0x33, 0x47, 0xaf, 0x20, 0x3b, 0x5d, 0x0f, 0xac,
	0x10, 0xa2, 0x49, 0x89, 0x97, 0x8b, 0xdb, 0x2b, 0x3f, 0x33, 0x2f, 0x28, 0xb5, 0xb0, 0x34, 0xb5,
	0xb8, 0x44, 0x69, 0x27, 0x23, 0x17, 0x0f, 0x84, 0x5f, 0x5c, 0x90, 0x9f, 0x57, 0x9c, 0x2a, 0x14,
	0xcb, 0xc5, 0x9d, 0x9b, 0x5a, 0x5c, 0x9c, 0x98, 0x9e, 0x1a, 0x52, 0x59, 0x90, 0x2a, 0xc1, 0xa8,
	0xc0, 0xa8, 0xc1, 0x67, 0x64, 0xad, 0x47, 0xac, 0xf9, 0x7a, 0xc8, 0x86, 0xe9, 0x81, 0x8c, 0x08,
	0x42, 0x36, 0x4f, 0x48, 0x84, 0x8b, 0xb5, 0xb0, 0x34, 0xb5, 0xa8, 0x52, 0x82, 0x49, 0x81, 0x51,
	0x83, 0x33, 0x08, 0xc2, 0x51, 0x32, 0xe1, 0x62, 0x01, 0xcb, 0xb2, 0x73, 0x31, 0x3b, 0x3a, 0x7b,
	0x0b, 0x30, 0x08, 0x09, 0x70, 0xf1, 0xb8, 0x38, 0xc5, 0xfb, 0x7a, 0xba, 0x07, 0x39, 0x86, 0x78,
	0xfa, 0xfb, 0x09, 0x30, 0x0a, 0x09, 0x72, 0xf1, 0xfa, 0xf8, 0xbb, 0x23, 0x09, 0x31, 0x19, 0x35,
	0x32, 0x72, 0xb1, 0x04, 0x25, 0xa6, 0x95, 0x08, 0x55, 0x72, 0xb1, 0x80, 0xac, 0x15, 0x32, 0x25,
	0xd5, 0x99, 0xe0, 0x30, 0x90, 0x32, 0x23, 0xcf, 0x77, 0x4a, 0x0c, 0x06, 0x8c, 0x4e, 0x1c, 0x51,
	0x6c, 0x10, 0xa9, 0x24, 0x36, 0x70, 0x4c, 0x18, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0xdd, 0x8b,
	0xe5, 0x95, 0xa7, 0x01, 0x00, 0x00,
}