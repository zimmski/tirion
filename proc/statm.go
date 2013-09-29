package proc

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type ProcStatm struct {
	Size     int
	Resident int
	Share    int
	Text     int
	Lib      int
	Data     int
	Dt       int
}

var ProcStatmIndizes = map[string]int{
	"proc.statm.size":     0,
	"proc.statm.resident": 1,
	"proc.statm.share":    2,
	"proc.statm.text":     3,
	"proc.statm.lib":      4,
	"proc.statm.data":     5,
	"proc.statm.dt":       6,
}

func readStatm(filename string) (string, error) {
	pStatFile, err := ioutil.ReadFile(filename)

	if err != nil {
		return "", err
	}

	return string(pStatFile), nil
}

func ReadStatmArray(filename string) ([]string, error) {
	pStatmRaw, err := readStatm(filename)

	if err != nil {
		return nil, err
	}

	return strings.Split(pStatmRaw, " "), nil
}

func ReadStatm(filename string) (*ProcStatm, error) {
	pStatmRaw, err := readStatm(filename)

	if err != nil {
		return nil, err
	}

	p, err := ParseStatm(pStatmRaw)

	if err != nil {
		return nil, err
	}

	return p, nil
}

func ParseStatm(statm string) (*ProcStatm, error) {
	var pStatm = new(ProcStatm)

	fmt.Sscanf(statm, "%d %d %d %d %d %d %d", &pStatm.Size, &pStatm.Resident, &pStatm.Share, &pStatm.Text, &pStatm.Lib, &pStatm.Data, &pStatm.Dt)

	return pStatm, nil
}
