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

func senderCommand() {
	w := fmt.Sprintf("%s:%d", setup.Set.AgServer.Host, setup.Set.AgServer.PortCommand)
	for {
		socket, err := net.Dial("tcp", w)
		if err != nil {
			logger.Error.Printf("connect %s %s ", w, err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		writer := bufio.NewWriter(socket)
		tenSecond := time.NewTicker(10 * time.Second)
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
			case cmd := <-CommandARM:
				cmd.User = setup.Set.MyName
				buffer, err := json.Marshal(cmd)
				if err != nil {
					logger.Error.Printf("%v %s", cmd, err.Error())
					break loop
				}
				writer.WriteString(string(buffer))
				writer.WriteString("\n")
				err = writer.Flush()
				if err != nil {
					logger.Error.Printf("%s %s", socket.RemoteAddr().String(), err.Error())
					break loop
				}
			}
		}
		socket.Close()
	}
}
