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
