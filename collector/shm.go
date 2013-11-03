package collector

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/zimmski/tirion/shm"
)

type CollectorShm struct {
	shm *shm.Shm
}

func (c *CollectorShm) InitAgent(pid int32, metricCount int32) (*url.URL, error) {
	var u = &url.URL{
		Scheme: "shm",
		Path:   fmt.Sprintf("/proc/%d", pid),
	}

	err := c.initShm(u.Path, true, metricCount)

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (c *CollectorShm) InitClient(u *url.URL, metricCount int32) error {
	if _, err := os.Stat(u.Path); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Cannot open shm path: %v", err))
	}

	return c.initShm(u.Path, false, metricCount)
}

func (c *CollectorShm) initShm(filename string, create bool, count int32) error {
	var err error

	c.shm, err = shm.NewShm(filename, create, count)

	if err != nil {
		return errors.New(fmt.Sprintf("Open shared memory: %v", err))
	}

	err = c.shm.Read()

	if err != nil {
		return errors.New(fmt.Sprintf("Read shared memory: %v", err))
	}

	return nil
}

func (c *CollectorShm) Data() []float32 {
	return c.shm.Data()
}

func (c *CollectorShm) Close() error {
	return c.shm.Close()
}

func (c *CollectorShm) Add(i int32, v float32) float32 {
	return c.shm.Add(i, v)
}

func (c *CollectorShm) Dec(i int32) float32 {
	return c.shm.Dec(i)
}

func (c *CollectorShm) Inc(i int32) float32 {
	return c.shm.Inc(i)
}

func (c *CollectorShm) Sub(i int32, v float32) float32 {
	return c.shm.Sub(i, v)
}
