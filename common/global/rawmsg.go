package global

import (
	"encoding/binary"
	"errors"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/network"
)

var tcpMemPool network.INetMempool = network.NewMemAreaPool()

type RawInputArgs struct {
	rawData      []byte //源数据
	msgType      uint16
	clientIdList []uint64
	msg          []byte
	protoMsg     interface{}
}

func (inputArgs *RawInputArgs) GetRawData() []byte {
	return inputArgs.rawData
}

func (inputArgs *RawInputArgs) GetMsgType() uint16 {
	return inputArgs.msgType
}

func (inputArgs *RawInputArgs) GetClientIdList() []uint64 {
	return inputArgs.clientIdList
}

func (inputArgs *RawInputArgs) GetMsg() []byte {
	return inputArgs.msg
}

func (inputArgs *RawInputArgs) SetProtoMsg(message interface{}){
	inputArgs.protoMsg = message
}

func (inputArgs *RawInputArgs) GetProtoMsg() interface{}{
	return inputArgs.protoMsg
}

func (inputArgs *RawInputArgs) DoFree() {

}

func (inputArgs *RawInputArgs) DoEscape() {

}

func (inputArgs *RawInputArgs) DoGc() {
	if len(inputArgs.rawData) < 2 {
		return
	}

	tcpMemPool.ReleaseByteSlice(inputArgs.rawData)
}

/*
func (inputArgs *RawInputArgs) MakeByteSlice(size int) []byte{
	inputArgs.rawData = tcpMemPool.MakeByteSlice(size)
	return inputArgs.rawData
}
*/
func (inputArgs *RawInputArgs) SetUint64(value uint64) []byte {
	inputArgs.rawData = tcpMemPool.MakeByteSlice(8)
	binary.BigEndian.PutUint64(inputArgs.rawData, value)

	return inputArgs.rawData
}

func (inputArgs *RawInputArgs) SetMsgHead(msgType uint16, clientId uint64) []byte {
	inputArgs.rawData = tcpMemPool.MakeByteSlice(10)
	binary.BigEndian.PutUint16(inputArgs.rawData, msgType)
	binary.BigEndian.PutUint64(inputArgs.rawData[2:], clientId)

	return inputArgs.rawData
}

func (inputArgs *RawInputArgs) SetMsg(msgType uint16, clientId uint64, msg []byte) []byte {
	inputArgs.rawData = tcpMemPool.MakeByteSlice(10 + len(msg))
	binary.BigEndian.PutUint16(inputArgs.rawData, msgType)
	binary.BigEndian.PutUint64(inputArgs.rawData[2:], clientId)
	copy(inputArgs.rawData[10:], msg)

	return inputArgs.rawData
}

func (inputArgs *RawInputArgs) ParseMsg(rawMsg []byte) error {
	if len(rawMsg) < 10 {
		err := errors.New("parseMsg error")
		log.Error(err.Error())
		return err
	}

	msgType := binary.BigEndian.Uint16(rawMsg)
	clientId := binary.BigEndian.Uint64(rawMsg[2:])
	msg := rawMsg[10:]

	inputArgs.msgType = msgType
	inputArgs.msg = msg
	inputArgs.clientIdList = []uint64{clientId}
	inputArgs.rawData = rawMsg
	return nil
}

func (inputArgs *RawInputArgs) ParseMsgHead(rawMsg []byte) (msgType uint16, clientId uint64, err error) {
	if len(rawMsg) < 10 {
		err = errors.New("parseMsg error")
		log.Error(err.Error())
		return 0, 0, err
	}
	msgType = binary.BigEndian.Uint16(rawMsg)
	clientId = binary.BigEndian.Uint64(rawMsg[2:])

	return
}

func (inputArgs *RawInputArgs) ParseUint64(rawMsg []byte) (val uint64, err error) {
	if len(rawMsg) < 10 {
		err = errors.New("parseMsg error")
		log.Error(err.Error())
		return 0, err
	}
	val = binary.BigEndian.Uint64(rawMsg)

	return
}
