-- 添加评论层级相关字段的数据库迁移脚本
-- 执行时间：2024-12-19

-- 1. 添加新字段到comments表
ALTER TABLE comments 
ADD COLUMN root_id INT UNSIGNED NULL COMMENT '根评论ID，用于快速查找评论树',
ADD COLUMN level INT NOT NULL DEFAULT 0 COMMENT '评论层级，0为顶级评论',
ADD COLUMN path VARCHAR(500) NOT NULL DEFAULT '' COMMENT '评论路径，如"1/2/3"，用于排序和查询',
ADD COLUMN replies_count INT NOT NULL DEFAULT 0 COMMENT '回复数量',
ADD COLUMN mentioned_users TEXT NULL COMMENT '@提及的用户ID列表，JSON格式',
ADD COLUMN is_hot BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否为热门评论';

-- 2. 添加索引以优化查询性能
ALTER TABLE comments 
ADD INDEX idx_comments_post_id (post_id),
ADD INDEX idx_comments_user_id (user_id),
ADD INDEX idx_comments_parent_id (parent_id),
ADD INDEX idx_comments_root_id (root_id),
ADD INDEX idx_comments_level (level),
ADD INDEX idx_comments_path (path),
ADD INDEX idx_comments_is_hot (is_hot),
ADD INDEX idx_comments_status (status);

-- 3. 添加外键约束（如果不存在）
-- 注意：在生产环境中，请根据实际情况决定是否添加外键约束
-- ALTER TABLE comments 
-- ADD CONSTRAINT fk_comments_root_id 
-- FOREIGN KEY (root_id) REFERENCES comments(id) ON DELETE CASCADE;

-- 4. 更新现有数据
-- 为现有的顶级评论设置path
UPDATE comments 
SET path = CAST(id AS CHAR), level = 0 
WHERE parent_id IS NULL AND path = '';

-- 为现有的回复评论设置层级信息
-- 这是一个递归更新，可能需要多次执行直到所有数据都更新完成
-- 第一层回复
UPDATE comments c1
JOIN comments c2 ON c1.parent_id = c2.id
SET c1.level = 1,
    c1.root_id = CASE WHEN c2.parent_id IS NULL THEN c2.id ELSE c2.root_id END,
    c1.path = CONCAT(CASE WHEN c2.path = '' THEN CAST(c2.id AS CHAR) ELSE c2.path END, '/', CAST(c1.id AS CHAR))
WHERE c1.parent_id IS NOT NULL AND c1.level = 0;

-- 第二层回复
UPDATE comments c1
JOIN comments c2 ON c1.parent_id = c2.id
SET c1.level = c2.level + 1,
    c1.root_id = c2.root_id,
    c1.path = CONCAT(c2.path, '/', CAST(c1.id AS CHAR))
WHERE c1.parent_id IS NOT NULL AND c1.level <= c2.level;

-- 第三层回复
UPDATE comments c1
JOIN comments c2 ON c1.parent_id = c2.id
SET c1.level = c2.level + 1,
    c1.root_id = c2.root_id,
    c1.path = CONCAT(c2.path, '/', CAST(c1.id AS CHAR))
WHERE c1.parent_id IS NOT NULL AND c1.level <= c2.level;

-- 第四层回复
UPDATE comments c1
JOIN comments c2 ON c1.parent_id = c2.id
SET c1.level = c2.level + 1,
    c1.root_id = c2.root_id,
    c1.path = CONCAT(c2.path, '/', CAST(c1.id AS CHAR))
WHERE c1.parent_id IS NOT NULL AND c1.level <= c2.level;

-- 第五层回复
UPDATE comments c1
JOIN comments c2 ON c1.parent_id = c2.id
SET c1.level = c2.level + 1,
    c1.root_id = c2.root_id,
    c1.path = CONCAT(c2.path, '/', CAST(c1.id AS CHAR))
WHERE c1.parent_id IS NOT NULL AND c1.level <= c2.level;

-- 5. 更新回复数量统计
UPDATE comments c1
SET replies_count = (
    SELECT COUNT(*)
    FROM comments c2
    WHERE c2.parent_id = c1.id AND c2.status = 1
)
WHERE c1.status = 1;

-- 6. 验证数据完整性
-- 检查是否有未设置path的评论
SELECT COUNT(*) as unset_path_count 
FROM comments 
WHERE path = '' AND status = 1;

-- 检查是否有层级设置错误的评论
SELECT COUNT(*) as invalid_level_count
FROM comments c1
JOIN comments c2 ON c1.parent_id = c2.id
WHERE c1.level != c2.level + 1 AND c1.status = 1;

-- 7. 创建触发器以自动维护回复数量（可选）
DELIMITER $$

CREATE TRIGGER update_replies_count_after_insert
AFTER INSERT ON comments
FOR EACH ROW
BEGIN
    IF NEW.parent_id IS NOT NULL THEN
        UPDATE comments 
        SET replies_count = replies_count + 1 
        WHERE id = NEW.parent_id;
    END IF;
END$$

CREATE TRIGGER update_replies_count_after_update
AFTER UPDATE ON comments
FOR EACH ROW
BEGIN
    -- 如果状态从正常变为删除
    IF OLD.status = 1 AND NEW.status = 0 AND NEW.parent_id IS NOT NULL THEN
        UPDATE comments 
        SET replies_count = replies_count - 1 
        WHERE id = NEW.parent_id;
    END IF;
    
    -- 如果状态从删除变为正常
    IF OLD.status = 0 AND NEW.status = 1 AND NEW.parent_id IS NOT NULL THEN
        UPDATE comments 
        SET replies_count = replies_count + 1 
        WHERE id = NEW.parent_id;
    END IF;
END$$

CREATE TRIGGER update_replies_count_after_delete
AFTER DELETE ON comments
FOR EACH ROW
BEGIN
    IF OLD.parent_id IS NOT NULL THEN
        UPDATE comments 
        SET replies_count = replies_count - 1 
        WHERE id = OLD.parent_id;
    END IF;
END$$

DELIMITER ;

-- 8. 创建用于评论统计的视图（可选）
CREATE VIEW comment_stats AS
SELECT 
    post_id,
    COUNT(*) as total_comments,
    COUNT(CASE WHEN parent_id IS NULL THEN 1 END) as root_comments,
    COUNT(CASE WHEN parent_id IS NOT NULL THEN 1 END) as reply_comments,
    MAX(level) as max_level,
    AVG(level) as avg_level
FROM comments 
WHERE status = 1
GROUP BY post_id;

-- 迁移完成
-- 请在执行此脚本后验证数据的完整性和正确性
-- 建议在生产环境执行前先在测试环境进行充分测试
