package common

import (
	"fmt"
	"github.com/duanhf2012/origin/log"
	originrpc "github.com/duanhf2012/origin/rpc"
	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/sysservice/tcpservice"
	"github.com/golang/protobuf/proto"
	"sunserver/common/global"
	"sunserver/common/proto/msg"
)

type GateProxyModule struct {
	service.Module
}

func NewGateProxyModule() *GateProxyModule {
	return &GateProxyModule{}
}

func (gate *GateProxyModule) OnInit() error{
	return nil
}

func (gate *GateProxyModule) SendToClient(clientId uint64, msgType msg.MsgType, msg proto.Message) error {
	//1.获取GateServiceNodeId
	var err error
	nodeId := tcpservice.GetNodeId(clientId)
	if nodeId < 0 || nodeId > tcpservice.MaxNodeId {
		err = fmt.Errorf("nodeid is error %d", nodeId)
		log.Error(err.Error())
		return err
	}

	//2.组装返回消息
	var msgBuff []byte
	if msg != nil {
		msgBuff, err = proto.Marshal(msg)
		if err != nil {
			log.Error("Marshal fail,msgType %d clientId %d.",msgType,clientId)
			return err
		}
	}

	var rawInputArgs global.RawInputArgs
	rawInputArgs.SetMsg(uint16(msgType),clientId,msgBuff)
	err = gate.RawGoNode(originrpc.RpcProcessorGoGoPB,nodeId,global.RawRpcMsgDispatch,global.GateService,&rawInputArgs)
	if err != nil {
		log.Error("RawGoNode fail :%s,msgType %d clientId %d.",err.Error(),msgType,clientId)
	}
	return err
}

