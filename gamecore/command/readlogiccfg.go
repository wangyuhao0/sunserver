package command

import (
	"github.com/duanhf2012/origin/node"
	"github.com/duanhf2012/origin/service"
)

type ReadLogicCfgModule struct {
	service.Module
	dir string
}

func NewReadLogicCfgModule() *ReadLogicCfgModule {
	return &ReadLogicCfgModule{}
}

func (slf *ReadLogicCfgModule) OnInit() error{
	slf.dir = node.GetConfigDir() + "/logicConfig"
	return nil
}

//func (slf *ReadLogicCfgModule) ReadAll
