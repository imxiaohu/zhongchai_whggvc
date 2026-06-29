package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xiaohu/pingjiao/config"
)

var (
	autoLoginRedisOnce   sync.Once
	autoLoginRedisClient *redis.Client
	autoLoginRedisOK     bool
	autoLoginLocalLocks  sync.Map
)

func getAutoLoginRedis() (*redis.Client, bool) {
	autoLoginRedisOnce.Do(func() {
		enabled := config.GetRedisEnabled() && config.GetCacheEnabled() && (config.GetCacheType() == "redis" || config.GetCacheType() == "hybrid")
		if !enabled {
			return
		}
		client := redis.NewClient(&redis.Options{
			Addr:         fmt.Sprintf("%s:%s", config.GetRedisHost(), config.GetRedisPort()),
			Password:     config.GetRedisPassword(),
			DB:           config.GetRedisDB(),
			PoolSize:     config.GetRedisPoolSize(),
			DialTimeout:  time.Duration(config.GetRedisTimeout()) * time.Second,
			ReadTimeout:  time.Duration(config.GetRedisTimeout()) * time.Second,
			WriteTimeout: time.Duration(config.GetRedisTimeout()) * time.Second,
		})
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := client.Ping(ctx).Err(); err != nil {
			_ = client.Close()
			return
		}
		autoLoginRedisClient = client
		autoLoginRedisOK = true
	})
	return autoLoginRedisClient, autoLoginRedisOK
}

func withAutoLoginLock(userID uint, fn func() error) error {
	if client, ok := getAutoLoginRedis(); ok {
		return withRedisLock(client, fmt.Sprintf("auto_login_%d", userID), 45*time.Second, fn)
	}
	muAny, _ := autoLoginLocalLocks.LoadOrStore(userID, &sync.Mutex{})
	mu := muAny.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()
	return fn()
}

func withRedisLock(client *redis.Client, key string, ttl time.Duration, fn func() error) error {
	token, _ := randomToken()
	ctx := context.Background()
	//nolint:staticcheck
	acquired, err := client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return fn()
	}
	if !acquired {
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			time.Sleep(200 * time.Millisecond)
			//nolint:staticcheck
			acquired, _ = client.SetNX(ctx, key, token, ttl).Result()
			if acquired {
				break
			}
		}
		if !acquired {
			return fn()
		}
	}

	err = fn()
	_ = releaseRedisLock(client, key, token)
	return err
}

func releaseRedisLock(client *redis.Client, key string, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	const script = "if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end"
	return client.Eval(ctx, script, []string{key}, token).Err()
}

func randomToken() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func recordSchoolUnreachable(userID uint) {
	client, ok := getAutoLoginRedis()
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = client.Incr(ctx, "metrics:school_unreachable_total").Err()
	_ = client.Incr(ctx, fmt.Sprintf("metrics:school_unreachable_user:%d", userID)).Err()
}

func isSchoolUnreachableErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	needles := []string{"eof", "timeout", "connection refused", "connection reset", "no such host", "network is unreachable", "发送请求失败", "自动登录失败"}
	for _, n := range needles {
		if strings.Contains(msg, strings.ToLower(n)) {
			return true
		}
	}
	return false
}

