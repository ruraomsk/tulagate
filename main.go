package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/tulagate/controller"

	"github.com/ruraomsk/tulagate/proto"

	"github.com/ruraomsk/tulagate/setup"

	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	status "google.golang.org/grpc/status"
)

// var (
// 	remoteHost = "89.20.134.2"
// 	remotePort = 34221
// )

var (
	//go:embed config
	config embed.FS
)
var ctrls map[string]controller.Controller

func init() {
	setup.Set = new(setup.Setup)
	if _, err := toml.DecodeFS(config, "config/config.toml", &setup.Set); err != nil {
		fmt.Println("Dissmis config.toml")
		os.Exit(-1)
		return
	}
	os.MkdirAll(setup.Set.LogPath, 0777)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := logger.Init(setup.Set.LogPath); err != nil {
		log.Panic("Error logger system", err.Error())
		return
	}
	ctrls = make(map[string]controller.Controller)
	/* Подключение к gRPC */
	grpcURL := fmt.Sprintf("%s:%d", setup.Set.RemoteHost, setup.Set.RemotePort)
	// grpcConn, err := grpc.Dial(grpcURL, grpc.WithInsecure(), grpc.WithBlock())
	grpcConn, err := grpc.Dial(grpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		logger.Error.Printf("Процедура подключения к gRPC завершена с ошибкой: %s", err.Error())
		return
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
			logger.Error.Printf("Процедура опроса данных завершена с ошибкой в связи с истекшим временем действия токена")
			return
		} else {
			logger.Error.Printf("Процедура опроса данных завершена с ошибкой: %s", err.Error())
			return
		}
	}
	if serverResp == nil {
		logger.Error.Println("Пустой ответ от сервера")
		return
	}
	for _, entity := range serverResp.Data {
		ctrl := controller.Controller{ID: entity.Id, AddressRu: entity.AddressRu.Text}
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
		ctrls[entity.Id] = ctrl

	}
	fmt.Println("start to ami")
	amiClient := proto.NewAMIClient(grpcConn)
	stream, err := amiClient.Run(ctx)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			fmt.Print(".")
		default:
			res, err := stream.Recv()
			if err != nil {
				fmt.Println(err.Error())
				panic(err)
			}
			fmt.Printf("stream read %v", res)
		}

	}
}
