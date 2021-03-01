package msghandler

import (
	"github.com/duanhf2012/origin/log"
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/queueservice/cycledo"
)

func handlerQuitQueue(qi *cycledo.QueueInterface,clientId uint64,message proto.Message)  {
	req := message.(*msg.MsgQuitQueueReq)
	roomUuid := req.GetRoomUuid()
	userId := req.GetUserId()
	log.Release("quitQueueReq:%s",roomUuid)
	room := qi.GetRoomQi(roomUuid)
	ownerId := room.GetOwner().GetUserId()
	//判断是否为房主
	if ownerId!= userId{
		//说明不是房主
		queueRes := msg.MsgAddQueueRes{Ret: msg.ErrCode_NotOwner}
		room.SendToClient(clientId,msg.MsgType_QuitQueueRes,&queueRes)
		return
	}

	qi.QuitQueueQi(room)
	//通知用户
	queueRes := msg.MsgQuitQueueRes{Ret: msg.ErrCode_OK}
	room.SendToClient(room.GetOwner().GetClientId(),msg.MsgType_QuitQueueRes,&queueRes)
	//给其他人推送
	for _, client := range room.GetOtherClients() {
		room.SendToClient(client.GetClientId(),msg.MsgType_QuitQueueRes,&queueRes)
	}
	qi.RemoveRoomQi(roomUuid)
}

