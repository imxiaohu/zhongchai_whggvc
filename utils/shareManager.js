/**
 * 微信小程序页面分享管理器
 * 统一管理页面分享功能，包含用户追踪和分析数据收集
 */

class ShareManager {
  constructor() {
    this.defaultShareConfig = {
      title: '众柴智慧校园 - 让评教更简单',
      path: '/pages/index/index',
      imageUrl: '/static/images/share-logo.png'
    };
  }

  /**
   * 生成分享配置
   * @param {Object} options 分享配置选项
   * @param {string} options.title 分享标题
   * @param {string} options.path 分享路径
   * @param {string} options.imageUrl 分享图片
   * @param {Object} options.trackingData 追踪数据
   * @returns {Object} 微信小程序分享配置
   */
  generateShareConfig(options = {}) {
    const {
      title = this.defaultShareConfig.title,
      path = this.defaultShareConfig.path,
      imageUrl = this.defaultShareConfig.imageUrl,
      trackingData = {}
    } = options;

    // 获取分享者信息
    const sharerInfo = this.getSharerInfo();
    
    // 生成追踪参数
    const trackingParams = this.generateTrackingParams(sharerInfo, trackingData);
    
    // 构建完整的分享路径
    const fullPath = this.buildSharePath(path, trackingParams);

    return {
      title,
      path: fullPath,
      imageUrl,
      success: (res) => {
        console.log('分享成功:', res);
        this.recordShareEvent('success', {
          title,
          path: fullPath,
          sharerInfo,
          trackingData
        });
      },
      fail: (error) => {
        console.error('分享失败:', error);
        this.recordShareEvent('fail', {
          title,
          path: fullPath,
          sharerInfo,
          trackingData,
          error
        });
      }
    };
  }

  /**
   * 生成朋友圈分享配置
   * @param {Object} options 分享配置选项
   * @returns {Object} 朋友圈分享配置
   */
  generateTimelineConfig(options = {}) {
    const {
      title = this.defaultShareConfig.title,
      query = '',
      imageUrl = this.defaultShareConfig.imageUrl,
      trackingData = {}
    } = options;

    // 获取分享者信息
    const sharerInfo = this.getSharerInfo();
    
    // 生成追踪参数
    const trackingParams = this.generateTrackingParams(sharerInfo, trackingData);
    
    // 构建查询字符串
    const fullQuery = this.buildShareQuery(query, trackingParams);

    return {
      title,
      query: fullQuery,
      imageUrl,
      success: (res) => {
        console.log('朋友圈分享成功:', res);
        this.recordShareEvent('timeline_success', {
          title,
          query: fullQuery,
          sharerInfo,
          trackingData
        });
      },
      fail: (error) => {
        console.error('朋友圈分享失败:', error);
        this.recordShareEvent('timeline_fail', {
          title,
          query: fullQuery,
          sharerInfo,
          trackingData,
          error
        });
      }
    };
  }

  /**
   * 获取分享者信息
   * @returns {Object} 分享者信息
   */
  getSharerInfo() {
    try {
      const userInfo = uni.getStorageSync('userInfo');
      const loginType = uni.getStorageSync('loginType');
      
      let sharerInfo = {
        userId: null,
        userType: 'anonymous',
        timestamp: Date.now()
      };

      if (userInfo) {
        const user = typeof userInfo === 'string' ? JSON.parse(userInfo) : userInfo;
        sharerInfo = {
          userId: user.id || user.openid || 'unknown',
          userType: loginType || 'unknown',
          userName: user.realname || user.nickname || 'unknown',
          timestamp: Date.now()
        };
      }

      return sharerInfo;
    } catch (error) {
      console.error('获取分享者信息失败:', error);
      return {
        userId: 'error',
        userType: 'error',
        timestamp: Date.now()
      };
    }
  }

  /**
   * 生成追踪参数
   * @param {Object} sharerInfo 分享者信息
   * @param {Object} trackingData 额外追踪数据
   * @returns {Object} 追踪参数
   */
  generateTrackingParams(sharerInfo, trackingData = {}) {
    const systemInfo = uni.getSystemInfoSync();
    
    return {
      // 分享者信息
      sharer_id: sharerInfo.userId,
      sharer_type: sharerInfo.userType,
      share_time: sharerInfo.timestamp,
      
      // 设备信息
      device_brand: systemInfo.brand || 'unknown',
      device_model: systemInfo.model || 'unknown',
      device_system: systemInfo.system || 'unknown',
      device_platform: systemInfo.platform || 'unknown',
      
      // 应用信息
      app_version: systemInfo.version || 'unknown',
      sdk_version: systemInfo.SDKVersion || 'unknown',
      
      // 额外追踪数据
      ...trackingData,
      
      // 分享标识
      from_share: 'true',
      share_id: this.generateShareId()
    };
  }

  /**
   * 构建分享路径
   * @param {string} basePath 基础路径
   * @param {Object} trackingParams 追踪参数
   * @returns {string} 完整分享路径
   */
  buildSharePath(basePath, trackingParams) {
    const queryString = Object.keys(trackingParams)
      .map(key => `${encodeURIComponent(key)}=${encodeURIComponent(trackingParams[key])}`)
      .join('&');
    
    const separator = basePath.includes('?') ? '&' : '?';
    return `${basePath}${separator}${queryString}`;
  }

  /**
   * 构建分享查询字符串
   * @param {string} baseQuery 基础查询
   * @param {Object} trackingParams 追踪参数
   * @returns {string} 完整查询字符串
   */
  buildShareQuery(baseQuery, trackingParams) {
    const queryString = Object.keys(trackingParams)
      .map(key => `${encodeURIComponent(key)}=${encodeURIComponent(trackingParams[key])}`)
      .join('&');
    
    if (baseQuery) {
      return `${baseQuery}&${queryString}`;
    }
    return queryString;
  }

  /**
   * 生成唯一分享ID
   * @returns {string} 分享ID
   */
  generateShareId() {
    return `share_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * 记录分享事件
   * @param {string} eventType 事件类型
   * @param {Object} eventData 事件数据
   */
  recordShareEvent(eventType, eventData) {
    try {
      const shareEvent = {
        type: eventType,
        data: eventData,
        timestamp: Date.now()
      };

      // 保存到本地存储
      const shareHistory = uni.getStorageSync('shareHistory') || [];
      shareHistory.push(shareEvent);
      
      // 只保留最近100条记录
      if (shareHistory.length > 100) {
        shareHistory.splice(0, shareHistory.length - 100);
      }
      
      uni.setStorageSync('shareHistory', shareHistory);

      // 如果有网络，尝试上报到服务器
      this.reportShareEvent(shareEvent);

    } catch (error) {
      console.error('记录分享事件失败:', error);
    }
  }

  /**
   * 上报分享事件到服务器
   * @param {Object} shareEvent 分享事件
   */
  async reportShareEvent(shareEvent) {
    try {
      const token = uni.getStorageSync('token');
      if (!token) {
        console.log('未登录，跳过分享事件上报');
        return;
      }

      // 发送到服务器
      await uni.request({
        url: '/api/analytics/share',
        method: 'POST',
        header: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        data: shareEvent
      });

      console.log('分享事件上报成功');
    } catch (error) {
      console.error('分享事件上报失败:', error);
    }
  }

  /**
   * 处理分享链接打开
   * @param {Object} query 页面查询参数
   */
  handleShareOpen(query) {
    try {
      if (query.from_share === 'true') {
        console.log('通过分享链接打开:', query);
        
        // 记录分享链接打开事件
        const openEvent = {
          type: 'share_open',
          data: {
            shareId: query.share_id,
            sharerId: query.sharer_id,
            sharerType: query.sharer_type,
            shareTime: query.share_time,
            openTime: Date.now(),
            deviceInfo: uni.getSystemInfoSync()
          },
          timestamp: Date.now()
        };

        this.recordShareEvent('share_open', openEvent.data);

        // 显示分享来源提示
        if (query.sharer_type !== 'anonymous') {
          setTimeout(() => {
            uni.showToast({
              title: '来自朋友的分享',
              icon: 'none',
              duration: 2000
            });
          }, 1000);
        }
      }
    } catch (error) {
      console.error('处理分享链接打开失败:', error);
    }
  }

  /**
   * 获取分享统计数据
   * @returns {Object} 分享统计
   */
  getShareStatistics() {
    try {
      const shareHistory = uni.getStorageSync('shareHistory') || [];
      
      const stats = {
        totalShares: 0,
        successfulShares: 0,
        failedShares: 0,
        timelineShares: 0,
        shareOpens: 0,
        lastShareTime: null
      };

      shareHistory.forEach(event => {
        switch (event.type) {
          case 'success':
            stats.totalShares++;
            stats.successfulShares++;
            stats.lastShareTime = event.timestamp;
            break;
          case 'fail':
            stats.totalShares++;
            stats.failedShares++;
            break;
          case 'timeline_success':
            stats.timelineShares++;
            break;
          case 'share_open':
            stats.shareOpens++;
            break;
        }
      });

      return stats;
    } catch (error) {
      console.error('获取分享统计失败:', error);
      return {};
    }
  }
}

// 创建单例实例
const shareManager = new ShareManager();

export default shareManager;
