//go:build windows

package shell

import (
	"os/exec"
)

func ExecShell(file string) *exec.Cmd {
	return exec.Command("CMD", "/C", file)
}
