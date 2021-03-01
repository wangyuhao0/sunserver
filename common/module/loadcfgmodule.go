package module

import (
	"errors"
	"github.com/duanhf2012/origin/event"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/node"
	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/util/coroutine"
	"sunserver/common/configdef"
	"sunserver/common/global"
	"sunserver/common/util"
)

//event数据
type ReadLogicCfgData struct {
	FileName 	string
	Record  	interface{}
}

//读取模块
type LoadCfgModule struct {
	service.Module

	dir 		string
	configMap	map[string]interface{}
}

func NewLoadCfgModule() *LoadCfgModule {
	return &LoadCfgModule{}
}

func (slf *LoadCfgModule) OnInit() error{
	slf.dir = node.GetConfigDir() + "/logicConfig"
	slf.configMap = map[string]interface{}{}
	return nil
}

func (slf *LoadCfgModule) LoadCfg(fileName string) error {
	fileObj, err := slf.loadCfgInfo(fileName)
	if err != nil {
		log.Error("LoadCfgModule.LoadCfg[%s], err:%+v", fileName, err)
		return err
	}

	slf.configMap[fileName] = fileObj
	log.Release("LoadCfgModule LoadCfg[%s] suc.", fileName)
	return nil
}

func (slf *LoadCfgModule) AsyncReLoadCfg() {
	coroutine.Go(slf.coroutinesLoadCfg)
}

func (slf *LoadCfgModule) SetLogicConfig(fileName string, cfgRecord interface{}) {
	slf.configMap[fileName] = cfgRecord
}

func (slf *LoadCfgModule) GetConfig(name string) interface{} {
	ret, ok := slf.configMap[name]
	if ok == false {
		return nil
	}

	return ret
}

func (slf *LoadCfgModule) loadCfgInfo(fileName string) (interface{}, error) {
	fileDir := slf.dir + "/" + fileName
	fileObj := configdef.NewICSVFile(fileName)
	if fileObj == nil {
		err := errors.New("no this csv struct")
		return nil, err
	}

	record, readErr := util.NewRecordFile(fileObj.NewRow())
	if readErr != nil {
		return nil, readErr
	}

	readErr = record.Read(fileDir)
	if readErr != nil {
		return nil, readErr
	}

	readErr = fileObj.OnLoadCSVFinish(record)
	if readErr != nil {
		return nil, readErr
	}

	fileObj.ResetLoadTime()
	return fileObj, nil
}

func (slf *LoadCfgModule) coroutinesLoadCfg() {
	log.Release("coroutinesLoadCfg start...")
	eventDataList := make([]ReadLogicCfgData, 0, len(slf.configMap))
	for fileName := range slf.configMap {
		fileObj, err := slf.loadCfgInfo(fileName)
		if err != nil {
			log.Error("LoadCfgModule.LoadCfg[%s], err:%+v", err)
			return
		}

		eventDataList = append(eventDataList, ReadLogicCfgData{FileName: fileName, Record: fileObj})
		log.Release("LoadCfgModule coroutinesLoadCfg[%s] suc.", fileName)
	}

	log.Release("coroutinesLoadCfg finish...")
	slf.GetEventHandler().NotifyEvent(&event.Event{Type:global.EventConfigComplete, Data: eventDataList})
}
