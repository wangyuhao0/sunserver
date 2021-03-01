package entity

type PlayerInfo struct {
	rank uint64
	nickName string
	avatar string
	sex int32
	userId uint64
	clientId uint64
	isOwner bool
	seatNum int32
}



func (p *PlayerInfo) GetRank() uint64{
	return p.rank
}

func (p *PlayerInfo) SetRank(rank uint64) {
	p.rank = rank
}

func (p *PlayerInfo) GetNickName() string{
	return p.nickName
}

func (p *PlayerInfo) SetNickName(nickName string) {
	p.nickName = nickName
}

func (p *PlayerInfo) GetAvatar() string{
	return p.avatar
}

func (p *PlayerInfo) SetAvatar(avatar string) {
	p.avatar = avatar
}

func (p *PlayerInfo) GetSex() int32{
	return p.sex
}

func (p *PlayerInfo) SetSex(sex int32) {
	p.sex = sex
}

func (p *PlayerInfo) GetUserId() uint64{
	return p.userId
}

func (p *PlayerInfo) SetUserId(userId uint64) {
	p.userId = userId
}

func (p *PlayerInfo) GetClientId() uint64{
	return p.clientId
}

func (p *PlayerInfo) SetClientId(clientId uint64) {
	p.clientId = clientId
}

func (p *PlayerInfo) IsOwner() bool{
	return p.isOwner
}

func (p *PlayerInfo) SetOwner(isOwner bool) {
	p.isOwner = isOwner
}

func (p *PlayerInfo) SetSeatNum(seatNum int32){
	p.seatNum = seatNum
}

func (p *PlayerInfo) GetSeatNum() int32{
	return p.seatNum
}