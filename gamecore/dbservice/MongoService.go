package dbservice

import (
	"fmt"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/node"
	"github.com/duanhf2012/origin/rpc"
	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/sysmodule/mongomodule"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"runtime"
	"sunserver/common/db"
	"time"
)

func init() {
	node.Setup(&MongoService{})
}

var emptyRes [][]byte

type MongoService struct {
	service.Service
	mongoModule    mongomodule.MongoModule
	channelOptData []chan MongoDBRequest
	url            string
	dbName         string
	goroutineNum   uint32
	sessionNum     int
	dialTimeout    int
	syncTimeout    int
	channelNum     int
}

func (mongoService *MongoService) ReadCfg() error {
	mapMongoServiceCfg, ok := mongoService.GetServiceCfg().(map[string]interface{})
	if ok == false {
		return fmt.Errorf("MongoService config is error!")
	}

	//parse MsgRouter
	url, ok := mapMongoServiceCfg["Url"]
	if ok == false {
		return fmt.Errorf("MongoService config is error!")
	}
	mongoService.url = url.(string)

	dbName, ok := mapMongoServiceCfg["DBName"]
	if ok == false {
		return fmt.Errorf("MongoService config is error!")
	}
	mongoService.dbName = dbName.(string)

	goroutineNum, ok := mapMongoServiceCfg["GoroutineNum"]
	if ok == false {
		return fmt.Errorf("MongoService config is error!")
	}
	mongoService.goroutineNum = uint32(goroutineNum.(float64))

	sessionNum, ok := mapMongoServiceCfg["SessionNum"]
	if ok == false {
		return fmt.Errorf("MongoService config is error!")
	}
	mongoService.sessionNum = int(sessionNum.(float64))

	dialTimeout, ok := mapMongoServiceCfg["DialTimeout"]
	if ok == false {
		return fmt.Errorf("MongoService config is error!")
	}
	mongoService.dialTimeout = int(dialTimeout.(float64))

	syncTimeout, ok := mapMongoServiceCfg["SyncTimeout"]
	if ok == false {
		return fmt.Errorf("MongoService config is error!")
	}
	mongoService.syncTimeout = int(syncTimeout.(float64))

	channelNum, ok := mapMongoServiceCfg["ChannelNum"]
	if ok == false {
		return fmt.Errorf("MongoService config is error!")
	}
	mongoService.channelNum = int(channelNum.(float64))

	return nil
}

func (mongoService *MongoService) OnInit() error {
	fmt.Println("start init MongoService")
	defer fmt.Println("finish init MongoService")

	err := mongoService.ReadCfg()
	if err != nil {
		return err
	}

	err = mongoService.mongoModule.Init(mongoService.url, uint32(mongoService.sessionNum), time.Duration(mongoService.dialTimeout)*time.Second, time.Duration(mongoService.syncTimeout)*time.Second)
	if err != nil {
		return err
	}

	mongoService.channelOptData = make([]chan MongoDBRequest, mongoService.goroutineNum)
	for i := uint32(0); i < mongoService.goroutineNum; i++ {
		mongoService.channelOptData[i] = make(chan MongoDBRequest, mongoService.channelNum)
		go mongoService.ExecuteOptData(mongoService.channelOptData[i])
	}

	return nil
}

func (mongoService *MongoService) ExecuteOptData(channelOptData chan MongoDBRequest) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			err := fmt.Errorf("%v: %s", r, buf[:l])
			log.Error("core dump info:%+v\n", err)
			mongoService.ExecuteOptData(channelOptData)
		}
	}()
	for {
		select {
		case optData := <-channelOptData:
			switch optData.request.GetType() {
			case db.OptType_Del:
				mongoService.DoDel(optData)
			case db.OptType_Update:
				mongoService.DoUpdate(optData)
			case db.OptType_Find:
				mongoService.DoFind(optData)
			case db.OptType_Insert:
				mongoService.DoInsert(optData)
			case db.OptType_Insert + db.OptType_Update:
				mongoService.DoInsertUpdate(optData)
			case db.OptType_SetOnInsert:
				mongoService.DoSetOnInsert(optData)
			case db.OptType_SetOnInsert + db.OptType_Find:
				mongoService.DoSetOnInsertFind(optData)
			default:
				log.Error("optype %d is error.", optData.request.GetType())
			}
		}
	}
}

func (mongoService *MongoService) DoSetOnInsert(dbReq MongoDBRequest) {
	//1.选择数据库与表
	dataBase := mongoService.mongoModule.Take().DB(mongoService.dbName)
	collect := dataBase.C(dbReq.request.GetCollectName())

	//2.设置数据
	if len(dbReq.request.Data) != 1 {
		err := fmt.Errorf("%s DoUpdate data len is error %d.", dbReq.request.CollectName, len(dbReq.request.Data))
		log.Error(err.Error())
		mongoService.responseRet(dbReq, err, 0)
		return
	}
	var data interface{}
	err := bson.Unmarshal(dbReq.request.Data[0], &data)
	if err != nil {
		err := fmt.Errorf("%s DoInsertUpdate data Unmarshal error %s.", dbReq.request.CollectName, err.Error())
		log.Error(err.Error())
		mongoService.responseRet(dbReq, err, 0)
		return
	}

	//3.设置条件
	var condition interface{}
	bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	changeInfo, err := collect.Upsert(condition, bson.M{"$setOnInsert": data})

	if dbReq.responder.IsInvalid() == false {
		mongoService.responseRet(dbReq, err, int32(changeInfo.Updated))
	}
}

func (mongoService *MongoService) DoSetOnInsertFind(dbReq MongoDBRequest) {
	//1.选择数据库与表
	dataBase := mongoService.mongoModule.Take().DB(mongoService.dbName)
	collect := dataBase.C(dbReq.request.GetCollectName())

	//2.设置数据
	if len(dbReq.request.Data) != 1 {
		err := fmt.Errorf("%s DoUpdate data len is error %d.", dbReq.request.CollectName, len(dbReq.request.Data))
		log.Error(err.Error())
		mongoService.responseRet(dbReq, err, 0)
		return
	}
	var data interface{}
	uErr := bson.Unmarshal(dbReq.request.Data[0], &data)
	if uErr != nil {
		uErr := fmt.Errorf("%s DoInsertUpdate data Unmarshal error %s.", dbReq.request.CollectName, uErr.Error())
		log.Error(uErr.Error())
		mongoService.responseRet(dbReq, uErr, 0)
		return
	}

	//3.设置条件
	var condition interface{}
	bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	_, usErr := collect.Upsert(condition, bson.M{"$setOnInsert": data})
	if dbReq.responder.IsInvalid() == false && usErr != nil {
		mongoService.responseRet(dbReq, usErr, 0)
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

func (mongoService *MongoService) upSet(dbReq MongoDBRequest) (info *mgo.ChangeInfo, err error) {
	//1.选择数据库与表
	dataBase := mongoService.mongoModule.Take().DB(mongoService.dbName)
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
		mongoService.responseRet(dbReq, err, 0)
		return nil, err
	}

	//3.设置条件
	var condition interface{}
	bson.Unmarshal(dbReq.request.GetCondition(), &condition)

	return collect.Upsert(condition, data)
}

func (mongoService *MongoService) DoInsertUpdate(dbReq MongoDBRequest) {
	//1.选择数据库与表
	dataBase := mongoService.mongoModule.Take().DB(mongoService.dbName)
	collect := dataBase.C(dbReq.request.GetCollectName())

	//2.设置数据
	if len(dbReq.request.Data) != 1 {
		err := fmt.Errorf("%s DoUpdate data len is error %d.", dbReq.request.CollectName, len(dbReq.request.Data))
		log.Error(err.Error())
		mongoService.responseRet(dbReq, err, 0)
		return
	}
	var data interface{}
	err := bson.Unmarshal(dbReq.request.Data[0], &data)
	if err != nil {
		err := fmt.Errorf("%s DoInsertUpdate data Unmarshal error %s.", dbReq.request.CollectName, err.Error())
		log.Error(err.Error())
		mongoService.responseRet(dbReq, err, 0)
		return
	}

	//3.设置条件
	var condition interface{}
	bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	changeInfo, err := collect.Upsert(condition, data)

	if dbReq.responder.IsInvalid() == false {
		mongoService.responseRet(dbReq, err, int32(changeInfo.Updated))
	}
}

func (mongoService *MongoService) DoDel(dbReq MongoDBRequest) {
	//1.选择数据库与表
	dataBase := mongoService.mongoModule.Take().DB(mongoService.dbName)
	collect := dataBase.C(dbReq.request.GetCollectName())

	//2.设置条件
	var condition interface{}
	bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	err := collect.Remove(condition)
	if err != nil {
		log.Error("%s DoUpdate fail error %s", dbReq.request.CollectName, err.Error())
	}
	mongoService.responseRet(dbReq, err, 0)
}

func (mongoService *MongoService) DoUpdate(dbReq MongoDBRequest) {
	//1.选择数据库与表
	dataBase := mongoService.mongoModule.Take().DB(mongoService.dbName)
	collect := dataBase.C(dbReq.request.GetCollectName())

	//2.设置条件
	var condition interface{}
	bson.Unmarshal(dbReq.request.GetCondition(), &condition)

	//3.设置数据
	if len(dbReq.request.Data) != 1 {
		err := fmt.Errorf("%s DoUpdate data len is error %d.", dbReq.request.CollectName, len(dbReq.request.Data))
		log.Error(err.Error())
		mongoService.responseRet(dbReq, err, 0)
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
		mongoService.responseRet(dbReq, err, 0)
	}
}

func (mongoService *MongoService) DoFind(dbReq MongoDBRequest) {
	//1.选择数据库与表
	dataBase := mongoService.mongoModule.Take().DB(mongoService.dbName)
	collect := dataBase.C(dbReq.request.GetCollectName())

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

func (mongoService *MongoService) responseRet(dbReq MongoDBRequest, err error, effectRow int32) {
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

func (mongoService *MongoService) DoInsert(dbReq MongoDBRequest) {
	//1.选择数据库与表
	dataBase := mongoService.mongoModule.Take().DB(mongoService.dbName)
	collect := dataBase.C(dbReq.request.GetCollectName())

	var data []interface{}
	data = make([]interface{}, len(dbReq.request.Data))
	for i := 0; i < len(data); i++ {
		err := bson.Unmarshal(dbReq.request.Data[i], &data[i])
		if err != nil {
			err := fmt.Errorf("%s DoInsert fail %s", dbReq.request.CollectName, err.Error())
			mongoService.responseRet(dbReq, err, 0)
			return
		}
	}

	err := collect.Insert(data)
	if err != nil {
		log.Error("%s DoInsert fail error %s", dbReq.request.CollectName, err.Error())
	}
	mongoService.responseRet(dbReq, err, 0)
}

type MongoDBRequest struct {
	request   *db.DBControllerReq
	responder rpc.Responder
}

func (mongoService *MongoService) RPC_MongoDBRequest(responder rpc.Responder, request *db.DBControllerReq) error {
	//从 LoginModule rpc发往db 进行数据处理
	index := request.GetKey() % uint64(mongoService.goroutineNum)
	if len(mongoService.channelOptData[index]) == cap(mongoService.channelOptData[index]) {
		log.Error("channel is full %d", index)

		responder(nil, rpc.RpcError("channel is full"))
		return nil
	}

	var MongoDBRequest MongoDBRequest
	MongoDBRequest.request = request
	MongoDBRequest.responder = responder

	//往管道发数据
	mongoService.channelOptData[index] <- MongoDBRequest
	return nil
}
