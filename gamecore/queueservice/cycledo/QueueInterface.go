package cycledo

import (
	"github.com/duanhf2012/origin/rpc"
	pbRpc "sunserver/common/proto/rpc"
	"sunserver/gamecore/roomservice/room"
)

type QueueInterface struct {
	QS queueService
}

func New(qs queueService) *QueueInterface {
	return &QueueInterface{
		QS: qs,
	}
}

type queueService interface {
	GetRpcHandler() rpc.IRpcHandler

	PackRoom(res *pbRpc.GetRoomRes) *room.Room

	AddQueue(room *room.Room) bool

	AddRoom(room *room.Room)

	GetRoom(roomUuid string) *room.Room

	RemoveRoom(roomUuid string)

	QuitQueue(room *room.Room)
}

func (qi *QueueInterface) GetRpcHandlerQi() rpc.IRpcHandler {
	return qi.QS.GetRpcHandler()
}

func (qi *QueueInterface) PackRoomQi(res *pbRpc.GetRoomRes) *room.Room {
	return qi.QS.PackRoom(res)
}

func (qi *QueueInterface) AddQueueQi(room *room.Room) bool {
	return qi.QS.AddQueue(room)
}

func (qi *QueueInterface) AddRoomQi(room *room.Room) {
	qi.QS.AddRoom(room)
}

func (qi *QueueInterface) GetRoomQi(roomUuid string) *room.Room {
	return qi.QS.GetRoom(roomUuid)
}

func (qi *QueueInterface) RemoveRoomQi(roomUuid string) {
	qi.QS.RemoveRoom(roomUuid)
}

func (qi *QueueInterface) QuitQueueQi(room *room.Room) {
	qi.QS.QuitQueue(room)
}
