package logic

import (
	"fmt"
	"strconv"
	"sunserver/originhelper/bot"
)

//testtimer 1 1
func TestTimer(helper IHelper,args ...string) error{
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

	//先检查
	for i:=startBotId;i<=endBotId;i++{
		pModule := helper.GetModule(int64(i))
		if pModule == nil {
			return fmt.Errorf("Cannot find bot[%d].", i)
		}
	}

	for i:=startBotId;i<=endBotId;i++{
		pModule := helper.GetModule(int64(i)).(*bot.BotModule)
		if pModule == nil {
			return fmt.Errorf("Cannot find bot[%d].", i)
		}

		pModule.TestTimer()
	}

	fmt.Printf("")
	return nil
}

func StatTimerInfo(helper IHelper,args ...string) error{
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

	//先检查
	for i:=startBotId;i<=endBotId;i++{
		pModule := helper.GetModule(int64(i))
		if pModule == nil {
			return fmt.Errorf("Cannot find bot[%d].", i)
		}
	}

	for i:=startBotId;i<=endBotId;i++{
		pModule := helper.GetModule(int64(i)).(*bot.BotModule)
		if pModule == nil {
			return fmt.Errorf("Cannot find bot[%d].", i)
		}

		pModule.StatTimerInfo()
	}

	return nil
}