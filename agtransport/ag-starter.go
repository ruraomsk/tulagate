package agtransport

import (
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
)

var (
	CommandARM         chan pudge.CommandARM
	SendCross          chan pudge.UserCross
	workSenderArrays   bool
	workSenderCommand  bool
	internalCommandARM chan pudge.CommandARM
	internalSendCross  chan pudge.UserCross
)

func Starter(next chan interface{}) {
	CommandARM = make(chan pudge.CommandARM, 1000)
	internalCommandARM = make(chan pudge.CommandARM)
	SendCross = make(chan pudge.UserCross, 100)
	internalSendCross = make(chan pudge.UserCross)
	internalCommandARM = make(chan pudge.CommandARM)
	go senderArrays()
	go senderCommand()
	go recieverPhases()
	logger.Info.Print("Транспорт для ag-server готов")
	next <- 1
	for {
		select {
		case sc := <-SendCross:
			if workSenderArrays {
				internalSendCross <- sc
			} else {
				logger.Error.Printf("нет связи с ag-server для %d %d %d", sc.State.Region, sc.State.Area, sc.State.ID)
			}
		case cm := <-CommandARM:
			if workSenderCommand {
				internalCommandARM <- cm
			} else {
				logger.Error.Printf("нет связи с ag-server для %d ", cm.ID)
			}
		}

	}
}
func ReadyAgTransport() bool {
	return workSenderArrays && workSenderCommand
}
