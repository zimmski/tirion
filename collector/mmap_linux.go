package collector

/*
#include <stdlib.h>
#include "mmap_linux.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"unsafe"
)

type CollectorMmap struct {
	addr     *C.float
	count    int32
	create   bool
	filename string
}

func (c *CollectorMmap) InitAgent(pid int32, metricCount int32) (*url.URL, error) {
	var u = &url.URL{
		Scheme: "mmap",
		Path:   fmt.Sprintf("%s/tirion-%d.mmap", os.TempDir(), pid),
	}

	err := c.initMmap(u.Path, true, metricCount)

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (c *CollectorMmap) InitClient(u *url.URL, metricCount int32) error {
	if _, err := os.Stat(u.Path); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Cannot open mmap file: %v", err))
	}

	return c.initMmap(u.Path, false, metricCount)
}

func (c *CollectorMmap) initMmap(filename string, create bool, count int32) error {
	c.count = count
	c.create = create
	c.filename = filename

	f := C.CString(filename)
	defer C.free(unsafe.Pointer(f))

	var cr C.char

	if create {
		cr = C.char(1)
	} else {
		cr = C.char(0)
	}

	/**
	 * TODO Map the address directly to a Go structure
	 * this would make it possible to use indizes to access
	 * the array elements.
	 */
	c.addr = C.mmapOpen(f, cr, C.long(count))

	if c.addr == nil {
		return errors.New("Cannot open mmap")
	}

	return nil
}

func (c *CollectorMmap) Data() []float32 {
	a := make([]float32, c.count)

	C.mmapCopy(c.addr, (*C.float)(unsafe.Pointer(&a[0])), C.long(c.count))

	return a
}

func (c *CollectorMmap) Close() error {
	f := C.CString(c.filename)
	defer C.free(unsafe.Pointer(f))

	var cr C.char

	if c.create {
		cr = C.char(1)
	} else {
		cr = C.char(0)
	}

	if C.mmapClose(c.addr, f, cr, C.long(c.count)) != 0 {
		return errors.New("Mmap close error")
	}

	return nil
}

func (c *CollectorMmap) Add(i int32, v float32) float32 {
	if i < 0 || i >= c.count {
		return 0.0
	}

	return float32(C.mmapAdd(c.addr, C.long(i), C.float(v)))
}

func (c *CollectorMmap) Dec(i int32) float32 {
	if i < 0 || i >= c.count {
		return 0.0
	}

	return float32(C.mmapDec(c.addr, C.long(i)))
}

func (c *CollectorMmap) Inc(i int32) float32 {
	if i < 0 || i >= c.count {
		return 0.0
	}

	return float32(C.mmapInc(c.addr, C.long(i)))
}

func (c *CollectorMmap) Sub(i int32, v float32) float32 {
	if i < 0 || i >= c.count {
		return 0.0
	}

	return float32(C.mmapSub(c.addr, C.long(i), C.float(v)))
}
