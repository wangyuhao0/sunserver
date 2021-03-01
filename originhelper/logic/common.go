package logic

import (
	"fmt"
	"github.com/duanhf2012/origin/event"
	"github.com/duanhf2012/origin/sysmodule/mongomodule"
	"io/ioutil"
	"strconv"
	"strings"
	"sunserver/common/collect"
	"sunserver/common/proto/rpc"
	"sunserver/originhelper/bot"
	"time"
)

func  ExecFile(helper IHelper,args... string) error{
	if len(args)!= 1 {
		return fmt.Errorf("First parameter error.")
	}

	file := strings.Trim(args[0]," ")
	context, err := ioutil.ReadFile(file) // just pass the file name
	if err != nil {
		return fmt.Errorf("Cannot open %s file.",args[0])
	}

	type cmdline struct {
		execRepeatNum int
		perWait time.Duration
		finishWait time.Duration
		cmd string
		//args []string
	}
	funParseUtilTime := func(strTime string,prefix string,duration time.Duration) (d time.Duration,err error){
		pos := strings.LastIndex(strTime,prefix)
		if pos != -1 {
			tm,err := strconv.Atoi(strings.TrimSpace(strTime[:pos]))
			if err != nil {
				return 0,err
			}
			return time.Duration(tm)*duration,nil
		}
		return 0,fmt.Errorf("cannot find %s",prefix)
	}

	funParseTime := func(strTime string)(d time.Duration,err error){
		d,err = funParseUtilTime(strTime,"ns",time.Nanosecond)
		if err== nil {
			return
		}
		d,err = funParseUtilTime(strTime,"us",time.Microsecond)
		if err== nil {
			return
		}

		d,err = funParseUtilTime(strTime,"ms",time.Millisecond)
		if err== nil {
			return
		}

		d,err = funParseUtilTime(strTime,"s",time.Second)
		if err== nil {
			return
		}
		return
	}

	var cmdLineList []cmdline
	lineList := strings.Split(string(context),"\n")
	for lineNum,line := range lineList {
		strLine:=strings.TrimSpace(line)
		strLine=strings.Trim(strLine,"\n")
		strLine=strings.Trim(strLine,"\r\n")
		strLine=strings.TrimSpace(line)
		if len(strLine) == 0 || strLine[0] == '#'{
			continue
		}

		var cline cmdline
		cmdInfo :=  strings.Split(strLine,",")
		if len(cmdInfo)<4{
			return fmt.Errorf("file line number %d is error.",lineNum)
		}

		execRepeatNum,err := strconv.Atoi(strings.TrimSpace(cmdInfo[0]))
		if err != nil || execRepeatNum<=0{
			return fmt.Errorf("file line number %d is error.",lineNum)
		}

		cline.execRepeatNum = execRepeatNum
		d,err := funParseTime(strings.TrimSpace(cmdInfo[1]))
		if err != nil {
			return fmt.Errorf("file line number %d is error.",lineNum)
		}
		cline.perWait = d
		d,err = funParseTime(strings.TrimSpace(cmdInfo[2]))
		if err != nil {
			return fmt.Errorf("file line number %d is error.",lineNum)
		}
		cline.finishWait = d

		strCmd := strings.TrimSpace(cmdInfo[3])
		cmdList := strings.Split(strCmd," ")
		if strCmd=="" || len(cmdList) == 0 {
			return fmt.Errorf("file line number %d is error.",lineNum)
		}

		cmd := strings.Trim(cmdList[0]," ")
		if helper.HashCmd(cmd) == false {
			return fmt.Errorf("file line number %d is error.",lineNum)
		}
		cline.cmd = strCmd
		//cline.cmd = cmdList[0]
		//cline.args =  cmdList[1:]

		cmdLineList = append(cmdLineList,cline)
	}
	if len(cmdLineList) == 0 {
		return fmt.Errorf("There are no commands to execute.")
	}

	//开始执行命令
	go func() {
		for idx,l := range cmdLineList {
			fmt.Printf("Execute cmd %-15s[%d%%].\n",l.cmd,(idx+1)*100/len(cmdLineList))
			for i:=0;i<l.execRepeatNum;i++{
				//err:= helper.ExecCmd(l.cmd,l.args...)
				helper.NotifyEvent(&event.Event{Type:bot.Cmd_Event_Input,Data: l.cmd})
				if l.perWait>0 {
					time.Sleep(l.perWait)
				}
				if err != nil {
					fmt.Printf("%s\n",err.Error())
				}
			}

			if l.finishWait>0 {
				time.Sleep(l.finishWait)
			}
		}
		/*for {
			input, err := inputReader.ReadString('\n')
			if err != nil {
				fmt.Printf("%s: Command not found.\n", input)
				cp.printInput()
			}

			input = strings.Trim(input, "\n")
			input = strings.Trim(input, "\r")
			input = strings.Trim(input, " ")

		}*/
	}()


	fmt.Printf("Execute cmd finish!\n")
	return nil
}



//syncindex 1 mongodb://admin:123456@127.0.0.1:27017/dbname
func SyncIndex(helper IHelper,args ...string) error{
	if len(args)!=2{
		return fmt.Errorf("Invalid parameter.")
	}

	strDbType := strings.TrimSpace(args[0])
	dbType,err := strconv.Atoi(strDbType)
	if err != nil || dbType<=0 {
		return fmt.Errorf("The first parameter is error.")
	}

	url := strings.TrimSpace(args[1])
	if url =="" {
		return fmt.Errorf("First parameter error.")
	}

	module := mongomodule.MongoModule{}
	index := strings.LastIndex(url, "/")

	connUrl := url[:index]
	dbName := url[index+1:]
	err = module.Init(connUrl, 1, 5*time.Second, 5*time.Second)
	if err != nil {
		return fmt.Errorf("Connection database fail:%s.", err.Error())
	}

	s := module.Take()
	if s == nil {
		return fmt.Errorf("Connection database fail:%s.", err.Error())
	}
	err = collect.SyncIndex(s,collect.DBType(dbType), dbName)
	if err != nil {
		return fmt.Errorf("SyncIndex fail:%s.", err.Error())
	}

	fmt.Println(" Synchronization index complet")
	return nil
}



func ReLoadCfg(helper IHelper,args ...string) error {
	if len(args) != 3 {
		return fmt.Errorf("Invalid parameter.")
	}

	serviceName := args[1]
	nodeID, err := strconv.Atoi(args[0])
	if err != nil || nodeID <= 0 {
		return fmt.Errorf("The first parameter is error.")
	}

	noticeInfo := rpc.ReloadCfgInfo{}
	resultInfo := rpc.ReloadCfgResult{}
	noticeInfo.FileNameList = strings.Split(args[2], ",")
	if len(noticeInfo.FileNameList) <= 0 {
		return fmt.Errorf("The third parameter is error: no file need reload.")
	}

	callFunc := serviceName + ".RPC_ReLoadCfg"
	err = helper.CallNode(nodeID, callFunc, &noticeInfo, &resultInfo)

	return err
}

func AddBot(helper IHelper,args ...string) error{
	if len(args)!= 2 {
		return fmt.Errorf("Must be 2 parameters.")
	}

	startBotId,err := strconv.Atoi(args[0])
	if err != nil || startBotId<=0 {
		return fmt.Errorf("The first parameter is error.")
	}

	endBotId,err := strconv.Atoi(args[1])
	if err != nil || endBotId<=0 {
		return fmt.Errorf("The second parameter[%d] is error:%+v.", endBotId, err)
	}

	//验证botid是否存在
	for i:=startBotId;i<=endBotId;i++{
		if helper.GetModule(int64(i))!= nil {
			return fmt.Errorf("Robot %d already exists.",i)
		}
	}

	//添加机器人
	for i:=startBotId;i<=endBotId;i++{
		bot := &bot.BotModule{}
		bot.SetModuleId(int64(i))
		helper.AddModule(bot)
	}

	fmt.Printf("Add robot complete[%d-%d]!\n",startBotId,endBotId)
	return nil
}

func CreateRoomBot(helper IHelper, args ...string) error {
	/*if len(args)!= 2 {
		return fmt.Errorf("Must be 2 parameters.")
	}*/
	botId,err := strconv.Atoi(args[0])
	if err != nil || botId<=0 {
		return fmt.Errorf("The first parameter is error.")
	}
	botIModule := helper.GetModule(int64(botId))
	botObj := botIModule.(*bot.BotModule)
	botObj.CreateRoom(uint64(botId))
	return nil
}


func AddRoomBot(helper IHelper, args ...string) error {
	/*if len(args)!= 2 {
		return fmt.Errorf("Must be 2 parameters.")
	}*/
	botId,err := strconv.Atoi(args[0])
	if err != nil || botId<=0 {
		return fmt.Errorf("The first parameter is error.")
	}

	if err != nil || len(args[1])<=0 {
		return fmt.Errorf("The second parameter[%d] is error:%+v.", args[1], err)
	}

	botIModule := helper.GetModule(int64(botId))
	botObj := botIModule.(*bot.BotModule)
	botObj.AddRoom(uint64(botId),string(args[1]))
	return nil
}

func QuitRoomBot(helper IHelper, args ...string) error {
	/*if len(args)!= 2 {
		return fmt.Errorf("Must be 2 parameters.")
	}*/
	botId,err := strconv.Atoi(args[0])
	if err != nil || botId<=0 {
		return fmt.Errorf("The first parameter is error.")
	}

	if err != nil || len(args[1])<=0 {
		return fmt.Errorf("The second parameter[%d] is error:%+v.", args[1], err)
	}

	botIModule := helper.GetModule(int64(botId))
	botObj := botIModule.(*bot.BotModule)
	botObj.QuitRoom(uint64(botId),string(args[1]))
	return nil
}

func AddQueueBot(helper IHelper, args ...string) error {
	/*if len(args)!= 2 {
		return fmt.Errorf("Must be 2 parameters.")
	}*/
	botId,err := strconv.Atoi(args[0])
	if err != nil || botId<=0 {
		return fmt.Errorf("The first parameter is error.")
	}

	if err != nil || len(args[1])<=0 {
		return fmt.Errorf("The second parameter[%d] is error:%+v.", args[1], err)
	}

	botIModule := helper.GetModule(int64(botId))
	botObj := botIModule.(*bot.BotModule)
	botObj.AddQueue(uint64(botId),string(args[1]))
	return nil
}

func QuitQueueBot(helper IHelper, args ...string) error {
	/*if len(args)!= 2 {
		return fmt.Errorf("Must be 2 parameters.")
	}*/
	botId,err := strconv.Atoi(args[0])
	if err != nil || botId<=0 {
		return fmt.Errorf("The first parameter is error.")
	}

	if err != nil || len(args[1])<=0 {
		return fmt.Errorf("The second parameter[%d] is error:%+v.", args[1], err)
	}

	botIModule := helper.GetModule(int64(botId))
	botObj := botIModule.(*bot.BotModule)
	botObj.QuitQueue(uint64(botId),string(args[1]))
	return nil
}

func AddTableBot(helper IHelper, args ...string) error {
	/*if len(args)!= 2 {
		return fmt.Errorf("Must be 2 parameters.")
	}*/
	botId,err := strconv.Atoi(args[0])
	if err != nil || botId<=0 {
		return fmt.Errorf("The first parameter is error.")
	}

	if err != nil || len(args[1])<=0 {
		return fmt.Errorf("The second parameter[%d] is error:%+v.", args[1], err)
	}

	botIModule := helper.GetModule(int64(botId))
	botObj := botIModule.(*bot.BotModule)
	botObj.AddTable(uint64(botId),string(args[1]))
	return nil
}

func BotStatus(helper IHelper, args ...string) error {
	if len(args)!= 2 {
		return fmt.Errorf("Must be 2 parameters.")
	}

	startBotId,err := strconv.Atoi(args[0])
	if err != nil || startBotId<=0 {
		return fmt.Errorf("The first parameter is error.")
	}

	endBotId,err := strconv.Atoi(args[1])
	if err != nil || endBotId<=0 {
		return fmt.Errorf("The second parameter[%d] is error:%+v.", endBotId, err)
	}

	//获取机器人状态
	statusMap := map[int]int{}
	for i := startBotId; i <= endBotId; i++{
		botIModule := helper.GetModule(int64(i))
		if botIModule == nil {
			if _, ok := statusMap[bot.Bot_Status_NO]; ok == false {
				statusMap[bot.Bot_Status_NO] = 0
			}
			statusMap[bot.Bot_Status_NO] += 1
			continue
		}

		botObj := botIModule.(*bot.BotModule)
		botStatus := botObj.GetBotStatus()
		if botStatus < bot.LoginedGameGate {
			if _, ok := statusMap[bot.Bot_Status_Logining]; ok == false {
				statusMap[bot.Bot_Status_Logining] = 0
			}
			statusMap[bot.Bot_Status_Logining] += 1
		} else {
			if _, ok := statusMap[bot.Bot_Status_Logined]; ok == false {
				statusMap[bot.Bot_Status_Logined] = 0
			}
			statusMap[bot.Bot_Status_Logined] += 1
		}
	}

	//输出
	printStr := ""
	for status, number := range statusMap {
		switch status {
		case bot.Bot_Status_NO:
			printStr = fmt.Sprintf("%s do not add robot num:%d\n", printStr, number)
		case bot.Bot_Status_Logining:
			printStr = fmt.Sprintf("%s landing robot num:%d\n", printStr, number)
		case bot.Bot_Status_Logined:
			printStr = fmt.Sprintf("%s login complete robot num:%d\n", printStr, number)
		default:
			continue
		}
	}

	fmt.Print(printStr)
	return nil
}