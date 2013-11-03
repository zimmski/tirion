package tirion

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"syscall"

	"github.com/zimmski/tirion/collector"
)

// TirionClient contains the state of a client.
type TirionClient struct {
	Tirion
	metricsCollector         collector.Collector
	PreferredMetricProtocoll string // which metric protocols should be tried first. default is "shm,mmap"
}

// NewTirionClient allocates a new TirionClient object
func NewTirionClient(socket string, verbose bool) *TirionClient {
	return &TirionClient{
		Tirion: Tirion{
			socket:    socket,
			verbose:   verbose,
			logPrefix: "[client]",
		},
		metricsCollector:         nil,
		PreferredMetricProtocoll: "shm,mmap",
	}
}

// Init initializes the client
func (c *TirionClient) Init() error {
	var err error

	if r, err := syscall.Setsid(); r == -1 {
		c.E("Cannot set new session and group id of process: %v")

		return err
	}

	c.V("Open unix socket to %s", c.socket)
	c.fd, err = net.Dial("unix", c.socket)

	if err != nil {
		if strings.HasSuffix(err.Error(), "use of closed network connection") || strings.HasSuffix(err.Error(), "no such file or directory") {
			c.E("Cannot open unix socket %s", c.socket)
		}

		return err
	} else {
		c.V("Request tirion protocol version v%s", Version)
		if err := c.send("tirion v" + Version + "\t" + c.PreferredMetricProtocoll); err != nil {
			c.E(err.Error())

			return err
		}

		m, err := c.receive()

		switch err {
		case nil:
			var t = strings.SplitN(m, "\t", 2)

			if len(t) == 1 || t[1] == "" {
				err := errors.New("Did not receive correct metric count and protocol URL")

				c.E(err.Error())

				return err
			}

			var metricCount, err = strconv.Atoi(t[0])

			if err != nil {
				c.E("Did not receive correct metric count")

				return err
			}

			u, err := url.Parse(t[1])

			if err != nil {
				c.E("Did not receive correct protocol URL")

				return err
			}

			c.V("Received metric count %d and protocol URL %v", metricCount, u)

			c.metricsCollector, err = collector.NewCollector(u.Scheme)

			if err != nil {
				c.E("Cannot create metric collector")

				return err
			}

			err = c.metricsCollector.InitClient(u, int32(metricCount))

			if err != nil {
				c.E("Cannot initialize metrics collector")

				return err
			}

			c.V("Initialized metric collector %s", u.Scheme)

			c.Running = true

			// we want to handle commands not in the main thread
			go c.handleCommands()
		case io.EOF:
			c.V("Unix socket got closed with EOF")

			return err
		default:
			if strings.HasSuffix(err.Error(), "use of closed network connection") {
				c.V("Unix socket suddenly got closed")
			}

			return err
		}
	}

	return nil
}

// Close uninitializes the client by closing all connections of the client.
func (c *TirionClient) Close() error {
	c.Running = false

	if c.metricsCollector != nil {
		if err := c.metricsCollector.Close(); err != nil {
			return err
		}

		c.metricsCollector = nil
	}

	if c.fd != nil {
		if err := c.fd.Close(); err != nil {
			return err
		}

		c.fd = nil
	}

	return nil
}

// Destroy deallocates all data of the client.
func (c *TirionClient) Destroy() error {
	return nil
}

func (c *TirionClient) handleCommands() {
	c.V("Start listening to commands")

	for c.Running {
		var data, err = c.receive()

		switch err {
		case nil:
			com := data[0]

			switch com {
			default:
				c.E("Unknown command '%c'", com)
			}
		case io.EOF:
			c.V("Unix socket got closed with EOF")

			c.Running = false
		default:
			if strings.HasSuffix(err.Error(), "use of closed network connection") {
				if c.Running {
					c.V("Unix socket suddenly got closed")
				}
			} else {
				c.E("%v", err)
			}

			c.Running = false

			break
		}
	}

	c.V("Stop listening to commands")
}

// Add adds a value to a metric
func (c *TirionClient) Add(i int32, v float32) float32 {
	return c.metricsCollector.Add(i, v)
}

// Dec decrements a metric by 1.0
func (c *TirionClient) Dec(i int32) float32 {
	return c.metricsCollector.Dec(i)
}

// Inc increments a metric by 1.0
func (c *TirionClient) Inc(i int32) float32 {
	return c.metricsCollector.Inc(i)
}

// Sub subtracts a value of a metric
func (c *TirionClient) Sub(i int32, v float32) float32 {
	return c.metricsCollector.Sub(i, v)
}

// Tag sends a tag to the agent
func (c *TirionClient) Tag(format string, a ...interface{}) {
	c.send(PrepareTag(fmt.Sprintf("t"+format, a...)))
}
