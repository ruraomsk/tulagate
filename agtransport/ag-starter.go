package agtransport

import (
	"time"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
)

var CommandARM chan pudge.CommandARM
var SendCross chan pudge.UserCross

func Starter(next chan interface{}) {
	CommandARM = make(chan pudge.CommandARM, 1000)
	SendCross = make(chan pudge.UserCross, 100)
	go senderArrays()
	go senderCommand()
	go recieverPhases()
	logger.Info.Print("Транспорт для ag-server готов")
	next <- 1
	for {
		time.Sleep(time.Second)
	}
}
