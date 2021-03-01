package originhelper

import (
	"bufio"
	"fmt"
	"github.com/duanhf2012/origin/event"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/util/timer"
	"os"
	"sort"
	"strconv"
	"strings"
	"sunserver/originhelper/bot"
	"sunserver/originhelper/logic"
	"time"
)

type CmdCB func(helper logic.IHelper, param ...string) error
type ICmdProcessor interface {
	RegCmd(command string, sortValue int, description string, cmdCB CmdCB)
	RegBotCmd(command string, sortValue int, description string, cmdCB bot.BotCmdCB)
	ExecCmd(cmd string, args ...string) error
	HashCmd(command string) bool
}

type CmdInfo struct {
	cmd   string
	desc  string
	cb    CmdCB
	cbBot bot.BotCmdCB
	sort  int
}

type CmdInfoSlice []CmdInfo
type CmdProcessor struct {
	service.Module

	mapRegCmd map[string]CmdInfo
}

func (cp *CmdProcessor) OnInit() error {

	cp.GetEventProcessor().RegEventReciverFunc(bot.Cmd_Event_Input, cp.GetEventHandler(), cp.DealCmd)

	cp.mapRegCmd = make(map[string]CmdInfo, 50)
	cp.AfterFunc(time.Second*1, cp.Input)
	cp.RegCmd("help", 0, "print this help page", cp.printHelp)
	return nil
}

func (s CmdInfoSlice) Len() int {
	return len(s)
}

func (s CmdInfoSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s CmdInfoSlice) Less(i, j int) bool {
	return s[i].sort < s[j].sort
}

func (cp *CmdProcessor) printWelcome() {

}

func (cp *CmdProcessor) printInput() {
	fmt.Printf(">")
}

func (cp *CmdProcessor) printHelp(helper logic.IHelper, param ...string) error {
	var cmdList []CmdInfo
	for _, v := range cp.mapRegCmd {
		cmdList = append(cmdList, v)
	}
	sort.Sort(CmdInfoSlice(cmdList))

	//
	fmt.Printf("----   Welcome to origin helper   ----\n")
	//fmt.Printf("This is an origin Engine console tool:\n")
	beforeSort := 0
	for idx, _ := range cmdList {
		if cmdList[idx].sort-beforeSort > 1 {
			fmt.Printf("\n")
		}
		beforeSort = cmdList[idx].sort
		fmt.Printf("%-15s%s\n", cmdList[idx].cmd, cmdList[idx].desc)
	}
	return nil
}

func (cp *CmdProcessor) Input(timer *timer.Timer) {
	fmt.Printf("\n")
	cp.printHelp(nil)
	cp.printInput()

	inputReader := bufio.NewReader(os.Stdin)

	go func() {
		for {
			input, err := inputReader.ReadString('\n')
			if err != nil {
				fmt.Printf("%s: Command not found.\n", input)
				cp.printInput()
			}

			input = strings.Trim(input, "\n")
			input = strings.Trim(input, "\r")
			input = strings.Trim(input, " ")
			cp.GetEventHandler().NotifyEvent(&event.Event{Type: bot.Cmd_Event_Input, Data: input})
		}
	}()
}

type HelpEvent struct {
	event *event.Event
}

func (he *HelpEvent) GetEventType() event.EventType {
	return he.event.GetEventType()
}

func (cp *CmdProcessor) DealCmd(ev event.IEvent) {
	cmdInput := ev.(*event.Event).Data.(string)
	if cp.Parse(cmdInput) == false {
		fmt.Printf("%s: Command not found.\n", cmdInput)
		cp.printInput()
	}
	cp.printInput()
}

func (cp *CmdProcessor) RegCmd(command string, sortValue int, description string, cmdCB CmdCB) {
	cp.mapRegCmd[command] = CmdInfo{cmd: command, desc: description, cb: cmdCB, sort: sortValue}
}

func (cp *CmdProcessor) RegBotCmd(command string, sortValue int, description string, cmdCB bot.BotCmdCB) {
	cp.mapRegCmd[command] = CmdInfo{cmd: command, desc: description, cbBot: cmdCB, sort: sortValue}
}

func (cp *CmdProcessor) ExecCmd(cmd string, args ...string) error {
	v, ok := cp.mapRegCmd[cmd]
	if ok == false {
		return fmt.Errorf("%s: Command not found.", cmd)
	}

	if v.cb != nil {
		err := v.cb(cp.GetService().(*OriginHelperService), args...)
		if err != nil {
			return fmt.Errorf("%s: %s\n", cmd, err.Error())
		}
	} else if v.cbBot != nil {
		serviceHelper := cp.GetService().(*OriginHelperService)
		if len(args) < 2 {
			return fmt.Errorf("%s: Parameter must be greater than 2\n", cmd)
		}

		startBotID, err := strconv.Atoi(args[0])
		if err != nil || startBotID <= 0 {
			return fmt.Errorf("The first parameter is error.")
		}

		endBotID, err := strconv.Atoi(args[1])
		if err != nil || endBotID <= 0 {
			return fmt.Errorf("The second parameter[%d] is error:%+v.", endBotID, err)
		}

		for i := startBotID; i <= endBotID; i++ {
			botIModule := serviceHelper.GetModule(int64(i))
			if botIModule == nil {
				log.Warning("%s: no this robot[%d]", cmd, i)
				continue
			}

			cmdEvent := bot.BotEvent{
				Type:     bot.Bot_Event_ReceiveCmd,
				Data:     nil,
				CallBack: v.cbBot,
			}
			cmdEvent.Data = append(make([]string, 0, len(args)-2), args[2:]...)

			errPut := botIModule.(*bot.BotModule).PutBotEvent(cmdEvent)
			if errPut != nil {
				log.Error("%s: put robot[%d] cmd[%+v] err:%+v", cmd, i, &cmdEvent, errPut)
				continue
			}
		}
	} else {
		return fmt.Errorf("%s: no this cmd\n", cmd)
	}

	return nil
}

func (cp *CmdProcessor) Parse(strCmd string) bool {
	cmdList := strings.Split(strCmd, " ")
	if strCmd == "" || len(cmdList) == 0 {
		return true
	}

	//err := v.cb(cp.GetService().(*OriginHelperService),cmdList[1:]...)
	err := cp.ExecCmd(cmdList[0], cmdList[1:]...)
	if err != nil {
		fmt.Printf("%s: %s\n", cmdList[0], err.Error())
	}
	return true
}

func (cp *CmdProcessor) HashCmd(command string) bool {
	_, ok := cp.mapRegCmd[command]
	return ok
}
