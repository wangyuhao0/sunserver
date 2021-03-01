package msghandler

import (
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/playerservice/player"
)

type CallBack func(player *player.Player,  msg proto.Message)

func OnRegisterMessage(register func(msgType msg.MsgType, message proto.Message, cb CallBack)) {
	register(msg.MsgType_Ping, nil, ping)
	register(msg.MsgType_ClientSyncTimeReq, &msg.MsgClientSyncTimeReq{}, handlerClientSyncTimeMsg)
	//playerservice.CallBack()
}
