package agtransport

import (
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
)

var CommandARM chan pudge.CommandARM
var SendCross chan pudge.UserCross

func Starter() error {
	CommandARM = make(chan pudge.CommandARM, 1000)
	SendCross = make(chan pudge.UserCross, 100)
	go senderArrays()
	go senderCommand()
	go recieverPhases()
	logger.Info.Print("Транспорт для ag-server готов")
	return nil
}