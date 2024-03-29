package controller

// Status Периодическое (например, раз в сек) получение текущего статуса контроллера. Инициатор действия клиент.В теле ответа возвращается структура:
type Status struct {
	Program_number         int     `json:"program_number"`         // Номер программы
	Phase_number           int     `json:"phase_number"`           // Номер фазы
	Tact_number            int     `json:"tact_number"`            // Номер такта в фазе
	Tact_tick              int     `json:"tact_tick"`              // Тик такта
	Hold_phase_number      int     `json:"hold_phase_number"`      // Номер фазы для удержания
	Host                   string  `json:"host"`                   // IP адрес подключенного устройства
	Port                   int     `json:"port"`                   // Порт подключенного устройства
	State                  int     `json:"state"`                  // Статус контроллера в сети: 0 - нет связи, 1 - есть связь нет управления, 2 - есть связь есть управление
	Errors                 Errors  `json:"errors"`                 // Описание ошибок
	Mode                   int     `json:"mode"`                   // Текущий режим: 1 - локальный, 2 - удаленное (центральное) управление, 3 - желтое мигание, 4 - кругом красный, 5 - все светофоры выключенны, 6 - ручное (местное) управление, 7 - удержание фазы, 8 - диспетчерское управление
	Timestamp              int64   `json:"timestamp"`              // Время на контроллере
	Has_Default_Programs   []int   `json:"has_default_programs"`   //Список номеров ПК загруженных верхом
	Has_Loaded_Programs    []int   `json:"has_loaded_programs"`    //Список номеров всех ПК
	VAC                    float64 `json:"vac"`                    // Текущее напряжение на контроллере
	Hold_phase_time_remain int     `json:"hold_phase_time_remain"` // Таймер обратного отсчета до конца удержания фазы, в сек
	Has_loaded_daily_cards []int   `json:"has_loaded_daily_cards"` //Перечень загруженных суточных карт
	Has_loaded_week_cards  []int   `json:"has_loaded_week_cards"`  //Перечень загруженных недельных карт
	Is_transition          bool    `json:"is_transition"`          //Признак перехода
}

// Errors Описание ошибок
type Errors struct {
	Hw_error       []string `json:"hw_error"`       // Массив ошибок, связанных с железом
	Sw_error       []string `json:"sw_error"`       // Массив ошибок, связанных с ПО
	Ec_error       []string `json:"ec_error"`       // Массив ошибок, связанных с электрикой (например, кз)
	Detector_fault []string `json:"detector_fault"` // Массив ошибок, связанных с детектором
	Is_door_opened bool     `json:"is_door_opened"` // Состояние открытия/закрытия двери
}

// SetMode Установка режима работы контроллера. Инициатор действия сервер. В теле запроса приходит следующая структура:
type SetMode struct {
	Mode       int  `json:"mode"`       // 3 - желтое мигание, 4 - кругом красный, 5 - все светофоры выключенны
	Is_enabled bool `json:"is_enabled"` // Включить / Выключить
}
type ChanelStat struct {
	Chanels [16]int `json:"chanels"`
}

// HoldPhase Включение удержания заданной фазы. Переводит поле mode в структуре ответа на действие Status в значение "удержание фазы". Инициатор действия сервер. В теле запроса приходит следующая структура:
type HoldPhase struct {
	Phase_number int  `json:"phase_number"` // Номер фазы
	Max_duration int  `json:"max_duration"` // Максимальное время удержания фазы в секундах
	Unhold_phase bool `json:"unhold_phase"` // Флаг снятия удержания фазы
}

// SwitchProgram  Установка программы на контроллере. Инициатор действия сервер. В теле запроса приходит следующая структура:
type SwitchProgram struct {
	Programm_number int  `json:"program_number"` // Номер программы
	Switch_default  bool `json:"switch_default"` // Флаг установки значения по умолчанию
}

type Programm struct {
	Number          int     `json:"number"`          // номер прогрммы
	Offset          int     `json:"offset"`          // смещение прогрммы (скорее всего тут это не нужно, так как для каждого плана координации смещение будет свое)
	Is_Coordination bool    `json:"is_coordination"` //True если ПК иначе ЛР
	IsDefault       bool    `json:"is_default"`      // признак того, используется прогрмма по умолчанию или нет
	Mode            int     `json:"mode"`            // 0 обычная программа 1 ЖМ 2 ОС
	Phases          []Phase `json:"phases"`          // массив фаз
}

type Phase struct {
	Number   int `json:"number"`   // номер фазы
	Duration int `json:"duration"` // общая длительность фазы, ключая длительность всех промтактов
	Type     int `json:"mode"`     // 0 обычная фаза 1 фаза с ТВП 2 Замещающая фаза
	// TLGroups []Group `json:"tl_groups"` // массив групп светофоров
}

type Group struct {
	Number int        `json:"number"`     // номер группы светофоров
	Color  int        `json:"color"`      // цвет группы светофоров в фазе
	PTs    []PromTact `json:"prom_tacts"` // массив промтактов
}

type PromTact struct {
	Number int `json:"number"` // номер промтакта
	Color  int `json:"color"`  // цвет промтакта
}
type ErrorString struct {
	Message string `json:"error"`
}
type MessageFromAmi struct {
	Action string
	Body   string
}
type MessageToAmi struct {
	IDExternal string
	Action     string
	Body       string
}

// UploadDailyCards -загрузка суточных карт
type UploadDailyCards struct {
	Cards []DailyCard `json:"daily_cards"`
}

// DailyCard -собственно суточная карта
type DailyCard struct {
	Number   int    `json:"number"`
	Programs []Line `json:"programs"`
}

// Line одно переключение суточной карты
type Line struct {
	Number int `json:"number"`
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

// UploadWeekCards загрузка недельных карт
type UploadWeekCards struct {
	Weeks []Week `json:"week_cards"`
}

// Week собственно недельная карта
type Week struct {
	Number     int   `json:"number"`
	DailyCards []int `json:"daily_cards"`
}
type Config struct {
	Programs []Programm  `json:"programs"`
	Cards    []DailyCard `json:"daily_cards"`
	Weeks    []Week      `json:"week_cards"`
}
