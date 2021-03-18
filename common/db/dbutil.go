package db

import (
	"gopkg.in/mgo.v2/bson"
)

func MakeUpsetId(collName string, id interface{}, data interface{}, key uint64, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_Upset
	req.Condition, _ = bson.Marshal(bson.D{{"_id", id}})
	req.Key = key
	out, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	req.Data = append(req.Data, out)
	return nil
}

func MakeRemoveId(collName string, id interface{}, key uint64, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_Del
	req.Condition, _ = bson.Marshal(bson.D{{"_id", id}})
	req.Key = key

	return nil
}

func MakeFind(collName string, condition bson.D, key uint64, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_Find
	req.Condition, _ = bson.Marshal(condition)
	req.Key = key

	return nil
}

func MakeSetOnInsertAndFind(collName string, id interface{}, updateInfo interface{}, key uint64, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_Upset
	req.Condition, _ = bson.Marshal(bson.D{{"_id", id}})
	req.Key = key
	out, err := bson.Marshal(updateInfo)
	if err != nil {
		return err
	}
	req.Data = append(req.Data, out)
	return nil
}

func MakeInsertId(collName string, data interface{}, key uint64, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_Insert
	req.Key = key
	out, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	req.Data = append(req.Data, out)
	return nil
}

//var mysqlData db.MysqlControllerReq
//	mysqlData.TableName = constpackage.UserTableName
//	mysqlData.Key = uint64(util.HashString2Number(loginInfo.PlatId))
//	mysqlData.Sql = constpackage.LoginSql
//	mysqlData.Args = []string{account, passWord}
//	mysqlData.Type = db.OptType_Find

func MakeMysql(tableName string, key uint64, sql string, args []string, optType OptType, req *MysqlControllerReq) {
	req.TableName = tableName
	req.Key = key
	req.Sql = sql
	req.Args = args
	req.Type = optType
}

//var req db.RedisControllerReq
//		req.Type = db.OptType_InsertNoFallBack
//		req.RKey = constpackage.UserRedisKey + res.Token
//		req.Key = uint64(util.HashString2Number(loginInfo.PlatId))
//		data, err := json.Marshal(user)
//		req.RValue = string(data)

func MakeRedis(optType OptType, rKey string, key uint64, val string, req *RedisControllerReq) {
	req.Type = optType
	req.RKey = rKey
	req.Key = key
	req.RValue = val
}
