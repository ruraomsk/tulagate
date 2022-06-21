package device

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ruraomsk/ag-server/binding"
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
	logger.Info.Printf("Начинаем работу %s %v", d.OneSet.IDExternal, d.Region)
	for !agtransport.ReadyAgTransport() {
		time.Sleep(time.Second)
	}
	d.loadData()
	baseCross, _ := db.GetStartCross(d.Region)
	db.MoveData(&d.Cross, &baseCross)
	agtransport.SendCross <- pudge.UserCross{State: d.Cross}
	logger.Info.Printf("Откатились на базовое состояние %v", d.Cross.IDevice)
	tickSFDK := time.NewTicker(time.Minute)
	tickOneSecond := time.NewTicker(time.Second)

	if d.Ctrl.IsConnected() {
		if setup.Set.MGRSet {
			//есть МГР
			d.LastMGR = 1
			agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 0x0f, Params: 0}
		} else {
			//нет МГР
			d.LastMGR = 0
			agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 0x0f, Params: 1}
		}
	}
	d.executeStartWork()
	// }
	for {
		select {
		case <-d.clear:
			//Пришла команда почистить свое состояние
			d.loadData()
			baseCross, _ := db.GetStartCross(d.Region)
			db.MoveData(&d.Cross, &baseCross)
			agtransport.SendCross <- pudge.UserCross{State: d.Cross}
			logger.Info.Printf("Откатились по разрыву связи с верхом %v", d.Cross.IDevice)
			if d.Ctrl.IsConnected() {
				if setup.Set.MGRSet {
					//нет МГР
					if d.LastMGR == 1 {
						agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 0x0f, Params: 0}
						d.LastMGR = 0
					}
				}
			}
		case <-tickOneSecond.C:
			ptime := TimeNowOfSecond()
			if ptime%d.Stat.interval == 0 {
				d.sendStatistics(ptime)
			}
			if !agtransport.ReadyAgTransport() {
				d.sendNotTransport()
				continue
			}
			if d.HoldPhase.Unhold_phase {
				//Есть команда на удержание фазы
				if d.DK.FDK == d.HoldPhase.Phase_number {
					d.CountHoldPhase++
					if d.CountHoldPhase >= d.HoldPhase.Max_duration {
						d.HoldPhase.Max_duration = 0
						d.HoldPhase.Phase_number = 0
						d.HoldPhase.Unhold_phase = false
						d.CountHoldPhase = 0
						agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: 9}
					} else {
						if d.CountHoldPhase%30 == 0 {
							agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: d.HoldPhase.Phase_number}
						}
					}
				}
			}
			if time.Since(d.LastSendStatus) > time.Minute {
				uptransport.SendToAmiChan <- d.sendStatus()
			}
			d.loadData()
			if !d.Ctrl.IsConnected() {
				continue
			}
			if setup.Set.MGRSet {
				mgr := d.Stat.nowStat.getMGRword()
				d.Stat.nowStat.init()
				if mgr != 0 {
					agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 0x0e, Params: mgr}
				}
			}
			if setup.Set.MGRSet {
				if time.Since(d.LastReciveStat) > time.Duration(setup.Set.TimeKeepStatistic)*time.Second {
					//нет МГР
					if d.LastMGR == 1 {
						agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 0x0f, Params: 1}
						d.LastMGR = 0
					}
				} else {
					//есть МГР
					if d.LastMGR == 0 {
						agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 0x0f, Params: 0}
						d.LastMGR = 1
					}
				}
			}

		case <-tickSFDK.C:
			if !agtransport.ReadyAgTransport() {
				d.sendNotTransport()
				continue
			}
			//Шлем устройству СФДК
			// logger.Debug.Printf("dev %v 1 minute", d.Region)
			if d.Ctrl.IsConnected() && !d.Ctrl.StatusCommandDU.IsReqSFDK1 {
				agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, User: setup.Set.MyName, Command: 4, Params: 1}
			}
		case dk := <-d.DevPhases:
			if !agtransport.ReadyAgTransport() {
				d.sendNotTransport()
				continue
			}
			//Пришло измение по фазам
			d.DK = dk.DK
			d.loadData()
			// logger.Debug.Printf("%v %v", d.Region, d.DK)
			uptransport.SendToAmiChan <- d.sendStatus()
		case message := <-d.MessageForMe:
			if !agtransport.ReadyAgTransport() {
				d.sendNotTransport()
				continue
			}
			logger.Debug.Printf("%v %v", d.Region, message)
			switch message.Action {
			case "SetMode":
				d.sendReplayToAmiWithStatus(d.executeSetMode(message))
			case "HoldPhase":
				d.sendReplayToAmiWithStatus(d.executeHoldPhase(message))
			case "SwitchProgram":
				d.sendReplayToAmiWithStatus(d.executeSwitchProgram(message))
			case "UploadPrograms":
				message = d.insertMGR(message)
				// d.sendReplayToAmi(d.executeUploadPrograms(message))
				d.executeUploadPrograms(message)
			case "GetCoordination":
				d.loadData()
				d.executeGetCoordination()
			case "ChanelStat":
				d.executeAddStat(message)
			case "UploadDailyCards":
				d.executeUploadDailyCards(message)
			case "UploadWeekCards":
				d.executeUploadWeekCards(message)
			case "Config":
				d.executeConfig(message)
			default:
				s := fmt.Sprintf("%s not supported", message.Action)
				logger.Error.Printf(s)
				d.sendReplayToAmiWithStatus(s)
			}
		}

	}
}
func (d *Device) loadData() {
	// logger.Debug.Printf("%d loadData", d.Cross.IDevice)
	d.Cross, err = db.GetCross(d.Region)
	if err != nil {
		logger.Error.Print(err.Error())
		return
	}
	d.Ctrl, err = db.GetController(d.Cross.IDevice)
	if err != nil {
		logger.Error.Print(err.Error())
		return
	}
}
func (d *Device) sendNotTransport() {
	message := controller.MessageToAmi{IDExternal: d.OneSet.IDExternal, Action: "status", Body: "{}"}
	status := controller.Status{State: 0}
	status.Timestamp = time.Now().Unix()
	status.Errors = controller.Errors{Hw_error: make([]string, 0), Sw_error: make([]string, 0), Ec_error: make([]string, 0), Detector_fault: make([]string, 0)}
	status.Errors.Sw_error = append(status.Errors.Sw_error, "Нет связи с системой управления ag-server!!!")
	body, _ := json.Marshal(&status)
	message.Body = string(body)
	d.LastSendStatus = time.Now()
	uptransport.SendToAmiChan <- message
}
func (d *Device) sendStatus() controller.MessageToAmi {
	message := controller.MessageToAmi{IDExternal: d.OneSet.IDExternal, Action: "status", Body: "{}"}
	status := controller.Status{Program_number: d.Cross.PK, Phase_number: d.DK.FDK, Has_Default_Programs: make([]int, 0), Has_Loaded_Programs: make([]int, 0)}
	if status.Phase_number > 8 {
		status.Phase_number = 0
	}
	if d.DK.FDK == 9 {
		status.Tact_number = 1
		status.Phase_number = d.DK.FTUDK
	} else {
		status.Tact_number = 2
	}
	status.Tact_tick = d.DK.TDK
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
	status.Errors.Sw_error = d.ErrorTech
	status.Errors.Is_door_opened = d.Ctrl.DK.ODK
	status.Channels_state = make([]controller.Channels_state, 0)
	status.Channels_powers = make([]float64, 0)
	status.Mode = 0
	switch d.DK.RDK {
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
	case 0:
		status.Mode = 2
	default:
		logger.Error.Printf("rdk=%d нужна перекодировка!", d.DK.RDK)
	}
	switch d.DK.FDK {
	case 10:
		status.Mode = 3
	case 14:
		status.Mode = 3
	case 11:
		status.Mode = 5
	case 15:
		status.Mode = 5
	case 12:
		status.Mode = 4
	default:
		if d.DK.FDK > 9 {
			logger.Error.Printf("fdk=%d нужна перекодировка!", d.DK.FDK)
		}
	}
	status.Has_Default_Programs = db.LoadBaseProgramm(d.Region)
	status.Has_Loaded_Programs = make([]int, 0)
	for _, v := range d.Cross.Arrays.SetDK.DK {
		if v.Tc > 3 {
			status.Has_Loaded_Programs = append(status.Has_Loaded_Programs, v.Pk)
		}
	}
	status.Has_loaded_daily_cards = make([]int, 0)
	for _, v := range d.Cross.Arrays.DaySets.DaySets {
		if v.Count != 0 {
			status.Has_loaded_daily_cards = append(status.Has_loaded_daily_cards, v.Number)
		}
	}
	status.Has_loaded_week_cards = make([]int, 0)
	for _, v := range d.Cross.Arrays.WeekSets.WeekSets {
		if !isWeekEmpty(v) {
			status.Has_loaded_week_cards = append(status.Has_loaded_week_cards, v.Number)
		}
	}
	status.Timestamp = time.Now().Unix()
	body, _ := json.Marshal(&status)
	message.Body = string(body)
	d.LastSendStatus = time.Now()
	return message
}
func isWeekEmpty(ow binding.OneWeek) bool {
	for _, v := range ow.Days {
		if v != 0 {
			return false
		}
	}
	return true
}
