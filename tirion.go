package tirion

import (
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/zimmski/tirion/shm"
)

const Version = "0.1"

const TIRION_TAG_SIZE = 513

type HighStockTag struct {
	X     int64  `json:"x"`
	Title string `json:"title"`
}

type Metric struct {
	Name string
	Type string
}

var MetricTypes = map[string]bool{
	"int":   true,
	"float": true,
}

type Program struct {
	Name string
}

type Run struct {
	Id            int32
	Name          string
	SubName       string
	Interval      int32
	Metrics       []Metric
	MetricCount   int32
	Prog          string
	ProgArguments string
	Start         *time.Time
	Stop          *time.Time
}

type Tag struct {
	Time time.Time
	Tag  string
}

type Tirion struct {
	fd        net.Conn
	Running   bool
	shm       *shm.Shm
	socket    string
	verbose   bool
	logPrefix string
}

func CheckMetrics(metrics []Metric) error {
	if len(metrics) == 0 {
		return errors.New("No metrics defined")
	} else if len(metrics) >= math.MaxInt32 {
		return errors.New(fmt.Sprintf("Maximum of %d metrics allowed", math.MaxInt32))
	}

	var metricNameRegex = regexp.MustCompile("[^a-zA-Z0-9.-_]")
	var metricNames = make(map[string]int32)

	for i, m := range metrics {
		if m.Name == "" {
			return errors.New(fmt.Sprintf("No name defined for metric[%d]", i))
		} else if len(m.Name) > 256 {
			return errors.New(fmt.Sprintf("Name of metric[%d] exceeds maximum of 256 characters", i))
		} else if metricNameRegex.MatchString(m.Name) {
			return errors.New(fmt.Sprintf("Name  of metric[%d] uses illegal characters. Only a-z, A-Z, 0-9, ., - and _ are allowed!", i))
		} else if v, ok := metricNames[m.Name]; ok {
			return errors.New(fmt.Sprintf("Name \"%s\" of metric[%d] alreay used for metric[%d]", m.Name, i, v))
		} else if m.Type == "" {
			return errors.New(fmt.Sprintf("No type defined for metric[%d]", i))
		} else if _, ok := MetricTypes[m.Type]; !ok {
			return errors.New(fmt.Sprintf("Unknown metric type \"%s\" for metric[%d]", m.Type, i))
		}

		metricNames[m.Name] = int32(i)
	}

	return nil
}

func PrepareTag(tag string) string {
	if len(tag) > TIRION_TAG_SIZE {
		tag = tag[:TIRION_TAG_SIZE]
	}

	return strings.Replace(tag, "\n", " ", -1)
}

func (t *Tirion) initShm(filename string, create bool, count int32) error {
	var err error

	t.V("Open shared memory")
	t.shm, err = shm.NewShm(filename, create, count)

	if err != nil {
		return err
	}

	t.V("Read shared memory")
	err = t.shm.Read()

	if err != nil {
		return err
	}

	return nil
}

func (t *Tirion) initSigHandler() {
	t.V("Create signal handler")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		s := <-sig
		t.V("Catched signal %v", s)

		t.Running = false
	}()
}

func (t *Tirion) receive() (string, error) {
	var buf = make([]byte, 4096)

	nr, err := t.fd.Read(buf)

	if err != nil {
		return "", err
	}

	return strings.Trim(string(buf[0:nr]), "\n"), nil
}

func (t *Tirion) send(msg string) error {
	_, err := t.fd.Write([]byte(msg + "\n"))

	return err
}

func (t *Tirion) D(format string, a ...interface{}) (n int, err error) {
	if !t.verbose {
		return
	}

	return fmt.Fprintf(os.Stderr, t.logPrefix+"[debug] "+format+"\n", a...)
}

func (t *Tirion) E(format string, a ...interface{}) (n int, err error) {
	if !t.verbose {
		return
	}

	return fmt.Fprintf(os.Stderr, t.logPrefix+"[error] "+format+"\n", a...)
}

func (t *Tirion) V(format string, a ...interface{}) (n int, err error) {
	if !t.verbose {
		return
	}

	return fmt.Fprintf(os.Stderr, t.logPrefix+"[verbose] "+format+"\n", a...)
}
