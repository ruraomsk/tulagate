package device

import (
	"encoding/json"

	"github.com/ruraomsk/tulagate/controller"
	"github.com/ruraomsk/tulagate/setup"
)

func (d *Device) insertMGR(message controller.MessageFromAmi) controller.MessageFromAmi {
	result := controller.MessageFromAmi{Action: message.Action}
	var setter controller.Programm
	err := json.Unmarshal([]byte(message.Body), &setter)
	if err != nil {
		return message
	}
	size := len(setter.Phases)
	phs := make([]controller.Phase, 0)
	if size != 0 {
		for i := 0; i < size; i++ {
			if !setup.Set.MGRSet {
				phs = append(phs, setter.Phases[i])
				continue
			}
			m, is := d.MGRS[setter.Phases[i].Number]
			if !is {
				phs = append(phs, setter.Phases[i])
				continue
			}
			if setter.Phases[i].Duration > (m.TLen + m.TMGR) {
				//Можно вставить МГР
				nph := controller.Phase{Number: 0, Duration: m.TMGR}
				setter.Phases[i].Duration = setter.Phases[i].Duration - m.TMGR
				phs = append(phs, setter.Phases[i])
				phs = append(phs, nph)
			} else {
				phs = append(phs, setter.Phases[i])
			}
		}

	}
	setter.Phases = phs
	buffer, err := json.Marshal(setter)
	if err != nil {
		return message
	}
	result.Body = string(buffer)
	return result
}
