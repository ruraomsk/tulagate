package tester

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/ruraomsk/tulagate/controller"
)

func Maker() {
	time.Sleep(time.Minute)
	tickOneSecond := time.NewTicker(time.Second)
	rand.Seed(time.Now().Unix())

	for {
		<-tickOneSecond.C
		s := controller.ChanelStat{}
		for i := 0; i < len(s.Chanels); i++ {
			s.Chanels[i] = rand.Intn(3)
		}
		buff, _ := json.Marshal(s)
		senderCommand("device3", "ChanelStat", string(buff))
	}
}
