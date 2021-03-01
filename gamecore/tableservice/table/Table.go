package table


type Table struct {
	uuid string //桌子uuid
	tableType int32 //table类型 和房间类型一致
	roomUuidList []string //组成对局的房间
	//连接的客户端
	clientList []uint64 // 客户端列表
	// 初始化应该连接的客户端
	shouldConnectedClientList []uint64 //应该连接客户端
	clientConnectedNum int32 // 连接人数
	playerNum int32 //玩的人数
}

func (t *Table) OnInit(uuid string,tableType int32,playerNum int32,roomUuidList []string,shouldConnectedClientList []uint64) {
	t.uuid = uuid
	t.tableType = tableType
	t.clientList = make([]uint64,0)
	t.roomUuidList = roomUuidList
	t.shouldConnectedClientList =shouldConnectedClientList
	t.playerNum = playerNum
}

func (t *Table) GetClientList() []uint64{
	return t.clientList
}

func (t *Table) SetClientList(clientList []uint64) {
	t.clientList = clientList
}

func (t *Table) GetRoomUuidList() []string{
	return t.roomUuidList
}


func (t *Table) SetRoomUuidList(roomUuidList []string)  {
	t.roomUuidList = roomUuidList
}

func (t *Table) GetShouldConnectedClientList() []uint64{
	return t.shouldConnectedClientList
}

func (t *Table) CheckClientCanConnect(clientId uint64) bool  {
	for _, i2 := range t.shouldConnectedClientList {
		if i2 == clientId{
			return true
		}
	}
	return false
}

func (t *Table) GetPlayerNum() int32{
	return t.playerNum
}

func (t *Table) SetPlayerNum(playerNum int32) {
	t.playerNum = playerNum
}

func (t *Table) GetClientConnectedNum() int32{
	return t.clientConnectedNum
}

func (t *Table) SetClientConnectedNum(clientConnectedNum int32) {
	t.clientConnectedNum = clientConnectedNum
}

func (t *Table) GetTableUuid() string  {
	return t.uuid
}