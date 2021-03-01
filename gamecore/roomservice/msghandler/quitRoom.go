package msghandler

import (
	"github.com/duanhf2012/origin/log"
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/roomservice/cycledo"
)

func handlerClientQuitRoom(ri *cycledo.RoomInterface,clientId uint64, message proto.Message) {
	log.Release("roomService-quitRoom")

	msgReq := message.(*msg.MsgQuitRoomReq)
	roomUuid := msgReq.GetRoomUuid()
	room, ok := ri.GetRoomRi(roomUuid)

	if ok {
		if room.GetRoomClientNum()==1{
			//说明单人房间直接销毁即可
			ri.RemoveRoomRi(roomUuid)
			room.SendToClient(clientId,msg.MsgType_QuitRoomRes,&msg.MsgQuitRoomRes{Ret: msg.ErrCode_OK})
			return
		}
		room.SetRoomClientNum(room.GetRoomClientNum()-1)
		//比对分配是否为房主 如果是房主 需要重置房主
		ownerId := room.GetOwner().GetClientId()
		room.SendToClient(clientId,msg.MsgType_QuitRoomRes,&msg.MsgQuitRoomRes{Ret: msg.ErrCode_OK})
		if ownerId == clientId{
			//说明是房主 进行房主分配
			rank := room.GetOwner().GetRank()
			otherClients := room.GetOtherClients()
			newOwner := otherClients[0]
			//设置房主
			newOwner.SetOwner(true)
			room.SetOwner(newOwner)
			//otherClients 重新赋值
			room.SetOtherClients(otherClients[1:])
			//刷新平均分
			room.SetAvgRank(room.GetAvgRank()* (uint64(room.GetRoomClientNum()))-rank/uint64(room.GetRoomClientNum()+1))
			//通知
			ri.RadioPlayerInfoRi(room)
			return
		}
		//不是房主
		otherClients := room.GetOtherClients()
		//比对是哪个用户
		i := -1
		rank := uint64(0)
		for j, info := range otherClients {
			if info.GetClientId() == clientId {
				//匹配上了
				rank = info.GetRank()
				i = j
			}
		}
		if i<0 {
			//说明没有匹配上 直接不回应
			log.Release("退出人不是该房间用户%d,房间:%s",clientId,room.GetUUid())
			return
		}
		//重新赋值
		room.SetOtherClients(append(otherClients[:i],otherClients[i+1:]...))
		//平均分
		room.SetAvgRank(room.GetAvgRank()* (uint64(room.GetRoomClientNum()))-rank/uint64(room.GetRoomClientNum()+1))
		//广播
		ri.RadioPlayerInfoRi(room)
	}


}

