package collector

import (
	"errors"
	"fmt"
	"net/url"
)

type Collector interface {
	InitAgent(pid int32, metricCount int32) (*url.URL, error)
	InitClient(u *url.URL, metricCount int32) error
	Data() []float32
	Close() error

	Add(i int32, v float32) float32
	Dec(i int32) float32
	Inc(i int32) float32
	Sub(i int32, v float32) float32
}

func NewCollector(typ string) (Collector, error) {
	switch typ {
	case "shm":
		return new(CollectorShm), nil
	default:
		return nil, errors.New(fmt.Sprintf("Unknown metric protocol \"%s\"", typ))
	}
}
