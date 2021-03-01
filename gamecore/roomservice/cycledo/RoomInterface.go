package cycledo

import (
	"github.com/golang/protobuf/proto"
	"sunserver/common/entity"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/common"
)

//进行循环引用解除

type RoomInterface struct {

	RS roomService
}

func New(rs roomService)  *RoomInterface{
	return &RoomInterface{
		RS: rs,
	}
}

type roomService interface {

	NewRoom() *common.Room

	NewPlayerInfo(playerInfoPb *msg.PlayerInfo) *entity.PlayerInfo

	SetRoom(roomUuid string, room *common.Room)

	GetProxy() *common.GateProxyModule

    GetRoom(roomUuid string) (*common.Room, bool)

    SendMsg(clientId uint64, msgType msg.MsgType, msg proto.Message)

    RadioPlayerInfo(room *common.Room)

	RemoveRoom(roomUuid string)

	PackRoom(room *common.Room) *msg.Room
}

func (ri *RoomInterface) NewRoomRi() *common.Room {
	return ri.RS.NewRoom()
}

func (ri *RoomInterface) NewPlayerInfoRi(playerInfoPb *msg.PlayerInfo) *entity.PlayerInfo {
	return ri.RS.NewPlayerInfo(playerInfoPb)
}

func (ri *RoomInterface) SetRoomRi(roomUuid string, room *common.Room) {
	ri.RS.SetRoom(roomUuid,room)
}

func (ri *RoomInterface) GetProxyRi() *common.GateProxyModule {
	return ri.RS.GetProxy()
}


func (ri *RoomInterface) GetRoomRi(roomUuid string) (*common.Room, bool) {
	return ri.RS.GetRoom(roomUuid)
}


func (ri *RoomInterface) SendMsgRi(clientId uint64, msgType msg.MsgType, msg proto.Message) {
	ri.RS.SendMsg(clientId, msgType, msg)
}

func (ri *RoomInterface) PackRoomRi(room *common.Room) *msg.Room {
	return ri.RS.PackRoom(room)
}


func (ri *RoomInterface) RadioPlayerInfoRi(room *common.Room)  {
	ri.RS.RadioPlayerInfo(room)
}

func (ri *RoomInterface) RemoveRoomRi(roomUuid string) {
	ri.RS.RemoveRoom(roomUuid)
}

