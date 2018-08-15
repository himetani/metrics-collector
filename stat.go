package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const vmstatMock = "2  0      0 411848  23620 1379292    0    0     1     3   39   84  0  0 100  0  0"

var (
	nowFn    = time.Now
	prodMode = true
)

type Vmstat struct {
	wg     sync.WaitGroup
	db     DB
	ticker int // second
}

func (v *Vmstat) Run(ctx context.Context) error {
	vmstatCh, err := v.exec(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case m := <-vmstatCh:
			v.db.Insert(m)
		case <-ctx.Done():
			fmt.Println("task.Run has been ended")
			v.wg.Done()
			return nil
		}
	}

	return nil
}

type metrics struct {
	Datetime      time.Time
	Running       uint64
	Blocking      uint64
	Swapped       uint64
	Free          uint64
	Buffer        uint64
	Cache         uint64
	SwapIn        uint64
	SwapOut       uint64
	BlockIn       uint64
	BlockOut      uint64
	Interapt      uint64
	ContextSwitch uint64
	CpuUser       uint64
	CpuSystem     uint64
	CpuIdle       uint64
	CpuIowait     uint64
	CpuSteal      uint64
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
		Datetime:      nowFn(),
		Running:       l0,
		Blocking:      l1,
		Swapped:       l2,
		Free:          l3,
		Buffer:        l4,
		Cache:         l5,
		SwapIn:        l6,
		SwapOut:       l7,
		BlockIn:       l8,
		BlockOut:      l9,
		Interapt:      l10,
		ContextSwitch: l11,
		CpuUser:       l12,
		CpuSystem:     l13,
		CpuIdle:       l14,
		CpuIowait:     l15,
		CpuSteal:      l16,
	}, nil
}

func (v *Vmstat) exec(ctx context.Context) (chan metrics, error) {
	ch := make(chan metrics)

	var stdout io.Reader
	if runtime.GOOS == "linux" && prodMode {
		cmd := exec.Command("vmstat", "-n", strconv.Itoa(v.ticker))
		stdout, _ = cmd.StdoutPipe()
		cmd.Start()
	} else {
		var w *io.PipeWriter
		stdout, w = io.Pipe()
		ticker := time.NewTicker(time.Duration(v.ticker) * time.Second)

		go func() {
			for {
				select {
				case <-ticker.C:
					fmt.Fprintf(w, "%s\n", vmstatMock)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

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

	return ch, nil
}
