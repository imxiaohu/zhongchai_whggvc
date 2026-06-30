-- 同步设置表复合索引优化
-- 解决 GetSyncSettingsForSchedule 查询的全表扫描问题
-- 支持 (enabled, next_sync_at, sync_status) 三字段组合查询

-- 如果索引已存在则跳过
SET @exist := (SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
    AND table_name = 'sync_settings'
    AND index_name = 'idx_sync_enabled_schedule');
SET @sqlstmt := IF(@exist > 0,
    'SELECT "Index already exists"',
    'ALTER TABLE sync_settings ADD INDEX idx_sync_enabled_schedule (enabled, next_sync_at, sync_status)');
PREPARE stmt FROM @sqlstmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 单独为 enabled 字段添加索引（加速 enabled=true 的过滤）
SET @exist_enabled := (SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
    AND table_name = 'sync_settings'
    AND index_name = 'idx_sync_settings_enabled');
SET @sqlstmt2 := IF(@exist_enabled > 0,
    'SELECT "idx_sync_settings_enabled already exists"',
    'ALTER TABLE sync_settings ADD INDEX idx_sync_settings_enabled (enabled)');
PREPARE stmt2 FROM @sqlstmt2;
EXECUTE stmt2;
DEALLOCATE PREPARE stmt2;
