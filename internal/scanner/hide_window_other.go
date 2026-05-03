//go:build !windows

package scanner

import "os/exec"

// hideWindow 非 Windows 平台空实现
func hideWindow(cmd *exec.Cmd) {}
