package uptransport

import (
	"context"
	"fmt"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/tulagate/controller"
	"github.com/ruraomsk/tulagate/db"
	"github.com/ruraomsk/tulagate/proto"
	"github.com/ruraomsk/tulagate/setup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var ExtCtrls map[string]controller.ExternalController
var SendToAmiChan chan controller.MessageToAmi

func Starter() error {
	ExtCtrls = make(map[string]controller.ExternalController)
	SendToAmiChan = make(chan controller.MessageToAmi, 1000)
	go emptyConnect()
	/* Подключение к gRPC */
	grpcURL := fmt.Sprintf("%s:%d", setup.Set.RemoteHost, setup.Set.RemotePort)
	// grpcConn, err := grpc.Dial(grpcURL, grpc.WithInsecure(), grpc.WithBlock())
	grpcConn, err := grpc.Dial(grpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("процедура подключения к gRPC завершена с ошибкой: %s", err.Error())

	}
	defer grpcConn.Close()
	logger.Info.Println("gRPC Соединение установлено")

	/* Инициализация клиента gRPC */
	grpcClient := proto.NewServiceClient(grpcConn)

	/* Создание контекста */
	ctx := context.Background()

	/* Опрос */
	serverResp, err := grpcClient.TrafficLightsControllersList(ctx, &proto.NoArguments{})
	if err != nil {
		statusCode := status.Code(err)
		if statusCode == codes.Unauthenticated {

			return fmt.Errorf("процедура опроса данных завершена с ошибкой в связи с истекшим временем действия токена")
		} else {
			return fmt.Errorf("процедура опроса данных завершена с ошибкой: %s", err.Error())
		}
	}
	if serverResp == nil {
		logger.Error.Println("Пустой ответ от сервера")
		return fmt.Errorf("пустой ответ от сервера")
	}

	for _, entity := range serverResp.Data {
		ctrl := controller.ExternalController{IDExternal: entity.Id, AddressRu: entity.AddressRu.Text}
		if entity.AddressLatin != nil {
			ctrl.AddressLatin = entity.AddressLatin.Text
		}
		if entity.Geom != nil {
			ctrl.Geom = controller.Geom{Latitude: entity.Geom.Latitude, Longitude: entity.Geom.Longitude}
		}
		if entity.LastProgrammId != nil {
			ctrl.LastProgrammId = entity.LastProgrammId.Id
		} else {
			ctrl.LastProgrammId = ""
		}
		ExtCtrls[entity.Id] = ctrl

	}
	amiClient := proto.NewAMIClient(grpcConn)
	stream, err := amiClient.Run(ctx)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	go RecieveFromAmi(stream)
	go SendToAmi(stream)
	return nil
}
func emptyConnect() {
	// logger.Debug.Print("empty")
	for {
		message := <-SendToAmiChan
		logger.Info.Printf("sendtoami %v", proto.RequestRun{ControllerId: message.IDExternal, Action: message.Action, Body: message.Body})
	}
}
func SendToAmi(stream proto.AMI_RunClient) {
	for {
		message := <-SendToAmiChan
		logger.Info.Printf("sendtoami %v", proto.RequestRun{ControllerId: message.IDExternal, Action: message.Action, Body: message.Body})
		err := stream.Send(&proto.RequestRun{ControllerId: message.IDExternal, Action: message.Action, Body: message.Body})
		if err != nil {
			logger.Error.Println(err.Error())
			return
		}

	}
}
func RecieveFromAmi(stream proto.AMI_RunClient) {
	for {
		recv, err := stream.Recv()
		if err != nil {
			logger.Error.Println(err.Error())
			return
		}
		ch, err := db.GetChanelForMessage(recv.ControllerId)
		if err != nil {
			logger.Error.Print(err.Error())
		}
		ch <- controller.MessageFromAmi{Action: recv.Action, Body: recv.Body}
	}

}
