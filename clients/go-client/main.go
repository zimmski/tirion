package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/zimmski/tirion"
)

func main() {
	var flagHelp bool
	var flagRuntime int
	var flagSocket string
	var flagVerbose bool

	flag.BoolVar(&flagHelp, "help", false, "Show this help")
	flag.IntVar(&flagRuntime, "runtime", 5, "Runtime of the example client in seconds")
	flag.StringVar(&flagSocket, "socket", "/tmp/tirion.sock", "Unix socket path for client<-->agent communication")
	flag.BoolVar(&flagVerbose, "verbose", false, "Verbose output of what is going on")

	flag.Parse()

	if flagSocket == "" || flagHelp {
		fmt.Printf("Tirion go example client v%s\n", tirion.Version)
		fmt.Printf("usage: %s [options]\n", os.Args[0])
		fmt.Printf("options\n")
		flag.PrintDefaults()

		if !flagHelp {
			fmt.Printf("ERROR: Wrong arguments")
		}

		os.Exit(1)
	}

	c := tirion.NewTirionClient(flagSocket, flagVerbose)

	if err := c.Init(); err != nil {
		panic(err)
	}

	time.AfterFunc(time.Second*time.Duration(flagRuntime), func() {
		c.D("Program ran for %d seconds, this is enough data.", flagRuntime)

		c.Close()
	})

	for c.Running {
		r := c.Inc(0)
		c.Dec(1)
		c.Add(2, 0.3)
		c.Sub(3, 0.3)

		time.Sleep(10 * time.Millisecond)

		if s := math.Mod(float64(r), 50.0); s == 0.0 {
			c.Tag("index 0 is %f", r)
		}
	}

	c.Close()

	c.V("Stopped")

	c.Destroy()

	return
}
