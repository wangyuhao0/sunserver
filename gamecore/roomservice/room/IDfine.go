package room

import (
	"github.com/duanhf2012/origin/util/timer"
	"github.com/golang/protobuf/proto"
	"sunserver/common/proto/msg"
	"time"
)

type IRoom interface {
	CloseRoom(uuid string, roomType int32)
}

type ITimer interface {
	TimerAfter(roomType int32, uuid string, d time.Duration, cb func(timer *timer.Timer)) *timer.Timer
	TimerTicker(roomType int32, uuid string, d time.Duration, cb func(timer *timer.Ticker)) *timer.Ticker
}

type ISender interface {
	SendToClient(clientId uint64, msgType msg.MsgType, msg proto.Message) error
}
