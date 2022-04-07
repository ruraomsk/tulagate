package device

import (
	"github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/tulagate/controller"
	"github.com/ruraomsk/tulagate/db"
	"github.com/ruraomsk/tulagate/setup"
)

func Starter(dks *controller.DKSet) error {
	idToRegion = make(map[int]pudge.Region)
	uidToRegion = make(map[string]pudge.Region)
	devices = make(map[pudge.Region]Device)

	for _, v := range dks.DKSets {
		region := pudge.Region{Region: setup.Set.Region, Area: v.Area, ID: v.ID}
		device := Device{OneSet: v, Region: region, DevPhases: make(chan comm.DevPhases),
			MessageForMe: make(chan controller.MessageFromAmi)}
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
		} else {
			device.Ctrl = ctrl
		}
		idToRegion[device.Cross.IDevice] = region
		uidToRegion[v.IDExternal] = region
		db.AddChanReceivePhases(device.Cross.IDevice, device.DevPhases)
		db.AddChanelForMessage(device.OneSet.IDExternal, device.MessageForMe)
		devices[region] = device
		go device.worker()
	}
	return nil
}
