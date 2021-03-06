package collector

import (
	"fmt"
	"net/url"
)

type Collector interface {
	InitAgent(pid int32, metricCount int32) (*url.URL, error)
	InitClient(u *url.URL, metricCount int32) error
	Data() []float32
	Close() error

	Get(i int32) float32
	Set(i int32, v float32) float32

	Add(i int32, v float32) float32
	Dec(i int32) float32
	Inc(i int32) float32
	Sub(i int32, v float32) float32
}

func NewCollector(typ string) (Collector, error) {
	switch typ {
	case "mmap":
		return new(CollectorMmap), nil
	case "shm":
		return new(CollectorShm), nil
	default:
		return nil, fmt.Errorf("unknown metric protocol \"%s\"", typ)
	}
}
