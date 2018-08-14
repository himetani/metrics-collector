package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"
)

func TestConvert(t *testing.T) {
	nowFn = func() time.Time {
		t, _ := time.Parse("2006-01-02", "2018-07-29")
		return t
	}

	line := "2  0      0 411848  23620 1379292    0    0     1     3   39   84  0  0 100  0  0"
	expected := helper()

	actual, err := convert(line)
	if err != nil {
		t.Errorf("Unexpected error happend. Msg: %s", err.Error())
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("TestConvert: expected = %+v, but got = %+v", expected, actual)
	}
}

func TestRun(t *testing.T) {
	prodMode = false

	buf := new(bytes.Buffer)
	mock := &dbMock{w: buf}
	vmstat := &Vmstat{
		db:     mock,
		ticker: 1,
	}

	vmstat.wg.Add(1)

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	vmstat.Run(ctx)
	vmstat.wg.Wait()

	actual := buf.String()
	expected := dumpedMetrics()

	if actual != expected {
		t.Errorf("Test Run: expected = %s, but got = %s", expected, actual)
	}
}

func helper() *metrics {
	return &metrics{
		Datetime:      nowFn(),
		Running:       2,
		Blocking:      0,
		Swapped:       0,
		Free:          411848,
		Buffer:        23620,
		Cache:         1379292,
		SwapIn:        0,
		SwapOut:       0,
		BlockIn:       1,
		BlockOut:      3,
		Interapt:      39,
		ContextSwitch: 84,
		CpuUser:       0,
		CpuSystem:     0,
		CpuIdle:       100,
		CpuIowait:     0,
		CpuSteal:      0,
	}

}

func dumpedMetrics() string {
	return fmt.Sprintf("%+v", *helper())
}

type dbMock struct {
	w io.Writer
}

func (d *dbMock) Insert(met metrics) error {
	fmt.Fprintf(d.w, "%+v", met)
	return nil
}
