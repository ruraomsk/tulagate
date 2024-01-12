package device

import (
	"time"

	"github.com/ruraomsk/ag-server/binding"
	"github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/tulagate/controller"
	"github.com/ruraomsk/tulagate/db"
	"github.com/ruraomsk/tulagate/setup"
)

func Starter(dks *controller.DKSet, stop chan interface{}, next chan interface{}) {
	idToRegion = make(map[int]pudge.Region)
	uidToRegion = make(map[string]pudge.Region)
	devices = make(map[pudge.Region]Device)

	for _, v := range dks.DKSets {
		if !v.Work {
			continue
		}
		region := pudge.Region{Region: setup.Set.Region, Area: v.Area, ID: v.ID}
		device := Device{OneSet: v, Region: region, DevPhases: make(chan comm.DevPhases),
			MessageForMe: make(chan controller.MessageFromAmi), ErrorTech: make([]string, 0), LastSendStatus: time.Now(),
			LastReciveStat: time.Now(), clear: make(chan interface{}), MGRS: make(map[int]binding.MGR)}
		device.Stat = statistic{tp: 1, interval: 300, count: 0}
		device.isMessage = false
		device.initStatistic()
		cross, err := db.GetCross(region)
		if err != nil {
			logger.Error.Print(err.Error())
			continue
		}
		device.Cross = cross
		ctrl, err := db.GetController(device.Cross.IDevice)
		if err != nil {
			logger.Error.Print(err.Error())
			device.Cross.StatusDevice = 18
			device.isMessage = true
		} else {
			device.Ctrl = ctrl
		}
		for _, m := range device.Cross.Arrays.MGRs {
			if m.TLen != 0 && m.TMGR != 0 {
				device.MGRS[m.Phase] = m
			}
		}
		idToRegion[device.Cross.IDevice] = region
		uidToRegion[v.IDExternal] = region
		db.AddChanReceivePhases(device.Cross.IDevice, device.DevPhases)
		db.AddChanelForMessage(device.OneSet.IDExternal, device.MessageForMe)
		db.AddChanToStop(region, device.clear)
		devices[region] = device
		go device.worker()
	}
	next <- 1
	for {
		<-stop
		for _, dev := range devices {
			dev.stop()
		}
	}
}

func ClearAllPKs() {
	for _, v := range devices {
		v.clear <- 1
	}
}
