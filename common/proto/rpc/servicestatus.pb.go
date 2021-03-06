// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proto/rpcproto/servicestatus.proto

package rpc

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

//PlayerService->CenterService同步负载情况
type PlayerServiceBalance struct {
	NodeId               int32    `protobuf:"varint,1,opt,name=NodeId,proto3" json:"NodeId,omitempty"`
	Weigh                int32    `protobuf:"varint,2,opt,name=Weigh,proto3" json:"Weigh,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PlayerServiceBalance) Reset()         { *m = PlayerServiceBalance{} }
func (m *PlayerServiceBalance) String() string { return proto.CompactTextString(m) }
func (*PlayerServiceBalance) ProtoMessage()    {}
func (*PlayerServiceBalance) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e2a4236feecff6a, []int{0}
}

func (m *PlayerServiceBalance) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PlayerServiceBalance.Unmarshal(m, b)
}
func (m *PlayerServiceBalance) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PlayerServiceBalance.Marshal(b, m, deterministic)
}
func (m *PlayerServiceBalance) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PlayerServiceBalance.Merge(m, src)
}
func (m *PlayerServiceBalance) XXX_Size() int {
	return xxx_messageInfo_PlayerServiceBalance.Size(m)
}
func (m *PlayerServiceBalance) XXX_DiscardUnknown() {
	xxx_messageInfo_PlayerServiceBalance.DiscardUnknown(m)
}

var xxx_messageInfo_PlayerServiceBalance proto.InternalMessageInfo

func (m *PlayerServiceBalance) GetNodeId() int32 {
	if m != nil {
		return m.NodeId
	}
	return 0
}

func (m *PlayerServiceBalance) GetWeigh() int32 {
	if m != nil {
		return m.Weigh
	}
	return 0
}

type CheckOnLineReq struct {
	ClientId             uint64   `protobuf:"varint,1,opt,name=ClientId,proto3" json:"ClientId,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CheckOnLineReq) Reset()         { *m = CheckOnLineReq{} }
func (m *CheckOnLineReq) String() string { return proto.CompactTextString(m) }
func (*CheckOnLineReq) ProtoMessage()    {}
func (*CheckOnLineReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e2a4236feecff6a, []int{1}
}

func (m *CheckOnLineReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CheckOnLineReq.Unmarshal(m, b)
}
func (m *CheckOnLineReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CheckOnLineReq.Marshal(b, m, deterministic)
}
func (m *CheckOnLineReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CheckOnLineReq.Merge(m, src)
}
func (m *CheckOnLineReq) XXX_Size() int {
	return xxx_messageInfo_CheckOnLineReq.Size(m)
}
func (m *CheckOnLineReq) XXX_DiscardUnknown() {
	xxx_messageInfo_CheckOnLineReq.DiscardUnknown(m)
}

var xxx_messageInfo_CheckOnLineReq proto.InternalMessageInfo

func (m *CheckOnLineReq) GetClientId() uint64 {
	if m != nil {
		return m.ClientId
	}
	return 0
}

type CheckOnLineRes struct {
	Flag                 bool     `protobuf:"varint,1,opt,name=flag,proto3" json:"flag,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CheckOnLineRes) Reset()         { *m = CheckOnLineRes{} }
func (m *CheckOnLineRes) String() string { return proto.CompactTextString(m) }
func (*CheckOnLineRes) ProtoMessage()    {}
func (*CheckOnLineRes) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e2a4236feecff6a, []int{2}
}

func (m *CheckOnLineRes) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CheckOnLineRes.Unmarshal(m, b)
}
func (m *CheckOnLineRes) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CheckOnLineRes.Marshal(b, m, deterministic)
}
func (m *CheckOnLineRes) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CheckOnLineRes.Merge(m, src)
}
func (m *CheckOnLineRes) XXX_Size() int {
	return xxx_messageInfo_CheckOnLineRes.Size(m)
}
func (m *CheckOnLineRes) XXX_DiscardUnknown() {
	xxx_messageInfo_CheckOnLineRes.DiscardUnknown(m)
}

var xxx_messageInfo_CheckOnLineRes proto.InternalMessageInfo

func (m *CheckOnLineRes) GetFlag() bool {
	if m != nil {
		return m.Flag
	}
	return false
}

//PlayerService->CenterService刷新所有玩家的列表
type UpdatePlayerList struct {
	NodeId               int32    `protobuf:"varint,1,opt,name=NodeId,proto3" json:"NodeId,omitempty"`
	UList                []uint64 `protobuf:"varint,2,rep,packed,name=UList,proto3" json:"UList,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UpdatePlayerList) Reset()         { *m = UpdatePlayerList{} }
func (m *UpdatePlayerList) String() string { return proto.CompactTextString(m) }
func (*UpdatePlayerList) ProtoMessage()    {}
func (*UpdatePlayerList) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e2a4236feecff6a, []int{3}
}

func (m *UpdatePlayerList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdatePlayerList.Unmarshal(m, b)
}
func (m *UpdatePlayerList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdatePlayerList.Marshal(b, m, deterministic)
}
func (m *UpdatePlayerList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdatePlayerList.Merge(m, src)
}
func (m *UpdatePlayerList) XXX_Size() int {
	return xxx_messageInfo_UpdatePlayerList.Size(m)
}
func (m *UpdatePlayerList) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdatePlayerList.DiscardUnknown(m)
}

var xxx_messageInfo_UpdatePlayerList proto.InternalMessageInfo

func (m *UpdatePlayerList) GetNodeId() int32 {
	if m != nil {
		return m.NodeId
	}
	return 0
}

func (m *UpdatePlayerList) GetUList() []uint64 {
	if m != nil {
		return m.UList
	}
	return nil
}

//RoomService->TableService 初始化创建房间
type CreateTable struct {
	TableUuid                string   `protobuf:"bytes,1,opt,name=TableUuid,proto3" json:"TableUuid,omitempty"`
	TableType                int32    `protobuf:"varint,2,opt,name=TableType,proto3" json:"TableType,omitempty"`
	PlayerNum                int32    `protobuf:"varint,3,opt,name=PlayerNum,proto3" json:"PlayerNum,omitempty"`
	RoomUuidList             []string `protobuf:"bytes,4,rep,name=roomUuidList,proto3" json:"roomUuidList,omitempty"`
	ShouldConnectedClintList []uint64 `protobuf:"varint,5,rep,packed,name=shouldConnectedClintList,proto3" json:"shouldConnectedClintList,omitempty"`
	XXX_NoUnkeyedLiteral     struct{} `json:"-"`
	XXX_unrecognized         []byte   `json:"-"`
	XXX_sizecache            int32    `json:"-"`
}

func (m *CreateTable) Reset()         { *m = CreateTable{} }
func (m *CreateTable) String() string { return proto.CompactTextString(m) }
func (*CreateTable) ProtoMessage()    {}
func (*CreateTable) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e2a4236feecff6a, []int{4}
}

func (m *CreateTable) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateTable.Unmarshal(m, b)
}
func (m *CreateTable) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateTable.Marshal(b, m, deterministic)
}
func (m *CreateTable) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateTable.Merge(m, src)
}
func (m *CreateTable) XXX_Size() int {
	return xxx_messageInfo_CreateTable.Size(m)
}
func (m *CreateTable) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateTable.DiscardUnknown(m)
}

var xxx_messageInfo_CreateTable proto.InternalMessageInfo

func (m *CreateTable) GetTableUuid() string {
	if m != nil {
		return m.TableUuid
	}
	return ""
}

func (m *CreateTable) GetTableType() int32 {
	if m != nil {
		return m.TableType
	}
	return 0
}

func (m *CreateTable) GetPlayerNum() int32 {
	if m != nil {
		return m.PlayerNum
	}
	return 0
}

func (m *CreateTable) GetRoomUuidList() []string {
	if m != nil {
		return m.RoomUuidList
	}
	return nil
}

func (m *CreateTable) GetShouldConnectedClintList() []uint64 {
	if m != nil {
		return m.ShouldConnectedClintList
	}
	return nil
}

//PlayerService->RoomService关闭单个房间
type RemoveOneRoom struct {
	RoomUuid             uint64   `protobuf:"varint,1,opt,name=RoomUuid,proto3" json:"RoomUuid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RemoveOneRoom) Reset()         { *m = RemoveOneRoom{} }
func (m *RemoveOneRoom) String() string { return proto.CompactTextString(m) }
func (*RemoveOneRoom) ProtoMessage()    {}
func (*RemoveOneRoom) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e2a4236feecff6a, []int{5}
}

func (m *RemoveOneRoom) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RemoveOneRoom.Unmarshal(m, b)
}
func (m *RemoveOneRoom) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RemoveOneRoom.Marshal(b, m, deterministic)
}
func (m *RemoveOneRoom) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemoveOneRoom.Merge(m, src)
}
func (m *RemoveOneRoom) XXX_Size() int {
	return xxx_messageInfo_RemoveOneRoom.Size(m)
}
func (m *RemoveOneRoom) XXX_DiscardUnknown() {
	xxx_messageInfo_RemoveOneRoom.DiscardUnknown(m)
}

var xxx_messageInfo_RemoveOneRoom proto.InternalMessageInfo

func (m *RemoveOneRoom) GetRoomUuid() uint64 {
	if m != nil {
		return m.RoomUuid
	}
	return 0
}

//GateService->HttpGateService同步负载
type GateBalance struct {
	NodeId               int32    `protobuf:"varint,1,opt,name=NodeId,proto3" json:"NodeId,omitempty"`
	Weigh                int32    `protobuf:"varint,2,opt,name=Weigh,proto3" json:"Weigh,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GateBalance) Reset()         { *m = GateBalance{} }
func (m *GateBalance) String() string { return proto.CompactTextString(m) }
func (*GateBalance) ProtoMessage()    {}
func (*GateBalance) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e2a4236feecff6a, []int{6}
}

func (m *GateBalance) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GateBalance.Unmarshal(m, b)
}
func (m *GateBalance) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GateBalance.Marshal(b, m, deterministic)
}
func (m *GateBalance) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GateBalance.Merge(m, src)
}
func (m *GateBalance) XXX_Size() int {
	return xxx_messageInfo_GateBalance.Size(m)
}
func (m *GateBalance) XXX_DiscardUnknown() {
	xxx_messageInfo_GateBalance.DiscardUnknown(m)
}

var xxx_messageInfo_GateBalance proto.InternalMessageInfo

func (m *GateBalance) GetNodeId() int32 {
	if m != nil {
		return m.NodeId
	}
	return 0
}

func (m *GateBalance) GetWeigh() int32 {
	if m != nil {
		return m.Weigh
	}
	return 0
}

//获取user所在playerService的NodeID
type QueryUserNodeID struct {
	UserID               uint64   `protobuf:"varint,1,opt,name=UserID,proto3" json:"UserID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *QueryUserNodeID) Reset()         { *m = QueryUserNodeID{} }
func (m *QueryUserNodeID) String() string { return proto.CompactTextString(m) }
func (*QueryUserNodeID) ProtoMessage()    {}
func (*QueryUserNodeID) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e2a4236feecff6a, []int{7}
}

func (m *QueryUserNodeID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryUserNodeID.Unmarshal(m, b)
}
func (m *QueryUserNodeID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryUserNodeID.Marshal(b, m, deterministic)
}
func (m *QueryUserNodeID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryUserNodeID.Merge(m, src)
}
func (m *QueryUserNodeID) XXX_Size() int {
	return xxx_messageInfo_QueryUserNodeID.Size(m)
}
func (m *QueryUserNodeID) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryUserNodeID.DiscardUnknown(m)
}

var xxx_messageInfo_QueryUserNodeID proto.InternalMessageInfo

func (m *QueryUserNodeID) GetUserID() uint64 {
	if m != nil {
		return m.UserID
	}
	return 0
}

type QueryUserNodeIDRet struct {
	NodeID               int32    `protobuf:"varint,1,opt,name=NodeID,proto3" json:"NodeID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *QueryUserNodeIDRet) Reset()         { *m = QueryUserNodeIDRet{} }
func (m *QueryUserNodeIDRet) String() string { return proto.CompactTextString(m) }
func (*QueryUserNodeIDRet) ProtoMessage()    {}
func (*QueryUserNodeIDRet) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e2a4236feecff6a, []int{8}
}

func (m *QueryUserNodeIDRet) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryUserNodeIDRet.Unmarshal(m, b)
}
func (m *QueryUserNodeIDRet) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryUserNodeIDRet.Marshal(b, m, deterministic)
}
func (m *QueryUserNodeIDRet) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryUserNodeIDRet.Merge(m, src)
}
func (m *QueryUserNodeIDRet) XXX_Size() int {
	return xxx_messageInfo_QueryUserNodeIDRet.Size(m)
}
func (m *QueryUserNodeIDRet) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryUserNodeIDRet.DiscardUnknown(m)
}

var xxx_messageInfo_QueryUserNodeIDRet proto.InternalMessageInfo

func (m *QueryUserNodeIDRet) GetNodeID() int32 {
	if m != nil {
		return m.NodeID
	}
	return 0
}

func init() {
	proto.RegisterType((*PlayerServiceBalance)(nil), "rpc.PlayerServiceBalance")
	proto.RegisterType((*CheckOnLineReq)(nil), "rpc.CheckOnLineReq")
	proto.RegisterType((*CheckOnLineRes)(nil), "rpc.CheckOnLineRes")
	proto.RegisterType((*UpdatePlayerList)(nil), "rpc.UpdatePlayerList")
	proto.RegisterType((*CreateTable)(nil), "rpc.CreateTable")
	proto.RegisterType((*RemoveOneRoom)(nil), "rpc.RemoveOneRoom")
	proto.RegisterType((*GateBalance)(nil), "rpc.GateBalance")
	proto.RegisterType((*QueryUserNodeID)(nil), "rpc.QueryUserNodeID")
	proto.RegisterType((*QueryUserNodeIDRet)(nil), "rpc.QueryUserNodeIDRet")
}

func init() { proto.RegisterFile("proto/rpcproto/servicestatus.proto", fileDescriptor_4e2a4236feecff6a) }

var fileDescriptor_4e2a4236feecff6a = []byte{
	// 366 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x52, 0xcf, 0x8b, 0xda, 0x40,
	0x14, 0x46, 0x93, 0x58, 0x7d, 0xf6, 0x17, 0x83, 0x94, 0x50, 0x7a, 0x90, 0xa1, 0x07, 0x4b, 0xc5,
	0x1e, 0x7a, 0xab, 0x97, 0x62, 0x02, 0x45, 0x10, 0x6d, 0xa7, 0x86, 0x85, 0xbd, 0x8d, 0xc9, 0x5b,
	0x0d, 0x9b, 0xcc, 0x64, 0x27, 0x13, 0xc1, 0xbf, 0x70, 0xff, 0xad, 0x25, 0x33, 0x51, 0x51, 0xd8,
	0x3d, 0xec, 0xed, 0x7d, 0x3f, 0xf8, 0xe6, 0xcb, 0x7b, 0x01, 0x5a, 0x28, 0xa9, 0xe5, 0x0f, 0x55,
	0xc4, 0x76, 0x28, 0x51, 0xed, 0xd3, 0x18, 0x4b, 0xcd, 0x75, 0x55, 0x4e, 0x0c, 0x47, 0x1c, 0x55,
	0xc4, 0x34, 0x84, 0xc1, 0xdf, 0x8c, 0x1f, 0x50, 0xfd, 0xb7, 0x8e, 0x19, 0xcf, 0xb8, 0x88, 0x91,
	0x7c, 0x82, 0xce, 0x52, 0x26, 0x38, 0x4f, 0xfc, 0xd6, 0xb0, 0x35, 0xf2, 0x58, 0x83, 0xc8, 0x00,
	0xbc, 0x1b, 0x4c, 0xb7, 0x3b, 0xbf, 0x6d, 0x68, 0x0b, 0xe8, 0x18, 0xde, 0x07, 0x3b, 0x8c, 0xef,
	0x57, 0x62, 0x91, 0x0a, 0x64, 0xf8, 0x40, 0x3e, 0x43, 0x37, 0xc8, 0x52, 0x14, 0xba, 0x49, 0x70,
	0xd9, 0x09, 0xd3, 0xaf, 0x57, 0xee, 0x92, 0x10, 0x70, 0xef, 0x32, 0xbe, 0x35, 0xce, 0x2e, 0x33,
	0x33, 0xfd, 0x0d, 0x1f, 0xa3, 0x22, 0xe1, 0x1a, 0x6d, 0xbf, 0x45, 0x5a, 0xea, 0x97, 0x5a, 0x45,
	0xb5, 0xc1, 0x6f, 0x0f, 0x9d, 0x91, 0xcb, 0x2c, 0xa0, 0x8f, 0x2d, 0xe8, 0x07, 0x0a, 0xb9, 0xc6,
	0x35, 0xdf, 0x64, 0x48, 0xbe, 0x40, 0xcf, 0x0c, 0x51, 0x95, 0xda, 0x80, 0x1e, 0x3b, 0x13, 0x27,
	0x75, 0x7d, 0x28, 0xb0, 0xf9, 0xba, 0x33, 0x51, 0xab, 0xb6, 0xc7, 0xb2, 0xca, 0x7d, 0xc7, 0xaa,
	0x27, 0x82, 0x50, 0x78, 0xab, 0xa4, 0xcc, 0xeb, 0x1c, 0x53, 0xc3, 0x1d, 0x3a, 0xa3, 0x1e, 0xbb,
	0xe0, 0xc8, 0x2f, 0xf0, 0xcb, 0x9d, 0xac, 0xb2, 0x24, 0x90, 0x42, 0x60, 0xac, 0x31, 0x09, 0xb2,
	0x54, 0x68, 0xe3, 0xf7, 0x4c, 0xed, 0x67, 0x75, 0xfa, 0x1d, 0xde, 0x31, 0xcc, 0xe5, 0x1e, 0x57,
	0x02, 0x99, 0x94, 0x79, 0xbd, 0x5e, 0xd6, 0x84, 0x1f, 0xd7, 0x7b, 0xc4, 0x74, 0x0a, 0xfd, 0x3f,
	0x5c, 0xbf, 0xf2, 0x92, 0xdf, 0xe0, 0xc3, 0xbf, 0x0a, 0xd5, 0x21, 0x2a, 0x51, 0x19, 0x63, 0x58,
	0x07, 0xd4, 0x68, 0x1e, 0x36, 0x2f, 0x35, 0x88, 0x8e, 0x81, 0x5c, 0x59, 0x19, 0x9e, 0x4f, 0x14,
	0x5e, 0x3c, 0x17, 0xce, 0xde, 0xdc, 0x7a, 0x93, 0xa9, 0x2a, 0xe2, 0x4d, 0xc7, 0xfc, 0x7d, 0x3f,
	0x9f, 0x02, 0x00, 0x00, 0xff, 0xff, 0x95, 0x06, 0x3a, 0x10, 0xa3, 0x02, 0x00, 0x00,
}
