package player

import (
	"fmt"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/util/timer"
	"sunserver/common/collect"
	"sunserver/common/proto/msg"
	"sunserver/common/proto/rpc"
	"time"
)

//用于存于不需要持久化的数据
type DataInfo struct {
	fromGateId int    	//关联网关nodeId
	clientId uint64   	//clientId
	isOnline bool     	//是否在线状态

	pingTime time.Time   //ping时间
	saveTime time.Time   //上次存档时间

	pintTicker *timer.Ticker
}


//清理内存变量
func (dataInfo *DataInfo) Clear(){
	dataInfo.fromGateId = 0
	dataInfo.clientId = 0
	dataInfo.isOnline = false
	dataInfo.pingTime = time.Now()
	dataInfo.saveTime =dataInfo.pingTime

	if dataInfo.pintTicker!= nil {
		log.Warning("ticker not close?")
		dataInfo.pintTicker.Cancel()
	}

	dataInfo.pintTicker = nil
}

func (p *Player) OnInitEnd(){

}

//对象释放时
func (p *Player) OnRelease() {
	if p.pintTicker!= nil {
		p.pintTicker.Cancel()
		p.pintTicker = nil
	}
}

//登陆完成，或者重登陆发送所有数据
func (p *Player) SendAllPlayerData(){
	//
	/*collectionType:= p.PlayerDB.GetCollectionType()
	p.PlayerDB.GetRowByKey(collectionType,p.)*/
}

func (p *Player) DealClientSyncTimeMsg(msgInfo *msg.MsgClientSyncTimeReq) {
	p.SyncTime = msgInfo.SyncInfo
	p.CUserInfo.MakeDirty()
}


func (p *Player) ReceiveMail(mailInfo *rpc.UserMailInfo){
	newMail := collect.CMailInfo{}
	collect.CopyMailDataFromPBData(mailInfo, &newMail)
	if p.InsertMultiRow(&newMail, false) == false {
		err := fmt.Errorf("PlayerService.RPC_NoticeUserMail mail[%+v], err:insert multiRow err", mailInfo)
		log.Error("%s", err.Error())
	}
}


