package collect

type User struct {
	Id         uint64 `json:"id"`       //生成最新id
	Rank       uint64 `json:"rank"`     //排位分
	Account    string `json:"account"`  //账号
	PassWord   string `json:"password"` //密码
	Coin       uint64 `json:"coin"`     //金币
	NickName   string `json:"nick_name"`
	Avatar     string `json:"avatar"`      //头像
	Sex        int    `json:"sex"`         //性别
	Grade      int    `json:"grade"`       // 等级
	CreateTime int64  `json:"create_time"` // 创建时间
	UpdateTime int64  `json:"update_time"` // 修改时间
}

/*func (user User) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id": user.Id,
		"rank":  user.Rank,
		"account": user.Account,
		"passWord":  user.PassWord,
		"coin": user.Coin,
		"nickName":  user.NickName,
		"avatar": user.Avatar,
		"sex":  user.Sex,
		"grade": user.Grade,
		"createTime":  user.CreateTime,
		"updateTime": user.UpdateTime,
	})
}*/
