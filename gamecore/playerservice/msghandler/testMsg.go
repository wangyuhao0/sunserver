package msghandler

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/playerservice/player"
	"time"
)

func handlerClientSyncTimeMsg(player *player.Player, message proto.Message) {
	receiveInfo := message.(*msg.MsgClientSyncTimeReq)
	player.DealClientSyncTimeMsg(receiveInfo)

	sendMsg := msg.MsgClientSyncTimeRes{Result: fmt.Sprintf("my cid is %d", player.GetClientId()), Test: time.Now().Unix()}
	player.SendToClient(player.GetClientId(), msg.MsgType_ClientSyncTimeRes, &sendMsg)
}
