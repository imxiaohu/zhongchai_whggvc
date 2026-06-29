/**
 * 社区相关API
 */

import { request } from '../../utils/request.js';

// 基础URL
const BASE_URL = '/api';

/**
 * 社团相关API
 */

// 获取社团列表
export function getClubsList(params = {}) {
  return request({
    url: `${BASE_URL}/clubs`,
    method: 'GET',
    data: {
      page: 1,
      pageSize: 10,
      ...params
    }
  });
}

// 获取我的社团列表
export function getMyClubs() {
  return request({
    url: `${BASE_URL}/clubs/my`,
    method: 'GET'
  });
}

// 获取社团详情
export function getClubDetail(clubId) {
  return request({
    url: `${BASE_URL}/clubs/${clubId}`,
    method: 'GET'
  });
}

// 创建社团
export function createClub(data) {
  return request({
    url: `${BASE_URL}/clubs`,
    method: 'POST',
    data
  });
}

// 更新社团信息
export function updateClub(clubId, data) {
  return request({
    url: `${BASE_URL}/clubs/${clubId}`,
    method: 'PUT',
    data
  });
}

// 删除社团
export function deleteClub(clubId) {
  return request({
    url: `${BASE_URL}/clubs/${clubId}`,
    method: 'DELETE'
  });
}

// 加入社团
export function joinClub(clubId) {
  return request({
    url: `${BASE_URL}/clubs/${clubId}/join`,
    method: 'POST'
  });
}

// 退出社团
export function leaveClub(clubId) {
  return request({
    url: `${BASE_URL}/clubs/${clubId}/leave`,
    method: 'POST'
  });
}

// 获取社团成员列表
export function getClubMembers(clubId, params = {}) {
  return request({
    url: `${BASE_URL}/clubs/${clubId}/members`,
    method: 'GET',
    data: {
      page: 1,
      pageSize: 20,
      ...params
    }
  });
}

// 更新成员角色
export function updateMemberRole(clubId, memberId, role) {
  return request({
    url: `${BASE_URL}/clubs/${clubId}/members/${memberId}/role`,
    method: 'PUT',
    data: { role }
  });
}

/**
 * 帖子相关API
 */

// 获取帖子列表
export function getPostsList(params = {}) {
  return request({
    url: `${BASE_URL}/posts`,
    method: 'GET',
    data: {
      page: 1,
      pageSize: 10,
      ...params
    }
  });
}

// 获取帖子详情
export function getPostDetail(postId) {
  return request({
    url: `${BASE_URL}/posts/${postId}`,
    method: 'GET'
  });
}

// 创建帖子
export function createPost(data) {
  return request({
    url: `${BASE_URL}/posts`,
    method: 'POST',
    data
  });
}

// 更新帖子
export function updatePost(postId, data) {
  return request({
    url: `${BASE_URL}/posts/${postId}`,
    method: 'PUT',
    data
  });
}

// 删除帖子
export function deletePost(postId) {
  return request({
    url: `${BASE_URL}/posts/${postId}`,
    method: 'DELETE'
  });
}

// 点赞帖子
export function likePost(postId) {
  return request({
    url: `${BASE_URL}/posts/${postId}/like`,
    method: 'POST'
  });
}

// 取消点赞帖子
export function unlikePost(postId) {
  return request({
    url: `${BASE_URL}/posts/${postId}/unlike`,
    method: 'POST'
  });
}

// 切换点赞状态 (别名)
export async function toggleLike(postId) {
  // 先获取详情判断状态，或者由后端处理。这里简单实现：
  // 实际建议后端提供单个 toggle 接口，或者前端维护状态。
  // 为了兼容现有代码，这里假设是通过 post 状态决定的，但 API 没传状态。
  // 我们可以尝试先点赞，如果报错说已点赞则取消点赞，或者反之。
  // 但最稳妥的是在组件内维护状态调用 likePost/unlikePost。
  // 鉴于 club-detail.vue 等处直接用了 toggleLike，我们在这里做一个简单的逻辑：
  // 尝试调用 likePost，如果成功返回新状态。
  try {
    const res = await likePost(postId);
    if (res && res.success) {
      const likesCount = res.result?.likesCount ?? res.result?.post?.likesCount;
      const isLiked = res.result?.isLiked ?? true;
      return { isLiked, likesCount };
    }
    return {
      isLiked: res?.result?.isLiked,
      likesCount: res?.result?.likesCount ?? res?.result?.post?.likesCount
    };
  } catch (err) {
    const res = await unlikePost(postId);
    const likesCount = res?.result?.likesCount ?? res?.result?.post?.likesCount;
    const isLiked = res?.result?.isLiked ?? false;
    return { isLiked, likesCount };
  }
}

// 切换收藏状态 (别名)
export async function toggleBookmark(postId) {
  try {
    const res = await bookmarkPost(postId);
    if (res && res.success) {
      const isBookmarked = res.result?.isBookmarked ?? true;
      const bookmarksCount = res.result?.bookmarksCount ?? res.result?.post?.bookmarksCount;
      return { isBookmarked, bookmarksCount };
    }
    return {
      isBookmarked: res?.result?.isBookmarked,
      bookmarksCount: res?.result?.bookmarksCount ?? res?.result?.post?.bookmarksCount
    };
  } catch (err) {
    const res = await unbookmarkPost(postId);
    const isBookmarked = res?.result?.isBookmarked ?? false;
    const bookmarksCount = res?.result?.bookmarksCount ?? res?.result?.post?.bookmarksCount;
    return { isBookmarked, bookmarksCount };
  }
}

/**
 * 评论相关API
 */

// 获取帖子列表 (别名)
export const getPosts = getPostsList;

// 获取评论列表（支持层级和排序）
export function getCommentsList(postId, params = {}) {
  return request({
    url: `${BASE_URL}/comments`,
    method: 'GET',
    data: {
      postId,
      page: 1,
      pageSize: 20,
      sortBy: 'time', // time, time_desc, likes, hot
      ...params
    }
  });
}

// 创建评论
export function createComment(data) {
  return request({
    url: `${BASE_URL}/comments`,
    method: 'POST',
    data
  });
}

// 更新评论
export function updateComment(commentId, data) {
  return request({
    url: `${BASE_URL}/comments/${commentId}`,
    method: 'PUT',
    data
  });
}

// 删除评论
export function deleteComment(commentId) {
  return request({
    url: `${BASE_URL}/comments/${commentId}`,
    method: 'DELETE'
  });
}

// 回复评论
export function replyToComment(commentId, data) {
  return request({
    url: `${BASE_URL}/comments/${commentId}/reply`,
    method: 'POST',
    data
  });
}

// 获取评论回复列表
export function getCommentReplies(commentId, params = {}) {
  return request({
    url: `${BASE_URL}/comments/${commentId}/replies`,
    method: 'GET',
    data: {
      page: 1,
      pageSize: 10,
      sortBy: 'time',
      ...params
    }
  });
}

// 点赞评论
export function likeComment(commentId) {
  return request({
    url: `${BASE_URL}/comments/${commentId}/like`,
    method: 'POST'
  });
}

// 取消点赞评论
export function unlikeComment(commentId) {
  return request({
    url: `${BASE_URL}/comments/${commentId}/unlike`,
    method: 'POST'
  });
}

// 获取热门评论
export function getHotComments(postId, params = {}) {
  return request({
    url: `${BASE_URL}/comments/hot`,
    method: 'GET',
    data: {
      postId,
      limit: 5,
      ...params
    }
  });
}

/**
 * 收藏相关API
 */

// 收藏帖子
export function bookmarkPost(postId) {
  return request({
    url: `${BASE_URL}/bookmarks`,
    method: 'POST',
    data: { postId }
  });
}

// 取消收藏帖子
export function unbookmarkPost(postId) {
  return request({
    url: `${BASE_URL}/bookmarks/${postId}`,
    method: 'DELETE'
  });
}

// 获取收藏列表
export function getBookmarksList(params = {}) {
  return request({
    url: `${BASE_URL}/bookmarks`,
    method: 'GET',
    data: {
      page: 1,
      pageSize: 10,
      ...params
    }
  });
}

// 检查收藏状态
export function checkBookmarkStatus(postId) {
  return request({
    url: `${BASE_URL}/bookmarks/status/${postId}`,
    method: 'GET'
  });
}

/**
 * 举报相关API
 */

// 提交举报
export function submitReport(data) {
  // 确保 targetId 是整数类型
  const requestData = {
    ...data,
    targetId: parseInt(data.targetId, 10)
  };

  console.log('提交举报数据:', requestData);

  return request({
    url: `${BASE_URL}/reports`,
    method: 'POST',
    data: requestData
  });
}

// 获取举报列表（管理员）
export function getReportsList(params = {}) {
  return request({
    url: `${BASE_URL}/reports`,
    method: 'GET',
    data: {
      page: 1,
      pageSize: 20,
      ...params
    }
  });
}

// 审核举报（管理员）
export function reviewReport(reportId, data) {
  return request({
    url: `${BASE_URL}/reports/${reportId}/review`,
    method: 'PUT',
    data
  });
}

/**
 * 通知相关API
 */

// 获取通知列表
export function getNotificationsList(params = {}) {
  return request({
    url: `${BASE_URL}/notifications`,
    method: 'GET',
    data: {
      page: 1,
      pageSize: 20,
      unreadOnly: false,
      ...params
    }
  });
}

// 标记通知为已读
export function markNotificationAsRead(notificationId) {
  return request({
    url: `${BASE_URL}/notifications/${notificationId}/read`,
    method: 'PUT'
  });
}

// 标记所有通知为已读
export function markAllNotificationsAsRead() {
  return request({
    url: `${BASE_URL}/notifications/read-all`,
    method: 'PUT'
  });
}

// 获取未读通知数量
export function getUnreadNotificationCount() {
  return request({
    url: `${BASE_URL}/notifications/unread-count`,
    method: 'GET'
  });
}

/**
 * 推荐相关API
 */

// 获取推荐数据 (包括推荐社团和推荐动态)
export function getRecommendData() {
  return request({
    url: `${BASE_URL}/community/recommend`,
    method: 'GET'
  });
}

// 创建通知
export function createNotification(data) {
  return request({
    url: `${BASE_URL}/notifications`,
    method: 'POST',
    data
  });
}

/**
 * 用户屏蔽相关API
 */

// 屏蔽用户
export function blockUser(userId) {
  return request({
    url: `${BASE_URL}/users/${userId}/block`,
    method: 'POST'
  });
}

// 取消屏蔽用户
export function unblockUser(userId) {
  return request({
    url: `${BASE_URL}/users/${userId}/unblock`,
    method: 'DELETE'
  });
}

// 获取屏蔽用户列表
export function getBlockedUsersList(params = {}) {
  return request({
    url: `${BASE_URL}/users/blocked`,
    method: 'GET',
    data: {
      page: 1,
      pageSize: 20,
      ...params
    }
  });
}

// 检查用户屏蔽状态
export function checkBlockStatus(userId) {
  return request({
    url: `${BASE_URL}/users/${userId}/block-status`,
    method: 'GET'
  });
}

/**
 * 文件上传相关API
 */

// 上传单张图片
export function uploadImage(file) {
  return request({
    url: `${BASE_URL}/upload/image`,
    method: 'POST',
    data: file,
    header: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

// 批量上传图片
export function uploadImages(files) {
  return request({
    url: `${BASE_URL}/upload/images`,
    method: 'POST',
    data: files,
    header: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

// 获取上传凭证
export function getUploadToken() {
  return request({
    url: `${BASE_URL}/upload/token`,
    method: 'GET'
  });
}

// 获取上传服务状态
export function getUploadStats() {
  return request({
    url: `${BASE_URL}/upload/stats`,
    method: 'GET'
  });
}

// 删除上传的文件
export function deleteUploadedFile(key) {
  return request({
    url: `${BASE_URL}/upload/file`,
    method: 'DELETE',
    data: { key }
  });
}

// 获取用户社区主页信息
export function getUserProfile(userId) {
  return request({
    url: `${BASE_URL}/users/${userId}/profile`,
    method: 'GET'
  });
}

// 获取用户发布的帖子列表
export function getUserPosts(userId, params = {}) {
  return request({
    url: `${BASE_URL}/users/${userId}/posts`,
    method: 'GET',
    data: {
      page: 1,
      pageSize: 20,
      ...params
    }
  });
}

// 关注用户
export function followUser(userId) {
  return request({
    url: `${BASE_URL}/users/${userId}/follow`,
    method: 'POST'
  });
}

// 取消关注用户
export function unfollowUser(userId) {
  return request({
    url: `${BASE_URL}/users/${userId}/unfollow`,
    method: 'POST'
  });
}

/**
 * 社区须知相关API
 */

// 同意社区须知
export function agreeCommunityTerms() {
  return request({
    url: `${BASE_URL}/community/terms/agree`,
    method: 'POST'
  });
}

// 获取社区须知状态
export function getCommunityTermsStatus() {
  return request({
    url: `${BASE_URL}/community/terms/status`,
    method: 'GET'
  });
}

/**
 * 工具函数
 */

// 格式化时间 - 仅支持中文 (已移除国际化)
export function formatTime(time) {
  if (!time) return '';

  const date = new Date(time);
  const now = new Date();
  const diff = now - date;

  // 小于1分钟
  if (diff < 60000) {
    return '刚刚';
  }

  // 小于1小时
  if (diff < 3600000) {
    return `${Math.floor(diff / 60000)}分钟前`;
  }

  // 小于1天
  if (diff < 86400000) {
    return `${Math.floor(diff / 3600000)}小时前`;
  }

  // 小于7天
  if (diff < 604800000) {
    return `${Math.floor(diff / 86400000)}天前`;
  }

  // 超过7天显示具体日期
  return date.toLocaleDateString('zh-CN');
}

// 格式化数字
export function formatNumber(num) {
  if (!num || num === 0) return '0';
  
  if (num < 1000) {
    return num.toString();
  }
  
  if (num < 10000) {
    return (num / 1000).toFixed(1) + 'k';
  }
  
  return (num / 10000).toFixed(1) + 'w';
}

// 处理图片URL
export function processImageUrl(url) {
  if (!url) return '';

  // 如果是相对路径，添加基础URL
  if (url.startsWith('/')) {
    try {
      const app = getApp();
      const baseUrl = app && app.globalData && app.globalData.baseUrl;
      return (baseUrl || '') + url;
    } catch (e) {
      return url;
    }
  }

  return url;
}

// 处理图片数组
export function processImages(images) {
  if (!images) return [];
  
  try {
    if (Array.isArray(images)) {
      return images
        .map(item => (typeof item === 'string' ? item.replace(/`/g, '').trim() : item))
        .filter(Boolean)
        .map(processImageUrl);
    }

    if (typeof images !== 'string') return [];

    const raw = images.trim();
    if (!raw) return [];

    const cleaned = raw.replace(/\\`/g, '').replace(/`/g, '').trim();

    try {
      const parsed = JSON.parse(cleaned);
      if (!Array.isArray(parsed)) return [];
      return parsed
        .map(item => (typeof item === 'string' ? item.replace(/\\`/g, '').replace(/`/g, '').trim() : item))
        .filter(Boolean)
        .map(processImageUrl);
    } catch (e) {
      const matches = cleaned.match(/https?:\/\/[^\s"'`\\\]]+/g) || [];
      return matches.map(item => item.replace(/\\`/g, '').replace(/`/g, '').trim()).filter(Boolean).map(processImageUrl);
    }
  } catch (error) {
    console.error('处理图片数组失败:', error);
    return [];
  }
}

export function normalizePostDetailResponse(payload) {
  if (!payload) return null;

  const data = payload.result && typeof payload.result === 'object' ? payload.result : payload;
  if (!data || typeof data !== 'object') return null;

  const post = data.post && typeof data.post === 'object' ? data.post : data;
  if (!post || typeof post !== 'object') return null;

  return {
    ...post,
    isLiked: typeof data.isLiked === 'boolean' ? data.isLiked : post.isLiked,
    isBookmarked: typeof data.isBookmarked === 'boolean' ? data.isBookmarked : post.isBookmarked
  };
}

/**
 * 关注相关API
 */

// 获取用户粉丝列表
export function getFollowers(userId, params = {}) {
  return request({
    url: `${BASE_URL}/users/${userId}/followers`,
    method: 'GET',
    data: { page: 1, pageSize: 20, ...params }
  });
}

// 获取用户关注列表
export function getFollowing(userId, params = {}) {
  return request({
    url: `${BASE_URL}/users/${userId}/following`,
    method: 'GET',
    data: { page: 1, pageSize: 20, ...params }
  });
}
