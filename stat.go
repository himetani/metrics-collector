package main

import (
	"bufio"
	"context"
	"errors"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const darwinVmstatMock = "2  0      0 411848  23620 1379292    0    0     1     3   39   84  0  0 100  0  0"

var nowFn = time.Now

type Vmstat struct {
	wg     sync.WaitGroup
	db     DB
	ticker int // second
}

func (v *Vmstat) Run(ctx context.Context) error {
	vmstatCh := v.exec(ctx)

	for {
		select {
		case m := <-vmstatCh:
			v.db.Insert(m)
		case <-ctx.Done():
			v.wg.Done()
			return nil
		}
	}

	return nil
}

type metrics struct {
	datetime      time.Time
	running       uint64
	blocking      uint64
	swapped       uint64
	free          uint64
	buffer        uint64
	cache         uint64
	swapIn        uint64
	swapOut       uint64
	blockIn       uint64
	blockOut      uint64
	interapt      uint64
	contextSwitch uint64
	cpuUser       uint64
	cpuSystem     uint64
	cpuIdle       uint64
	cpuIowait     uint64
	cpuSteal      uint64
}

func convert(line string) (*metrics, error) {
	tmp := strings.Split(line, " ")
	lines := []string{}
	for _, v := range tmp {
		if v != "" {
			lines = append(lines, v)
		}
	}

	if len(lines) != 17 {
		return nil, errors.New("Invalid input")
	}

	l0, _ := strconv.ParseUint(lines[0], 10, 32)
	l1, _ := strconv.ParseUint(lines[1], 10, 32)
	l2, _ := strconv.ParseUint(lines[2], 10, 32)
	l3, _ := strconv.ParseUint(lines[3], 10, 32)
	l4, _ := strconv.ParseUint(lines[4], 10, 32)
	l5, _ := strconv.ParseUint(lines[5], 10, 32)
	l6, _ := strconv.ParseUint(lines[6], 10, 32)
	l7, _ := strconv.ParseUint(lines[7], 10, 32)
	l8, _ := strconv.ParseUint(lines[8], 10, 32)
	l9, _ := strconv.ParseUint(lines[9], 10, 32)
	l10, _ := strconv.ParseUint(lines[10], 10, 32)
	l11, _ := strconv.ParseUint(lines[11], 10, 32)
	l12, _ := strconv.ParseUint(lines[12], 10, 32)
	l13, _ := strconv.ParseUint(lines[13], 10, 32)
	l14, _ := strconv.ParseUint(lines[14], 10, 32)
	l15, _ := strconv.ParseUint(lines[15], 10, 32)
	l16, _ := strconv.ParseUint(lines[16], 10, 32)

	return &metrics{
		datetime:      nowFn(),
		running:       l0,
		blocking:      l1,
		swapped:       l2,
		free:          l3,
		buffer:        l4,
		cache:         l5,
		swapIn:        l6,
		swapOut:       l7,
		blockIn:       l8,
		blockOut:      l9,
		interapt:      l10,
		contextSwitch: l11,
		cpuUser:       l12,
		cpuSystem:     l13,
		cpuIdle:       l14,
		cpuIowait:     l15,
		cpuSteal:      l16,
	}, nil
}

func (v *Vmstat) exec(ctx context.Context) chan metrics {
	ch := make(chan metrics)

	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("vmstat", "-n", strconv.Itoa(v.ticker))
		stdout, _ := cmd.StdoutPipe()
		cmd.Start()

		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "procs -----------memory---------- ---swap-- -----io---- -system-- ------cpu-----") {
					continue
				}

				if strings.Contains(line, "r  b   swpd   free   buff  cache   si   so    bi    bo   in   cs us sy id wa st") {
					continue
				}

				vmstat, err := convert(line)
				if err != nil {
					panic(err)
				}
				ch <- *vmstat

			}
		}()
	case "darwin":
		go func() {
			ticker := time.NewTicker(time.Duration(v.ticker) * time.Second)
			for {
				select {
				case <-ticker.C:
					vmstat, err := convert(darwinVmstatMock)
					if err != nil {
						panic(err)
					}
					ch <- *vmstat
				case <-ctx.Done():
					return
				}
			}
		}()
	default:
		panic("Unsupported OS")
	}

	return ch
}