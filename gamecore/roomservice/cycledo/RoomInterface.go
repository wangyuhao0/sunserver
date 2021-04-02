package cycledo

import (
	"github.com/duanhf2012/origin/util/timer"
	"github.com/golang/protobuf/proto"
	"sunserver/common/entity"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/common"
	"sunserver/gamecore/roomservice/room"
	"time"
)

//进行循环引用解除

type RoomInterface struct {
	RS roomService
}

func New(rs roomService) *RoomInterface {
	return &RoomInterface{
		RS: rs,
	}
}

type roomService interface {
	TimerAfter(roomType int32, uuid string, d time.Duration, cb func(timer *timer.Timer)) *timer.Timer

	TimerTicker(roomType int32, uuid string, d time.Duration, cb func(timer *timer.Ticker)) *timer.Ticker

	NewRoom() *room.Room

	NewPlayerInfo(playerInfoPb *msg.PlayerInfo) *entity.PlayerInfo

	SetRoom(roomUuid string, roomType int32, room *room.Room)

	GetProxy() *common.GateProxyModule

	GetRoom(roomUuid string, roomType int32) (*room.Room, bool)

	SendMsg(clientId uint64, msgType msg.MsgType, msg proto.Message)

	RadioPlayerInfo(room *room.Room)

	RemoveRoom(roomUuid string, roomType int32)

	PackRoom(room *room.Room) *msg.Room

	SimpleRoomList(roomType int32) []*msg.SimpleRoom

	CheckCreateRoom(userId uint64) bool

	RemoveRoomStatus(userId uint64)

	AddRoomStatus(userId uint64, roomUuid string)
}

func (ri *RoomInterface) TimerAfter(roomType int32, uuid string, d time.Duration, cb func(timer *timer.Timer)) *timer.Timer {
	return ri.RS.TimerAfter(roomType, uuid, d, cb)
}

func (ri *RoomInterface) TimerTicker(roomType int32, uuid string, d time.Duration, cb func(timer *timer.Ticker)) *timer.Ticker {
	return ri.RS.TimerTicker(roomType, uuid, d, cb)

}

func (ri *RoomInterface) NewRoomRi() *room.Room {
	return ri.RS.NewRoom()
}

func (ri *RoomInterface) NewPlayerInfoRi(playerInfoPb *msg.PlayerInfo) *entity.PlayerInfo {
	return ri.RS.NewPlayerInfo(playerInfoPb)
}

func (ri *RoomInterface) SetRoomRi(roomUuid string, roomType int32, room *room.Room) {
	ri.RS.SetRoom(roomUuid, roomType, room)
}

func (ri *RoomInterface) GetProxyRi() *common.GateProxyModule {
	return ri.RS.GetProxy()
}

func (ri *RoomInterface) GetRoomRi(roomUuid string, roomType int32) (*room.Room, bool) {
	return ri.RS.GetRoom(roomUuid, roomType)
}

func (ri *RoomInterface) SendMsgRi(clientId uint64, msgType msg.MsgType, msg proto.Message) {
	ri.RS.SendMsg(clientId, msgType, msg)
}

func (ri *RoomInterface) PackRoomRi(room *room.Room) *msg.Room {
	return ri.RS.PackRoom(room)
}

func (ri *RoomInterface) RadioPlayerInfoRi(room *room.Room) {
	ri.RS.RadioPlayerInfo(room)
}

func (ri *RoomInterface) RemoveRoomRi(roomUuid string, roomType int32) {
	ri.RS.RemoveRoom(roomUuid, roomType)
}

func (ri *RoomInterface) SimpleRoomList(roomType int32) []*msg.SimpleRoom {
	return ri.RS.SimpleRoomList(roomType)
}

func (ri *RoomInterface) RemoveRoomStatus(userId uint64) {
	ri.RS.RemoveRoomStatus(userId)
}

func (ri *RoomInterface) CheckCreateRoom(userId uint64) bool {
	return ri.RS.CheckCreateRoom(userId)
}

func (ri *RoomInterface) AddRoomStatus(userId uint64, roomUuid string) {
	ri.RS.AddRoomStatus(userId, roomUuid)
}
