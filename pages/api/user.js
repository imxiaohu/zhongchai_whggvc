/**
 * 用户相关API
 */

import { request } from '../../utils/request.js';

/**
 * 获取用户信息
 * @returns {Promise} 返回用户信息
 */
export const getUserInfo = () => {
  return request({
    url: '/api/user/info',
    method: 'GET'
  });
};

/**
 * 更新用户信息
 * @param {Object} data 用户信息
 * @returns {Promise} 返回更新结果
 */
export const updateUserInfo = (data) => {
  return request({
    url: '/api/user/info',
    method: 'PUT',
    data
  });
};

/**
 * 获取用户设置
 * @returns {Promise} 返回用户设置
 */
export const getUserSettings = () => {
  return request({
    url: '/api/user/settings',
    method: 'GET'
  });
};

/**
 * 更新用户设置
 * @param {Object} settings 设置数据
 * @param {string} settings.language 语言设置
 * @param {string} settings.theme 主题设置
 * @param {boolean} settings.syncEnabled 是否启用同步
 * @param {string} settings.syncFrequency 同步频率
 * @param {string} settings.syncTimeRange 同步时间范围
 * @param {boolean} settings.notificationEnabled 是否启用通知
 * @param {Object} settings.clientInfo 客户端信息
 * @returns {Promise} 返回更新结果
 */
export const updateUserSettings = (settings) => {
  return request({
    url: '/api/user/settings',
    method: 'PUT',
    data: settings
  });
};

/**
 * 重置用户设置
 * @returns {Promise} 返回重置结果
 */
export const resetUserSettings = () => {
  return request({
    url: '/api/user/settings/reset',
    method: 'POST'
  });
};

/**
 * 获取用户统计信息
 * @returns {Promise} 返回统计信息
 */
export const getUserStatistics = () => {
  return request({
    url: '/api/user/statistics',
    method: 'GET'
  });
};

/**
 * 上传用户头像
 * @param {File} file 头像文件
 * @returns {Promise} 返回上传结果
 */
export const uploadAvatar = (file) => {
  return request({
    url: '/api/user/avatar',
    method: 'POST',
    data: {
      file
    },
    header: {
      'Content-Type': 'multipart/form-data'
    }
  });
};

/**
 * 绑定学校账号
 * @param {Object} data 绑定数据
 * @param {string} data.username 学校账号用户名
 * @param {string} data.password 学校账号密码
 * @returns {Promise} 返回绑定结果
 */
export const bindSchoolAccount = (data) => {
  return request({
    url: '/api/user/bind-school',
    method: 'POST',
    data
  });
};

/**
 * 解绑学校账号
 * @returns {Promise} 返回解绑结果
 */
export const unbindSchoolAccount = () => {
  return request({
    url: '/api/user/unbind-school',
    method: 'POST'
  });
};

/**
 * 验证学校账号
 * @param {Object} data 验证数据
 * @param {string} data.username 学校账号用户名
 * @param {string} data.password 学校账号密码
 * @returns {Promise} 返回验证结果
 */
export const verifySchoolAccount = (data) => {
  return request({
    url: '/api/user/verify-school',
    method: 'POST',
    data
  });
};

/**
 * 获取用户活动日志
 * @param {Object} params 查询参数
 * @param {number} params.page 页码
 * @param {number} params.pageSize 每页数量
 * @param {string} params.type 日志类型
 * @returns {Promise} 返回活动日志
 */
export const getUserActivityLog = (params = {}) => {
  return request({
    url: '/api/user/activity-log',
    method: 'GET',
    data: params
  });
};

/**
 * 清除用户缓存
 * @param {Array} cacheTypes 要清除的缓存类型
 * @returns {Promise} 返回清除结果
 */
export const clearUserCache = (cacheTypes = []) => {
  return request({
    url: '/api/user/clear-cache',
    method: 'POST',
    data: {
      cacheTypes
    }
  });
};

/**
 * 导出用户数据
 * @param {Array} dataTypes 要导出的数据类型
 * @returns {Promise} 返回导出结果
 */
export const exportUserData = (dataTypes = []) => {
  return request({
    url: '/api/user/export-data',
    method: 'POST',
    data: {
      dataTypes
    }
  });
};

/**
 * 删除用户账号
 * @param {Object} data 删除确认数据
 * @param {string} data.password 用户密码确认
 * @param {string} data.reason 删除原因
 * @returns {Promise} 返回删除结果
 */
export const deleteUserAccount = (data) => {
  return request({
    url: '/api/user/delete-account',
    method: 'DELETE',
    data
  });
};

/**
 * 获取用户偏好设置
 * @returns {Promise} 返回偏好设置
 */
export const getUserPreferences = () => {
  return request({
    url: '/api/user/preferences',
    method: 'GET'
  });
};

/**
 * 更新用户偏好设置
 * @param {Object} preferences 偏好设置
 * @returns {Promise} 返回更新结果
 */
export const updateUserPreferences = (preferences) => {
  return request({
    url: '/api/user/preferences',
    method: 'PUT',
    data: preferences
  });
};

/**
 * 获取用户通知设置
 * @returns {Promise} 返回通知设置
 */
export const getNotificationSettings = () => {
  return request({
    url: '/api/user/notification-settings',
    method: 'GET'
  });
};

/**
 * 更新用户通知设置
 * @param {Object} settings 通知设置
 * @returns {Promise} 返回更新结果
 */
export const updateNotificationSettings = (settings) => {
  return request({
    url: '/api/user/notification-settings',
    method: 'PUT',
    data: settings
  });
};

/**
 * 获取用户隐私设置
 * @returns {Promise} 返回隐私设置
 */
export const getPrivacySettings = () => {
  return request({
    url: '/api/user/privacy-settings',
    method: 'GET'
  });
};

/**
 * 更新用户隐私设置
 * @param {Object} settings 隐私设置
 * @returns {Promise} 返回更新结果
 */
export const updatePrivacySettings = (settings) => {
  return request({
    url: '/api/user/privacy-settings',
    method: 'PUT',
    data: settings
  });
};
