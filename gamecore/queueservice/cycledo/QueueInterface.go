package cycledo

import (
	"github.com/duanhf2012/origin/rpc"
	pbRpc "sunserver/common/proto/rpc"
	"sunserver/gamecore/common"
)

type QueueInterface struct {

	QS queueService

}

func New(qs queueService)  *QueueInterface{
	return &QueueInterface{
		QS: qs,
	}
}

type queueService interface {
	GetRpcHandler() rpc.IRpcHandler

	PackRoom(res *pbRpc.GetRoomRes)  *common.Room

	AddQueue(room *common.Room) bool

	AddRoom(room *common.Room)

	GetRoom(roomUuid string)  *common.Room

	RemoveRoom(roomUuid string)

	QuitQueue(room *common.Room)
}

func (qi *QueueInterface) GetRpcHandlerQi() rpc.IRpcHandler {
	return qi.QS.GetRpcHandler()
}

func (qi *QueueInterface) PackRoomQi(res *pbRpc.GetRoomRes)  *common.Room{
	return qi.QS.PackRoom(res)
}

func (qi *QueueInterface) AddQueueQi(room *common.Room) bool {
	return qi.QS.AddQueue(room)
}

func (qi *QueueInterface) AddRoomQi(room *common.Room)  {
	qi.QS.AddRoom(room)
}

func (qi *QueueInterface) GetRoomQi(roomUuid string)  *common.Room{
	return qi.QS.GetRoom(roomUuid)
}

func (qi *QueueInterface) RemoveRoomQi(roomUuid string) {
	qi.QS.RemoveRoom(roomUuid)
}

func (qi *QueueInterface) QuitQueueQi(room *common.Room) {
	qi.QS.QuitQueue(room)
}


