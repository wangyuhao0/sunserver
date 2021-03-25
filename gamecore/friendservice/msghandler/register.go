package msghandler

import (
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/friendservice/cycledo"
)

type CallBack func(ti *cycledo.FriendInterface, clientId uint64, msg proto.Message)

func OnRegisterMessage(register func(msgType msg.MsgType, message proto.Message, cb CallBack)) {
}
