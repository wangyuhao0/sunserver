package roomservice

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
	"sunserver/gamecore/roomservice/cycledo"
	"sunserver/gamecore/roomservice/msghandler"
	"sync"
	"time"
)

var roomService RoomService
var roomPool sync.Pool
var playerInfoPool sync.Pool

func init() {
	node.Setup(&roomService)
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
	roomService.mapRegisterMsg[msgType] = &regMsgInfo
}

type RoomService struct {
	service.Service

	mapRegisterMsg map[msg.MsgType]*RegMsgInfo //消息注册
	// k 为 roomUuid

	//第一个k为房间类型 第二个key 房间号 做房间类型区分
	room map[int32]map[string]*common.Room //
	//room map[string]*common.Room //在线房间数 还有房主信息

	cycleDo *cycledo.RoomInterface

	gateProxy *common.GateProxyModule //网关代理
}

func (rs *RoomService) OnInit() error {
	//1.初始化变量与模块
	rs.room = make(map[int32]map[string]*common.Room)

	rs.mapRegisterMsg = make(map[msg.MsgType]*RegMsgInfo, 512)
	roomPool = sync.Pool{New: func() interface{} {
		return &common.Room{}
	}}
	playerInfoPool = sync.Pool{New: func() interface{} {
		return &entity.PlayerInfo{}
	}}
	rs.gateProxy = common.NewGateProxyModule()
	rs.AddModule(rs.gateProxy)
	//2.设置注册函数回调
	msghandler.OnRegisterMessage(RegisterMessage)
	rs.OnLoadCfg()
	rs.OpenProfiler()
	rs.GetProfiler().SetOverTime(time.Millisecond * 50)
	rs.GetProfiler().SetMaxOverTime(time.Second * 10)
	//打印房间人数
	//rs.NewTicker(10*time.Second,rs.FmtRoom)

	// 注册原始套接字回调
	rs.RegRawRpc(global.RawRpcOnRecv, &RpcOnRecvCallBack{})

	roomInterface := cycledo.New(rs)
	rs.cycleDo = roomInterface
	return nil
}

func (rs *RoomService) OnLoadCfg() {

	//写死3种
	/*	rs.room[1] = make(map[string]*common.Room,3000)
		rs.room[2] = make(map[string]*common.Room,3000)
		rs.room[3] = make(map[string]*common.Room,3000)*/

	cfg := rs.GetServiceCfg()
	configMap, ok := cfg.(map[string]interface{})

	room, ok := configMap["Room"]
	if ok == false {
		return
	}

	roomList, ok := room.([]interface{})
	if ok == false {
		//error...
		return
	}

	for _, v := range roomList {
		mapItem := v.(map[string]interface{})
		var iRoomType interface{}
		//var iUserNum interface{}
		var iRoomNum interface{}
		if iRoomType, ok = mapItem["RoomType"]; ok == false {
			//error ...
			continue
		}
		/*if iUserNum, ok = mapItem["UserNum"]; ok == false {
			//error ...
			continue
		}*/
		if iRoomNum, ok = mapItem["RoomNum"]; ok == false {
			//error ...
			continue
		}
		//依据这几个数值进行区间换算
		//组装列表
		roomType := int32(iRoomType.(float64))
		//userNum := int32(iUserNum.(float64))
		roomNum := int32(iRoomNum.(float64))
		rs.room[roomType] = make(map[string]*common.Room, roomNum)
	}

}

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

	msgInfo := roomService.mapRegisterMsg[msg.MsgType(rawInput.GetMsgType())]
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

	msgInfo, ok := roomService.mapRegisterMsg[msgType]
	if ok == false {
		log.Warning("close client %d,message type %d is not  register.", clientIdList[0], msgType)
		return
	}
	msgInfo.msgCallBack(roomService.cycleDo, clientIdList[0], args.GetProtoMsg().(*protoMsg).msg)
	msgInfo.ReleaseMsg(args.GetProtoMsg().(*protoMsg))
}

/*func (rs *RoomService) CreateRoom(clientId uint64,roomType int32,info *entity.PlayerInfo) {
	//先向 play发送验证是否登录
	log.Release("roomService-createRoom")
	onlineFlag := rs.CheckOnline(clientId)
	//登录平台了 然后创建房间放入
	if onlineFlag {
		rs.Create(clientId,roomType,info)
	}
}
*/
/*func (rs *RoomService) AddRoom(clientId uint64, roomId uint64, info *entity.PlayerInfo) {
	//先验证是否登录
	log.Release("roomService-AddRoom")
	onlineFlag := rs.CheckOnline(clientId)
	if onlineFlag {
		//进行下一步
		rs.Add(clientId, roomId, info)
	}
}*/

func (rs *RoomService) FmtRoom(timer *timer.Ticker) {
	log.Release("打印房间---------")
	/*for k, v := range rs.room {
		log.Release("room 房间号--%s,用户数--%d,owner--%d", k, v.GetRoomClientNum(), v.GetOwner().GetUserId())
	}*/
}

func (rs *RoomService) CheckOnline(clientId uint64) bool {
	log.Release("roomService-CheckOnline")
	var req rpc.CheckOnLineReq
	req.ClientId = clientId
	err := rs.GetService().GetRpcHandler().AsyncCall("PlayerService.RPC_CheckOnline", &req, func(res *rpc.CheckOnLineRes, err error) {
		flag := res.Flag
		if err != nil {
			log.Error("callPlayerService.CheckOnline fail %s,clientId:%s!", err.Error(), clientId)
			rs.SendMsg(clientId, msg.MsgType_CreateRoomRes, &msg.MsgAddQueueRes{Ret: msg.ErrCode_InterNalError})
			return
		}

		if !flag {
			log.Warning(" callPlayerService.CheckOnline 未登录playService，clientId:%d", clientId)
			rs.SendMsg(clientId, msg.MsgType_CreateRoomRes, &msg.MsgAddQueueRes{Ret: msg.ErrCode_NotLoginPlayerService})
			return
		}

	})

	if err != nil {
		log.Error("callPlayerService.CheckOnline fail %s,clientId:%s!", err.Error(), clientId)
		rs.SendMsg(clientId, msg.MsgType_CreateRoomRes, &msg.MsgAddQueueRes{Ret: msg.ErrCode_NotLoginPlayerService})
		return false
	}
	return true
}

/*func (rs *RoomService) Create(clientId uint64, roomType int32, info *msg.PlayerInfo) *room.Room {
	newRoom := rs.NewRoom()
	playerInfo := rs.NewPlayerInfo(info)
	//设置为房主
	playerInfo.SetOwner(true)
	newRoom.OnInit(rs.gateProxy, string(clientId), string(clientId), clientId, 1, playerInfo, roomType)
	rs.room[clientId] = newRoom
	return newRoom
}*/

func (rs *RoomService) GetRoom(roomUuid string, roomType int32) (*common.Room, bool) {

	if roomType == 0 {
		//代表没传那就全局检索
		for _, roomMap := range rs.room {
			room, ok := roomMap[roomUuid]
			if ok {
				return room, ok
			}
		}
	} else {
		roomMap, ok := rs.room[roomType]
		if !ok {
			return nil, ok
		}
		room, ok := roomMap[roomUuid]
		return room, ok
	}
	return nil, false
}

func (rs *RoomService) RPC_GetPbRoom(req *rpc.GetRoomReq, res *rpc.GetRoomRes) error {
	roomUuid := req.GetRoomUuid()
	roomType := req.GetRoomType()
	roomMap, ok := rs.room[roomType]
	if !ok {
		return nil
	}
	room := roomMap[roomUuid]
	res.Room = rs.PackRpcRoom(room)
	return nil
}

func (rs *RoomService) SetRoom(roomUuid string, roomType int32, room *common.Room) {
	roomMap, ok := rs.room[roomType]
	if !ok {
		return
	}
	roomMap[roomUuid] = room
}

func (rs *RoomService) GetProxy() *common.GateProxyModule {
	return rs.gateProxy
}

/*func (rs *RoomService) Add(clientId uint64, roomId uint64, info *msg.PlayerInfo) {
	room, ok := rs.room[roomId]
	if !ok {
		log.Release("房间不存在%s", room)
		//不存在
		rs.SendMsg(clientId, msg.MsgType_AddRoomRes, &msg.MsgAddRoomRes{Ret: msg.ErrCode_RoomIdNotExist})
		return
	}
	//放入道理吗
	playerInfo := rs.NewPlayerInfo(info)
	infos := make([]*msg.Player, room.GetRoomClientNum())
	infos = append(infos, rs.packPlayer(playerInfo))
	for _, user := range room.GetOtherUsers() {
		infos = append(infos, rs.packPlayer(user))
	}
	//向房间里面的所有人广播
	log.Release("向房间广播---roomID:%d,%s-%d加入房间", roomId, info.GetNickName(), info.GetClientId())
	for _, user := range infos {
		rs.SendMsg(user.GetClientId(), msg.MsgType_RadioOtherAddRoomRes, &msg.MsgOtherAddRoomRes{PlayerInfo: rs.packPlayer(playerInfo)})
		//组装一个
	}
	//给加入的人广播其他人的消息
	rs.SendMsg(clientId, msg.MsgType_CreateRoomRes, &msg.MsgAddRoomRes{Ret: msg.ErrCode_NotLoginPlayerService, ClientId: info.GetClientId(), Room: rs.packRoom(room), Player: infos})
}*/

func (rs *RoomService) SendMsg(clientId uint64, msgType msg.MsgType, msg proto.Message) {
	rs.gateProxy.SendToClient(clientId, msgType, msg)
}

func (rs *RoomService) PackPlayerInfo(info *entity.PlayerInfo) *msg.PlayerInfo {
	return &msg.PlayerInfo{UserId: info.GetUserId(), Rank: info.GetRank(), NickName: info.GetNickName(), Sex: info.GetSex(), Avatar: info.GetAvatar(), ClientId: info.GetClientId(), IsOwner: info.IsOwner(), SeatNum: info.GetSeatNum()}
}

func (rs *RoomService) PackRpcPlayerInfo(info *entity.PlayerInfo) *rpc.PlayerInfo {
	return &rpc.PlayerInfo{UserId: info.GetUserId(), Rank: info.GetRank(), NickName: info.GetNickName(), Sex: info.GetSex(), Avatar: info.GetAvatar(), ClientId: info.GetClientId(), IsOwner: info.IsOwner(), SeatNum: info.GetSeatNum()}
}

func (rs *RoomService) PackRoom(room *common.Room) *msg.Room {
	owner := rs.PackPlayerInfo(room.GetOwner())
	otherClients := room.GetOtherClients()
	playerInfos := make([]*msg.PlayerInfo, 0)
	for _, client := range otherClients {
		playerInfos = append(playerInfos, rs.PackPlayerInfo(client))
	}
	return &msg.Room{Uuid: room.GetUUid(), RoomName: room.GetRoomName(), RoomType: room.GetRoomType(), AvgRank: room.GetAvgRank(), RoomClientNum: room.GetRoomClientNum(), Owner: owner, OtherClients: playerInfos}
}

func (rs *RoomService) PackRpcRoom(room *common.Room) *rpc.Room {
	owner := rs.PackRpcPlayerInfo(room.GetOwner())
	otherClients := room.GetOtherClients()
	playerInfos := make([]*rpc.PlayerInfo, 0)
	for _, client := range otherClients {
		playerInfos = append(playerInfos, rs.PackRpcPlayerInfo(client))
	}
	return &rpc.Room{Uuid: room.GetUUid(), RoomName: room.GetRoomName(), RoomType: room.GetRoomType(), AvgRank: room.GetAvgRank(), RoomClientNum: room.GetRoomClientNum(), Owner: owner, OtherClients: playerInfos}
}

/*func (rs *RoomService) RPC_RemoveRoom(clientId uint64) {
	log.Release("移除房间--%d", clientId)
	_, ok := rs.room[clientId]
	if ok {
		delete(rs.room, clientId)
	}
}*/

func (rs *RoomService) RemoveRoom(roomUuid string, roomType int32) {
	log.Release("移除房间--%d", roomUuid)
	roomMap, ok := rs.room[roomType]
	if !ok {
		return
	}
	delete(roomMap, roomUuid)
}

func (rs *RoomService) RadioPlayerInfo(room *common.Room) {
	playerInfos := make([]*entity.PlayerInfo, 0)
	resPlayers := make([]*msg.PlayerInfo, 0)
	playerInfos = append(playerInfos, room.GetOwner())
	playerInfos = append(playerInfos, room.GetOtherClients()...)
	for _, user := range playerInfos {
		resPlayers = append(resPlayers, rs.PackPlayerInfo(user))
	}
	//向房间里面的所有人广播
	for _, user := range playerInfos {
		room.SendToClient(user.GetClientId(), msg.MsgType_RadioOtherAddRoomRes, &msg.MsgClientOnRoomRes{RoomPlayer: resPlayers})
	}
}

func (rs *RoomService) NewRoom() *common.Room {
	room := roomPool.Get().(*common.Room)
	return room
}

func (rs *RoomService) SimpleRoomList(roomType int32) []*msg.SimpleRoom {
	simpleRooms := make([]*msg.SimpleRoom, 0)
	if roomType == 0 {
		//全局的返回数据
		for _, roomMap := range rs.room {
			for _, room := range roomMap {
				simpleRooms = append(simpleRooms, rs.NewSimpleRoom(room))
			}
		}
	} else {
		roomMap, ok := rs.room[roomType]
		if ok {
			for _, room := range roomMap {
				simpleRooms = append(simpleRooms, rs.NewSimpleRoom(room))
			}
		}
	}
	return simpleRooms
}

func (rs *RoomService) NewSimpleRoom(room *common.Room) *msg.SimpleRoom {
	return &msg.SimpleRoom{RoomType: room.GetRoomType(), RoomName: room.GetRoomName(), Uuid: room.GetUUid(), RoomClientNum: room.GetRoomClientNum(), AvgRank: room.GetAvgRank()}
}

func (rs *RoomService) NewPlayerInfo(playerInfoPb *msg.PlayerInfo) *entity.PlayerInfo {
	playerInfo := playerInfoPool.Get().(*entity.PlayerInfo)

	playerInfo.SetRank(playerInfoPb.GetRank())
	playerInfo.SetClientId(playerInfoPb.ClientId)
	playerInfo.SetAvatar(playerInfoPb.GetAvatar())
	playerInfo.SetNickName(playerInfoPb.GetNickName())
	playerInfo.SetSex(playerInfo.GetSex())
	playerInfo.SetUserId(playerInfoPb.GetUserId())
	return playerInfo
}
