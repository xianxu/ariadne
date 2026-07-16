package project

import "testing"

func TestParsePhaseAStates(t *testing.T) {
	tests := []struct {
		name, body string
		hours      float64
		present    bool
		wantErr    bool
	}{
		{"absent", "notes only", 0, false, false},
		{"valid", "**phase-a:** 40h", 40, true, false},
		{"fraction", "**phase-a:** .5h", .5, true, false},
		{"malformed", "**phase-a:** soon", 0, true, true},
		{"zero", "**phase-a:** 0h", 0, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hours, present, err := ParsePhaseA(tt.body)
			if hours != tt.hours || present != tt.present || (err != nil) != tt.wantErr {
				t.Fatalf("ParsePhaseA = (%g,%v,%v), want (%g,%v,err=%v)", hours, present, err, tt.hours, tt.present, tt.wantErr)
			}
		})
	}
}
