package browser

import (
	"os/exec"
	"runtime"
)

// Open opens the specified URL in the default browser.
// On Windows, uses rundll32 instead of cmd /c start to prevent command injection.
func Open(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		// Use rundll32 to avoid shell injection via cmd /c start
		// rundll32 passes URL directly to the handler without shell interpretation
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}
