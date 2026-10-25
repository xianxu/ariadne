package main

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/unix"
)

// hostLoadAverage is the 1, 5 and 15 minute load (#630), from vm.loadavg:
// struct loadavg { fixpt_t ldavg[3]; long fscale; } (fixpt_t is uint32).
func hostLoadAverage() ([3]float64, error) {
	raw, err := unix.SysctlRaw("vm.loadavg")
	if err != nil {
		return [3]float64{}, err
	}
	if len(raw) < 16 {
		return [3]float64{}, fmt.Errorf("vm.loadavg is %d bytes", len(raw))
	}
	scale := float64(binary.LittleEndian.Uint64(raw[len(raw)-8:]))
	if scale == 0 {
		return [3]float64{}, fmt.Errorf("vm.loadavg has no fscale")
	}
	var out [3]float64
	for i := range out {
		out[i] = float64(binary.LittleEndian.Uint32(raw[4*i:])) / scale
	}
	return out, nil
}
