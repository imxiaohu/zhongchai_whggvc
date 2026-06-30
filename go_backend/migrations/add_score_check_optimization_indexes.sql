-- 成绩检查系统性能优化索引
-- 1. notification_channels 表：加速调度器查询 "哪些用户启用了成绩检查"
-- 2. score_snapshots 表：确保 upsert 操作的唯一性约束

-- ============================================================
-- 1. notification_channels 表：联合索引加速调度器全量扫描
-- 调度器查询: SELECT * FROM notification_channels WHERE score_check_enabled = true
-- ============================================================

SET @exist_nc := (SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
    AND table_name = 'notification_channels'
    AND index_name = 'idx_score_check_enabled');
IF @exist_nc = 0 THEN
    ALTER TABLE notification_channels
    ADD INDEX idx_score_check_enabled (score_check_enabled, last_score_check);
END IF;

-- ============================================================
-- 2. score_snapshots 表：联合唯一索引
-- 防止同一用户、学期、课程、成绩类型出现多条 current 版本记录
-- 这也是 INSERT ON DUPLICATE KEY UPDATE 的前提
-- ============================================================

SET @exist_ss := (SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
    AND table_name = 'score_snapshots'
    AND index_name = 'idx_score_snapshot_unique_current');
IF @exist_ss = 0 THEN
    -- 先检查是否有重复数据（如果有，先清理）
    -- 然后添加唯一索引
    ALTER TABLE score_snapshots
    ADD UNIQUE INDEX idx_score_snapshot_unique_current
        (user_id, semester, course_code, score_type, version, is_active);
END IF;

-- ============================================================
-- 3. score_snapshots 表：为 check_sum 字段添加复合索引
-- 用于快速对比同一课程的成绩是否变化
-- ============================================================

SET @exist_cs := (SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
    AND table_name = 'score_snapshots'
    AND index_name = 'idx_score_snapshot_checksum');
IF @exist_cs = 0 THEN
    ALTER TABLE score_snapshots
    ADD INDEX idx_score_snapshot_checksum (user_id, semester, course_code, score_type, check_sum);
END IF;

-- ============================================================
-- 4. 验证索引创建结果
-- ============================================================
-- SHOW INDEX FROM notification_channels;
-- SHOW INDEX FROM score_snapshots;
