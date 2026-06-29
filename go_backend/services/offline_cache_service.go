package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/xiaohu/pingjiao/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OfflineCacheService struct {
	redis *RedisService
}

type CacheMeta struct {
	FetchedAt time.Time
	ExpiresAt time.Time
}

type CacheResult struct {
	Data      []byte
	Meta      CacheMeta
	FromCache bool
	Stale     bool
}

type PersonalScoreBundle struct {
	ScoreList            json.RawMessage   `json:"scoreList,omitempty"`
	SemesterScore        json.RawMessage   `json:"semesterScore,omitempty"`
	EvaluationConfigList json.RawMessage   `json:"evaluationConfigList,omitempty"`
	UpdatedAt            map[string]string `json:"updatedAt,omitempty"`
}

type ClassScheduleBundle struct {
	Weeks     map[int][]models.Course `json:"weeks,omitempty"`
	UpdatedAt map[string]string       `json:"updatedAt,omitempty"`
}

type redisEnvelope struct {
	Data      string `json:"data"`
	FetchedAt int64  `json:"fetchedAt"`
	ExpiresAt int64  `json:"expiresAt"`
}

var globalOfflineCacheService *OfflineCacheService

func InitOfflineCacheService() {
	globalOfflineCacheService = &OfflineCacheService{redis: GetRedisService()}
}

func GetOfflineCacheService() *OfflineCacheService {
	if globalOfflineCacheService == nil {
		InitOfflineCacheService()
	}
	return globalOfflineCacheService
}

func (s *OfflineCacheService) now() time.Time { return time.Now() }

func (s *OfflineCacheService) newsRedisKey(cacheKey string) string {
	return fmt.Sprintf("offline_cache:news:%s", cacheKey)
}

func (s *OfflineCacheService) currentTimeRedisKey() string {
	return "offline_cache:current_time"
}

func (s *OfflineCacheService) courseTimetableRedisKey(userID uint, semester string, week int) string {
	return fmt.Sprintf("offline_cache:course_timetable_week:%d:%s:%d", userID, semester, week)
}

func (s *OfflineCacheService) classScheduleRedisKey(classID, term string) string {
	return fmt.Sprintf("offline_cache:class_schedule:%s:%s", classID, term)
}

func (s *OfflineCacheService) personalScoreRedisKey(userID uint, term string) string {
	return fmt.Sprintf("offline_cache:personal_score:%d:%s", userID, term)
}

func (s *OfflineCacheService) GetNewsCache(cacheKey string) (CacheResult, bool, error) {
	if cacheKey == "" {
		return CacheResult{}, false, errors.New("cacheKey is empty")
	}

	if s.redis != nil && s.redis.IsEnabled() {
		if raw, ok := s.redis.Get(s.newsRedisKey(cacheKey)); ok {
			var env redisEnvelope
			if err := json.Unmarshal(raw, &env); err == nil {
				return CacheResult{
					Data:      []byte(env.Data),
					Meta:      CacheMeta{FetchedAt: time.Unix(env.FetchedAt, 0), ExpiresAt: time.Unix(env.ExpiresAt, 0)},
					FromCache: true,
				}, true, nil
			}
		}
	}

	var row models.NewsCache
	err := models.DB.Where("cache_key = ?", cacheKey).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return CacheResult{}, false, nil
		}
		return CacheResult{}, false, err
	}

	res := CacheResult{
		Data:      []byte(row.Data),
		Meta:      CacheMeta{FetchedAt: row.FetchedAt, ExpiresAt: row.ExpiresAt},
		FromCache: true,
	}

	if s.redis != nil && s.redis.IsEnabled() {
		ttl := time.Until(row.ExpiresAt)
		if ttl > 0 {
			envRaw, _ := json.Marshal(redisEnvelope{Data: string(res.Data), FetchedAt: row.FetchedAt.Unix(), ExpiresAt: row.ExpiresAt.Unix()})
			_ = s.redis.Set(s.newsRedisKey(cacheKey), envRaw, ttl)
		}
	}

	return res, true, nil
}

func (s *OfflineCacheService) PutNewsCache(cacheKey string, data []byte, ttl time.Duration) error {
	if cacheKey == "" {
		return errors.New("cacheKey is empty")
	}
	now := s.now()
	expiresAt := now.Add(ttl)
	row := models.NewsCache{
		CacheKey:  cacheKey,
		Data:      string(data),
		FetchedAt: now,
		ExpiresAt: expiresAt,
	}
	err := models.DB.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "cache_key"}},
			DoUpdates: clause.AssignmentColumns([]string{"data", "fetched_at", "expires_at", "updated_at"}),
		},
	).Create(&row).Error
	if err != nil {
		return err
	}

	if s.redis != nil && s.redis.IsEnabled() {
		envRaw, _ := json.Marshal(redisEnvelope{Data: string(data), FetchedAt: now.Unix(), ExpiresAt: now.Add(ttl).Unix()})
		_ = s.redis.Set(s.newsRedisKey(cacheKey), envRaw, ttl)
	}
	return nil
}

func (s *OfflineCacheService) GetOrRefreshNews(cacheKey string, ttl time.Duration, fetch func() ([]byte, error)) (CacheResult, error) {
	cached, found, err := s.GetNewsCache(cacheKey)
	if err != nil {
		return CacheResult{}, err
	}
	if found && !cached.Meta.ExpiresAt.IsZero() && s.now().Before(cached.Meta.ExpiresAt) {
		return cached, nil
	}

	fresh, fetchErr := fetch()
	if fetchErr == nil {
		if err := s.PutNewsCache(cacheKey, fresh, ttl); err != nil {
			log.Printf("WARN offline cache: put news_cache failed: %v", err)
		}
		return CacheResult{Data: fresh, Meta: CacheMeta{FetchedAt: s.now(), ExpiresAt: s.now().Add(ttl)}, FromCache: false}, nil
	}

	if found {
		cached.Stale = true
		log.Printf("WARN offline cache: news fetch failed, served stale cache_key=%s err=%v", cacheKey, fetchErr)
		return cached, nil
	}

	return CacheResult{}, fetchErr
}

func (s *OfflineCacheService) GetPersonalScoreBundle(userID uint, term string) (PersonalScoreBundle, CacheMeta, bool, error) {
	if term == "" {
		term = "__default__"
	}
	if s.redis != nil && s.redis.IsEnabled() {
		if raw, ok := s.redis.Get(s.personalScoreRedisKey(userID, term)); ok {
			var env redisEnvelope
			if err := json.Unmarshal(raw, &env); err == nil {
				var bundle PersonalScoreBundle
				_ = json.Unmarshal([]byte(env.Data), &bundle)
				return bundle, CacheMeta{FetchedAt: time.Unix(env.FetchedAt, 0), ExpiresAt: time.Unix(env.ExpiresAt, 0)}, true, nil
			}
		}
	}

	var row models.PersonalScoreCache
	err := models.DB.Where("user_id = ? AND term = ?", userID, term).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return PersonalScoreBundle{}, CacheMeta{}, false, nil
		}
		return PersonalScoreBundle{}, CacheMeta{}, false, err
	}

	var bundle PersonalScoreBundle
	_ = json.Unmarshal([]byte(row.Data), &bundle)
	meta := CacheMeta{FetchedAt: row.FetchedAt, ExpiresAt: row.ExpiresAt}

	if s.redis != nil && s.redis.IsEnabled() {
		ttl := time.Until(row.ExpiresAt)
		if ttl > 0 {
			envRaw, _ := json.Marshal(redisEnvelope{Data: row.Data, FetchedAt: row.FetchedAt.Unix(), ExpiresAt: row.ExpiresAt.Unix()})
			_ = s.redis.Set(s.personalScoreRedisKey(userID, term), envRaw, ttl)
		}
	}

	return bundle, meta, true, nil
}

func (s *OfflineCacheService) PutPersonalScoreBundle(userID uint, term string, bundle PersonalScoreBundle, ttl time.Duration) (CacheMeta, error) {
	if term == "" {
		term = "__default__"
	}
	now := s.now()
	expiresAt := now.Add(ttl)
	if bundle.UpdatedAt == nil {
		bundle.UpdatedAt = map[string]string{}
	}
	raw, err := json.Marshal(bundle)
	if err != nil {
		return CacheMeta{}, err
	}
	row := models.PersonalScoreCache{
		UserID:    userID,
		Term:      term,
		Data:      string(raw),
		FetchedAt: now,
		ExpiresAt: expiresAt,
	}
	err = models.DB.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "term"}},
			DoUpdates: clause.AssignmentColumns([]string{"data", "fetched_at", "expires_at", "updated_at"}),
		},
	).Create(&row).Error
	if err != nil {
		return CacheMeta{}, err
	}
	if s.redis != nil && s.redis.IsEnabled() {
		envRaw, _ := json.Marshal(redisEnvelope{Data: string(raw), FetchedAt: now.Unix(), ExpiresAt: now.Add(ttl).Unix()})
		_ = s.redis.Set(s.personalScoreRedisKey(userID, term), envRaw, ttl)
	}
	return CacheMeta{FetchedAt: now, ExpiresAt: now.Add(ttl)}, nil
}

func (s *OfflineCacheService) UpsertPersonalScorePart(userID uint, term string, part string, payload []byte, ttl time.Duration) (CacheMeta, error) {
	if part == "" {
		return CacheMeta{}, errors.New("part is empty")
	}

	return s.upsertPersonalScorePartTx(userID, term, part, payload, ttl)
}

func (s *OfflineCacheService) upsertPersonalScorePartTx(userID uint, term string, part string, payload []byte, ttl time.Duration) (CacheMeta, error) {
	if term == "" {
		term = "__default__"
	}
	now := s.now()
	expiresAt := now.Add(ttl)

	returnMeta := CacheMeta{}
	err := models.DB.Transaction(func(tx *gorm.DB) error {
		var row models.PersonalScoreCache
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ? AND term = ?", userID, term).First(&row).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		bundle := PersonalScoreBundle{}
		if err == nil {
			_ = json.Unmarshal([]byte(row.Data), &bundle)
		}
		if bundle.UpdatedAt == nil {
			bundle.UpdatedAt = map[string]string{}
		}

		switch part {
		case "scoreList":
			bundle.ScoreList = append([]byte(nil), payload...)
		case "semesterScore":
			bundle.SemesterScore = append([]byte(nil), payload...)
		case "evaluationConfigList":
			bundle.EvaluationConfigList = append([]byte(nil), payload...)
		default:
			return fmt.Errorf("unknown personal score part: %s", part)
		}

		bundle.UpdatedAt[part] = now.Format("2006-01-02 15:04:05")
		raw, err := json.Marshal(bundle)
		if err != nil {
			return err
		}

		row = models.PersonalScoreCache{
			UserID:    userID,
			Term:      term,
			Data:      string(raw),
			FetchedAt: now,
			ExpiresAt: expiresAt,
		}
		err = tx.Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "term"}},
				DoUpdates: clause.AssignmentColumns([]string{"data", "fetched_at", "expires_at", "updated_at"}),
			},
		).Create(&row).Error
		if err != nil {
			return err
		}
		returnMeta = CacheMeta{FetchedAt: now, ExpiresAt: expiresAt}
		return nil
	})
	if err != nil {
		return CacheMeta{}, err
	}
	if s.redis != nil && s.redis.IsEnabled() {
		bundle, meta, _, _ := s.GetPersonalScoreBundle(userID, term)
		raw, _ := json.Marshal(bundle)
		envRaw, _ := json.Marshal(redisEnvelope{Data: string(raw), FetchedAt: meta.FetchedAt.Unix(), ExpiresAt: meta.ExpiresAt.Unix()})
		_ = s.redis.Set(s.personalScoreRedisKey(userID, term), envRaw, ttl)
	}
	return returnMeta, nil
}

func (s *OfflineCacheService) GetClassScheduleBundle(classID, term string) (ClassScheduleBundle, CacheMeta, bool, error) {
	if classID == "" {
		return ClassScheduleBundle{}, CacheMeta{}, false, errors.New("classID is empty")
	}
	if term == "" {
		term = "__default__"
	}
	if s.redis != nil && s.redis.IsEnabled() {
		if raw, ok := s.redis.Get(s.classScheduleRedisKey(classID, term)); ok {
			var env redisEnvelope
			if err := json.Unmarshal(raw, &env); err == nil {
				var bundle ClassScheduleBundle
				_ = json.Unmarshal([]byte(env.Data), &bundle)
				return bundle, CacheMeta{FetchedAt: time.Unix(env.FetchedAt, 0), ExpiresAt: time.Unix(env.ExpiresAt, 0)}, true, nil
			}
		}
	}

	var row models.ClassScheduleCache
	err := models.DB.Where("class_id = ? AND term = ?", classID, term).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ClassScheduleBundle{}, CacheMeta{}, false, nil
		}
		return ClassScheduleBundle{}, CacheMeta{}, false, err
	}

	var bundle ClassScheduleBundle
	_ = json.Unmarshal([]byte(row.Data), &bundle)
	meta := CacheMeta{FetchedAt: row.FetchedAt, ExpiresAt: row.ExpiresAt}
	if s.redis != nil && s.redis.IsEnabled() {
		ttl := time.Until(row.ExpiresAt)
		if ttl > 0 {
			envRaw, _ := json.Marshal(redisEnvelope{Data: row.Data, FetchedAt: row.FetchedAt.Unix(), ExpiresAt: row.ExpiresAt.Unix()})
			_ = s.redis.Set(s.classScheduleRedisKey(classID, term), envRaw, ttl)
		}
	}
	return bundle, meta, true, nil
}

func (s *OfflineCacheService) PutClassScheduleWeek(classID, term string, week int, courses []models.Course) (CacheMeta, error) {
	if classID == "" {
		return CacheMeta{}, errors.New("classID is empty")
	}
	if term == "" {
		term = "__default__"
	}
	now := s.now()
	expiresAt := endOfDayRefreshWindowEnd(now)

	var returnMeta CacheMeta
	err := models.DB.Transaction(func(tx *gorm.DB) error {
		var row models.ClassScheduleCache
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("class_id = ? AND term = ?", classID, term).First(&row).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		bundle := ClassScheduleBundle{}
		if err == nil {
			_ = json.Unmarshal([]byte(row.Data), &bundle)
		}
		if bundle.Weeks == nil {
			bundle.Weeks = map[int][]models.Course{}
		}
		if bundle.UpdatedAt == nil {
			bundle.UpdatedAt = map[string]string{}
		}
		bundle.Weeks[week] = courses
		bundle.UpdatedAt[fmt.Sprintf("week_%d", week)] = now.Format("2006-01-02 15:04:05")

		raw, err := json.Marshal(bundle)
		if err != nil {
			return err
		}
		row = models.ClassScheduleCache{
			ClassID:   classID,
			Term:      term,
			Data:      string(raw),
			FetchedAt: now,
			ExpiresAt: expiresAt,
		}
		err = tx.Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "class_id"}, {Name: "term"}},
				DoUpdates: clause.AssignmentColumns([]string{"data", "fetched_at", "expires_at", "updated_at"}),
			},
		).Create(&row).Error
		if err != nil {
			return err
		}
		returnMeta = CacheMeta{FetchedAt: now, ExpiresAt: expiresAt}
		return nil
	})
	if err != nil {
		return CacheMeta{}, err
	}
	if s.redis != nil && s.redis.IsEnabled() {
		bundle, _, _, _ := s.GetClassScheduleBundle(classID, term)
		raw, _ := json.Marshal(bundle)
		ttl := time.Until(returnMeta.ExpiresAt)
		if ttl > 0 {
			envRaw, _ := json.Marshal(redisEnvelope{Data: string(raw), FetchedAt: returnMeta.FetchedAt.Unix(), ExpiresAt: returnMeta.ExpiresAt.Unix()})
			_ = s.redis.Set(s.classScheduleRedisKey(classID, term), envRaw, ttl)
		}
	}
	return returnMeta, nil
}

func endOfDayRefreshWindowEnd(t time.Time) time.Time {
	y, m, d := t.Date()
	loc := t.Location()
	end := time.Date(y, m, d, 23, 30, 0, 0, loc)
	if t.After(end) {
		end = end.Add(24 * time.Hour)
	}
	return end
}

func IsWithinDailyRefreshWindow(t time.Time, startHour, startMin, endHour, endMin int) bool {
	y, m, d := t.Date()
	loc := t.Location()
	start := time.Date(y, m, d, startHour, startMin, 0, 0, loc)
	end := time.Date(y, m, d, endHour, endMin, 0, 0, loc)
	if end.Before(start) {
		end = end.Add(24 * time.Hour)
		if t.Before(start) {
			t = t.Add(24 * time.Hour)
		}
	}
	return (t.Equal(start) || t.After(start)) && t.Before(end)
}

// GetCurrentTimeCache 获取当前时间缓存（Redis优先，再DB）
func (s *OfflineCacheService) GetCurrentTimeCache() (CacheResult, bool, error) {
	const cacheKey = "global_current_time"

	if s.redis != nil && s.redis.IsEnabled() {
		if raw, ok := s.redis.Get(s.currentTimeRedisKey()); ok {
			var env redisEnvelope
			if err := json.Unmarshal(raw, &env); err == nil {
				return CacheResult{
					Data:      []byte(env.Data),
					Meta:      CacheMeta{FetchedAt: time.Unix(env.FetchedAt, 0), ExpiresAt: time.Unix(env.ExpiresAt, 0)},
					FromCache: true,
				}, true, nil
			}
		}
	}

	var row models.CurrentTimeCache
	err := models.DB.Where("cache_key = ?", cacheKey).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return CacheResult{}, false, nil
		}
		return CacheResult{}, false, err
	}

	res := CacheResult{
		Data:      []byte(row.Data),
		Meta:      CacheMeta{FetchedAt: row.FetchedAt, ExpiresAt: row.ExpiresAt},
		FromCache: true,
	}

	if s.redis != nil && s.redis.IsEnabled() {
		ttl := time.Until(row.ExpiresAt)
		if ttl > 0 {
			envRaw, _ := json.Marshal(redisEnvelope{Data: string(res.Data), FetchedAt: row.FetchedAt.Unix(), ExpiresAt: row.ExpiresAt.Unix()})
			_ = s.redis.Set(s.currentTimeRedisKey(), envRaw, ttl)
		}
	}

	return res, true, nil
}

// PutCurrentTimeCache 写入当前时间缓存
func (s *OfflineCacheService) PutCurrentTimeCache(data []byte, ttl time.Duration) error {
	const cacheKey = "global_current_time"
	now := s.now()
	expiresAt := now.Add(ttl)
	row := models.CurrentTimeCache{
		CacheKey:  cacheKey,
		Data:      string(data),
		FetchedAt: now,
		ExpiresAt: expiresAt,
	}
	err := models.DB.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "cache_key"}},
			DoUpdates: clause.AssignmentColumns([]string{"data", "fetched_at", "expires_at", "updated_at"}),
		},
	).Create(&row).Error
	if err != nil {
		return err
	}

	if s.redis != nil && s.redis.IsEnabled() {
		envRaw, _ := json.Marshal(redisEnvelope{Data: string(data), FetchedAt: now.Unix(), ExpiresAt: now.Add(ttl).Unix()})
		_ = s.redis.Set(s.currentTimeRedisKey(), envRaw, ttl)
	}
	return nil
}

// GetCourseTimetableWeekCache 获取课表周缓存（Redis优先，再DB）
func (s *OfflineCacheService) GetCourseTimetableWeekCache(userID uint, semester string, week int) (CacheResult, bool, error) {
	if semester == "" {
		semester = "__default__"
	}

	redisKey := s.courseTimetableRedisKey(userID, semester, week)
	if s.redis != nil && s.redis.IsEnabled() {
		if raw, ok := s.redis.Get(redisKey); ok {
			var env redisEnvelope
			if err := json.Unmarshal(raw, &env); err == nil {
				return CacheResult{
					Data:      []byte(env.Data),
					Meta:      CacheMeta{FetchedAt: time.Unix(env.FetchedAt, 0), ExpiresAt: time.Unix(env.ExpiresAt, 0)},
					FromCache: true,
				}, true, nil
			}
		}
	}

	var row models.CourseTimetableWeekCache
	err := models.DB.Where("user_id = ? AND semester = ? AND week = ?", userID, semester, week).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return CacheResult{}, false, nil
		}
		return CacheResult{}, false, err
	}

	res := CacheResult{
		Data:      []byte(row.Data),
		Meta:      CacheMeta{FetchedAt: row.FetchedAt, ExpiresAt: row.ExpiresAt},
		FromCache: true,
	}

	if s.redis != nil && s.redis.IsEnabled() {
		ttl := time.Until(row.ExpiresAt)
		if ttl > 0 {
			envRaw, _ := json.Marshal(redisEnvelope{Data: string(res.Data), FetchedAt: row.FetchedAt.Unix(), ExpiresAt: row.ExpiresAt.Unix()})
			_ = s.redis.Set(redisKey, envRaw, ttl)
		}
	}

	return res, true, nil
}

// PutCourseTimetableWeekCache 写入课表周缓存
func (s *OfflineCacheService) PutCourseTimetableWeekCache(userID uint, semester string, week int, data []byte, ttl time.Duration) error {
	if semester == "" {
		semester = "__default__"
	}
	now := s.now()
	expiresAt := now.Add(ttl)
	row := models.CourseTimetableWeekCache{
		UserID:    userID,
		Semester:  semester,
		Week:      week,
		Data:      string(data),
		FetchedAt: now,
		ExpiresAt: expiresAt,
	}
	err := models.DB.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "semester"}, {Name: "week"}},
			DoUpdates: clause.AssignmentColumns([]string{"data", "fetched_at", "expires_at", "updated_at"}),
		},
	).Create(&row).Error
	if err != nil {
		return err
	}

	if s.redis != nil && s.redis.IsEnabled() {
		envRaw, _ := json.Marshal(redisEnvelope{Data: string(data), FetchedAt: now.Unix(), ExpiresAt: now.Add(ttl).Unix()})
		_ = s.redis.Set(s.courseTimetableRedisKey(userID, semester, week), envRaw, ttl)
	}
	return nil
}
