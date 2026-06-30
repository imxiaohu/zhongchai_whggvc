-- 添加个人基础信息缓存相关字段
-- 1. sync_settings 表：个人基础信息缓存开关和状态
ALTER TABLE sync_settings
ADD COLUMN IF NOT EXISTS personal_info_sync_enabled TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否开启个人基础信息缓存',
ADD COLUMN IF NOT EXISTS personal_info_cache_status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '缓存状态: active活跃 paused暂停 resumed恢复',
ADD COLUMN IF NOT EXISTS personal_info_last_cached_at DATETIME(3) NULL COMMENT '上次个人基础信息缓存时间';

-- 2. users 表：活跃度追踪字段
ALTER TABLE users
ADD COLUMN IF NOT EXISTS last_active_at DATETIME(3) NULL COMMENT '用户最后活跃时间';

-- 3. 添加索引加速活跃度查询
ALTER TABLE users ADD INDEX IF NOT EXISTS idx_users_last_active_at (last_active_at);

-- 4. 添加 sync_logs 表的分类字段（可选，便于日志分类查询）
-- ALTER TABLE sync_logs ADD COLUMN IF NOT EXISTS sync_category VARCHAR(20) NOT NULL DEFAULT 'course' COMMENT '同步分类: course课程 personal_info个人信息';
