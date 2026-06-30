-- 添加社区互动通知偏好字段
ALTER TABLE notification_channels 
ADD COLUMN community_like_email BOOLEAN DEFAULT FALSE COMMENT '点赞邮件通知',
ADD COLUMN community_like_ding_talk BOOLEAN DEFAULT FALSE COMMENT '点赞钉钉通知',
ADD COLUMN community_bookmark_email BOOLEAN DEFAULT FALSE COMMENT '收藏邮件通知',
ADD COLUMN community_bookmark_ding_talk BOOLEAN DEFAULT FALSE COMMENT '收藏钉钉通知',
ADD COLUMN community_comment_email BOOLEAN DEFAULT FALSE COMMENT '评论邮件通知',
ADD COLUMN community_comment_ding_talk BOOLEAN DEFAULT FALSE COMMENT '评论钉钉通知';
