package device

import (
	"encoding/json"
	"sort"

	"github.com/ruraomsk/tulagate/controller"
)

func (d *Device) insertMGR(message controller.MessageFromAmi) controller.MessageFromAmi {
	result := controller.MessageFromAmi{Action: message.Action}
	var setter controller.StartCoordination
	err := json.Unmarshal([]byte(message.Body), &setter)
	if err != nil {
		return message
	}
	size := len(setter.Phases)
	for i := 0; i < size; i++ {
		setter.Phases[i].Phase_order = setter.Phases[i].Phase_order * 10
	}
	sort.Slice(setter.Phases, func(i, j int) bool {
		return setter.Phases[i].Phase_order < setter.Phases[j].Phase_order
	})
	for i := 0; i < size; i++ {
		m, is := d.MGRS[setter.Phases[i].Phase_number]
		if !is {
			continue
		}
		if setter.Phases[i].Phase_duration > (m.TLen + m.TMGR) {
			//Можно вставить МГР
			nph := controller.Phase{Phase_number: 0, Phase_order: setter.Phases[i].Phase_order + 1, Phase_duration: m.TMGR}
			setter.Phases[i].Phase_duration = setter.Phases[i].Phase_duration - m.TMGR
			setter.Phases = append(setter.Phases, nph)
		}
	}
	buffer, err := json.Marshal(setter)
	if err != nil {
		return message
	}
	result.Body = string(buffer)
	return result

}
