package tcp

import "time"

type Config struct {
	Port            uint16
	ShutdownTimeout time.Duration
	Callbacks       ServerCallbacks
}
