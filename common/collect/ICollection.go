package collect

import "gopkg.in/mgo.v2/bson"

type CBaseCollection struct {
	dirty bool
}

func (slf *CBaseCollection) MakeDirty() {
	slf.dirty = true
}

func (slf *CBaseCollection) IsDirty() bool {
	return slf.dirty
}

//ICollection MongoDB中仅一条数据使用，如用户数据
type ICollection interface {
	Clean()										//清理所有数据
	GetId() interface{}							//获取主键值，一般是userID
	GetCollectionType() CollectionType			//获取数据类型
	GetCollName() string						//获取MongoDB表名
	MakeDirty()									//数据置脏，确保每次修改数据都需要调用该接口
	IsDirty() bool								//获取数据是否是脏数据
	GetSelf() ICollection						//返回自身数据对象指针 ps：在MongoDB中仅一个数据使用，如玩家数据
	OnLoadSucc(notFound bool, userID uint64)	//MongoDB数据加载成功后调用
	OnSave()									//数据存放MongoDB成功后调用
	GetCondition(value interface{}) bson.D		//获取编辑查询条件
}

//IMultiCollection MongoDB中多条数据时使用，如用户的mail数据
type IMultiCollection interface {
	Clean()										//清理所有数据
	GetId() interface{}							//获取主键值，一般是userID
	GetCollectionType() MultiCollectionType		//获取数据类型
	GetCollName() string						//获取MongoDB表名
	MakeRow() IMultiCollection				//返回一个新的数据对象指针 ps：在MongoDB中位多行数据使用，如用户的邮件数据
	OnLoadSucc(notFound bool, userID uint64)	//MongoDB数据加载成功后调用
	OnSave()									//数据存放MongoDB成功后调用
	GetCondition(value interface{}) bson.D		//获取编辑查询条件
}
