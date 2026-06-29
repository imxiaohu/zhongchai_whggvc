package utils

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/xiaohu/pingjiao/config"
)

// EnsureOCRService 检查并启动 OCR 服务。
// 优先级：已有服务 > 原生 Python > Docker
// 如果所有方案都不可用，仅记录警告，不阻塞主程序启动。
func EnsureOCRService() {
	ocrURL := config.GetOCRServURL() + "/health"

	// 1. 检查是否已经运行
	if isOCRServiceUp(ocrURL) {
		log.Println("[OCR] ddddocr 服务已就绪:", config.GetOCRServURL())
		return
	}

	log.Println("[OCR] 检测到 OCR 服务未运行，尝试自动启动...")

	// 2. 尝试原生 Python 启动（macOS/Linux）
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if err := tryStartNativeOCR(); err == nil {
			if waitForOCRService(ocrURL, 60*time.Second) {
				log.Println("[OCR] 原生 Python OCR 服务启动成功:", config.GetOCRServURL())
				return
			}
			log.Printf("[OCR] 原生 OCR 启动后健康检查未通过，尝试其他方案...")
		} else {
			log.Printf("[OCR] 原生启动失败: %v", err)
		}
	}

	// 3. 尝试 Docker 启动
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if err := tryStartDockerOCR(); err == nil {
			if waitForOCRService(ocrURL, 60*time.Second) {
				log.Println("[OCR] Docker OCR 服务启动成功:", config.GetOCRServURL())
				return
			}
			log.Printf("[OCR] Docker OCR 启动后健康检查未通过...")
		} else {
			log.Printf("[OCR] Docker 启动失败: %v", err)
		}
	}

	log.Println("[OCR] 警告: OCR 服务不可用，自动登录将降级为手动验证模式")
	log.Printf("[OCR] 如需启用 OCR，可手动运行: cd %s && docker compose --profile prod up -d", getDDDDOCRScriptDir())
}

// isOCRServiceUp 快速健康检查
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
	return resp.StatusCode == http.StatusOK
}

// waitForOCRService 等待服务就绪
func waitForOCRService(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if isOCRServiceUp(url) {
			return true
		}
		time.Sleep(2 * time.Second)
	}
	return false
}

// tryStartNativeOCR 尝试用 Python 原生启动 ddddocr
func tryStartNativeOCR() error {
	// 检查 Python3 是否可用
	pythonPath, err := exec.LookPath("python3")
	if err != nil {
		return fmt.Errorf("未找到 python3: %w", err)
	}

	// 检查 ddddocr 是否已安装并可用
	check := exec.Command(pythonPath, "-c", "import ddddocr; print('ok')")
	if err := check.Run(); err != nil {
		log.Println("[OCR] ddddocr 未安装，正在安装（可能需要几分钟）...")
		if err := installDDDDOCR(pythonPath); err != nil {
			return fmt.Errorf("安装 ddddocr 失败: %w", err)
		}
	}

	// 检查端口是否被占用
	if isPortInUse(8899) {
		return fmt.Errorf("端口 8899 被占用但服务不可用")
	}

	// 启动 ddddocr API 服务（后台运行，继承 stdout/stderr 以便查看日志）
	scriptDir := getDDDDOCRScriptDir()
	cmd := exec.Command(pythonPath, "-m", "ddddocr", "api",
		"--host=0.0.0.0", "--port=8899", "--workers=1")
	cmd.Dir = scriptDir
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 ddddocr 进程失败: %w", err)
	}

	log.Printf("[OCR] ddddocr 进程已启动 (pid=%d)，等待模型加载...", cmd.Process.Pid)
	return nil
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
	// macOS/Homebrew Python 需要 --break-system-packages
	// 不固定版本避免因 Python 版本差异导致无 wheel 可用而触发源码编译
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
	// 检查 docker 是否可用
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("未找到 docker: %w", err)
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

// getDDDDOCRScriptDir 获取 ddddocr-api 目录路径
func getDDDDOCRScriptDir() string {
	// 假设 go_backend 与 ddddocr-api 同级目录
	exe, _ := os.Executable()
	if exe != "" {
		base := filepath.Dir(exe)
		ddddocrDir := filepath.Join(base, "..", "ddddocr-api")
		if _, err := os.Stat(ddddocrDir); err == nil {
			abs, _ := filepath.Abs(ddddocrDir)
			return abs
		}
	}

	// 回退：尝试当前工作目录
	cwd, _ := os.Getwd()
	ddddocrDir := filepath.Join(cwd, "..", "ddddocr-api")
	if _, err := os.Stat(ddddocrDir); err == nil {
		abs, _ := filepath.Abs(ddddocrDir)
		return abs
	}

	// 再回退：当前目录
	return cwd
}

// isPortInUse 检查端口是否被占用
func isPortInUse(port int) bool {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
