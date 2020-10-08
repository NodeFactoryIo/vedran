package configuration

type Configuration struct {
	AuthSecret string
	Name       string
	Capacity   int64
	Whitelist  []string
	Fee        float32
	Selection  string
	Port       int32
}

var Config Configuration
