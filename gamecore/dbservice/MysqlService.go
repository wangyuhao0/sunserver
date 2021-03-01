package dbservice
/*
import (
	"fmt"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/node"
	"github.com/duanhf2012/origin/rpc"
	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/sysmodule/mysqlmondule"
	"github.com/duanhf2012/origin/util/timer"
	"go/token"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"runtime"
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
	channelOptData []chan DBRequest
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
	log.Release("start init MysqlService")

	err := mysqlService.ReadCfg()
	if err != nil {
		return err
	}

	err = mysqlService.mysqlModule.Init(mysqlService.url, mysqlService.userName, mysqlService.passWord, mysqlService.dbName, mysqlService.maxConn)
	if err != nil {
		return err
	}

	mysqlService.channelOptData = make([]chan DBRequest, mysqlService.goroutineNum)
	for i := uint32(0); i < mysqlService.goroutineNum; i++ {
		mysqlService.channelOptData[i] = make(chan DBRequest, mysqlService.channelNum)
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

	log.Release("finish init MysqlService")
	return nil
}



func (MysqlService *MysqlService) PrintDBCost(tm *timer.Ticker) {
	averageCostTime := int64(0)
	if MysqlService.dbDealCount != 0 {
		averageCostTime = MysqlService.dbAllCostTime / int64(MysqlService.dbDealCount)
	}
	log.Release("MysqlService dbDealCount[%d], dbMaxCostTime[%d], averageCostTime[%d]", MysqlService.dbDealCount, MysqlService.dbMaxCostTime, averageCostTime)
}

func (mysqlService *MysqlService) ExecuteOptData(channelOptData chan DBRequest) {
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
			case db.OptType_Insert + db.OptType_Update:
				mysqlService.DoInsertUpdate(optData)
			case db.OptType_SetOnInsert:
				mysqlService.DoSetOnInsert(optData)
			case db.OptType_SetOnInsert + db.OptType_Find:
				mysqlService.DoSetOnInsertFind(optData)
			case db.OptType_Upset:
				mysqlService.DoUpSet(optData)
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


func (mysqlService *MysqlService) DoSetOnInsert(dbReq DBRequest) {
	//1.选择数据库与表
	request := dbReq.request
	collectName := request.CollectName
	request.
	mysqlService.mysqlModule.Exec()

	//2.设置数据
	if len(dbReq.request.Data) != 1 {
		err := fmt.Errorf("%s DoUpdate data len is error %d.", dbReq.request.CollectName, len(dbReq.request.Data))
		log.Error(err.Error())
		MysqlService.responseRet(dbReq, err, 0)
		return
	}
	var data interface{}
	err := bson.Unmarshal(dbReq.request.Data[0], &data)
	if err != nil {
		err := fmt.Errorf("%s DoInsertUpdate data Unmarshal error %s.", dbReq.request.CollectName, err.Error())
		log.Error(err.Error())
		MysqlService.responseRet(dbReq, err, 0)
		return
	}

	//3.设置条件
	var condition interface{}
	bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	changeInfo, err := collect.Upsert(condition, bson.M{"$setOnInsert": data})

	if dbReq.responder.IsInvalid() == false {
		MysqlService.responseRet(dbReq, err, int32(changeInfo.Updated))
	}
}

func (MysqlService *MysqlService) DoSetOnInsertFind(dbReq DBRequest) {
	//1.选择数据库与表
	req := dbReq.request
	collectName := req.CollectName
	req.
	dataBase := MysqlService.mongoModule.Take().DB(MysqlService.dbName)
	collect := dataBase.C(dbReq.request.GetCollectName())

	//2.设置数据
	if len(dbReq.request.Data) != 1 {
		err := fmt.Errorf("%s DoUpdate data len is error %d.", dbReq.request.CollectName, len(dbReq.request.Data))
		log.Error(err.Error())
		MysqlService.responseRet(dbReq, err, 0)
		return
	}
	var data interface{}
	uErr := bson.Unmarshal(dbReq.request.Data[0], &data)
	if uErr != nil {
		uErr := fmt.Errorf("%s DoInsertUpdate data Unmarshal error %s.", dbReq.request.CollectName, uErr.Error())
		log.Error(uErr.Error())
		MysqlService.responseRet(dbReq, uErr, 0)
		return
	}

	//3.设置条件
	var condition interface{}
	bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	_, usErr := collect.Upsert(condition, bson.M{"$setOnInsert": data})
	if dbReq.responder.IsInvalid() == false && usErr != nil {
		MysqlService.responseRet(dbReq, usErr, 0)
		return
	}

	if dbReq.responder == nil {
		return
	}

	//4.设置条件
	finds := collect.Find(condition)

	//5.排序
	if dbReq.request.GetSort() != "" {
		finds = finds.Sort(dbReq.request.GetSort())
	}

	//6.limit
	var res []interface{}
	if dbReq.request.GetMaxRow() > 0 {
		finds = finds.Limit(int(dbReq.request.GetMaxRow()))
	}

	//7.获取结果集
	var dbRet db.DBControllerRet
	var err error
	//if dbReq.request.GetMaxRow() != 0 {
	finds.All(&res)
	//序列化结果
	dbRet.Type = dbReq.request.Type
	dbRet.Res = make([][]byte, len(res))
	var rpcErr rpc.RpcError
	for i := 0; i < len(res); i++ {
		dbRet.Res[i], err = bson.Marshal(res[i])
		if err != nil {
			rpcErr = rpc.RpcError(err.Error())
			dbRet.Res = emptyRes
			break
		}
	}
	//}
	var rowNum int
	rowNum, err = finds.Count()
	if err != nil {
		rpcErr = rpc.RpcError(err.Error())
		dbRet.Res = emptyRes
	}

	dbRet.RowNum = int32(rowNum)
	dbReq.responder(&dbRet, rpcErr)
}

func (MysqlService *MysqlService) DoUpSet(dbReq DBRequest) (info *mgo.ChangeInfo, err error) {
	//1.选择数据库与表
	dataBase := MysqlService.mongoModule.Take().DB(MysqlService.dbName)
	collect := dataBase.C(dbReq.request.GetCollectName())

	//2.设置数据
	if len(dbReq.request.Data) != 1 {
		err := fmt.Errorf("%s DoUpdate data len is error %d.", dbReq.request.CollectName, len(dbReq.request.Data))
		log.Error(err.Error())
		return nil, err
	}
	var data interface{}
	err = bson.Unmarshal(dbReq.request.Data[0], &data)
	if err != nil {
		err := fmt.Errorf("%s DoInsertUpdate data Unmarshal error %s.", dbReq.request.CollectName, err.Error())
		log.Error(err.Error())
		MysqlService.responseRet(dbReq, err, 0)
		return nil, err
	}

	//3.设置条件
	var condition interface{}
	bson.Unmarshal(dbReq.request.GetCondition(), &condition)

	return collect.Upsert(condition, data)
}

func (MysqlService *MysqlService) DoInsertUpdate(dbReq DBRequest) {
	//1.选择数据库与表
	dataBase := MysqlService.mongoModule.Take().DB(MysqlService.dbName)
	collect := dataBase.C(dbReq.request.GetCollectName())

	//2.设置数据
	if len(dbReq.request.Data) != 1 {
		err := fmt.Errorf("%s DoUpdate data len is error %d.", dbReq.request.CollectName, len(dbReq.request.Data))
		log.Error(err.Error())
		MysqlService.responseRet(dbReq, err, 0)
		return
	}
	var data interface{}
	err := bson.Unmarshal(dbReq.request.Data[0], &data)
	if err != nil {
		err := fmt.Errorf("%s DoInsertUpdate data Unmarshal error %s.", dbReq.request.CollectName, err.Error())
		log.Error(err.Error())
		MysqlService.responseRet(dbReq, err, 0)
		return
	}

	//3.设置条件
	var condition interface{}
	bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	changeInfo, err := collect.Upsert(condition, data)

	if dbReq.responder.IsInvalid() == false {
		MysqlService.responseRet(dbReq, err, int32(changeInfo.Updated))
	}
}

func (MysqlService *MysqlService) DoDel(dbReq DBRequest) {
	//1.选择数据库与表
	dataBase := MysqlService.mongoModule.Take().DB(MysqlService.dbName)
	collect := dataBase.C(dbReq.request.GetCollectName())

	//2.设置条件
	var condition interface{}
	bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	err := collect.Remove(condition)
	if err != nil {
		log.Error("%s DoUpdate fail error %s", dbReq.request.CollectName, err.Error())
	}
	MysqlService.responseRet(dbReq, err, 0)
}

func (MysqlService *MysqlService) DoUpdate(dbReq DBRequest) {
	//1.选择数据库与表
	dataBase := MysqlService.mongoModule.Take().DB(MysqlService.dbName)
	collect := dataBase.C(dbReq.request.GetCollectName())

	//2.设置条件
	var condition interface{}
	bson.Unmarshal(dbReq.request.GetCondition(), &condition)

	//3.设置数据
	if len(dbReq.request.Data) != 1 {
		err := fmt.Errorf("%s DoUpdate data len is error %d.", dbReq.request.CollectName, len(dbReq.request.Data))
		log.Error(err.Error())
		MysqlService.responseRet(dbReq, err, 0)
		return
	}
	var data interface{}
	bson.Unmarshal(dbReq.request.Data[0], &data)
	//3.更新
	err := collect.Update(condition, data)
	if err != nil {
		log.Error("%s DoUpdate fail error %s", dbReq.request.CollectName, err.Error())
	}

	if dbReq.responder.IsInvalid() == false {
		MysqlService.responseRet(dbReq, err, 0)
	}
}




func (mysqlService *MysqlService) DoFind(dbReq DBRequest) {
	//1.选择数据库与表
	result, err := mysqlService.mysqlModule.Query(dbReq.request.Sql, dbReq.request.Args)

	if err!=nil {
		fmt.Print(err)
	}else{
		dbRet :=

		//从结构集中返序列化数据到结构体切片中，UnMarshal可以支持多个结果集
		err = result.UnMarshal(&dbRet)
		if err !=nil {
			fmt.Print(err)
		}
	}


	//2.设置条件
	var dbRet db.DBControllerRet
	var condition interface{}
	err := bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	if err != nil {
		dbRet.Res = emptyRes
		dbReq.responder(&dbRet, rpc.RpcError(err.Error()))
	}
	finds := collect.Find(condition)

	//3.排序
	if dbReq.request.GetSort() != "" {
		finds = finds.Sort(dbReq.request.GetSort())
	}

	//4.limit
	var res []interface{}
	if dbReq.request.GetMaxRow() > 0 {
		finds = finds.Limit(int(dbReq.request.GetMaxRow()))
	}

	//5.获取结果集
	var rpcErr rpc.RpcError
	if dbReq.request.GetMaxRow() != 0 {
		finds.All(&res)
		//序列化结果
		dbRet.Type = dbReq.request.Type
		dbRet.Res = make([][]byte, len(res))
		for i := 0; i < len(res); i++ {
			dbRet.Res[i], err = bson.Marshal(res[i])
			if err != nil {
				rpcErr = rpc.RpcError(err.Error())
				dbRet.Res = emptyRes
				break
			}
		}
	}
	var rowNum int
	rowNum, err = finds.Count()
	if err != nil {
		rpcErr = rpc.RpcError(err.Error())
		dbRet.Res = emptyRes
	}

	dbRet.RowNum = int32(rowNum)
	dbReq.responder(&dbRet, rpcErr)
}

func (MysqlService *MysqlService) responseRet(dbReq DBRequest, err error, effectRow int32) {
	var dbRet db.DBControllerRet
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

func (MysqlService *MysqlService) DoInsert(dbReq DBRequest) {
	//1.选择数据库与表
	dataBase := MysqlService.mongoModule.Take().DB(MysqlService.dbName)
	collect := dataBase.C(dbReq.request.GetCollectName())

	var data []interface{}
	data = make([]interface{}, len(dbReq.request.Data))
	for i := 0; i < len(data); i++ {
		err := bson.Unmarshal(dbReq.request.Data[i], &data[i])
		if err != nil {
			err := fmt.Errorf("%s DoInsert fail %s", dbReq.request.CollectName, err.Error())
			MysqlService.responseRet(dbReq, err, 0)
			return
		}
	}

	err := collect.Insert(data...)
	if err != nil {
		log.Error("%s DoInsert fail error %s", dbReq.request.CollectName, err.Error())
	}
	MysqlService.responseRet(dbReq, err, 0)
}

type DBRequest struct {
	request   *db.MysqlControllerReq
	responder rpc.Responder
}

func (MysqlService *MysqlService) RPC_DBRequest(responder rpc.Responder, request *db.DBControllerReq) error {
	log.Release("进入到了MysqlService-RPC_DBRequest")
	index := request.GetKey() % uint64(MysqlService.goroutineNum)
	if len(MysqlService.channelOptData[index]) == cap(MysqlService.channelOptData[index]) {
		log.Error("channel is full %d", index)

		responder(nil, rpc.RpcError("channel is full"))
		return nil
	}

	var dbRequest DBRequest
	dbRequest.request = request
	dbRequest.responder = responder

	MysqlService.channelOptData[index] <- dbRequest
	return nil
}
*/