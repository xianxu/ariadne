package sysvshm

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func kernelWitness(pid, pgid int) (Witness, bool) {
	raw, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil || len(raw) > MaxBytes {
		return Witness{}, false
	}
	s := string(raw)
	first, last := strings.IndexByte(s, '('), strings.LastIndexByte(s, ')')
	if first < 0 || last < first {
		return Witness{}, false
	}
	fields := strings.Fields(s[last+1:])
	if len(fields) < 20 {
		return Witness{}, false
	}
	group, err := strconv.Atoi(fields[2])
	if err != nil || group != pgid {
		return Witness{}, false
	}
	return Witness{Name: s[first+1 : last], Start: fields[19]}, true
}
