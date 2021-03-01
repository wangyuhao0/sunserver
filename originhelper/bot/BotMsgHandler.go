package bot

import (
	"github.com/duanhf2012/origin/log"
	"github.com/gogo/protobuf/proto"
	"sunserver/common/proto/msg"
	"sunserver/common/util"
)


func (botModule *BotModule) HandlerLoginRes(clientid uint64, msgPb proto.Message) {
	tcpStruct := msgPb.(*msg.MsgLoginRes)
	log.Release("botName[%s] clientID[%d], recive login msg[%+v]", botModule.data.BotName, clientid, tcpStruct)
	log.Release("修改bot状态HandlerLoginRes")
	botModule.status = LoginedGameGate
}

func (botModule *BotModule) HandlerLoginFinish(clientid uint64, msg proto.Message) {
	log.Release("botName[%s] clientID[%d], LoginOK", botModule.data.BotName, clientid)
	botModule.status = LoadDataFinish
}

func (botModule *BotModule) HandlerCreateRoomRes(clientid uint64, msgPb proto.Message) {
	//发送创建房间请求
	msgRes := msgPb.(*msg.MsgCreateRoomRes)
	log.Release("创建房间回包数据---%s---错误码:%d", msgRes.RoomUuid, msgRes.Ret)

	//偷懒 直接加入队列 后续删除
	//botModule.AddQueue(clientid,msgRes.RoomUuid)
}

func (botModule *BotModule) HandlerAddRoomRes(clientid uint64, msgPb proto.Message) {
	//发送创建房间请求
	//str:= ""
	//util.PB2JSON(msgPb,&str)
	msgRes := msgPb.(*msg.MsgAddRoomRes)

	log.Release("加入房间回包数据，错误码:%d", msgRes.Ret)
}

func (botModule *BotModule) HandlerRadioRoomRes(clientid uint64, msgPb proto.Message) {
	str := ""
	util.PB2JSON(msgPb, &str)
	log.Release("房间数据广播--客户端%d,数据:%s", clientid, str)
}

func (botModule *BotModule) HandlerAddQueueRes(clientid uint64, msgPb proto.Message) {
	msgRes := msgPb.(*msg.MsgAddQueueRes)
	log.Release("加入队列 cid:%d,状态码:%d", clientid, msgRes.Ret)
}

func (botModule *BotModule) HandlerQuitQueueRes(clientid uint64, msgPb proto.Message) {
	msgRes := msgPb.(*msg.MsgQuitQueueRes)
	log.Release("退出队列 cid:%d,状态码:%d", clientid, msgRes.Ret)
}

func (botModule *BotModule) HandlerMatchRes(clientid uint64, msgPb proto.Message) {
	msgRes := msgPb.(*msg.MsgMatchRes)
	log.Release("匹配成功 cid:%d,状态码:%d,tableId:%s", clientid, msgRes.Ret, msgRes.TableId)
}

func (botModule *BotModule) HandlerAddTableRes(clientid uint64, msgPb proto.Message) {
	msgRes := msgPb.(*msg.MsgAddTableRes)
	log.Release("加入Table成功 cid:%d,状态码:%d", clientid, msgRes.Ret)
}

func (botModule *BotModule) HandleClientConnectStatusRes(clientid uint64, msgPb proto.Message) {
	msgRes := msgPb.(*msg.MsgClientConnectedStatusRes)
	log.Release("table 连接状态 cid:%d,状态码:%d", clientid, msgRes.Ret)
}

func (botModule *BotModule) HandlerTestRes(clientid uint64, msgPb proto.Message) {
	tcpStruct := msgPb.(*msg.MsgClientSyncTimeRes)
	log.Release("botName[%s] clientID[%d], recive test msg[%+v]", botModule.data.BotName, clientid, tcpStruct)
}

func (botModule *BotModule) HandlerPong(clientid uint64, msgPb proto.Message) {
	log.Release("botName[%s] clientID[%d], HandlerPong", botModule.data.BotName, clientid)
}
