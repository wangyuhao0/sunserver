package logic

import (
	"fmt"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/util/uuid"
	"strconv"
	"sunserver/common/proto/rpc"
	"sunserver/common/util"
)

func TestCallRPC(helper IHelper, args ...string) error {
	if len(args) != 1 {
		return fmt.Errorf("Must be 1 parameters.")
	}

	end, err := strconv.Atoi(args[0])
	if err != nil || end<=0 {
		return fmt.Errorf("The second parameter[%d] is error:%+v.", end, err)
	}

	for i := 0; i <= end; i++ {
		go func() {
			argInfo := rpc.LoginInfo{
				PlatType:             rpc.LoginType(util.RandNum(10000)),
				PlatId:               "",
				AccessToken:          uuid.Rand().HexEx(),
			}
			retInfo := rpc.LoginResult{}
			helper.Call("AuthService.c", &argInfo, &retInfo)
			if argInfo.PlatType != retInfo.PlatType || argInfo.AccessToken != retInfo.AccessToken || argInfo.PlatId != retInfo.PlatId {
				log.Error("TestCallRPC err: arg[%+v] != ret[%+v]", &argInfo, &retInfo)
			} else {
				log.Release("TestCallRPC arg[%+v] == ret[%+v]", &argInfo, &retInfo)
			}
		}()
	}

	return nil
}