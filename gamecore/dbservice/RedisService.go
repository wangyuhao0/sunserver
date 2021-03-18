package dbservice

import (
	"fmt"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/node"
	"github.com/duanhf2012/origin/rpc"
	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/sysmodule/redismodule"
	"gopkg.in/mgo.v2/bson"
	"runtime"
	"sunserver/common/db"
)

func init() {
	node.Setup(&RedisService{})
}

var redisEmptyRes []byte

type RedisService struct {
	service.Service
	redisModule    redismodule.RedisModule
	channelOptData []chan RedisRequest
	ip             string
	port           int
	password       string
	dbIndex        int
	maxIdle        int //最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态。
	maxActive      int //最大的激活连接数，表示同时最多有N个连接
	idleTimeout    int //最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭
	goroutineNum   uint32
	channelNum     int
}

func (redisService *RedisService) ReadCfg() error {
	mapRedisServiceCfg, ok := redisService.GetServiceCfg().(map[string]interface{})
	if ok == false {
		return fmt.Errorf("RedisService config is error!")
	}

	//parse MsgRouter
	url, ok := mapRedisServiceCfg["Url"]
	if ok == false {
		return fmt.Errorf("RedisService config is error!")
	}
	redisService.ip = url.(string)

	port, ok := mapRedisServiceCfg["Port"]
	if ok == false {
		return fmt.Errorf("RedisService config is error!")
	}
	redisService.port = int(port.(float64))

	password, ok := mapRedisServiceCfg["Password"]
	if ok == false {
		return fmt.Errorf("RedisService config is error!")
	}
	redisService.password = password.(string)

	dbIndex, ok := mapRedisServiceCfg["DbIndex"]
	if ok == false {
		return fmt.Errorf("RedisService config is error!")
	}
	redisService.dbIndex = int(dbIndex.(float64))

	maxIdle, ok := mapRedisServiceCfg["MaxIdle"]
	if ok == false {
		return fmt.Errorf("RedisService config is error!")
	}
	redisService.maxIdle = int(maxIdle.(float64))

	maxActive, ok := mapRedisServiceCfg["MaxActive"]
	if ok == false {
		return fmt.Errorf("RedisService config is error!")
	}
	redisService.maxActive = int(maxActive.(float64))

	idleTimeout, ok := mapRedisServiceCfg["IdleTimeout"]
	if ok == false {
		return fmt.Errorf("RedisService config is error!")
	}
	redisService.idleTimeout = int(idleTimeout.(float64))

	goroutineNum, ok := mapRedisServiceCfg["GoroutineNum"]
	if ok == false {
		return fmt.Errorf("RedisService config is error!")
	}
	redisService.goroutineNum = uint32(goroutineNum.(float64))

	channelNum, ok := mapRedisServiceCfg["ChannelNum"]
	if ok == false {
		return fmt.Errorf("RedisService config is error!")
	}
	redisService.channelNum = int(channelNum.(float64))

	return nil
}

func (redisService *RedisService) OnInit() error {
	fmt.Println("start init RedisService")
	defer fmt.Println("finish init RedisService")

	err := redisService.ReadCfg()
	if err != nil {
		return err
	}
	var redisCfg redismodule.ConfigRedis
	redisCfg.DbIndex = redisService.dbIndex         //数据库索引
	redisCfg.IdleTimeout = redisService.idleTimeout //最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭
	redisCfg.MaxIdle = redisService.maxIdle         //最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态。
	redisCfg.MaxActive = redisService.maxActive     //最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭
	redisCfg.IP = redisService.ip                   //redis服务器IP
	redisCfg.Port = redisService.port               //redis服务器端口
	redisCfg.Password = redisService.password
	redisService.redisModule.Init(&redisCfg)

	redisService.channelOptData = make([]chan RedisRequest, redisService.goroutineNum)
	for i := uint32(0); i < redisService.goroutineNum; i++ {
		redisService.channelOptData[i] = make(chan RedisRequest, redisService.channelNum)
		go redisService.ExecuteOptData(redisService.channelOptData[i])
	}

	return nil
}

func (redisService *RedisService) ExecuteOptData(channelOptData chan RedisRequest) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			err := fmt.Errorf("%v: %s", r, buf[:l])
			log.Error("core dump info:%+v\n", err)
			redisService.ExecuteOptData(channelOptData)
		}
	}()
	for {
		select {
		case optData := <-channelOptData:
			switch optData.request.GetType() {
			case db.OptType_Find:
				redisService.DoFind(optData)
			case db.OptType_Insert:
				redisService.DoInsert(optData)
			case db.OptType_Del:
				redisService.DoDel(optData)
			case db.OptType_InsertNoFallBack:
				redisService.DoInsertNoFallBack(optData)

			default:
				log.Error("optype %d is error.", optData.request.GetType())
			}
		}
	}
}

func (redisService *RedisService) DoDel(dbReq RedisRequest) {
	redisModule := redisService.redisModule
	//redis key
	rKey := dbReq.request.RKey

	err := redisModule.DelString(rKey)

	if err != nil {
		redisService.responseRet(dbReq, err, 0)
		return
	}
	redisService.responseRet(dbReq, nil, 1)

}

func (redisService *RedisService) responseRet(dbReq RedisRequest, err error, effectRow int32) {
	var dbRet db.RedisControllerRet
	if effectRow > 0 {
		dbRet.RowNum = effectRow
	}

	if dbReq.responder.IsInvalid() == false {
		if err == nil {
			dbReq.responder(&dbRet, rpc.NilError)
		} else {
			dbReq.responder(&dbRet, rpc.RpcError(err.Error()))
		}

	}
}

func (redisService *RedisService) DoFind(dbReq RedisRequest) {
	//1.选择数据库与表

	redisModule := redisService.redisModule
	//redis key
	rKey := dbReq.request.RKey
	var dbRet db.RedisControllerRet
	var rpcErr rpc.RpcError
	flag, err := redisModule.ExistsKey(rKey)
	if !flag || err != nil {
		//说明不存在
		dbRet.Res = redisEmptyRes
		dbReq.responder(&dbRet, rpc.RpcError(err.Error()))
		dbRet.RowNum = 0
		return
	}

	//从key里面拿东西
	result, err1 := redisModule.GetString(rKey)

	if err1 != nil {
		dbRet.Res = redisEmptyRes
		dbReq.responder(&dbRet, rpc.RpcError(err1.Error()))
		dbRet.RowNum = 0
		return
	}

	out, err2 := bson.Marshal(result)
	if err2 != nil {
		dbRet.Res = redisEmptyRes
		dbReq.responder(&dbRet, rpc.RpcError(err2.Error()))
		dbRet.RowNum = 0
		return
	}

	dbRet.Res = out
	dbRet.RowNum = 1
	dbReq.responder(&dbRet, rpcErr)

}

func (redisService *RedisService) DoInsert(dbReq RedisRequest) {
	//1.选择数据库与表
	rKey := dbReq.request.RKey
	rValue := dbReq.request.RValue

	redisModule := redisService.redisModule
	err := redisModule.SetString(rKey, rValue)

	if err != nil {
		redisService.responseRet(dbReq, err, 0)
		return
	}
	redisService.responseRet(dbReq, nil, 1)

}

func (redisService *RedisService) DoInsertNoFallBack(dbReq RedisRequest) {
	//1.选择数据库与表
	rKey := dbReq.request.RKey
	rValue := dbReq.request.RValue

	redisModule := redisService.redisModule
	err := redisModule.SetString(rKey, rValue)

	if err != nil {
		//redisService.responseRet(dbReq, err, 0)
		log.Error("DoInsertNoFallBack，%s", rKey)
		return
	}
	//redisService.responseRet(dbReq, nil, 1)

}

type RedisRequest struct {
	request   *db.RedisControllerReq
	responder rpc.Responder
}

func (redisService *RedisService) RPC_RedisRequest(responder rpc.Responder, request *db.RedisControllerReq) error {
	//从 LoginModule rpc发往db 进行数据处理
	index := request.GetKey() % uint64(redisService.goroutineNum)
	if len(redisService.channelOptData[index]) == cap(redisService.channelOptData[index]) {
		log.Error("channel is full %d", index)

		responder(nil, rpc.RpcError("channel is full"))
		return nil
	}

	var redisRequest RedisRequest
	redisRequest.request = request
	redisRequest.responder = responder

	//往管道发数据
	redisService.channelOptData[index] <- redisRequest
	return nil
}

func (redisService *RedisService) RPC_InitDataRequest(request *db.RedisControllerReq) error {
	//从 LoginModule rpc发往db 进行数据处理
	index := request.GetKey() % uint64(redisService.goroutineNum)
	if len(redisService.channelOptData[index]) == cap(redisService.channelOptData[index]) {
		log.Error("channel is full %d", index)

		return nil
	}

	var redisRequest RedisRequest
	redisRequest.request = request
	//往管道发数据
	redisService.channelOptData[index] <- redisRequest
	return nil
}
