package util

import (
	"github.com/duanhf2012/origin/node"
	"time"
)

const (
	MaxNodeId = 1<<10 - 1 //Uint10
	MaxSeed   = 1<<22 - 1 //MaxUint24
)

type UnionId struct {
	seed uint32
}

//非协程安全
func (uId *UnionId) GenUnionId() uint64{
	if node.GetNodeId()>MaxNodeId{
		panic("nodeId exceeds the maximum!")
	}

	seed := (uId.seed+1)%MaxSeed
	nowTime := uint64(time.Now().Second())
	return (uint64(node.GetNodeId())<<54)|(nowTime<<22)|uint64(seed)
}
