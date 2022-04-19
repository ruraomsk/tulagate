package device

import (
	"github.com/ruraomsk/tulagate/controller"
)

func (d *Device) executeGetCoordination() []controller.StartCoordination {
	result := make([]controller.StartCoordination, 0)
	for _, v := range d.Cross.Arrays.SetDK.DK {
		plan := controller.StartCoordination{Programm_number: v.Pk, Offset: v.Shift, Phases: make([]controller.Phase, 0)}
		plan.IsEnabled = true
		if v.Tc == 0 {
			plan.IsEnabled = false
		}
		phase := controller.Phase{Phase_order: 1}
		for _, ph := range v.Stages {
			if ph.Number == 0 && ph.Start == 0 && ph.Stop == 0 {
				break
			}
			if ph.Tf == 1 {
				// Значит у нас МГР
				phase = plan.Phases[len(plan.Phases)-1]
				phase.Phase_duration += ph.Stop - ph.Start + ph.Dt
				plan.Phases[len(plan.Phases)-1] = phase
				// logger.Debug.Print(plan.Phases)
				continue
			}
			phase = controller.Phase{Phase_order: ph.Nline, Phase_duration: ph.Stop - ph.Start + ph.Dt, Phase_number: ph.Number}
			plan.Phases = append(plan.Phases, phase)
		}
		result = append(result, plan)
	}
	return result
}
