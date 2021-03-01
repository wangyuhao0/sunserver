package dbservice

/*mport (
	"fmt"
	"github.com/duanhf2012/origin/sysmodule/mongomodule"
	"github.com/golang/protobuf/proto"
	"gopkg.in/mgo.v2/bson"
	"sunserver/common/db"
	"testing"
	"time"
)

type Student struct {
	ID     bson.ObjectId `bson:"_id"`
	Name   string        `bson: "name"`
	Age    int           `bson: "age"`
	Sid    string        `bson: "sid"`
	Status int           `bson: "status"`
}

func findRequest(tableName string, key uint64, condition bson.M, selectField bson.M, sort string, maxRow int, request *db.DBControllerReq) (err error) {
	request.Condition, err = bson.Marshal(condition)
	if err != nil {
		return
	}

	if selectField != nil {
		request.SelectField, err = bson.Marshal(selectField)
		if err != nil {
			return
		}
	}

	request.Type = db.OptType_Find.Enum()
	request.CollectName = proto.String(tableName)
	request.Sort = proto.String(sort)
	request.MaxRow = proto.Int32(int32(maxRow))
	request.Key = proto.Uint64(key)
	return
}

func findRespone(ret *db.DBControllerRet) {
	fmt.Print(ret.GetRowNum())
	fmt.Print(ret.GetError())

	var s Student
	for i := 0; i < len(ret.Res); i++ {
		bson.Unmarshal(ret.Res[i], &s)
		fmt.Print(s)
	}
}

func findDo(session *mongomodule.Session, request *db.DBControllerReq) {
	collect := session.DB("test2").C(request.GetCollectName())
	var condition interface{}
	bson.Unmarshal(request.GetCondition(), &condition)
	finds := collect.Find(condition)
	if request.GetSort() != "" {
		finds = finds.Sort(request.GetSort())
	}

	var res []interface{}
	if request.GetMaxRow() > 0 {
		finds = finds.Limit(int(request.GetMaxRow()))
	}

	var ret db.DBControllerRet
	var err error
	if request.GetMaxRow() != 0 {
		finds.All(&res)
		//序列化结果
		ret.Type = request.Type
		ret.Res = make([][]byte, len(res))
		for i := 0; i < len(res); i++ {
			ret.Res[i], err = bson.Marshal(res[i])
			if err != nil {
				ret.Error = proto.String(err.Error())
				ret.Res = emptyRes
				break
			}
		}
	}
	var rowNum int
	rowNum, err = finds.Count()
	if err != nil {
		ret.Error = proto.String(err.Error())
		ret.Res = emptyRes
	}

	ret.RowNum = proto.Int32(int32(rowNum))
	findRespone(&ret)
}

func Test_Example(t *testing.T) {
	module := mongomodule.MongoModule{}
	//mongodb://admin@123456:127.0.0.1:27017
	module.Init("mongodb://admin:123456@49.232.105.112:27017/test2", 100, 5*time.Second, 5*time.Second)

	// take session
	s := module.Take()
	var request db.DBControllerReq
	findRequest("t_student", 11, bson.M{"_id": bson.ObjectIdHex("5fa532214af0e635547207d0")}, nil, "", -1, &request)
	findDo(s, &request)

	//c := s.DB("test2").C("t_student")

}
*/