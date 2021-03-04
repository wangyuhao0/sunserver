// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proto/msgproto/playerinfo.proto

package msg

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
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

//用户基本信息
type PlayerInfo struct {
	UserId               uint64   `protobuf:"varint,1,opt,name=userId,proto3" json:"userId,omitempty"`
	Rank                 uint64   `protobuf:"varint,2,opt,name=rank,proto3" json:"rank,omitempty"`
	NickName             string   `protobuf:"bytes,3,opt,name=nickName,proto3" json:"nickName,omitempty"`
	Sex                  int32    `protobuf:"varint,4,opt,name=sex,proto3" json:"sex,omitempty"`
	Avatar               string   `protobuf:"bytes,5,opt,name=avatar,proto3" json:"avatar,omitempty"`
	ClientId             uint64   `protobuf:"varint,6,opt,name=clientId,proto3" json:"clientId,omitempty"`
	IsOwner              bool     `protobuf:"varint,7,opt,name=isOwner,proto3" json:"isOwner,omitempty"`
	SeatNum              int32    `protobuf:"varint,8,opt,name=seatNum,proto3" json:"seatNum,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PlayerInfo) Reset()         { *m = PlayerInfo{} }
func (m *PlayerInfo) String() string { return proto.CompactTextString(m) }
func (*PlayerInfo) ProtoMessage()    {}
func (*PlayerInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_f81d971039664abd, []int{0}
}

func (m *PlayerInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PlayerInfo.Unmarshal(m, b)
}
func (m *PlayerInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PlayerInfo.Marshal(b, m, deterministic)
}
func (m *PlayerInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PlayerInfo.Merge(m, src)
}
func (m *PlayerInfo) XXX_Size() int {
	return xxx_messageInfo_PlayerInfo.Size(m)
}
func (m *PlayerInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_PlayerInfo.DiscardUnknown(m)
}

var xxx_messageInfo_PlayerInfo proto.InternalMessageInfo

func (m *PlayerInfo) GetUserId() uint64 {
	if m != nil {
		return m.UserId
	}
	return 0
}

func (m *PlayerInfo) GetRank() uint64 {
	if m != nil {
		return m.Rank
	}
	return 0
}

func (m *PlayerInfo) GetNickName() string {
	if m != nil {
		return m.NickName
	}
	return ""
}

func (m *PlayerInfo) GetSex() int32 {
	if m != nil {
		return m.Sex
	}
	return 0
}

func (m *PlayerInfo) GetAvatar() string {
	if m != nil {
		return m.Avatar
	}
	return ""
}

func (m *PlayerInfo) GetClientId() uint64 {
	if m != nil {
		return m.ClientId
	}
	return 0
}

func (m *PlayerInfo) GetIsOwner() bool {
	if m != nil {
		return m.IsOwner
	}
	return false
}

func (m *PlayerInfo) GetSeatNum() int32 {
	if m != nil {
		return m.SeatNum
	}
	return 0
}

func init() {
	proto.RegisterType((*PlayerInfo)(nil), "msg.PlayerInfo")
}

func init() { proto.RegisterFile("proto/msgproto/playerinfo.proto", fileDescriptor_f81d971039664abd) }

var fileDescriptor_f81d971039664abd = []byte{
	// 208 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0x8f, 0xbd, 0x4e, 0x04, 0x21,
	0x14, 0x85, 0x83, 0xf3, 0xeb, 0xad, 0x0c, 0x85, 0xb9, 0xb1, 0x91, 0x58, 0x51, 0xad, 0x85, 0xa5,
	0x9d, 0xdd, 0x34, 0xab, 0xa1, 0xb4, 0xc3, 0x5d, 0x76, 0x42, 0x76, 0x81, 0x09, 0x30, 0xfe, 0xbc,
	0xa4, 0xcf, 0x64, 0xb8, 0xce, 0x4c, 0x77, 0xbe, 0x73, 0xc8, 0x47, 0x2e, 0xdc, 0x4f, 0x31, 0xe4,
	0xf0, 0xe8, 0xd2, 0xf8, 0x1f, 0xa6, 0x8b, 0xfe, 0x31, 0xd1, 0xfa, 0x53, 0xd8, 0x51, 0xc1, 0x2b,
	0x97, 0xc6, 0x87, 0x5f, 0x06, 0xf0, 0x46, 0xcb, 0xe0, 0x4f, 0x81, 0xdf, 0x42, 0x3b, 0x27, 0x13,
	0x87, 0x23, 0x32, 0xc1, 0x64, 0xad, 0x16, 0xe2, 0x1c, 0xea, 0xa8, 0xfd, 0x19, 0xaf, 0xa8, 0xa5,
	0xcc, 0xef, 0xa0, 0xf7, 0xf6, 0x70, 0xde, 0x6b, 0x67, 0xb0, 0x12, 0x4c, 0x5e, 0xab, 0x8d, 0xf9,
	0x0d, 0x54, 0xc9, 0x7c, 0x63, 0x2d, 0x98, 0x6c, 0x54, 0x89, 0xc5, 0xac, 0x3f, 0x75, 0xd6, 0x11,
	0x1b, 0x7a, 0xbb, 0x50, 0xb1, 0x1c, 0x2e, 0xd6, 0xf8, 0x3c, 0x1c, 0xb1, 0x25, 0xfb, 0xc6, 0x1c,
	0xa1, 0xb3, 0xe9, 0xf5, 0xcb, 0x9b, 0x88, 0x9d, 0x60, 0xb2, 0x57, 0x2b, 0x96, 0x25, 0x19, 0x9d,
	0xf7, 0xb3, 0xc3, 0x9e, 0xfe, 0x58, 0xf1, 0xa5, 0x7b, 0x6f, 0x76, 0xcf, 0x2e, 0x8d, 0x1f, 0x2d,
	0x5d, 0xf9, 0xf4, 0x17, 0x00, 0x00, 0xff, 0xff, 0xe8, 0x5f, 0x0f, 0xbf, 0x08, 0x01, 0x00, 0x00,
}