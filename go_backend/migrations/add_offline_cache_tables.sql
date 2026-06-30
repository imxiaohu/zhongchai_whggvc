CREATE TABLE IF NOT EXISTS class_schedule_cache (
  class_id VARCHAR(64) NOT NULL,
  term VARCHAR(64) NOT NULL,
  data LONGTEXT,
  fetched_at DATETIME,
  expires_at DATETIME,
  created_at DATETIME,
  updated_at DATETIME,
  PRIMARY KEY (class_id, term),
  INDEX idx_class_schedule_expires (expires_at),
  INDEX idx_class_schedule_fetched (fetched_at)
);

CREATE TABLE IF NOT EXISTS news_cache (
  cache_key VARCHAR(32) NOT NULL,
  data LONGTEXT,
  fetched_at DATETIME,
  expires_at DATETIME,
  created_at DATETIME,
  updated_at DATETIME,
  PRIMARY KEY (cache_key),
  INDEX idx_news_expires (expires_at),
  INDEX idx_news_fetched (fetched_at)
);

CREATE TABLE IF NOT EXISTS personal_score_cache (
  user_id BIGINT UNSIGNED NOT NULL,
  term VARCHAR(64) NOT NULL,
  data LONGTEXT,
  fetched_at DATETIME,
  expires_at DATETIME,
  created_at DATETIME,
  updated_at DATETIME,
  PRIMARY KEY (user_id, term),
  INDEX idx_personal_score_expires (expires_at),
  INDEX idx_personal_score_fetched (fetched_at)
);
