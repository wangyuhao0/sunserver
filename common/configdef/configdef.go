package configdef

import (
	"sunserver/common/util"
	"time"
)

const (
	FileTemplate string = "template.csv"
)

type IBaseFile interface {
	ResetLoadTime()
	GetLoadTime() int64
}

type ICSVFile interface {
	IBaseFile

	NewRow() interface{}
	OnLoadCSVFinish(recordFile *util.RecordFile) error
}

func NewICSVFile(fileName string) ICSVFile {
	switch fileName {
	case FileTemplate:
		return &TemplateCfg{}
	default:
		return nil
	}
}

type BaseCSVCfg struct {
	*util.RecordFile
	loadCompleteTime int64
}

func (slf *BaseCSVCfg) GetItemByFirstIndex(index interface{}) interface{} {
	return slf.RecordFile.Index(index)
}

func (slf *BaseCSVCfg) GetItemByChooseIndex(chooseIndex int, index interface{}) interface{} {
	indexMap := slf.RecordFile.Indexes(chooseIndex)
	if indexMap == nil {
		return nil
	}

	return indexMap[index]
}

func (slf *BaseCSVCfg) GetItemList(chooseIndex int) map[interface{}]interface{} {
	indexMap := slf.RecordFile.Indexes(chooseIndex)
	if indexMap == nil {
		return nil
	}

	return indexMap
}

func (slf *BaseCSVCfg) ResetLoadTime() {
	slf.loadCompleteTime = time.Now().Unix()
}

func (slf *BaseCSVCfg) GetLoadTime() int64  {
	return slf.loadCompleteTime
}
