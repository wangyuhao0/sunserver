package originhelper

import (
	"fmt"
	"github.com/duanhf2012/origin/event"
	"github.com/duanhf2012/origin/node"
	"github.com/duanhf2012/origin/service"
	"sunserver/originhelper/bot"
)

type OriginHelperService struct {
	service.Service
	ICmdProcessor

	useNodeId int
}

func init(){
	node.Setup(&OriginHelperService{})
}

func (os *OriginHelperService) OnInit() error {
	processor := &CmdProcessor{}
	os.ICmdProcessor = processor
	os.AddModule(processor)
	os.OnRegCommand()

	//机器人事件注册
	os.RegEventReceiverFunc(bot.Cmd_Event_BotStop, os.GetEventHandler(), os.BotEventStop)

	return nil
}

func (os *OriginHelperService) GetEventType() event.EventType{
	return os.GetEventType()
}

func (os *OriginHelperService) BotEventStop(ev event.IEvent) {
	botID := ev.(*event.Event).Data.(int64)
	os.ReleaseModule(botID)
	fmt.Printf("stop robot[%d]\n", botID)
}
