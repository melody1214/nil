// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pkg/swim/swim.proto

/*
Package swim is a generated protocol buffer package.

It is generated from these files:
	pkg/swim/swim.proto

It has these top-level messages:
	Member
	Ping
	PingRequest
	Ack
	Test
*/
package swim

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Status int32

const (
	Status_ALIVE   Status = 0
	Status_SUSPECT Status = 1
	Status_FAULTY  Status = 2
)

var Status_name = map[int32]string{
	0: "ALIVE",
	1: "SUSPECT",
	2: "FAULTY",
}
var Status_value = map[string]int32{
	"ALIVE":   0,
	"SUSPECT": 1,
	"FAULTY":  2,
}

func (x Status) String() string {
	return proto.EnumName(Status_name, int32(x))
}
func (Status) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Member struct {
	Uuid        string `protobuf:"bytes,1,opt,name=uuid" json:"uuid,omitempty"`
	Addr        string `protobuf:"bytes,2,opt,name=addr" json:"addr,omitempty"`
	Port        string `protobuf:"bytes,3,opt,name=port" json:"port,omitempty"`
	Status      Status `protobuf:"varint,4,opt,name=status,enum=github.com.chanyoung.nil.pkg.swim.Status" json:"status,omitempty"`
	Incarnation uint32 `protobuf:"varint,5,opt,name=incarnation" json:"incarnation,omitempty"`
}

func (m *Member) Reset()                    { *m = Member{} }
func (m *Member) String() string            { return proto.CompactTextString(m) }
func (*Member) ProtoMessage()               {}
func (*Member) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Member) GetUuid() string {
	if m != nil {
		return m.Uuid
	}
	return ""
}

func (m *Member) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

func (m *Member) GetPort() string {
	if m != nil {
		return m.Port
	}
	return ""
}

func (m *Member) GetStatus() Status {
	if m != nil {
		return m.Status
	}
	return Status_ALIVE
}

func (m *Member) GetIncarnation() uint32 {
	if m != nil {
		return m.Incarnation
	}
	return 0
}

type Ping struct {
	Memlist []*Member `protobuf:"bytes,1,rep,name=memlist" json:"memlist,omitempty"`
}

func (m *Ping) Reset()                    { *m = Ping{} }
func (m *Ping) String() string            { return proto.CompactTextString(m) }
func (*Ping) ProtoMessage()               {}
func (*Ping) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Ping) GetMemlist() []*Member {
	if m != nil {
		return m.Memlist
	}
	return nil
}

type PingRequest struct {
	Memlist []*Member `protobuf:"bytes,1,rep,name=memlist" json:"memlist,omitempty"`
	Target  *Member   `protobuf:"bytes,2,opt,name=target" json:"target,omitempty"`
}

func (m *PingRequest) Reset()                    { *m = PingRequest{} }
func (m *PingRequest) String() string            { return proto.CompactTextString(m) }
func (*PingRequest) ProtoMessage()               {}
func (*PingRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *PingRequest) GetMemlist() []*Member {
	if m != nil {
		return m.Memlist
	}
	return nil
}

func (m *PingRequest) GetTarget() *Member {
	if m != nil {
		return m.Target
	}
	return nil
}

type Ack struct {
}

func (m *Ack) Reset()                    { *m = Ack{} }
func (m *Ack) String() string            { return proto.CompactTextString(m) }
func (*Ack) ProtoMessage()               {}
func (*Ack) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type Test struct {
	Content string `protobuf:"bytes,1,opt,name=content" json:"content,omitempty"`
}

func (m *Test) Reset()                    { *m = Test{} }
func (m *Test) String() string            { return proto.CompactTextString(m) }
func (*Test) ProtoMessage()               {}
func (*Test) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *Test) GetContent() string {
	if m != nil {
		return m.Content
	}
	return ""
}

func init() {
	proto.RegisterType((*Member)(nil), "github.com.chanyoung.nil.pkg.swim.Member")
	proto.RegisterType((*Ping)(nil), "github.com.chanyoung.nil.pkg.swim.Ping")
	proto.RegisterType((*PingRequest)(nil), "github.com.chanyoung.nil.pkg.swim.PingRequest")
	proto.RegisterType((*Ack)(nil), "github.com.chanyoung.nil.pkg.swim.Ack")
	proto.RegisterType((*Test)(nil), "github.com.chanyoung.nil.pkg.swim.Test")
	proto.RegisterEnum("github.com.chanyoung.nil.pkg.swim.Status", Status_name, Status_value)
}

func init() { proto.RegisterFile("pkg/swim/swim.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 310 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x52, 0xd1, 0x4a, 0xc3, 0x30,
	0x14, 0x35, 0x5b, 0xd7, 0xb1, 0x5b, 0x94, 0x11, 0x5f, 0xf2, 0x58, 0xfb, 0x54, 0x45, 0x22, 0xe8,
	0x17, 0xd4, 0x31, 0x41, 0x9c, 0x30, 0xba, 0x4d, 0xd0, 0xb7, 0xae, 0x0b, 0x5d, 0xd8, 0x9a, 0xd4,
	0xf4, 0x06, 0xf1, 0x43, 0xfc, 0x08, 0xff, 0x52, 0xd2, 0x76, 0xe0, 0x9b, 0x0a, 0xbe, 0x84, 0x93,
	0x43, 0xce, 0x39, 0xc9, 0xc9, 0x85, 0xd3, 0x6a, 0x57, 0x5c, 0xd5, 0x6f, 0xb2, 0x6c, 0x16, 0x5e,
	0x19, 0x8d, 0x9a, 0x9e, 0x15, 0x12, 0xb7, 0x76, 0xcd, 0x73, 0x5d, 0xf2, 0x7c, 0x9b, 0xa9, 0x77,
	0x6d, 0x55, 0xc1, 0x95, 0xdc, 0xf3, 0x6a, 0x57, 0x70, 0x77, 0x30, 0xfa, 0x24, 0xe0, 0x3f, 0x8a,
	0x72, 0x2d, 0x0c, 0xa5, 0xe0, 0x59, 0x2b, 0x37, 0x8c, 0x84, 0x24, 0x1e, 0xa5, 0x0d, 0x76, 0x5c,
	0xb6, 0xd9, 0x18, 0xd6, 0x6b, 0x39, 0x87, 0x1d, 0x57, 0x69, 0x83, 0xac, 0xdf, 0x72, 0x0e, 0xd3,
	0x04, 0xfc, 0x1a, 0x33, 0xb4, 0x35, 0xf3, 0x42, 0x12, 0x9f, 0x5c, 0x9f, 0xf3, 0x1f, 0xa3, 0xf9,
	0xa2, 0x11, 0xa4, 0x9d, 0x90, 0x86, 0x10, 0x48, 0x95, 0x67, 0x46, 0x65, 0x28, 0xb5, 0x62, 0x83,
	0x90, 0xc4, 0xc7, 0xe9, 0x77, 0x2a, 0x7a, 0x00, 0x6f, 0x2e, 0x55, 0x41, 0x27, 0x30, 0x2c, 0x45,
	0xb9, 0x97, 0x35, 0x32, 0x12, 0xf6, 0xe3, 0xe0, 0x57, 0x69, 0xed, 0x23, 0xd3, 0x83, 0x32, 0xfa,
	0x20, 0x10, 0x38, 0xb7, 0x54, 0xbc, 0x5a, 0x51, 0xe3, 0xbf, 0x98, 0xba, 0x1a, 0x30, 0x33, 0x85,
	0xc0, 0xa6, 0xb0, 0x3f, 0x79, 0x74, 0xc2, 0x68, 0x00, 0xfd, 0x24, 0xdf, 0x45, 0x21, 0x78, 0x4b,
	0x77, 0x2d, 0x06, 0xc3, 0x5c, 0x2b, 0x14, 0x0a, 0xbb, 0x7f, 0x39, 0x6c, 0x2f, 0x2e, 0xc1, 0x6f,
	0x1b, 0xa4, 0x23, 0x18, 0x24, 0xb3, 0xfb, 0xa7, 0xe9, 0xf8, 0x88, 0x06, 0x30, 0x5c, 0xac, 0x16,
	0xf3, 0xe9, 0x64, 0x39, 0x26, 0x14, 0xc0, 0xbf, 0x4b, 0x56, 0xb3, 0xe5, 0xf3, 0xb8, 0x77, 0xeb,
	0xbf, 0x78, 0x2e, 0x6d, 0xed, 0x37, 0x93, 0x71, 0xf3, 0x15, 0x00, 0x00, 0xff, 0xff, 0x3b, 0x53,
	0xe4, 0xd3, 0x30, 0x02, 0x00, 0x00,
}
