package device

import (
	"encoding/json"
	"fmt"
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
	// if d.Ctrl.IsConnected() && !d.Ctrl.StatusCommandDU.IsReqSFDK1 {
	tickOneSecond := time.NewTicker(time.Second)
	agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, User: setup.Set.MyName, Command: 4, Params: 1}
	// }
	for {
		select {
		case <-tickOneSecond.C:
			d.loadData()
			if d.HoldPhase.Unhold_phase {
				if d.Ctrl.DK.FTSDK == d.HoldPhase.Phase_number {
					d.CountHoldPhase++
					if d.CountHoldPhase >= d.HoldPhase.Max_duration {
						d.HoldPhase.Max_duration = 0
						d.HoldPhase.Phase_number = 0
						d.HoldPhase.Unhold_phase = false
						d.CountHoldPhase = 0
						agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: 9}
					} else {
						if d.CountHoldPhase%50 == 0 {
							agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: d.HoldPhase.Phase_number}
						}
					}
				}
			}

		case <-tickSFDK.C:
			//Шлем устройству СФДК
			// logger.Debug.Printf("dev %v 1 minute", d.Region)
			agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, User: setup.Set.MyName, Command: 4, Params: 1}
		case dk := <-d.DevPhases:
			//Пришло измение по фазам
			d.loadData()
			d.Ctrl.DK = dk.DK
			uptransport.SendToAmiChan <- d.sendStatus()
		case message := <-d.MessageForMe:
			logger.Debug.Print(message)
			switch message.Action {
			case "SetMode":
				d.sendReplayToAmi(d.executeSetMode(message))
			case "HoldPhase":
				d.sendReplayToAmi(d.executeHoldPhase(message))
			case "SwitchProgram":
				d.sendReplayToAmi(d.executeSwitchProgram(message))
			default:
				logger.Error.Printf("not found %v", message)
				d.sendReplayToAmi(fmt.Sprintf("%s not supported", message.Action))
			}
		}

	}
}
func (d *Device) loadData() {
	// logger.Debug.Printf("%d loadData", d.Cross.IDevice)
	d.ErrorDB = make([]string, 0)
	d.Cross, err = db.GetCross(d.Region)
	if err != nil {
		logger.Error.Print(err.Error())
		d.ErrorDB = append(d.ErrorDB, err.Error())
		return
	}
	d.Ctrl, err = db.GetController(d.Cross.IDevice)
	if err != nil {
		logger.Error.Print(err.Error())
		d.ErrorDB = append(d.ErrorDB, err.Error())
		return
	}
}
func (d *Device) sendStatus() controller.MessageToAmi {
	message := controller.MessageToAmi{IDExternal: d.OneSet.IDExternal, Action: "status", Body: "{}"}
	status := controller.Status{Program_number: d.Cross.PK, Phase_number: d.Ctrl.DK.FTUDK}
	if d.Ctrl.DK.FTSDK == 9 {
		status.Tact_number = 1
	} else {
		status.Tact_number = 2
	}
	status.Tact_tick = d.Ctrl.DK.TDK
	if d.HoldPhase.Unhold_phase {
		status.Hold_phase_number = d.HoldPhase.Phase_number
		status.Hold_phase_time_remain = d.HoldPhase.Max_duration - d.CountHoldPhase
	} else {
		status.Hold_phase_number = 0
		status.Hold_phase_time_remain = 0
	}
	ip := strings.Split(d.Ctrl.IPHost, ":")
	// logger.Debug.Print(d.Ctrl)
	if len(ip) == 2 {
		status.Host = ip[0]
		status.Port, _ = strconv.Atoi(ip[1])
	}
	status.State = 0
	if d.Cross.StatusDevice != 18 {
		if db.GetControlStatus(d.Cross.StatusDevice) {
			status.State = 2
		} else {
			status.State = 1
		}
	}
	status.Errors = controller.Errors{Hw_error: make([]string, 0), Sw_error: make([]string, 0), Ec_error: make([]string, 0), Detector_fault: make([]string, 0)}
	if d.Ctrl.Error.RTC {
		status.Errors.Hw_error = append(status.Errors.Hw_error, "error RTC")
	}
	if d.Ctrl.Error.FRAM {
		status.Errors.Hw_error = append(status.Errors.Hw_error, "error FRAM")
	}
	if d.Ctrl.Error.TVP1 {
		status.Errors.Hw_error = append(status.Errors.Hw_error, "error TVP1")
	}
	if d.Ctrl.Error.TVP2 {
		status.Errors.Hw_error = append(status.Errors.Hw_error, "error TVP2")
	}
	if d.Ctrl.Error.V220DK1 || d.Ctrl.Error.V220DK2 {
		status.Errors.Hw_error = append(status.Errors.Hw_error, "error V220")
	}
	if d.Ctrl.GPS.E01 {
		status.Errors.Hw_error = append(status.Errors.Hw_error, "Нет связи с приемником GPS")
	}
	if d.Ctrl.GPS.E02 {
		status.Errors.Hw_error = append(status.Errors.Hw_error, "CRC связи с приемником GPS")
	}
	if d.Ctrl.GPS.E03 {
		status.Errors.Hw_error = append(status.Errors.Hw_error, "Нет валидного времени от GPS")
	}
	if d.Ctrl.GPS.E04 {
		status.Errors.Hw_error = append(status.Errors.Hw_error, "мало спутников GPS")
	}
	if d.Ctrl.GPS.Seek {
		status.Errors.Hw_error = append(status.Errors.Hw_error, "Поиск спутников GPS")
	}

	status.Errors.Is_door_opened = d.Ctrl.DK.ODK
	status.Channels_state = make([]controller.Channels_state, 0)
	status.Channels_powers = make([]float64, 0)
	status.Mode = 0
	switch d.Ctrl.DK.RDK {
	case 1:
		status.Mode = 6
	case 2:
		status.Mode = 6
	case 3:
		status.Mode = 8
	case 4:
		status.Mode = 7
	case 5:
		status.Mode = 1
	case 6:
		status.Mode = 1
	case 8:
		status.Mode = 2
	case 9:
		status.Mode = 2
	default:
		logger.Error.Printf("rdk=%d нужна перекодировка!", d.Ctrl.DK.RDK)
	}
	switch d.Ctrl.DK.FDK {
	case 10:
		status.Mode = 3
	case 11:
		status.Mode = 5
	case 12:
		status.Mode = 4
	default:
		if d.Ctrl.DK.FDK > 9 {
			logger.Error.Printf("fdk=%d нужна перекодировка!", d.Ctrl.DK.FDK)
		}
	}
	status.Timestamp = time.Now().Unix()
	body, _ := json.Marshal(&status)
	message.Body = string(body)
	// logger.Debug.Printf("message %v", message)
	return message

}
