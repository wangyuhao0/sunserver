package room

import (
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/util/timer"
)

//用于存于不需要持久化的数据
type DataInfo struct {

	//pingTime time.Time   //ping时间
	//saveTime time.Time   //上次存档时间

	pintTicker *timer.Ticker
}

//清理内存变量
func (dataInfo *DataInfo) Clear() {

	if dataInfo.pintTicker != nil {
		log.Warning("ticker not close?")
		dataInfo.pintTicker.Cancel()
	}

	dataInfo.pintTicker = nil
}

//对象释放时
func (dataInfo *DataInfo) OnRelease() {
	if dataInfo.pintTicker != nil {
		dataInfo.pintTicker.Cancel()
		dataInfo.pintTicker = nil
	}
}
