package originhelper

import (
	"fmt"
	"github.com/duanhf2012/origin/cluster"
	"github.com/duanhf2012/origin/sysmodule/mongomodule"
	"io/ioutil"
	"strconv"
	"strings"
	"sunserver/common/collect"
	"sunserver/common/proto/rpc"
	"sunserver/originhelper/bot"
	"sunserver/originhelper/logic"
	"time"
)

func (os *OriginHelperService) OnRegCommand(){
	//系统全局功能
	os.RegCmd("execfile",1,"<execfile file> Execute command script file.",logic.ExecFile)
	os.RegCmd("syncindex",1,"<syncindex mongourl> Synchironize database index.",logic.SyncIndex)

	//功能测试
	//os.RegCmd("test2",100,"this is a test2",os.Test1)

	//配置命令
	os.RegCmd("reloadcfg", 200, "<reloadcfg nodeID serviceName fileNameList> notice nodeID.serviceName reload fileNameList", logic.ReLoadCfg)

	//压力测试
	os.RegCmd("testRpc", 1000, "<testRpc maxunm>", logic.TestCallRPC)
	os.RegCmd("addbot",1001,"<addbot startid endid> Add",logic.AddBot)

	//os.RegCmd("testtimer",1000,"<testtimer startid endid> Add",os.testTimer)
	os.RegCmd("botstatus", 1002, "<botstatus startid endid> startid~endid robots's status statistics", logic.BotStatus)

	os.RegBotCmd("stopbot", 1003, "<stopbot startid endid> startid~endid robots stop", logic.BotStop)
	os.RegBotCmd("botsendmsg", 1004, "<botsendmsg startid endid msgid msgstruct> startid~endid robots send msg to server", logic.BotSendMsg)

	os.RegCmd("testtimer",1000,"<testtimer startid endid> test timer",logic.TestTimer)
	os.RegCmd("st",1000,"<stattimer startid endid> stat timer info",logic.StatTimerInfo)

	//针对玩家命令
	os.RegCmd("sendmail", 2000, "<sendmail userid title content attachment> send a mail to user.", logic.SendMailToUser)

	os.RegCmd("createRoom",2001,"<createRoom botId> ",logic.CreateRoomBot)
	os.RegCmd("addRoom",2002,"<addRoom botId roomUuid> ",logic.AddRoomBot)
	os.RegCmd("quitRoom",2003,"<quitRoom botId roomUuid> ",logic.QuitRoomBot)
	os.RegCmd("addQueue",2004,"<addQueue botId roomUuid> ",logic.AddQueueBot)
	os.RegCmd("quitQueue",2005,"<quitQueue botId roomUuid> ",logic.QuitQueueBot)
	os.RegCmd("addTable",2006,"<addTable botId tableUuid> ",logic.AddTableBot)



}

func (os *OriginHelperService) useNode(args... string) error{
	if len(args)!= 1 {
		return fmt.Errorf("First parameter error.")
	}

	nodeId,err := strconv.Atoi(args[0])
	if err != nil || nodeId<=0 {
		return fmt.Errorf("First parameter error.")
	}

	bConnect := cluster.GetCluster().IsNodeConnected(nodeId)
	if bConnect == false {
		return fmt.Errorf("Cannot switch to nodeid %d.",nodeId)
	}

	os.useNodeId = nodeId
	return nil
}

func (os *OriginHelperService) Test1(args... string) error{
	return nil
}

func (os *OriginHelperService) reLoadCfg(args ...string) error {
	if len(args) != 2{
		return fmt.Errorf("Invalid parameter.")
	}

	serviceName := args[1]
	nodeID, err := strconv.Atoi(args[0])
	if err != nil || nodeID <= 0 {
		return fmt.Errorf("The first parameter is error.")
	}

	noticeInfo := rpc.PlaceHolders{}
	resultInfo := rpc.ReloadCfgResult{}
	callFunc := serviceName + ".RPC_ReLoadCfg"
	err = os.CallNode(nodeID, callFunc, &noticeInfo, &resultInfo)

	return err
}

func (os *OriginHelperService) execFile(args... string) error{
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
		args []string
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
		if os.HashCmd(cmd) == false {
			return fmt.Errorf("file line number %d is error.",lineNum)
		}
		cline.cmd = cmdList[0]
		cline.args =  cmdList[1:]

		cmdLineList = append(cmdLineList,cline)
	}
	if len(cmdLineList) == 0 {
		return fmt.Errorf("There are no commands to execute.")
	}

	//开始执行命令
	for idx,l := range cmdLineList {
		fmt.Printf("Execute cmd %-15s[%d%%].\n",l.cmd,(idx+1)*100/len(cmdLineList))
		for i:=0;i<l.execRepeatNum;i++{
			err:= os.ExecCmd(l.cmd,l.args...)
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

	fmt.Printf("Execute cmd finish!\n")
	return nil
}

func (os *OriginHelperService) addBot(args ...string) error{
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
		if os.GetModule(int64(i))!= nil {
			return fmt.Errorf("Robot %d already exists.",i)
		}
	}

	//添加机器人
	for i:=startBotId;i<=endBotId;i++{
		bot := &bot.BotModule{}
		bot.SetModuleId(int64(i))
		os.AddModule(bot)
	}

	fmt.Printf("Add robot complete[%d-%d]!\n",startBotId,endBotId)
	return nil
}

//syncindex 1 mongodb://admin:123456@127.0.0.1:27017/dbname
func (os *OriginHelperService) syncIndex(args ...string) error{
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
