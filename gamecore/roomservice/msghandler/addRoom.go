package msghandler

import (
	"github.com/duanhf2012/origin/log"
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/roomservice/cycledo"
)

func handlerClientAddRoom(ri *cycledo.RoomInterface,clientId uint64, message proto.Message) {
	log.Release("roomService-createRoom")

	msgReq := message.(*msg.MsgAddRoomReq)
	roomUuid := msgReq.GetRoomUuid()
	info := msgReq.GetPlayerInfo()
	//登录平台了 然后创建房间放入
	room, ok := ri.GetRoomRi(roomUuid)
	if !ok {
		log.Release("房间不存在%s", room)
		//不存在
		ri.SendMsgRi(clientId, msg.MsgType_AddRoomRes, &msg.MsgAddRoomRes{Ret: msg.ErrCode_RoomIdNotExist})
		return
	}
	otherClients := room.GetOtherClients()
	for _, v := range otherClients {
		if clientId == v.GetClientId(){
			log.Release("重复加入房间")
			return
		}
	}
	//平衡一下房间的平均分
	room.SetAvgRank(room.GetAvgRank()* (uint64(room.GetRoomClientNum())+info.GetRank())/uint64(room.GetRoomClientNum()+1))
	//放入道理吗
	playerInfo := ri.NewPlayerInfoRi(info)
	//查看房间有几个人 以及其他人的座位号
	playerInfo.SetSeatNum(room.GetSeatNum())

	playerInfo.SetOwner(false)

	playerInfo.SetClientId(clientId)
	//增加人数
	num := room.GetRoomClientNum()
	room.SetRoomClientNum(num+1)
	// 增加用户
	otherClients = append(otherClients, playerInfo)
	room.SetOtherClients(otherClients)
	log.Release("向房间广播---roomID:%d,%s-%d加入房间", roomUuid, info.GetNickName(), info.GetClientId())
	//发送加入成功
	//packRoomRi := ri.PackRoomRi(room)
	ri.SendMsgRi(clientId, msg.MsgType_AddRoomRes, &msg.MsgAddRoomRes{Ret: msg.ErrCode_OK})
	//先开辟一个空间 把用户全部放进去 包括房主
	ri.RadioPlayerInfoRi(room)

}
