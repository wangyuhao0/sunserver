package msghandler

import (
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/playerservice/player"
)

func ping(player *player.Player,message proto.Message) {
	player.Ping()
	player.SendToClient(player.GetClientId(),msg.MsgType_Pong,&msg.MsgPong{})
}
