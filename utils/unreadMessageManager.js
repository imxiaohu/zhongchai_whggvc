/**
 * 未读消息管理工具
 * 用于全局管理未读消息数量和 tabbar 角标
 */

import { getUnreadNotificationCount } from '@/pages/api/notifications.js';

class UnreadMessageManager {
  constructor() {
    this.unreadCount = 0;
    this.isInitialized = false;
    this.listeners = [];
    this.isFetching = false; // 添加防重复请求标志
  }

  /**
   * 初始化未读消息管理器
   */
  async init() {
    if (this.isInitialized) {
      return;
    }
    
    console.log('UnreadMessageManager: 初始化未读消息管理器');
    this.isInitialized = true;
    
    // 获取初始未读消息数量
    await this.fetchUnreadCount();
  }

  /**
   * 获取未读消息数量
   */
  async fetchUnreadCount() {
    try {
      // 检查登录状态
      const token = uni.getStorageSync('token');
      if (!token) {
        console.log('UnreadMessageManager: 用户未登录，设置未读数量为 0');
        this.setUnreadCount(0);
        return;
      }
  
      // 防止短时间内重复请求
      if (this.isFetching) {
        console.log('UnreadMessageManager: 正在获取中，跳过本次请求');
        return;
      }
  
      this.isFetching = true;
      console.log('UnreadMessageManager: 开始获取未读消息数量');
  
      // 调用 API 获取未读消息数量
      const count = await getUnreadNotificationCount();
      console.log('UnreadMessageManager: 获取到未读消息数量:', count);
      this.setUnreadCount(count);
        
      this.isFetching = false;
    } catch (error) {
      console.error('UnreadMessageManager: 获取未读消息数量出错:', error);
      this.isFetching = false;
      this.setUnreadCount(0);
      // 不抛出错误，避免触发全局错误处理器
    }
  }

  /**
   * 设置未读消息数量
   * @param {number} count 未读消息数量
   */
  setUnreadCount(count) {
    const oldCount = this.unreadCount;
    this.unreadCount = count;
    
    console.log('UnreadMessageManager: 设置未读消息数量:', count);
    
    // 更新 tabbar 角标
    this.updateTabBarBadge();
    
    // 通知监听器
    this.notifyListeners(count, oldCount);
  }

  /**
   * 获取当前未读消息数量
   */
  getUnreadCount() {
    return this.unreadCount;
  }

  /**
   * 更新 tabbar 角标
   */
  updateTabBarBadge() {
    try {
      // 获取当前页面路径，判断是否在 TabBar 页面
      const pages = getCurrentPages();
      if (!pages || pages.length === 0) {
        return;
      }
      
      const currentPage = pages[pages.length - 1];
      const currentRoute = currentPage.route;
      
      // TabBar 页面列表（移除知识库后）
      const tabBarPages = [
        'pages/index/index',      // 索引 0 - 首页
        'pages/schedule/index',   // 索引 1 - 课表
        'pages/evaluation/list',  // 索引 2 - 评教
        'pages/user/index'        // 索引 3 - 我的
      ];
      
      const tabBarIndex = tabBarPages.indexOf(currentRoute);
      
      // 如果当前不在 TabBar 页面，不更新角标
      if (tabBarIndex === -1) {
        return;
      }
      
      if (this.unreadCount > 0) {
        // 在"我的"页面（索引3）显示角标
        uni.setTabBarBadge({
          index: 3,
          text: this.unreadCount.toString(),
          success: () => {
            console.log('UnreadMessageManager: 设置我的页面角标成功:', this.unreadCount);
          },
          fail: (error) => {
            // 静默失败，不显示警告
          }
        });
      } else {
        // 移除角标
        uni.removeTabBarBadge({
          index: 3,
          success: () => {
            console.log('UnreadMessageManager: 移除我的页面角标成功');
          },
          fail: (error) => {
            // 静默失败，不显示警告
          }
        });
      }
    } catch (error) {
      console.error('UnreadMessageManager: 更新 tabbar 角标失败:', error);
    }
  }

  /**
   * 添加监听器
   * @param {Function} listener 监听器函数，接收 (newCount, oldCount) 参数
   */
  addListener(listener) {
    if (typeof listener === 'function') {
      this.listeners.push(listener);
    }
  }

  /**
   * 移除监听器
   * @param {Function} listener 要移除的监听器函数
   */
  removeListener(listener) {
    const index = this.listeners.indexOf(listener);
    if (index > -1) {
      this.listeners.splice(index, 1);
    }
  }

  /**
   * 通知所有监听器
   * @param {number} newCount 新的未读数量
   * @param {number} oldCount 旧的未读数量
   */
  notifyListeners(newCount, oldCount) {
    this.listeners.forEach(listener => {
      try {
        listener(newCount, oldCount);
      } catch (error) {
        console.error('UnreadMessageManager: 监听器执行出错:', error);
      }
    });
  }

  /**
   * 标记消息为已读（减少未读数量）
   * @param {number} count 已读消息数量，默认为1
   */
  markAsRead(count = 1) {
    const newCount = Math.max(0, this.unreadCount - count);
    this.setUnreadCount(newCount);
  }

  /**
   * 增加未读消息数量
   * @param {number} count 新增的未读消息数量，默认为1
   */
  addUnreadCount(count = 1) {
    const newCount = this.unreadCount + count;
    this.setUnreadCount(newCount);
  }

  /**
   * 重置未读消息数量
   */
  reset() {
    this.setUnreadCount(0);
  }

  /**
   * 销毁管理器
   */
  destroy() {
    this.listeners = [];
    this.unreadCount = 0;
    this.isInitialized = false;

    // 移除 tabbar 角标
    try {
      uni.removeTabBarBadge({
        index: 3,
        success: () => {
          console.log('UnreadMessageManager: 销毁时移除角标成功');
        },
        fail: (error) => {
          // 静默失败，不显示警告
        }
      });
    } catch (error) {
      console.error('UnreadMessageManager: 销毁时移除角标失败:', error);
    }
  }
}

// 创建全局单例
const unreadMessageManager = new UnreadMessageManager();

export default unreadMessageManager;
