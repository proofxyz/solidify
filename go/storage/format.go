package storage

import (
	"fmt"
	"os/exec"
)

// FormatSol formats a list of solidity files using the forge formatter.
func FormatSol(files []string) error {
	cmd := exec.Command("forge", "fmt")
	cmd.Args = append(cmd.Args, files...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("formatting sol files: %v", err)
	}
	return nil
}
