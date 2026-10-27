package platform

import (
	"path/filepath"
	"testing"
)

func TestDirsLinux(t *testing.T) {
	home := t.TempDir()

	tests := []struct {
		name      string
		dataHome  string
		stateHome string
		wantData  string
		wantLog   string
	}{
		{
			name:      "set",
			dataHome:  "/xdg/data",
			stateHome: "/xdg/state",
			wantData:  "/xdg/data/chiron",
			wantLog:   "/xdg/state/chiron",
		},
		{
			name:     "unset",
			wantData: filepath.Join(home, ".local", "share", "chiron"),
			wantLog:  filepath.Join(home, ".local", "state", "chiron"),
		},
		{
			name:      "relative",
			dataHome:  "xdg/data",
			stateHome: "./xdg/state",
			wantData:  filepath.Join(home, ".local", "share", "chiron"),
			wantLog:   filepath.Join(home, ".local", "state", "chiron"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", home)
			t.Setenv("XDG_DATA_HOME", tt.dataHome)
			t.Setenv("XDG_STATE_HOME", tt.stateHome)

			data, err := DataDir()
			if err != nil {
				t.Fatalf("DataDir: %v", err)
			}
			if data != tt.wantData {
				t.Errorf("DataDir = %q, want %q", data, tt.wantData)
			}

			logDir, err := LogDir()
			if err != nil {
				t.Fatalf("LogDir: %v", err)
			}
			if logDir != tt.wantLog {
				t.Errorf("LogDir = %q, want %q", logDir, tt.wantLog)
			}
		})
	}
}
