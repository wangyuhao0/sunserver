package msghandler

import (
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/queueservice/cycledo"
)

type CallBack func(qi *cycledo.QueueInterface,clientId uint64, msg proto.Message)

func OnRegisterMessage(register func(msgType msg.MsgType, message proto.Message, cb CallBack)) {
	register(msg.MsgType_AddQueueReq, &msg.MsgAddQueueReq{}, handlerAddQueue)
	register(msg.MsgType_QuitQueueReq, &msg.MsgQuitQueueReq{}, handlerQuitQueue)
}
