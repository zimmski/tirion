package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/zimmski/tirion"
)

func main() {
	var flagExec string
	var flagExecArguments string
	var flagHelp bool
	var flagInterval int
	var flagLimitMemory int
	var flagLimitMemoryInterval int
	var flagLimitTime int
	var flagMetrics string
	var flagMetricsFile string
	var flagName string
	var flagPid int
	var flagSendInterval int
	var flagServer string
	var flagSocket string
	var flagSubName string
	var flagVerbose bool

	flag.BoolVar(&flagHelp, "help", false, "Show this help")
	flag.StringVar(&flagExec, "exec", "", "Execute this command")
	flag.StringVar(&flagExecArguments, "exec-arguments", "", "Arguments for the command")
	flag.IntVar(&flagInterval, "interval", 250, "How often metrics are fetched (in milliseconds)")
	flag.IntVar(&flagLimitMemory, "limit-memory", 0, "Limit the memory of the program and its children (in MB)")
	flag.IntVar(&flagLimitMemoryInterval, "limit-memory-interval", 5, "Interval for checking the memory limit (in milliseconds)")
	flag.IntVar(&flagLimitTime, "limit-time", 0, "Limit the runtime of the program (in seconds)")
	flag.StringVar(&flagMetrics, "metrics", "", "Definition of needed program metrics")
	flag.StringVar(&flagMetricsFile, "metrics-file", "", "Definition of needed program metrics as a JSON file")
	flag.StringVar(&flagName, "name", "", "The name of this run (defaults to exec)")
	flag.IntVar(&flagPid, "pid", -1, "PID of program which should be monitored")
	flag.IntVar(&flagSendInterval, "send-interval", 5, "How often data is pushed to the server (in seconds)")
	flag.StringVar(&flagServer, "server", "", "Server address for agent<-->server communication")
	flag.StringVar(&flagSocket, "socket", "", "Unix socket path for client<-->agent communication")
	flag.StringVar(&flagSubName, "sub-name", "", "The subname of this run")
	flag.BoolVar(&flagVerbose, "verbose", false, "Verbose output of what is going on")

	flag.Parse()

	if (flagPid == -1 && flagExec == "") || (flagMetrics == "" && flagMetricsFile == "") || flagHelp {
		fmt.Printf("Tirion agent v%s\n", tirion.Version)
		fmt.Printf("usage:\n")
		fmt.Printf("\t%s -pid <pid> -metrics <metrics> [other options]\n", os.Args[0])
		fmt.Printf("\t%s -pid <pid> -metrics-file <metrics json file> [other options]\n", os.Args[0])
		fmt.Printf("\t%s -exec <program> -metrics <metrics> [other options]\n", os.Args[0])
		fmt.Printf("\t%s -exec <program> -metrics-file <metrics json file> [other options]\n", os.Args[0])
		fmt.Printf("options\n")
		flag.PrintDefaults()
		fmt.Printf("\n")

		if !flagHelp {
			fmt.Printf("ERROR: Wrong arguments\n")
		}

		os.Exit(1)
	}

	if flagName == "" && flagExec != "" {
		flagName = flagExec
	}
	if flagInterval <= 0 {
		panic("ERROR: Argument -interval must be a positive number")
	}
	if flagSendInterval <= 0 {
		panic("ERROR: Argument -send-interval must be a positive number")
	}
	if flagLimitTime < 0 {
		panic("ERROR: Argument -limit-time must be a positive number")
	} else if flagLimitTime > 0 && flagExec == "" {
		panic("ERROR: -limit-time only works in combination with -exec")
	}
	if flagLimitMemory < 0 {
		panic("ERROR: Argument -limit-memory must be a positive number")
	} else if flagLimitMemory > 0 && flagExec == "" {
		panic("ERROR: -limit-memory only works in combination with -exec")
	}
	if flagLimitMemoryInterval <= 0 {
		panic("ERROR: Argument -limit-memory-interval must be a positive number")
	}

	var execArguments []string

	if flagExecArguments != "" {
		execArguments = strings.Split(flagExecArguments, " ")
	}

	var metrics []tirion.Metric

	if flagMetrics != "" {
		metrics = make([]tirion.Metric, 0)

		for _, m := range strings.Split(flagMetrics, ";") {
			mi := strings.Split(m, ",")

			if len(mi) != 2 {
				panic("wrong format for metrics argument")
			}

			metrics = append(metrics, tirion.Metric{Name: mi[0], Type: mi[1]})
		}
	} else {
		jsonFile, err := ioutil.ReadFile(flagMetricsFile)

		if err != nil {
			panic(fmt.Sprintf("Read metrics file: %v", err))
		}

		err = json.Unmarshal(jsonFile, &metrics)

		if err != nil {
			panic(fmt.Sprintf("Parse metrics file: %v", err))
		}
	}

	a := tirion.NewTirionAgent(
		flagName,
		flagSubName,
		flagServer,
		int32(flagSendInterval),
		int32(flagPid),
		metrics,
		flagExec,
		execArguments,
		int32(flagInterval),
		flagSocket,
		flagVerbose,
		int64(flagLimitMemory),
		int32(flagLimitMemoryInterval),
		int32(flagLimitTime),
	)

	a.Init()
	// defer close in case of errors
	defer a.Close()

	a.Run()

	a.Close()

	// TODO try to remove this. it is only here to complete all defers!
	a.V("Terminate in a second")
	time.Sleep(time.Second)

	a.V("Stopped")

	return
}
