//go:build windows

package scanner

import (
	"os/exec"
	"syscall"
)

// hideWindow 设置子进程不弹出控制台窗口（仅 Windows 生效）
func hideWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= 0x08000000 // CREATE_NO_WINDOW
}
