package controllers

import (
	"encoding/json"
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

func initTestDB(t *testing.T) {
	t.Helper()
	//nolint:errcheck
	os.Setenv("DB_TYPE", "sqlite")
	dbFile := filepath.Join(t.TempDir(), "test.db")
	//nolint:errcheck
	//nolint:errcheck
	os.Setenv("DB_PATH", dbFile)
	//nolint:errcheck
	//nolint:errcheck
	os.Setenv("JWT_SECRET", "test_secret")
	//nolint:errcheck
	os.Setenv("SCHOOL_PASSWORD_ENC_KEY", "0123456789abcdef0123456789abcdef")
	models.InitDB()
}

func TestScoreList_School503_ReturnsCached200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDB(t)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/scloudoa/scs/course/tCourseScore/getCourseScore" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"success":false,"code":503,"message":"down","result":null}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"code":0,"message":"ok","result":{}}`))
	//nolint:errcheck
	}))
	//nolint:errcheck
	defer upstream.Close()
	//nolint:errcheck
	os.Setenv("SCHOOL_SERVER_URL", upstream.URL)

	user := &models.User{Username: "u1", UserType: "student", Status: 1}
	if err := user.SetPassword("p1"); err != nil {
		t.Fatal(err)
	}
	_ = user.SetSchoolPassword("p1")
	user.SchoolToken = "token"
	user.TokenExpireAt = time.Now().Add(24 * time.Hour)
	user.LastLoginAt = time.Now()
	if err := models.CreateUser(user); err != nil {
		t.Fatal(err)
	}

	term := "2024-1"
	cachedRaw := []byte(`{"success":true,"code":0,"message":"ok","result":{"marker":"from_cache"}}`)
	offlineCache := services.GetOfflineCacheService()
	_, err := offlineCache.UpsertPersonalScorePart(user.ID, term, "scoreList", cachedRaw, 12*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	pc := NewProxyController()
	r := gin.New()
	r.GET("/score", func(c *gin.Context) {
		c.Set("userId", fmt.Sprintf("%d", user.ID))
		c.Request.URL.RawQuery = "currentSemester=" + term
		pc.GetScoreList(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/score?currentSemester="+term, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v body=%s", err, w.Body.String())
	}
	if resp["dataSourceType"] != string(DataSourceDatabase) {
		t.Fatalf("expected dataSourceType=database, got %v", resp["dataSourceType"])
	}
	result, _ := resp["result"].(map[string]interface{})
	if result["marker"] != "from_cache" {
		t.Fatalf("expected cached marker, got %v", result["marker"])
	}
}

