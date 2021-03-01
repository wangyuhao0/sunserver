package logic

import (
	"github.com/duanhf2012/origin/rpc"
	"github.com/duanhf2012/origin/service"
)

type IHelper interface {
	ExecCmd(cmd string,args... string) error
	HashCmd(command string) bool
	rpc.IRpcHandler
	service.IModule
}

