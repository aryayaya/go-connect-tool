package checker

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"go-connect-tool/store"
)

type CheckResult struct {
	SiteID    string        `json:"site_id"`
	Status    string        `json:"status"` // "up", "down", "error"
	Latency   time.Duration `json:"latency"`
	Message   string        `json:"message"`
	Timestamp time.Time     `json:"timestamp"`
}

func CheckSite(site store.Site, proxyConfig store.ProxyConfig) CheckResult {
	start := time.Now()
	result := CheckResult{
		SiteID:    site.ID,
		Timestamp: start,
	}

	var err error
	var latency time.Duration

	switch site.Method {
	case "http", "https":
		latency, err = checkHTTP(site.URL, proxyConfig)
	case "tcp":
		latency, err = checkTCP(site.URL, proxyConfig)
	case "ping":
		latency, err = checkPing(site.URL)
	default:
		// Default to HTTP if unknown
		latency, err = checkHTTP(site.URL, proxyConfig)
	}

	result.Latency = latency
	if err != nil {
		result.Status = "down"
		result.Message = err.Error()
	} else {
		result.Status = "up"
		result.Message = "OK"
	}

	return result
}

func checkHTTP(targetURL string, proxyConfig store.ProxyConfig) (time.Duration, error) {
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "http://" + targetURL
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if proxyConfig.Enabled && proxyConfig.URL != "" {
		proxyURL, err := url.Parse(proxyConfig.URL)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	start := time.Now()
	resp, err := client.Get(targetURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return time.Since(start), fmt.Errorf("status code: %d", resp.StatusCode)
	}

	return time.Since(start), nil
}

func checkTCP(target string, proxyConfig store.ProxyConfig) (time.Duration, error) {
	// Simple TCP check.
	// Using proxy for TCP dialing typically requires a SOCKS client lib if we want to channel
	// raw TCP through the proxy. Since Go's net/http transport handles it for HTTP,
	// for raw TCP we might need "golang.org/x/net/proxy" if we want full SOCKS support here.
	// For simplicity, this initial version will do direct TCP if no proxy lib is added,
	// OR we can rely on system environment variables if standard net.Dial respects them (it usually doesn't for raw TCP).
	// If the user wants to test "connectivity" via proxy, HTTP check is best.
	// As a "small tool" request, I will adhere to standard lib.
	// If proxy is HTTP, we can't tunnel raw TCP easily without CONNECT.
	// SOCKS5 is supported by `x/net/proxy`. I will stick to direct connection for TCP for now
	// to keep dependencies zero, but mention it.

	// Actually, let's strip protocol if present
	if strings.Contains(target, "://") {
		parts := strings.Split(target, "://")
		if len(parts) > 1 {
			target = parts[1]
		}
	}

	// Assume port 80 if none
	if !strings.Contains(target, ":") {
		target = target + ":80"
	}

	d := net.Dialer{Timeout: 5 * time.Second}
	start := time.Now()
	conn, err := d.Dial("tcp", target)
	if err != nil {
		return 0, err
	}
	conn.Close()
	return time.Since(start), nil
}

func checkPing(target string) (time.Duration, error) {
	// OS execution of ping command.
	// Windows: ping -n 1 -w 1000 target
	// Linux/Mac: ping -c 1 -W 1 target

	if strings.Contains(target, "://") {
		parts := strings.Split(target, "://")
		if len(parts) > 1 {
			target = parts[1]
		}
	}
	// Remove port if present
	if strings.Contains(target, ":") {
		parts := strings.Split(target, ":")
		target = parts[0]
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", "2000", target)
	} else {
		cmd = exec.Command("ping", "-c", "1", "-W", "2", target)
	}

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		return 0, err
	}
	return duration, nil
}
