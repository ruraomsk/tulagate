package setup

var (
	Set *Setup
)

type Setup struct {
	Region            int      `toml:"region"`
	LogPath           string   `toml:"logpath"`
	RemoteHost        string   `toml:"remoteHost"`
	RemotePort        int      `toml:"remotePort"`
	DataBase          DataBase `toml:"dataBase"`
	AgServer          AgServer `toml:"agserver"`
	MyName            string   `toml:"myname"`
	TimeKeepAliveAmi  int      `toml:"keepalive"`
	MGRSet            bool     `toml:"mgrset"`
	TimeKeepStatistic int      `toml:"timestat"` //Интервал времени отстутствия статистики
	Test              bool     `toml:"test"`
}

//DataBase настройки базы данных postresql
type DataBase struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DBname   string `toml:"dbname"`
	Step     int    `toml:"step"`
}

//CommServer настройки для сервера коммуникации
type AgServer struct {
	Host        string `toml:"host"`
	PortCommand int    `toml:"portc"` //Порт приема команд от сервера АРМ
	PortArray   int    `toml:"porta"` //Порт приема массивов привязки от сервера АРМ
	PortDevices int    `toml:"portd"` //Порт передачи номера фазы и времени фазы серверу АРМ
}

func init() {
}
