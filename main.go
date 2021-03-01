package main

import (
	"github.com/duanhf2012/origin/node"
	_ "sunserver/gamecore/command"
	_ "sunserver/gamecore/dbservice"
	_ "sunserver/gamecore/gateservice"
	_ "sunserver/gamecore/msgcenterservice"
	_ "sunserver/gamecore/playerservice"
	_ "sunserver/gamecore/queueservice"
	_ "sunserver/gamecore/roomservice"
	_ "sunserver/gamecore/tableservice"
	_ "sunserver/gamemaster/authservice"
	_ "sunserver/gamemaster/centerservice"
	_ "sunserver/gamemaster/httpgateservice"
	_ "sunserver/originhelper"
	"time"
)



func main() {
	//打开性能分析报告功能，并设置10秒汇报一次
	node.OpenProfilerReport(time.Second * 10)
	node.Start()


}






