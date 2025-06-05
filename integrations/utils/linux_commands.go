package utils

import (
	"os/exec"
)

func IsMwAgentRunning() bool {
	cmd := exec.Command("pidof", "mw-agent")
	err := cmd.Run()
	return err == nil
}
