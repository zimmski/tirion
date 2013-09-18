package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/robfig/revel"
	"github.com/zimmski/tirion"
	"github.com/zimmski/tirion/tirion-server/app"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	programs, err := app.Db.SearchPrograms()

	if err != nil {
		panic(err)
	}

	return c.Render(programs)
}

func (c App) ProgramIndex(programName string) revel.Result {
	runs, err := app.Db.SearchRuns(programName)

	if err != nil {
		panic(err)
	}

	if len(runs) == 0 {
		return c.NotFound("Program \"%s\" does not exists", programName)
	}

	return c.Render(programName, runs)
}

func (c App) ProgramRunIndex(programName string, runId int) revel.Result {
	run, err := app.Db.FindRun(programName, runId)

	if err != nil {
		panic(err)
	} else if run == nil {
		return c.NotFound("Run %d of program \"%s\" does not exists", runId, programName)
	}

	return c.Render(programName, run)
}

func (c App) ProgramRunMetric(programName string, runId int, metricName string) revel.Result {
	run, err := app.Db.FindRun(programName, runId)

	if err != nil {
		panic(err)
	} else if run == nil {
		return c.NotFound("Run %d of program \"%s\" does not exists", runId, programName)
	}

	metric, err := app.Db.SearchMetricOfRun(run, metricName)

	if err != nil {
		panic(err)
	}

	return c.RenderJson(metric)
}

func (c App) ProgramRunStart(programName string) revel.Result {
	var interval, err = strconv.ParseInt(c.Params.Get("interval"), 10, 32)

	if err != nil {
		return c.RenderJson(tirion.MessageReturnStart{Error: fmt.Sprintf("Cannot parse interval: %v", err)})
	}

	var run = tirion.Run{
		Name:          c.Params.Get("name"),
		SubName:       c.Params.Get("sub_name"),
		Interval:      int(interval),
		Prog:          c.Params.Get("prog"),
		ProgArguments: c.Params.Get("prog_arguments"),
	}

	if run.Name == "" {
		return c.RenderJson(tirion.MessageReturnStart{Error: fmt.Sprintf("No name defined")})
	} else if run.Prog == "" {
		return c.RenderJson(tirion.MessageReturnStart{Error: fmt.Sprintf("No prog defined")})
	}

	err = json.Unmarshal([]byte(c.Params.Get("metrics")), &run.Metrics)

	if err != nil {
		return c.RenderJson(tirion.MessageReturnStart{Error: fmt.Sprintf("Parse metrics file: %v", err)})
	}

	if len(run.Metrics) == 0 {
		return c.RenderJson(tirion.MessageReturnStart{Error: fmt.Sprintf("No metrics defined")})
	}

	var metricNames = make(map[string]int)

	for i, m := range run.Metrics {
		if m.Name == "" {
			panic(fmt.Sprintf("No name defined for metric[%d]", i))
		} else if v, ok := metricNames[m.Name]; ok {
			panic(fmt.Sprintf("Name \"%s\" of metric[%d] alreay used for metric[%d]", m.Name, i, v))
		} else if m.Type == "" {
			panic(fmt.Sprintf("No type defined for metric[%d]", i))
		} else if _, ok := tirion.MetricTypes[m.Type]; !ok {
			panic(fmt.Sprintf("Unknown metric type \"%s\" for metric[%d]", m.Type, i))
		}

		metricNames[m.Name] = i
	}

	err = app.Db.StartRun(&run)

	if err != nil {
		return c.RenderJson(tirion.MessageReturnStart{Error: fmt.Sprintf("%+v", err)})
	} else {
		return c.RenderJson(tirion.MessageReturnStart{Run: run.Id, Error: ""})
	}
}

func (c App) ProgramRunInsert(programName string, runId int) revel.Result {
	var metrics []tirion.MessageData

	var err = json.Unmarshal([]byte(c.Params.Get("metrics")), &metrics)

	if err != nil {
		return c.RenderJson(tirion.MessageReturnStart{Error: fmt.Sprintf("Parse metrics: %v", err)})
	}

	err = app.Db.CreateMetrics(runId, metrics)

	if err != nil {
		return c.RenderJson(tirion.MessageReturnInsert{Error: fmt.Sprintf("%+v", err)})
	} else {
		return c.RenderJson(tirion.MessageReturnInsert{Error: ""})
	}
}

func (c App) ProgramRunStop(programName string, runId int) revel.Result {
	var err = app.Db.StopRun(runId)

	if err != nil {
		return c.RenderJson(tirion.MessageReturnStop{Error: fmt.Sprintf("%+v", err)})
	} else {
		return c.RenderJson(tirion.MessageReturnStop{Error: ""})
	}
}

func (c App) ProgramRunTag(programName string, runId int) revel.Result {
	var t, err = strconv.ParseInt(c.Params.Get("time"), 10, 64)

	var tag = tirion.Tag{
		Tag:  c.Params.Get("tag"),
		Time: time.Unix(0, t),
	}

	err = app.Db.CreateTag(runId, &tag)

	if err != nil {
		return c.RenderJson(tirion.MessageReturnStop{Error: fmt.Sprintf("%+v", err)})
	} else {
		return c.RenderJson(tirion.MessageReturnStop{Error: ""})
	}
}
