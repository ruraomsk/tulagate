package device

import (
	"encoding/json"
	"time"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/tulagate/controller"
	"github.com/ruraomsk/tulagate/uptransport"
)

func (d *Device) executeGetCoordination() []controller.Programm {
	result := make([]controller.Programm, 0)
	for _, v := range d.Cross.Arrays.SetDK.DK {
		plan := controller.Programm{Number: v.Pk, Offset: v.Shift, Phases: make([]controller.Phase, 0), Is_Coordination: false}
		if v.TypePU == 1 {
			plan.Is_Coordination = true
		}
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
func (d *Device) executeConfig(message controller.MessageFromAmi) string {
	c := controller.Config{Programs: d.executeGetCoordination(),
		Cards: make([]controller.DailyCard, 0),
		Weeks: make([]controller.Week, 0)}
	for _, v := range d.Cross.Arrays.WeekSets.WeekSets {
		if !isWeekEmpty(v) {
			c.Weeks = append(c.Weeks, controller.Week{Number: v.Number, DailyCards: v.Days})
		}
	}
	for _, v := range d.Cross.Arrays.DaySets.DaySets {
		if v.Count != 0 {
			crd := controller.DailyCard{Number: v.Number, Programs: make([]controller.Line, 0)}
			for _, l := range v.Lines {
				if l.Hour != 0 || l.Min != 0 {
					crd.Programs = append(crd.Programs, controller.Line{Number: l.PKNom, Hour: l.Hour, Minute: l.Min})
				}
			}
			c.Cards = append(c.Cards, crd)
		}
	}
	replay := controller.MessageToAmi{IDExternal: d.OneSet.IDExternal, Action: "Config", Body: "{}"}
	body, _ := json.Marshal(&c)
	replay.Body = string(body)
	d.LastSendStatus = time.Now()
	logger.Debug.Println(replay)
	uptransport.SendToAmiChan <- replay
	return "ok"
}
