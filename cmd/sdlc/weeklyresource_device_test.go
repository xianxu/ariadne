package main

import (
	"strings"
	"testing"
)

func TestParseDeviceGauges(t *testing.T) {
	ioreg := `+-o IOBlockStorageDriver <class IOBlockStorageDriver, id 0x1>
  "Statistics" = {"Bytes (Write)"=0}
  +-o Other Media <class IOMedia>
    "BSD Name" = "disk4"
    "Removable" = Yes
+-o IOBlockStorageDriver <class IOBlockStorageDriver, id 0x2>
  "Statistics" = {"Bytes (Write)"=35732193742848}
  +-o APPLE SSD Media <class IOMedia>
    "BSD Name" = "disk0"
    "Removable" = No
`
	bytes, err := parseIORegWriteBytes(ioreg, "disk0")
	if err != nil || bytes != 35732193742848 {
		t.Fatalf("internal disk write counter=%d err=%v", bytes, err)
	}
	if _, err := parseIORegWriteBytes(ioreg, "disk3"); err == nil {
		t.Fatal("missing disk must remain unknown")
	}
	if _, err := parseIORegWriteBytes(ioreg+strings.ReplaceAll(ioreg, "35732193742848", "1"), "disk0"); err == nil {
		t.Fatal("ambiguous disk must remain unknown")
	}
	smart := "Percentage Used:                    8%\nRead 1 entries from Error Information Log failed\n"
	if wear, err := parseSmartWear(smart); err != nil || wear != 8 {
		t.Fatalf("wear should be usable despite unrelated smartctl exit error: %d %v", wear, err)
	}
	if _, err := parseSmartWear("Permission denied"); err == nil {
		t.Fatal("unreadable wear must remain unknown")
	}
}

func TestCompareDeviceCounterRequiresSameEpoch(t *testing.T) {
	prior := DeviceGauge{Device: "disk0", Epoch: "boot-1", Value: 100, Observed: true, State: "baseline"}
	for _, tc := range []struct {
		name    string
		current DeviceGauge
		state   string
		delta   *int64
	}{
		{"comparable", DeviceGauge{Device: "disk0", Epoch: "boot-1", Value: 130, Observed: true, State: "observed"}, "compared", ptrInt64(30)},
		{"reboot", DeviceGauge{Device: "disk0", Epoch: "boot-2", Value: 130, Observed: true, State: "observed"}, "unknown", nil},
		{"reset", DeviceGauge{Device: "disk0", Epoch: "boot-1", Value: 90, Observed: true, State: "observed"}, "unknown", nil},
		{"device", DeviceGauge{Device: "disk1", Epoch: "boot-1", Value: 130, Observed: true, State: "observed"}, "unknown", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := compareDeviceCounter(tc.current, &prior)
			if got.State != tc.state || !sameInt64(got.Delta, tc.delta) {
				t.Fatalf("comparison=%+v", got)
			}
		})
	}
}
