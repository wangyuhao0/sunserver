package msghandler

import (
	"github.com/duanhf2012/origin/log"
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/roomservice/cycledo"
)

func handlerClientListRoom(ri *cycledo.RoomInterface, clientId uint64, message proto.Message) {
	log.Release("roomService-listRoom")

	msgReq := message.(*msg.MsgRoomListReq)
	roomType := msgReq.GetRoomType()

	roomList := ri.SimpleRoomList(roomType)
	// 发送
	ri.SendMsgRi(clientId, msg.MsgType_RoomListRes, &msg.MsgRoomListRes{Ret: msg.ErrCode_OK, RoomList: roomList})

}
