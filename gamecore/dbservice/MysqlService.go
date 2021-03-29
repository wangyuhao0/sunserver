package dbservice

import (
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/node"
	"github.com/duanhf2012/origin/rpc"
	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/sysmodule/mysqlmondule"
	"runtime"
	"sunserver/common/collect"
	"sunserver/common/const"
	"sunserver/common/db"
	"sync/atomic"
	"time"
)

func init() {
	node.Setup(&MysqlService{})
}

const slowTime = int64(500) //毫秒

type MysqlService struct {
	service.Service
	mysqlModule    mysqlmondule.MySQLModule
	channelOptData []chan MysqlDBRequest
	url            string
	userName       string
	passWord       string
	dbName         string
	goroutineNum   uint32
	maxConn        int
	channelNum     int

	dbDealCount   int32
	dbAllCostTime int64
	dbMaxCostTime int64
}

func (mysqlService *MysqlService) ReadCfg() error {
	mapMysqlServiceCfg, ok := mysqlService.GetServiceCfg().(map[string]interface{})
	if ok == false {
		return fmt.Errorf("MysqlService config is error!")
	}

	//parse MsgRouter
	url, ok := mapMysqlServiceCfg["Url"]
	if ok == false {
		return fmt.Errorf("MysqlService config is error!")
	}
	mysqlService.url = url.(string)

	userName, ok := mapMysqlServiceCfg["UserName"]
	if ok == false {
		return fmt.Errorf("MysqlService config is error!")
	}
	mysqlService.userName = userName.(string)

	passWord, ok := mapMysqlServiceCfg["PassWord"]
	if ok == false {
		return fmt.Errorf("MysqlService config is error!")
	}
	mysqlService.passWord = passWord.(string)

	dbName, ok := mapMysqlServiceCfg["DBName"]
	if ok == false {
		return fmt.Errorf("MysqlService config is error!")
	}
	mysqlService.dbName = dbName.(string)

	goroutineNum, ok := mapMysqlServiceCfg["GoroutineNum"]
	if ok == false {
		return fmt.Errorf("MysqlService config is error!")
	}
	mysqlService.goroutineNum = uint32(goroutineNum.(float64))

	maxConn, ok := mapMysqlServiceCfg["MaxConn"]
	if ok == false {
		return fmt.Errorf("MysqlService config is error!")
	}
	mysqlService.maxConn = int(maxConn.(float64))

	channelNum, ok := mapMysqlServiceCfg["ChannelNum"]
	if ok == false {
		return fmt.Errorf("MongoService config is error!")
	}
	mysqlService.channelNum = int(channelNum.(float64))
	return nil
}

func (mysqlService *MysqlService) OnInit() error {
	fmt.Println("start init MysqlService")
	defer fmt.Println("finish init MysqlService")
	err := mysqlService.ReadCfg()
	if err != nil {
		return err
	}

	err = mysqlService.mysqlModule.Init(mysqlService.url, mysqlService.userName, mysqlService.passWord, mysqlService.dbName, mysqlService.maxConn)
	if err != nil {
		return err
	}

	mysqlService.channelOptData = make([]chan MysqlDBRequest, mysqlService.goroutineNum)
	for i := uint32(0); i < mysqlService.goroutineNum; i++ {
		mysqlService.channelOptData[i] = make(chan MysqlDBRequest, mysqlService.channelNum)
		go mysqlService.ExecuteOptData(mysqlService.channelOptData[i])
	}

	mysqlService.dbDealCount = 0
	mysqlService.dbAllCostTime = 0
	mysqlService.dbMaxCostTime = 0

	//MysqlService.NewTicker(time.Second*5, MysqlService.PrintDBCost)

	//性能监控
	mysqlService.OpenProfiler()
	mysqlService.GetProfiler().SetOverTime(time.Millisecond * 500)
	mysqlService.GetProfiler().SetMaxOverTime(time.Second * 10)

	return nil
}

func (mysqlService *MysqlService) ExecuteOptData(channelOptData chan MysqlDBRequest) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			err := fmt.Errorf("%v: %s", r, buf[:l])
			log.Error("core dump info:%+v\n", err)
			mysqlService.ExecuteOptData(channelOptData)
		}
	}()
	for {
		select {
		case optData := <-channelOptData:
			timeNow := time.Now()
			switch optData.request.GetType() {
			case db.OptType_Del:
				mysqlService.DoDel(optData)
			case db.OptType_Update:
				mysqlService.DoUpdate(optData)
			case db.OptType_Find:
				mysqlService.DoFind(optData)
			case db.OptType_Insert:
				mysqlService.DoInsert(optData)
			default:
				log.Error("optype %d is error.", optData.request.GetType())
			}

			costTime := time.Now().Sub(timeNow).Milliseconds()
			if atomic.LoadInt64(&mysqlService.dbMaxCostTime) < costTime {
				atomic.StoreInt64(&mysqlService.dbMaxCostTime, costTime)
			}
			atomic.AddInt64(&mysqlService.dbAllCostTime, costTime)
			atomic.AddInt32(&mysqlService.dbDealCount, 1)

			if costTime >= slowTime {
				log.Warning("MysqlService.ExecuteOptData[%+v] slow[%d]", &optData, costTime)
			}
		}
	}
}

func (mysqlService *MysqlService) DoDel(dbReq MysqlDBRequest) {
	//1.选择数据库与表
	request := dbReq.request
	sql := request.Sql
	args := request.Args
	exec, err := mysqlService.mysqlModule.Exec(sql, mysqlService.SetArgs(args)...)

	if err != nil {
		log.Error("%s DoDel fail error %s", dbReq.request.TableName, err.Error())
	}
	mysqlService.responseRet(dbReq, err, int32(exec.RowsAffected))
}

func (mysqlService *MysqlService) DoUpdate(dbReq MysqlDBRequest) {
	//1.选择数据库与表
	request := dbReq.request
	sql := request.Sql
	args := request.Args
	exec, err := mysqlService.mysqlModule.Exec(sql, mysqlService.SetArgs(args)...)
	if err != nil {
		log.Error("%s DoUpdate fail error %s", dbReq.request.TableName, err.Error())
	}

	if request.GetCallBack() {
		if dbReq.responder.IsInvalid() == false {
			mysqlService.responseRet(dbReq, err, int32(exec.RowsAffected))
		}
	}
}

func (mysqlService *MysqlService) DoFind(dbReq MysqlDBRequest) {
	//1.选择数据库与表
	request := dbReq.request
	tableName := request.TableName
	args := request.Args
	var dbRet db.MysqlControllerRet
	var rowNum int

	var rpcErr rpc.RpcError
	//序列化结果
	dbRet.Type = dbReq.request.Type
	dbRet.Res = make([][]byte, 0)
	result, err := mysqlService.mysqlModule.Query(request.Sql, mysqlService.SetArgs(args)...)
	if err != nil {
		fmt.Print(err)
		rpcErr = rpc.RpcError(err.Error())
		dbReq.responder(&dbRet, rpcErr)
		return
	} else {
		if tableName == "user" {
			var dbRetData []collect.User
			err = result.UnMarshal(&dbRetData)
			if err != nil {
				fmt.Print(err)
				rpcErr = rpc.RpcError(err.Error())
				dbReq.responder(&dbRet, rpcErr)
				return
			}
			rowNum = len(dbRetData)

			for i := 0; i < rowNum; i++ {
				bytes, err := json.Marshal(dbRetData[i])
				if err != nil {
					rpcErr = rpc.RpcError(err.Error())
					dbRet.Res = emptyRes
					break
				}
				dbRet.Res = append(dbRet.Res, bytes)
			}
		}
		//从结构集中返序列化数据到结构体切片中，UnMarshal可以支持多个结果集
	}

	dbRet.RowNum = int32(rowNum)
	dbReq.responder(&dbRet, rpcErr)
}

func (mysqlService *MysqlService) DoInsert(dbReq MysqlDBRequest) {
	//1.选择数据库与表
	request := dbReq.request
	sql := request.Sql
	args := request.Args
	execResult, err := mysqlService.mysqlModule.Exec(sql, mysqlService.SetArgs(args)...)
	log.Release("insert %s", execResult)
	tableName := request.TableName
	var rpcErr rpc.RpcError
	var dbRet db.MysqlControllerRet
	var rowNum int
	findArgs := make([]interface{}, 0)
	findArgs = append(findArgs, request.Args[1])
	findArgs = append(findArgs, request.Args[2])
	if err != nil {
		log.Error("%s DoInsert fail error %s", dbReq.request.TableName, err.Error())
	}
	dbRet.Type = dbReq.request.Type
	dbRet.Res = make([][]byte, 0)
	findSql := constpackage.LoginSql
	result, err := mysqlService.mysqlModule.Query(findSql, findArgs...)
	if err != nil {
		fmt.Print(err)
	} else {
		if tableName == "user" {
			var dbRetData []collect.User
			err = result.UnMarshal(&dbRetData)
			if err != nil {
				fmt.Print(err)
			}
			rowNum = len(dbRetData)

			for i := 0; i < rowNum; i++ {
				bytes, err := json.Marshal(dbRetData[i])
				if err != nil {
					rpcErr = rpc.RpcError(err.Error())
					dbRet.Res = emptyRes
					break
				}
				dbRet.Res = append(dbRet.Res, bytes)
			}
		}
		//从结构集中返序列化数据到结构体切片中，UnMarshal可以支持多个结果集
	}

	//mysqlService.responseRet(dbReq, err, int32(exec.RowsAffected))
	dbRet.RowNum = int32(rowNum)
	dbReq.responder(&dbRet, rpcErr)
}

func (mysqlService *MysqlService) responseRet(dbReq MysqlDBRequest, err error, effectRow int32) {
	var dbRet db.MysqlControllerRet
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

type MysqlDBRequest struct {
	request   *db.MysqlControllerReq
	responder rpc.Responder
}

func (mysqlService *MysqlService) RPC_MysqlDBRequest(responder rpc.Responder, request *db.MysqlControllerReq) error {
	log.Release("进入到了MysqlService-RPC_DBRequest")
	index := request.GetKey() % uint64(mysqlService.goroutineNum)
	if len(mysqlService.channelOptData[index]) == cap(mysqlService.channelOptData[index]) {
		log.Error("channel is full %d", index)

		responder(nil, rpc.RpcError("channel is full"))
		return nil
	}

	var dbRequest MysqlDBRequest
	dbRequest.request = request
	dbRequest.responder = responder

	mysqlService.channelOptData[index] <- dbRequest
	return nil
}

func (mysqlService *MysqlService) SetArgs(args []string) []interface{} {
	data := make([]interface{}, 0)
	for _, arg := range args {
		data = append(data, arg)
	}
	return data
}
