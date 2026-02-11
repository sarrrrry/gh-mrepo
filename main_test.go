package main

import (
	"testing"
)

func TestExtractAllFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantRest []string
		wantAll  bool
	}{
		{
			name:     "no flags",
			args:     []string{},
			wantRest: nil,
			wantAll:  false,
		},
		{
			name:     "--all flag",
			args:     []string{"--all"},
			wantRest: nil,
			wantAll:  true,
		},
		{
			name:     "-a flag",
			args:     []string{"-a"},
			wantRest: nil,
			wantAll:  true,
		},
		{
			name:     "--all with other args",
			args:     []string{"--all", "--limit", "5"},
			wantRest: []string{"--limit", "5"},
			wantAll:  true,
		},
		{
			name:     "-a with other args",
			args:     []string{"-a", "--limit", "5"},
			wantRest: []string{"--limit", "5"},
			wantAll:  true,
		},
		{
			name:     "other args only",
			args:     []string{"--limit", "5"},
			wantRest: []string{"--limit", "5"},
			wantAll:  false,
		},
		{
			name:     "--all between other args",
			args:     []string{"--limit", "5", "--all", "--json", "name"},
			wantRest: []string{"--limit", "5", "--json", "name"},
			wantAll:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRest, gotAll := extractAllFlag(tt.args)
			if gotAll != tt.wantAll {
				t.Errorf("all = %v, want %v", gotAll, tt.wantAll)
			}
			if len(gotRest) != len(tt.wantRest) {
				t.Errorf("rest = %v, want %v", gotRest, tt.wantRest)
				return
			}
			for i := range gotRest {
				if gotRest[i] != tt.wantRest[i] {
					t.Errorf("rest[%d] = %q, want %q", i, gotRest[i], tt.wantRest[i])
				}
			}
		})
	}
}

func TestExtractLlsFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantAll  bool
		wantJSON bool
	}{
		{name: "no flags", args: []string{}, wantAll: false, wantJSON: false},
		{name: "--all", args: []string{"--all"}, wantAll: true, wantJSON: false},
		{name: "-a", args: []string{"-a"}, wantAll: true, wantJSON: false},
		{name: "--json", args: []string{"--json"}, wantAll: false, wantJSON: true},
		{name: "-j", args: []string{"-j"}, wantAll: false, wantJSON: true},
		{name: "-a -j separate", args: []string{"-a", "-j"}, wantAll: true, wantJSON: true},
		{name: "-aj combined", args: []string{"-aj"}, wantAll: true, wantJSON: true},
		{name: "-ja combined", args: []string{"-ja"}, wantAll: true, wantJSON: true},
		{name: "--all --json", args: []string{"--all", "--json"}, wantAll: true, wantJSON: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAll, gotJSON := extractLlsFlags(tt.args)
			if gotAll != tt.wantAll {
				t.Errorf("all = %v, want %v", gotAll, tt.wantAll)
			}
			if gotJSON != tt.wantJSON {
				t.Errorf("json = %v, want %v", gotJSON, tt.wantJSON)
			}
		})
	}
}
