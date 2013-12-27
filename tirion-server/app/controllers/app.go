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

func (c *App) Index() revel.Result {
	programs, err := app.Db.SearchPrograms()

	if err != nil {
		panic(err)
	}

	return c.Render(programs)
}

func (c *App) ProgramIndex(programName string) revel.Result {
	runs, err := app.Db.SearchRuns(programName)

	if err != nil {
		panic(err)
	}

	if len(runs) == 0 {
		return c.NotFound("Program \"%s\" does not exists", programName)
	}

	return c.Render(programName, runs)
}

func (c *App) ProgramRunIndex(programName string, runID int32) revel.Result {
	run, err := app.Db.FindRun(programName, runID)

	if err != nil {
		panic(err)
	} else if run == nil {
		return c.NotFound("Run %d of program \"%s\" does not exists", runID, programName)
	}

	return c.Render(programName, run)
}

func (c *App) ProgramRunMetric(programName string, runID int32, metricName string) revel.Result {
	run, err := app.Db.FindRun(programName, runID)

	if err != nil {
		panic(err)
	} else if run == nil {
		return c.NotFound("Run %d of program \"%s\" does not exists", runID, programName)
	}

	metric, err := app.Db.SearchMetricOfRun(run, metricName)

	if err != nil {
		panic(err)
	}

	return c.RenderJson(metric)
}

func (c *App) ProgramRunStart(programName string) revel.Result {
	var interval, err = strconv.ParseInt(c.Params.Get("interval"), 10, 32)

	if err != nil {
		return c.RenderJson(tirion.MessageReturnStart{Error: fmt.Sprintf("Cannot parse interval: %v", err)})
	}

	var run = tirion.Run{
		Name:          c.Params.Get("name"),
		SubName:       c.Params.Get("sub_name"),
		Interval:      int32(interval),
		Prog:          c.Params.Get("prog"),
		ProgArguments: c.Params.Get("prog_arguments"),
	}

	if run.Name == "" {
		return c.RenderJson(tirion.MessageReturnStart{Error: fmt.Sprintf("No name defined")})
	}

	err = json.Unmarshal([]byte(c.Params.Get("metrics")), &run.Metrics)

	if err != nil {
		return c.RenderJson(tirion.MessageReturnStart{Error: fmt.Sprintf("Parse metrics file: %v", err)})
	}

	if err := tirion.CheckMetrics(run.Metrics); err != nil {
		return c.RenderJson(tirion.MessageReturnStart{Error: err.Error()})
	}

	err = app.Db.StartRun(&run)
	if err != nil {
		return c.RenderJson(tirion.MessageReturnStart{Error: fmt.Sprintf("%+v", err)})
	}

	return c.RenderJson(tirion.MessageReturnStart{Run: run.ID, Error: ""})
}

func (c *App) ProgramRunInsert(programName string, runID int32) revel.Result {
	var metrics []tirion.MessageData

	var err = json.Unmarshal([]byte(c.Params.Get("metrics")), &metrics)

	if err != nil {
		return c.RenderJson(tirion.MessageReturnStart{Error: fmt.Sprintf("Parse metrics: %v", err)})
	}

	err = app.Db.CreateMetrics(runID, metrics)
	if err != nil {
		return c.RenderJson(tirion.MessageReturnInsert{Error: fmt.Sprintf("%+v", err)})
	}

	return c.RenderJson(tirion.MessageReturnInsert{Error: ""})
}

func (c *App) ProgramRunStop(programName string, runID int32) revel.Result {
	var err = app.Db.StopRun(runID)

	if err != nil {
		return c.RenderJson(tirion.MessageReturnStop{Error: fmt.Sprintf("%+v", err)})
	}

	return c.RenderJson(tirion.MessageReturnStop{Error: ""})
}

func (c *App) ProgramRunTag(programName string, runID int32) revel.Result {
	var t, err = strconv.ParseInt(c.Params.Get("time"), 10, 64)

	var tag = tirion.Tag{
		Tag:  tirion.PrepareTag(c.Params.Get("tag")),
		Time: time.Unix(0, t),
	}

	err = app.Db.CreateTag(runID, &tag)

	if err != nil {
		return c.RenderJson(tirion.MessageReturnTag{Error: fmt.Sprintf("%+v", err)})
	}

	return c.RenderJson(tirion.MessageReturnTag{Error: ""})
}

func (c *App) ProgramRunTags(programName string, runID int32) revel.Result {
	run, err := app.Db.FindRun(programName, runID)

	if err != nil {
		panic(err)
	} else if run == nil {
		return c.NotFound("Run %d of program \"%s\" does not exists", runID, programName)
	}

	tags, err := app.Db.SearchTagsOfRun(run)
	if err != nil {
		panic(err)
	}

	return c.RenderJson(tags)
}
