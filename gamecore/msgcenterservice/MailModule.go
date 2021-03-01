package msgcenterservice

import (
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/util/uuid"
	"sunserver/common/collect"
	"sunserver/common/db"
	"sunserver/common/proto/rpc"
	"sunserver/common/util"
	"time"
)

type MailModule struct {
	service.Module
}

func (slf *MailModule) OnInit() error{
	return nil
}

func (slf *MailModule) DealNewMail(mailInfo *rpc.UserMailInfo) error {
	//1.初始值设定
	mailInfo.Id = uuid.Rand().HexEx()
	mailInfo.Status = collect.MailStatusNotReceived
	mailInfo.SendTime = time.Now().Unix()
	//临时放俩附件

	//2.存放数据库
	mailColl := collect.CMailInfo{}
	collect.CopyMailDataFromPBData(mailInfo, &mailColl)
	var req db.DBControllerReq
	err := db.MakeInsertId(mailColl.GetCollName(), mailColl, mailInfo.SendToUser, &req)
	if err != nil {
		log.Error("make upsetid fail %s", err.Error())
		return err
	}
	dbNodeID := util.GetBestNodeId("DBService.RPC_DBRequest", mailInfo.SendToUser)
	slf.AsyncCallNode(dbNodeID, "DBService.RPC_DBRequest", &req, func(res *db.DBControllerRet,err error) {
		if err != nil {
			log.Error("AsyncCall error :%+v\n",err)
			return
		}

		//3.邮件存放DB成功,调用回调,通知到center服
		slf.mailSaveToDBCallBack(mailInfo)
	})

	return nil
}

func (slf *MailModule) mailSaveToDBCallBack(mailInfo *rpc.UserMailInfo) {
	//1.先获取玩家所在的NodeID
	queryInfo := rpc.QueryUserNodeID{UserID: mailInfo.SendToUser}
	centerID := util.GetSlaveCenterNodeId()

	errCallCenter := slf.AsyncCallNode(centerID, "CenterService.RPC_QueryUserNodeID", &queryInfo, func(resultInfo *rpc.QueryUserNodeIDRet, err error) {
		if err != nil {
			log.Error("MsgCenterService.mailSaveToDBCallBack, call[%d][CenterService.RPC_QueryUserNodeID], user[%d], err:%+v", centerID, mailInfo.SendToUser, err)
			return
		}

		//2.发送到玩家身上
		if resultInfo.NodeID > 0 {
			errSendMail := slf.GoNode(int(resultInfo.NodeID), "PlayerService.RPC_NoticeUserMail", mailInfo)
			if errSendMail != nil {
				log.Error("MsgCenterService.mailSaveToDBCallBack, go[%d][PlayerService.RPC_NoticeUserMail], mail[%+v], err:%+v", resultInfo.NodeID, mailInfo, errSendMail)
			}
		}
	})

	if errCallCenter != nil {
		log.Error("MsgCenterService.mailSaveToDBCallBack, call[%d][CenterService.RPC_QueryUserNodeID], user[%d], err:%+v", centerID, mailInfo.SendToUser, errCallCenter)
		return
	}
}