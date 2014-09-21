package backend

import (
	"fmt"

	"github.com/zimmski/tirion"
)

type Backend interface {
	Init(params Parameters) error

	SearchPrograms() ([]tirion.Program, error)

	FindRun(programName string, runID int32) (*tirion.Run, error)
	SearchRuns(programName string) ([]tirion.Run, error)
	StartRun(run *tirion.Run) error
	StopRun(runID int32) error

	CreateMetrics(runID int32, metrics []tirion.MessageData) error
	SearchMetricOfRun(run *tirion.Run, metric string) ([][]interface{}, error)
	SearchMetricsOfRun(run *tirion.Run) ([][]float32, error)

	CreateTag(runID int32, tag *tirion.Tag) error
	SearchTagsOfRun(run *tirion.Run) ([]tirion.HighStockTag, error)
}

type Parameters struct {
	Spec         string
	MaxIdleConns int
	MaxOpenConns int
}

func NewBackend(name string) (Backend, error) {
	if name == "postgresql" {
		return NewBackendPostgresql(), nil
	}

	return nil, fmt.Errorf("unknown backend \"%s\"", name)
}
