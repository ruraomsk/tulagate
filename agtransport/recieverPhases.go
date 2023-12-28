package agtransport

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/tulagate/db"
	"github.com/ruraomsk/tulagate/setup"
)

var sending map[int]bool

func recieverPhases() {
	sending = make(map[int]bool)
	w := fmt.Sprintf("%s:%d", setup.Set.AgServer.Host, setup.Set.AgServer.PortDevices)
	for {
		time.Sleep(15 * time.Second)
		socket, err := net.Dial("tcp", w)
		if err != nil {
			logger.Error.Printf("connect %s %s ", w, err.Error())
			continue
		}
		reader := bufio.NewReader(socket)
		logger.Info.Print("recieverPhases ready")

		var phases comm.DevPhases
	loop:
		for {
			str, err := reader.ReadString('\n')
			if err != nil {
				logger.Error.Printf("%s %s", socket.RemoteAddr().String(), err.Error())
				break loop
			}
			str = strings.ReplaceAll(str, "\n", "")
			if strings.Compare(str, "0") == 0 {
				//keep alive
				continue loop
			}
			err = json.Unmarshal([]byte(str), &phases)
			if err != nil {
				logger.Error.Printf("%s %s", socket.RemoteAddr().String(), err.Error())
				break loop
			}
			// logger.Debug.Printf("receive %v", phases)
			ch, err := db.GetChanReceivePhases(phases.ID)
			if err != nil {
				_, is := sending[phases.ID]
				if !is {
					logger.Error.Printf("%s %s", socket.RemoteAddr().String(), err.Error())
					sending[phases.ID] = true
				}
				continue loop
			}
			_, is := sending[phases.ID]
			if is {
				delete(sending, phases.ID)
			}

			ch <- phases
		}
		socket.Close()
	}

}
