package agtransport

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/tulagate/setup"
)

func senderArrays() {
	workSenderArrays = false
	w := fmt.Sprintf("%s:%d", setup.Set.AgServer.Host, setup.Set.AgServer.PortArray)
	for {
		socket, err := net.Dial("tcp", w)
		if err != nil {
			logger.Error.Printf("connect %s %s ", w, err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		writer := bufio.NewWriter(socket)
		tenSecond := time.NewTicker(10 * time.Second)
		logger.Info.Print("senderArrays ready")
		workSenderArrays = true
	loop:
		for {
			select {
			case <-tenSecond.C:
				writer.WriteString("0\n")
				err := writer.Flush()
				if err != nil {
					logger.Error.Printf("%s %s", socket.RemoteAddr().String(), err.Error())
					break loop
				}
			case cross := <-internalSendCross:
				cross.User = setup.Set.MyName
				buffer, err := json.Marshal(cross)
				if err != nil {
					logger.Error.Printf("%v %s", cross, err.Error())
					break loop
				}
				// logger.Debug.Printf("send cross %d %d %d", cross.State.Region, cross.State.Area, cross.State.ID)
				writer.WriteString(string(buffer))
				writer.WriteString("\n")
				err = writer.Flush()
				if err != nil {
					logger.Error.Printf("%s %s", socket.RemoteAddr().String(), err.Error())
					break loop
				}
			}
		}
		workSenderArrays = false
		socket.Close()
		for range internalSendCross {
		}
	}

}
