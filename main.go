package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/BurntSushi/toml"

	"github.com/ruraomsk/ag-server/logger"

	"github.com/ruraomsk/tulagate/controller"
	"github.com/ruraomsk/tulagate/setup"
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

}
