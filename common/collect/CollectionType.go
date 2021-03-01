package collect

//新增表需要新增类型
type CollectionType int

const (
	CTUserInfo CollectionType = iota

	CTMax
)

//新增多行表
type MultiCollectionType int

const (
	MCTUserMail MultiCollectionType = iota

	MCTMax
)