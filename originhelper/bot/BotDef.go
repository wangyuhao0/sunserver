package bot

import (
	"github.com/duanhf2012/origin/event"
	"github.com/golang/protobuf/proto"
	"math"
	"sunserver/common/proto/msg"
)

//服务事件类型
const (
	Cmd_Event_Input   event.EventType = 1001

	//以下是机器人事件
	Cmd_Event_BotStop 	  event.EventType = 2002
)

//机器人事件类型
type BotEventType = int32
const (
	Bot_Event_Connected  	BotEventType = 0 //TCP连结上
	Bot_Event_ReceiveMsg 	BotEventType = 1 //收到TCP消息
	Bot_Event_ReceiveCmd 	BotEventType = 2 //收到CMD指令
	Bot_Event_Disconnected	BotEventType = 3 //TCP连结断开
)

const MaxBotEventChan = 50

const (
	Bot_Status_NO 		int = 0 //机器人尚未添加
	Bot_Status_Logining int = 2 //机器人登陆中
	Bot_Status_Logined	int = 3 //机器人登陆完成
)

type StatusType = int32

const(
	StopRobot StatusType = iota //停止

	LoginingHttpGate   		//HTTP请求中
	LoginedHttpGate 		//HTTP请求完成，准备链接TCP
	ConnectingGameGate  	//TCP链接中
	ConnectedGameGate   	//TCP链接完成，准备登陆
	LoginingGameGate   		//登陆TCP
	LoginedGameGate			//登陆TCP完成，开始接收推送
	LoadDataFinish			//数据加载完成
)

type BotData struct {
	BotName    string

	UserID 		uint64
	Token  		string
	TcpGateUrl 	string
}

//机器人事件及其回调
type BotCmdCB func(robotObj *BotModule, param ...string)
type BotEvent struct {
	Type        BotEventType
	Data 		interface{}
	CallBack 	BotCmdCB
}

//机器人发送消息结构
type BotSendMsgStruct struct {
	MsgID 	msg.MsgType
	MsgBody proto.Message
}

type GateServer struct {
	Weight int
	Url    string
}

type LoginHttpGateData struct {
	UserId 			uint64
	Token  			string
	GateServerUrl 	[]GateServer
}

func (slf *LoginHttpGateData) randServerByWeight() *GateServer {
	if len(slf.GateServerUrl) <= 0 {
		return nil
	}
	minWeight := math.MaxInt32
	index := -1
	for i := 0; i < len(slf.GateServerUrl); i++ {
		if slf.GateServerUrl[i].Weight < minWeight {
			minWeight = slf.GateServerUrl[i].Weight
			index = i
		}
	}

	if index < 0 || index >= len(slf.GateServerUrl) {
		return nil
	}

	return &slf.GateServerUrl[index]
}

//机器人需要发送的消息在此添加
func GetProtoStructByMsgID(msgID msg.MsgType) proto.Message {
	switch msgID {
	case msg.MsgType_ClientSyncTimeReq:
		return &msg.MsgClientSyncTimeReq{}
	default:
		return nil
	}
}