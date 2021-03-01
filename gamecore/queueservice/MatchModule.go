package queueservice

import (
	"fmt"
	"github.com/duanhf2012/origin/log"
	originrpc "github.com/duanhf2012/origin/rpc"
	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/sysservice/tcpservice"
	"github.com/duanhf2012/origin/util/uuid"
	"github.com/golang/protobuf/proto"
	"sunserver/common/global"
	"sunserver/common/proto/msg"
	"sunserver/common/proto/rpc"
	"sunserver/common/util"
	"sunserver/gamecore/common"
	"sunserver/gamecore/queueservice/def"
)

type MatchModule struct {
	service.Module
}

func NewMatchModule() *MatchModule {
	return &MatchModule{}
}

func (match *MatchModule) OnInit() error {
	return nil
}

//动态规划--递归 解决排队问题

func (match *MatchModule) Match(mapList *def.MapList, need int32) {
	list := mapList.GetDataList()
	clientNums := make([]int, 0)
	tableServiceNodeId := util.GetNodeIdByService("TableService")
	for elem := list.Front(); elem != nil; elem = elem.Next() {
		room1 := elem.Value.(*common.Room)
		//拿到人数
		num := room1.GetRoomClientNum()
		clientNums = append(clientNums, int(num))
	}

	//传入递归
	for {
		sum := 0
		x := make([]bool, len(clientNums))
		result := make([]int, 0)
		trace := backTrace(0, sum, int(need), clientNums, x, len(clientNums), result)
		if trace != nil {
			//说明存在  先从list里面剔除掉 放入到排队队列
			clientList := make([]uint64,0)
			roomUuidList := make([]string,0)
			var roomType int32
			for i := 0; i < len(trace); i++ {
				//往匹配队列里面加
				room := mapList.GetRoomByIndex(i)
				if room==nil {
					log.Release("匹配出现异常-----")
					break
				}
				roomType = room.GetRoomType()
				clientList = append(clientList, room.GetOwnerCid())
				for _, other := range room.GetOtherClients() {
					clientList = append(clientList, other.GetClientId())
				}
				roomUuidList = append(roomUuidList,room.GetUUid())
			}
			//通知匹配成功
			//生成一个预支id -- 用户确认后由客户端传回 然后 服务端组成一个局
			tableUuId := uuid.Rand().HexEx()
			log.Release("tableUuid-%s", tableUuId)
			for _, clientId := range clientList {
				match.SendToClient(clientId, msg.MsgType_MatchRes, &msg.MsgMatchRes{Ret: msg.ErrCode_OK, TableId: tableUuId})
			}
			//移除room在匹配队列
			for _, v := range roomUuidList {
				log.Release("匹配成功移除房间-%s",v)
				mapList.Remove(v)
			}

			//往tableService发送然后先预置一个房间
			var createTable rpc.CreateTable
			createTable.TableType = roomType
			createTable.TableUuid = tableUuId
			createTable.PlayerNum = need
			createTable.RoomUuidList = roomUuidList
			createTable.ShouldConnectedClintList = clientList
			//往桌子里面初始化个桌子先
			err := match.GoNode(tableServiceNodeId, "TableService.RPC_CreateTable", &createTable)
			if err != nil {
				log.Error("go TableService.RPC_CreateTable fail %s", err.Error())
				return
			}
			return
		} else {
			//代表没有数据了可以结束匹配了
			fmt.Println("匹配完成")
			break
		}
	}
}

//递归
func backTrace(n int, sum int, m int, a []int, x []bool, num int, result []int) []int {
	if sum > m {
		return nil
	}
	if len(result) > 0 {
		//出现了结果
		return nil
	}
	if sum == m { //当前和
		for i := 0; i < n; i++ {
			if x[i] {
				result = append(result, i)
			}
		}
		return result
	}
	if n == num {
		return nil
	}

	for i := n; i < num; i++ {
		if x[i] == false {
			x[i] = true
			sum += a[i]
			trace := backTrace(i+1, sum, m, a, x, num, result)
			//回退状态
			if trace != nil {
				return trace
			}
			x[i] = false
			sum -= a[i]
			if i < num-1 && a[i] == a[i+1] {
				i++
				continue
			}
		}
	}
	return nil
}

func (match *MatchModule) SendToClient(clientId uint64, msgType msg.MsgType, msg proto.Message) error {
	//1.获取GateServiceNodeId
	var err error
	nodeId := tcpservice.GetNodeId(clientId)
	if nodeId < 0 || nodeId > tcpservice.MaxNodeId {
		err = fmt.Errorf("nodeid is error %d", nodeId)
		log.Error(err.Error())
		return err
	}

	//2.组装返回消息
	var msgBuff []byte
	if msg != nil {
		msgBuff, err = proto.Marshal(msg)
		if err != nil {
			log.Error("Marshal fail,msgType %d clientId %d.", msgType, clientId)
			return err
		}
	}

	var rawInputArgs global.RawInputArgs
	rawInputArgs.SetMsg(uint16(msgType), clientId, msgBuff)
	err = match.RawGoNode(originrpc.RpcProcessorGoGoPB, nodeId, global.RawRpcMsgDispatch, global.GateService, &rawInputArgs)
	if err != nil {
		log.Error("RawGoNode fail :%s,msgType %d clientId %d.", err.Error(), msgType, clientId)
	}
	return err
}
