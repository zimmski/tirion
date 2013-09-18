package shm

/*
#include <stdlib.h>
#include "shm_linux.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

type Shm struct {
	id     int
	create bool
	addr   *C.float
	count  int
}

func NewShm(filename string, create bool, count int) (*Shm, error) {
	f := C.CString(filename)
	defer C.free(unsafe.Pointer(f))

	var c C.int

	if create {
		c = C.int(1)
	} else {
		c = C.int(0)
	}

	id := int(C.shmOpen(f, c, C.int(count)))

	if id == -1 {
		return nil, errors.New("Shm open error")
	}

	return &Shm{id, create, nil, count}, nil
}

func (shm *Shm) Close() error {
	C.shmDetach(shm.addr)

	if shm.create {
		if C.shmClose(C.int(shm.id)) != 0 {
			return errors.New("Shm close error")
		}
	}

	return nil
}

func (shm *Shm) Data() []float32 {
	a := make([]float32, shm.count)

	C.shmCopy(shm.addr, (*C.float)(unsafe.Pointer(&a[0])), C.int(shm.count))

	return a
}

func (shm *Shm) Get(i int) float32 {
	var v C.float = C.shmGet(shm.addr, C.int(i))

	return float32(v)
}

func (shm *Shm) Read() error {
	if shm.id == 0 {
		return errors.New("No shm id defined")
	}

	// TODO map the shared memory directly to a go structure so we can use it via index directly. well, how?!
	shm.addr = C.shmAttach(C.int(shm.id))

	return nil
}

func (shm *Shm) Set(i int, v float32) float32 {
	return float32(C.shmSet(shm.addr, C.int(i), C.float(v)))
}

func (shm *Shm) Add(i int, v float32) float32 {
	return float32(C.shmAdd(shm.addr, C.int(i), C.float(v)))
}

func (shm *Shm) Dec(i int) float32 {
	return float32(C.shmDec(shm.addr, C.int(i)))
}

func (shm *Shm) Inc(i int) float32 {
	return float32(C.shmInc(shm.addr, C.int(i)))
}

func (shm *Shm) Sub(i int, v float32) float32 {
	return float32(C.shmAdd(shm.addr, C.int(i), C.float(v)))
}
