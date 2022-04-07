package device

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/tulagate/agtransport"
	"github.com/ruraomsk/tulagate/controller"
	"github.com/ruraomsk/tulagate/db"
	"github.com/ruraomsk/tulagate/setup"
	"github.com/ruraomsk/tulagate/uptransport"
)

var err error

func (d *Device) worker() {
	//При запуске сразу шлем СФДК
	logger.Debug.Printf("device %v", d.Region)
	tickSFDK := time.NewTicker(time.Minute)
	// agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, User: setup.Set.MyName, Command: 4, Params: 1}
	for {
		select {
		case <-tickSFDK.C:
			//Шлем устройству СФДК исли оно есть и не установлен запрос СФДК
			if d.Ctrl.IsConnected() && !d.Ctrl.StatusCommandDU.IsReqSFDK1 {
				agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, User: setup.Set.MyName, Command: 4, Params: 1}
			}
		case dk := <-d.DevPhases:
			// logger.Debug.Print("there!")
			//Пришло измение по фазам
			d.loadData()
			// logger.Debug.Print("there!~!!")
			d.Ctrl.DK = dk.DK
			uptransport.SendToAmiChan <- d.sendStatus()
			// logger.Debug.Print("there!~!!111")
		}

	}
}
func (d *Device) loadData() {
	// logger.Debug.Printf("%d loadData", d.Cross.IDevice)
	d.ErrorDB = make([]string, 0)
	d.Cross, err = db.GetCross(d.Region)
	if err != nil {
		d.ErrorDB = append(d.ErrorDB, err.Error())
		return
	}
	d.Ctrl, err = db.GetController(d.Cross.IDevice)
	if err != nil {
		d.ErrorDB = append(d.ErrorDB, err.Error())
		return
	}
}
func (d *Device) sendStatus() controller.MessageToAmi {
	message := controller.MessageToAmi{IDExternal: d.OneSet.IDExternal, Action: "status", Body: "{}"}
	status := controller.Status{Program_number: d.Cross.PK, Phase_number: d.Ctrl.DK.FTSDK}
	if d.HoldPhase.Unhold_phase {
		status.Hold_phase_number = d.HoldPhase.Phase_number
		status.Hold_phase_time_remain = d.HoldPhase.Max_duration - d.CountHoldPhase
	} else {
		status.Hold_phase_number = 0
		status.Hold_phase_time_remain = 0
	}
	if d.Ctrl.IsConnected() {
		ip := strings.Split(d.Ctrl.IPHost, ":")
		status.Host = ip[0]
		status.Port, _ = strconv.Atoi(ip[1])
	}
	status.State = 0
	if d.Ctrl.IsConnected() {
		if db.GetControlStatus(d.Cross.StatusDevice) {
			status.State = 2
		} else {
			status.State = 1
		}
	}
	status.Errors = controller.Errors{Hw_error: make([]string, 0), Sw_error: make([]string, 0), Ec_error: make([]string, 0), Detector_fault: make([]string, 0)}
	status.Errors.Is_door_opened = d.Ctrl.DK.ODK
	status.Channels_state = make([]controller.Channels_state, 0)
	status.Channels_powers = make([]float64, 0)
	status.Mode = d.State
	status.Timestamp = time.Now().Unix()
	body, _ := json.Marshal(&status)
	message.Body = string(body)
	// logger.Debug.Printf("message %v", message)
	return message

}
