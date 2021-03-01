package centerservice

import (
	"fmt"
	"github.com/duanhf2012/origin/cluster"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/node"
	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/util/timer"
	"github.com/duanhf2012/origin/util/uuid"
	"sunserver/common/proto/rpc"
	"time"
)

type UserInfo struct {
	Status rpc.LoginStatus //0正在登录，1登录成功
	Token string
	PlayerServiceNodeId int
}

type BalanceInfo struct {
	refreshTime time.Time
	weight int32
}

type CenterService struct {
	service.Service

	mapUserInfo map[uint64]UserInfo
	mapLogining map[uint64]time.Time
	mapBalance  map[int]*BalanceInfo
}

func init(){
	node.Setup(&CenterService{})
}

func (cs *CenterService) OnInit() error {
	cs.mapUserInfo = make(map[uint64]UserInfo,100000)
	cs.mapBalance = make(map[int]*BalanceInfo,20)
	cs.mapLogining = make(map[uint64]time.Time,100)
	cs.NewTicker(time.Minute*1,cs.CheckLoginTimeout)

	//性能监控
	cs.OpenProfiler()
	cs.GetProfiler().SetOverTime(time.Millisecond * 100)
	cs.GetProfiler().SetMaxOverTime(time.Second * 10)

	return nil
}

// 定时检查x分钟内，状态依然没有成功的玩家
func (cs *CenterService) CheckLoginTimeout(timer *timer.Ticker){
	now:=time.Now()
	for uId,v := range cs.mapLogining {
		if now.Sub(v) > time.Minute*6{
			delete( cs.mapLogining,uId)
			v,ok := cs.mapUserInfo[uId]
			if ok == true && v.Status == rpc.LoginStatus_Logined {
				//这种状态不应该发生
				log.Error("mapUserInfo status is %d,but the userid still exists mapLogining",v.Status)
			}else{
				delete(cs.mapUserInfo,uId)
			}
		}
	}
}

// 查找负载最低的PlayerService
func (cs *CenterService) getBestPlayerServiceNodeId() int {
	now := time.Now()
	bestNodeId := 0
	weight := int32(0)
	for nodeId,balance := range cs.mapBalance{
		//如果是断开连接或者30秒内未汇报负载情况的服务不再分配
		if cluster.GetCluster().IsNodeConnected(nodeId) == false ||
			now.Sub(balance.refreshTime)> time.Second*30 {
			continue
		}

		//选择负载最小
		if bestNodeId == 0 || balance.weight<weight {
			bestNodeId = nodeId
			weight = balance.weight
		}
	}

	return bestNodeId
}

// HttpGateService登陆选服
func (cs *CenterService) RPC_ChoseServer(req *rpc.ChoseServerReq,res *rpc.ChoseServerRet) error {
	log.Release("进入到CenterService-RPC_ChoseServer")
	v,ok := cs.mapUserInfo[req.UserId]
	if ok == false {
		//生成Token
		res.Token = uuid.Rand().HexEx()
		//选择最优NodeId
		nodeId := cs.getBestPlayerServiceNodeId()
		if nodeId == 0 {
			res.Ret = 1
			return nil
		}
		//缓存登陆状态
		cs.mapUserInfo[req.UserId] = UserInfo{Status:rpc.LoginStatus_LoginStart,Token:res.Token,PlayerServiceNodeId: nodeId}
		cs.mapLogining[req.UserId] = time.Now()
		return nil
	}

	//如果已经存在，直接返回之前结果
	v.Status = rpc.LoginStatus_LoginStart
	res.Ret = 0
	res.Token = v.Token
	cs.mapLogining[req.UserId] = time.Now()

	return nil
}

// GateService登陆验证token
func (cs *CenterService) RPC_Login(req *rpc.LoginGateCheckReq,res *rpc.LoginGateCheckRet) error {
	// 1.查找登陆缓存
	log.Release("进入到了CenterService-RPCLogin")
	v,ok := cs.mapUserInfo[req.UserId]
	if ok == false {
		res.Ret = 1
		return nil
	}

	//2.验证Token
	if v.Token!= req.Token {
		res.Ret = 2
		return nil
	}

	//3.返回成功，并修改状态正在登陆中
	v.Status = rpc.LoginStatus_Logining
	cs.mapUserInfo[req.UserId] = v
	cs.mapLogining[req.UserId] = time.Now()
	res.NodeId = int32(v.PlayerServiceNodeId)

	return nil
}

//登陆PlayerService与Player对象释放时
func (cs *CenterService) RPC_UpdateStatus(playerStatus *rpc.UpdatePlayerStatus) error {
	//只能这两种状态才能同步
	log.Release("进入到CenterService-RPC_UpdateStatus")
	if playerStatus.Status != rpc.LoginStatus_Logined && playerStatus.Status!= rpc.LoginStatus_LoginOut {
		return fmt.Errorf("status is error")
	}

	//如果是登出，删除相应的缓存
	//cs.mapLogining删除可能会遇到客户端正在请求登陆的情况，这样客户端登陆失败，重新走登陆流程即可
	delete(cs.mapLogining,playerStatus.UserId)
	if playerStatus.Status == rpc.LoginStatus_LoginOut {
		delete(cs.mapUserInfo,playerStatus.UserId)
		return nil
	}

	//如果是登陆成功，修改保存状态
	v,ok := cs.mapUserInfo[playerStatus.UserId]
	if ok == false {
		cs.mapUserInfo[playerStatus.UserId] =UserInfo{Status:rpc.LoginStatus_Logined,Token: uuid.Rand().HexEx(),PlayerServiceNodeId: int(playerStatus.NodeId)}
	}else{
		v.Status = rpc.LoginStatus_Logined
		v.PlayerServiceNodeId = int(playerStatus.NodeId)
		cs.mapUserInfo[playerStatus.UserId] = v
	}

	return nil
}

// PlayerService服同步负载情况
func (cs *CenterService) RPC_UpdateBalance(balance *rpc.PlayerServiceBalance) error{
	log.Release("进入到CenterSerice-RPC_UpdateBalance")
	nodeId := int(balance.NodeId)
	v,ok := cs.mapBalance[nodeId]
	if ok == false {
		cs.mapBalance[nodeId] = &BalanceInfo{
			refreshTime: time.Now(),
			weight:      balance.Weigh,
		}
		return nil
	}

	v.weight = balance.Weigh
	v.refreshTime = time.Now()

	return nil
}

// PlayerService服全同步所有玩家列表
func (cs *CenterService) RPC_UpdateUserList(playerList *rpc.UpdatePlayerList) error{
	//清理掉所有的player对象
	for uId,UInfo := range cs.mapUserInfo {
		if UInfo.PlayerServiceNodeId == int(playerList.NodeId) {
			delete(cs.mapUserInfo,uId)
			delete(cs.mapLogining,uId)
		}
	}

	//重新同步
	nodeId := int(playerList.NodeId)
	for _,uId := range playerList.UList{
		cs.mapUserInfo[uId] = UserInfo{Status:rpc.LoginStatus_Logined,Token:uuid.Rand().HexEx(),PlayerServiceNodeId:nodeId}
	}

	return nil
}

// 查询UserId对应的PlayeServiceNodeId
func (cs *CenterService) RPC_QueryUserNodeID(query *rpc.QueryUserNodeID, result *rpc.QueryUserNodeIDRet) error {
	userInfo, ok := cs.mapUserInfo[query.UserID]
	if ok == false {
		result.NodeID = 0
		return nil
	}

	result.NodeID = int32(userInfo.PlayerServiceNodeId)
	return nil
}
