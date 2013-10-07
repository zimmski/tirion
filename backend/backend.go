package backend

import (
	"errors"
	"fmt"

	"github.com/zimmski/tirion"
)

type Backend interface {
	Init(spec string) error

	SearchPrograms() ([]tirion.Program, error)

	FindRun(programName string, runId int32) (*tirion.Run, error)
	SearchRuns(programName string) ([]tirion.Run, error)
	StartRun(run *tirion.Run) error
	StopRun(runId int32) error

	CreateMetrics(runId int32, metrics []tirion.MessageData) error
	SearchMetricOfRun(run *tirion.Run, metric string) ([][]interface{}, error)
	SearchMetricsOfRun(run *tirion.Run) ([][]float32, error)

	CreateTag(runId int32, tag *tirion.Tag) error
	SearchTagsOfRun(run *tirion.Run) ([]tirion.HighStockTag, error)
}

func NewBackend(name string) (Backend, error) {
	if name == "postgresql" {
		return NewBackendPostgresql(), nil
	} else {
		return nil, errors.New(fmt.Sprintf("Unknown backend \"%s\"", name))
	}
}
