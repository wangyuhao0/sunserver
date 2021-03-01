package msghandler

import (
	"github.com/duanhf2012/origin/log"
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/tableservice/cycledo"
)

func handlerClientAddTable(ti *cycledo.TableInterface,clientId uint64, message proto.Message) {
	msgReq := message.(*msg.MsgAddTableReq)

	userId:= msgReq.GetUserId()
	tableUuid := msgReq.GetTableUuid()
	flag := msgReq.GetFlag()
	log.Release("tableService-addTable,userId-%d,tableUuid-%s,flag-%d",userId,tableUuid,flag)
	ti.AddTableTi(clientId,tableUuid,flag)
}

