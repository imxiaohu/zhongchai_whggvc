/**
 * 通知相关API
 */

import { request } from '../../utils/request.js';

/**
 * 获取未读消息数量
 * @returns {Promise<number>} 未读消息数量
 */
export async function getUnreadNotificationCount() {
  try {
    console.log('API: 获取未读消息数量');
    
    const response = await request({
      url: '/api/notifications/unread-count',
      method: 'GET'
    });

    console.log('API: 未读消息数量响应:', response);

    if (response && response.success) {
      const count = response.result?.unreadCount || 0;
      console.log('API: 解析到未读消息数量:', count);
      return count;
    } else {
      console.warn('API: 获取未读消息数量失败:', response);
      return 0;
    }
  } catch (error) {
    console.error('API: 获取未读消息数量出错:', error);
    return 0;
  }
}

/**
 * 标记消息为已读
 * @param {string|number} messageId 消息ID
 * @returns {Promise<boolean>} 是否成功
 */
export async function markMessageAsRead(messageId) {
  try {
    console.log('API: 标记消息为已读:', messageId);
    
    const response = await request({
      url: `/api/notifications/${messageId}/read`,
      method: 'PUT'
    });

    console.log('API: 标记已读响应:', response);

    if (response && response.success) {
      console.log('API: 标记消息已读成功');
      return true;
    } else {
      console.warn('API: 标记消息已读失败:', response);
      return false;
    }
  } catch (error) {
    console.error('API: 标记消息已读出错:', error);
    return false;
  }
}

/**
 * 批量标记消息为已读
 * @param {Array<string|number>} messageIds 消息ID数组
 * @returns {Promise<boolean>} 是否成功
 */
export async function markMessagesAsRead(messageIds) {
  try {
    console.log('API: 批量标记消息为已读:', messageIds);
    
    const response = await request({
      url: '/api/notifications/read-all',
      method: 'PUT'
    });

    console.log('API: 批量标记已读响应:', response);

    if (response && response.success) {
      console.log('API: 批量标记消息已读成功');
      return true;
    } else {
      console.warn('API: 批量标记消息已读失败:', response);
      return false;
    }
  } catch (error) {
    console.error('API: 批量标记消息已读出错:', error);
    return false;
  }
}

/**
 * 获取通知列表
 * @param {Object} options 查询选项
 * @param {number} options.page 页码，默认1
 * @param {number} options.pageSize 每页数量，默认20
 * @param {string} options.type 通知类型，可选
 * @returns {Promise<Object>} 通知列表数据
 */
export async function getNotificationList(options = {}) {
  try {
    const {
      page = 1,
      pageSize = 20,
      type = ''
    } = options;

    console.log('API: 获取通知列表:', { page, pageSize, type });
    
    const response = await request({
      url: '/api/notifications/list',
      method: 'GET',
      data: {
        page,
        pageSize,
        type
      }
    });

    console.log('API: 通知列表响应:', response);

    if (response && response.success) {
      const result = {
        list: response.result?.records || [],
        total: response.result?.total || 0,
        page: response.result?.current || page,
        pageSize: response.result?.size || pageSize,
        hasMore: (response.result?.current || page) < (response.result?.pages || 1)
      };
      console.log('API: 解析通知列表数据:', result);
      return result;
    } else {
      console.warn('API: 获取通知列表失败:', response);
      return {
        list: [],
        total: 0,
        page: 1,
        pageSize: 20,
        hasMore: false
      };
    }
  } catch (error) {
    console.error('API: 获取通知列表出错:', error);
    return {
      list: [],
      total: 0,
      page: 1,
      pageSize: 20,
      hasMore: false
    };
  }
}
