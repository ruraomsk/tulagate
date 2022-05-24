package device

import (
	"github.com/ruraomsk/tulagate/controller"
)

func (d *Device) executeGetCoordination() []controller.Programm {
	result := make([]controller.Programm, 0)
	for _, v := range d.Cross.Arrays.SetDK.DK {
		plan := controller.Programm{Number: v.Pk, Offset: v.Shift, Phases: make([]controller.Phase, 0)}
		phase := controller.Phase{}
		for _, ph := range v.Stages {
			if ph.Number == 0 && ph.Start == 0 && ph.Stop == 0 {
				break
			}
			if ph.Tf == 1 {
				// Значит у нас МГР
				phase = plan.Phases[len(plan.Phases)-1]
				phase.Duration += ph.Stop - ph.Start + ph.Dt
				plan.Phases[len(plan.Phases)-1] = phase
				// logger.Debug.Print(plan.Phases)
				continue
			}
			phase = controller.Phase{Duration: ph.Stop - ph.Start + ph.Dt, Number: ph.Number}
			plan.Phases = append(plan.Phases, phase)
		}
		result = append(result, plan)
	}
	return result
}
