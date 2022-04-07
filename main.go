package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/ruraomsk/ag-server/logger"

	"github.com/ruraomsk/tulagate/agtransport"
	"github.com/ruraomsk/tulagate/controller"
	"github.com/ruraomsk/tulagate/db"
	"github.com/ruraomsk/tulagate/device"
	"github.com/ruraomsk/tulagate/setup"
	"github.com/ruraomsk/tulagate/uptransport"
)

// var (
// 	remoteHost = "89.20.134.2"
// 	remotePort = 34221
// )

var (
	//go:embed config
	config embed.FS
)
var dkset controller.DKSet

func init() {
	setup.Set = new(setup.Setup)
	if _, err := toml.DecodeFS(config, "config/config.toml", &setup.Set); err != nil {
		fmt.Println("Dissmis config.toml")
		os.Exit(-1)
		return
	}
	os.MkdirAll(setup.Set.LogPath, 0777)
	if err := logger.Init(setup.Set.LogPath); err != nil {
		log.Panic("Error logger system", err.Error())
		return
	}
	buffer, err := config.ReadFile("config/dkset.json")
	if err != nil {
		logger.Error.Println(err.Error())
		fmt.Println(err.Error())
		os.Exit(-1)
		return
	}
	err = json.Unmarshal(buffer, &dkset)
	if err != nil {
		logger.Error.Println(err.Error())
		os.Exit(-1)
		return
	}

}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logger.Info.Print("Start tulagate")
	db.Starter(&dkset)
	device.Starter(&dkset)
	agtransport.Starter()
	uptransport.Starter()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
loop:
	for {
		<-c
		fmt.Println("\nWait make abort...")
		time.Sleep(3 * time.Second)
		break loop
	}
	logger.Info.Print("Stop tulagate")
}
