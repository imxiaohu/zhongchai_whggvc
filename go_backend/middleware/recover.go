package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/utils"
)

// RecoverPanic returns a Gin middleware that recovers from any panic,
// logs the stack trace, and returns a generic 500 error to the client
// without exposing internal details.
func RecoverPanic() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] Recovered from panic: %v\n%s", r, debug.Stack())
				utils.ErrorResponse(c, http.StatusInternalServerError, "服务异常，请稍后重试")
				c.Abort()
			}
		}()
		c.Next()
	}
}
