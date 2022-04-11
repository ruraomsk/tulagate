package device

import (
	"encoding/json"
	"fmt"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/tulagate/agtransport"
	"github.com/ruraomsk/tulagate/controller"
)

func (d *Device) sendReplayToAmi(message string) {
	repl := controller.ErrorString{Message: message}
	buffer, err := json.Marshal(&repl)
	if err != nil {
		logger.Error.Print(err.Error())
		return
	}
	logger.Debug.Print(controller.MessageToAmi{IDExternal: d.OneSet.IDExternal, Action: "error", Body: string(buffer)})
	// uptransport.SendToAmiChan <- controller.MessageToAmi{IDExternal: d.OneSet.IDExternal, Action: "error", Body: string(buffer)}
}
func (d *Device) executeSetMode(message controller.MessageFromAmi) string {
	var setter controller.SetMode
	err := json.Unmarshal([]byte(message.Body), &setter)
	if err != nil {
		return err.Error()
	}
	if !setter.Is_enabled {
		agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: 9}
		return "ok"
	}
	switch setter.Mode {
	case 3:
		agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: 0x0a}
		return "ok"
	case 5:
		agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: 0x0b}
		return "ok"
	}
	return "unsupported"
}
func (d *Device) executeHoldPhase(message controller.MessageFromAmi) string {
	var setter controller.HoldPhase
	err := json.Unmarshal([]byte(message.Body), &setter)
	if err != nil {
		return err.Error()
	}
	if !setter.Unhold_phase {
		d.HoldPhase.Max_duration = 0
		d.HoldPhase.Phase_number = 0
		d.HoldPhase.Unhold_phase = false
		d.CountHoldPhase = 0
		agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: 9}
		return "ok"
	}
	phs := make([]int, 0)
	for i := 1; i < 13; i++ {
		phs = append(phs, d.Cross.Arrays.SetDK.GetPhases(i)...)
	}
	found := false
	for _, v := range phs {
		if v == setter.Phase_number {
			found = true
			break
		}
	}
	if !found {
		return fmt.Sprintf("unsupported phase %d", setter.Phase_number)
	}
	d.HoldPhase.Max_duration = setter.Max_duration
	d.HoldPhase.Phase_number = setter.Phase_number
	d.HoldPhase.Unhold_phase = true
	d.CountHoldPhase = 0
	agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 9, Params: setter.Phase_number}
	return "ok"
}
func (d *Device) executeSwitchProgram(message controller.MessageFromAmi) string {
	var setter controller.SwitchProgram
	err := json.Unmarshal([]byte(message.Body), &setter)
	if err != nil {
		return err.Error()
	}
	if !setter.Switch_default {
		agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 5, Params: 0}
		return "ok"
	}
	if setter.Programm_number > 0 && setter.Programm_number <= 12 {
		agtransport.CommandARM <- pudge.CommandARM{ID: d.Cross.IDevice, Command: 5, Params: setter.Programm_number}
		return "ok"
	}
	return "unsupported"
}
