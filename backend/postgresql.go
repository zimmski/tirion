package backend

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/jbarham/gopgsqldriver"
	"github.com/zimmski/tirion"
)

type Postgresql struct {
	Db *sql.DB
}

func (p *Postgresql) Init() error {
	var err error

	p.Db, err = sql.Open("postgres", "user=zimmski dbname=tirion sslmode=disable")

	if err != nil {
		panic(fmt.Sprintf("Cannot connect to database: %v", err))
	}

	err = p.Db.Ping()

	if err != nil {
		panic(fmt.Sprintf("Cannot ping database: %v", err))
	}

	p.Db.SetMaxIdleConns(100)

	return nil
}

func (p *Postgresql) SearchPrograms() ([]tirion.Program, error) {
	tx, err := p.Db.Begin()

	if err != nil {
		return nil, err
	}

	programs := make([]tirion.Program, 0)

	rows, err := tx.Query("SELECT name FROM run GROUP BY name ORDER BY name")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var program = tirion.Program{}

		if err := rows.Scan(&program.Name); err != nil {
			return nil, err
		}

		programs = append(programs, program)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	return programs, nil
}

func (p *Postgresql) FindRun(programName string, runId int) (*tirion.Run, error) {
	tx, err := p.Db.Begin()

	if err != nil {
		return nil, err
	}

	row := tx.QueryRow("SELECT id, name, sub_name, interval, metrics, prog, prog_arguments, extract(epoch from start), extract(epoch from stop) FROM run WHERE name = $1 and id = $2", programName, runId)

	var run = tirion.Run{}
	var metrics, start string
	var stop *string

	if err := row.Scan(&run.Id, &run.Name, &run.SubName, &run.Interval, &metrics, &run.Prog, &run.ProgArguments, &start, &stop); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	json.Unmarshal([]byte(metrics), &run.Metrics)

	var sta, _ = strconv.ParseFloat(start, 64)
	var stat = time.Unix(int64(sta), 0)
	run.Start = &stat

	if stop != nil {
		var sto, _ = strconv.ParseFloat(*stop, 64)
		var stot = time.Unix(int64(sto), 0)
		run.Stop = &stot
	} else {
		run.Stop = nil
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	return &run, nil
}
func (p *Postgresql) SearchRuns(programName string) ([]tirion.Run, error) {
	tx, err := p.Db.Begin()

	if err != nil {
		return nil, err
	}

	runs := make([]tirion.Run, 0)

	rows, err := tx.Query("SELECT id, name, sub_name, interval, prog, prog_arguments, extract(epoch from start), extract(epoch from stop) FROM run WHERE name = $1 ORDER BY start desc", programName)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var run = tirion.Run{}

		var start, stop *string

		if err := rows.Scan(&run.Id, &run.Name, &run.SubName, &run.Interval, &run.Prog, &run.ProgArguments, &start, &stop); err != nil {
			return nil, err
		}

		var sta, _ = strconv.ParseFloat(*start, 64)
		var stat = time.Unix(int64(sta), 0)
		run.Start = &stat

		if stop != nil {
			var sto, _ = strconv.ParseFloat(*stop, 64)
			var stot = time.Unix(int64(sto), 0)
			run.Stop = &stot
		} else {
			run.Stop = nil
		}

		runs = append(runs, run)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	return runs, nil
}
func (p *Postgresql) StartRun(run *tirion.Run) error {
	tx, err := p.Db.Begin()

	if err != nil {
		return err
	}

	var metrics, _ = json.Marshal(run.Metrics)

	err = tx.QueryRow("INSERT INTO run(name, sub_name, interval, metrics, prog, prog_arguments, start) VALUES($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP) RETURNING id", run.Name, run.SubName, run.Interval, string(metrics), run.Prog, run.ProgArguments).Scan(&run.Id)

	if err != nil {
		return err
	}

	var columns = make([]string, len(run.Metrics))

	for i, m := range run.Metrics {
		var n = strings.Replace(m.Name, ".", "_", -1)

		switch m.Type {
		case "float":
			columns[i] = n + " REAL NOT NULL"
		default:
			columns[i] = n + " " + m.Type + " NOT NULL"
		}
	}

	_, err = tx.Exec("CREATE TABLE r" + strconv.FormatInt(int64(run.Id), 10) + "(t TIMESTAMP NOT NULL, " + strings.Join(columns, ",") + ", PRIMARY KEY(t))")

	if err != nil {
		return err
	}

	_, err = tx.Exec("CREATE TABLE rt" + strconv.FormatInt(int64(run.Id), 10) + "(t TIMESTAMP NOT NULL, message TEXT NOT NULL, PRIMARY KEY(t))")

	if err != nil {
		return err
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}
func (p *Postgresql) StopRun(runId int) error {
	tx, err := p.Db.Begin()

	if err != nil {
		return err
	}

	rows, err := tx.Query("SELECT id FROM run WHERE id = $1", runId)

	if err != nil {
		panic(err)
	} else if !rows.Next() {
		panic(errors.New("Cannot find run"))
	}

	defer rows.Close()

	_, err = tx.Exec("UPDATE run SET stop = CURRENT_TIMESTAMP WHERE id = $1", runId)

	if err != nil {
		panic(err)
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}

func (p *Postgresql) CreateMetrics(runId int, metrics []tirion.MessageData) error {
	tx, err := p.Db.Begin()

	if err != nil {
		return err
	}

	rows, err := tx.Query("SELECT id FROM run WHERE id = $1", runId)

	if err != nil {
		panic(err)
	} else if !rows.Next() {
		panic(errors.New("Cannot find run"))
	}

	defer rows.Close()

	for _, m := range metrics {
		var ffff bytes.Buffer

		ffff.WriteString("INSERT INTO r" + strconv.FormatInt(int64(runId), 10) + " VALUES(TO_TIMESTAMP($1)")

		for _, i := range m.Data {
			ffff.WriteString("," + strconv.FormatFloat(float64(i), 'f', 5, 32))
		}

		ffff.WriteString(")")

		_, err = tx.Exec(ffff.String(), float64(m.Time.UnixNano())/1000000000.0)

		if err != nil {
			return err
		}
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}
func (p *Postgresql) SearchMetricOfRun(run *tirion.Run, metricName string) ([][]interface{}, error) {
	var found = false

	for _, m := range run.Metrics {
		if m.Name == metricName {
			found = true

			break
		}
	}

	if !found {
		return nil, errors.New("metric name not found")
	}

	tx, err := p.Db.Begin()

	if err != nil {
		return nil, err
	}

	metrics := make([][]interface{}, 0)

	rows, err := tx.Query("SELECT EXTRACT(EPOCH FROM t) * 1000.0, " + strings.Replace(metricName, ".", "_", -1) + " FROM r" + strconv.FormatInt(int64(run.Id), 10) + " ORDER BY t")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var metric = make([]interface{}, 2)

		var m float32
		var tt float64

		if err := rows.Scan(&tt, &m); err != nil {
			return nil, err
		}

		var t = int64(tt)

		metric[0] = &t
		metric[1] = &m

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	return metrics, nil
}
func (p *Postgresql) SearchMetricsOfRun(run *tirion.Run) ([][]float32, error) {
	tx, err := p.Db.Begin()

	if err != nil {
		return nil, err
	}

	pointers := make([]interface{}, len(run.Metrics)+1)
	metrics := make([][]float32, 0)

	var t string

	rows, err := tx.Query("SELECT * FROM r" + strconv.FormatInt(int64(run.Id), 10) + " ORDER BY t")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var metric = make([]float32, len(run.Metrics)+1)

		for i := range pointers {
			pointers[i] = &metric[i]
		}

		pointers[0] = &t

		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}

		tt, err := time.Parse("2006-01-02 15:04:05.9999", t)

		if err != nil {
			return nil, err
		}

		metric[0] = float32(tt.UnixNano()) / 1000000000.0

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (p *Postgresql) CreateTag(runId int, tag *tirion.Tag) error {
	tx, err := p.Db.Begin()

	if err != nil {
		return err
	}

	rows, err := tx.Query("SELECT id FROM run WHERE id = $1", runId)

	if err != nil {
		panic(err)
	} else if !rows.Next() {
		panic(errors.New("Cannot find run"))
	}

	defer rows.Close()

	_, err = tx.Exec("INSERT INTO rt"+strconv.FormatInt(int64(runId), 10)+"(t, message) VALUES(TO_TIMESTAMP($1), $2)", float64(tag.Time.UnixNano())/1000000000.0, tag.Tag)

	if err != nil {
		panic(err)
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}
func (p *Postgresql) SearchTagsOfRun(run *tirion.Run) ([]tirion.Tag, error) {
	tx, err := p.Db.Begin()

	if err != nil {
		return nil, err
	}

	tags := make([]tirion.Tag, 0)
	var t string

	rows, err := tx.Query("SELECT extract(epoch from t), message FROM rt" + strconv.FormatInt(int64(run.Id), 10) + " ORDER BY t")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var tag = tirion.Tag{}

		if err := rows.Scan(&t, &tag.Tag); err != nil {
			return nil, err
		}

		var tt, _ = strconv.ParseFloat(t, 64)
		tag.Time = time.Unix(int64(tt), 0)

		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	return tags, nil
}
