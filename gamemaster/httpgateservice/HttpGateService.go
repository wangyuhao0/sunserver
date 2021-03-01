package httpgateservice

import (
	"github.com/duanhf2012/origin/node"
	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/sysservice/httpservice"
	"github.com/duanhf2012/origin/util/timer"
	"sunserver/common/proto/rpc"
	"time"
)

func init() {
	node.Setup(&httpservice.HttpService{})
	node.Setup(&HttpGateService{})
}

type HttpGateService struct {
	service.Service
	loginModule *LoginModule
	mapTcpGate map[int]*TcpGateInfo

	tcpGateInfo []GateInfoResp
	dirty bool
}

type TcpGateInfo struct {
	Weight int32
	Url string
	refresh time.Time
}

func (gate *HttpGateService) OnInit() error {
	gate.mapTcpGate = map[int]*TcpGateInfo{}

	tcpGateList := gate.GetServiceCfg().([]interface{})
	for _,g := range tcpGateList{
		mapGate := g.(map[string]interface{})
		nodeId := int(mapGate["NodeId"].(float64))
		addr := mapGate["Addr"].(string)
		gate.mapTcpGate[nodeId] = &TcpGateInfo{Weight:-1,Url:addr}
	}

	//获取系统httpService服务
	httpService := node.GetService("HttpService").(*httpservice.HttpService)

	//新建并设置路由对象
	httpRouter := httpservice.NewHttpHttpRouter()
	httpService.SetHttpRouter(httpRouter, gate.GetEventHandler())

	gate.loginModule = &LoginModule{}
	gate.loginModule.funcGetGateUrl = gate.GetGateInfoUrl
	gate.AddModule(gate.loginModule)

	//性能监控
	gate.OpenProfiler()
	gate.GetProfiler().SetOverTime(time.Millisecond * 100)
	gate.GetProfiler().SetMaxOverTime(time.Second * 10)

	gate.NewTicker(time.Second*5,gate.PrepareGateService)

	//POST方法 请求url:http://127.0.0.1:9402/login
	//返回结果为：{"msg":"hello world"}
	httpRouter.POST("/login", gate.loginModule.Login)
	return nil
}

// GateService->HttpGateService同步负载
func (gate *HttpGateService) RPC_SetTcpGateBalance(balance *rpc.GateBalance) error{
	v,ok := gate.mapTcpGate[int(balance.NodeId)]
	if ok == false{
		return nil
	}
	if v.Weight != balance.Weigh {
		gate.dirty = true
	}

	v.Weight = balance.Weigh
	v.refresh = time.Now()

	return nil
}

// 定时预处理网关列表
func (gate *HttpGateService) PrepareGateService(timer *timer.Ticker){
	if gate.dirty == false {
		return
	}
	gate.tcpGateInfo = make([]GateInfoResp,0,len(gate.mapTcpGate))
	for _,info := range gate.mapTcpGate {
		if time.Now().Sub(info.refresh) > 10*time.Second { //10秒都没有同步
			continue
		}

		gate.tcpGateInfo = append(gate.tcpGateInfo,GateInfoResp{Weight:info.Weight,Url: info.Url})
	}
	gate.dirty = false
}

func (gate *HttpGateService) GetGateInfoUrl() []GateInfoResp{
	return gate.tcpGateInfo
}
