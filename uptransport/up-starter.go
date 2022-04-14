package uptransport

import (
	"context"
	"fmt"
	"time"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/tulagate/agtransport"
	"github.com/ruraomsk/tulagate/controller"
	"github.com/ruraomsk/tulagate/db"
	"github.com/ruraomsk/tulagate/proto"
	"github.com/ruraomsk/tulagate/setup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var ExtCtrls map[string]controller.ExternalController
var SendToAmiChan chan controller.MessageToAmi

// var amiClient proto.AMIClient  //= proto.NewAMIClient(grpcConn)
// var ctx context.Context        //= metadata.NewOutgoingContext(context.Background(), metadata.Pairs("protocol", "dkst"))
// var stream proto.AMI_RunClient //= amiClient.Run(ctx)

func Starter() {
	ExtCtrls = make(map[string]controller.ExternalController)
	SendToAmiChan = make(chan controller.MessageToAmi, 1000)
	// go emptyConnect()
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
		defer grpcConn.Close()
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

		go SendToAmi(stream)
		for {
			recv, err := stream.Recv()
			if err != nil {
				logger.Error.Println(err.Error())
				break
			}
			if agtransport.ReadyAgTransport() {
				ch, err := db.GetChanelForMessage(recv.ControllerId)
				if err != nil {
					logger.Error.Print(err.Error())
					continue
				}
				ch <- controller.MessageFromAmi{Action: recv.Action, Body: recv.Body}
				logger.Debug.Printf("recv %v", controller.MessageFromAmi{Action: recv.Action, Body: recv.Body})
			} else {
				logger.Error.Printf("нет связи с ag-server игнорируем")
			}
		}
		grpcConn.Close()
		logger.Error.Printf("Связь с верхом оборвалась")
		time.Sleep(5 * time.Second)
	}

}
func emptyConnect() {
	// logger.Debug.Print("empty")
	for {
		message := <-SendToAmiChan
		logger.Info.Printf("no connect %v", proto.RequestRun{ControllerId: message.IDExternal, Action: message.Action, Body: message.Body})
	}
}
func SendToAmi(stream proto.AMI_RunClient) {
	for {
		message := <-SendToAmiChan
		// logger.Debug.Print(message.Body)
		// if message.Action == "replay" {
		// 	var statusCode = codes.OK
		// 	// if message.Body == "ok" {
		// 	// } else {
		// 	// 	statusCode = codes.NotFound
		// 	// }
		// 	err := status.Error(statusCode, message.Body)
		// 	if err != nil {
		// 		logger.Error.Println(err.Error())
		// 		return
		// 	}
		// 	continue
		// }
		// logger.Info.Printf("sendtoami %v", proto.RequestRun{ControllerId: message.IDExternal, Action: message.Action, Body: message.Body})
		err := stream.Send(&proto.RequestRun{ControllerId: message.IDExternal, Action: message.Action, Body: message.Body, Protocol: "DKST"})
		if err != nil {
			logger.Error.Println(err.Error())
			return
		}

	}
}
