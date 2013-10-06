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
	id     int32
	create bool
	addr   *C.float
	count  int32
}

func NewShm(filename string, create bool, count int32) (*Shm, error) {
	f := C.CString(filename)
	defer C.free(unsafe.Pointer(f))

	var c C.char

	if create {
		c = C.char(1)
	} else {
		c = C.char(0)
	}

	id := int32(C.shmOpen(f, c, C.long(count)))

	if id == -1 {
		return nil, errors.New("Shm open error")
	}

	return &Shm{id, create, nil, count}, nil
}

func (shm *Shm) Close() error {
	C.shmDetach(shm.addr)

	if shm.create {
		if C.shmClose(C.long(shm.id)) != 0 {
			return errors.New("Shm close error")
		}
	}

	return nil
}

func (shm *Shm) Data() []float32 {
	a := make([]float32, shm.count)

	C.shmCopy(shm.addr, (*C.float)(unsafe.Pointer(&a[0])), C.long(shm.count))

	return a
}

func (shm *Shm) Get(i int32) float32 {
	if i < 0 || i >= shm.count {
		return 0.0
	}

	var v C.float = C.shmGet(shm.addr, C.long(i))

	return float32(v)
}

func (shm *Shm) Read() error {
	if shm.id == 0 {
		return errors.New("No shm id defined")
	}

	// TODO map the shared memory directly to a go structure so we can use it via index directly. well, how?!
	shm.addr = C.shmAttach(C.long(shm.id))

	return nil
}

func (shm *Shm) Set(i int32, v float32) float32 {
	if i < 0 || i >= shm.count {
		return 0.0
	}

	return float32(C.shmSet(shm.addr, C.long(i), C.float(v)))
}

func (shm *Shm) Add(i int32, v float32) float32 {
	if i < 0 || i >= shm.count {
		return 0.0
	}

	return float32(C.shmAdd(shm.addr, C.long(i), C.float(v)))
}

func (shm *Shm) Dec(i int32) float32 {
	if i < 0 || i >= shm.count {
		return 0.0
	}

	return float32(C.shmDec(shm.addr, C.long(i)))
}

func (shm *Shm) Inc(i int32) float32 {
	if i < 0 || i >= shm.count {
		return 0.0
	}

	return float32(C.shmInc(shm.addr, C.long(i)))
}

func (shm *Shm) Sub(i int32, v float32) float32 {
	if i < 0 || i >= shm.count {
		return 0.0
	}

	return float32(C.shmSub(shm.addr, C.long(i), C.float(v)))
}
