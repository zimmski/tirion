package collector

/*
#include <stdlib.h>
#include "shm_linux.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"unsafe"
)

type CollectorShm struct {
	id     int32
	create bool
	addr   *C.float
	count  int32
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
	c.create = create
	c.count = count

	f := C.CString(filename)
	defer C.free(unsafe.Pointer(f))

	var cr C.char

	if c.create {
		cr = C.char(1)
	} else {
		cr = C.char(0)
	}

	c.id = int32(C.shmOpen(f, cr, C.long(c.count)))

	if c.id == -1 {
		return errors.New("Shm open error")
	}

	/**
	 * TODO Map the address directly to a Go structure
	 * this would make it possible to use indizes to access
	 * the array elements.
	 */
	c.addr = C.shmAttach(C.long(c.id))

	return nil
}

func (c *CollectorShm) Data() []float32 {
	a := make([]float32, c.count)

	C.shmCopy(c.addr, (*C.float)(unsafe.Pointer(&a[0])), C.long(c.count))

	return a
}

func (c *CollectorShm) Close() error {
	C.shmDetach(c.addr)

	if c.create {
		if C.shmClose(C.long(c.id)) != 0 {
			return errors.New("Shm close error")
		}
	}

	return nil
}

func (c *CollectorShm) Get(i int32) float32 {
	if i < 0 || i >= c.count {
		return 0.0
	}

	var v C.float = C.shmGet(c.addr, C.long(i))

	return float32(v)
}

func (c *CollectorShm) Set(i int32, v float32) float32 {
	if i < 0 || i >= c.count {
		return 0.0
	}

	return float32(C.shmSet(c.addr, C.long(i), C.float(v)))
}

func (c *CollectorShm) Add(i int32, v float32) float32 {
	if i < 0 || i >= c.count {
		return 0.0
	}

	return float32(C.shmAdd(c.addr, C.long(i), C.float(v)))
}

func (c *CollectorShm) Dec(i int32) float32 {
	return c.Add(i, -1.0)
}

func (c *CollectorShm) Inc(i int32) float32 {
	return c.Add(i, 1.0)
}

func (c *CollectorShm) Sub(i int32, v float32) float32 {
	return c.Add(i, -v)
}
