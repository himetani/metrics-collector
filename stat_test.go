package main

import (
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

func helper() *metrics {
	return &metrics{
		datetime:      nowFn(),
		running:       2,
		blocking:      0,
		swapped:       0,
		free:          411848,
		buffer:        23620,
		cache:         1379292,
		swapIn:        0,
		swapOut:       0,
		blockIn:       1,
		blockOut:      3,
		interapt:      39,
		contextSwitch: 84,
		cpuUser:       0,
		cpuSystem:     0,
		cpuIdle:       100,
		cpuIowait:     0,
		cpuSteal:      0,
	}

}
