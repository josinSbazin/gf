package browser

import (
	"os/exec"
	"runtime"
)

// Open opens the specified URL in the default browser
func Open(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}
