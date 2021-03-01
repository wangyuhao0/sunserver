// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proto/msgproto/table.proto

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

type MsgAddTableReq struct {
	UserId               uint64   `protobuf:"varint,1,opt,name=UserId,proto3" json:"UserId,omitempty"`
	TableUuid            string   `protobuf:"bytes,2,opt,name=TableUuid,proto3" json:"TableUuid,omitempty"`
	Flag                 int32    `protobuf:"varint,3,opt,name=flag,proto3" json:"flag,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MsgAddTableReq) Reset()         { *m = MsgAddTableReq{} }
func (m *MsgAddTableReq) String() string { return proto.CompactTextString(m) }
func (*MsgAddTableReq) ProtoMessage()    {}
func (*MsgAddTableReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_6cf7e9aa369046db, []int{0}
}

func (m *MsgAddTableReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MsgAddTableReq.Unmarshal(m, b)
}
func (m *MsgAddTableReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MsgAddTableReq.Marshal(b, m, deterministic)
}
func (m *MsgAddTableReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgAddTableReq.Merge(m, src)
}
func (m *MsgAddTableReq) XXX_Size() int {
	return xxx_messageInfo_MsgAddTableReq.Size(m)
}
func (m *MsgAddTableReq) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgAddTableReq.DiscardUnknown(m)
}

var xxx_messageInfo_MsgAddTableReq proto.InternalMessageInfo

func (m *MsgAddTableReq) GetUserId() uint64 {
	if m != nil {
		return m.UserId
	}
	return 0
}

func (m *MsgAddTableReq) GetTableUuid() string {
	if m != nil {
		return m.TableUuid
	}
	return ""
}

func (m *MsgAddTableReq) GetFlag() int32 {
	if m != nil {
		return m.Flag
	}
	return 0
}

type MsgAddTableRes struct {
	Ret                  ErrCode  `protobuf:"varint,1,opt,name=Ret,proto3,enum=msg.ErrCode" json:"Ret,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MsgAddTableRes) Reset()         { *m = MsgAddTableRes{} }
func (m *MsgAddTableRes) String() string { return proto.CompactTextString(m) }
func (*MsgAddTableRes) ProtoMessage()    {}
func (*MsgAddTableRes) Descriptor() ([]byte, []int) {
	return fileDescriptor_6cf7e9aa369046db, []int{1}
}

func (m *MsgAddTableRes) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MsgAddTableRes.Unmarshal(m, b)
}
func (m *MsgAddTableRes) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MsgAddTableRes.Marshal(b, m, deterministic)
}
func (m *MsgAddTableRes) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgAddTableRes.Merge(m, src)
}
func (m *MsgAddTableRes) XXX_Size() int {
	return xxx_messageInfo_MsgAddTableRes.Size(m)
}
func (m *MsgAddTableRes) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgAddTableRes.DiscardUnknown(m)
}

var xxx_messageInfo_MsgAddTableRes proto.InternalMessageInfo

func (m *MsgAddTableRes) GetRet() ErrCode {
	if m != nil {
		return m.Ret
	}
	return ErrCode_OK
}

type MsgClientConnectedStatusRes struct {
	Ret                  ErrCode  `protobuf:"varint,1,opt,name=Ret,proto3,enum=msg.ErrCode" json:"Ret,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MsgClientConnectedStatusRes) Reset()         { *m = MsgClientConnectedStatusRes{} }
func (m *MsgClientConnectedStatusRes) String() string { return proto.CompactTextString(m) }
func (*MsgClientConnectedStatusRes) ProtoMessage()    {}
func (*MsgClientConnectedStatusRes) Descriptor() ([]byte, []int) {
	return fileDescriptor_6cf7e9aa369046db, []int{2}
}

func (m *MsgClientConnectedStatusRes) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MsgClientConnectedStatusRes.Unmarshal(m, b)
}
func (m *MsgClientConnectedStatusRes) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MsgClientConnectedStatusRes.Marshal(b, m, deterministic)
}
func (m *MsgClientConnectedStatusRes) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgClientConnectedStatusRes.Merge(m, src)
}
func (m *MsgClientConnectedStatusRes) XXX_Size() int {
	return xxx_messageInfo_MsgClientConnectedStatusRes.Size(m)
}
func (m *MsgClientConnectedStatusRes) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgClientConnectedStatusRes.DiscardUnknown(m)
}

var xxx_messageInfo_MsgClientConnectedStatusRes proto.InternalMessageInfo

func (m *MsgClientConnectedStatusRes) GetRet() ErrCode {
	if m != nil {
		return m.Ret
	}
	return ErrCode_OK
}

func init() {
	proto.RegisterType((*MsgAddTableReq)(nil), "msg.MsgAddTableReq")
	proto.RegisterType((*MsgAddTableRes)(nil), "msg.MsgAddTableRes")
	proto.RegisterType((*MsgClientConnectedStatusRes)(nil), "msg.MsgClientConnectedStatusRes")
}

func init() { proto.RegisterFile("proto/msgproto/table.proto", fileDescriptor_6cf7e9aa369046db) }

var fileDescriptor_6cf7e9aa369046db = []byte{
	// 209 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2a, 0x28, 0xca, 0x2f,
	0xc9, 0xd7, 0xcf, 0x2d, 0x4e, 0x87, 0x30, 0x4a, 0x12, 0x93, 0x72, 0x52, 0xf5, 0xc0, 0x6c, 0x21,
	0xe6, 0xdc, 0xe2, 0x74, 0x29, 0x19, 0x34, 0x05, 0xa9, 0x45, 0x45, 0xc9, 0xf9, 0x29, 0x50, 0x25,
	0x4a, 0x51, 0x5c, 0x7c, 0xbe, 0xc5, 0xe9, 0x8e, 0x29, 0x29, 0x21, 0x20, 0x7d, 0x41, 0xa9, 0x85,
	0x42, 0x62, 0x5c, 0x6c, 0xa1, 0xc5, 0xa9, 0x45, 0x9e, 0x29, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x2c,
	0x41, 0x50, 0x9e, 0x90, 0x0c, 0x17, 0x27, 0x58, 0x4d, 0x68, 0x69, 0x66, 0x8a, 0x04, 0x93, 0x02,
	0xa3, 0x06, 0x67, 0x10, 0x42, 0x40, 0x48, 0x88, 0x8b, 0x25, 0x2d, 0x27, 0x31, 0x5d, 0x82, 0x59,
	0x81, 0x51, 0x83, 0x35, 0x08, 0xcc, 0x56, 0x32, 0x40, 0x33, 0xbb, 0x58, 0x48, 0x8e, 0x8b, 0x39,
	0x28, 0xb5, 0x04, 0x6c, 0x30, 0x9f, 0x11, 0x8f, 0x5e, 0x6e, 0x71, 0xba, 0x9e, 0x6b, 0x51, 0x91,
	0x73, 0x7e, 0x4a, 0x6a, 0x10, 0x48, 0x42, 0xc9, 0x96, 0x4b, 0xda, 0xb7, 0x38, 0xdd, 0x39, 0x27,
	0x33, 0x35, 0xaf, 0xc4, 0x39, 0x3f, 0x2f, 0x2f, 0x35, 0xb9, 0x24, 0x35, 0x25, 0xb8, 0x24, 0xb1,
	0xa4, 0xb4, 0x98, 0x08, 0xed, 0x4e, 0xec, 0x51, 0xac, 0x7a, 0xd6, 0xb9, 0xc5, 0xe9, 0x49, 0x6c,
	0x60, 0xcf, 0x19, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0xf3, 0x59, 0xd8, 0x46, 0x1d, 0x01, 0x00,
	0x00,
}
