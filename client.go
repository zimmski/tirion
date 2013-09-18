package tirion

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type TirionClient struct {
	Tirion
}

func NewTirionClient(socket string, verbose bool) *TirionClient {
	return &TirionClient{
		Tirion{
			socket:    socket,
			verbose:   verbose,
			logPrefix: "[client]",
		},
	}
}

func (c *TirionClient) Init() error {
	var err error

	c.initSigHandler()

	c.V("Open unix socket to %s", c.socket)
	c.fd, err = net.Dial("unix", c.socket)

	if err != nil {
		if strings.HasSuffix(err.Error(), "use of closed network connection") || strings.HasSuffix(err.Error(), "no such file or directory") {
			c.E("Cannot open unix socket %s", c.socket)
		}

		return err
	} else {
		c.V("Request tirion protocol version v%s", Version)
		c.send("tirion v" + Version)

		m, err := c.receive()

		switch err {
		case nil:
			var metricCount, err = strconv.Atoi(m)

			if err != nil {
				c.E("Did not receive metric count")

				return err
			}

			c.V("Received metric count %d", metricCount)

			err = c.initShm("/tmp", false, metricCount)

			if err != nil {
				c.E("Cannot initialize shared memory")

				return err
			}

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

// TODO add Close function like in the C client
// TODO add Destroy function like in the C client

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
				c.V("Unix socket suddenly got closed")

				c.Running = false
			} else {
				panic(err)
			}
		}
	}

	c.V("Stop listening to commands")
}

func (c *TirionClient) Add(i int, v float32) float32 {
	return c.shm.Add(i, v)
}

func (c *TirionClient) Dec(i int) float32 {
	return c.shm.Dec(i)
}

func (c *TirionClient) Inc(i int) float32 {
	return c.shm.Inc(i)
}

func (c *TirionClient) Sub(i int, v float32) float32 {
	return c.shm.Sub(i, v)
}

func (c *TirionClient) Tag(format string, a ...interface{}) {
	c.send(fmt.Sprintf("t"+format, a...))
}
