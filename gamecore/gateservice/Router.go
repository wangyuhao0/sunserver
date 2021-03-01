package gateservice

import (
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/network/processor"
	"github.com/duanhf2012/origin/node"
	originrpc "github.com/duanhf2012/origin/rpc"
	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/sysservice/tcpservice"
	"github.com/golang/protobuf/proto"
	"strconv"
	"strings"
	"sunserver/common/global"
	"sunserver/common/proto/msg"
	"sunserver/common/proto/rpc"
	"sunserver/common/util"
)

type RegCallBack func(cliId uint64, msg []byte)
type StatusType int

const (
	LoginStart StatusType = 0
	Logining  StatusType = 1
	Logined    StatusType = 2
)

type nodeInfo struct {
	nodeId int
	status StatusType
}

type Router struct {
	service.Module

	tcpService           *tcpservice.TcpService     //TcpService服务指针
	mapRegMessage        map[uint16]RegCallBack     //注册消息
	mapRouterCache       map[uint64]*nodeInfo       //map[clientId]nodeId,Client连接信息
	gateService *GateService                        //GateService服务指针
	mapMsgRouterInfo map[*MsgIdRange]*MsgRouterInfo // 存入配置载入的发送消息

	maxConnect int                                  //最大连接数,配置读取
	sendRawPack processor.PBRawPackInfo             //用于发送消息的临时变量
}

type MsgIdRange struct {
	startMsgId uint16
	endMsgId uint16
}

type MsgRouterInfo struct {
	ServiceName string
}


func (r *Router) OnInit() error{
	//1.初始化变量
	r.mapRegMessage = make(map[uint16]RegCallBack, 64)
	r.mapRouterCache = make(map[uint64]*nodeInfo, 4096)
	r.gateService = r.GetService().(*GateService)
	r.tcpService = node.GetService("TcpService").(*tcpservice.TcpService)
	//2.加载配置 配置最大连接数 以及 消息转发
	r.OnLoadCfg()

	//3.注册消息
	r.RegMessage(msg.MsgType_LoginReq,r.loginReq)

	return nil
}

func (r *Router)  OnLoadCfg(){
	r.mapMsgRouterInfo = map[*MsgIdRange]*MsgRouterInfo{}

	cfg := r.gateService.GetServiceCfg()
	configMap,ok := cfg.(map[string]interface{})

	maxConnect,ok := configMap["MaxConnect"]
	if ok == false {
		return
	}
	r.maxConnect = int(maxConnect.(float64))
	if ok == false{
		//error....
		return
	}

	//parse MsgRouter
	routerInfo,ok := configMap["MsgRouter"]
	if ok == false{
		//error...
		return
	}

	//ar routerList []RouterItem
	routerList,ok := routerInfo.([]interface{})
	if ok == false{
		//error...
		return
	}

	for _,v := range routerList{
		mapItem := v.(map[string]interface{})
		var iMsgId interface{}
		var iServiceName interface{}

		if iMsgId,ok = mapItem["MsgId"];ok == false {
			//error ...
			continue
		}
		if iServiceName,ok = mapItem["ServiceName"];ok == false {
			//error ...
			continue
		}
		//消息分类
		msgIdStr:= strings.Split(iMsgId.(string), "-")
		log.Release("msgIdStr:%s",msgIdStr)
		if len(msgIdStr)!=2 {
			continue
		}
		//msgId,ok  := iMsgId.(float64)
		if ok == false {
			//error ...
			continue
		}

		/*//strService := strings.Split(iRpc.(string),".")
		if len(strService)!=2 {
			//error ...
			continue
		}*/
		startMsgId, err := strconv.ParseUint(msgIdStr[0], 10, 16)
		if err !=nil{
			continue
		}
		endMsgId, err := strconv.ParseUint(msgIdStr[1], 10, 16)
		if err !=nil{
			continue
		}
		r.mapMsgRouterInfo[&MsgIdRange{uint16(startMsgId),uint16(endMsgId)}] = &MsgRouterInfo{ServiceName: iServiceName.(string)}
	}


}

func (r *Router) RegMessage(msgType msg.MsgType, cb RegCallBack) {
	r.mapRegMessage[uint16(msgType)] = cb
}

func (r *Router) GetRouterId(clientId uint64) int {
	nodeInfo, ok := r.mapRouterCache[clientId]
	if ok == false {
		return 0
	}
	return nodeInfo.nodeId
}

func (r *Router) RouterMessage(cliId uint64, msgType uint16, msgBuff []byte) {
	log.Release("lfy------RouterMessage cid[%d], msgType[%d], msg[%+v]", cliId, msgType, msgType)
	//1.查找是否为注册消息
	if cb, ok := r.mapRegMessage[msgType]; ok == true {
		cb(cliId, msgBuff[2:])
		r.tcpService.ReleaseNetMem(msgBuff)
		return
	}

	//2.通过clientId获取nodeId
	nodeId := r.GetRouterId(cliId)
	if nodeId == 0 {
		r.tcpService.ReleaseNetMem(msgBuff)
		log.Warning("cannot find clientId %d",cliId)
		r.tcpService.Close(cliId)
		return
	}

	//3.组装原始Rpc参数用于转发
	var inputArgs global.RawInputArgs
	inputArgs.SetMsg(msgType,cliId,msgBuff[2:])
	r.tcpService.ReleaseNetMem(msgBuff)

	//4.转发消息
	r.SendMsgToRpc(nodeId,msgType,&inputArgs)
	/*err := r.gateService.RawGoNode(originrpc.RpcProcessorPb, nodeId, global.RawRpcOnRecv, global.PlayerService, &inputArgs)
	if err != nil {
		log.Error("RawGoNode fail %s",err.Error())
	}*/
}

func (r *Router) SendMsgToRpc(nodeId int,msgType uint16,args *global.RawInputArgs){
	msgRouterInfo := r.mapMsgRouterInfo
	// 通过msgType 找到需要转发的
	for k,v  :=range msgRouterInfo{
		if msgType>=k.startMsgId&&msgType<=k.endMsgId {
			//在范围之中
			log.Release("msgType:%d,sendMsgToRpc-startMsgId:%d , endMsgId:%d , serviceName:%s",msgType,k.startMsgId,k.endMsgId,v.ServiceName)
			err := r.gateService.RawGoNode(originrpc.RpcProcessorGoGoPB, nodeId, global.RawRpcOnRecv, v.ServiceName, args)
			if err != nil {
				log.Error("RawGoNode fail %s",err.Error())
			}
			return
		}
	}
	log.Error("SendMsgToRpc fail----")
}


func (r *Router) OnConnected(clientId uint64){
	log.Debug("connect clientId %d.",clientId)
	r.mapRouterCache[clientId] = &nodeInfo{status: LoginStart,nodeId: 0}
}

func (r *Router) OnDisconnected(clientId uint64) {
	log.Debug("disconnect clientId %d.",clientId)

	//1.查找路由
	nodeId := r.GetRouterId(clientId)
	delete(r.mapRouterCache,clientId)
	if nodeId == 0 {
		log.Error("cannot find clientId %d",clientId)
		return
	}

	//2.转发客户端连接断开
	var inputArgs global.RawInputArgs
	inputArgs.SetUint64(clientId)
	r.gateService.RawGoNode(originrpc.RpcProcessorGoGoPB, nodeId, global.RawRpcOnClose, global.PlayerService, &inputArgs)
}

func (r *Router) loginOk(cliId uint64,nodeId int,userId uint64){
	var req rpc.LoginToPlayerServiceReq
	req.UserId = userId
	req.NodeId = int32(node.GetNodeId())
	req.ClientId = cliId

	//1.修改状态为Logining
	v,ok := r.mapRouterCache[cliId]
	if ok == false {
		log.Release("Client is close cancel login ")
		return
	}

	//不允许重入
	if v.status != LoginStart {
		var loginRes msg.MsgLoginRes
		log.Error("status error.%d",v.status)
		loginRes.Ret = msg.ErrCode_RepeatLoginReq
		r.SendMsg(cliId,msg.MsgType_LoginRes,&loginRes)
		return
	}

	v.nodeId = nodeId
	v.status = Logining

	//2.向选择好的PlayerService服发起登陆
	err := r.gateService.AsyncCallNode(nodeId,"PlayerService.RPC_Login",&req,func(res *rpc.LoginToPlayerServiceRet,err error){
		var loginRes msg.MsgLoginRes
		if err != nil {
			loginRes.Ret = msg.ErrCode_InterNalError
			r.SendMsg(cliId,msg.MsgType_LoginRes,&loginRes)
			return
		}

		if res.Ret!=0 {
			loginRes.Ret = msg.ErrCode_InterNalError
			r.SendMsg(cliId,msg.MsgType_LoginRes,&loginRes)
			return
		}

		//判断连接是否存在
		v,ok := r.mapRouterCache[cliId]
		if ok == false {
			log.Warning("Client is close cancel login ")
			return
		}

		//登陆成功
		v.status = Logined
		loginRes.Ret = msg.ErrCode_OK
		r.SendMsg(cliId,msg.MsgType_LoginRes,&loginRes)
	})

	if err != nil {
		log.Error("PlayerService.RPC_Login is error :%s",err.Error())
		var loginRes msg.MsgLoginRes
		loginRes.Ret = msg.ErrCode_InterNalError
		r.SendMsg(cliId,msg.MsgType_LoginRes,&loginRes)
	}
}

func (r *Router) SendMsg(clientId uint64,msgType msg.MsgType,msg proto.Message) error{
	bytes,err := proto.Marshal(msg)
	if err != nil {
		log.Error("proto.Marshal fail :%s",err.Error())
		return err
	}
	//log.Release("lfy------send to client[%d], type[%d], body[%+v]", clientId, msgType, msg)

	r.sendRawPack.SetPackInfo(uint16(msgType),bytes)
	err = r.tcpService.SendMsg(clientId,&r.sendRawPack)
	if err != nil {
		log.Error("SendMsg fail %s",err.Error())
	}

	return nil
}

func (r *Router) loginReq(cliId uint64,msgBuff []byte){
	log.Release("进入到Router-loginReq")
	//1.判断最大连接数，如果该情况发生客户端应该等待重试
	if len(r.mapRouterCache) >= r.maxConnect {
		var loginRes msg.MsgLoginRes
		loginRes.Ret = msg.ErrCode_ConnExceeded
		r.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
		log.Error("Maximum number of connections exceeded")
	}

	//2.解析消息
	var msgLoginReq msg.MsgLoginReq
	err := proto.Unmarshal(msgBuff,&msgLoginReq)
	if err != nil {
		log.Error("LoginReq fail,Unmarshal error:%s",err.Error())
		r.gateService.Close(cliId)
		return
	}
	//log.Release("lfy -------------LoginReq[%d]:receive msg[%+v]", msgLoginReq.UserId, &msgLoginReq)

	//3.选择主中心服
	masterNodeId := util.GetMasterCenterNodeId()
	if masterNodeId == 0 {
		log.Error("Cannot get centerservice service.")
		r.gateService.Close(cliId)
		return
	}

	//4.从中心服验证Token
	var req rpc.LoginGateCheckReq
	req.UserId = msgLoginReq.UserId
	req.Token = msgLoginReq.Token
	err = r.gateService.AsyncCallNode(masterNodeId,"CenterService.RPC_Login",&req,func(res *rpc.LoginGateCheckRet,err error) {
		var loginRes msg.MsgLoginRes
		if err != nil {
			loginRes.Ret = msg.ErrCode_InterNalError
			log.Release("Route发出信息，在bot模块回调修改状态")
			r.SendMsg(cliId,msg.MsgType_LoginRes,&loginRes)
			log.Error("AsyncCallNode CenterService.RPC_Login fail :%s",err.Error())
			return
		}

		if res.Ret!= 0 {
			loginRes.Ret = msg.ErrCode_TokenError
			r.SendMsg(cliId,msg.MsgType_LoginRes,&loginRes)
			log.Error("AsyncCallNode CenterService.RPC_Login fail ret:%d",res.Ret)
			return
		}

		if res.NodeId <= 0 {
			loginRes.Ret = msg.ErrCode_InterNalError
			r.SendMsg(cliId,msg.MsgType_LoginRes,&loginRes)
			log.Error("AsyncCallNode CenterService.RPC_Login fail cannot find nodeid")
			return
		}

		//验证通过
		r.loginOk(cliId,int(res.NodeId), msgLoginReq.UserId)
	})

	//5.失败返回失败结果
	if err!= nil {
		var loginRes msg.MsgLoginRes
		log.Error("AsyncCallNode CenterService.RPC_Login fail :%s",err.Error())
		loginRes.Ret = msg.ErrCode_InterNalError
		r.SendMsg(cliId,msg.MsgType_LoginRes,&loginRes)
	}
}