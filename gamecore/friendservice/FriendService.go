package friendservice

import (
	"fmt"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/node"
	"github.com/duanhf2012/origin/service"
	sync2 "github.com/duanhf2012/origin/util/sync"
	"github.com/golang/protobuf/proto"
	"sunserver/common/constpackage"
	"sunserver/common/db"
	"sunserver/common/global"
	"sunserver/common/proto/msg"
	"sunserver/common/util"
	"sunserver/gamecore/common"
	"sunserver/gamecore/friendservice/cycledo"
	"sunserver/gamecore/friendservice/msghandler"
	"time"
)

var friendService FriendService

func init() {
	node.Setup(&friendService)
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
	friendService.mapRegisterMsg[msgType] = &regMsgInfo
}

type FriendService struct {
	service.Service

	mapRegisterMsg map[msg.MsgType]*RegMsgInfo //消息注册

	cycleDo *cycledo.FriendInterface

	gateProxy *common.GateProxyModule //网关代理
}

func (fs *FriendService) OnInit() error {
	//1.初始化变量与模块
	fs.mapRegisterMsg = make(map[msg.MsgType]*RegMsgInfo, 512)

	fs.gateProxy = common.NewGateProxyModule()
	fs.AddModule(fs.gateProxy)
	//2.设置注册函数回调
	msghandler.OnRegisterMessage(RegisterMessage)

	// 注册原始套接字回调
	fs.RegRawRpc(global.RawRpcOnRecv, &RpcOnRecvCallBack{})

	fs.OpenProfiler()
	fs.GetProfiler().SetOverTime(time.Millisecond * 50)
	fs.GetProfiler().SetMaxOverTime(time.Second * 10)

	friendInterface := cycledo.New(fs)
	fs.cycleDo = friendInterface

	return nil
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

	msgInfo := friendService.mapRegisterMsg[msg.MsgType(rawInput.GetMsgType())]
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

	msgInfo, ok := friendService.mapRegisterMsg[msgType]
	if ok == false {
		log.Warning("close client %d,message type %d is not  register.", clientIdList[0], msgType)
		return
	}
	msgInfo.msgCallBack(friendService.cycleDo, clientIdList[0], args.GetProtoMsg().(*protoMsg).msg)
	msgInfo.ReleaseMsg(args.GetProtoMsg().(*protoMsg))
}

func (fs *FriendService) GetFriendList(req *msg.MsgFriendListReq) {

	userId := req.UserId
	platId := req.PlatId

	var mysqlData db.MysqlControllerReq
	mysqlData.TableName = constpackage.FriendTableName
	mysqlData.Key = uint64(util.HashString2Number(platId))
	mysqlData.Sql = "SELECT b.*, a.remark FROM(SELECT a.id, a.user_id,a.add_user_id,a.remark,a.`status`,a.create_time,a.update_time FROM friend AS a WHERE a.user_id = ? AND `status` = 1) a LEFT JOIN `user` b ON a.add_user_id = b.id"
	mysqlData.Args = []string{string(userId)}
	mysqlData.Type = db.OptType_Find
	fs.GetRpcHandler().AsyncCall("MysqlService.RPC_MysqlDBRequest", &mysqlData, func(ret *db.MysqlControllerRet, err error) {

	})

}
