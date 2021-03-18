package global

import "github.com/duanhf2012/origin/event"

const GateService string = "GateService"
const CenterService string = "CenterService"
const TcpService string = "TcpService"
const PlayerService string = "PlayerService"
const QueueService string = "QueueService"
const MySqlService string = "MySqlService"

//原始Rpc的MethodId定义
const (
	RawRpcMsgDispatch uint32 = 1 //其他服(PlayerService或其他)->GateService->Client,转发消息
	RawRpcCloseClient uint32 = 2 //其他服(PlayerService或其他)->GateService->Client,断开与Client连接
	RawRpcOnRecv      uint32 = 3 //Client->GateService->其他服(PlayerService或其他),转发消息
	RawRpcOnClose     uint32 = 4 //Client->GateService->其他服(PlayerService或其他),转发Client连接断开事件
)

const (
	//配置加载事件
	EventConfigComplete event.EventType = 3001
)
