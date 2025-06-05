package utils

import (
	"bytes"
	"os/exec"
	"strings"
)

func IsMwAgentRunning() bool {
	cmd := exec.Command("pgrep", "-fl", "mw-agent")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false
	}

	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if strings.Contains(line, "mw-agent") && !strings.Contains(line, "grep") {
			return true
		}
	}
	return false
}
