package queueservice

import (
	"fmt"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/node"
	"github.com/duanhf2012/origin/service"
	sync2 "github.com/duanhf2012/origin/util/sync"
	"github.com/duanhf2012/origin/util/timer"
	"github.com/golang/protobuf/proto"
	"sunserver/common/entity"
	"sunserver/common/global"
	"sunserver/common/proto/msg"
	"sunserver/common/proto/rpc"
	"sunserver/gamecore/common"
	"sunserver/gamecore/queueservice/cycledo"
	"sunserver/gamecore/queueservice/def"
	"sunserver/gamecore/queueservice/msghandler"
	"sunserver/gamecore/roomservice/room"
	"sync"
	"time"
)

var queueService QueueService
var roomPool sync.Pool
var playerInfoPool sync.Pool

func init() {
	node.Setup(&queueService)
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
	queueService.mapRegisterMsg[msgType] = &regMsgInfo
}

type QueueService struct {
	service.Service

	mapRegisterMsg map[msg.MsgType]*RegMsgInfo //消息注册
	//mapClientOnline        map[int32]*ClientPlayer   // 用来校验用户是否登录
	//依据房间类型
	mapRoomOnQueue map[int32]*QueueTypeList // 在线排队 roomType->[ {a-b:[实际的人]}  ]

	mapRoom map[string]*room.Room //存放房间的 判断房间是否存在

	maxPlayerNum uint64 //每个队列排队最大人数

	//groupPlayerNum uint64 // 每组匹配人数
	gateProxy *common.GateProxyModule //网关代理

	cycleDo *cycledo.QueueInterface

	match *MatchModule // 匹配模块

}

type RankRange struct {
	startRank uint64
	endRank   uint64
}

/*type ClientPlayer struct {
	refreshTime     time.Time
	mapClientPlayer map[uint64]*user.QueueUser //clientID 从PlayService同步过来
}
*/
type QueueTypeList struct {
	/*{
		"qUEUEType": 1,
		"playerNum": 10,
		"minRank": 0,
		"maxRank": 100000,
		"rankInterval": 1000
	},*/
	queueType    int32
	playerNum    int32
	minRank      uint64
	maxRank      uint64
	rankInterval uint64
	//依据分数房间队列
	queueList map[*RankRange]*QueueList
}

type QueueList struct {
	waitPlayerNum uint64
	roomList      *def.MapList //clientID 从PlayService同步过来
}

const UINT_MAX = ^uint64(0)

func (qs *QueueService) OnInit() error {
	//1.初始化变量与模块
	qs.mapRoomOnQueue = make(map[int32]*QueueTypeList, 4096)
	qs.mapRoom = make(map[string]*room.Room, 4096)
	qs.mapRegisterMsg = make(map[msg.MsgType]*RegMsgInfo, 512)
	//qs.mapClientOnline = make(map[int32]*ClientPlayer, 20)
	roomPool = sync.Pool{New: func() interface{} {
		return &room.Room{}
	}}

	playerInfoPool = sync.Pool{New: func() interface{} {
		return &entity.PlayerInfo{}
	}}

	qs.gateProxy = common.NewGateProxyModule()
	qs.AddModule(qs.gateProxy)

	qs.match = NewMatchModule()
	qs.AddModule(qs.match)

	qs.OnLoadCfg()

	//2.设置注册函数回调
	msghandler.OnRegisterMessage(RegisterMessage)

	//10 s 扫描一次匹配队列
	qs.NewTicker(30*time.Second, qs.matchProcess)

	//10 s 打印一次队列
	//qs.NewTicker(10*time.Second,qs.fmtQueue)

	qs.OpenProfiler()
	qs.GetProfiler().SetOverTime(time.Millisecond * 50)
	qs.GetProfiler().SetMaxOverTime(time.Second * 10)

	// 注册原始套接字回调
	qs.RegRawRpc(global.RawRpcOnRecv, &RpcOnRecvCallBack{})

	//绑定
	queueInterface := cycledo.New(qs)
	qs.cycleDo = queueInterface
	return nil
}

//扫描不同的rank组 进行组装匹配
func (qs *QueueService) matchProcess(timer *timer.Ticker) {
	log.Release("执行匹配开始时间--%s", time.Now().String())
	// 按照rank 组 分类匹配
	for queueType, queueTypeInfo := range qs.mapRoomOnQueue {
		log.Release("目前执行匹配类型%d", queueType)
		queueList := queueTypeInfo.queueList
		playerNum := queueTypeInfo.playerNum
		for rankRange, playerList := range queueList {
			log.Release("目前执行匹配rank段为[%d-%d]", rankRange.startRank, rankRange.endRank)
			num := playerList.waitPlayerNum
			roomList := playerList.roomList
			if num < uint64(playerNum) {
				//说明不够 遍历下一个组合
				continue
			}
			//匹配模块的匹配方法
			qs.match.Match(roomList, queueTypeInfo.playerNum)
		}
	}
}

func (qs *QueueService) OnLoadCfg() {
	cfg := qs.GetServiceCfg()
	configMap, ok := cfg.(map[string]interface{})
	/*""QueueService": {
	      "MaxQueueNum": 5000,
	      "Queue": [
	        {
	          "QueueType": 1,
	          "PlayerNum": 10,
	          "MinRank": 0,
	          "MaxRank": 100000,
	          "RankInterval": 1000
	        },
	}*/
	//队列人数
	maxQueueNum, ok := configMap["MaxQueueNum"]
	if ok == false {
		return
	}
	qs.maxPlayerNum = uint64(maxQueueNum.(float64))
	queue, ok := configMap["Queue"]
	if ok == false {
		return
	}

	queueList, ok := queue.([]interface{})
	if ok == false {
		//error...
		return
	}
	// "Queue": [
	//	        {
	//	          "QueueType": 1,
	//	          "PlayerNum": 10,
	//	          "MinRank": 0,
	//	          "MaxRank": 100000,
	//	          "RankInterval": 1000
	//	        },
	for _, v := range queueList {
		mapItem := v.(map[string]interface{})
		var iQueueType interface{}
		var iPlayerNum interface{}
		var iMinRank interface{}
		var iMaxRank interface{}
		var iRankInterval interface{}

		if iQueueType, ok = mapItem["RoomType"]; ok == false {
			//error ...
			continue
		}
		if iPlayerNum, ok = mapItem["PlayerNum"]; ok == false {
			//error ...
			continue
		}
		if iMinRank, ok = mapItem["MinRank"]; ok == false {
			//error ...
			continue
		}
		if iMaxRank, ok = mapItem["MaxRank"]; ok == false {
			//error ...
			continue
		}
		if iRankInterval, ok = mapItem["RankInterval"]; ok == false {
			//error ...
			continue
		}
		//依据这几个数值进行区间换算
		var rank uint64
		rankMap := make(map[*RankRange]*QueueList)
		//组装列表
		roomType := int32(iQueueType.(float64))
		playNum := int32(iPlayerNum.(float64))
		MinRank := uint64(iMinRank.(float64))
		MaxRank := uint64(iMaxRank.(float64))
		RankInterval := uint64(iRankInterval.(float64))
		for {
			if MinRank >= MaxRank {
				//放入剩余部分
				if MaxRank%RankInterval != 0 {
					//没有余数 有余数进行处理
					rankMap[&RankRange{MinRank - RankInterval, MaxRank}] = &QueueList{0, def.NewMapList()}
				}
				//对数据放入 maxRank-MAX
				rankMap[&RankRange{MaxRank, UINT_MAX}] = &QueueList{0, def.NewMapList()}
				break
			}
			rank = MinRank + RankInterval
			rankMap[&RankRange{MinRank, rank}] = &QueueList{0, def.NewMapList()}
			MinRank = rank
		}
		//组装对象
		qs.mapRoomOnQueue[roomType] = &QueueTypeList{roomType, playNum, uint64(iMinRank.(float64)), uint64(iMaxRank.(float64)), uint64(iRankInterval.(float64)), rankMap}
	}

}

/*func (qs *QueueService) NewQueueUser(clientId uint64) *user.QueueUser {
	user := queueUserPool.Get().(*user.QueueUser)
	user.SetClientId(clientId)
	user.ISender = qs.gateProxy
	return user
}
*/

/*func (qs *QueueService) NewMatchUser(clientId uint64) *user.QueueUser {
	user := queueUserPool.Get().(*user.QueueUser)
	user.SetClientId(clientId)
	user.ISender = qs.gateProxy
	return user
}*/

/*
func (qs *QueueService) CheckClientIdOnLine(clientId uint64) bool {

	log.Release("检查客户端是否在线---")
	var flag bool
	for nodeId, clientPlayer := range qs.mapClientOnline {
		log.Release("nodeId:%d,clientPlayer人数:%d", nodeId, clientPlayer)
		_, ok := clientPlayer.mapClientPlayer[clientId]
		if ok {
			flag = true
		}
	}
	return flag
}*/
func (qs *QueueService) NewRoom() *room.Room {
	room := roomPool.Get().(*room.Room)
	return room
}

func (qs *QueueService) NewPlayerInfo() *entity.PlayerInfo {
	playerInfo := playerInfoPool.Get().(*entity.PlayerInfo)
	return playerInfo
}

func (qs *QueueService) fmtQueue(timer *timer.Ticker) {
	for k, v := range qs.mapRoomOnQueue {
		log.Release("打印排队队列 k--%d", k)
		for rang, list := range v.queueList {
			roomList := list.roomList
			log.Release("rank段:%d-%d,房间数:%d", rang.startRank, rang.endRank, roomList.Size())
		}
	}
}

func (qs *QueueService) PackPlayerInfo(pbData *rpc.PlayerInfo) *entity.PlayerInfo {
	playerInfo := qs.NewPlayerInfo()
	playerInfo.SetUserId(pbData.GetUserId())
	playerInfo.SetSex(pbData.GetSex())
	playerInfo.SetNickName(pbData.GetNickName())
	playerInfo.SetAvatar(pbData.GetAvatar())
	playerInfo.SetClientId(pbData.ClientId)
	playerInfo.SetRank(pbData.GetRank())
	playerInfo.SetOwner(pbData.GetIsOwner())
	playerInfo.SetSeatNum(pbData.GetSeatNum())

	return playerInfo
}

func (qs *QueueService) PackRoom(res *rpc.GetRoomRes) *room.Room {
	room := qs.NewRoom()
	pbRoom := res.GetRoom()
	pbOwner := pbRoom.GetOwner()
	owner := qs.PackPlayerInfo(pbOwner)
	//初始化其他用户
	otherClients := pbRoom.GetOtherClients()
	other := make([]*entity.PlayerInfo, len(otherClients))
	for _, client := range otherClients {
		other = append(other, qs.PackPlayerInfo(client))
	}
	room.PackFromPb(qs.gateProxy, pbRoom.GetUuid(), pbRoom.GetRoomName(), pbRoom.GetRoomClientNum(), owner, other, pbRoom.GetRoomType(), pbRoom.GetAvgRank())
	return room
}

func (qs *QueueService) AddQueue(room *room.Room) bool {
	log.Release("加入队列---%d", room.GetUUid())
	rank := room.GetAvgRank()
	roomType := room.GetRoomType()
	var flag bool
	for queueType, info := range qs.mapRoomOnQueue {
		if int32(queueType) == roomType {
			list := info.queueList
			for rankRange, playerList := range list {
				//取左不取右
				if rankRange.startRank <= rank && rankRange.endRank > rank {
					//在区间里面
					playerList.waitPlayerNum += uint64(room.GetRoomClientNum())
					playerList.roomList.Push(room.GetUUid(), room)
					flag = true
				}
			}
		}
	}
	return flag
}

func (qs *QueueService) AddRoom(room *room.Room) {
	qs.mapRoom[room.GetUUid()] = room
}

func (qs *QueueService) GetRoom(roomUuid string) *room.Room {
	return qs.mapRoom[roomUuid]
}

func (qs *QueueService) RemoveRoom(roomUuid string) {
	delete(qs.mapRoom, roomUuid)
}

func (qs *QueueService) QuitQueue(room *room.Room) {
	log.Release("移除队列---%d", room.GetUUid())
	rank := room.GetAvgRank()
	roomType := room.GetRoomType()
	for queueType, info := range qs.mapRoomOnQueue {
		if queueType == roomType {
			list := info.queueList
			for rankRange, playerList := range list {
				if rankRange.startRank <= rank && rankRange.endRank >= rank {
					//在区间里面
					playerList.waitPlayerNum -= uint64(room.GetRoomClientNum())
					playerList.roomList.Remove(room.GetUUid())
				}
			}
		}
	}
}

/*// PlayerService服同步负载情况
func (qs *QueueService) RPC_UpdateBalance(clist *rpc.UpdateClientList) error {
	log.Release("进入到QueueService-RPC_UpdateBalance")
	cNodeId := clist.NodeId
	for nodeId, _ := range qs.mapClientOnline {
		if nodeId == cNodeId {
			delete(qs.mapClientOnline, nodeId)
		}
	}

	//重新同步
	clientMap := make(map[uint64]*user.QueueUser)
	for cId := range clist.CList {
		clientMap[uint64(cId)] = qs.NewQueueUser(uint64(cId))
	}
	if len(clientMap) > 0 {
		qs.mapClientOnline[cNodeId] = &ClientPlayer{time.Now(), clientMap}
	}
	log.Release("同步之后---:%d", len(qs.mapClientOnline))
	return nil
}*/

/*// PlayerService服同步负载情况
func (qs *QueueService) RPC_UpdateBalance_One(client *rpc.UpdateClientOne) error {
	log.Release("进入到QueueService-RPC_UpdateBalance_One")
	cNodeId := client.NodeId
	clientPlayer := qs.mapClientOnline[cNodeId]
	if clientPlayer == nil {
		clientPlayer = &ClientPlayer{
			time.Now(), make(map[uint64]*user.QueueUser),
		}
	}
	clientPlayer.refreshTime = time.Now()
	clientPlayer.mapClientPlayer[client.ClientId] = qs.NewQueueUser(client.ClientId)
	log.Release("大厅在线人数---%d", len(clientPlayer.mapClientPlayer))
	return nil
}
*/

// 收到来自GateService通知下线
/*func (qs *QueueService) RPCOnClose(byteBuffer []byte) {
	var rawInput global.RawInputArgs
	_,err := rawInput.ParseUint64(byteBuffer)
	if err!=nil {
		log.Error("error data :%s!",err.Error())
		return
	}

	return
}*/

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

	msgInfo := queueService.mapRegisterMsg[msg.MsgType(rawInput.GetMsgType())]
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

	msgType := msg.MsgType(args.GetMsgType())

	msgInfo, ok := queueService.mapRegisterMsg[msgType]
	if ok == false {
		log.Warning("close client %d,message type %d is not  register.", clientIdList[0], msgType)
		return
	}
	msgInfo.msgCallBack(queueService.cycleDo, clientIdList[0], args.GetProtoMsg().(*protoMsg).msg)
	msgInfo.ReleaseMsg(args.GetProtoMsg().(*protoMsg))
}

func (qs *QueueService) SendMsg(clientId uint64, msgType msg.MsgType, msg proto.Message) {
	qs.gateProxy.SendToClient(clientId, msgType, msg)
}
