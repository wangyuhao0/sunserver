package tableService

import (
	"fmt"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/node"
	"github.com/duanhf2012/origin/service"
	"github.com/golang/protobuf/proto"
	"sunserver/common/global"
	"sunserver/common/proto/msg"
	"sunserver/common/proto/rpc"
	"sunserver/gamecore/common"
	"sunserver/gamecore/tableservice/cycledo"
	"sunserver/gamecore/tableservice/msghandler"
	"sunserver/gamecore/tableservice/table"
	"sync"
	"time"
)

var tableService TableService
var tablePool sync.Pool

func init() {
	node.Setup(&tableService)
}

type protoMsg struct {
	ref bool
	msg proto.Message
}

func (m *protoMsg) Reset() {
	m.msg.Reset()
}

func (m *protoMsg) IsRef() bool {
	return m.ref
}

func (m *protoMsg) Ref() {
	m.ref = true
}

func (m *protoMsg) UnRef() {
	m.ref = false
}

type RegMsgInfo struct {
	protoMsg    *protoMsg
	msgPool     *sync.Pool
	msgCallBack msghandler.CallBack
}

func (r *RegMsgInfo) NewMsg() *protoMsg {
	pMsg := r.msgPool.Get().(*protoMsg)
	return pMsg
}

func (r *RegMsgInfo) ReleaseMsg(msg *protoMsg) {
	r.msgPool.Put(msg)
}

func RegisterMessage(msgType msg.MsgType, message proto.Message, cb msghandler.CallBack) {
	var regMsgInfo RegMsgInfo
	regMsgInfo.protoMsg = &protoMsg{}
	regMsgInfo.protoMsg.msg = message
	regMsgInfo.msgPool = &sync.Pool{
		New: func() interface{} {
			protoMsg := protoMsg{}
			protoMsg.msg = proto.Clone(regMsgInfo.protoMsg.msg)
			return &protoMsg
		},
	}
	regMsgInfo.msgCallBack = cb
	tableService.mapRegisterMsg[msgType] = &regMsgInfo
}

type TableService struct {
	service.Service

	mapRegisterMsg map[msg.MsgType]*RegMsgInfo //消息注册
	table           map[string]*table.Table   //在线房间数 还有房主信息

	cycleDo *cycledo.TableInterface


	gateProxy *common.GateProxyModule //网关代理
}

func (ts *TableService) OnInit() error {
	//1.初始化变量与模块
	ts.table = make(map[string]*table.Table, 3000)
	ts.mapRegisterMsg = make(map[msg.MsgType]*RegMsgInfo, 512)
	tablePool = sync.Pool{New: func() interface{} {
		return &table.Table{}
	}}

	ts.gateProxy = common.NewGateProxyModule()
	ts.AddModule(ts.gateProxy)
	//2.设置注册函数回调
	msghandler.OnRegisterMessage(RegisterMessage)


	// 注册原始套接字回调
	ts.RegRawRpc(global.RawRpcOnRecv, &RpcOnRecvCallBack{})

	ts.OpenProfiler()
	ts.GetProfiler().SetOverTime(time.Millisecond * 50)
	ts.GetProfiler().SetMaxOverTime(time.Second * 10)

	tableInterface := cycledo.New(ts)
	ts.cycleDo = tableInterface

	return nil
}

func (ts *TableService) RPC_CreateTable(createTable *rpc.CreateTable) error{
	tableUuid := createTable.GetTableUuid()
	tableType := createTable.GetTableType()
	playerNum := createTable.GetPlayerNum()
	roomUuidList := createTable.GetRoomUuidList()
	shouldConnectedClintList := createTable.GetShouldConnectedClintList()
	log.Release("从QueueService过来的创建对局,uuid:%s,房间类型:%d,人数:%d",tableUuid,tableType,playerNum)
	existFlag := ts.CheckTableExist(createTable.GetTableUuid())

	if existFlag {
		//旧房间置空
		log.Release("对局:%s,已经创建,重新初始化--",tableUuid)
		return fmt.Errorf("对局已经创建了")
	}
	log.Release("创建对局:%s",tableUuid)
	ts.Create(tableUuid,tableType,playerNum,roomUuidList,shouldConnectedClintList)
	return nil
}

func (ts *TableService) AddTable(clientId uint64, tableUuid string,addFlag int32) {
	// 进入table
	log.Release("进入TableService-AddTable")
	log.Release("加入table,客户端id:%d,房间号:%s,状态:%d",clientId,tableUuid,addFlag)

	flag := ts.CheckTableExist(tableUuid)
	if flag {
		//加入房间
		numFlag := ts.CheckTableClientNum(tableUuid)
		if !numFlag {
			log.Release("对局人数不够")
			ts.Add(clientId,tableUuid,addFlag)
		}else {
			log.Release("对局出现错误")
			ts.AddError(clientId,msg.ErrCode_TableIsEnough)
		}
	}else {
		// 通知加入失败 然后给房间已有用户进行推送房间已销毁
		ts.AddError(clientId,msg.ErrCode_TableIdNotExist)
	}
}

func (ts *TableService) Add(clientId uint64, tableUuid string,addFlag int32)  {
	log.Release("加入对局%s,%d",tableUuid,clientId)
	table := ts.table[tableUuid]
	connectFlag := table.CheckClientCanConnect(clientId)
	if !connectFlag {
		//说明他不能连接
		ts.gateProxy.SendToClient(clientId,msg.MsgType_AddTableRes,&msg.MsgAddTableRes{Ret: msg.ErrCode_InterNalError})
		return
	}

	if addFlag==0 {
		//说明拒绝了 推送给其他用户退出该对局 重新加入匹配
		list:= table.GetShouldConnectedClientList()
		for _, clientId := range list {
			ts.gateProxy.SendToClient(clientId,msg.MsgType_ClientConnectedStatus,&msg.MsgClientConnectedStatusRes{Ret: msg.ErrCode_OK})
		}
		//移除该对局
		ts.RemoveTable(table.GetTableUuid())
		return

	}

	clientList := table.GetClientList()
	clientList = append(clientList,clientId)
	table.SetClientList(clientList)
	//设置连接数
	table.SetClientConnectedNum(table.GetClientConnectedNum()+1)


	ts.gateProxy.SendToClient(clientId,msg.MsgType_AddTableRes,&msg.MsgAddTableRes{Ret: msg.ErrCode_OK})

	//需要去判断是否人全了 全了发一个消息切换场景
	if table.GetClientConnectedNum() == table.GetPlayerNum() {
		list:= table.GetClientList()
		for _, clientId := range list {
			ts.gateProxy.SendToClient(clientId,msg.MsgType_ClientConnectedStatus,&msg.MsgClientConnectedStatusRes{Ret: msg.ErrCode_OK})
		}
	}
}

func (ts *TableService) AddError(clientId uint64,code msg.ErrCode)  {

	ts.gateProxy.SendToClient(clientId,msg.MsgType_AddTableRes,&msg.MsgAddTableRes{Ret: code})

}

//检查房间是否创建了
func (ts *TableService) CheckTableExist(tableUuid string) bool {
	log.Release("tableService-CheckOnline")
	_,ok := ts.table[tableUuid]
	return ok
}

//检查房间人数是否满了
func (ts *TableService) CheckTableClientNum(tableUuid string) bool {
	log.Release("tableService-CheckTableClientNum")
	table,_ := ts.table[tableUuid]
	return table.GetClientConnectedNum()==table.GetPlayerNum()
}

func (ts *TableService) Create(tableUuid string,tableType int32,playerNum int32,roomUuidList []string,shouldConnectClientList []uint64) {
	newTable := ts.NewTable()
	newTable.OnInit(tableUuid,tableType,playerNum,roomUuidList,shouldConnectClientList)
	ts.table[tableUuid] = newTable
}


func (ts *TableService) NewTable() *table.Table {
	table := tablePool.Get().(*table.Table)
	return table
}


// 来自GateService转发消息
// 来自GateService转发消息
type RpcOnRecvCallBack struct {
}

func (cb *RpcOnRecvCallBack) Unmarshal(data []byte) (interface{}, error) {
	var rawInput global.RawInputArgs

	err := rawInput.ParseMsg(data)
	if err != nil {
		log.Error("parse message is error:%s!", err.Error())
		return nil, err
	}
	clientIdList := rawInput.GetClientIdList()

	msgInfo := tableService.mapRegisterMsg[msg.MsgType(rawInput.GetMsgType())]
	if msgInfo == nil {
		err = fmt.Errorf("message type %d is not  register.", rawInput.GetMsgType())
		log.Warning("close client %+v,message type %d is not  register.", clientIdList, rawInput.GetMsgType())
		return nil, err
	}

	protoMsg := msgInfo.NewMsg()
	if protoMsg.msg !=nil {
		err = proto.Unmarshal(data[2+len(clientIdList)*8:], protoMsg.msg)
		if err != nil {
			err = fmt.Errorf("message type %d is not  register.", rawInput.GetMsgType())
			log.Warning("close client %+v,message type %d is not  register.", clientIdList, rawInput.GetMsgType())
			return nil, err
		}
	}

	rawInput.SetProtoMsg(protoMsg)

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


	msgType := msg.MsgType(args.GetMsgType())

	msgInfo, ok := tableService.mapRegisterMsg[msgType]
	if ok == false {
		log.Warning("close client %d,message type %d is not  register.", clientIdList[0], msgType)
		return
	}
	msgInfo.msgCallBack(tableService.cycleDo,clientIdList[0], args.GetProtoMsg().(*protoMsg).msg)
	msgInfo.ReleaseMsg(args.GetProtoMsg().(*protoMsg))
}

func (ts *TableService) RemoveTable(tableUuid string)  {
	delete(ts.table,tableUuid)
}

