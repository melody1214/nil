// Code generated by protoc-gen-go.
// source: pkg/raft/raftpb/raft.proto
// DO NOT EDIT!

/*
Package raftpb is a generated protocol buffer package.

It is generated from these files:
	pkg/raft/raftpb/raft.proto

It has these top-level messages:
	JoinRequest
	JoinResponse
	RequestVoteRequest
	RequestVoteResponse
	AppendEntriesRequest
	AppendEntriesResponse
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

//
// Basic RPC messages structure in the Diego's paper.
//
// Diego Ongaro and John Ousterhout. 2014. In search of and understandable consensus algorithm.
// In Proceedings of the 2014 USENIX Annual Technical Conference (ATC'14). 305-319.
//
type RequestVoteRequest struct {
	// Candidate's term.
	Term uint64 `protobuf:"varint,1,opt,name=term" json:"term,omitempty"`
	// Candidate's requesting vote.
	CandidateId string `protobuf:"bytes,2,opt,name=candidateId" json:"candidateId,omitempty"`
	// Index of candidate's last log entry.
	LastLogIndex uint64 `protobuf:"varint,3,opt,name=lastLogIndex" json:"lastLogIndex,omitempty"`
	// Term of candidate's last log entry.
	LastLogTerm uint64 `protobuf:"varint,4,opt,name=lastLogTerm" json:"lastLogTerm,omitempty"`
}

func (m *RequestVoteRequest) Reset()                    { *m = RequestVoteRequest{} }
func (m *RequestVoteRequest) String() string            { return proto.CompactTextString(m) }
func (*RequestVoteRequest) ProtoMessage()               {}
func (*RequestVoteRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *RequestVoteRequest) GetTerm() uint64 {
	if m != nil {
		return m.Term
	}
	return 0
}

func (m *RequestVoteRequest) GetCandidateId() string {
	if m != nil {
		return m.CandidateId
	}
	return ""
}

func (m *RequestVoteRequest) GetLastLogIndex() uint64 {
	if m != nil {
		return m.LastLogIndex
	}
	return 0
}

func (m *RequestVoteRequest) GetLastLogTerm() uint64 {
	if m != nil {
		return m.LastLogTerm
	}
	return 0
}

type RequestVoteResponse struct {
	// Current term, for candidate to update itself.
	Term uint64 `protobuf:"varint,1,opt,name=term" json:"term,omitempty"`
	// True means candidate received vote.
	VoteGranted bool `protobuf:"varint,2,opt,name=voteGranted" json:"voteGranted,omitempty"`
}

func (m *RequestVoteResponse) Reset()                    { *m = RequestVoteResponse{} }
func (m *RequestVoteResponse) String() string            { return proto.CompactTextString(m) }
func (*RequestVoteResponse) ProtoMessage()               {}
func (*RequestVoteResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *RequestVoteResponse) GetTerm() uint64 {
	if m != nil {
		return m.Term
	}
	return 0
}

func (m *RequestVoteResponse) GetVoteGranted() bool {
	if m != nil {
		return m.VoteGranted
	}
	return false
}

type AppendEntriesRequest struct {
	// Leader's term.
	Term uint64 `protobuf:"varint,1,opt,name=term" json:"term,omitempty"`
	// So follower can redirect clients.
	LeaderId string `protobuf:"bytes,2,opt,name=leaderId" json:"leaderId,omitempty"`
	// Index of log entry immediately preceding new ones.
	PrevLogIndex uint64 `protobuf:"varint,3,opt,name=prevLogIndex" json:"prevLogIndex,omitempty"`
	// Term of prevLogIndex entry.
	PrevLogTerm uint64 `protobuf:"varint,4,opt,name=prevLogTerm" json:"prevLogTerm,omitempty"`
	// Log entries to store (empty for heartbeat; may send more than one for efficiency)
	Entries []string `protobuf:"bytes,5,rep,name=entries" json:"entries,omitempty"`
	// Leader's commit index.
	LeaderCommit uint64 `protobuf:"varint,6,opt,name=leaderCommit" json:"leaderCommit,omitempty"`
}

func (m *AppendEntriesRequest) Reset()                    { *m = AppendEntriesRequest{} }
func (m *AppendEntriesRequest) String() string            { return proto.CompactTextString(m) }
func (*AppendEntriesRequest) ProtoMessage()               {}
func (*AppendEntriesRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *AppendEntriesRequest) GetTerm() uint64 {
	if m != nil {
		return m.Term
	}
	return 0
}

func (m *AppendEntriesRequest) GetLeaderId() string {
	if m != nil {
		return m.LeaderId
	}
	return ""
}

func (m *AppendEntriesRequest) GetPrevLogIndex() uint64 {
	if m != nil {
		return m.PrevLogIndex
	}
	return 0
}

func (m *AppendEntriesRequest) GetPrevLogTerm() uint64 {
	if m != nil {
		return m.PrevLogTerm
	}
	return 0
}

func (m *AppendEntriesRequest) GetEntries() []string {
	if m != nil {
		return m.Entries
	}
	return nil
}

func (m *AppendEntriesRequest) GetLeaderCommit() uint64 {
	if m != nil {
		return m.LeaderCommit
	}
	return 0
}

type AppendEntriesResponse struct {
	// Current term, for leader to update itself.
	Term uint64 `protobuf:"varint,1,opt,name=term" json:"term,omitempty"`
	// True if follower contained entry matching prevLogIndex and prevLogTerm.
	Success bool `protobuf:"varint,2,opt,name=success" json:"success,omitempty"`
}

func (m *AppendEntriesResponse) Reset()                    { *m = AppendEntriesResponse{} }
func (m *AppendEntriesResponse) String() string            { return proto.CompactTextString(m) }
func (*AppendEntriesResponse) ProtoMessage()               {}
func (*AppendEntriesResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *AppendEntriesResponse) GetTerm() uint64 {
	if m != nil {
		return m.Term
	}
	return 0
}

func (m *AppendEntriesResponse) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func init() {
	proto.RegisterType((*JoinRequest)(nil), "github.com.chanyoung.nil.pkg.raft.raftpb.JoinRequest")
	proto.RegisterType((*JoinResponse)(nil), "github.com.chanyoung.nil.pkg.raft.raftpb.JoinResponse")
	proto.RegisterType((*RequestVoteRequest)(nil), "github.com.chanyoung.nil.pkg.raft.raftpb.RequestVoteRequest")
	proto.RegisterType((*RequestVoteResponse)(nil), "github.com.chanyoung.nil.pkg.raft.raftpb.RequestVoteResponse")
	proto.RegisterType((*AppendEntriesRequest)(nil), "github.com.chanyoung.nil.pkg.raft.raftpb.AppendEntriesRequest")
	proto.RegisterType((*AppendEntriesResponse)(nil), "github.com.chanyoung.nil.pkg.raft.raftpb.AppendEntriesResponse")
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
	// 440 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x9c, 0x53, 0x5d, 0x6f, 0xd3, 0x30,
	0x14, 0x9d, 0xd7, 0xac, 0xed, 0x6e, 0x3b, 0x54, 0xcc, 0x90, 0xa2, 0x3e, 0x55, 0x7e, 0xea, 0x93,
	0x41, 0xe3, 0xe3, 0x85, 0xa7, 0x6e, 0x4c, 0x55, 0xd9, 0x60, 0x52, 0x54, 0xf1, 0x80, 0x84, 0x90,
	0x9b, 0xdc, 0x65, 0xd1, 0x1a, 0xdb, 0x8b, 0x9d, 0x89, 0x3e, 0xf2, 0x0f, 0xf8, 0x3b, 0xfc, 0x05,
	0x7e, 0x15, 0x8a, 0xe3, 0x4d, 0xde, 0x24, 0x26, 0xd8, 0x4b, 0xe2, 0x73, 0x6c, 0xdf, 0x7b, 0xce,
	0xd1, 0x35, 0x8c, 0xf5, 0x65, 0xfe, 0xa2, 0x12, 0xe7, 0xd6, 0x7d, 0xf4, 0xca, 0xfd, 0xb8, 0xae,
	0x94, 0x55, 0x74, 0x9a, 0x17, 0xf6, 0xa2, 0x5e, 0xf1, 0x54, 0x95, 0x3c, 0xbd, 0x10, 0x72, 0xa3,
	0x6a, 0x99, 0x73, 0x59, 0xac, 0xb9, 0xbe, 0xcc, 0xb9, 0x3b, 0xd8, 0x5e, 0x62, 0x7b, 0x30, 0xf8,
	0xa0, 0x0a, 0x99, 0xe0, 0x55, 0x8d, 0xc6, 0xb2, 0x5f, 0x04, 0x86, 0x2d, 0x36, 0x5a, 0x49, 0x83,
	0xf4, 0x2b, 0x0c, 0x4a, 0x34, 0x46, 0xe4, 0xb8, 0xdc, 0x68, 0x8c, 0xc9, 0x84, 0x4c, 0x9f, 0x1c,
	0xbc, 0xe3, 0xff, 0x5a, 0x9f, 0x87, 0xc5, 0x78, 0x53, 0x22, 0x09, 0xeb, 0xd1, 0x7d, 0xd8, 0xb9,
	0xaa, 0xb1, 0xda, 0xc4, 0xdb, 0x13, 0x32, 0xdd, 0x4d, 0x5a, 0xc0, 0x5e, 0x43, 0xe4, 0x76, 0x7b,
	0xd0, 0x99, 0x1d, 0x9d, 0x8c, 0xb6, 0xe8, 0x08, 0x86, 0xef, 0x0f, 0xbf, 0x7d, 0x5c, 0xcc, 0x93,
	0xd9, 0x72, 0x71, 0xf6, 0x69, 0x44, 0xe8, 0x53, 0xd8, 0x3b, 0x3d, 0x9b, 0x07, 0xd4, 0x36, 0xfb,
	0x49, 0x80, 0x7a, 0x1f, 0x9f, 0x95, 0x45, 0xbf, 0xa4, 0x14, 0x22, 0x8b, 0x55, 0xe9, 0xa4, 0x47,
	0x89, 0x5b, 0xd3, 0x09, 0x0c, 0x52, 0x21, 0xb3, 0x22, 0x13, 0x16, 0x17, 0x99, 0x6f, 0x1e, 0x52,
	0x94, 0xc1, 0x70, 0x2d, 0x8c, 0x3d, 0x55, 0xf9, 0x42, 0x66, 0xf8, 0x3d, 0xee, 0xb8, 0xdb, 0x77,
	0xb8, 0xa6, 0x8a, 0xc7, 0xcb, 0xa6, 0x41, 0xe4, 0x8e, 0x84, 0x14, 0x3b, 0x81, 0x67, 0x77, 0x14,
	0xf9, 0x50, 0xff, 0x22, 0xe9, 0x5a, 0x59, 0x9c, 0x57, 0x42, 0x5a, 0x6c, 0x25, 0xf5, 0x93, 0x90,
	0x62, 0xbf, 0x09, 0xec, 0xcf, 0xb4, 0x46, 0x99, 0x1d, 0x4b, 0x5b, 0x15, 0x68, 0x1e, 0x72, 0x38,
	0x86, 0xfe, 0x1a, 0x45, 0x86, 0xd5, 0xad, 0xbd, 0x5b, 0xdc, 0x78, 0xd3, 0x15, 0x5e, 0xdf, 0xf7,
	0x16, 0x72, 0x8d, 0x1c, 0x8f, 0x43, 0x6f, 0x01, 0x45, 0x63, 0xe8, 0x61, 0xab, 0x23, 0xde, 0x99,
	0x74, 0xa6, 0xbb, 0xc9, 0x0d, 0x74, 0xd9, 0xb9, 0x5e, 0x47, 0xaa, 0x2c, 0x0b, 0x1b, 0x77, 0x7d,
	0x76, 0x01, 0xc7, 0x8e, 0xe1, 0xf9, 0x3d, 0x2f, 0x0f, 0x64, 0x13, 0x43, 0xcf, 0xd4, 0x69, 0x8a,
	0xc6, 0xf8, 0x5c, 0x6e, 0xe0, 0xc1, 0x0f, 0x02, 0x51, 0x22, 0xce, 0x2d, 0xdd, 0x40, 0xd4, 0x8c,
	0x1a, 0x7d, 0xf3, 0xbf, 0xa3, 0xe9, 0x22, 0x1c, 0xbf, 0x7d, 0xdc, 0x44, 0xb3, 0xad, 0x97, 0xe4,
	0xb0, 0xff, 0xa5, 0xdb, 0x6e, 0xad, 0xba, 0xee, 0xf5, 0xbd, 0xfa, 0x13, 0x00, 0x00, 0xff, 0xff,
	0xbf, 0xb1, 0x86, 0x88, 0x9b, 0x03, 0x00, 0x00,
}
