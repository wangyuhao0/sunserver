package msghandler

import (
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/roomservice/cycledo"
)

type CallBack func(mrs *cycledo.RoomInterface,clientId uint64, msg proto.Message)

func OnRegisterMessage(register func(msgType msg.MsgType, message proto.Message, cb CallBack)) {
	register(msg.MsgType_CreateRoomReq, &msg.MsgCreateRoomReq{}, handlerClientCreateRoom)
	register(msg.MsgType_AddRoomReq, &msg.MsgAddRoomReq{}, handlerClientAddRoom)
	register(msg.MsgType_QuitRoomReq, &msg.MsgQuitRoomReq{}, handlerClientQuitRoom)
}
