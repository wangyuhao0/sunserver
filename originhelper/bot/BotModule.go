package bot

import (
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/event"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/network"
	"github.com/duanhf2012/origin/network/processor"
	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/sysmodule/httpclientmodule"
	"github.com/duanhf2012/origin/util/timer"
	"github.com/golang/protobuf/proto"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"sunserver/common/proto/msg"
	"sync/atomic"
	"time"
)

type BotModule struct {
	service.Module
	LoginUrl string

	httpClient   *httpclientmodule.HttpClientModule
	tcpClient    *network.TCPClient
	tcpConn      *network.TCPConn
	pbProcessor  *processor.PBProcessor
	botEventChan chan BotEvent

	//机器人数据
	data BotData
	//机器人状态
	status StatusType

	mapTicker map[*timer.Ticker]*TestData
	mapTimer  map[int]*TestData
	mapCron   map[*timer.Cron]*TestData

	genId int
}

func (botModule *BotModule) OnInit() error {
	botModule.data.BotName = fmt.Sprintf("robot%d", botModule.GetModuleId())
	botModule.botEventChan = make(chan BotEvent, 50)

	botModule.LoginUrl = "http://127.0.0.1:9101/login"
	botModule.httpClient = &httpclientmodule.HttpClientModule{}
	botModule.httpClient.Init(1, "")
	botModule.Start()
	botModule.SetBotStatus(LoginingHttpGate)
	botModule.pbProcessor = processor.NewPBProcessor()

	//消息注册
	botModule.pbProcessor.Register(uint16(msg.MsgType_LoginRes), &msg.MsgLoginRes{}, botModule.HandlerLoginRes)
	botModule.pbProcessor.Register(uint16(msg.MsgType_LoadFinish), &msg.MsgNil{}, botModule.HandlerLoginFinish)
	botModule.pbProcessor.Register(uint16(msg.MsgType_ClientSyncTimeRes), &msg.MsgClientSyncTimeRes{}, botModule.HandlerTestRes)
	botModule.pbProcessor.Register(uint16(msg.MsgType_Pong), &msg.MsgNil{}, botModule.HandlerPong)
	botModule.pbProcessor.Register(uint16(msg.MsgType_CreateRoomRes), &msg.MsgCreateRoomRes{}, botModule.HandlerCreateRoomRes)
	botModule.pbProcessor.Register(uint16(msg.MsgType_AddRoomRes), &msg.MsgAddRoomRes{}, botModule.HandlerAddRoomRes)
	botModule.pbProcessor.Register(uint16(msg.MsgType_RadioOtherAddRoomRes), &msg.MsgClientOnRoomRes{}, botModule.HandlerRadioRoomRes)
	botModule.pbProcessor.Register(uint16(msg.MsgType_AddQueueRes), &msg.MsgAddQueueRes{}, botModule.HandlerAddQueueRes)
	botModule.pbProcessor.Register(uint16(msg.MsgType_QuitQueueRes), &msg.MsgQuitQueueRes{}, botModule.HandlerQuitQueueRes)
	botModule.pbProcessor.Register(uint16(msg.MsgType_MatchRes), &msg.MsgMatchRes{}, botModule.HandlerMatchRes)
	botModule.pbProcessor.Register(uint16(msg.MsgType_AddTableRes), &msg.MsgAddTableRes{}, botModule.HandlerAddTableRes)
	botModule.pbProcessor.Register(uint16(msg.MsgType_ClientConnectedStatus), &msg.MsgClientConnectedStatusRes{}, botModule.HandleClientConnectStatusRes)

	return nil
}

func (botModule *BotModule) Start() {
	go botModule.RunBot()
}

func (botModule *BotModule) GetBotStatus() StatusType {
	return atomic.LoadInt32(&botModule.status)
}

func (botModule *BotModule) SetBotStatus(value StatusType) {
	atomic.StoreInt32(&botModule.status, value)
}

func (botModule *BotModule) PutBotEvent(event BotEvent) error {
	if len(botModule.botEventChan) > MaxBotEventChan {
		return fmt.Errorf("bot[%d] event chan is full", botModule.GetModuleId())
	}
	botModule.botEventChan <- event
	return nil
}

func (botModule *BotModule) RunBot() {
	log.Release("5分钟的定时器")
	t := timer.NewTimer(5 * time.Second)
	for {
		if botModule.GetBotStatus() == StopRobot {
			break
		}

		select {
		case botEvent := <-botModule.botEventChan:
			botModule.dealBotEvent(botEvent)
		case timeC := <-t.C:
			botModule.ping()
			log.Release("现在触发ping")
			timeC.SetupTimer(time.Now())
		default:
			botModule.checkRunStatus()
		}
		time.Sleep(20 * time.Millisecond)
	}

	//循环结束,表示机器人停止
	t.Cancel()
	timer.ReleaseTimer(t)
	botModule.tcpClient.Close(false)
	botModule.GetEventHandler().NotifyEvent(&event.Event{Type: Cmd_Event_BotStop, Data: botModule.GetModuleId()})
}

func (botModule *BotModule) Run() {
	botModule.PutBotEvent(BotEvent{Type: Bot_Event_Connected, Data: nil, CallBack: nil})
	for {
		bytes, err := botModule.tcpConn.ReadMsg()
		if err != nil {
			botModule.tcpConn = nil
			fmt.Printf("read client is error:%+v\n", err)
			break
		}

		botModule.PutBotEvent(BotEvent{Type: Bot_Event_ReceiveMsg, Data: bytes, CallBack: nil})
	}
}

func (botModule *BotModule) OnClose() {
	fmt.Printf("bot[%s] tcp close\n", botModule.data.BotName)
	botModule.tcpClient.Close(false)
	botModule.PutBotEvent(BotEvent{Type: Bot_Event_Disconnected, Data: nil, CallBack: nil})
}

func (botModule *BotModule) NewClientAgent(conn *network.TCPConn) network.Agent {
	botModule.tcpConn = conn
	return botModule
}

func (botModule *BotModule) SendMsg(msgType msg.MsgType, message proto.Message) error {
	pbPackInfo := botModule.pbProcessor.MakeMsg(uint16(msgType), message)
	bData, err := botModule.pbProcessor.Marshal(pbPackInfo)
	if err != nil {
		log.Error("botModule[%s] userid[%d]  send msg err:%+v", botModule.data.BotName, botModule.data.UserID, err)
		return err
	}

	if botModule.tcpConn == nil {
		log.Warning("botModule[%s] userid[%d] disconnect", botModule.data.BotName, botModule.data.UserID)
		return nil
	}

	err = botModule.tcpConn.WriteMsg(bData)
	if err != nil {
		log.Error("")
	}
	return err
}

func (botModule *BotModule) dealBotEvent(event BotEvent) {
	switch event.Type {
	case Bot_Event_Connected:
		botModule.SetBotStatus(ConnectedGameGate)
	case Bot_Event_ReceiveMsg:
		botModule.receiveTcpMsg(event.Data)
	case Bot_Event_ReceiveCmd:
		dataList := event.Data.([]string)
		event.CallBack(botModule, dataList...)
	case Bot_Event_Disconnected:
		if botModule.GetBotStatus() != StopRobot {
			botModule.SetBotStatus(LoginingHttpGate)
		}
	}
}

func (botModule *BotModule) checkRunStatus() {
	switch botModule.GetBotStatus() {
	case LoginingHttpGate:
		log.Release("bot-loginHttpGate")
		botModule.loginHttpGate()
		time.Sleep(1 * time.Second)
	case LoginedHttpGate:
		log.Release("bot-connectGameGate")
		botModule.connectGameGate()
	case ConnectedGameGate:
		log.Release("ConnectedGameGate-connectGameGate")
		botModule.loginGameGate()
	}
}

func (botModule *BotModule) loginHttpGate() {
	loginBody := fmt.Sprintf(`{"PlatType":1, "PlatId":"%s", "AccessToken":"%s","Account":"%s","PassWord":"%s"}`, botModule.data.BotName, botModule.data.BotName, botModule.data.BotName, botModule.data.BotName)

	response := botModule.httpClient.Request(http.MethodPost, botModule.LoginUrl, []byte(loginBody), nil)
	if response.Err != nil {
		fmt.Printf("robot[%s], loginHttpGate err:%+v\n", botModule.data.BotName, response.Err)
		return
	}
	if strings.Index(response.Status, "200") == -1 {
		fmt.Printf("robot[%s], loginHttpGate err:%s\n", botModule.data.BotName, response.Status)
		return
	}
	logHttpGateRet := LoginHttpGateData{}
	err := json.Unmarshal(response.Body, &logHttpGateRet)
	if err != nil {
		fmt.Printf("robot[%s], loginHttpGate err:%+v\n", botModule.data.BotName, err)
		return
	}

	randServer := logHttpGateRet.randServerByWeight()
	if randServer == nil {
		fmt.Printf("robot[%s], loginHttpGate retData[%+v] err:rand server is nil \n", botModule.data.BotName, err)
		return
	}
	botModule.data.TcpGateUrl = randServer.Url
	botModule.data.UserID = logHttpGateRet.UserId
	botModule.data.Token = logHttpGateRet.Token

	botModule.SetBotStatus(LoginedHttpGate)
}

func (botModule *BotModule) connectGameGate() {
	botModule.SetBotStatus(ConnectingGameGate)

	if botModule.tcpClient == nil {
		botModule.tcpClient = &network.TCPClient{}
	}
	botModule.tcpClient.Addr = botModule.data.TcpGateUrl
	botModule.tcpClient.ConnNum = 1
	botModule.tcpClient.ConnectInterval = time.Second * 2
	botModule.tcpClient.PendingWriteNum = 1000
	botModule.tcpClient.AutoReconnect = true
	botModule.tcpClient.LenMsgLen = 2
	botModule.tcpClient.MinMsgLen = 2
	botModule.tcpClient.MaxMsgLen = math.MaxUint16
	botModule.tcpClient.NewAgent = botModule.NewClientAgent
	botModule.tcpClient.LittleEndian = false
	botModule.tcpClient.Start()
}

func (botModule *BotModule) loginGameGate() {
	msgLoginReq := msg.MsgLoginReq{}
	msgLoginReq.UserId = botModule.data.UserID
	msgLoginReq.Token = botModule.data.Token

	err := botModule.SendMsg(msg.MsgType_LoginReq, &msgLoginReq)
	if err == nil {
		botModule.SetBotStatus(LoginingGameGate)
	} else {
		fmt.Printf("error....\n")
	}
}

func (botModule *BotModule) CreateRoom(clientId uint64) {
	log.Release("bot 创建房间--%d", clientId)
	playerInfo := msg.PlayerInfo{UserId: clientId, Rank: uint64(100), NickName: "王宇豪" + string(clientId), Sex: 1, Avatar: "123", ClientId: clientId, IsOwner: true, SeatNum: int32(0)}
	roomReq := msg.MsgCreateRoomReq{PlayerInfo: &playerInfo, RoomType: int32(2)}
	botModule.SendMsg(msg.MsgType_CreateRoomReq, &roomReq)
}

func (botModule *BotModule) AddRoom(clientId uint64, roomUuid string) {
	log.Release("bot 加入房间--%d,房间号--%d", clientId, roomUuid)

	playerInfo := msg.PlayerInfo{UserId: clientId, Rank: uint64(100), NickName: "王宇豪" + string(clientId), Sex: 1, Avatar: "123", ClientId: clientId, IsOwner: false, SeatNum: int32(0)}
	req := msg.MsgAddRoomReq{RoomUuid: roomUuid, PlayerInfo: &playerInfo}
	botModule.SendMsg(msg.MsgType_AddRoomReq, &req)
}

func (botModule *BotModule) QuitRoom(userId uint64, roomUuid string) {
	log.Release("bot 退出房间--%d,房间号--%d", userId, roomUuid)
	req := msg.MsgQuitRoomReq{RoomUuid: roomUuid}
	botModule.SendMsg(msg.MsgType_QuitRoomReq, &req)
}

func (botModule *BotModule) AddQueue(userId uint64, roomUuid string) {
	log.Release("用户 加入队列--%d,房间号--%s", userId, roomUuid)
	req := msg.MsgAddQueueReq{RoomUuid: roomUuid, UserId: userId}
	botModule.SendMsg(msg.MsgType_AddQueueReq, &req)
}

func (botModule *BotModule) QuitQueue(userId uint64, roomUuid string) {
	log.Release("用户 退出队列--%d,房间号--%s", userId, roomUuid)
	req := msg.MsgQuitQueueReq{RoomUuid: roomUuid, UserId: userId}
	botModule.SendMsg(msg.MsgType_QuitQueueReq, &req)
}

func (botModule *BotModule) AddTable(userId uint64, tableUuid string) {
	log.Release("用户 加入对局--%d,对局号--%s", userId, tableUuid)
	req := msg.MsgAddTableReq{TableUuid: tableUuid, UserId: userId}
	botModule.SendMsg(msg.MsgType_AddTableReq, &req)
}

func (botModule *BotModule) receiveTcpMsg(data interface{}) {
	pbPackage, err := botModule.pbProcessor.Unmarshal(data.([]byte))
	if err != nil {
		fmt.Printf("receiveTcpMsg is error:%+v\n", err)
		return
	}

	botModule.pbProcessor.MsgRoute(pbPackage, uint64(botModule.GetModuleId()))
}

func (botModule *BotModule) ping() {
	if botModule.status < LoginedGameGate {
		return
	}

	botModule.SendMsg(msg.MsgType_Ping, nil)
}

type TestData struct {
	d         time.Duration
	count     int
	trigger   time.Time
	startTime time.Time

	timer   *timer.Timer
	strCron string
}

var timerNum int
var timerDur time.Duration

var tickerNum int
var tickerDur time.Duration

var cronNum int
var cronDur time.Duration

var lastTm time.Time = time.Now()

func Add(typ int, dur time.Duration) {
	if typ == 0 {
		timerNum++
		timerDur += dur
	} else if typ == 1 {
		tickerNum++
		tickerDur += dur
	} else {
		cronNum++
		cronDur += dur
	}
	diffTm := time.Now().Sub(lastTm)
	if diffTm > 5*time.Second && timerNum > 0 && tickerNum > 0 && cronNum > 0 {
		fmt.Printf("diff:%dms->timer[%d,%dms] ticker[%d,%dms] cron[%d,%dms]\n", diffTm.Milliseconds(), timerNum, time.Duration(int64(timerDur)/int64(timerNum)).Milliseconds(),
			tickerNum, time.Duration(int64(tickerDur)/int64(tickerNum)).Milliseconds(),
			cronNum, time.Duration(int64(cronDur)/int64(cronNum)).Milliseconds())
		lastTm = time.Now()
		timerNum = 0
		timerDur = 0
		tickerNum = 0
		tickerDur = 0
		cronNum = 0
		cronDur = 0
	}

}

func (botModule *BotModule) TestTimer() {

	rand.Seed(time.Now().UnixNano())
	//1.Ticker测试
	if botModule.mapTicker == nil {
		botModule.mapTicker = make(map[*timer.Ticker]*TestData)
	}

	if botModule.mapTimer == nil {
		botModule.mapTimer = make(map[int]*TestData)
	}

	if botModule.mapCron == nil {
		botModule.mapCron = make(map[*timer.Cron]*TestData)
	}
	return
	//安装ticker,3600*2内的测试
	for i := 1; i <= 200; i++ {
		dur := time.Duration(rand.Intn(60*1000)) * time.Millisecond //生成0-99随机整数
		ticker := botModule.NewTicker(dur, func(ticker *timer.Ticker) {
			v, ok := botModule.mapTicker[ticker]
			if ok == false {
				log.Error("Error....xxxxxxxxxxxx")
				return
			}

			//fmt.Printf("trigger:%+v\n",v.trigger)
			diff := timer.Now().Sub(v.trigger) - dur
			Add(1, diff)
			v.count += 1
			v.trigger = time.Now()
		})

		botModule.mapTicker[ticker] = &TestData{d: dur, startTime: time.Now(), trigger: time.Now()}
	}

	//安装timer,3600*2内的测试
	for i := 1; i <= 200; i++ {
		dur := time.Duration(rand.Intn(60*1000)) * time.Millisecond //生成0-99随机整数
		fmt.Print(dur, "\n")
		botModule.genId++
		index := botModule.genId
		botModule.TimerCB(dur, index)
	}

	//测试cron,
	for i := 1; i <= 200; i++ {
		str := "* * * * * *"
		pCron, err := timer.NewCronExpr(str)
		if err != nil {
			log.Error("Error....pppppp")
			return
		}

		//开始定时器
		crons := botModule.CronFunc(pCron, func(cron *timer.Cron) {
			v, ok := botModule.mapCron[cron]
			if ok == false {
				log.Error("Error....qqqqqqq")
				return
			}

			diff := timer.Now().Sub(v.trigger) - time.Second
			Add(2, diff)
			v.count += 1
			v.trigger = time.Now()
			//v.trigger = time.Now()
		})
		botModule.mapCron[crons] = &TestData{startTime: time.Now(), strCron: str, trigger: time.Now()}
	}

	fmt.Printf("test bot finish.\n")
}

func (botModule *BotModule) StatTimerInfo() {
	//fmt.Printf("start check timer....\n")
	for k, v := range botModule.mapTimer {
		sub := time.Now().Sub(v.trigger)
		if sub > v.d {
			fmt.Printf("timer,id:%d,sub:%d\n", k, (sub - v.d).Milliseconds())
		}
	}

	//	fmt.Printf("start check ticker....\n")
	for _, v := range botModule.mapTicker {
		sub := time.Now().Sub(v.trigger)
		if sub > v.d {
			fmt.Printf("ticker,sub:%d\n", (sub - v.d).Milliseconds())
		}
	}

	//fmt.Printf("start check cron....\n")

	for _, v := range botModule.mapCron {
		sub := time.Now().Sub(v.trigger)
		if sub > time.Minute {
			fmt.Printf("ticker,sub:%d\n", (sub - time.Second).Milliseconds())
		}
	}

	//fmt.Printf("finish check.....\n")
	/*
		mapTicker map[*timer.Ticker] *TestData
		mapTimer map[int] *TestData
		mapCron map[*timer.Cron] *TestData
	*/

}

func (botModule *BotModule) TimerCB(dur time.Duration, index int) {
	tm := botModule.AfterFunc(dur, func(timer2 *timer.Timer) {
		v, ok := botModule.mapTimer[index]
		if ok == false {
			log.Error("Error....yyyyyyyyy")
			return
		}
		v.count += 1
		diff := timer.Now().Sub(v.trigger) - dur
		Add(0, diff)
		v.timer.SetupTimer(time.Now())
		v.trigger = time.Now()
	})

	botModule.mapTimer[index] = &TestData{d: dur, timer: tm, startTime: time.Now(), trigger: time.Now()}
}
