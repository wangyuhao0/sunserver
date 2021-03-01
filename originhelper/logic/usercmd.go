package logic

import (
	"fmt"
	"strconv"
	"sunserver/common/proto/rpc"
)

func SendMailToUser(helper IHelper,args... string) error {
	if len(args) <= 3 {
		return fmt.Errorf("Invalid parameter.")
	}

	sendToUser, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || sendToUser <= 0 {
		return fmt.Errorf("The first parameter is error.")
	}

	mailInfo := rpc.UserMailInfo{
		MailType:             0,
		FromUser:             0,
		SendToUser:           uint64(sendToUser),
		Title:                args[1],
		Content:              args[2],
		Attachment:			  []byte{},
	}

	if len(args) == 4 {
		mailInfo.Attachment = []byte(args[3])
	}

	err = helper.Go("MsgCenterService.RPC_SendMailToUser", &mailInfo)
	if err != nil {
		return err
	}

	return nil
}
