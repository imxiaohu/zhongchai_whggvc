-- 添加成绩快照版本控制字段
-- 用于支持成绩检测系统的版本对比功能

-- 1. 添加版本控制字段到 score_snapshots 表
ALTER TABLE score_snapshots 
ADD COLUMN version VARCHAR(20) NOT NULL DEFAULT 'current' COMMENT '版本标识: current, previous';

ALTER TABLE score_snapshots 
ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否为活跃版本';

-- 2. 添加索引以提高查询性能
CREATE INDEX idx_score_snapshots_version ON score_snapshots(user_id, semester, version, is_active);
CREATE INDEX idx_score_snapshots_user_version ON score_snapshots(user_id, version, is_active);

-- 3. 更新现有数据，将所有现有快照标记为 current 版本
UPDATE score_snapshots 
SET version = 'current', is_active = TRUE 
WHERE version IS NULL OR version = '';

-- 4. 创建版本轮换的存储过程（MySQL）
DELIMITER //

CREATE PROCEDURE RotateScoreSnapshots(
    IN p_user_id INT UNSIGNED,
    IN p_semester VARCHAR(50)
)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    START TRANSACTION;
    
    -- 将所有 previous 版本标记为非活跃
    UPDATE score_snapshots 
    SET is_active = FALSE 
    WHERE user_id = p_user_id 
      AND semester = p_semester 
      AND version = 'previous';
    
    -- 将 current 版本改为 previous 版本
    UPDATE score_snapshots 
    SET version = 'previous' 
    WHERE user_id = p_user_id 
      AND semester = p_semester 
      AND version = 'current';
    
    COMMIT;
END //

DELIMITER ;

-- 5. 创建清理过期快照的存储过程
DELIMITER //

CREATE PROCEDURE CleanupExpiredSnapshots(
    IN p_days_to_keep INT DEFAULT 30
)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    START TRANSACTION;
    
    -- 删除超过指定天数的非活跃快照
    DELETE FROM score_snapshots 
    WHERE is_active = FALSE 
      AND created_at < DATE_SUB(NOW(), INTERVAL p_days_to_keep DAY);
    
    COMMIT;
END //

DELIMITER ;

-- 6. 创建获取用户成绩快照统计的视图
CREATE VIEW v_score_snapshot_stats AS
SELECT 
    user_id,
    semester,
    version,
    COUNT(*) as snapshot_count,
    MAX(created_at) as latest_created_at,
    MAX(updated_at) as latest_updated_at
FROM score_snapshots 
WHERE is_active = TRUE
GROUP BY user_id, semester, version;

-- 7. 插入一些示例数据用于测试（可选）
-- INSERT INTO score_snapshots (user_id, semester, course_code, course_name, score_type, score, credit, gpa, version, is_active, check_sum) VALUES
-- (10, '2024-2025学年第二学期', 'MATH101', '高等数学', 'final', '85', 4.0, 3.5, 'previous', TRUE, MD5('test_checksum_1')),
-- (10, '2024-2025学年第二学期', 'MATH101', '高等数学', 'final', '88', 4.0, 3.8, 'current', TRUE, MD5('test_checksum_2'));

-- 8. 验证数据完整性的查询
-- 检查是否有重复的 current 版本
-- SELECT user_id, semester, course_code, score_type, COUNT(*) as count
-- FROM score_snapshots 
-- WHERE version = 'current' AND is_active = TRUE
-- GROUP BY user_id, semester, course_code, score_type
-- HAVING COUNT(*) > 1;

-- 9. 创建定期清理任务的事件（MySQL Event Scheduler）
-- 注意：需要确保 event_scheduler 已启用
-- SET GLOBAL event_scheduler = ON;

CREATE EVENT IF NOT EXISTS cleanup_expired_snapshots
ON SCHEDULE EVERY 1 DAY
STARTS CURRENT_TIMESTAMP
DO
  CALL CleanupExpiredSnapshots(30);

-- 10. 添加注释说明
ALTER TABLE score_snapshots COMMENT = '成绩快照表 - 支持版本控制的成绩检测系统';

-- 执行完成后的验证查询
-- SELECT 
--     COLUMN_NAME, 
--     DATA_TYPE, 
--     IS_NULLABLE, 
--     COLUMN_DEFAULT, 
--     COLUMN_COMMENT 
-- FROM INFORMATION_SCHEMA.COLUMNS 
-- WHERE TABLE_NAME = 'score_snapshots' 
--   AND COLUMN_NAME IN ('version', 'is_active')
-- ORDER BY ORDINAL_POSITION;
