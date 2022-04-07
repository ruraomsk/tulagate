package device

import (
	"github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/tulagate/controller"
)

var (
	idToRegion  map[int]pudge.Region
	uidToRegion map[string]pudge.Region
	devices     map[pudge.Region]Device
)

type Device struct {
	OneSet         controller.OneSet
	Region         pudge.Region
	Cross          pudge.Cross
	Ctrl           pudge.Controller
	DevPhases      chan comm.DevPhases
	MessageForMe   chan controller.MessageFromAmi
	HoldPhase      controller.HoldPhase
	CountHoldPhase int
	State          int //State  в их понимании
	ErrorDB        []string
}
