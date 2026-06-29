package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xiaohu/pingjiao/models"
)

func initTestDB(t *testing.T) {
	t.Helper()
	//nolint:errcheck
	os.Setenv("DB_TYPE", "sqlite")
	//nolint:errcheck
	os.Setenv("DB_PATH", filepath.Join(t.TempDir(), "test.db"))
	//nolint:errcheck
	//nolint:errcheck
	//nolint:errcheck
	os.Setenv("JWT_SECRET", "test_secret")
	//nolint:errcheck
	//nolint:errcheck
	//nolint:errcheck
	os.Setenv("SCHOOL_PASSWORD_ENC_KEY", "0123456789abcdef0123456789abcdef")
	//nolint:errcheck
	//nolint:errcheck
	os.Setenv("REDIS_ENABLED", "false")
	//nolint:errcheck
	os.Setenv("CACHE_ENABLED", "false")
	models.InitDB()
}

func TestAutoLogin_ConcurrentOnlyOneSchoolLogin(t *testing.T) {
	initTestDB(t)

	var loginHits int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/scloudoa/sys/mLogin" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true,"code":0,"message":"ok","result":{}}`))
			return
		}
		atomic.AddInt32(&loginHits, 1)
		var req map[string]string
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"code":0,"message":"ok","result":{"userInfo":{"id":"1","username":"u1","realname":"n","avatar":"","birthday":"","sex":0,"email":"","phone":"","className":"c","schoolId":1,"professionId":0,"facultyId":0,"gradeId":0,"currentSemester":"2024","identityCard":""},"token":"T"},"timestamp":0}`))
	}))
	defer upstream.Close()
	//nolint:errcheck
	os.Setenv("SCHOOL_SERVER_URL", upstream.URL)

	u := &models.User{Username: "u1", UserType: "student", Status: 1}
	if err := u.SetPassword("p1"); err != nil {
		t.Fatal(err)
	}
	if err := u.SetSchoolPassword("p1"); err != nil {
		t.Fatal(err)
	}
	u.TokenExpireAt = time.Now().Add(-time.Hour)
	if err := models.CreateUser(u); err != nil {
		t.Fatal(err)
	}

	pc := NewProxyClient()

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			userCopy, _ := models.FindUserByID(u.ID)
			_ = pc.AutoLogin(userCopy, false)
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&loginHits); got != 1 {
		t.Fatalf("expected 1 login hit, got %d", got)
	}

	updated, err := models.FindUserByID(u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.SchoolToken != "T" {
		t.Fatalf("expected token updated, got %q", updated.SchoolToken)
	}
}

