package utils

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/xiaohu/pingjiao/config"
)

// ocrChildProc 用于持有原生 Python 启动的 ddddocr 子进程句柄，
// 便于在主进程退出或健康检查失败时主动清理，避免遗留孤儿进程占端口。
type ocrChildProc struct {
	cmd    *exec.Cmd
	pid    int
	exited atomic.Bool // 标记是否已退出，防止重复 Wait/清理
}

func (p *ocrChildProc) isAlive() bool {
	if p == nil || p.cmd == nil || p.cmd.Process == nil {
		return false
	}
	if p.exited.Load() {
		return false
	}
	// ProcessState 非空说明已被 Wait 过
	if p.cmd.ProcessState != nil {
		return false
	}
	return true
}

// EnsureOCRService 检查并启动 OCR 服务。
// 优先级：已有服务 > 原生 Python（start_ocr.sh > ocr_server.py > python -m ddddocr）
// 如果所有方案都不可用，仅记录警告，不阻塞主程序启动。
func EnsureOCRService() {
	ocrURL := config.GetOCRServURL() + "/health"

	// 1. 已经运行则直接返回
	if isOCRServiceUp(ocrURL) {
		log.Println("[OCR] ddddocr 服务已就绪:", config.GetOCRServURL())
		return
	}

	log.Println("[OCR] 检测到 OCR 服务未运行，尝试自动启动...")

	// 解析超时/轮询配置（env > 默认）
	timeout := envDurationSec("OCR_START_TIMEOUT_SEC", 180)
	pollInterval := envDurationSec("OCR_START_POLL_SEC", 2)
	if timeout < 30*time.Second {
		timeout = 30 * time.Second
	}
	if pollInterval < 500*time.Millisecond {
		pollInterval = 500 * time.Millisecond
	}

	// 2. 尝试原生 Python 启动
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		child, err := tryStartNativeOCR()
		if err != nil {
			log.Printf("[OCR] 原生启动失败: %v", err)
		} else if waitForOCRServiceWithProcess(ocrURL, child, timeout, pollInterval) {
			log.Println("[OCR] 原生 Python OCR 服务启动成功:", config.GetOCRServURL())
			return
		} else {
			log.Printf("[OCR] 原生 OCR 启动后健康检查未通过（超时 %s）...", timeout)
			killChild(child)
		}
	}

	// 3. 尝试 Docker 启动
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if err := tryStartDockerOCR(); err == nil {
			if waitForOCRService(ocrURL, timeout, pollInterval) {
				log.Println("[OCR] Docker OCR 服务启动成功:", config.GetOCRServURL())
				return
			}
			log.Printf("[OCR] Docker OCR 启动后健康检查未通过（超时 %s）...", timeout)
		} else {
			log.Printf("[OCR] Docker 启动失败: %v", err)
		}
	}

	log.Println("[OCR] 警告: OCR 服务不可用，自动登录将降级为手动验证模式")
	log.Printf("[OCR] 如需启用 OCR，可手动运行: cd %s && ./start_ocr.sh", getDDDDOCRScriptDir())
}

// envDurationSec 读取秒级时间配置，不合法时使用默认值。
func envDurationSec(key string, defSec int) time.Duration {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return time.Duration(defSec) * time.Second
}

// isOCRServiceUp 快速健康检查（单次，不重试）
func isOCRServiceUp(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	// 额外确认：读取响应体，确保不是服务占着端口但 /health 返回空体。
	// ddddocr 的 /health 应当返回非空文本。
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
	return len(body) > 0
}

// waitForOCRService 轮询等待 OCR 服务就绪
func waitForOCRService(url string, timeout, pollInterval time.Duration) bool {
	return waitForOCRServiceWithProcess(url, nil, timeout, pollInterval)
}

// waitForOCRServiceWithProcess 在轮询健康检查的同时监听子进程是否已退出；
// 如果子进程先挂了，立刻返回 false，避免长时间等待一个已经死掉的进程。
func waitForOCRServiceWithProcess(url string, child *ocrChildProc, timeout, pollInterval time.Duration) bool {
	deadline := time.Now().Add(timeout)
	attempt := 0
	for time.Now().Before(deadline) {
		// 子进程监控：先于 HTTP 检查，避免对已死进程做无用探测
		if child != nil && !child.isAlive() {
			log.Printf("[OCR] ddddocr 子进程 (pid=%d) 已退出，放弃等待", child.pid)
			return false
		}

		if isOCRServiceUp(url) {
			return true
		}

		attempt++
		// 退避策略：前 5 次每 pollInterval 一次，之后翻倍，最大 10s
		sleep := pollInterval
		if attempt > 5 {
			sleep *= 2
			if sleep > 10*time.Second {
				sleep = 10 * time.Second
			}
		}
		time.Sleep(sleep)
	}
	return false
}

// tryStartNativeOCR 尝试用 Python 原生启动 ddddocr。
// 优先级：start_ocr.sh（推荐） > python ocr_server.py > python -m ddddocr api
// 不依赖 setsid，统一通过 bash 脚本内部 nohup 后台执行，macOS/Linux 兼容。
// 返回 *ocrChildProc 用于后续存活监控和失败时清理。
func tryStartNativeOCR() (*ocrChildProc, error) {
	pythonPath, err := exec.LookPath("python3")
	if err != nil {
		return nil, fmt.Errorf("未找到 python3: %w", err)
	}

	// 检查 ddddocr 是否已安装
	check := exec.Command(pythonPath, "-c", "import ddddocr; print('ok')")
	if err := check.Run(); err != nil {
		log.Println("[OCR] ddddocr 未安装，正在安装（可能需要几分钟）...")
		if err := installDDDDOCR(pythonPath); err != nil {
			return nil, fmt.Errorf("安装 ddddocr 失败: %w", err)
		}
	}

	// 检查端口是否已被占用
	if isPortInUse(8899) {
		return nil, fmt.Errorf("端口 8899 被占用但 /health 检查未通过，可能有残留进程")
	}

	scriptDir := getDDDDOCRScriptDir()

	// 独立日志文件：写到部署目录，避免 stdout 被 Go 接管
	logFile := filepath.Join(scriptDir, "ocr.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Printf("[OCR] 无法打开日志文件 %s（继续使用 stdout）: %v", logFile, err)
		f = nil
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()

	// 构建启动命令，分三级兜底
	tryMethods := []struct {
		name    string
		probe   func() bool
		makeCmd func() *exec.Cmd
	}{
		{
			name: "start_ocr.sh",
			probe: func() bool {
				sh := filepath.Join(scriptDir, "start_ocr.sh")
				info, _ := os.Stat(sh)
				return info != nil && info.Mode().IsRegular()
			},
			// 脚本内部通过 nohup 后台执行，跨平台兼容（Linux + macOS）
			makeCmd: func() *exec.Cmd {
				return exec.Command("bash", filepath.Join(scriptDir, "start_ocr.sh"))
			},
		},
		{
			name: "ocr_server.py",
			probe: func() bool {
				py := filepath.Join(scriptDir, "ocr_server.py")
				info, _ := os.Stat(py)
				return info != nil && info.Mode().IsRegular()
			},
			makeCmd: func() *exec.Cmd {
				cmd := exec.Command(pythonPath, filepath.Join(scriptDir, "ocr_server.py"),
					"--host", "0.0.0.0", "--port", "8899", "--workers", "1")
				cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
				return cmd
			},
		},
		{
			name:  "python -m ddddocr api",
			probe: func() bool { return true },
			makeCmd: func() *exec.Cmd {
				cmd := exec.Command(pythonPath, "-m", "ddddocr", "api",
					"--host=0.0.0.0", "--port=8899", "--workers=1")
				cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
				return cmd
			},
		},
	}

	for _, m := range tryMethods {
		if !m.probe() {
			log.Printf("[OCR] 跳过 %s（文件不存在）", m.name)
			continue
		}
		log.Printf("[OCR] 尝试启动方式: %s", m.name)
		cmd := m.makeCmd()
		cmd.Dir = scriptDir
		cmd.Env = append(os.Environ(),
			"OCR_HOST=0.0.0.0",
			"OCR_PORT=8899",
			"OCR_WORKERS=1",
		)

		if f != nil {
			cmd.Stdout = f
			cmd.Stderr = f
		} else {
			cmd.Stdout = prefixWriter("[OCR-py.stdout] ", os.Stdout)
			cmd.Stderr = prefixWriter("[OCR-py.stderr] ", os.Stderr)
		}

		if err := cmd.Start(); err != nil {
			log.Printf("[OCR] 启动 %s 失败: %v", m.name, err)
			continue
		}

		child := &ocrChildProc{cmd: cmd, pid: cmd.Process.Pid}

		go func(name string, c *ocrChildProc) {
			waitErr := c.cmd.Wait()
			c.exited.Store(true)
			if waitErr != nil {
				log.Printf("[OCR] %s 子进程 (pid=%d) 已退出: %v", name, c.pid, waitErr)
			} else {
				log.Printf("[OCR] %s 子进程 (pid=%d) 已退出", name, c.pid)
			}
		}(m.name, child)

		log.Printf("[OCR] %s 进程已启动 (pid=%d)，等待服务就绪...", m.name, child.pid)
		return child, nil
	}

	return nil, fmt.Errorf("所有启动方式均不可用")
}

// prefixWriter 返回一个 io.Writer，给每行加 prefix 后写入底层 w。
// 用于把子进程的输出和 Go 主进程的日志区分开，便于排障。
func prefixWriter(prefix string, w io.Writer) io.Writer {
	pr, pw := io.Pipe()
	go func() {
		scanner := bufio.NewScanner(pr)
		// 单行最大 1MB，ddddocr 在加载模型时可能打印较长的 tensor shape
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			// 注意：log.Println 不保证原子写，但单条 line 是一次 Scan()
			// 这里直接用 fmt.Fprintf 写到原 w，避免引入多 goroutine 竞争。
			fmt.Fprintf(w, "%s%s\n", prefix, line)
		}
	}()
	return pw
}

// killChild 在启动失败或主进程退出时清理子进程，
// 特别处理 setsid 创建的会话首进程（SIGTERM→SIGKILL）。
func killChild(child *ocrChildProc) {
	if child == nil || child.cmd == nil || child.cmd.Process == nil {
		return
	}
	if !child.isAlive() {
		return
	}
	if runtime.GOOS == "windows" {
		_ = child.cmd.Process.Kill()
	} else {
		// setsid 启动的进程是会话首进程，Process.Kill() 只杀主进程；
		// 为保险先 SIGTERM 让进程组内所有进程收到终止信号。
		_ = syscall.Kill(-child.pid, syscall.SIGTERM)
		time.Sleep(500 * time.Millisecond)
		if child.isAlive() {
			_ = syscall.Kill(-child.pid, syscall.SIGKILL)
		}
	}
	child.exited.Store(true)
}

// installDDDDOCR 通过 pip 安装 ddddocr 及其 API 服务依赖
func installDDDDOCR(pythonPath string) error {
	// 先安装系统依赖（Linux 需要）
	if runtime.GOOS == "linux" {
		installDeps := exec.Command("sudo", "apt-get", "install", "-y",
			"libgl1-mesa-glx", "libglib2.0-0", "libgomp1")
		installDeps.Stdout = os.Stdout
		installDeps.Stderr = os.Stderr
		if err := installDeps.Run(); err != nil {
			log.Printf("[OCR] 安装系统依赖失败（继续尝试）: %v", err)
		}
	}

	// pip 安装 ddddocr 及 API 服务所需依赖
	pipBase := []string{pythonPath, "-m", "pip", "install", "--quiet"}
	packages := []string{
		"ddddocr",
		"fastapi>=0.100",
		"uvicorn[standard]",
		"python-multipart",
	}
	if runtime.GOOS == "darwin" {
		pipBase = append(pipBase, "--break-system-packages")
	}

	for _, pkg := range packages {
		args := append(pipBase, pkg)
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("安装 %s 失败: %w", pkg, err)
		}
	}

	return nil
}

// tryStartDockerOCR 尝试用 Docker 启动
func tryStartDockerOCR() error {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("未找到 docker: %w（PATH=%s）", err, os.Getenv("PATH"))
	}

	// 检查 docker daemon 是否运行
	checkDocker := exec.Command(dockerPath, "info")
	if err := checkDocker.Run(); err != nil {
		return fmt.Errorf("docker daemon 未运行: %w", err)
	}

	scriptDir := getDDDDOCRScriptDir()
	cmd := exec.Command(dockerPath, "compose", "--profile", "prod", "up", "-d")
	cmd.Dir = scriptDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose up 失败: %w", err)
	}

	return nil
}

// getDDDDOCRScriptDir 获取 ddddocr-api 目录路径。
// 优先级：
//  1. 部署包结构：<deploy>/ddddocr-api/（Go 服务在 <deploy>/pingjiao_server）
//  2. 开发结构：<repo>/go_backend/ 的上一级（Go 服务在 <repo>/go_backend/pingjiao_server）
//  3. 当前工作目录
func getDDDDOCRScriptDir() string {
	// 1. 部署包结构：exe 同级目录下有 ddddocr-api 子目录
	//    场景：/www/wwwroot/zhongchai/pingjiao_server 与
	//          /www/wwwroot/zhongchai/ddddocr-api/ 同级
	exe, _ := os.Executable()
	if exe != "" {
		base := filepath.Dir(exe)

		// 尝试 <exe_dir>/ddddocr-api
		ddddocrDir := filepath.Join(base, "ddddocr-api")
		if info, err := os.Stat(ddddocrDir); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(ddddocrDir)
			return abs
		}

		// 尝试 <exe_dir>/../ddddocr-api（Go 服务在子目录的情况）
		parentDir := filepath.Join(base, "..")
		ddddocrDir = filepath.Join(parentDir, "ddddocr-api")
		if info, err := os.Stat(ddddocrDir); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(ddddocrDir)
			return abs
		}
	}

	// 2. 开发结构：go_backend 与 ddddocr-api 同级
	cwd, _ := os.Getwd()

	// 尝试 <cwd>/ddddocr-api（Go 服务在 go_backend/）
	ddddocrDir := filepath.Join(cwd, "ddddocr-api")
	if info, err := os.Stat(ddddocrDir); err == nil && info.IsDir() {
		abs, _ := filepath.Abs(ddddocrDir)
		return abs
	}

	// 尝试 <cwd>/../ddddocr-api（Go 服务在 go_backend/）
	ddddocrDir = filepath.Join(cwd, "..", "ddddocr-api")
	if info, err := os.Stat(ddddocrDir); err == nil && info.IsDir() {
		abs, _ := filepath.Abs(ddddocrDir)
		return abs
	}

	// 3. 回退：当前目录
	return cwd
}

// isPortInUse 检查端口是否被占用（TCP connect 探测）
func isPortInUse(port int) bool {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
