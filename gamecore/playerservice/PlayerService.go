package playerservice

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/cluster"
	"github.com/duanhf2012/origin/event"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/node"
	originrpc "github.com/duanhf2012/origin/rpc"
	"github.com/duanhf2012/origin/service"
	sync2 "github.com/duanhf2012/origin/util/sync"
	"github.com/duanhf2012/origin/util/timer"
	"github.com/golang/protobuf/proto"
	"strconv"
	"sunserver/common/configdef"
	"sunserver/common/const"
	"sunserver/common/db"
	"sunserver/common/global"
	"sunserver/common/module"
	"sunserver/common/proto/msg"
	"sunserver/common/proto/rpc"
	"sunserver/common/util"
	"sunserver/gamecore/common"
	"sunserver/gamecore/playerservice/msghandler"
	"sunserver/gamecore/playerservice/player"
	"sync"
	"time"
)

var playerService PlayerService

func init() {
	node.Setup(&playerService)
}

type protoMsg struct {
	ref bool
	msg proto.Message
}

func (m *protoMsg) Reset() {
	if m.msg != nil {
		m.msg.Reset()
	}
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
	msgPool     *sync2.PoolEx
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
	regMsgInfo.msgPool = sync2.NewPoolEx(make(chan sync2.IPoolData, 1000), func() sync2.IPoolData {
		protoMsg := protoMsg{}
		protoMsg.msg = proto.Clone(regMsgInfo.protoMsg.msg)
		return &protoMsg
	})
	regMsgInfo.msgCallBack = cb
	playerService.mapRegisterMsg[msgType] = &regMsgInfo
}

var playerPool sync.Pool

//var protoMsgPool sync.Pool

var GateService string = "GateService"

type PlayerService struct {
	service.Service

	mapRegisterMsg  map[msg.MsgType]*RegMsgInfo //消息注册
	mapPlayer       map[uint64]*player.Player   //map[userId]*Player
	mapClientPlayer map[uint64]*player.Player   //map[clientId]*Player

	gateProxy *common.GateProxyModule //网关代理
	cfgModule *module.LoadCfgModule   //加载配置模块

	balance         rpc.PlayerServiceBalance //负载同步变量
	mapCenterNodeId map[int]interface{}      //存储所有的CenterService的NodeId
}

type MsgRouterInfo struct {
	Rpc             string
	ServiceName     string
	LoadBalanceType string
}

func (ps *PlayerService) NewPlayer(id uint64) *player.Player {
	player := playerPool.Get().(*player.Player)
	player.Clear()
	player.Id = id
	return player
}

func (ps *PlayerService) ReleasePlayer(player *player.Player) {
	log.Release("ReleasePlayer")
	var playerStatus rpc.UpdatePlayerStatus
	playerStatus.UserId = player.GetUserId()
	playerStatus.Status = rpc.LoginStatus_LoginOut
	playerStatus.NodeId = int32(node.GetNodeId())
	masterNodeId := util.GetMasterCenterNodeId()
	if masterNodeId != 0 {
		log.Release("释放用户时候触发CenterService.RPC_UpdateStatus")
		err := ps.GoNode(masterNodeId, "CenterService.RPC_UpdateStatus", &playerStatus)
		if err != nil {
			log.Error("Go CenterService.RPC_UpdateStatus fail")
		}
	} else {
		log.Error("cannot find masterNodeId")
	}

	if player.GetClientId() > 0 {
		//断开连接
		log.Warning("close client  %d,release player", player.GetClientId())
		ps.CloseClient(player.GetClientId())
	}

	//断开关系
	delete(ps.mapClientPlayer, player.GetClientId())
	delete(ps.mapPlayer, player.GetUserId())

	//回收对象
	player.Clear()
	playerPool.Put(player)
}

func (ps *PlayerService) OnInit() error {
	//1.初始化变量与模块
	ps.mapPlayer = make(map[uint64]*player.Player, 4096)
	ps.mapClientPlayer = make(map[uint64]*player.Player, 4096)
	ps.mapRegisterMsg = make(map[msg.MsgType]*RegMsgInfo, 512)
	ps.mapCenterNodeId = make(map[int]interface{}, 4)
	ps.gateProxy = common.NewGateProxyModule()
	ps.AddModule(ps.gateProxy)

	playerPool = sync.Pool{New: func() interface{} {
		return &player.Player{}
	}}

	/*protoMsgPool = sync.Pool{New: func() interface{} {
		return &protoMsg{false,nil}
	}}*/

	//2.设置注册函数回调
	msghandler.OnRegisterMessage(RegisterMessage)
	cluster.GetCluster().RegisterRpcListener(ps)

	//3.打开定时器，定时向CenterService同步负载
	ps.balance.NodeId = int32(node.GetNodeId())
	//10s  同步一次负载
	ps.NewTicker(10*time.Second, ps.timerUpdateBalance)
	// 和队列 5s  同步一次
	//ps.NewTicker(5*time.Second,ps.timerUpdateBalanceQueue)

	//4.获得所有的CenterService的NodeId
	var rpcClientList [4]*originrpc.Client
	err, num := cluster.GetCluster().GetNodeIdByService("CenterService", rpcClientList[:], true)
	if err != nil {
		return err
	}
	if num == 0 {
		return fmt.Errorf("cannot find CenterService nodeId")
	}
	for i := 0; i < num; i++ {
		ps.mapCenterNodeId[rpcClientList[i].GetId()] = nil
	}

	//5.打开性能监控
	ps.OpenProfiler()
	ps.GetProfiler().SetOverTime(time.Millisecond * 50)
	ps.GetProfiler().SetMaxOverTime(time.Second * 10)

	//6.事件注册
	ps.OnRegisterEvent()

	//7. 逻辑配置加载模块
	err = ps.OnLoadCfg()
	if err != nil {
		return err
	}

	//8. 注册原始套接字回调
	ps.RegRawRpc(global.RawRpcOnRecv, &RpcOnRecvCallBack{})
	ps.RegRawRpc(global.RawRpcOnClose, &RpcOnCloseCallBack{})

	return nil
}

func (ps *PlayerService) OnLoadCfg() error {
	//1.添加配置加载模块
	ps.cfgModule = module.NewLoadCfgModule()
	ps.AddModule(ps.cfgModule)

	//2.具体加载的配置表
	err := ps.cfgModule.LoadCfg(configdef.FileTemplate)
	if err != nil {
		return err
	}

	/*fileObj := ps.cfgModule.GetConfig(configdef.FileTemplate).(*configdef.TemplateCfg)
	cfgItem := fileObj.GetItemByFirstIndex(1).(*configdef.TemplateCSVCfg)
	log.Release(cfgItem)

	cfgItem2 := fileObj.GetItemByChooseIndex(2, configdef.TemplateStructKey{	A: 1, B: "2" }).(*configdef.TemplateCSVCfg)
	log.Release(cfgItem2)*/
	return nil
}

func (ps *PlayerService) OnRegisterEvent() {
	//注册监听加载配置完成事件
	ps.RegEventReceiverFunc(global.EventConfigComplete, ps.GetEventHandler(), ps.ConfigComplete)
}

// 主动关闭连接
func (ps *PlayerService) CloseClient(clientId uint64) {
	log.Release("CloseClient")
	if clientId == 0 {
		log.Error("clientId is error.")
		return
	}

	//查找Player对象
	p, ok := ps.mapClientPlayer[clientId]
	if ok == false {
		log.Error("clientId %d not found", clientId)
		return
	}

	//设置为离线状态
	p.SetOnline(false)
	nodeId := p.GetFromGateId()

	//通知网关关闭连接
	var inputArgs global.RawInputArgs
	inputArgs.SetUint64(clientId)
	ps.RawGoNode(originrpc.RpcProcessorGoGoPB, nodeId, global.RawRpcCloseClient, global.GateService, &inputArgs)
	//通知房间如果有就下线
	delete(ps.mapClientPlayer, clientId)

}

// 收到来自GateService通知下线
func (ps *PlayerService) RPCOnClose(byteBuffer []byte) {
	var rawInput global.RawInputArgs
	clientId, err := rawInput.ParseUint64(byteBuffer)
	if err != nil {
		log.Error("error data :%s!", err.Error())
		return
	}

	v, ok := ps.mapClientPlayer[clientId]
	if ok == false {
		log.Warning("Cannot find clientId %d.", clientId)
		return
	}
	v.SetOnline(false)
	//发送给数据库
	var mysqlData db.MysqlControllerReq

	sql := "update `user` set last_login_time = ?,is_login=? where id = ?"
	args := []string{strconv.FormatInt(timer.Now().Unix(), 10), "0", strconv.FormatUint(v.Id, 10)}
	db.MakeMysql(constpackage.UserTableName, uint64(util.HashString2Number(v.PlatId)), sql, args, db.OptType_Update, &mysqlData)
	ps.SendMsgToMysql(&mysqlData)

	//往队列服发送通知下线
	//ps.UpdateBalanceQueue(clientId,2)
	return
}

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
	v, ok := playerService.mapClientPlayer[clientId]
	if ok == false {
		log.Warning("Cannot find clientId %d.", clientId)
		return
	}

	v.SetOnline(false)
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
	clientIdList := rawInput.GetClientIdList()

	msgInfo := playerService.mapRegisterMsg[msg.MsgType(rawInput.GetMsgType())]
	if msgInfo == nil {
		err = fmt.Errorf("message type %d is not  register.", rawInput.GetMsgType())
		log.Warning("close client %+v,message type %d is not  register.", clientIdList, rawInput.GetMsgType())
		return nil, err
	}

	protoMsg := msgInfo.NewMsg()
	if protoMsg.msg != nil {
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

	clientId := clientIdList[0]
	player, ok := playerService.mapClientPlayer[clientId]
	if ok == false {
		log.Warning("close client %d,mapClientPlayer not exists clientId", clientId)
		log.Release("RpcOnRecvCallBack CB 1")
		playerService.CloseClient(clientId)
		return
	}

	msgType := msg.MsgType(args.GetMsgType())
	if msgType != msg.MsgType_Ping && player.IsLoadFinish() == false {
		log.Warning("close client %d, Player data has not been loaded yet", clientId)
		//ps.CloseClient(clientId)
		return
	}

	msgInfo, ok := playerService.mapRegisterMsg[msgType]
	if ok == false {
		log.Release("RpcOnRecvCallBack CB 2")
		playerService.CloseClient(clientId)
		log.Warning("close client %d,message type %d is not  register.", player.GetClientId(), msgType)
		return
	}
	msgInfo.msgCallBack(player, args.GetProtoMsg().(*protoMsg).msg)
	msgInfo.ReleaseMsg(args.GetProtoMsg().(*protoMsg))
}

// 重载配置
func (ps *PlayerService) RPC_ReLoadCfg(argInfo *rpc.PlaceHolders, retInfo *rpc.ReloadCfgResult) error {
	ps.cfgModule.AsyncReLoadCfg()
	retInfo.Status = 0
	return nil
}

func (ps *PlayerService) RPC_Login(req *rpc.LoginToPlayerServiceReq, res *rpc.LoginToPlayerServiceRet) error {
	//1.查找Player对象
	log.Release("进入到PlayerService-RPC_Login")
	res.Ret = 0
	p, ok := ps.mapPlayer[req.UserId]
	masterNodeId := util.GetMasterCenterNodeId()
	if masterNodeId == 0 {
		res.Ret = 1
		log.Error("getBestMasterNodeId is fail")
		return nil
	}

	//2.先同步Player登陆状态
	var playerStatus rpc.UpdatePlayerStatus
	playerStatus.UserId = req.UserId
	playerStatus.Status = rpc.LoginStatus_Logined
	playerStatus.NodeId = int32(node.GetNodeId())
	//往中心服发送 同步用户
	log.Release("发往CenterService.RPC_UpdateStatus")
	err := ps.GoNode(masterNodeId, "CenterService.RPC_UpdateStatus", &playerStatus)
	if err != nil {
		res.Ret = 2
		log.Error("go CenterService.RPC_UpdateStatus fail %s", err.Error())
		return nil
	}
	//往队列服务 同步客户端连接

	//3.创建或初始化玩家
	if ok == false {
		p = ps.NewPlayer(req.UserId)
		ps.ResetConn(req.ClientId, req.UserId, int(req.NodeId), p)
		p.OnInit(ps.GetRpcHandler(), ps, ps.gateProxy, ps)
		p.StartLogin(req.ClientId, req.UserId, int(req.NodeId))
	} else {
		//关闭老连接
		log.Warning("close client %d,player reLogin.", p.GetClientId())
		ps.CloseClient(p.GetClientId())
		//重新关联新连接
		ps.ResetConn(req.ClientId, req.UserId, int(req.NodeId), p)
		//重登陆
		p.ReLogin(req.ClientId, req.UserId, int(req.NodeId))

	}
	log.Release("出来了++++++++++")
	return nil
}

func (ps *PlayerService) ResetConn(cliId uint64, userId uint64, fromGateId int, p *player.Player) {
	log.Release("ResetConn %d", cliId)
	p.SetOnline(true)
	ps.mapPlayer[userId] = p
	ps.mapClientPlayer[cliId] = p
	/*var mysqlData db.MysqlControllerReq
	mysqlData.TableName = constpackage.UserTableName
	sql := "update `user` set last_login_time = ?,is_login=? where id = ?"
	args := []string{strconv.FormatInt(timer.Now().Unix(), 10), "1", strconv.FormatUint(userId, 10)}
	db.MakeMysql(constpackage.UserTableName, uint64(util.HashString2Number(p.PlatId)), sql, args, db.OptType_Update, &mysqlData)
	ps.SendMsgToMysql(&mysqlData)*/
}

func (ps *PlayerService) TimerAfter(userId uint64, d time.Duration, cb func(ticker *timer.Timer)) *timer.Timer {
	return ps.AfterFunc(d, func(ticker *timer.Timer) {
		_, ok := ps.mapPlayer[userId]
		if ok == true {
			cb(ticker)
		}
	})
}

func (ps *PlayerService) TimerTicker(userId uint64, d time.Duration, cb func(ticker *timer.Ticker)) *timer.Ticker {
	ticker := ps.NewTicker(d, func(ticker *timer.Ticker) {
		_, ok := ps.mapPlayer[userId]
		if ok == true {
			cb(ticker)
		} else {
			ticker.Cancel()
		}
	})

	return ticker
}

// 向中心服同步PlayerService负载
// 向队列服务同步在线用户
func (ps *PlayerService) timerUpdateBalance(timer *timer.Ticker) {
	nodeId := util.GetMasterCenterNodeId()
	queueNodeId := util.GetNodeIdByService(global.CenterService)
	if nodeId == 0 {
		fmt.Errorf("cannot find best centerservice nodeid")
		return
	}

	if queueNodeId == 0 {
		fmt.Errorf("cannot find best centerservice nodeid")
		return
	}

	ps.balance.Weigh = int32(len(ps.mapPlayer))
	err := ps.GoNode(nodeId, "CenterService.RPC_UpdateBalance", &ps.balance)
	if err != nil {
		log.Error("RPC_UpdateBalance fail %s", err.Error())
	}
}

/*func (ps *PlayerService) timerUpdateBalanceQueue(timer *timer.Ticker){
	nodeId := util.GetNodeIdByService(global.QueueService)
	if nodeId == 0 {
		fmt.Errorf("cannot find best queueservice nodeid")
		return
	}

	var clientList rpc.UpdateClientList
	clientList.NodeId = int32(node.GetNodeId())
	clientList.CList = make([]uint64, 0, len(ps.mapClientPlayer))
	for clientId,player := range ps.mapClientPlayer {
		if player.GetOnline() {
			clientList.CList = append(clientList.CList ,clientId)
		}
	}

	err := ps.GoNode(nodeId,"QueueService.RPC_UpdateBalance",&clientList)
	if err != nil {
		log.Error("RPC_UpdateBalance fail %s",err.Error())
	}
}

func (ps *PlayerService) updateBalanceQueue(){
	nodeId := util.GetNodeIdByService(global.QueueService)
	if nodeId == 0 {
		fmt.Errorf("cannot find best queueservice nodeid")
		return
	}

	var clientList rpc.UpdateClientList
	clientList.NodeId = int32(node.GetNodeId())
	clientList.CList = make([]uint64, 0, len(ps.mapClientPlayer))
	for clientId,_ := range ps.mapClientPlayer {
		clientList.CList = append(clientList.CList ,clientId)
	}

	err := ps.GoNode(nodeId,"QueueService.RPC_UpdateBalance",&clientList)
	if err != nil {
		log.Error("RPC_UpdateBalance fail %s",err.Error())
	}
}

func (ps *PlayerService) UpdateBalanceQueue(clientId uint64,flag int32){
	nodeId := util.GetNodeIdByService(global.QueueService)
	if nodeId == 0 {
		fmt.Errorf("cannot find best queueservice nodeid")
		return
	}

	var clientOne rpc.UpdateClientOne
	clientOne.NodeId = int32(node.GetNodeId())
	clientOne.ClientId = clientId
	// 登入
	clientOne.Flag = flag

	err := ps.GoNode(nodeId,"QueueService.RPC_UpdateBalance_One",&clientOne)
	if err != nil {
		log.Error("RPC_UpdateBalance fail %s",err.Error())
	}
}
//通知房间关闭
func (ps *PlayerService) NoticeRemoveRoom(clientId uint64){
	nodeId := util.GetNodeIdByService(global.QueueService)
	if nodeId == 0 {
		fmt.Errorf("cannot find best queueservice nodeid")
		return
	}

	var removeRoom rpc.RemoveOneRoom
	removeRoom.RoomUuid = clientId
	// 登入

	err := ps.GoNode(nodeId,"RoomService.RPC_RemoveRoom",&removeRoom)
	if err != nil {
		log.Error("RPC_UpdateBalance fail %s",err.Error())
	}
}*/

func (ps *PlayerService) IsCenterNode(nodeId int) bool {
	if _, ok := ps.mapCenterNodeId[nodeId]; ok == true {
		return true
	}
	return false
}

func (ps *PlayerService) OnRpcConnected(nodeId int) {
	//如果是重新连接上中心服，重新同步玩家列表
	if ps.IsCenterNode(nodeId) == true {
		nodeId := util.GetMasterCenterNodeId()
		if nodeId == 0 {
			fmt.Errorf("cannot find best centerservice nodeid")
			return
		}

		var playerList rpc.UpdatePlayerList
		playerList.NodeId = int32(node.GetNodeId())
		playerList.UList = make([]uint64, 0, len(ps.mapPlayer))
		for uId, _ := range ps.mapPlayer {
			playerList.UList = append(playerList.UList, uId)
		}

		ps.GoNode(nodeId, "CenterService.RPC_UpdateUserList", &playerList)
		//向队列服同步信息
		//ps.updateBalanceQueue()
	}
}

func (ps *PlayerService) RPC_CheckOnline(req *rpc.CheckOnLineReq, res *rpc.CheckOnLineRes) error {
	_, ok := ps.mapClientPlayer[req.ClientId]
	res.Flag = ok
	return nil
}

func (ps *PlayerService) OnRpcDisconnect(nodeId int) {
}

func (ps *PlayerService) ConfigComplete(ev event.IEvent) {
	loadFileList := ev.(*event.Event).Data.([]module.ReadLogicCfgData)
	for _, fileItem := range loadFileList {
		ps.cfgModule.SetLogicConfig(fileItem.FileName, fileItem.Record)
	}
}

func (ps *PlayerService) SendMsgToMysql(req *db.MysqlControllerReq) {
	mysqlServiceNodeId := util.GetNodeIdByService("MysqlService")
	ps.GoNode(mysqlServiceNodeId, "MysqlService.RPC_MysqlDBRequest", req)
}
