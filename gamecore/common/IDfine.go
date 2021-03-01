package common

import (
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
)

type ISender interface {
	SendToClient(clientId uint64, msgType msg.MsgType, msg proto.Message) error
}
