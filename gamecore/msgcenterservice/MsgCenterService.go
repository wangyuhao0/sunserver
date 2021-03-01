package msgcenterservice

import (
	"github.com/duanhf2012/origin/node"
	"github.com/duanhf2012/origin/service"
	"sunserver/common/proto/rpc"
)

func init() {
	node.Setup(&MsgCenterService{})
}

type MsgCenterService struct {
	service.Service

	mailModule MailModule
}

func (slf *MsgCenterService) OnInit() error {
	slf.AddModule(&slf.mailModule)
	return nil
}

// 给指定玩家发送邮件
func (slf *MsgCenterService) RPC_SendMailToUser(mailInfo *rpc.UserMailInfo) error {
	return slf.mailModule.DealNewMail(mailInfo)
}
