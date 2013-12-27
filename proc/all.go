package proc

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type ProcAll struct {
	RSSize int64
	VSize  int64
}

var ProcAllIndizes = map[string]int{
	"proc.all.rssize": 0,
	"proc.all.vsize":  1,
}

func ReadAllArray(pid int) ([]string, error) {
	pAllRaw, err := ReadAll(pid)

	if err != nil {
		return nil, err
	}

	return []string{strconv.FormatInt(pAllRaw.RSSize, 10), strconv.FormatInt(pAllRaw.VSize, 10)}, nil
}

func ReadAll(pid int) (*ProcAll, error) {
	var pAllRaw ProcAll

	var cmd = exec.Command("ps", "-e", "--forest", "-o", "pid=", "-o", "ucmd=", "-o", "rss=", "-o", "vsz=")

	var out, err = cmd.CombinedOutput()

	if err != nil {
		return nil, err
	}

	var r = regexp.MustCompile(`^\s*(\d+)((?:\s+\|)?(?:\s+\\_)?)\s+.+?(\d+)\s+(\d+)$`)

	var pidS = strconv.FormatInt(int64(pid), 10)

	var inParent = false
	var parentDepth = ""

	for _, i := range strings.Split(string(out), "\n") {
		if i == "" {
			continue
		}

		o := r.FindStringSubmatch(i)
		if o == nil {
			return nil, fmt.Errorf("cannot match: %s\n", i)
		}

		if inParent || o[1] == pidS {
			if !inParent {
				parentDepth = o[2]
			} else if o[2] == parentDepth {
				// out of parent again

				break
			}

			inParent = true

			rssize, _ := strconv.ParseInt(o[3], 10, 64)
			pAllRaw.RSSize += rssize
			vsize, _ := strconv.ParseInt(o[4], 10, 64)
			pAllRaw.VSize += vsize
		}
	}

	return &pAllRaw, nil
}
