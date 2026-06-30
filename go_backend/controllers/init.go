package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math/big"
	mathrand "math/rand/v2"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/utils"
)

// captchaStore 验证码内存存储：key=token, value=code；一次性使用，验证成功立即删除
// 采用 token 而非 cookie 存储的原因：
//  1. 避免 CSRF：cookie 会被浏览器自动携带，无法区分"用户主动提交"与"跨站请求"
//  2. 避免多用户共享 cookie 互相覆盖
//  3. 与 sessionId/JSESSIONID 解耦，移动端/H5/小程序均能可靠使用
type captchaEntry struct {
	code      string
	createdAt time.Time
}

var (
	captchaStoreMu sync.RWMutex
	captchaStore   = make(map[string]captchaEntry)
)

const (
	captchaTTL         = 5 * time.Minute
	captchaStoreGCHook = 1 * time.Minute
)

// init 启动后台 goroutine 定期清理过期验证码
func init() {
	go func() {
		ticker := time.NewTicker(captchaStoreGCHook)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			captchaStoreMu.Lock()
			for k, v := range captchaStore {
				if now.Sub(v.createdAt) > captchaTTL {
					delete(captchaStore, k)
				}
			}
			captchaStoreMu.Unlock()
		}
	}()
}

// newCaptchaToken 生成不透明的随机 token
func newCaptchaToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// putCaptcha 写入验证码到内存
func putCaptcha(token, code string) {
	captchaStoreMu.Lock()
	captchaStore[token] = captchaEntry{code: code, createdAt: time.Now()}
	captchaStoreMu.Unlock()
}

// consumeCaptcha 校验并消费（一次性），校验成功立即删除
func consumeCaptcha(token, code string) bool {
	if token == "" || code == "" {
		return false
	}
	captchaStoreMu.Lock()
	defer captchaStoreMu.Unlock()
	entry, ok := captchaStore[token]
	if !ok {
		return false
	}
	if time.Since(entry.createdAt) > captchaTTL {
		delete(captchaStore, token)
		return false
	}
	delete(captchaStore, token)
	return strings.EqualFold(entry.code, code)
}

// ScloudInit 处理初始化请求
func ScloudInit(c *gin.Context) {
	// 生成随机clientId
	clientId, err := generateRandomString(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("生成clientId失败", 500))
		return
	}

	// 返回初始化数据
	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"clientId":  clientId,
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	}))
}

// ScloudValidateCode 生成验证码图片
// 响应中返回 captchaToken，前端必须在登录时一并提交（替代原 cookie 方案）
func ScloudValidateCode(c *gin.Context) {
	// 生成随机验证码
	code, err := generateRandomCode(4)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("生成验证码失败", 500))
		return
	}

	// 生成 token 并存储验证码
	token, err := newCaptchaToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("生成captcha token失败", 500))
		return
	}
	putCaptcha(token, code)

	// 设置响应头（图片本身）
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	// 通过响应头返回 token（非 Set-Cookie，CSRF 无法利用）
	c.Header("X-Captcha-Token", token)

	// 生成验证码图片
	img := generateCaptchaImage(code)

	// 输出图片
	if err := png.Encode(c.Writer, img); err != nil {
		return
	}
}

// generateRandomString 生成指定长度的随机字符串
func generateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}

// generateRandomCode 生成指定长度的随机数字验证码
func generateRandomCode(length int) (string, error) {
	const charset = "0123456789"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}

// generateCaptchaImage 生成验证码图片
// 内部使用 math/rand/v2（并发安全），不再使用共享锁保护的 math/rand
func generateCaptchaImage(code string) *image.RGBA {
	// 创建图片
	width, height := 120, 40
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 设置背景色
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{240, 240, 240, 255})
		}
	}

	// 添加干扰线
	for i := 0; i < 5; i++ {
		x1 := mathrand.IntN(width)
		y1 := mathrand.IntN(height)
		x2 := mathrand.IntN(width)
		y2 := mathrand.IntN(height)
		addLine(img, x1, y1, x2, y2, color.RGBA{uint8(mathrand.IntN(150)), uint8(mathrand.IntN(150)), uint8(mathrand.IntN(150)), 255})
	}

	// 添加干扰点
	for i := 0; i < 100; i++ {
		x := mathrand.IntN(width)
		y := mathrand.IntN(height)
		img.Set(x, y, color.RGBA{uint8(mathrand.IntN(255)), uint8(mathrand.IntN(255)), uint8(mathrand.IntN(255)), 255})
	}

	// 绘制验证码
	for i, char := range code {
		x := 20 + i*20 + mathrand.IntN(5)
		y := 20 + mathrand.IntN(10)
		addChar(img, x, y, char, color.RGBA{0, 0, 200, 255})
	}

	return img
}

// addLine 添加线条
func addLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := x2 - x1
	dy := y2 - y1
	steps := 0
	if abs(dx) > abs(dy) {
		steps = abs(dx)
	} else {
		steps = abs(dy)
	}

	xIncrement := float64(dx) / float64(steps)
	yIncrement := float64(dy) / float64(steps)
	x := float64(x1)
	y := float64(y1)

	for i := 0; i < steps; i++ {
		img.Set(int(x), int(y), c)
		x += xIncrement
		y += yIncrement
	}
}

// addChar 添加字符
func addChar(img *image.RGBA, x, y int, char rune, c color.RGBA) {
	// 简单的字符绘制，实际项目中可以使用字体库
	switch strings.ToUpper(string(char)) {
	case "0":
		addCircle(img, x, y, 8, c)
	case "1":
		addLine(img, x, y-10, x, y+10, c)
	case "2":
		addLine(img, x-5, y-10, x+5, y-10, c)
		addLine(img, x+5, y-10, x+5, y, c)
		addLine(img, x+5, y, x-5, y, c)
		addLine(img, x-5, y, x-5, y+10, c)
		addLine(img, x-5, y+10, x+5, y+10, c)
	case "3":
		addLine(img, x-5, y-10, x+5, y-10, c)
		addLine(img, x+5, y-10, x+5, y+10, c)
		addLine(img, x-5, y+10, x+5, y+10, c)
		addLine(img, x-5, y, x+5, y, c)
	case "4":
		addLine(img, x-5, y-10, x-5, y, c)
		addLine(img, x-5, y, x+5, y, c)
		addLine(img, x+5, y-10, x+5, y+10, c)
	case "5":
		addLine(img, x-5, y-10, x+5, y-10, c)
		addLine(img, x-5, y-10, x-5, y, c)
		addLine(img, x-5, y, x+5, y, c)
		addLine(img, x+5, y, x+5, y+10, c)
		addLine(img, x-5, y+10, x+5, y+10, c)
	case "6":
		addLine(img, x-5, y-10, x+5, y-10, c)
		addLine(img, x-5, y-10, x-5, y+10, c)
		addLine(img, x-5, y+10, x+5, y+10, c)
		addLine(img, x+5, y+10, x+5, y, c)
		addLine(img, x-5, y, x+5, y, c)
	case "7":
		addLine(img, x-5, y-10, x+5, y-10, c)
		addLine(img, x+5, y-10, x+5, y+10, c)
	case "8":
		addCircle(img, x, y-5, 5, c)
		addCircle(img, x, y+5, 5, c)
	case "9":
		addLine(img, x-5, y-10, x+5, y-10, c)
		addLine(img, x+5, y-10, x+5, y+10, c)
		addLine(img, x-5, y-10, x-5, y, c)
		addLine(img, x-5, y, x+5, y, c)
	default:
		addLine(img, x-5, y-10, x+5, y+10, c)
		addLine(img, x+5, y-10, x-5, y+10, c)
	}
}

// addCircle 添加圆形
func addCircle(img *image.RGBA, x, y, r int, c color.RGBA) {
	for i := -r; i <= r; i++ {
		for j := -r; j <= r; j++ {
			if i*i+j*j <= r*r {
				img.Set(x+i, y+j, c)
			}
		}
	}
}

// abs 取绝对值
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// GenerateRandomAvatar 生成随机头像
func GenerateRandomAvatar() string {
	// 生成随机颜色的头像
	width, height := 100, 100
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 设置背景色
	bgColor := color.RGBA{uint8(mathrand.IntN(200)), uint8(mathrand.IntN(200)), uint8(mathrand.IntN(200)), 255}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, bgColor)
		}
	}

	// 添加随机图案
	patternColor := color.RGBA{uint8(mathrand.IntN(100)), uint8(mathrand.IntN(100)), uint8(mathrand.IntN(100)), 255}
	for i := 0; i < 5; i++ {
		x := width / 2
		y := height / 2
		r := 10 + mathrand.IntN(30)
		addCircle(img, x, y, r, patternColor)
	}

	// 将图像转换为Base64字符串
	var buf strings.Builder
	_ = png.Encode(base64.NewEncoder(base64.StdEncoding, &buf), img)

	return "data:image/png;base64," + buf.String()
}
