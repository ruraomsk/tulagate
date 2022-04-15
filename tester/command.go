package tester

import (
	"time"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/tulagate/controller"
	"github.com/ruraomsk/tulagate/db"
)

// type SetMode struct {
// 	Mode       int  `json:"mode"`       // 3 - желтое мигание, 4 - кругом красный, 5 - все светофоры выключенны
// 	Is_enabled bool `json:"is_enabled"` // Включить / Выключить
// }
//HoldPhase Включение удержания заданной фазы. Переводит поле mode в структуре ответа на действие Status в значение "удержание фазы". Инициатор действия сервер. В теле запроса приходит следующая структура:
// type HoldPhase struct {
// 	Phase_number int  `json:"phase_number"` // Номер фазы
// 	Max_duration int  `json:"max_duration"` // Максимальное время удержания фазы в секундах
// 	Unhold_phase bool `json:"unhold_phase"` // Флаг снятия удержания фазы
// }
//SwitchProgram  Установка программы на контроллере. Инициатор действия сервер. В теле запроса приходит следующая структура:
// type SwitchProgram struct {
// 	Programm_number int  `json:"program_number"` // Номер программы
// 	Switch_default  bool `json:"switch_default"` // Флаг установки значения по умолчанию
// }
//StartCoordination Включение плана координации. Инициатор действия сервер. В теле запроса приходит следующая структура:
// type StartCoordination struct {
// 	Programm_number int     `json:"program_number"` // Номер программы
// 	Phases          []Phase `json:"phases"`         // Список фаз
// 	Offset          int     `json:"offset"`         // Смещение начала программы в сек
// 	IsEnabled       bool    `json:"isEnabled"`      // Вкл / Выкл
// }

// type Phase struct {
// 	Phase_number   int `json:"phase_number"`   //Номер фазы
// 	Phase_duration int `json:"phase_duration"` //Время фазы в секундах
// 	Phase_order    int `json:"phase_order"`    //Порядок фазы в программе
// 	Max_time       int `json:"max_time"`       //Максимальная граница
// 	Min_time       int `json:"min_time"`       //Минимальная граница
// }
var pl = `{
	"program_number":1,
	"offset":10,
	"isEnabled":true,
	"phases":[
		{
			"phase_number":1,
			"phase_duration":20,
			"phase_order":1
		},
		{
			"phase_number":2,
			"phase_duration":20,
			"phase_order":2
		},
		{
			"phase_number":1,
			"phase_duration":30,
			"phase_order":3
		},
		{
			"phase_number":2,
			"phase_duration":30,
			"phase_order":4
		}
	]
}`
var pl1 = `{
	"program_number":1,
	"offset":20,
	"isEnabled":true,
	"phases":[
		{
			"phase_number":1,
			"phase_duration":25,
			"phase_order":1
		},
		{
			"phase_number":2,
			"phase_duration":25,
			"phase_order":2
		}
	]
}`

var sl = `{
	"program_number":1,
	"offset":0,
	"isEnabled":false,
	"phases":[]
}
`

func TestCommand() {
	for {
		time.Sleep(5 * time.Second)

		senderCommand("device3", "StartCoordination", pl)
		senderCommand("device5", "StartCoordination", pl)
		senderCommand("device3", "SwitchProgram", `{"program_number":1,"switch_default":true}`)
		senderCommand("device5", "SwitchProgram", `{"program_number":1,"switch_default":true}`)
		time.Sleep(300 * time.Second)

		senderCommand("device3", "HoldPhase", `{"phase_number":1,"max_duration":70,"unhold_phase":true}`)
		senderCommand("device5", "HoldPhase", `{"phase_number":1,"max_duration":70,"unhold_phase":true}`)
		time.Sleep(120 * time.Second)
		senderCommand("device3", "HoldPhase", `{"phase_number":2,"max_duration":75,"unhold_phase":true}`)
		senderCommand("device5", "HoldPhase", `{"phase_number":2,"max_duration":75,"unhold_phase":true}`)
		time.Sleep(120 * time.Second)

		senderCommand("device3", "SwitchProgram", `{"program_number":1,"switch_default":true}`)
		senderCommand("device5", "SwitchProgram", `{"program_number":1,"switch_default":true}`)
		time.Sleep(60 * time.Second)
		senderCommand("device3", "SwitchProgram", `{"program_number":2,"switch_default":true}`)
		senderCommand("device5", "SwitchProgram", `{"program_number":2,"switch_default":true}`)
		time.Sleep(60 * time.Second)
		senderCommand("device3", "SwitchProgram", `{"program_number":0,"switch_default":false}`)
		senderCommand("device5", "SwitchProgram", `{"program_number":0,"switch_default":false}`)
		time.Sleep(60 * time.Second)

		senderCommand("device3", "SetMode", `{"mode":3,"is_enabled":true}`)
		senderCommand("device5", "SetMode", `{"mode":3,"is_enabled":true}`)
		time.Sleep(60 * time.Second)
		senderCommand("device3", "SetMode", `{"mode":4,"is_enabled":true}`)
		senderCommand("device5", "SetMode", `{"mode":4,"is_enabled":true}`)
		time.Sleep(60 * time.Second)
		senderCommand("device3", "SetMode", `{"mode":5,"is_enabled":true}`)
		senderCommand("device5", "SetMode", `{"mode":5,"is_enabled":true}`)
		time.Sleep(60 * time.Second)
		senderCommand("device3", "SetMode", `{"mode":0,"is_enabled":false}`)
		senderCommand("device5", "SetMode", `{"mode":0,"is_enabled":false}`)
		time.Sleep(60 * time.Second)

		senderCommand("device3", "StartCoordination", pl1)
		senderCommand("device5", "StartCoordination", pl1)
		time.Sleep(600 * time.Second)

		senderCommand("device3", "StartCoordination", sl)
		senderCommand("device5", "StartCoordination", sl)

	}

}
func senderCommand(id string, action string, body string) error {
	ch, err := db.GetChanelForMessage(id)
	if err != nil {
		logger.Error.Print(err.Error())
	}
	message := controller.MessageFromAmi{Action: action, Body: body}
	ch <- message
	return nil
}
