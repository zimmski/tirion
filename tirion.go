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

// Version of Tirion.
// The version is also used to dictated the used
// protocol between agent and client communication.
const Version = "0.1"

const tirion_tag_size = 513

// HighStockTag contains all data of a tag used with the HighStock library.
type HighStockTag struct {
	X     int64  `json:"x"`
	Title string `json:"title"`
}

// Metric contains all data of a metric.
type Metric struct {
	Name string
	Type string
}

// metricTypes holds all useable metric types.
var metricTypes = map[string]bool{
	"int":   true,
	"float": true,
}

// Program contains all data of a program.
type Program struct {
	Name string
}

// Run contains all data of a run.
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

// Tag contains all data of a tag.
type Tag struct {
	Time time.Time
	Tag  string
}

// Tirion contains all common data of a Tirion object like TirionAgent and TirionClient.
type Tirion struct {
	fd        net.Conn
	Running   bool // states if the given Tirion object is still running
	shm       *shm.Shm
	socket    string
	verbose   bool
	logPrefix string
}

// CheckMetrics validates a array of metrics.
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
		} else if _, ok := metricTypes[m.Type]; !ok {
			return errors.New(fmt.Sprintf("Unknown metric type \"%s\" for metric[%d]", m.Type, i))
		}

		metricNames[m.Name] = int32(i)
	}

	return nil
}

// PrepareTag modifies a raw tag to a valid state.
func PrepareTag(tag string) string {
	if len(tag) > tirion_tag_size {
		tag = tag[:tirion_tag_size]
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

// D outputs a Tirion debug message.
func (t *Tirion) D(format string, a ...interface{}) (n int, err error) {
	if !t.verbose {
		return
	}

	return fmt.Fprintf(os.Stderr, t.logPrefix+"[debug] "+format+"\n", a...)
}

// E outputs a Tirion error message.
func (t *Tirion) E(format string, a ...interface{}) (n int, err error) {
	if !t.verbose {
		return
	}

	return fmt.Fprintf(os.Stderr, t.logPrefix+"[error] "+format+"\n", a...)
}

// V outputs a Tirion verbose message.
func (t *Tirion) V(format string, a ...interface{}) (n int, err error) {
	if !t.verbose {
		return
	}

	return fmt.Fprintf(os.Stderr, t.logPrefix+"[verbose] "+format+"\n", a...)
}
