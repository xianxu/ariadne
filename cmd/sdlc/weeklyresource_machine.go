package main

import (
	"context"
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
)

const resourceCommandTimeout = 5 * time.Second
const resourceCacheEntryLimit = 2_000_000

var resourceBootSeconds = regexp.MustCompile(`\bsec = ([0-9]+)\b`)

type resourceCommandRunner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

type resourceExecRunner struct{}

func (resourceExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, resourceCommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	processgroup.Configure(cmd)
	cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
	cmd.WaitDelay = time.Second
	var stdout, stderr weeklyLimitWriter
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	_ = processgroup.Terminate(cmd, true)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if stdout.exceeded || stderr.exceeded {
		return nil, fmt.Errorf("%s output exceeded 8 MiB", name)
	}
	if err != nil {
		return stdout.Bytes(), fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

type resourceMachineDeps struct {
	Runner   resourceCommandRunner
	Capacity func(string) (total, available uint64, err error)
}

type ResourceMachine struct {
	GoCache           ResourceObservation `json:"go_cache"`
	AgentScratchFree  ResourceObservation `json:"agentscratch_free"`
	AgentScratchTotal ResourceObservation `json:"agentscratch_total"`
	BytesWritten      DeviceGauge         `json:"bytes_written"`
	SSDWear           DeviceGauge         `json:"ssd_wear"`
}

type ResourceMachineTrend struct {
	GoCache           ResourceComparison `json:"go_cache"`
	AgentScratchFree  ResourceComparison `json:"agentscratch_free"`
	AgentScratchTotal ResourceComparison `json:"agentscratch_total"`
	BytesWritten      DeviceGauge        `json:"bytes_written"`
	SSDWear           DeviceGauge        `json:"ssd_wear"`
}

func compareResourceMachine(current ResourceMachine, prior previousResourceResult, threshold float64) ResourceMachineTrend {
	var old ResourceMachineTrend
	if prior.Section != nil {
		old = prior.Section.Machine
	}
	priorBytes := func(row ResourceComparison) *ResourceObservation {
		if prior.State == "complete" && row.Kind != "" {
			return &row.ResourceObservation
		}
		return nil
	}
	result := ResourceMachineTrend{
		GoCache:           compareResourceBytes(current.GoCache, priorBytes(old.GoCache), threshold),
		AgentScratchFree:  compareResourceBytes(current.AgentScratchFree, priorBytes(old.AgentScratchFree), threshold),
		AgentScratchTotal: compareResourceBytes(current.AgentScratchTotal, priorBytes(old.AgentScratchTotal), threshold),
	}
	if prior.State == "complete" && old.BytesWritten.Device != "" {
		result.BytesWritten = compareDeviceCounter(current.BytesWritten, &old.BytesWritten)
	} else {
		result.BytesWritten = compareDeviceCounter(current.BytesWritten, nil)
	}
	if prior.State == "complete" && old.SSDWear.Device != "" {
		result.SSDWear = compareWearGauge(current.SSDWear, &old.SSDWear)
	} else {
		result.SSDWear = compareWearGauge(current.SSDWear, nil)
	}
	if prior.State == "unknown" {
		for _, row := range []*ResourceComparison{&result.GoCache, &result.AgentScratchFree, &result.AgentScratchTotal} {
			row.State, row.Reason = "unknown", prior.Reason
			row.PreviousBytes, row.DeltaBytes, row.GrowthShare = nil, nil, nil
			row.Alert = false
		}
		for _, row := range []*DeviceGauge{&result.BytesWritten, &result.SSDWear} {
			row.State, row.Reason, row.Delta = "unknown", prior.Reason, nil
		}
	}
	return result
}

func compareWearGauge(current DeviceGauge, prior *DeviceGauge) DeviceGauge {
	if current.State != "observed" {
		current.State = "unknown"
		return current
	}
	if prior == nil {
		current.State = "baseline"
		return current
	}
	if prior.Device != current.Device || (prior.State != "observed" && prior.State != "compared" && prior.State != "baseline") || current.Value < prior.Value || current.Value-prior.Value > math.MaxInt64 {
		current.State, current.Reason = "unknown", "previous wear unavailable, device changed, or counter reset"
		return current
	}
	delta := int64(current.Value - prior.Value)
	current.State, current.Delta = "compared", &delta
	return current
}

func resourceStatfsCapacity(path string) (uint64, uint64, error) {
	var stats syscall.Statfs_t
	if err := syscall.Statfs(path, &stats); err != nil {
		return 0, 0, err
	}
	return uint64(stats.Blocks) * uint64(stats.Bsize), uint64(stats.Bavail) * uint64(stats.Bsize), nil
}

func collectResourceMachine(ctx context.Context, deps resourceMachineDeps, scratchPath, disk string) ResourceMachine {
	result := ResourceMachine{}
	if deps.Runner == nil || deps.Capacity == nil {
		result.GoCache = ResourceObservation{Kind: "go-cache", Reason: "machine probe dependency missing"}
		result.AgentScratchFree = ResourceObservation{Kind: "agentscratch-free", Reason: "machine probe dependency missing"}
		result.AgentScratchTotal = ResourceObservation{Kind: "agentscratch-total", Reason: "machine probe dependency missing"}
		result.BytesWritten = DeviceGauge{Device: disk, State: "unknown", Reason: "machine probe dependency missing"}
		result.SSDWear = DeviceGauge{Device: disk, State: "unknown", Reason: "machine probe dependency missing"}
		return result
	}
	cacheOutput, cacheErr := deps.Runner.Run(ctx, "go", "env", "GOCACHE")
	cachePath := strings.TrimSpace(string(cacheOutput))
	result.GoCache = measureResourceDirectory("go-cache", cachePath)
	if cacheErr != nil {
		markResourceUnknown(&result.GoCache, "go env GOCACHE: "+cacheErr.Error())
	}
	total, free, capacityErr := deps.Capacity(scratchPath)
	result.AgentScratchFree = ResourceObservation{Kind: "agentscratch-free", Definition: scratchPath, Complete: capacityErr == nil}
	result.AgentScratchTotal = ResourceObservation{Kind: "agentscratch-total", Definition: scratchPath, Complete: capacityErr == nil}
	if capacityErr != nil || total > math.MaxInt64 || free > math.MaxInt64 {
		reason := fmt.Sprintf("AgentScratch capacity: %v", capacityErr)
		if capacityErr == nil {
			reason = "AgentScratch capacity exceeds int64"
		}
		markResourceUnknown(&result.AgentScratchFree, reason)
		markResourceUnknown(&result.AgentScratchTotal, reason)
	} else {
		result.AgentScratchFree.Bytes = int64(free)
		result.AgentScratchTotal.Bytes = int64(total)
	}
	ioregOutput, ioregErr := deps.Runner.Run(ctx, "ioreg", "-r", "-c", "IOBlockStorageDriver", "-w", "0", "-l", "-d", "2")
	bootOutput, bootErr := deps.Runner.Run(ctx, "sysctl", "-n", "kern.boottime")
	result.BytesWritten = DeviceGauge{Device: disk, State: "unknown"}
	if bytes, err := parseIORegWriteBytes(string(ioregOutput), disk); err == nil && ioregErr == nil {
		result.BytesWritten.Value = bytes
		result.BytesWritten.State = "observed"
		if bootErr == nil {
			if match := resourceBootSeconds.FindStringSubmatch(string(bootOutput)); len(match) == 2 {
				result.BytesWritten.Epoch = match[1]
			}
		}
		if result.BytesWritten.Epoch == "" {
			result.BytesWritten.Reason = "boot epoch unavailable; weekly delta will be unknown"
		}
	} else {
		result.BytesWritten.Reason = fmt.Sprintf("ioreg write counter: %v %v", err, ioregErr)
	}
	smartOutput, smartErr := deps.Runner.Run(ctx, "smartctl", "-a", "/dev/"+disk)
	result.SSDWear = DeviceGauge{Device: disk, State: "unknown"}
	if wear, err := parseSmartWear(string(smartOutput)); err == nil {
		result.SSDWear.Value, result.SSDWear.State = wear, "observed"
	} else {
		result.SSDWear.Reason = fmt.Sprintf("smartctl wear: %v %v", err, smartErr)
	}
	return result
}

func measureResourceDirectory(kind, root string) ResourceObservation {
	row := ResourceObservation{Kind: kind, Definition: root, Complete: true}
	if !filepath.IsAbs(root) {
		markResourceUnknown(&row, "absolute directory path unavailable")
		return row
	}
	visited := 0
	err := filepath.WalkDir(root, func(_ string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		visited++
		if visited > resourceCacheEntryLimit {
			return fmt.Errorf("directory exceeded %d entries", resourceCacheEntryLimit)
		}
		if entry.Type()&os.ModeSymlink != 0 || entry.IsDir() || !entry.Type().IsRegular() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		row.Bytes += info.Size()
		row.Files++
		return nil
	})
	if err != nil {
		markResourceUnknown(&row, err.Error())
	}
	return row
}
