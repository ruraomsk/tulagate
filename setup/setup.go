package setup

var (
	Set *Setup
)

type Setup struct {
	LogPath    string `toml:"logpath"`
	RemoteHost string `toml:"remoteHost"`
	RemotePort int    `toml:"remotePort"`
}

func init() {
}
