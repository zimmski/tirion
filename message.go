package tirion

import (
	"time"
)

// Message contains all common data of a Tirion message.
type Message struct {
	Time time.Time
}

// MessageData contains all data of data message.
type MessageData struct {
	Message
	Data []float32
}

// MessageReturnInsert contains all data of the result of an Insert call.
type MessageReturnInsert struct {
	Error string
}

// MessageReturnStart contains all data of the result of a Start call.
type MessageReturnStart struct {
	Run   int32
	Error string
}

// MessageReturnStop contains all data of the result of a Stop call.
type MessageReturnStop struct {
	Error string
}

// MessageReturnTag contains all data of the result of a Tag call.
type MessageReturnTag struct {
	Error string
}

// MessageTag contains all data of tag message.
type MessageTag struct {
	Message
	Tag string
}
