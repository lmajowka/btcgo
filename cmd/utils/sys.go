/**
 * BTCGO
 *
 * Modulo : System
 */

package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Get Exe Path
func GetPath() (string, error) {
	// Find app path
	exePath, err := os.Executable()
	if err == nil {
		return filepath.Dir(exePath), nil
	}
	return "", err
}

// Clear Console
func ClearConsole() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

// Get End of line
func GetEndLineChar() rune {
	charReadline := '\n'
	if runtime.GOOS == "windows" {
		charReadline = '\r'
	}
	return charReadline
}
