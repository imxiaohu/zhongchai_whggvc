-- 添加 PC 端会话凭证相关字段
-- 用于存储用户 PC 端登录后的 JSESSIONID，与移动端 Token 完全独立

ALTER TABLE users
ADD COLUMN IF NOT EXISTS pc_jsession_id VARCHAR(256) DEFAULT '' COMMENT 'PC端会话凭证JSESSIONID',
ADD COLUMN IF NOT EXISTS pc_login_time DATETIME(3) NULL COMMENT 'PC端登录时间',
ADD COLUMN IF NOT EXISTS pc_expire_time DATETIME(3) NULL COMMENT 'PC端会话过期时间',
ADD COLUMN IF NOT EXISTS pc_user_agent VARCHAR(512) DEFAULT '' COMMENT 'PC端登录时浏览器UA';

-- 添加索引加速会话查询
ALTER TABLE users ADD INDEX IF NOT EXISTS idx_users_pc_jsession_id (pc_jsession_id);
