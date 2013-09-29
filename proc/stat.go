package proc

import (
	"io/ioutil"
	"strconv"
	"strings"
)

type ProcStat struct {
	Pid                 int
	Comm                string
	State               byte
	Ppid                int
	Pgrp                int
	Session             int
	TtyNr               int
	Tpgid               int
	Flags               int
	Minflt              int
	Cminflt             int
	Majflt              int
	Cmajflt             int
	Utime               int
	Stime               int
	Cutime              int
	Cstime              int
	Priority            int
	Nice                int
	NumThreads          int
	Itrealvalue         int
	Starttime           int
	Vsize               int
	Rss                 int
	Rsslim              int
	Startcode           int
	Endcode             int
	Startstack          int
	Kstkesp             int
	Kstkeip             int
	Signal              int
	Blocked             int
	Sigignore           int
	Sigcatch            int
	Wchan               int
	Nswap               int
	Cnswap              int
	ExitSignal          int
	Processor           int
	RtPriority          int
	Policy              int
	DelayacctBlkioTicks int
	GuestTime           int
	CguestTime          int
}

var ProcStatIndizes = map[string]int{
	"proc.stat.pid":                   0,
	"proc.stat.comm":                  1,
	"proc.stat.state":                 2,
	"proc.stat.ppid":                  3,
	"proc.stat.pgrp":                  4,
	"proc.stat.session":               5,
	"proc.stat.tty_nr":                6,
	"proc.stat.tpgid":                 7,
	"proc.stat.flags":                 8,
	"proc.stat.minflt":                9,
	"proc.stat.cminflt":               10,
	"proc.stat.majflt":                11,
	"proc.stat.cmajflt":               12,
	"proc.stat.utime":                 13,
	"proc.stat.stime":                 14,
	"proc.stat.cutime":                15,
	"proc.stat.cstime":                16,
	"proc.stat.priority":              17,
	"proc.stat.nice":                  18,
	"proc.stat.num_threads":           19,
	"proc.stat.itrealvalue":           20,
	"proc.stat.starttime":             21,
	"proc.stat.vsize":                 22,
	"proc.stat.rss":                   23,
	"proc.stat.rsslim":                24,
	"proc.stat.startcode":             25,
	"proc.stat.endcode":               26,
	"proc.stat.startstack":            27,
	"proc.stat.kstkesp":               28,
	"proc.stat.kstkeip":               29,
	"proc.stat.signal":                30,
	"proc.stat.blocked":               31,
	"proc.stat.sigignore":             32,
	"proc.stat.sigcatch":              33,
	"proc.stat.wchan":                 34,
	"proc.stat.nswap":                 35,
	"proc.stat.cnswap":                36,
	"proc.stat.exit_signal":           37,
	"proc.stat.processor":             38,
	"proc.stat.rt_priority":           39,
	"proc.stat.policy":                40,
	"proc.stat.delayacct_blkio_ticks": 41,
	"proc.stat.guest_time":            42,
	"proc.stat.cguest_time":           43,
}

func readStat(filename string) (string, error) {
	pStatFile, err := ioutil.ReadFile(filename)

	if err != nil {
		return "", err
	}

	return string(pStatFile), nil
}

func ReadStatArray(filename string) ([]string, error) {
	pStatRaw, err := readStat(filename)

	if err != nil {
		return nil, err
	}

	return strings.Split(pStatRaw, " "), nil
}

func ReadStat(filename string) (*ProcStat, error) {
	pStatRaw, err := readStat(filename)

	if err != nil {
		return nil, err
	}

	p, err := ParseStat(pStatRaw)

	if err != nil {
		return nil, err
	}

	return p, nil
}

func ParseStat(stat string) (*ProcStat, error) {
	pStat := new(ProcStat)
	pStatRaw := strings.Split(stat, " ")

	// original scan format "%d (%s) %c %d %d %d %d %d %u %lu %lu %lu %lu %lu %lu %ld %ld %ld %ld %ld %ld %llu %lu %ld %lu %lu %lu %lu %lu %lu %lu %lu %lu %lu %lu %lu %lu %d %d %u %u %llu %lu %ld"

	pStat.Pid, _ = strconv.Atoi(pStatRaw[0])
	pStat.Comm = pStatRaw[1][1 : len(pStatRaw[1])-1]
	pStat.State = pStatRaw[2][0]
	pStat.Ppid, _ = strconv.Atoi(pStatRaw[3])
	pStat.Pgrp, _ = strconv.Atoi(pStatRaw[4])
	pStat.Session, _ = strconv.Atoi(pStatRaw[5])
	pStat.TtyNr, _ = strconv.Atoi(pStatRaw[6])
	pStat.Tpgid, _ = strconv.Atoi(pStatRaw[7])
	pStat.Flags, _ = strconv.Atoi(pStatRaw[8])
	pStat.Minflt, _ = strconv.Atoi(pStatRaw[9])
	pStat.Cminflt, _ = strconv.Atoi(pStatRaw[10])
	pStat.Majflt, _ = strconv.Atoi(pStatRaw[11])
	pStat.Cmajflt, _ = strconv.Atoi(pStatRaw[12])
	pStat.Utime, _ = strconv.Atoi(pStatRaw[13])
	pStat.Stime, _ = strconv.Atoi(pStatRaw[14])
	pStat.Cutime, _ = strconv.Atoi(pStatRaw[15])
	pStat.Cstime, _ = strconv.Atoi(pStatRaw[16])
	pStat.Priority, _ = strconv.Atoi(pStatRaw[17])
	pStat.Nice, _ = strconv.Atoi(pStatRaw[18])
	pStat.NumThreads, _ = strconv.Atoi(pStatRaw[19])
	pStat.Itrealvalue, _ = strconv.Atoi(pStatRaw[20])
	pStat.Starttime, _ = strconv.Atoi(pStatRaw[21])
	pStat.Vsize, _ = strconv.Atoi(pStatRaw[22])
	pStat.Rss, _ = strconv.Atoi(pStatRaw[23])
	pStat.Rsslim, _ = strconv.Atoi(pStatRaw[24])
	pStat.Startcode, _ = strconv.Atoi(pStatRaw[25])
	pStat.Endcode, _ = strconv.Atoi(pStatRaw[26])
	pStat.Startstack, _ = strconv.Atoi(pStatRaw[27])
	pStat.Kstkesp, _ = strconv.Atoi(pStatRaw[28])
	pStat.Kstkeip, _ = strconv.Atoi(pStatRaw[29])
	pStat.Signal, _ = strconv.Atoi(pStatRaw[30])
	pStat.Blocked, _ = strconv.Atoi(pStatRaw[31])
	pStat.Sigignore, _ = strconv.Atoi(pStatRaw[32])
	pStat.Sigcatch, _ = strconv.Atoi(pStatRaw[33])
	pStat.Wchan, _ = strconv.Atoi(pStatRaw[34])
	pStat.Nswap, _ = strconv.Atoi(pStatRaw[35])
	pStat.Cnswap, _ = strconv.Atoi(pStatRaw[36])
	pStat.ExitSignal, _ = strconv.Atoi(pStatRaw[37])
	pStat.Processor, _ = strconv.Atoi(pStatRaw[38])
	pStat.RtPriority, _ = strconv.Atoi(pStatRaw[39])
	pStat.Policy, _ = strconv.Atoi(pStatRaw[40])
	pStat.DelayacctBlkioTicks, _ = strconv.Atoi(pStatRaw[41])
	pStat.GuestTime, _ = strconv.Atoi(pStatRaw[42])
	pStat.CguestTime, _ = strconv.Atoi(pStatRaw[43])

	return pStat, nil
}
