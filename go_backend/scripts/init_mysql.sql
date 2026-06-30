-- 平教小程序MySQL数据库初始化脚本
-- 使用方法：mysql -u root -p < init_mysql.sql

-- 创建数据库
CREATE DATABASE IF NOT EXISTS pingjiao CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE pingjiao;

-- 创建用户（可选）
-- CREATE USER IF NOT EXISTS 'pingjiao'@'localhost' IDENTIFIED BY 'your_password';
-- GRANT ALL PRIVILEGES ON pingjiao.* TO 'pingjiao'@'localhost';
-- FLUSH PRIVILEGES;

-- 注意：表结构会由GORM自动创建，这里只是创建数据库
-- 如果需要手动创建表，可以参考以下结构：

/*
-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    username VARCHAR(191) UNIQUE NOT NULL,
    nickname VARCHAR(191) NOT NULL,
    avatar_url TEXT,
    openid VARCHAR(191) UNIQUE,
    unionid VARCHAR(191),
    phone VARCHAR(191),
    email VARCHAR(191),
    real_name VARCHAR(191),
    student_id VARCHAR(191),
    class_name VARCHAR(191),
    school_id BIGINT UNSIGNED,
    status TINYINT DEFAULT 1,
    has_school_account BOOLEAN DEFAULT FALSE,
    school_username VARCHAR(191),
    school_password VARCHAR(191),
    INDEX idx_users_deleted_at (deleted_at),
    INDEX idx_users_username (username),
    INDEX idx_users_openid (openid)
);

-- 学校表
CREATE TABLE IF NOT EXISTS schools (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    name VARCHAR(191) NOT NULL,
    code VARCHAR(191) UNIQUE NOT NULL,
    description TEXT,
    logo_url TEXT,
    website VARCHAR(191),
    address TEXT,
    status TINYINT DEFAULT 1,
    INDEX idx_schools_deleted_at (deleted_at),
    INDEX idx_schools_code (code)
);

-- 新闻表
CREATE TABLE IF NOT EXISTS news (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    title VARCHAR(191) NOT NULL,
    content LONGTEXT,
    summary TEXT,
    author VARCHAR(191),
    type_id BIGINT UNSIGNED,
    school_id BIGINT UNSIGNED,
    view_count BIGINT UNSIGNED DEFAULT 0,
    attachment_url TEXT,
    attachment_name VARCHAR(191),
    publish_time DATETIME(3),
    status TINYINT DEFAULT 1,
    INDEX idx_news_deleted_at (deleted_at),
    INDEX idx_news_type_id (type_id),
    INDEX idx_news_school_id (school_id)
);

-- 学期表
CREATE TABLE IF NOT EXISTS semesters (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    name VARCHAR(191) NOT NULL,
    start_date DATE,
    end_date DATE,
    is_current BOOLEAN DEFAULT FALSE,
    school_id BIGINT UNSIGNED,
    INDEX idx_semesters_deleted_at (deleted_at),
    INDEX idx_semesters_school_id (school_id)
);

-- 课程表
CREATE TABLE IF NOT EXISTS courses (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    name VARCHAR(191) NOT NULL,
    teacher VARCHAR(191),
    classroom VARCHAR(191),
    weekday TINYINT,
    start_time VARCHAR(191),
    end_time VARCHAR(191),
    start_week TINYINT,
    end_week TINYINT,
    user_id BIGINT UNSIGNED,
    semester_id BIGINT UNSIGNED,
    course_code VARCHAR(191),
    credits DECIMAL(3,1),
    INDEX idx_courses_deleted_at (deleted_at),
    INDEX idx_courses_user_id (user_id),
    INDEX idx_courses_semester_id (semester_id)
);

-- 评教表
CREATE TABLE IF NOT EXISTS evaluations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    title VARCHAR(191) NOT NULL,
    description TEXT,
    start_time DATETIME(3),
    end_time DATETIME(3),
    semester_id BIGINT UNSIGNED,
    status TINYINT DEFAULT 1,
    INDEX idx_evaluations_deleted_at (deleted_at),
    INDEX idx_evaluations_semester_id (semester_id)
);
*/

-- 显示创建结果
SHOW DATABASES LIKE 'pingjiao';
SELECT 'MySQL数据库初始化完成！' AS message;
