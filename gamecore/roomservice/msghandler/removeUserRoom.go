package msghandler

import (
	"github.com/duanhf2012/origin/log"
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/roomservice/cycledo"
)

func handlerClientRemoveUserRoom(ri *cycledo.RoomInterface, clientId uint64, message proto.Message) {
	log.Release("roomService-addRoom")

	msgReq := message.(*msg.MsgRemoveUserReq)
	roomUuid := msgReq.GetRoomUuid()
	roomType := msgReq.GetRoomType()
	removeUserId := msgReq.GetRemoveUserId()
	//校验执行者权限
	room, b := ri.GetRoomRi(roomUuid, roomType)
	if !b {
		log.Release("移除失败 房间不存在-%d", removeUserId)
		ri.SendMsgRi(clientId, msg.MsgType_RemoveUserRes, &msg.MsgRemoveUserRes{Ret: msg.ErrCode_RoomIdNotExist})
		return
	}
	if room.GetOwner().GetClientId() != clientId {
		log.Release("移除失败 不是房主-%d", removeUserId)
		ri.SendMsgRi(clientId, msg.MsgType_RemoveUserRes, &msg.MsgRemoveUserRes{Ret: msg.ErrCode_NotOwner})
		return
	}
	if room.GetOwner().GetUserId() == removeUserId {
		log.Release("房主移除自己 不可以-%d", removeUserId)
		ri.SendMsgRi(clientId, msg.MsgType_RemoveUserRes, &msg.MsgRemoveUserRes{Ret: msg.ErrCode_InterNalError})
		return
	}
	//判断是否存在这个UserId
	clients := room.GetOtherClients()
	var flag = false
	var j = 0
	for i := 0; i < len(clients); i++ {
		info := clients[i]
		if info.GetUserId() == removeUserId {
			flag = true
			j = i
		}
	}
	if flag {
		log.Release("移除用户不在房间-%d", removeUserId)
		ri.SendMsgRi(clientId, msg.MsgType_RemoveUserRes, &msg.MsgRemoveUserRes{Ret: msg.ErrCode_InterNalError})
		return
	}
	//进行正常逻辑
	//----------------------1-----------------------------
	info := clients[j]
	room.SetOtherClients(append(clients[:j], clients[j+1:]...))
	//然后修改房间基础信息
	room.SetAvgRank((room.GetAvgRank()*(uint64(room.GetRoomClientNum())) - info.GetRank()) / uint64(room.GetRoomClientNum()-1))
	room.SetRoomClientNum(room.GetRoomClientNum() - 1)
	//重置用户房间权限
	ri.RemoveRoomStatus(removeUserId)
	packRoom := ri.PackRoomRi(room)
	//1. 移除用户 返回给房主
	ri.SendMsgRi(clientId, msg.MsgType_RemoveUserRes, &msg.MsgRemoveUserRes{Ret: msg.ErrCode_OK, Room: packRoom})
	//2. 广播给其他用户
	ri.RadioPlayerInfoRi(clientId, room)
	//3. 给被移除用户推送踢出房间
	if info.GetClientId() > 0 {
		ri.SendMsgRi(info.GetClientId(), msg.MsgType_UserRemovedRes, &msg.MsgUserRemovedRes{RoomUuid: roomUuid, RoomType: roomType})
	}

}
