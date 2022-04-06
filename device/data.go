package device

import (
	"github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/tulagate/controller"
)

type Device struct {
	OneSet    controller.OneSet
	Region    pudge.Region
	Cross     *pudge.Cross
	Ctrl      *pudge.Controller
	DevPhases chan comm.DevPhases
}
