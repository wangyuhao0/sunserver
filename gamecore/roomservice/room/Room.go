package room

import (
	"github.com/duanhf2012/origin/util/timer"
	"sunserver/common/entity"
	"time"
)

type Room struct {
	ISender
	ITimer
	DataInfo
	IRoom
	uuid          string //房间uuid
	roomName      string //房间名字
	ownerCid      uint64
	roomType      int32  //房间类型
	avgRank       uint64 //房间平均分数
	roomClientNum int32  //房间人数
	owner         *entity.PlayerInfo
	otherClients  []*entity.PlayerInfo
}

func (r *Room) OnInit(sender ISender, timer ITimer, uuid string, roomName string, roomClientNum int32, owner *entity.PlayerInfo, roomType int32) {
	r.ISender = sender
	r.uuid = uuid
	r.roomName = roomName
	r.ownerCid = owner.GetClientId()
	r.roomClientNum = roomClientNum
	r.owner = owner
	r.avgRank = r.owner.GetRank()
	r.roomType = roomType
	r.ITimer = timer
	r.pintTicker = r.TimerTicker(r.GetRoomType(), r.GetUUid(), time.Second*5, r.CheckTimeout)
}

func (r *Room) CheckTimeout(ticker *timer.Ticker) {
	//1.说明没人
	if r.roomClientNum == 0 {
		r.CloseRoom(r.GetUUid(), r.roomType)
		r.Clear()
	}
}

func (r *Room) PackFromPb(sender ISender, uuid string, roomName string, roomClientNum int32, owner *entity.PlayerInfo, other []*entity.PlayerInfo, roomType int32, avgRank uint64) {
	r.ISender = sender
	r.uuid = uuid
	r.roomName = roomName
	r.ownerCid = owner.GetClientId()
	r.roomClientNum = roomClientNum
	r.owner = owner
	r.otherClients = other
	r.avgRank = avgRank
	r.roomType = roomType
}

func (r *Room) GetUUid() string {
	return r.uuid
}

func (r *Room) SetUUid(uuid string) {
	r.uuid = uuid
}

func (r *Room) GetRoomName() string {
	return r.roomName
}

func (r *Room) SetRoomName(roomName string) {
	r.roomName = roomName
}

func (r *Room) GetOwnerCid() uint64 {
	return r.ownerCid
}

func (r *Room) SetOwnerCid(ownerCid uint64) {
	r.ownerCid = ownerCid
}

func (r *Room) GetRoomClientNum() int32 {
	return r.roomClientNum
}

func (r *Room) SetRoomClientNum(roomClientNum int32) {
	r.roomClientNum = roomClientNum
}

func (r *Room) GetOwner() *entity.PlayerInfo {
	return r.owner
}

func (r *Room) SetOwner(owner *entity.PlayerInfo) {
	r.owner = owner
}

func (r *Room) GetOtherClients() []*entity.PlayerInfo {
	return r.otherClients
}

func (r *Room) SetOtherClients(otherClients []*entity.PlayerInfo) {
	r.otherClients = otherClients
}

func (r *Room) GetAvgRank() uint64 {
	return r.avgRank
}

func (r *Room) SetAvgRank(avgRank uint64) {
	r.avgRank = avgRank
}

func (r *Room) GetRoomType() int32 {
	return r.roomType
}

func (r *Room) SetRoomType(roomType int32) {
	r.roomType = roomType
}

//匹配出来座位号
func (r *Room) GetSeatNum() int32 {
	owner := r.GetOwner()
	clients := r.GetOtherClients()
	num := owner.GetSeatNum()
	maxNum := int32(0)
	if num > maxNum {
		maxNum = num
	}
	for _, client := range clients {
		if maxNum < client.GetSeatNum() {
			maxNum = client.GetSeatNum()
		}
	}
	return maxNum + 1
}
