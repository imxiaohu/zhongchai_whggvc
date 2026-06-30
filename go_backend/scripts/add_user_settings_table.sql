-- 用户设置表创建脚本
-- 用于手动创建用户设置表（如果自动迁移失败）

-- 创建用户设置表
CREATE TABLE IF NOT EXISTS user_settings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    
    -- 用户关联
    user_id BIGINT UNSIGNED NOT NULL,
    
    -- 基本设置
    client_type VARCHAR(20) DEFAULT 'web' COMMENT '客户端类型: web, mobile, wechat, app',
    language VARCHAR(10) DEFAULT 'zh-CN' COMMENT '语言设置: zh-CN, en-US',
    theme VARCHAR(20) DEFAULT 'auto' COMMENT '主题设置: light, dark, auto',
    nickname VARCHAR(50) DEFAULT '' COMMENT '用户昵称',
    error_notification BOOLEAN DEFAULT TRUE COMMENT '错误提示管理：是否显示错误通知',
    error_notification_mode VARCHAR(20) DEFAULT 'popup' COMMENT '错误提示模式: popup, toast, silent',
    
    -- 服务器同步配置
    sync_enabled BOOLEAN DEFAULT FALSE COMMENT '是否启用服务器同步',
    sync_frequency VARCHAR(20) DEFAULT 'daily' COMMENT '同步频率: daily, weekly, every2days, every3days',
    sync_time_range VARCHAR(50) DEFAULT '08:30-22:20' COMMENT '同步时间范围',
    sync_auto_retry BOOLEAN DEFAULT TRUE COMMENT '是否启用自动重试',
    sync_notification BOOLEAN DEFAULT TRUE COMMENT '同步完成通知',
    
    -- 界面设置
    show_welcome_guide BOOLEAN DEFAULT TRUE COMMENT '是否显示欢迎引导',
    compact_mode BOOLEAN DEFAULT FALSE COMMENT '紧凑模式',
    show_avatar_in_header BOOLEAN DEFAULT TRUE COMMENT '头部显示头像',
    
    -- 隐私设置
    data_collection BOOLEAN DEFAULT TRUE COMMENT '数据收集同意',
    analytics_enabled BOOLEAN DEFAULT TRUE COMMENT '分析统计启用',
    
    -- 通知设置
    push_notification BOOLEAN DEFAULT TRUE COMMENT '推送通知',
    email_notification BOOLEAN DEFAULT FALSE COMMENT '邮件通知',
    news_notification BOOLEAN DEFAULT TRUE COMMENT '新闻通知',
    score_notification BOOLEAN DEFAULT TRUE COMMENT '成绩通知',
    schedule_notification BOOLEAN DEFAULT TRUE COMMENT '课程表通知',
    
    -- 扩展字段
    custom_settings TEXT COMMENT '自定义设置JSON字符串',
    last_modified_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '最后修改时间',
    
    -- 索引
    UNIQUE KEY idx_user_settings_user_id (user_id),
    KEY idx_user_settings_deleted_at (deleted_at),
    KEY idx_user_settings_last_modified (last_modified_at),
    
    -- 外键约束
    CONSTRAINT fk_user_settings_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户设置表';

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_user_settings_client_type ON user_settings(client_type);
CREATE INDEX IF NOT EXISTS idx_user_settings_language ON user_settings(language);
CREATE INDEX IF NOT EXISTS idx_user_settings_theme ON user_settings(theme);
CREATE INDEX IF NOT EXISTS idx_user_settings_sync_enabled ON user_settings(sync_enabled);

-- 插入说明注释
INSERT INTO information_schema.table_comment 
VALUES ('zhongchai_go', 'user_settings', '用户设置表 - 存储用户的个性化设置信息，包括界面、同步、通知等配置')
ON DUPLICATE KEY UPDATE table_comment = VALUES(table_comment);
