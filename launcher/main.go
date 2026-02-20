package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/getlantern/systray"
)

const (
	appExe = "go-connect-tool.exe"
	appURL = "http://localhost:8080"
)

var (
	mu          sync.Mutex
	mainProcess *exec.Cmd
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(makeIcon())
	systray.SetTooltip("Go Connect Tool")

	mOpen := systray.AddMenuItem("打开管理界面", "在浏览器中打开 http://localhost:8080")
	mRestart := systray.AddMenuItem("重启服务", "重启主程序")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出", "停止服务并退出")

	// 启动主程序，稍后自动打开浏览器
	go func() {
		startMain()
	}()
	go func() {
		time.Sleep(1500 * time.Millisecond)
		openBrowser(appURL)
	}()

	// 菜单事件循环
	for {
		select {
		case <-mOpen.ClickedCh:
			openBrowser(appURL)
		case <-mRestart.ClickedCh:
			stopMain()
			go startMain()
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func onExit() {
	stopMain()
}

// exeDir 返回当前可执行文件所在目录
func exeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// startMain 启动主程序（阻塞直到进程结束）
func startMain() {
	dir := exeDir()
	appPath := filepath.Join(dir, appExe)

	cmd := exec.Command(appPath)
	cmd.Dir = dir
	hideWindow(cmd) // 平台相关：隐藏控制台窗口

	mu.Lock()
	if err := cmd.Start(); err != nil {
		mu.Unlock()
		return
	}
	mainProcess = cmd
	mu.Unlock()

	cmd.Wait()
}

// stopMain 强制终止主程序进程
func stopMain() {
	mu.Lock()
	defer mu.Unlock()
	if mainProcess != nil && mainProcess.Process != nil {
		mainProcess.Process.Kill()
		mainProcess = nil
	}
}

// openBrowser 用系统默认浏览器打开 URL
func openBrowser(url string) {
	switch runtime.GOOS {
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
		exec.Command("xdg-open", url).Start()
	}
}
