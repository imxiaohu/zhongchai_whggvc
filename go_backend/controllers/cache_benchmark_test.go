package controllers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
)

func BenchmarkScoreList_CacheHit(b *testing.B) {
	gin.SetMode(gin.TestMode)
	//nolint:errcheck
	os.Setenv("DB_TYPE", "sqlite")
	//nolint:errcheck
	os.Setenv("DB_PATH", filepath.Join(b.TempDir(), "bench.db"))
	//nolint:errcheck
	os.Setenv("JWT_SECRET", "bench_secret")
	//nolint:errcheck
	os.Setenv("SCHOOL_PASSWORD_ENC_KEY", "0123456789abcdef0123456789abcdef")
	//nolint:errcheck
	os.Setenv("REDIS_ENABLED", "false")
	//nolint:errcheck
	os.Setenv("CACHE_ENABLED", "false")
	//nolint:errcheck
	os.Setenv("SCHOOL_SERVER_URL", "http://127.0.0.1:1")
	models.InitDB()

	user := &models.User{Username: "u1", UserType: "student", Status: 1}
	_ = user.SetPassword("p1")
	_ = user.SetSchoolPassword("p1")
	user.SchoolToken = "token"
	user.TokenExpireAt = time.Now().Add(24 * time.Hour)
	user.LastLoginAt = time.Now()
	_ = models.CreateUser(user)

	term := "2024-1"
	cachedRaw := []byte(`{"success":true,"code":0,"message":"ok","result":{"marker":"from_cache"}}`)
	offlineCache := services.GetOfflineCacheService()
	_, _ = offlineCache.UpsertPersonalScorePart(user.ID, term, "scoreList", cachedRaw, 12*time.Hour)

	pc := NewProxyController()
	r := gin.New()
	r.GET("/score", func(c *gin.Context) {
		c.Set("userId", fmt.Sprintf("%d", user.ID))
		pc.GetScoreList(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/score?currentSemester="+term, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("status=%d", w.Code)
		}
	}
}

