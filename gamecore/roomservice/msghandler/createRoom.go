package msghandler

import (
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/util/uuid"
	"github.com/golang/protobuf/proto"
	"strconv"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/roomservice/cycledo"
)

func handlerClientCreateRoom(ri *cycledo.RoomInterface, clientId uint64, message proto.Message) {
	log.Release("roomService-createRoom")
	msgReq := message.(*msg.MsgCreateRoomReq)
	playerInfoPb := msgReq.GetPlayerInfo()
	roomType := msgReq.GetRoomType()
	userId := playerInfoPb.GetUserId()
	//rClientId := msgReq.GetClientId()
	//登录平台了 然后创建房间放入
	newRoom := ri.NewRoomRi()
	playerInfo := ri.NewPlayerInfoRi(playerInfoPb)
	//设置为房主
	playerInfo.SetOwner(true)
	//设置位置
	playerInfo.SetSeatNum(1)

	playerInfo.SetClientId(clientId)

	uuid := uuid.Rand().HexEx()
	newRoom.OnInit(ri.GetProxyRi(), uuid, "testRoom"+strconv.FormatUint(userId, 10), 1, playerInfo, roomType)
	ri.SetRoomRi(uuid, roomType, newRoom)
	//通知客户端
	newRoom.SendToClient(clientId, msg.MsgType_CreateRoomRes, &msg.MsgCreateRoomRes{Ret: msg.ErrCode_OK, RoomUuid: uuid})

}
