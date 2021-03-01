package msghandler

import (
	"github.com/duanhf2012/origin/log"
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"sunserver/common/proto/rpc"
	"sunserver/gamecore/queueservice/cycledo"
)

func handlerAddQueue(qi *cycledo.QueueInterface,clientId uint64,message proto.Message)  {
	req := message.(*msg.MsgAddQueueReq)
	log.Release("addQueueReq:%d",clientId)
	roomUuid := req.GetRoomUuid()
	userId := req.GetUserId()
	//向room获取信息
	var getRoomReq rpc.GetRoomReq
	getRoomReq.RoomUuid = roomUuid


	err := qi.GetRpcHandlerQi().AsyncCall("RoomService.RPC_GetPbRoom",&getRoomReq,func(res *rpc.GetRoomRes,err error) {
		//打包好的房间加入队列
		ownerId := res.GetRoom().GetOwner().GetUserId()
		room := qi.PackRoomQi(res)
		//判断是否为房主
		if ownerId!= userId{
			//说明不是房主
			queueRes := msg.MsgAddQueueRes{Ret: msg.ErrCode_NotOwner}
			room.SendToClient(clientId,msg.MsgType_AddQueueRes,&queueRes)
			return
		}
		qi.AddRoomQi(room)
		flag := qi.AddQueueQi(room)
		if flag {
			//加入成功
			//向房间所有人推送
			queueRes := msg.MsgAddQueueRes{Ret: msg.ErrCode_OK}
			room.SendToClient(room.GetOwner().GetClientId(),msg.MsgType_AddQueueRes,&queueRes)
			//给其他人推送
			for _, client := range room.GetOtherClients() {
				room.SendToClient(client.GetClientId(),msg.MsgType_AddQueueRes,&queueRes)
			}
		}else {
			//失败
			queueRes := msg.MsgAddQueueRes{Ret: msg.ErrCode_InterNalError}
			room.SendToClient(room.GetOwner().GetClientId(),msg.MsgType_AddQueueRes,&queueRes)
			//给其他人推送
			for _, client := range room.GetOtherClients() {
				room.SendToClient(client.GetClientId(),msg.MsgType_AddQueueRes,&queueRes)
			}
			return
		}
	})
	if err != nil {
		//失败
		log.Release("房间加入队列失败----")
		return
	}
	return
}

