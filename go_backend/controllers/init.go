package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math/big"
	mathrand "math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/utils"
)

// 创建一个本地随机数生成器
var localRand = mathrand.New(mathrand.NewSource(time.Now().UnixNano()))

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
func ScloudValidateCode(c *gin.Context) {
	// 生成随机验证码
	code, err := generateRandomCode(4)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("生成验证码失败", 500))
		return
	}

	// 将验证码存储到会话中（实际项目中应使用Redis或其他方式存储）
	// 这里简化处理，将验证码放入cookie
	c.SetCookie("captcha_code", code, 300, "/", "", false, true)

	// 生成验证码图片
	img := generateCaptchaImage(code)

	// 设置响应头
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// 输出图片
	_ = png.Encode(c.Writer, img)
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
		x1 := localRand.Intn(width)
		y1 := localRand.Intn(height)
		x2 := localRand.Intn(width)
		y2 := localRand.Intn(height)
		addLine(img, x1, y1, x2, y2, color.RGBA{uint8(localRand.Intn(150)), uint8(localRand.Intn(150)), uint8(localRand.Intn(150)), 255})
	}

	// 添加干扰点
	for i := 0; i < 100; i++ {
		x := localRand.Intn(width)
		y := localRand.Intn(height)
		img.Set(x, y, color.RGBA{uint8(localRand.Intn(255)), uint8(localRand.Intn(255)), uint8(localRand.Intn(255)), 255})
	}

	// 绘制验证码
	for i, char := range code {
		x := 20 + i*20 + localRand.Intn(5)
		y := 20 + localRand.Intn(10)
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
	bgColor := color.RGBA{uint8(localRand.Intn(200)), uint8(localRand.Intn(200)), uint8(localRand.Intn(200)), 255}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, bgColor)
		}
	}

	// 添加随机图案
	patternColor := color.RGBA{uint8(localRand.Intn(100)), uint8(localRand.Intn(100)), uint8(localRand.Intn(100)), 255}
	for i := 0; i < 5; i++ {
		x := width / 2
		y := height / 2
		r := 10 + localRand.Intn(30)
		addCircle(img, x, y, r, patternColor)
	}

	// 将图像转换为Base64字符串
	var buf strings.Builder
	_ = png.Encode(base64.NewEncoder(base64.StdEncoding, &buf), img)

	return "data:image/png;base64," + buf.String()
}