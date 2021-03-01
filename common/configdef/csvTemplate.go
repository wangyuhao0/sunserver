package configdef

import (
	"sunserver/common/util"
)

type TemplateStructKey struct {
	A int
	B string
}

type TemplateCSVCfg struct {
	// index 0
	IndexInt int "index"
	// index 1
	IndexStr string "index"
	// index 2
	IndexStruct TemplateStructKey "index"
	Number   int32
	Str      string
	Arr1     [2]int
	Arr2     [3][2]int
	Arr3     []int
	St       []struct {
		Name string "name"
		Num  int    "num"
		Test int    "test"
	}
	M map[string]int
}

type TemplateCfg struct {
	BaseCSVCfg

	//自定义数据
	complete bool
}

func (tc *TemplateCfg) NewRow() interface{}{
	return TemplateCSVCfg{}
}

func (tc *TemplateCfg) OnLoadCSVFinish(recordFile *util.RecordFile) error{
	tc.RecordFile = recordFile
	tc.complete = true
	return nil
}

//GetItemNumber 自定义方法，获取number
func (tc *TemplateCfg) GetItemNumber(index string) int32 {
	return tc.Indexes(1)[index].(*TemplateCSVCfg).Number
}

//IsComplete 获取是否加载完成
func (tc *TemplateCfg) IsComplete() bool {
	return tc.complete
}
