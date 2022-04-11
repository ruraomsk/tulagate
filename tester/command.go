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

func TestCommand() {

	time.Sleep(20 * time.Second)
	senderCommand("device3", "SwitchProgram", "{\"program_number\":1,\"switch_default\":true}")
	time.Sleep(60 * time.Second)
	senderCommand("device3", "SwitchProgram", "{\"program_number\":2,\"switch_default\":true}")
	time.Sleep(60 * time.Second)
	senderCommand("device3", "SwitchProgram", "{\"program_number\":0,\"switch_default\":false}")
	time.Sleep(60 * time.Second)

	senderCommand("device3", "HoldPhase", "{\"phase_number\":1,\"max_duration\":45,\"unhold_phase\":true}")
	time.Sleep(60 * time.Second)
	senderCommand("device3", "HoldPhase", "{\"phase_number\":2,\"max_duration\":45,\"unhold_phase\":true}")
	time.Sleep(60 * time.Second)
	senderCommand("device3", "HoldPhase", "{\"phase_number\":3,\"max_duration\":45,\"unhold_phase\":true}")
	time.Sleep(60 * time.Second)
	senderCommand("device3", "HoldPhase", "{\"phase_number\":4,\"max_duration\":45,\"unhold_phase\":true}")
	time.Sleep(60 * time.Second)
	senderCommand("device3", "HoldPhase", "{\"phase_number\":4,\"max_duration\":45,\"unhold_phase\":false}")
	time.Sleep(60 * time.Second)

	senderCommand("device3", "SetMode", "{\"mode\":3,\"is_enabled\":true}")
	time.Sleep(60 * time.Second)
	senderCommand("device3", "SetMode", "{\"mode\":4,\"is_enabled\":true}")
	time.Sleep(60 * time.Second)
	senderCommand("device3", "SetMode", "{\"mode\":5,\"is_enabled\":true}")
	time.Sleep(60 * time.Second)
	senderCommand("device3", "SetMode", "{\"mode\":0,\"is_enabled\":false}")
	time.Sleep(60 * time.Second)

}
func senderCommand(id string, action string, body string) error {
	ch, err := db.GetChanelForMessage(id)
	if err != nil {
		logger.Error.Print(err.Error())
	}
	message := controller.MessageFromAmi{Action: action, Body: body}
	logger.Debug.Printf("From server %v", message)
	// time.Sleep(5 * time.Second)
	ch <- message
	return nil
}
