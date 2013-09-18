package backend

import (
	"fmt"

	"github.com/zimmski/tirion"
)

type Backend interface {
	Init() error

	SearchPrograms() ([]tirion.Program, error)

	FindRun(runId int) (*tirion.Run, error)
	SearchRuns(programName string) ([]tirion.Run, error)
	StartRun(run *tirion.Run) error
	StopRun(runId int) error

	CreateMetrics(runId int, metrics []tirion.MessageData) error
	SearchMetricOfRun(run *tirion.Run, metric string) ([][]interface{}, error)
	SearchMetricsOfRun(run *tirion.Run) ([][]float32, error)

	CreateTag(runId int, tag *tirion.Tag) error
	SearchTagsOfRun(run *tirion.Run) ([]tirion.Tag, error)
}

func NewBackend(name string) Backend {
	if name == "postgresql" {
		return new(Postgresql)
	} else {
		panic(fmt.Sprintf("Unknown backend \"%s\""))
	}
}
