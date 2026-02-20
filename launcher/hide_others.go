//go:build !windows

package main

import "os/exec"

// hideWindow 非 Windows 平台无需特殊处理
func hideWindow(cmd *exec.Cmd) {}
