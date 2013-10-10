package tirion

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
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
		if err := c.send("tirion v" + Version); err != nil {
			c.E(err.Error())

			return err
		}

		m, err := c.receive()

		switch err {
		case nil:
			var t = strings.SplitN(m, "\t", 2)

			if len(t) == 1 || t[1] == "" {
				err := errors.New("Did not receive correct metric count and shm path")

				c.E(err.Error())

				return err
			}

			var metricCount, err = strconv.Atoi(t[0])
			var shmPath = t[1]

			if err != nil {
				c.E("Did not receive correct metric count")

				return err
			} else if _, err := os.Stat(shmPath); os.IsNotExist(err) {
				c.E("Did not receive correct shm path")

				return err
			}

			c.V("Received metric count %d and shm path %s", metricCount, shmPath)

			err = c.initShm(shmPath, false, int32(metricCount))

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

func (c *TirionClient) Close() error {
	c.Running = false

	if c.shm != nil {
		if err := c.shm.Close(); err != nil {
			return err
		}

		c.shm = nil
	}

	if c.fd != nil {
		if err := c.fd.Close(); err != nil {
			return err
		}

		c.fd = nil
	}

	return nil
}

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

func (c *TirionClient) Add(i int32, v float32) float32 {
	return c.shm.Add(i, v)
}

func (c *TirionClient) Dec(i int32) float32 {
	return c.shm.Dec(i)
}

func (c *TirionClient) Inc(i int32) float32 {
	return c.shm.Inc(i)
}

func (c *TirionClient) Sub(i int32, v float32) float32 {
	return c.shm.Sub(i, v)
}

func (c *TirionClient) Tag(format string, a ...interface{}) {
	c.send(PrepareTag(fmt.Sprintf("t"+format, a...)))
}
