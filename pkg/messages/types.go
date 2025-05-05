package messages

import "github.com/Knoblauchpilze/chat-server/pkg/persistence"

type Processor interface {
	Start() error
	Stop() error

	Enqueue(msg persistence.Message)
}

type StartCallback func() error
type MessageCallback func(msg persistence.Message) error
type FinishCallback func() error

type Callbacks struct {
	Start   StartCallback
	Message MessageCallback
	Finish  FinishCallback
}
