package tirion

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/zimmski/tirion/shm"
)

const Version = "0.1"

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
	Id            int
	Name          string
	SubName       string
	Interval      int
	Metrics       []Metric
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

func (t *Tirion) initShm(filename string, create bool, count int) error {
	var err error

	t.V("Open shared memory")
	t.shm, err = shm.NewShm(filename, create, count)

	if err != nil {
		return err
	}

	t.V("Read shared memory")
	err = t.shm.Read()

	if err != nil {
		panic(err)
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

func (t *Tirion) send(msg string) {
	_, err := t.fd.Write([]byte(msg + "\n"))

	if err != nil {
		panic(err)
	}
}

// TODO remove this or maybe we should put them in a logging package or use another logging package

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
