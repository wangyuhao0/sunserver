package dbcollection

import (
	"container/list"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/rpc"
	"gopkg.in/mgo.v2/bson"
	"sunserver/common/collect"
	"sunserver/common/db"
	"sunserver/common/util"
)

const MaxRowNum int32 = 2000 //单集合支持最大行数
const strRpcDBRequest string = "MongoService.RPC_MongoDBRequest"

//玩家存档数据
type PlayerDB struct {
	rpc.IRpcHandler                                      //rpc操作接口
	Id uint64                                            //唯一标识，即UserId
	collection      [collect.CTMax]collect.ICollection   //单行集合
	multiCollection [collect.MCTMax]MultiRowData          //多行实在存档
	onLoadDBEnd     func(succ bool)                      //加载结束回调
	totalProgress   int                                  //总进度
	loadProgress    int                                  //当前加载的进度

	//新增表需要新增以下
	collect.CUserInfo                                    //用户表

	//多行表处理
	mailTemplate collect.CMailInfo						 //邮件表
}

//多行数据
type MultiRowData struct {
	template        collect.IMultiCollection
	mapCollection   map[interface{}]*list.Element //key,value
	listICollection list.List
}

func (playerDB *PlayerDB) OnInit(rpcHandler rpc.IRpcHandler,onLoadDBEnd func(ucc bool)) {
	//以下加入注册新表
	playerDB.IRpcHandler = rpcHandler
	playerDB.RegCollection(&playerDB.CUserInfo)
	playerDB.RegMultiCollection(&playerDB.mailTemplate)

	playerDB.onLoadDBEnd = onLoadDBEnd
	playerDB.loadProgress = 0
}

func (playerDB *PlayerDB) RegCollection(coll collect.ICollection) {
	if coll.GetCollectionType() >= collect.CTMax {
		panic("collection error!")
	}

	playerDB.collection[coll.GetCollectionType()] = coll
	playerDB.totalProgress++
}

func (playerDB *PlayerDB) RegMultiCollection(coll collect.IMultiCollection) {
	if coll.GetCollectionType() >= collect.MCTMax {
		panic("collection error!")
	}

	multiRowData := MultiRowData{}
	multiRowData.template = coll
	multiRowData.mapCollection = make(map[interface{}]*list.Element, MaxRowNum)
	playerDB.multiCollection[coll.GetCollectionType()] = multiRowData
	playerDB.totalProgress++
}

func (playerDB *PlayerDB) LoadFromDB() {
	//1.Load单行表
	for i := collect.CollectionType(0); i < collect.CTMax; i++ {
		if playerDB.collection[i] != nil {
			playerDB.load(i)
		}
	}

	//2.Load多行表
	for i := collect.MultiCollectionType(0); i < collect.MCTMax; i++ {
		if playerDB.getMultiCollection(i).template != nil {
			playerDB.loadMulti(i)
		}
	}
}

func (playerDB *PlayerDB) SaveToDB(bForce bool) {
	//没加载完不允许存档
	if playerDB.IsLoadFinish() == false {
		log.Warning("userid:%d not load finish.",playerDB.Id)
		return
	}

	for i := collect.CollectionType(0); i < collect.CTMax; i++ {
		coll := playerDB.collection[i]
		if coll == nil || (bForce == true || coll.IsDirty() == false) {
			continue
		}
		coll.OnSave()
		playerDB.upsetCollection(coll)
	}
}

func (playerDB *PlayerDB) loadMulti(typ collect.MultiCollectionType) {
	var coll collect.IMultiCollection
	multiCollection:= playerDB.getMultiCollection(typ)
	if multiCollection != nil && multiCollection.template != nil {
		coll = multiCollection.template
	}

	if coll == nil {
		playerDB.onLoadDBEnd(false)
		log.Error("loadMulti cannot find collection,load fail!")
		return
	}

	var req db.DBControllerReq
	db.MakeFind(coll.GetCollName(), coll.GetCondition(playerDB.Id), playerDB.Id, &req)
	req.MaxRow = MaxRowNum

	nodeId := util.GetBestNodeId(strRpcDBRequest,playerDB.Id)
	if nodeId == 0 {
		log.Error("Cannot find %s nodeId!",strRpcDBRequest)
		return
	}

	playerDB.AsyncCallNode(nodeId,strRpcDBRequest,&req,func(res *db.DBControllerRet,err error){
		if err != nil {
			log.Error("AsyncCall userid:%d,type:%d,error :%+v\n",playerDB.Id,typ,err)
			playerDB.onLoadDBEnd(false)
			return
		}
		playerDB.dbMultiLoadCallBack(typ,res,err)
	})
}

func (playerDB *PlayerDB) load(typ collect.CollectionType) {
	var coll collect.ICollection
	if playerDB.collection[typ] != nil {
		coll = playerDB.collection[typ]
	}

	if coll == nil {
		playerDB.onLoadDBEnd(false)
		log.Error("load cannot find collection,load fail!")
		return
	}

	var req db.DBControllerReq
	db.MakeFind(coll.GetCollName(), coll.GetCondition(playerDB.Id), playerDB.Id, &req)

	nodeId := util.GetBestNodeId(strRpcDBRequest,playerDB.Id)
	if nodeId == 0 {
		log.Error("Cannot find %s nodeId!",strRpcDBRequest)
		return
	}

	playerDB.AsyncCallNode(nodeId,strRpcDBRequest,&req,func(res *db.DBControllerRet,err error){
		if err != nil {
			log.Error("AsyncCall userid:%d,type:%d,error :%+v\n",playerDB.Id,typ,err)
			playerDB.onLoadDBEnd(false)
			return
		}
		playerDB.dbLoadCallBack(typ,res,err)
	})
}

func (playerDB *PlayerDB) dbLoadCallBack(collType collect.CollectionType, res *db.DBControllerRet,err error) {
	if err != nil {
		log.Error("load userid %d db collType %d is error!", playerDB.Id,collType)
		playerDB.onLoadDBEnd(false)
		return
	}

	if playerDB.collection[collType] != nil {
		if len(res.Res)>0 {
			//加载单行数据
			err := bson.Unmarshal(res.Res[0], playerDB.collection[collType])
			if err != nil {
				log.Error("bson.Unmarshal fail %s,userid %d collType %d",err.Error(),playerDB.Id,collType)
				playerDB.onLoadDBEnd(false)
				return
			}
		}
		playerDB.collection[collType].OnLoadSucc(len(res.Res)==0, playerDB.Id)
	}
	playerDB.loadProgress++

	if playerDB.loadProgress >= playerDB.totalProgress {
		playerDB.onLoadDBEnd(true)
	}
}

func (playerDB *PlayerDB) dbMultiLoadCallBack(collType collect.MultiCollectionType, res *db.DBControllerRet, err error) {
	if err != nil {
		log.Error("load userid %d db multi collType %d is error!", playerDB.Id,collType)
		playerDB.onLoadDBEnd(false)
		return
	}

	multiCollection := playerDB.getMultiCollection(collType)
	if multiCollection != nil && multiCollection.template != nil {
		//加载多行数据
		err := multiCollection.loadFromDB(res)
		if err != nil {
			log.Error("load multiCollection fail %s,userid %d collType %d",err.Error(),playerDB.Id,collType)
			playerDB.onLoadDBEnd(false)
			return
		}
	}
	playerDB.loadProgress++

	if playerDB.loadProgress >= playerDB.totalProgress {
		playerDB.onLoadDBEnd(true)
	}
}

func (multiRowData *MultiRowData) loadFromDB(res *db.DBControllerRet) error {
	for _, data := range res.Res {
		rowData := multiRowData.template.MakeRow()
		err := bson.Unmarshal(data, rowData)
		if err != nil {
			return err
		}

		pElem := multiRowData.listICollection.PushBack(rowData)
		multiRowData.mapCollection[rowData.GetId()] = pElem
	}

	return nil
}

func (multiRowData *MultiRowData) Clean(){
	multiRowData.template = nil
	multiRowData.mapCollection = nil
	multiRowData.listICollection = list.List{}
}

func (playerDB *PlayerDB) ExecDB(req *db.DBControllerReq) bool{
	nodeId := util.GetBestNodeId(strRpcDBRequest,playerDB.Id)
	if nodeId <= 0 {
		log.Error("cannot find nodeId from rpcMethod %s,userid %d,collectName %s.",strRpcDBRequest,playerDB.Id,req.CollectName)
		return false
	}

	err := playerDB.GoNode(nodeId,strRpcDBRequest,req)
	if err != nil {
		log.Error("ExecDB fail:go node error :%s,userid %d,collectName %s.",err.Error(),playerDB.Id,req.CollectName)
		return false
	}

	return true
}

func (playerDB *PlayerDB) RemoveMultiRow(collType collect.MultiCollectionType, key interface{}) bool {
	collection := playerDB.getMultiCollection(collType)
	if collection.template != nil {
		elem, ok := collection.mapCollection[key]
		if ok == false {
			log.Warning("cannot find key %+v,userid %d,collType %d.",key,playerDB.Id,collType)
			return false
		}

		//实时存档
		coll := elem.Value.(collect.IMultiCollection)
		collection.listICollection.Remove(elem)
		delete(collection.mapCollection, key)
		var req db.DBControllerReq
		db.MakeRemoveId(coll.GetCollName(), coll.GetId(), playerDB.Id, &req)
		return playerDB.ExecDB(&req)
	}

	return false
}

func (playerDB *PlayerDB) GetRowByKey(collType collect.MultiCollectionType, key interface{}) collect.IMultiCollection {
	collection := playerDB.getMultiCollection(collType)
	if collection.template != nil {
		log.Error("template is nil,userid %d,collType %d",playerDB.Id,collType)
		return nil
	}
	elem, ok := collection.mapCollection[key]
	if ok == false {
		log.Error("cannot find key %+v,userid %d,collType %d",key,playerDB.Id,collType)
		return nil
	}

	return elem.Value.(collect.IMultiCollection)
}

func (playerDB *PlayerDB) upsetCollection(coll collect.ICollection) bool {
	var req db.DBControllerReq
	err := db.MakeUpsetId(coll.GetCollName(), coll.GetId(), coll, playerDB.Id, &req)
	if err != nil {
		log.Error("make upsetid fail %s", err.Error())
		return false
	}
	return playerDB.ExecDB(&req)
}

func (playerDB *PlayerDB) upsetMultiCollection(coll collect.IMultiCollection) bool {
	var req db.DBControllerReq
	err := db.MakeUpsetId(coll.GetCollName(), coll.GetId(), coll, playerDB.Id, &req)
	if err != nil {
		log.Error("make multi upsetid fail %s", err.Error())
		return false
	}
	return playerDB.ExecDB(&req)
}


func (playerDB *PlayerDB) ApplyMultiRow(coll collect.ICollection) bool {
	return playerDB.upsetCollection(coll)
}

func (playerDB *PlayerDB) InsertMultiRow(coll collect.IMultiCollection, needSave bool) bool {
	collection := playerDB.getMultiCollection(coll.GetCollectionType())
	if collection.template == nil {
		log.Error("template is nil,userid %d,collType %d",playerDB.Id,coll.GetCollectionType())
		return false
	}

	elem := collection.listICollection.PushBack(coll)
	collection.mapCollection[coll.GetId()] = elem

	//超过最大条数，自动删除超出的
	if collection.listICollection.Len() > int(MaxRowNum) {
		frontElem := collection.listICollection.Front()
		coll := frontElem.Value.(collect.IMultiCollection)
		playerDB.RemoveMultiRow(coll.GetCollectionType(),coll.GetId())
		log.Release("lfy------- remove item: %+v", coll)
	}

	//存档
	if needSave {
		return playerDB.upsetMultiCollection(coll)
	}

	return true
}

func (playerDB *PlayerDB) GetMultiRow(collType collect.MultiCollectionType) list.List {
	return playerDB.multiCollection[collType].listICollection
}

func (playerDB *PlayerDB) IsLoadFinish() bool{
	return playerDB.loadProgress >= playerDB.totalProgress
}

func (playerDB *PlayerDB) Clear(){
	playerDB.Id = 0
	playerDB.totalProgress = 0
	playerDB.loadProgress = 0

	for i:=collect.MultiCollectionType(0);i<collect.MCTMax;i++{
		playerDB.getMultiCollection(i).Clean()
	}

	playerDB.CUserInfo.Clean()
}

func (playerDB *PlayerDB) getMultiCollection(collType collect.MultiCollectionType) *MultiRowData {
	return &playerDB.multiCollection[collType]
}
