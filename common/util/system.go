package util

import (
	"encoding/json"
	"github.com/duanhf2012/origin/cluster"
	"github.com/duanhf2012/origin/rpc"
	"github.com/golang/protobuf/proto"
	"hash/fnv"
	"io/ioutil"
	"math/rand"
	"strings"
	"sunserver/common/global"
	"time"
	"unsafe"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetMasterCenterNodeId() int {
	var clientList [4]*rpc.Client
	err, num := cluster.GetCluster().GetNodeIdByService(global.CenterService, clientList[:], false)
	if err != nil || num == 0 {
		return 0
	}

	minId := 0
	for i := 0; i < num; i++ {
		if clientList[i] != nil {
			if minId == 0 || clientList[i].GetId() < minId {
				minId = clientList[i].GetId()
			}
		}
	}

	return minId
}

func GetNodeIdByService(name string) int {
	var clientList [4]*rpc.Client
	err, num := cluster.GetCluster().GetNodeIdByService(name, clientList[:], false)
	if err != nil || num == 0 {
		return 0
	}

	minId := 0
	for i := 0; i < num; i++ {
		if clientList[i] != nil {
			if minId == 0 || clientList[i].GetId() < minId {
				minId = clientList[i].GetId()
			}
		}
	}

	return minId
}

//主从centerService，后续实现
func GetSlaveCenterNodeId() int {
	var clientList [4]*rpc.Client
	err, num := cluster.GetCluster().GetNodeIdByService("CenterService", clientList[:], false)
	if err != nil || num == 0 {
		return 0
	}

	minId := 0
	for i := 0; i < num; i++ {
		if clientList[i] != nil {
			if minId == 0 || clientList[i].GetId() < minId {
				minId = clientList[i].GetId()
			}
		}
	}

	return minId
}

func GetBestNodeId(serviceMethod string, key uint64) int {
	var clientList [4]*rpc.Client
	err, num := cluster.GetRpcClient(0, serviceMethod, clientList[:])
	if err != nil || num == 0 {
		return 0
	}

	return clientList[key%uint64(num)].GetId()
}

func RandNum(max int) int {
	r := rand.Intn(max)
	return r
}

func RandNumRange(min, max int64) int64 {
	r := rand.Int63n(max - min)
	return r + min
}

func RandArrayByWeight(array []int) int {
	var sumWeight int
	var weightArray []int

	for _, v := range array {
		sumWeight += v
		weightArray = append(weightArray, sumWeight)
	}

	return RandArrayBySumWeight(weightArray)
}

func RandArrayBySumWeight(array []int) int {
	arraylen := len(array)
	if arraylen <= 0 {
		return -1
	}
	weight := array[arraylen-1]
	r := RandNum(weight)
	for i, v := range array {
		if r < v {
			return i
		}
	}

	return -1
}

func GetDirAllFileName(fileDir string, fileType string) []string {
	files, _ := ioutil.ReadDir(fileDir)

	retFileNameList := make([]string, 0, len(files))
	for _, oneFile := range files {
		if oneFile.IsDir() {
			continue
		}

		if fileType != "" {
			fileName := oneFile.Name()
			vName := strings.Split(fileName, ".")
			if len(vName) <= 0 || strings.ToLower(vName[len(vName)-1]) != strings.ToLower(fileType) {
				continue
			}
		}

		retFileNameList = append(retFileNameList, oneFile.Name())
	}

	return retFileNameList
}

func JSON2PB(formJsonStr string, toPb proto.Message) error {
	// json字符串转pb
	return json.Unmarshal([]byte(formJsonStr), &toPb)
}

func PB2JSON(fromPb proto.Message, toJsonStr *string) error {
	// pb转json字符串
	jsonStr, err := json.Marshal(fromPb)
	if err == nil {
		*toJsonStr = string(jsonStr)
	}

	return err
}

func HashString2Number(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
