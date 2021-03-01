package collect

import (
	"encoding/json"
	"github.com/duanhf2012/origin/log"
	"gopkg.in/mgo.v2/bson"
	"sunserver/common/proto/rpc"
)

type MailStatus = int32

const (
	MailStatusDel 			MailStatus = 0 //邮件已经删除
	MailStatusNotReceived	MailStatus = 1 //邮件未领取
	MailStatusReceived		MailStatus = 2 //邮件已经领取
)

const (
	AttachmentsItem uint16 = 1
	AttachmentsHero uint16 = 2
)

type CMailInfo struct {
	Id    		string  					`bson:"_id"`   			//生成最新id
	MailType    int32						`bson:"MailType"`		//邮件类型
	FromUser    uint64						`bson:"FromUser"` 		//发送方User的ID
	SendToUser  uint64  					`bson:"SendToUser"` 	//接收方User的ID
	Title		string  					`bson:"Title"`			//邮件标题
	Content     string  					`bson:"Content"`		//邮件内容
	Attachments []map[string]interface{} 	`bson:"Attachments"`	//附件
	SendTime    int64						`bson:"SendTime"`		//发送时间戳
	Status      MailStatus					`bson:"Status"`			//邮件状态
}

var nameMailInfo = "MailInfo"

func (mail *CMailInfo) GetCollName() string {
	return nameMailInfo
}

func (mail *CMailInfo) Clean() {

}

func (mail *CMailInfo) GetId() interface{} {
	return mail.Id
}

func (mail *CMailInfo) GetCollectionType() MultiCollectionType {
	return MCTUserMail
}

func (mail *CMailInfo) MakeRow() IMultiCollection {
	return &CMailInfo{}
}

func (mail *CMailInfo) GetCondition(value interface{}) bson.D {
	return bson.D{{"SendToUser", value}}
}

func (mail *CMailInfo) OnLoadSucc(notFound bool, userID uint64) {
	log.Debug("CMailInfo OnLoadEnd,mailid:%d notFound:%t.", mail.Id, notFound)
}

func (mail *CMailInfo) OnSave()  {
	log.Debug("CMailInfo OnSave ,mailid:%d.",mail.Id)
}

func ByteToAttachments(data []byte) []map[string]interface{} {
	list := []map[string]interface{}{}
	err := json.Unmarshal(data, &list)
	if err != nil {
		log.Error("ByteToAttachments[%s], err:%+v", string(data), err)
		return nil
	}

	return list
}

//GetIAttachmentsByType 类型注册
func GetIAttachmentsByType(typeInfo uint16) interface{} {
	switch typeInfo {
	case AttachmentsItem:
		return &ItemInfo{}
	case AttachmentsHero:
		return &HeroInfo{}
	default:
		return nil
	}
}

//ItemInfo 附件类型——道具
type ItemInfo struct {
	Id 		uint64	`json:"id" bson:"Id"`
	Type	uint16	`json:"type" bson:"Type"`
	Count   int64	`json:"count" bson:"Count"`
}

//HeroInfo 附件类型——英雄
type HeroInfo struct {
	Id 		uint64	`json:"id" bson:"Id"`
	Type	uint16	`json:"type" bson:"Type"`
}

// CopyMailDataFromPBData rpc.mail数据拷贝到CMailinfo对象中
func CopyMailDataFromPBData(pbData *rpc.UserMailInfo, mailInfo *CMailInfo) {
	mailInfo.Id = pbData.Id
	mailInfo.MailType = pbData.MailType
	mailInfo.FromUser = pbData.FromUser
	mailInfo.SendToUser = pbData.SendToUser
	mailInfo.Title = pbData.Title
	mailInfo.Content = pbData.Content
	mailInfo.SendTime = pbData.SendTime
	mailInfo.Status = pbData.Status
	mailInfo.Attachments = ByteToAttachments(pbData.Attachment)
}
