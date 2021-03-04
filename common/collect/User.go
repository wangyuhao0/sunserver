package collect

type User struct {
	Id         uint64 `json:"id"`   //生成最新id
	Rank       uint64 `json:"rank"` //排位分
	Coin       uint64 `json:"coin"` //金币
	NickName   string `json:"nick_name"`
	Avatar     string `json:"avatar"`      //头像
	Sex        int    `json:"sex"`         //性别
	Grade      int    `json:"grade"`       // 等级
	CreateTime int64  `json:"create_time"` // 创建时间
	UpdateTime int64  `json:"update_time"` // 修改时间
}
