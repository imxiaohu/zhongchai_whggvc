package services

import (
	"errors"
	"os"
	"path/filepath"
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
	os.Setenv("JWT_SECRET", "test_secret")
	//nolint:errcheck
	os.Setenv("REDIS_ENABLED", "false")
	//nolint:errcheck
	os.Setenv("CACHE_ENABLED", "false")
	models.InitDB()
}

func TestGetOrRefreshNews_UsesFreshCache(t *testing.T) {
	initTestDB(t)
	s := GetOfflineCacheService()

	if err := s.PutNewsCache("k1", []byte(`{"ok":true}`), 30*time.Minute); err != nil {
		t.Fatal(err)
	}

	called := false
	res, err := s.GetOrRefreshNews("k1", 30*time.Minute, func() ([]byte, error) {
		called = true
		return []byte(`{"ok":false}`), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("expected fetch not called")
	}
	if string(res.Data) != `{"ok":true}` {
		t.Fatalf("unexpected data: %s", string(res.Data))
	}
}

func TestGetOrRefreshNews_FallsBackToStaleOnFetchError(t *testing.T) {
	initTestDB(t)
	s := GetOfflineCacheService()

	if err := s.PutNewsCache("k2", []byte(`{"v":1}`), 1*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond)

	res, err := s.GetOrRefreshNews("k2", 30*time.Minute, func() ([]byte, error) {
		return nil, errors.New("upstream down")
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(res.Data) != `{"v":1}` {
		t.Fatalf("unexpected data: %s", string(res.Data))
	}
	if !res.Stale {
		t.Fatal("expected stale=true")
	}
}

