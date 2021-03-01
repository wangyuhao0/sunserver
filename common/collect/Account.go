package collect


var AccountCollectName = "Account"
type CAccount struct {
	UserId    	uint64  `bson:"_id"`   //生成最新id
	PlatType 	int  	`bson:"PlatType"` //平台类型
	PlatId 		string 	`bson:"PlatId"`   //平台id
}


