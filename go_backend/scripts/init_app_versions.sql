-- 创建应用版本表
CREATE TABLE IF NOT EXISTS app_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    version_code INTEGER NOT NULL,
    platform VARCHAR(50) NOT NULL,
    download_url TEXT,
    size VARCHAR(50),
    release_notes TEXT,
    is_forced BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_app_versions_app_id ON app_versions(app_id);
CREATE INDEX IF NOT EXISTS idx_app_versions_platform ON app_versions(platform);
CREATE INDEX IF NOT EXISTS idx_app_versions_version_code ON app_versions(version_code);
CREATE INDEX IF NOT EXISTS idx_app_versions_is_active ON app_versions(is_active);

-- 插入示例数据
INSERT INTO app_versions (app_id, version, version_code, platform, download_url, size, release_notes, is_forced, is_active) VALUES
-- Android 版本
('__UNI__84DD641', '1.0.0', 100, 'Android', 'https://example.com/downloads/pingjiao-v1.0.0.apk', '15.2MB', '初始版本\n- 基础功能实现\n- 课程表查看\n- 成绩查询\n- 评教功能', FALSE, TRUE),
('__UNI__84DD641', '1.0.1', 101, 'Android', 'https://example.com/downloads/pingjiao-v1.0.1.apk', '15.5MB', '版本更新\n- 修复已知问题\n- 优化用户体验\n- 新增应用更新功能', FALSE, TRUE),
('__UNI__84DD641', '1.1.0', 110, 'Android', 'https://example.com/downloads/pingjiao-v1.1.0.apk', '16.8MB', '重大更新\n- 新增社区功能\n- 优化界面设计\n- 提升性能表现\n- 修复多个Bug', FALSE, TRUE),

-- iOS 版本
('__UNI__84DD641', '1.0.0', 100, 'iOS', 'https://apps.apple.com/app/pingjiao/id123456789', '18.3MB', '初始版本\n- 基础功能实现\n- 课程表查看\n- 成绩查询\n- 评教功能', FALSE, TRUE),
('__UNI__84DD641', '1.0.1', 101, 'iOS', 'https://apps.apple.com/app/pingjiao/id123456789', '18.6MB', '版本更新\n- 修复已知问题\n- 优化用户体验\n- 新增应用更新功能', FALSE, TRUE),

-- H5 版本
('__UNI__84DD641', '1.0.0', 100, 'H5', 'https://app.example.com/', '0MB', '初始版本\n- 基础功能实现\n- 课程表查看\n- 成绩查询\n- 评教功能', FALSE, TRUE),
('__UNI__84DD641', '1.0.1', 101, 'H5', 'https://app.example.com/', '0MB', '版本更新\n- 修复已知问题\n- 优化用户体验\n- 新增应用更新功能', FALSE, TRUE),

-- 微信小程序版本
('YOUR_WX_APPID2_REMOVED', '1.0.0', 100, 'WeChat', '', '2.5MB', '初始版本\n- 基础功能实现\n- 课程表查看\n- 成绩查询\n- 评教功能', FALSE, TRUE),
('YOUR_WX_APPID2_REMOVED', '1.0.1', 101, 'WeChat', '', '2.6MB', '版本更新\n- 修复已知问题\n- 优化用户体验\n- 新增应用更新功能', FALSE, TRUE);

-- 更新时间戳触发器（SQLite）
CREATE TRIGGER IF NOT EXISTS update_app_versions_updated_at 
    AFTER UPDATE ON app_versions
    FOR EACH ROW
BEGIN
    UPDATE app_versions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
