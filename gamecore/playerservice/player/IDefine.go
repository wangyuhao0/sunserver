package player

import (
	"github.com/duanhf2012/origin/util/timer"
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"time"
)

type IPlayer interface {
	ReleasePlayer(p *Player)
	CloseClient(clientId uint64)
}

type ITimer interface {
	TimerAfter(userId uint64,d time.Duration,cb func(timer *timer.Timer)) *timer.Timer
	TimerTicker(userId uint64,d time.Duration,cb func(timer *timer.Ticker)) *timer.Ticker
}

type ISender interface {
	SendToClient(clientId uint64, msgType msg.MsgType, msg proto.Message) error
}

