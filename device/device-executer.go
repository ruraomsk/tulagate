package device

import (
	"encoding/json"
	"fmt"

	"github.com/ruraomsk/ag-server/binding"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/tulagate/agtransport"
	"github.com/ruraomsk/tulagate/controller"
	"github.com/ruraomsk/tulagate/db"
	"github.com/ruraomsk/tulagate/uptransport"
)

func (d *Device) executeStartWork() {
	d.offMessage()
	agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: 9}
	agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 5, Params: 0}
	agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 4, Params: 1}

}
func (d *Device) stop() {
	d.offMessage()
	agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: 9}
	agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 5, Params: 0}
	agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 4, Params: 0}

}
func (d *Device) sendReplayToAmiWithStatus(message string) {
	d.ErrorTech = make([]string, 0)
	if len(message) != 0 {
		d.ErrorTech = append(d.ErrorTech, message)
	}
	uptransport.SendToAmiChan <- d.sendStatus()
	d.ErrorTech = make([]string, 0)
}
func (d *Device) executeSetMode(message controller.MessageFromAmi) string {
	var setter controller.SetMode
	var err1 = "device offline"
	var err2 = "device out of control"
	if !d.Ctrl.IsConnected() {
		d.offMessage()
		logger.Error.Println(err1)
		return err1
	}
	if !db.GetControlStatus(d.Cross.StatusDevice) {
		logger.Error.Println(err2)
		return err2
	}
	err := json.Unmarshal([]byte(message.Body), &setter)
	if err != nil {
		logger.Error.Println(err.Error())
		return err.Error()
	}
	if !setter.Is_enabled {
		agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: 9}
		return ""
	}
	switch setter.Mode {
	case 3:
		agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: 0x0a}
		return ""
	case 5:
		agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: 0x0b}
		return ""
	}
	return fmt.Sprintf("unsupported %d set mode", setter.Mode)
}
func (d *Device) executeHoldPhase(message controller.MessageFromAmi) string {
	var setter controller.HoldPhase
	var err1 = "device offline"
	var err2 = "device out of control"
	if !d.Ctrl.IsConnected() {
		d.offMessage()
		logger.Error.Println(err1)
		return err1
	}
	if !db.GetControlStatus(d.Cross.StatusDevice) {
		logger.Error.Println(err2)
		return err2
	}
	err := json.Unmarshal([]byte(message.Body), &setter)
	if err != nil {
		logger.Error.Println(err.Error())
		return err.Error()
	}
	if !setter.Unhold_phase {
		d.HoldPhase.Max_duration = 0
		d.HoldPhase.Phase_number = 0
		d.HoldPhase.Unhold_phase = false
		d.CountHoldPhase = 0
		agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: 9}
		return ""
	}
	d.HoldPhase.Max_duration = setter.Max_duration
	d.HoldPhase.Phase_number = setter.Phase_number
	d.HoldPhase.Unhold_phase = true
	d.CountHoldPhase = 0
	agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: setter.Phase_number}
	return ""
}

func (d *Device) executeSwitchProgram(message controller.MessageFromAmi) string {
	var setter controller.SwitchProgram
	var err1 = "device offline"
	var err2 = "device out of control"
	if !d.Ctrl.IsConnected() {
		logger.Error.Println(err1)
		return err1
	}
	if !db.GetControlStatus(d.Cross.StatusDevice) {
		logger.Error.Println(err2)
		return err2
	}
	err := json.Unmarshal([]byte(message.Body), &setter)
	if err != nil {
		logger.Error.Println(err.Error())
		return err.Error()
	}
	if !setter.Switch_default {
		d.isDUPK = false
		agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 5, Params: 0}
		return ""
	}
	if setter.Programm_number > 0 && setter.Programm_number <= 12 {
		d.isDUPK = true
		agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 5, Params: setter.Programm_number}
		return ""
	}
	err3 := fmt.Sprintf("unsupported %d programm", setter.Programm_number)
	logger.Error.Printf(err3)
	return err3
}

func (d *Device) executeUploadPrograms(message controller.MessageFromAmi) string {
	var setter controller.Programm
	// logger.Debug.Println(message)
	err := json.Unmarshal([]byte(message.Body), &setter)
	if err != nil {
		logger.Error.Printf("%s %s", message.Body, err.Error())
		return err.Error()
	}
	// logger.Debug.Println(setter)
	if setter.Number < 1 || setter.Number > 12 {
		err1 := fmt.Sprintf("unsupported %d programm", setter.Number)
		logger.Error.Printf(err1)
		return err1
	}
	if setter.Mode != 0 {
		for i, v := range d.Cross.Arrays.SetDK.DK {
			if v.Pk == setter.Number {
				d.Cross.Arrays.SetDK.DK[i] = binding.NewSetPk(v.Pk)
				d.Cross.Arrays.SetDK.DK[i].Tc = setter.Mode
				d.Cross.Arrays.SetDK.DK[i].Shift = setter.Offset
				d.Cross.Arrays.SetDK.DK[i].TypePU = 1
				if setter.Is_Coordination {
					d.Cross.Arrays.SetDK.DK[i].TypePU = 0
				}
				// logger.Debug.Print(d.Cross.Arrays.SetDK.DK[i])

				agtransport.SendCross <- pudge.UserCross{State: d.Cross}
				if setter.IsDefault {
					db.SetBasePlan(d.Region, d.Cross.Arrays.SetDK, setter.Number)
				}
				return ""
			}
		}
		err5 := fmt.Sprintf("%d нет такого плана в системе", setter.Number)
		logger.Error.Printf(err5)
		return err5

	}
	if len(setter.Phases) > 12 {
		err2 := fmt.Sprintf("слишком много фаз в %d  не больше 12", setter.Number)
		logger.Error.Printf(err2)
		return err2
	}
	// if setter.Mode != 0 {

	// }
	// if !setter.IsEnabled {
	// 	//Удаляем план создаем в нем ЛР
	// 	for i, v := range d.Cross.Arrays.SetDK.DK {
	// 		if v.Pk == setter.Programm_number {
	// 			if v.Pk == setter.Programm_number {
	// 				d.Cross.Arrays.SetDK.DK[i] = binding.NewSetPk(v.Pk)
	// 				d.Cross.Arrays.SetDK.DK[i].Tc = 0 //Локальный режим
	// 				agtransport.SendCross <- pudge.UserCross{State: d.Cross}
	// 				return ""
	// 			}
	// 		}
	// 	}
	// 	return fmt.Sprintf("%d нет такого плана в системе", setter.Programm_number)
	// }
	// if len(setter.Phases) < 1 {
	// 	return fmt.Sprintf("слишком мало фаз в %d ", setter.Number)
	// }
	//считаем время цикла
	tcycle := 0
	for _, v := range setter.Phases {
		if v.Type != 2 {
			tcycle += v.Duration
		}
	}
	if tcycle == 0 {
		for i, v := range d.Cross.Arrays.SetDK.DK {
			if v.Pk == setter.Number {
				d.Cross.Arrays.SetDK.DK[i] = binding.NewSetPk(v.Pk)
				d.Cross.Arrays.SetDK.DK[i].Tc = tcycle
				d.Cross.Arrays.SetDK.DK[i].Shift = setter.Offset
				d.Cross.Arrays.SetDK.DK[i].TypePU = 1
				if setter.Is_Coordination {
					d.Cross.Arrays.SetDK.DK[i].TypePU = 0
				}
				// logger.Debug.Print(d.Cross.Arrays.SetDK.DK[i])

				agtransport.SendCross <- pudge.UserCross{State: d.Cross}
				if setter.IsDefault {
					db.SetBasePlan(d.Region, d.Cross.Arrays.SetDK, setter.Number)
				}
				return ""
			}
		}
		err5 := fmt.Sprintf("%d нет такого плана в системе", setter.Number)
		logger.Error.Printf(err5)
		return err5

	}
	if setter.Offset >= tcycle {
		err3 := fmt.Sprintf("смещение цикла не должно быть больше или равно времени цикла в %d программе", setter.Number)
		logger.Error.Printf(err3)
		return err3
	}
	// sort.Slice(setter.Phases, func(i, j int) bool {
	// 	return setter.Phases[i].Phase_order < setter.Phases[j].Phase_order
	// })
	if setter.Phases[0].Number != 1 {
		err4 := fmt.Sprintf("первая фаза в цикле всегда должна быть первая в %d программе", setter.Number)
		logger.Error.Printf(err4)
		return err4
	}
	for i, v := range d.Cross.Arrays.SetDK.DK {
		if v.Pk == setter.Number {
			d.Cross.Arrays.SetDK.DK[i] = binding.NewSetPk(v.Pk)
			d.Cross.Arrays.SetDK.DK[i].Tc = tcycle
			d.Cross.Arrays.SetDK.DK[i].Shift = setter.Offset
			d.Cross.Arrays.SetDK.DK[i].TypePU = 1
			if setter.Is_Coordination {
				d.Cross.Arrays.SetDK.DK[i].TypePU = 0
			}
			tnow := setter.Offset
			for j, v := range setter.Phases {
				d.Cross.Arrays.SetDK.DK[i].Stages[j].Start = tnow
				if v.Number == 0 {
					d.Cross.Arrays.SetDK.DK[i].Stages[j].Tf = 1
				} else {
					if v.Type == 0 {
						d.Cross.Arrays.SetDK.DK[i].Stages[j].Tf = 0
					} else if v.Type == 1 {
						d.Cross.Arrays.SetDK.DK[i].Stages[j].Tf = 2
					} else if v.Type == 2 {
						d.Cross.Arrays.SetDK.DK[i].Stages[j].Tf = 7
					}
				}
				d.Cross.Arrays.SetDK.DK[i].Stages[j].Number = v.Number
				if tnow+v.Duration >= tcycle {
					// logger.Debug.Print(d.Cross.Arrays.SetDK.DK[i].Stages[j])
					d.Cross.Arrays.SetDK.DK[i].Stages[j].Trs = true
					d.Cross.Arrays.SetDK.DK[i].Stages[j].Stop = tcycle
					tnow += v.Duration - tcycle
					d.Cross.Arrays.SetDK.DK[i].Stages[j].Dt = tnow
				} else {
					tnow += v.Duration
					d.Cross.Arrays.SetDK.DK[i].Stages[j].Stop = tnow
					if v.Type == 2 {
						tnow -= v.Duration
					}
				}
			}

			// logger.Debug.Print(d.Cross.Arrays.SetDK.DK[i])

			agtransport.SendCross <- pudge.UserCross{State: d.Cross}
			if setter.IsDefault {
				db.SetBasePlan(d.Region, d.Cross.Arrays.SetDK, setter.Number)
			}
			return ""
		}
	}
	err5 := fmt.Sprintf("%d нет такого плана в системе", setter.Number)
	logger.Error.Printf(err5)
	return err5
}
func (d *Device) offMessage() {
	if !d.Ctrl.IsConnected() {
		logger.Error.Printf("Устройство %s %v не на связи ", d.OneSet.IDExternal, d.Region)
	}
}

func (d *Device) executeUploadDailyCards(message controller.MessageFromAmi) string {
	var setter []controller.DailyCard
	err := json.Unmarshal([]byte(message.Body), &setter)
	if err != nil {
		logger.Error.Println(err.Error())
		return err.Error()
	}
	send := false
	// logger.Debug.Println(setter)
	for _, v := range setter {
		send = true
		err := v.ToDaySet(&d.Cross.Arrays.DaySets)
		if err != nil {
			return err.Error()
		}
	}
	if send {
		agtransport.SendCross <- pudge.UserCross{State: d.Cross}
	}
	return ""
}
func (d *Device) executeUploadWeekCards(message controller.MessageFromAmi) string {
	var setter []controller.Week
	err := json.Unmarshal([]byte(message.Body), &setter)
	if err != nil {
		logger.Error.Println(err.Error())
		return err.Error()
	}
	for _, v := range setter {
		count := 0
		if v.Number == 1 {
			for _, wd := range v.DailyCards {
				if wd != 0 {
					count++
				}
			}
			if count != 7 {
				logger.Error.Println("В недельной карте 1 нельзя указывать нулевую суточную")
				return "error!"
			}
		}
	}
	send := false
	for _, v := range setter {
		send = true
		err := v.ToWeekSet(&d.Cross.Arrays.WeekSets)
		if err != nil {
			return err.Error()
		}
	}
	if send {
		agtransport.SendCross <- pudge.UserCross{State: d.Cross}
	}
	return ""
}
