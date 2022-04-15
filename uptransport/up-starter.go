package uptransport

import (
	"context"
	"fmt"
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
var timeKeepAliveAmi = time.Duration(1 * time.Minute)

// var amiClient proto.AMIClient  //= proto.NewAMIClient(grpcConn)
// var ctx context.Context        //= metadata.NewOutgoingContext(context.Background(), metadata.Pairs("protocol", "dkst"))
// var stream proto.AMI_RunClient //= amiClient.Run(ctx)

func Starter() {
	SendToAmiChan = make(chan controller.MessageToAmi, 1000)
	internalSendToAmiChan = make(chan controller.MessageToAmi, 1000)
	DebugStopAmi = make(chan interface{})

	workAmi = false
	go controlConnect()
	/* Подключение к gRPC */
	grpcURL := fmt.Sprintf("%s:%d", setup.Set.RemoteHost, setup.Set.RemotePort)
	// grpcConn, err := grpc.Dial(grpcURL, grpc.WithInsecure(), grpc.WithBlock())
	for {
		grpcConn, err := grpc.Dial(grpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Error.Panicf("процедура подключения к gRPC завершена с ошибкой: %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		logger.Info.Println("gRPC Соединение установлено")
		amiClient := proto.NewAMIClient(grpcConn)
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("protocol", "dkst"))
		stream, err := amiClient.Run(ctx)
		if err != nil {
			logger.Error.Print(err.Error())
			grpcConn.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		workAmi = true
		go SendToAmi(stream)
		for {
			recv, err := stream.Recv()
			if err != nil {
				logger.Error.Println(err.Error())
				break
			}
			ch, err := db.GetChanelForMessage(recv.ControllerId)
			if err != nil {
				logger.Error.Print(err.Error())
				continue
			}
			ch <- controller.MessageFromAmi{Action: recv.Action, Body: recv.Body}
			logger.Debug.Printf("recv %v", controller.MessageFromAmi{Action: recv.Action, Body: recv.Body})
		}
		workAmi = false
		grpcConn.Close()
		logger.Error.Printf("Связь с верхом оборвалась")
		time.Sleep(15 * time.Second)
	}
}
func controlConnect() {
	// logger.Debug.Print("empty")
	time.Sleep(10 * time.Second)
	intervalTime := time.NewTimer(timeKeepAliveAmi)
	intervalTime.Stop()
	oneSecond := time.NewTicker(time.Second)
	sended := false
	for {
		select {
		case <-oneSecond.C:
			if !workAmi {
				if !sended {
					// logger.Debug.Println("Начинаем считать время ")
					intervalTime = time.NewTimer(timeKeepAliveAmi)
					sended = true
				}
			}
		case message := <-SendToAmiChan:
			if workAmi {
				// logger.Debug.Println("Начинаем считать время заново")
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
	for {
		select {
		case <-DebugStopAmi:
			workAmi = false
			return
		case message := <-internalSendToAmiChan:
			err := stream.Send(&proto.RequestRun{ControllerId: message.IDExternal, Action: message.Action, Body: message.Body, Protocol: "DKST"})
			if err != nil {
				logger.Error.Println(err.Error())
				workAmi = false
				return
			}

		}

	}
}
