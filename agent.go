package tirion

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zimmski/tirion/proc"
)

type ExecProgram struct {
	pid           int32
	exec          string
	execArguments []string
	limitTime     int32
}

type TirionAgent struct {
	Tirion
	chMessages           chan interface{}
	cmd                  *exec.Cmd
	interval             int32
	l                    net.Listener
	program              ExecProgram
	metrics              []Metric
	metricsExternal      []int32
	metricsExternalIO    map[int32]int32
	metricsExternalStat  map[int32]int32
	metricsExternalStatm map[int32]int32
	metricsInternal      []int32
	metricsFilename      string
	name                 string
	run                  int32
	sendInterval         int32
	server               string
	serverConn           net.Conn
	serverClient         *httputil.ClientConn
	subName              string
	writerCSV            *csv.Writer
}

func NewTirionAgent(name string, subName string, server string, sendInterval int32, pid int32, metricsFilename string, exec string, execArguments []string, interval int32, socket string, verbose bool, limitTime int32) *TirionAgent {
	return &TirionAgent{
		Tirion: Tirion{
			socket:    socket,
			verbose:   verbose,
			logPrefix: "[agent]",
		},
		name:            name,
		subName:         subName,
		server:          server,
		sendInterval:    sendInterval,
		interval:        interval,
		metricsFilename: metricsFilename,
		program: ExecProgram{
			pid:           pid,
			exec:          exec,
			execArguments: execArguments,
			limitTime:     limitTime,
		},
	}
}

// TODO this is kind of ugly. find a better way to clean up
func (a *TirionAgent) Close() {
	a.closeProgram()
	a.closeSocket()
}

func (a *TirionAgent) closeServerConn() {
	if a.serverClient != nil {
		a.serverClient.Close()

		a.serverClient = nil
		a.serverConn = nil
	}
}

func (a *TirionAgent) closeProgram() {
	if a.cmd != nil {
		if a.cmd.ProcessState == nil {
			closeCmd := time.AfterFunc(2*time.Second, func() {
				a.E("Timeout waiting for program to exit. Let's kill it.")

				a.cmd.Process.Kill()
			})

			a.V("Wait for program to close")

			a.cmd.Wait()

			closeCmd.Stop()
		} else {
			a.V("Program already terminated")
		}

		a.cmd = nil
	}
}

func (a *TirionAgent) closeSocket() {
	if a.fd != nil {
		a.fd.Close()
	}
	if a.l != nil {
		a.l.Close()
	}
}

func (a *TirionAgent) Init() {
	a.initSigHandler()

	a.V("Read metrics file %s", a.metricsFilename)

	jsonFile, err := ioutil.ReadFile(a.metricsFilename)

	if err != nil {
		a.sPanic(fmt.Sprintf("Read metrics file: %v", err))
	}

	err = json.Unmarshal(jsonFile, &a.metrics)

	if err != nil {
		a.sPanic(fmt.Sprintf("Parse metrics file: %v", err))
	}

	if err := CheckMetrics(a.metrics); err != nil {
		a.sPanic(err.Error())
	}

	a.metricsExternalIO = make(map[int32]int32)
	a.metricsExternalStat = make(map[int32]int32)
	a.metricsExternalStatm = make(map[int32]int32)

	for i, m := range a.metrics {
		if strings.HasPrefix(m.Name, "proc") {
			a.V("External metric %+v", m)

			a.metricsExternal = append(a.metricsExternal, int32(i))

			if k, ok := proc.ProcIOIndizes[m.Name]; ok {
				a.metricsExternalIO[int32(k)] = int32(i)
			} else if k, ok := proc.ProcStatIndizes[m.Name]; ok {
				a.metricsExternalStat[int32(k)] = int32(i)
			} else if k, ok := proc.ProcStatmIndizes[m.Name]; ok {
				a.metricsExternalStatm[int32(k)] = int32(i)
			} else {
				a.sPanic(fmt.Sprintf("Unknown metric \"%s\"", m.Name))
			}
		} else {
			a.V("Internal metric %+v", m)

			a.metricsInternal = append(a.metricsInternal, int32(i))
		}
	}

	a.chMessages = make(chan interface{}, 100)

	if a.server != "" {
		a.V("Open server connection to %s", a.server)

		// TODO there is a bug in go 1.1.2. if net.Dial is fed an unexisting address, it will uncatchable panic. so we have to work around that
		a.serverConn, err = net.Dial("tcp", a.server)

		if err != nil {
			a.sPanic(fmt.Sprintf("Cannot connect to server: %v", err))
		}

		a.serverClient = httputil.NewClientConn(a.serverConn, nil)

		a.V("Request new run ID")

		m, _ := json.Marshal(a.metrics)

		runRequestData := url.Values{
			"name":           []string{a.name},
			"sub_name":       []string{a.subName},
			"interval":       []string{strconv.FormatInt(int64(a.interval), 10)},
			"metrics":        []string{string(m)},
			"prog":           []string{a.program.exec},
			"prog_arguments": a.program.execArguments,
		}
		runRequest, err := http.NewRequest("POST", "/program/"+a.name+"/run/start", ioutil.NopCloser(strings.NewReader(runRequestData.Encode())))
		var runRequestResult MessageReturnStart

		runRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		if err != nil {
			a.sPanic(fmt.Sprintf("Cannot create run request %v", err))
		}

		resp, err := a.serverClient.Do(runRequest)

		if err != nil {
			a.sPanic(fmt.Sprintf("Cannot do run request %v", err))
		} else if resp.StatusCode != 200 {
			a.sPanic(fmt.Sprintf("Run request failed with status %v", resp.StatusCode))
		}

		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		json.Unmarshal(body, &runRequestResult)

		if runRequestResult.Error != "" {
			a.sPanic(fmt.Sprintf("Run request failed with returned error: %v", runRequestResult.Error))
		}

		a.V("Received run ID %d", runRequestResult.Run)
		a.run = runRequestResult.Run
	} else {
		var tagNames = make([]string, len(a.metrics))

		for i, m := range a.metrics {
			tagNames[i] = m.Name
		}

		a.writerCSV = csv.NewWriter(os.Stdout)
		a.writerCSV.Comma = ';'
		a.writerCSV.Write(append([]string{"time", "tag"}, tagNames...))
	}

	if a.socket != "" {
		os.Remove(a.socket)

		var err error

		a.V("Open unix socket to %s", a.socket)
		a.l, err = net.Listen("unix", a.socket)

		if err != nil {
			a.sPanic(fmt.Sprintf("Listen to unix socket: %v", err))
		}
	}

	if a.program.exec != "" {
		a.V("Execute external program: %s %s", a.program.exec, strings.Join(a.program.execArguments, " "))
		a.cmd = exec.Command(a.program.exec, a.program.execArguments...)

		a.cmd.Stderr = os.Stderr
		a.cmd.Stdout = os.Stdout
	} else if _, err := os.Stat(fmt.Sprintf("/proc/%d/", a.program.pid)); os.IsNotExist(err) {
		a.sPanic(fmt.Sprintf("PID %d does not exists", a.program.pid))
	}
}

func (a *TirionAgent) handleCommands(c chan bool) {
	a.V("Start listening to commands")

	for a.Running {
		data, err := a.receive()

		switch err {
		case nil:
			com := data[0]

			switch com {
			case 't':
				a.chMessages <- MessageTag{Message{time.Now()}, PrepareTag(data)[1:]}
			default:
				a.E("Unknown command '%c'", com)
			}
		case io.EOF:
			a.V("Unix socket got closed with EOF")

			a.Running = false
		default:
			if strings.HasSuffix(err.Error(), "use of closed network connection") {
				a.V("Unix socket suddenly got closed")

				a.Running = false
			} else {
				a.sPanic(err)
			}
		}

	}

	a.V("Stop listening to commands")

	c <- true
}

func (a *TirionAgent) handleMessages(c chan bool) {
	a.V("Start handling messages")

	var arr = make([]string, len(a.metrics))
	var emptyMetric = make([]string, len(a.metrics))

	var metrics []MessageData
	var metricsQueue chan MessageData

	var tagRequest *http.Request
	var tagRequestData url.Values
	var tagRequestResult MessageReturnTag

	var metricsRequest *http.Request
	var metricsRequestData url.Values
	var metricsRequestResult MessageReturnInsert

	var sendMetrics func()
	var sendMetricsTicker *time.Ticker

	if a.writerCSV == nil {
		var err error

		tagRequest, err = http.NewRequest("POST", fmt.Sprintf("/program/%s/run/%d/tag", a.name, a.run), nil)
		tagRequestData = url.Values{"tag": nil, "time": nil}

		if err != nil {
			a.sPanic(fmt.Sprintf("Cannot create tag request %v", err))
		}

		tagRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		metrics = make([]MessageData, 0, 100)
		metricsQueue = make(chan MessageData, 1000)

		metricsRequest, err = http.NewRequest("POST", fmt.Sprintf("/program/%s/run/%d/insert", a.name, a.run), nil)
		metricsRequestData = url.Values{"metrics": nil}

		if err != nil {
			a.sPanic(fmt.Sprintf("Cannot create metrics request %v", err))
		}

		metricsRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		sendMetrics = func() {
			count := len(metricsQueue)

			if count == 0 {
				return
			}

			metrics = metrics[:0]

			for i := 0; i < count; i++ {
				metrics = append(metrics, <-metricsQueue)
			}

			a.D("Send metrics to server: %v", metrics)

			j, _ := json.Marshal(metrics)
			metricsRequestData.Set("metrics", string(j))

			metricsRequest.Body = ioutil.NopCloser(strings.NewReader(metricsRequestData.Encode()))

			resp, err := a.serverClient.Do(metricsRequest)

			if err != nil {
				a.sPanic(fmt.Sprintf("Cannot do metrics request %v", err))
			} else if resp.StatusCode != 200 {
				a.sPanic(fmt.Sprintf("Insert request failed with status %v", resp.StatusCode))
			}

			body, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()

			json.Unmarshal(body, &metricsRequestResult)

			if metricsRequestResult.Error != "" {
				a.sPanic(fmt.Sprintf("Insert request failed with error %v", metricsRequestResult.Error))
			}
		}
		sendMetricsTicker = time.NewTicker(time.Duration(a.sendInterval) * time.Second)
		go func() {
			for _ = range sendMetricsTicker.C {
				sendMetrics()
			}
		}()
	}

	for m := range a.chMessages {
		switch m.(type) {
		case MessageData:
			data, _ := m.(MessageData)

			if a.writerCSV != nil {
				for i, v := range data.Data {
					arr[i] = strconv.FormatFloat(float64(v), 'f', 3, 32)
				}

				a.writerCSV.Write(append([]string{strconv.FormatInt(data.Time.UnixNano(), 10), ""}, arr...))
				a.writerCSV.Flush()
			} else {
				metricsQueue <- data
			}
		case MessageTag:
			tag, _ := m.(MessageTag)

			if a.writerCSV != nil {
				a.writerCSV.Write(append([]string{strconv.FormatInt(tag.Time.UnixNano(), 10), tag.Tag}, emptyMetric...))
				a.writerCSV.Flush()
			} else {
				a.D("Send tag to server %+v", tag)

				tagRequestData.Set("tag", tag.Tag)
				tagRequestData.Set("time", strconv.FormatInt(tag.Time.UnixNano(), 10))

				tagRequest.Body = ioutil.NopCloser(strings.NewReader(tagRequestData.Encode()))

				resp, err := a.serverClient.Do(tagRequest)

				if err != nil {
					a.sPanic(fmt.Sprintf("Cannot do tag request %v", err))
				} else if resp.StatusCode != 200 {
					a.sPanic(fmt.Sprintf("Tag request failed with status %v", resp.StatusCode))
				}

				body, _ := ioutil.ReadAll(resp.Body)
				resp.Body.Close()

				json.Unmarshal(body, &tagRequestResult)

				if tagRequestResult.Error != "" {
					a.sPanic(fmt.Sprintf("Tag request failed with error %v", tagRequestResult.Error))
				}
			}
		}
	}

	if a.writerCSV == nil {
		sendMetricsTicker.Stop()
		sendMetrics()
	}

	a.V("Stop handling messages")

	c <- true
}

func (a *TirionAgent) handleMetrics(c chan bool) {
	pidFolder := fmt.Sprintf("/proc/%d/", a.program.pid)

	a.V("Start fetching metrics")

	for a.Running {
		if _, err := os.Stat(pidFolder); os.IsNotExist(err) || (a.cmd != nil && a.cmd.ProcessState != nil) {
			a.V("PID disappeared")

			a.Running = false

			break
		}

		// NOTE: we have to create this metrics slice everytime because otherwise it would be just a pointer :-)
		var metrics = make([]float32, len(a.metrics))
		var now = time.Now()

		if len(a.metricsExternalIO) > 0 {
			pIO, err := proc.ReadIOArray(pidFolder + "io")

			if err != nil {
				a.E("read io: " + err.Error())

				break
			}

			for k, v := range a.metricsExternalIO {
				f, _ := strconv.ParseFloat(pIO[k], 32)
				metrics[v] = float32(f)
			}
		}

		if len(a.metricsExternalStat) > 0 {
			pStat, err := proc.ReadStatArray(pidFolder + "stat")

			if err != nil {
				a.E("read stat: " + err.Error())

				break
			}

			for k, v := range a.metricsExternalStat {
				f, _ := strconv.ParseFloat(pStat[k], 32)
				metrics[v] = float32(f)
			}
		}

		if len(a.metricsExternalStatm) > 0 {
			pStatm, err := proc.ReadStatmArray(pidFolder + "statm")

			if err != nil {
				a.E("read statm: " + err.Error())

				break
			}

			for k, v := range a.metricsExternalStatm {
				f, _ := strconv.ParseFloat(pStatm[k], 32)
				metrics[v] = float32(f)
			}
		}

		if a.shm != nil {
			for i, v := range a.shm.Data() {
				metrics[a.metricsInternal[i]] = v
			}
		}

		a.chMessages <- MessageData{Message{now}, metrics}

		time.Sleep(time.Duration(a.interval) * time.Millisecond)
	}

	a.V("Stop fetching metrics")

	c <- true
}

func (a *TirionAgent) Run() {
	var err error

	if a.cmd != nil {
		err := a.cmd.Start()

		if err != nil {
			a.sPanic(err)
		}

		// if the program exits on its own we immediately want to know about it
		go a.cmd.Wait()

		defer a.closeProgram()

		a.program.pid = int32(a.cmd.Process.Pid)

		if a.program.limitTime > 0 {
			time.AfterFunc(time.Duration(a.program.limitTime)*time.Second, func() {
				a.V("Limit reached. Program ran for %d seconds. Let's kill it.", a.program.limitTime)

				a.cmd.Process.Kill()
			})
		}
	}

	a.V("Monitor program with PID %d", a.program.pid)

	if a.l != nil {
		closeUnix := time.AfterFunc(1*time.Second, func() {
			a.E("Timeout reading unix socket")

			a.l.Close()
		})

		a.fd, err = a.l.Accept()

		closeUnix.Stop()

		defer a.closeSocket()

		if err != nil {
			if strings.HasSuffix(err.Error(), "use of closed network connection") {
				a.sPanic("Unix socket got closed already")
			} else {
				a.sPanic(fmt.Sprintf("Accept %v", err))
			}
		}

		clientVersion, err := a.receive()

		if err != nil {
			a.sPanic(err)
		}

		var matchClientVersion = regexp.MustCompile("^tirion v([0-9.]+)$").FindStringSubmatch(clientVersion)

		if matchClientVersion == nil {
			a.sPanic("Client did not send tirion protocol version")
		}

		a.V("Requested tirion protocol version v%s", matchClientVersion[1])
		a.V("Using tirion protocol version v" + Version)

		var shmPath = fmt.Sprintf("/proc/%d", a.program.pid)

		err = a.initShm(shmPath, true, int32(len(a.metricsInternal)))

		if err != nil {
			// TODO better error handling, most of the time the shared memory does already exists
			a.sPanic(fmt.Sprintf("Cannot open shared memory: %v", err))
		}

		defer a.shm.Close()

		a.V("Send metric count %d and shm path %s", len(a.metricsInternal), shmPath)
		if err := a.send(fmt.Sprintf("%d\t%s", len(a.metricsInternal), shmPath)); err != nil {
			a.sPanic(fmt.Sprintf("Send error: %v", err))
		}
	}

	a.Running = true

	var chHandleCommands chan bool
	var chHandleMessages = make(chan bool)
	var chHandleMetrics = make(chan bool)

	go a.handleMessages(chHandleMessages)
	if a.l != nil {
		chHandleCommands = make(chan bool)

		go a.handleCommands(chHandleCommands)
	}
	go a.handleMetrics(chHandleMetrics)

	<-chHandleMetrics
	if a.l != nil {
		<-chHandleCommands
	}

	close(a.chMessages)

	<-chHandleMessages

	a.V("Request stop of run")

	stopRequest, err := http.NewRequest("GET", fmt.Sprintf("/program/%s/run/%d/stop", a.name, a.run), nil)
	var stopRequestResult MessageReturnStop

	if err != nil {
		a.sPanic(fmt.Sprintf("Cannot create stop request %v", err))
	}

	resp, err := a.serverClient.Do(stopRequest)

	if err != nil {
		a.sPanic(fmt.Sprintf("Cannot do stop request %v", err))
	} else if resp.StatusCode != 200 {
		a.sPanic(fmt.Sprintf("Stop request failed with status %v", resp.StatusCode))
	}

	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	json.Unmarshal(body, &stopRequestResult)

	if stopRequestResult.Error != "" {
		a.sPanic(fmt.Sprintf("Stop request failed with error %v", stopRequestResult.Error))
	}

	a.V("Stopped run")
}

// TODO remove this somehow
func (a *TirionAgent) sPanic(err interface{}) {
	if a.shm != nil {
		// this is needed before a a.sPanic because otherwise the shm stays open
		// there must be a better solution to this...
		a.shm.Close()
	}

	panic(err)
}
