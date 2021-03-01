package gateservice

import (
	"encoding/binary"
	"errors"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/network/processor"
	"github.com/duanhf2012/origin/node"
	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/sysservice/tcpservice"
	"github.com/duanhf2012/origin/util/timer"
	"sunserver/common/global"
	"sunserver/common/proto/rpc"
	"time"
)


var gateService GateService

func init() {
	node.Setup(&tcpservice.TcpService{})
	node.Setup(&gateService)
}

type GateService struct {
	service.Service

	processor  processor.IRawProcessor
	tcpService *tcpservice.TcpService
	router     *Router

	rawPackInfo processor.PBRawPackInfo
}

func (gateService *GateService) OnInit() error {
	gateService.router = &Router{}
	_, err := gateService.AddModule(gateService.router)
	if err != nil {
		return err
	}
	gateService.processor = &processor.PBRawProcessor{}
	//注册监听客户连接断开事件
	gateService.processor.SetDisConnectedHandler(gateService.router.OnDisconnected)
	//注册监听客户连接事件
	gateService.processor.SetConnectedHandler(gateService.router.OnConnected)
	//注册监听消息类型MsgType_MsgReq，并注册回调
	gateService.processor.SetRawMsgHandler(gateService.router.RouterMessage)
	//将protobuf消息处理器设置到TcpService服务中
	gateService.tcpService = node.GetService(global.TcpService).(*tcpservice.TcpService)
	gateService.tcpService.SetProcessor(gateService.processor, gateService.GetEventHandler())
	gateService.NewTicker(time.Second*3, gateService.SyncBalance)
	//gateService.NewTicker(time.Second*5, gateService.PrintMemPool)

	//性能监控
	gateService.OpenProfiler()
	gateService.GetProfiler().SetOverTime(time.Millisecond * 100)
	gateService.GetProfiler().SetMaxOverTime(time.Second * 10)

	//注册原始RPC
	gateService.RegRawRpc(global.RawRpcMsgDispatch,&RpcOnRecvCallBack{})
	gateService.RegRawRpc(global.RawRpcCloseClient,&RpcOnCloseCallBack{})
	return nil
}




func (gateService *GateService) SetEventChannel(channelNum int) {
	gateService.GetEventProcessor().SetEventChannel(channelNum)
}

func (gateService *GateService) SetRawProcessor(processor processor.IRawProcessor) {
	gateService.processor = processor
}

func (gateService *GateService) SetTcpGateService(tcpService *tcpservice.TcpService) {
	gateService.tcpService = tcpService
}

func (gateService *GateService) Close(clientId uint64) {
	gateService.tcpService.Close(clientId)
}

func (gateService *GateService) SyncBalance(timer *timer.Ticker) {
	var balance rpc.GateBalance
	balance.Weigh = int32(gateService.tcpService.GetConnNum())
	balance.NodeId = int32(node.GetNodeId())

	gateService.CastGo("HttpGateService.RPC_SetTcpGateBalance", &balance)
}

func (gateService *GateService) PrintMemPool(timer *timer.Ticker) {
	//log.Release("lfy----------------RPC-----------------------[%d]", network.GetUseMemCount())
}

/*func (gateService *GateService) RPCCloseClient(byteBuff []byte)  {
	var rawInput global.RawInputArgs
	clientId,err := rawInput.ParseUint64(byteBuff)
	if err!=nil {
		log.Error("msg is error:%s!",err.Error())
		return
	}

	gateService.Close(clientId)
}
*/
/*type RawRpcCallBack interface {
	Unmarshal(data []byte) (interface{},error)
	CB(data interface{})
}*/

/*func (gateService *GateService) RpcDispatch(byteBuff []byte) {
	var rawInput global.RawInputArgs
	msgType,clientId,msgBuff,err := rawInput.ParseMsg(byteBuff)
	if err != nil {
		log.Error("msg is error:%s!",err.Error())
		return
	}

	gateService.rawPackInfo.SetPackInfo(msgType,msgBuff)
	err = gateService.tcpService.SendMsg(clientId, &gateService.rawPackInfo)
	if err != nil {
		log.Error("SendRawMsg fail:%+v!", err)
	}
}*/

type RpcOnCloseCallBack struct {
}

func (cb *RpcOnCloseCallBack) Unmarshal(data []byte) (interface{}, error) {
	return data, nil
}

func (cb *RpcOnCloseCallBack) CB(data interface{}) {
	var rawInput []byte = data.([]byte)

	if len(rawInput) < 8 {
		err := errors.New("parseMsg error")
		log.Error(err.Error())
		return
	}

	clientId := binary.BigEndian.Uint64(rawInput)
	gateService.Close(clientId)
}

type RpcOnRecvCallBack struct {
}

func (cb *RpcOnRecvCallBack) Unmarshal(data []byte) (interface{}, error) {
	var rawInput global.RawInputArgs

	err := rawInput.ParseMsg(data)
	if err != nil {
		log.Error("parse message is error:%s!", err.Error())
		return nil, err
	}
	//clientIdList := rawInput.GetClientIdList()

	gateService.rawPackInfo.SetPackInfo(rawInput.GetMsgType(),rawInput.GetMsg())
	rawInput.SetProtoMsg(&gateService.rawPackInfo)

	return &rawInput, err
}

func (cb *RpcOnRecvCallBack) CB(data interface{}) {
	args := data.(*global.RawInputArgs)

	clientIdList := args.GetClientIdList()
	if len(clientIdList) != 1 {
		//收消息只可能有一个clientid
		log.Release("RpcOnRecvCallBack receive client len[%d] > 1", len(clientIdList))
		return
	}

	err := gateService.tcpService.SendMsg(clientIdList[0], args.GetProtoMsg())
	if err != nil {
		log.Error("SendRawMsg fail:%+v!", err)
	}

}
