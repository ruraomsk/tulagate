package uptransport

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/tulagate/controller"
	"github.com/ruraomsk/tulagate/db"
	"github.com/ruraomsk/tulagate/proto"
	"github.com/ruraomsk/tulagate/setup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var SendToAmiChan chan controller.MessageToAmi
var internalSendToAmiChan chan controller.MessageToAmi
var DebugStopAmi chan interface{}
var workAmi bool
var timeKeepAliveAmi time.Duration

// var amiClient proto.AMIClient  //= proto.NewAMIClient(grpcConn)
// var ctx context.Context        //= metadata.NewOutgoingContext(context.Background(), metadata.Pairs("protocol", "dkst"))
// var stream proto.AMI_RunClient //= amiClient.Run(ctx)

func Starter() {
	count := 0
	SendToAmiChan = make(chan controller.MessageToAmi)
	internalSendToAmiChan = make(chan controller.MessageToAmi)
	DebugStopAmi = make(chan interface{})
	timeKeepAliveAmi = time.Duration(setup.Set.TimeKeepAliveAmi) * time.Second
	workAmi = false
	go controlConnect()
	/* Подключение к gRPC */
	grpcURL := fmt.Sprintf("%s:%d", setup.Set.RemoteHost, setup.Set.RemotePort)
	// grpcConn, err := grpc.Dial(grpcURL, grpc.WithInsecure(), grpc.WithBlock())
	for {
		ctx1, _ := context.WithTimeout(context.Background(), time.Second*60)
		grpcConn, err := grpc.DialContext(ctx1, grpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Error.Printf("процедура подключения к gRPC завершена с ошибкой: %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		// logger.Info.Println("gRPC активно")
		amiClient := proto.NewAMIClient(grpcConn)
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("protocol", "dkst"))
		stream, err := amiClient.Run(ctx)
		if err != nil {
			if count%100 == 0 {
				logger.Error.Print(err.Error())
			}
			count++
			grpcConn.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		workAmi = true
		logger.Info.Println("Связь есть с верхом")
		count = 0
		go SendToAmi(stream)
		for {
			recv, err := stream.Recv()
			if err != nil {
				logger.Error.Println(err.Error())
				break
			}
			if !workAmi {
				break
			}
			ch, err := db.GetChanelForMessage(recv.ControllerId)
			if err != nil {
				logger.Error.Print(err.Error())
				continue
			}
			ch <- controller.MessageFromAmi{Action: recv.Action, Body: recv.Body}
			// logger.Debug.Printf("recv %v", controller.MessageFromAmi{Action: recv.Action, Body: recv.Body})
		}
		stream.CloseSend()
		workAmi = false
		grpcConn.Close()
		logger.Error.Printf("Связь с верхом оборвалась")
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}
func controlConnect() {
	// logger.Debug.Print("empty")
	// time.Sleep(10 * time.Second)
	intervalTime := time.NewTimer(timeKeepAliveAmi)
	oneSecond := time.NewTicker(time.Second)
	sended := false
	for {
		select {
		case <-oneSecond.C:
			if !workAmi {
				if !sended {
					logger.Debug.Println("Начинаем считать время ")
					intervalTime = time.NewTimer(timeKeepAliveAmi)
					sended = true
				}
			} else {
				if sended {
					sended = false
					intervalTime = time.NewTimer(timeKeepAliveAmi)
				}
			}
		case message := <-SendToAmiChan:
			if workAmi {
				// logger.Debug.Println("Пришло сообщение... Начинаем считать время заново")
				internalSendToAmiChan <- message
				intervalTime.Stop()
				intervalTime = time.NewTimer(timeKeepAliveAmi)
			}
		case <-intervalTime.C:
			db.AllStops()
		}
	}
}
func SendToAmi(stream proto.AMI_RunClient) {
	defer stream.CloseSend()
	oneSecond := time.NewTicker(time.Second)
	for {
		select {
		case <-oneSecond.C:
			if !workAmi {
				return
			}
		case <-DebugStopAmi:
			workAmi = false
			return
		case message := <-internalSendToAmiChan:
			logger.Debug.Printf("send to %v", message)
			err := stream.Send(&proto.RequestRun{ControllerId: message.IDExternal, Action: message.Action, Body: message.Body, Protocol: "DKST"})
			if err != nil {
				logger.Error.Println(err.Error())
				workAmi = false
				return
			}

		}

	}
}
