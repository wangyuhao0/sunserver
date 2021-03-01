package msghandler

import (
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/tableservice/cycledo"
)

type CallBack func(ti *cycledo.TableInterface,clientId uint64, msg proto.Message)

func OnRegisterMessage(register func(msgType msg.MsgType, message proto.Message, cb CallBack)) {
	register(msg.MsgType_AddTableReq, &msg.MsgAddTableReq{}, handlerClientAddTable)
}
