package playerservice

import (
	"fmt"
	"github.com/duanhf2012/origin/log"
	"sunserver/common/proto/rpc"
)

//向玩家发送邮件服
func (ps *PlayerService) RPC_NoticeUserMail(mailInfo *rpc.UserMailInfo) error {
	p, ok := ps.mapPlayer[mailInfo.SendToUser]
	if ok == false || p == nil {
		err := fmt.Errorf("PlayerService.RPC_NoticeUserMail mail[%+v], err:no this user[%d]", mailInfo, mailInfo.SendToUser)
		log.Error("%s", err.Error())
		return err
	}

	p.ReceiveMail(mailInfo)

	return nil
}
