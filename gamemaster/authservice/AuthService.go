package authservice

import (
	"fmt"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/node"
	"github.com/duanhf2012/origin/service"
	"sunserver/common/proto/rpc"
	"time"
)

func init(){
	node.Setup(&AuthService{})
}

type AuthService struct {
	service.Service
}

func (auth *AuthService) OnInit() error {
	cfg := auth.GetServiceCfg().(map[string]interface{})
	v,ok := cfg["GoRoutineNum"]
	if ok == false {
		return fmt.Errorf("Cannot find authService.goRoutineNum config!")
	}
	auth.SetGoRoutineNum(int32(v.(float64)))

	//性能监控
	auth.OpenProfiler()
	auth.GetProfiler().SetOverTime(time.Second * 2)
	auth.GetProfiler().SetMaxOverTime(time.Second * 10)

	return nil
}

func (auth *AuthService) RPC_Check(loginInfo *rpc.LoginInfo,loginResult *rpc.LoginResult) error{
	//loginResult.Ret = 0
	log.Release("进入到了AuthService-RPC_Check")
	/*var req db.RedisControllerReq
	req.Type = db.OptType_Find
	req.RKey = loginInfo.AccessToken
	req.Key = uint64(util.HashString2Number(loginInfo.PlatId))

	err := auth.GetService().GetRpcHandler().AsyncCall("RedisService.RPC_RedisRequest",&req,func(res *db.RedisControllerRet,err error){
		//返回账号创建结果
		if err != nil || res.RowNum ==0 {
			//说明不存在
			loginResult.Ret = 0
			loginResult.AccessToken = loginInfo.AccessToken
			loginResult.PlatId = loginInfo.PlatId
			loginResult.PlatType = loginInfo.PlatType
			return
		}
	})

	if err!=nil {
		return err
	}*/

	return nil
}

func (auth *AuthService) RPC_TestCall(argInfo *rpc.LoginInfo, retInfo *rpc.LoginResult) error {
	retInfo.PlatId = argInfo.PlatId
	retInfo.AccessToken = argInfo.AccessToken
	retInfo.PlatType = argInfo.PlatType

	return nil
}
