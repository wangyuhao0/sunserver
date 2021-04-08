package player

import (
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/rpc"
	"github.com/duanhf2012/origin/util/timer"
	"sunserver/common/proto/msg"
	"sunserver/gamecore/playerservice/dbcollection"
	"time"
)

type Player struct {
	dbcollection.PlayerDB
	DataInfo
	ITimer
	ISender
	IPlayer
}

func (p *Player) OnInit(rpcHandler rpc.IRpcHandler, timer ITimer, sender ISender, iPlayer IPlayer) {
	p.PlayerDB.OnInit(rpcHandler, p.OnLoadDBEnd)
	p.ISender = sender
	p.ITimer = timer
	p.IPlayer = iPlayer
	p.pintTicker = p.TimerTicker(p.GetUserId(), time.Second*10, p.CheckTimeout)

	p.OnInitEnd()
}

func (p *Player) Load() {
	p.PlayerDB.LoadFromDB()
	/*baseCollection:= p.PlayerDB.
	collection := p.PlayerDB.GetRowByKey(collect.MCTMax, p.clientId)
	collection.
	p.Rank = collection.*/
}

func (p *Player) OnLoadDBEnd(succ bool) {
	//1.加载失败，释放对象
	if succ == false {
		log.Release("加载失败，释放对象")
		p.ReleasePlayer(p)
		return
	}

	//2.向玩家发送所有数据
	p.SendAllPlayerData()

	//3.通知所有数据加载完成
	p.SendLoadFinish()

	//4.记录当前存档时间
	p.saveTime = time.Now()
}

func (p *Player) Clear() {
	p.PlayerDB.Clear()
	p.DataInfo.Clear()
}

func (p *Player) StartLogin(cliId uint64, userId uint64, fromGateId int) {
	p.fromGateId = fromGateId
	p.clientId = cliId
	p.Id = userId
	p.Ping()
	p.Load()
}

func (p *Player) GetFromGateId() int {
	return p.fromGateId
}

func (p *Player) GetClientId() uint64 {
	return p.clientId
}

func (p *Player) GetRank() uint64 {
	return p.Rank
}

func (p *Player) GetUserId() uint64 {
	return p.PlayerDB.Id
}

func (p *Player) SetUserId(id uint64) {
	p.PlayerDB.Id = id
}

func (p *Player) ReLogin(cliId uint64, userId uint64, fromGateId int) {
	//1.重置之前的连接
	log.Release("ReLogin %d", cliId)
	p.fromGateId = fromGateId
	p.clientId = cliId
	p.Id = userId
	//重置用户连接时间
	p.Ping()
	//2.发送重登陆数据
	if p.PlayerDB.IsLoadFinish() {
		p.SendAllPlayerData()
	}
}

func (p *Player) SetOnline(online bool) {
	p.isOnline = online
}

func (p *Player) GetOnline() bool {
	return p.isOnline
}

func (p *Player) downLineTimeout() {
	//下线超时释放
	log.Release("下线超时释放")
	if p.PlayerDB.IsLoadFinish() {
		p.OnRelease()
	}

	//下线存档
	p.PlayerDB.SaveToDB(true)

	//释放对象
	p.ReleasePlayer(p)
}

func (p *Player) SendLoadFinish() {
	p.SendToClient(p.GetClientId(), msg.
		MsgType_LoadFinish, &msg.MsgNil{})
}

func (p *Player) Close() {
	if p.GetClientId() > 0 {
		log.Release("player Close")
		p.CloseClient(p.GetClientId())
		p.clientId = 0
	}
}

func (p *Player) CheckTimeout(ticker *timer.Ticker) {
	//1.检查ping/pong超时
	now := time.Now()
	timeOut := now.Sub(p.pingTime)

	//ping/pong超过x秒，断开连接
	if p.GetClientId() > 0 && timeOut > time.Second*10 {
		p.Close()
	} else if timeOut > time.Minute*30 {
		//离线超过x分钟，释放玩家
		p.downLineTimeout()
		return
	}

	//2.检查是否需要存档
	if p.saveTime.IsZero() == false && now.Sub(p.saveTime) > time.Minute*2 {
		p.saveTime = now
		p.SaveToDB(false)
	}
}

func (p *Player) Ping() {
	p.pingTime = time.Now()
}
