package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	ioregWriteBytes = regexp.MustCompile(`"Bytes \(Write\)"=([0-9]+)`)
	smartWearLine   = regexp.MustCompile(`(?m)^Percentage Used:\s*([0-9]+)%`)
)

type DeviceGauge struct {
	Device   string `json:"device"`
	Epoch    string `json:"epoch,omitempty"`
	Value    uint64 `json:"value"`
	Observed bool   `json:"observed"`
	State    string `json:"state"`
	Reason   string `json:"reason,omitempty"`
	Delta    *int64 `json:"delta"`
}

// parseIORegWriteBytes binds a driver statistic to its whole, nonremovable
// IOMedia child, rather than picking the largest counter among disk images.
func parseIORegWriteBytes(output, disk string) (uint64, error) {
	var found []uint64
	for _, block := range strings.Split(output, "+-o IOBlockStorageDriver") {
		if !strings.Contains(block, `"BSD Name" = "`+disk+`"`) || !strings.Contains(block, `"Removable" = No`) {
			continue
		}
		match := ioregWriteBytes.FindStringSubmatch(block)
		if len(match) != 2 {
			return 0, fmt.Errorf("ioreg has no Bytes (Write) for %s", disk)
		}
		value, err := strconv.ParseUint(match[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("ioreg Bytes (Write) for %s: %w", disk, err)
		}
		found = append(found, value)
	}
	if len(found) != 1 {
		return 0, fmt.Errorf("ioreg matched %d nonremovable drivers for %s", len(found), disk)
	}
	return found[0], nil
}

func parseSmartWear(output string) (uint64, error) {
	match := smartWearLine.FindStringSubmatch(output)
	if len(match) != 2 {
		return 0, fmt.Errorf("smartctl did not report Percentage Used")
	}
	value, err := strconv.ParseUint(match[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("smartctl Percentage Used: %w", err)
	}
	return value, nil
}

func compareDeviceCounter(current DeviceGauge, prior *DeviceGauge) DeviceGauge {
	if !current.Observed {
		current.State = "unknown"
		return current
	}
	if prior == nil {
		current.State = "baseline"
		return current
	}
	if !prior.Observed {
		current.State = "unknown"
		current.Reason = "previous device counter unknown"
		return current
	}
	if current.Device != prior.Device || current.Epoch == "" || current.Epoch != prior.Epoch || current.Value < prior.Value || current.Value-prior.Value > math.MaxInt64 {
		current.State = "unknown"
		current.Reason = "device changed, counter epoch changed, or counter reset"
		return current
	}
	delta := int64(current.Value - prior.Value)
	current.Delta = &delta
	current.State = "compared"
	return current
}
