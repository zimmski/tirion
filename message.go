package tirion

import (
	"time"
)

type Message struct {
	Time time.Time
}

type MessageData struct {
	Message
	Data []float32
}

type MessageReturnInsert struct {
	Error string
}

type MessageReturnStart struct {
	Run   int32
	Error string
}

type MessageReturnStop struct {
	Error string
}

type MessageReturnTag struct {
	Error string
}

type MessageTag struct {
	Message
	Tag string
}
