package httpgateservice

import (
	"encoding/json"
	"github.com/duanhf2012/origin/cluster"
	"github.com/duanhf2012/origin/log"
	originrpc "github.com/duanhf2012/origin/rpc"
	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/sysservice/httpservice"
	"github.com/duanhf2012/origin/util/timer"
	"net/http"
	"strconv"
	"sunserver/common/collect"
	"sunserver/common/constPackage"
	"sunserver/common/db"
	"sunserver/common/proto/msg"
	"sunserver/common/proto/rpc"
	"sunserver/common/util"
)

type GateInfoResp struct {
	Weight int32
	Url    string
}

type LoginModule struct {
	service.Module
	seed           int
	funcGetGateUrl func() []GateInfoResp

	unionId util.UnionId
}

func (login *LoginModule) OnInit() error {
	return nil
}

func (login *LoginModule) OnRelease() {
}

type HttpRespone struct {
	ECode         int
	UserId        uint64
	GateServerUrl []GateInfoResp
	Token         string
}

func (login *LoginModule) loginCheck(session *httpservice.HttpSession, loginInfo rpc.LoginInfo) {
	//1.验证平台类型和Id
	log.Release("进入到了loginCheck")
	platId := loginInfo.PlatId
	if loginInfo.PlatType < 0 || loginInfo.PlatType >= rpc.LoginType_LoginType_Max {
		log.Warning("plat type %d is error!", loginInfo.PlatType)
		login.WriteResponseError(session, msg.ErrCode_PlatTypeError)
		return
	}

	if len(platId) == 0 {
		log.Warning("plat type %d is error!", loginInfo.PlatType)
		login.WriteResponseError(session, msg.ErrCode_PlatIdError)
		return
	}

	//2.向验证服检查登陆
	err := login.GetService().GetRpcHandler().AsyncCall("AuthService.RPC_Check", &loginInfo, func(loginResult *rpc.LoginResult, err error) {
		if err != nil {
			log.Error("call AuthService.RPC_Check fail %s,platid:%s!", err.Error(), platId)
			login.WriteResponseError(session, msg.ErrCode_InterNalError)
			return
		}

		if loginResult.Ret != 0 {
			log.Warning("AuthService.RPC_Check fail Ret:%d,platid:%s.", loginResult.Ret, platId)
			login.WriteResponseError(session, msg.ErrCode_TokenError)
		}

		//验证通过从数据库生成或获取账号信息
		login.loginToDB(session, loginInfo)
	})

	//3.服务内部错误
	if err != nil {
		login.WriteResponseError(session, msg.ErrCode_InterNalError)
		log.Error("AsyncCall AuthService.RPC_Check fail %s,platid:%s!", err.Error(), platId)
	}
}

func (login *LoginModule) GetBestNodeId(serviceMethod string) int {
	var clientList [4]*originrpc.Client
	err, num := cluster.GetCluster().GetNodeIdByService("CenterService", clientList[:], false)
	if err != nil || num == 0 {
		return 0
	}

	for i := 0; i < num; i++ {
		if clientList[i] != nil {
			return clientList[i].GetId()
		}
	}

	return 0
}

func (login *LoginModule) choseServer(session *httpservice.HttpSession, user *collect.User, loginInfo rpc.LoginInfo) {
	//1.查找最优的CenterService
	log.Release("进入到LoginModule-choseServer")
	//获取到中心服
	bestNodeId := util.GetMasterCenterNodeId() // login.GetBestNodeId("CenterService.RPC_ChoseServer")
	if bestNodeId == 0 {
		login.WriteResponseError(session, msg.ErrCode_InterNalError)
		log.Error("Cannot find CenterService.RPC_ChoseServer best node id!")
		return
	}
	redisNodeId := util.GetNodeIdByService("RedisService")
	//2.登陆到中心服
	var req rpc.ChoseServerReq
	req.UserId = user.Id
	// 从中心服进行数据
	err := login.GetService().GetRpcHandler().AsyncCallNode(bestNodeId, "CenterService.RPC_ChoseServer", &req, func(res *rpc.ChoseServerRet, err error) {
		if err != nil {
			login.WriteResponseError(session, msg.ErrCode_InterNalError)
			log.Error("chose server fail %s!", err.Error())
			return
		}

		if res.Ret != 0 {
			login.WriteResponseError(session, msg.ErrCode_InterNalError)
			log.Error("chose server fail %d!", res.Ret)
			return
		}

		//登陆成功,返回结果
		var resp HttpRespone
		resp.Token = res.Token
		resp.UserId = user.Id
		resp.GateServerUrl = login.funcGetGateUrl()
		session.WriteJsonDone(http.StatusOK, &resp)
		//存入 redis 数据缓存
		//异步存储
		var req db.RedisControllerReq
		req.Type = db.OptType_InsertNoFallBack
		req.RKey = constpackage.UserRedisKey + res.Token
		req.Key = uint64(util.HashString2Number(loginInfo.PlatId))
		data, err := json.Marshal(user)
		req.RValue = string(data)
		err = login.GoNode(redisNodeId, "RedisService.RPC_InitDataRequest", &req)
		if err != nil {
			log.Error("set Redis fail %d!", res.Ret)
			//login.WriteResponseError(session,msg.ErrCode_InterNalError)
			return
		}

	})

	if err != nil {
		login.WriteResponseError(session, msg.ErrCode_InterNalError)
		return
	}
}

func (login *LoginModule) loginToDB(session *httpservice.HttpSession, loginInfo rpc.LoginInfo) {
	//1.生成数据库请求
	log.Release("LoginModule-loginTOdb")
	//先去mysql进行核验 然后mongodb留存
	platId := loginInfo.PlatId
	account := loginInfo.Account
	passWord := loginInfo.PassWord
	var mysqlData db.MysqlControllerReq
	mysqlData.TableName = "user"
	mysqlData.Key = uint64(util.HashString2Number(loginInfo.PlatId))
	mysqlData.Sql = constpackage.LoginSql
	mysqlData.Args = []string{account, passWord}
	mysqlData.Type = db.OptType_Find
	err := login.GetService().GetRpcHandler().AsyncCall("MysqlService.RPC_MysqlDBRequest", &mysqlData, func(ret *db.MysqlControllerRet, err error) {
		//返回账号创建结果
		var user collect.User
		if err != nil {
			login.WriteResponseError(session, msg.ErrCode_InterNalError)
			if err != nil {
				log.Error("Call MysqlService.RPC_MysqlDBRequest platid:%s, fail %s!", platId, err.Error())
			} else {
				log.Error("Call MysqlService.RPC_MysqlDBRequest platid:%s, fail res is empty!", platId)
			}
			return
		}
		if ret.RowNum == 0 {
			//代表不存在 初始化
			log.Release("初始化用户------")
			var mysqlData db.MysqlControllerReq
			mysqlData.TableName = "user"
			mysqlData.Key = uint64(util.HashString2Number(loginInfo.PlatId))
			mysqlData.Sql = "insert `user`(nick_name,account,`password`,create_time) values(?,?,?,?)"
			mysqlData.Args = []string{account, account, passWord, strconv.FormatInt(timer.Now().Unix(), 10)}
			mysqlData.Type = db.OptType_Insert
			err := login.GetService().GetRpcHandler().AsyncCall("MysqlService.RPC_MysqlDBRequest", &mysqlData, func(ret *db.MysqlControllerRet, err error) {
				//返回账号创建结果
				if err != nil {
					login.WriteResponseError(session, msg.ErrCode_InterNalError)
					if err != nil {
						log.Error("Call MysqlService.RPC_MysqlDBRequest platid:%s, fail %s!", platId, err.Error())
					} else {
						log.Error("Call MysqlService.RPC_MysqlDBRequest platid:%s, fail res is empty!", platId)
					}
					return
				}
				err = json.Unmarshal(ret.Res[0], &user)
				if err != nil {
					login.WriteResponseError(session, msg.ErrCode_InterNalError)
					log.Error("Unmarshal fail %s,platid:%s!", err.Error(), platId)
					return
				}
			})
			if err != nil {
				login.WriteResponseError(session, msg.ErrCode_InterNalError)
				log.Error("Call MysqlService.RPC_MysqlDBRequest platid:%s, fail %s!", err.Error(), platId)
				return
			}

		} else {
			log.Release("更新用户------")
			//先处理原数据
			err = json.Unmarshal(ret.Res[0], &user)
			if err != nil {
				login.WriteResponseError(session, msg.ErrCode_InterNalError)
				log.Error("Unmarshal fail %s,platid:%s!", err.Error(), platId)
				return
			}
		}
		//向centerService登陆
		login.choseServer(session, &user, loginInfo)

	})
	if err != nil {
		login.WriteResponseError(session, msg.ErrCode_InterNalError)
		log.Error("Call MysqlService.RPC_MysqlDBRequest platid:%s, fail %s!", err.Error(), platId)
		return
	}
	/*var req db.DBControllerReq
	req.CollectName = collect.AccountCollectName
	req.Type = db.OptType_SetOnInsert+db.OptType_Find
	req.Condition, _ = bson.Marshal(bson.D{{"PlatId", platId}})
	req.Key = uint64(util.HashString2Number(loginInfo.PlatId))

	var cAccount collect.CAccount
	cAccount.PlatId = loginInfo.PlatId
	cAccount.PlatType = int(loginInfo.PlatType)
	cAccount.UserId = login.unionId.GenUnionId() //如果账号不存在，即使用当前生成的唯一Id
	out, err := bson.Marshal(cAccount)
	if err != nil {
		login.WriteResponseError(session,msg.ErrCode_InterNalError)
		log.Error("LoginToDB fail:%s,platId:%s!",err.Error(),platId)
		return
	}
	req.Data = append(req.Data, out)

	//2.平台登陆验证成功，去DB创建或者查询账号
	err = login.GetService().GetRpcHandler().AsyncCall("MongoService.RPC_MongoDBRequest",&req,func(res *db.DBControllerRet,err error){
		//返回账号创建结果
		if err != nil || len(res.Res) ==0 {
			login.WriteResponseError(session,msg.ErrCode_InterNalError)
			if err != nil {
				log.Error("Call MongoService.RPC_DBRequest platid:%s, fail %s!",platId,err.Error())
			}else {
				log.Error("Call MongoService.RPC_DBRequest platid:%s, fail res is empty!",platId)
			}
			return
		}

		//解析数据
		var account collect.CAccount
		err = bson.Unmarshal(res.Res[0],&account)
		if err != nil {
			login.WriteResponseError(session,msg.ErrCode_InterNalError)
			log.Error("Unmarshal fail %s,platid:%s!",err.Error(),platId)
			return
		}

		//向centerService登陆
		login.choseServer(session,account.UserId)
	})

	if err != nil {
		login.WriteResponseError(session,msg.ErrCode_InterNalError)
		log.Error("AsyncCall MongoService.RPC_MongoDBRequest fail %s,platid:%s!",err.Error(),platId)
	}*/
}

func (login *LoginModule) WriteResponseError(session *httpservice.HttpSession, eCode msg.ErrCode) {
	var resp HttpRespone
	resp.ECode = int(eCode)

	session.WriteJsonDone(http.StatusOK, &resp)
}

/*
Http登陆会返回以下错误：
OK            = 0 验证通过
InterNalError = 1 服务器内部错误
TokenError    = 2 Token验证不通过
PlatTypeError = 5 平台类型错误
PlatIdError   = 6 平台id错误

//请求
{
    "PlatType":0,
    "PlatId":"0_xxxxxxx",
    "AccessToken":"token"
}

//响应
{
    "ECode": 0,
    "UserId": 18014398547230721,
    "GateServerUrl": [
        {
            "Weight": 0,
            "Url": "127.0.0.1:9401"
        }
    ],
    "Token": "8c63d73978a645a2b4d1ae77254e2a66"
}
*/
func (login *LoginModule) Login(session *httpservice.HttpSession) {
	//1.验证Body请求内容
	var loginInfo rpc.LoginInfo
	err := json.Unmarshal(session.GetBody(), &loginInfo)
	if err != nil || loginInfo.AccessToken == "" {
		login.WriteResponseError(session, msg.ErrCode_TokenError)
		log.Warning("The body content of the HTTP request is incorrect:%s!", string(session.GetBody()))
		return
	}

	//2.平台登陆验证
	login.loginCheck(session, loginInfo)
}
