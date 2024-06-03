package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func CopyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "echo "+text+" | clip")
	case "darwin":
		cmd = exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
	case "linux":
		// Check if wl-copy is available
		_, err := exec.LookPath("wl-copy")
		if err == nil {
			cmd = exec.Command("wl-copy")
			cmd.Stdin = strings.NewReader(text)
		} else {
			// Fall back to xclip if wl-copy is not available
			cmd = exec.Command("xclip", "-selection", "clipboard")
			cmd.Stdin = strings.NewReader(text)
		}
	default:
		return fmt.Errorf("unsupported platform")
	}

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to copy to clipboard: %v", err)
	}

	return nil
}
