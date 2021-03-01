package logic

import (
	"github.com/duanhf2012/origin/log"
	"strconv"
	"sunserver/common/proto/msg"
	"sunserver/common/util"
	"sunserver/originhelper/bot"
)

func BotStop(robotObj *bot.BotModule, param ...string) {
	robotObj.SetBotStatus(bot.StopRobot)
}

func BotSendMsg(robotObj *bot.BotModule, param ...string) {
	if len(param) != 2 {
		log.Error("Must be 2 parameters.")
		return
	}

	msgID, err := strconv.Atoi(param[0])
	if err != nil || msgID < 0 {
		log.Error("The first parameter[%d] is error:%+v.", msgID, err)
		return
	}
	msgType := msg.MsgType(msgID)

	msgJsonStr := param[1]
	sendPBStruct := bot.GetProtoStructByMsgID(msgType)
	if sendPBStruct == nil {
		log.Error("The second parameter[%s] is error:no this msg[%d] struct.", msgJsonStr, msgType)
		return
	}
	err = util.JSON2PB(msgJsonStr, sendPBStruct)
	if err != nil {
		log.Error("The second parameter[%s] is error:%+v.", msgJsonStr, err)
		return
	}

	errSend := robotObj.SendMsg(msgType, sendPBStruct)
	if errSend != nil {
		log.Error("robot[%d] send msg to game server")
	}
}