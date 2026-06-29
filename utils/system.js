/**
 * 系统信息工具类
 * 用于获取设备系统信息，包括状态栏高度、安全区域等
 */

class SystemUtils {
  constructor() {
    this.systemInfo = null;
    this.statusBarHeight = 0;
    this.safeAreaInsets = {
      top: 0,
      bottom: 0,
      left: 0,
      right: 0
    };
  }

  /**
   * 获取系统信息
   * @returns {Object} 系统信息对象
   */
  getSystemInfo() {
    if (this.systemInfo) {
      return this.systemInfo;
    }

    try {
      this.systemInfo = uni.getSystemInfoSync();
      this.processSystemInfo();
      return this.systemInfo;
    } catch (error) {
      console.warn('获取系统信息失败:', error);
      this.setFallbackValues();
      return this.systemInfo;
    }
  }

  /**
   * 处理系统信息，设置各平台特定值
   */
  processSystemInfo() {
    if (!this.systemInfo) return;

    // 设置状态栏高度
    this.statusBarHeight = this.systemInfo.statusBarHeight || 0;

    // 针对不同平台进行调整
    // #ifdef MP-WEIXIN
    console.log('微信小程序状态栏高度:', this.statusBarHeight);
    this.processMiniProgramInfo();
    // #endif

    // #ifdef APP-PLUS
    console.log('App状态栏高度:', this.statusBarHeight);
    this.processAppInfo();
    // #endif

    // #ifdef H5
    this.statusBarHeight = 0; // H5不需要状态栏高度
    console.log('H5平台，状态栏高度设为0');
    // #endif

    // 设置安全区域
    this.setSafeAreaInsets();
  }

  /**
   * 处理微信小程序特定信息
   */
  processMiniProgramInfo() {
    // 微信小程序可能需要特殊处理
    if (this.systemInfo.model && this.systemInfo.model.includes('iPhone')) {
      // iPhone设备可能需要额外处理
      this.handleiPhoneSpecialCase();
    }
  }

  /**
   * 处理App特定信息
   */
  processAppInfo() {
    // App平台可能需要特殊处理
    // 可以根据需要添加特定逻辑
  }

  /**
   * 处理iPhone特殊情况
   */
  handleiPhoneSpecialCase() {
    // 针对iPhone的刘海屏等特殊情况处理
    if (this.systemInfo.screenHeight >= 812) {
      // 可能是刘海屏设备
      console.log('检测到可能的刘海屏设备');
    }
  }

  /**
   * 设置安全区域内边距
   */
  setSafeAreaInsets() {
    if (this.systemInfo.safeArea) {
      const safeArea = this.systemInfo.safeArea;
      this.safeAreaInsets = {
        top: safeArea.top || 0,
        bottom: this.systemInfo.screenHeight - (safeArea.bottom || this.systemInfo.screenHeight),
        left: safeArea.left || 0,
        right: this.systemInfo.screenWidth - (safeArea.right || this.systemInfo.screenWidth)
      };
    }
  }

  /**
   * 设置默认值（当获取系统信息失败时）
   */
  setFallbackValues() {
    this.statusBarHeight = 20; // 默认状态栏高度
    this.systemInfo = {
      statusBarHeight: this.statusBarHeight,
      screenWidth: 375,
      screenHeight: 667,
      platform: 'unknown'
    };
    console.warn('使用默认系统信息值');
  }

  /**
   * 获取状态栏高度
   * @returns {number} 状态栏高度（像素）
   */
  getStatusBarHeight() {
    if (!this.systemInfo) {
      this.getSystemInfo();
    }
    return this.statusBarHeight;
  }

  /**
   * 获取安全区域内边距
   * @returns {Object} 安全区域内边距对象
   */
  getSafeAreaInsets() {
    if (!this.systemInfo) {
      this.getSystemInfo();
    }
    return this.safeAreaInsets;
  }

  /**
   * 获取导航栏高度（状态栏 + 标题栏）
   * @returns {number} 导航栏总高度
   */
  getNavigationBarHeight() {
    const titleBarHeight = 44; // 标准标题栏高度
    return this.getStatusBarHeight() + titleBarHeight;
  }

  /**
   * 判断是否为刘海屏设备
   * @returns {boolean} 是否为刘海屏
   */
  isNotchDevice() {
    if (!this.systemInfo) {
      this.getSystemInfo();
    }
    
    // 简单判断：状态栏高度大于24px通常表示刘海屏
    return this.statusBarHeight > 24;
  }

  /**
   * 获取可用屏幕高度（去除状态栏和安全区域）
   * @returns {number} 可用屏幕高度
   */
  getAvailableScreenHeight() {
    if (!this.systemInfo) {
      this.getSystemInfo();
    }
    
    return this.systemInfo.screenHeight - this.statusBarHeight - this.safeAreaInsets.bottom;
  }

  /**
   * 重置系统信息（强制重新获取）
   */
  reset() {
    this.systemInfo = null;
    this.statusBarHeight = 0;
    this.safeAreaInsets = {
      top: 0,
      bottom: 0,
      left: 0,
      right: 0
    };
  }
}

export function normalizeNavHeightReadyPayload(payload) {
  if (payload == null) return 0;
  if (typeof payload === 'number' && Number.isFinite(payload)) return payload;
  if (typeof payload === 'string') return parsePxNumber(payload);
  if (typeof payload === 'object') {
    const height = payload.height;
    if (typeof height === 'number' && Number.isFinite(height)) return height;
    const heightPx = payload.heightPx;
    if (typeof heightPx === 'string') return parsePxNumber(heightPx);
  }
  return 0;
}

export function getTabBarHeightFromSystemInfo(systemInfo, baseHeightPx = 50) {
  if (!systemInfo) return baseHeightPx;
  const safeBottom = getSafeAreaBottomFromSystemInfo(systemInfo);
  return baseHeightPx + safeBottom;
}

function parsePxNumber(value) {
  const n = parseFloat(String(value).trim().replace(/px$/i, ''));
  return Number.isFinite(n) ? n : 0;
}

function getSafeAreaBottomFromSystemInfo(systemInfo) {
  if (!systemInfo) return 0;
  if (systemInfo.safeAreaInsets && typeof systemInfo.safeAreaInsets.bottom === 'number') {
    return systemInfo.safeAreaInsets.bottom;
  }
  if (systemInfo.safeArea && typeof systemInfo.safeArea.bottom === 'number' && typeof systemInfo.screenHeight === 'number') {
    const bottom = systemInfo.screenHeight - systemInfo.safeArea.bottom;
    return Number.isFinite(bottom) ? Math.max(0, bottom) : 0;
  }
  return 0;
}

// 创建单例实例
const systemUtils = new SystemUtils();

// 导出单例和类
export default systemUtils;
export { SystemUtils };

// 兼容CommonJS导出
if (typeof module !== 'undefined' && module.exports) {
  module.exports = systemUtils;
  module.exports.SystemUtils = SystemUtils;
  module.exports.normalizeNavHeightReadyPayload = normalizeNavHeightReadyPayload;
  module.exports.getTabBarHeightFromSystemInfo = getTabBarHeightFromSystemInfo;
}
