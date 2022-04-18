package device

import (
	"time"

	"github.com/ruraomsk/ag-server/binding"
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
	clear          chan interface{}
	CountHoldPhase int
	State          int //State  в их понимании
	ErrorTech      []string
	LastSendStatus time.Time
	DK             pudge.DK
	MGRS           map[int]binding.MGR
}
