package def

import (
	"container/list"
	"fmt"
	"sunserver/gamecore/roomservice/room"
)

// 为了 队列使用 满足查找以及删除
//用锁 防止并发

type Keyer interface {
	GetKey() string
	GetValue() string
}

type MapList struct {
	dataMap  map[string]*list.Element
	dataList *list.List
}

func (mapList *MapList) GetDataList() *list.List {
	return mapList.dataList
}

func NewMapList() *MapList {
	return &MapList{
		dataMap:  make(map[string]*list.Element),
		dataList: list.New(),
	}
}

func (mapList *MapList) Exists(roomUuid string) bool {
	_, exists := mapList.dataMap[roomUuid]
	return exists
}

func (mapList *MapList) Push(roomUuid string, room *room.Room) bool {
	if mapList.Exists(roomUuid) {
		return false
	}
	elem := mapList.dataList.PushBack(room)
	mapList.dataMap[roomUuid] = elem
	return true
}

func (mapList *MapList) Remove(roomUuid string) {
	if !mapList.Exists(roomUuid) {
		return
	}
	mapList.dataList.Remove(mapList.dataMap[roomUuid])
	delete(mapList.dataMap, roomUuid)
}

func (mapList *MapList) Size() int {
	return mapList.dataList.Len()
}

func (mapList *MapList) Walk() {
	for elem := mapList.dataList.Front(); elem != nil; elem = elem.Next() {
		fmt.Print(elem.Value)
	}
}

func (mapList *MapList) GetRoomByIndex(index int) *room.Room {
	i := 0
	for elem := mapList.dataList.Front(); elem != nil; elem = elem.Next() {
		if i == index {
			return elem.Value.(*room.Room)
		}
		i++
	}
	return nil
}

type Elements struct {
	key   string
	value *room.Room
}

func (e Elements) GetKey() string {
	return e.key
}

func (e Elements) GetValue() *room.Room {
	return e.value
}

/*type MapList struct {
	DataMap  sync.Map
	DataList []*common.Room
}


func NewMapList() *MapList {

	var dataMap sync.Map
	return &MapList{
		DataMap:  dataMap,
		DataList: make([]*common.Room,0),
	}
}*/

/*func (mapList *MapList) Exists(roomUuid string) bool {
	_, exists := mapList.DataMap.Load(roomUuid)
	return exists
}

func (mapList *MapList) Push(roomUuid string,room *common.Room) bool {
	if mapList.Exists(roomUuid) {
		return false
	}
	mapList.DataList = append(mapList.DataList,room)
	mapList.DataMap.Store(roomUuid,len(mapList.DataList)-1)
	return true
}

func (mapList *MapList) Remove(roomUuid string) {
	load, ok := mapList.DataMap.Load(roomUuid)
	if !ok {
		return
	}
	index := load.(int)
	mapList.DataList = append(mapList.DataList[:index],mapList.DataList[index+1:]...)
	mapList.DataMap.Delete(roomUuid)
}

func (mapList *MapList) GetRoomByIndex(index int) *common.Room{
	return mapList.DataList[index]
}

/*func (mapList *MapList) RemoveRoomByIndex(index int) {
	mapList.DataList = append(mapList.DataList[:index],mapList.DataList[index+1:]...)
}*/

/*func (mapList *MapList) Size() int {
	return len(mapList.DataList)
}

type Elements struct {
	value string
}

func (e Elements) GetKey() string {
	return e.value
}

*/
