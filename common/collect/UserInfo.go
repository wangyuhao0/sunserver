package collect

import (
	"github.com/duanhf2012/origin/log"
	"gopkg.in/mgo.v2/bson"
)

type CUserInfo struct {
	CBaseCollection `bson:"-"`

	Id       uint64 `bson:"_id"`      //生成最新id
	PlatType int    `bson:"PlatType"` //平台类型
	PlatId   string `bson:"PlatId"`   //平台id
	SyncTime int64  `bson:"SyncTime"` //客户端同步时间——测试
	Rank     uint64 `bson:"Rank"`     //排位分
	Coin     uint64 `bson:"Coin"`     //金币
	NickName string `bson:"NickName"` //NickName
	Avatar   string `bson:"Avatar"`   //头像
	Sex      int    `bson:"PlatType"` //性别
	Grade    int    `bson:"Grade"`    // 等级
}

var nameUserInfo = "UserInfo"

func (userInfo *CUserInfo) GetCollName() string {
	return nameUserInfo
}

func (userInfo *CUserInfo) Clean() {
	userInfo.Id = 0
	userInfo.PlatType = 0
	userInfo.PlatId = ""
	userInfo.SyncTime = 0
	userInfo.Avatar = ""
	userInfo.Sex = 0
	userInfo.Grade = 0
	userInfo.Coin = 0
	userInfo.Rank = 0
}

func (userInfo *CUserInfo) GetId() interface{} {
	return userInfo.Id
}

func (userInfo *CUserInfo) GetCollectionType() CollectionType {
	return CTUserInfo
}

func (userInfo *CUserInfo) GetSelf() ICollection {
	return userInfo
}

func (userInfo *CUserInfo) GetCondition(value interface{}) bson.D {
	return bson.D{{"_id", value}}
}

func (userInfo *CUserInfo) OnLoadSucc(notFound bool, userID uint64) {
	if notFound == true {
		userInfo.Id = userID
	} else if userInfo.Id == 0 {
		log.Error("CUserInfo OnLoadSucc err:notFound[%t], userID[%d]", userID)
	}
	//log.Debug("CUserInfo OnLoadEnd,userid:%d notFound:%t.",userInfo.Id,notFound)
}

func (userInfo *CUserInfo) OnSave() {
	//log.Debug("CUserInfo OnSave ,userid:%d.",userInfo.Id)
}
