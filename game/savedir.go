package game

import (
	"os"
	"path/filepath"
	"runtime"
)

// SaveDir is the folder "Save & Quit" writes to and the start screen lists
// saved games from: $XDG_STATE_HOME/el-poblador, else the platform's usual
// place (~/.local/state/el-poblador, ~/Library/Application Support/el-poblador,
// %LocalAppData%\el-poblador). If no home folder can be found it falls back
// to the current folder, where saves went before.
func SaveDir() string {
	if d := os.Getenv("XDG_STATE_HOME"); d != "" {
		return filepath.Join(d, "el-poblador")
	}
	switch runtime.GOOS {
	case "windows":
		if d := os.Getenv("LocalAppData"); d != "" {
			return filepath.Join(d, "el-poblador")
		}
	case "darwin":
		if d, err := os.UserConfigDir(); err == nil {
			return filepath.Join(d, "el-poblador")
		}
	default:
		if h, err := os.UserHomeDir(); err == nil {
			return filepath.Join(h, ".local", "state", "el-poblador")
		}
	}
	return "."
}
