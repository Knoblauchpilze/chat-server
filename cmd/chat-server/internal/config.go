package internal

type Configuration struct {
	Port uint16
}

func DefaultConfig() Configuration {
	return Configuration{
		Port: uint16(80),
	}
}
