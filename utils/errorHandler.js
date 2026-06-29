/**
 * 全局错误处理工具
 * 用于统一管理服务器错误和超时的处理
 */

// 错误类型常量
export const ERROR_TYPES = {
  TIMEOUT: 'timeout',
  PROXY_FAILED: 'proxy_failed',
  SERVER_DOWN: 'server_down',
  NETWORK_ERROR: 'network_error'
}

// 需要显示ServerDownPage的错误条件
const SHOW_SERVER_DOWN_CONDITIONS = [
  'timeout',
  '超时',
  'TIMEOUT',
  'ETIMEDOUT',
  '代理请求失败',
  'proxy failed',
  'PROXY_FAILED',
  'server down',
  'SERVER_DOWN',
  '服务器无响应',
  '连接超时',
  'Connection timeout',
  'Request timeout',
  '网络连接失败',
  'Network request failed',
  'ERR_NETWORK',
  'ERR_TIMEOUT'
]

/**
 * 判断是否应该显示ServerDownPage
 * @param {Object} error - 错误对象
 * @param {string} errorMessage - 错误消息
 * @returns {boolean} 是否显示ServerDownPage
 */
export function shouldShowServerDownPage(error, errorMessage = '') {
  const message = errorMessage || error?.message || error?.errMsg || ''
  const statusCode = error?.statusCode || error?.status
  
  // 检查HTTP状态码
  if (statusCode >= 500 && statusCode < 600) {
    return true
  }

  // 检查408超时状态码
  if (statusCode === 408) {
    return true
  }

  // 检查是否为超时错误
  if (error?.isTimeout) {
    return true
  }
  
  // 检查错误消息
  return SHOW_SERVER_DOWN_CONDITIONS.some(condition => 
    message.toLowerCase().includes(condition.toLowerCase())
  )
}

// 防重复跳转的状态管理
let isNavigatingToErrorPage = false;
let lastErrorPageNavigation = 0;
const ERROR_PAGE_NAVIGATION_COOLDOWN = 3000; // 3秒冷却时间

/**
 * 显示ServerDownPage
 * @param {Object} options - 配置选项
 * @param {string} options.errorMessage - 错误消息
 * @param {boolean} options.showDetails - 是否显示错误详情
 * @param {Function} options.retryFunction - 重试函数
 */
export function showServerDownPage(options = {}) {
  const {
    errorMessage = '服务器连接超时或暂时不可用',
    showDetails = false,
    retryFunction = null
  } = options

  console.log('showServerDownPage: 检测到服务器错误，但已禁用自动跳转', {
    errorMessage,
    showDetails,
    hasRetryFunction: !!retryFunction
  })

  // 检查错误提示是否被抑制
  if (isErrorSuppressed()) {
    console.log('showServerDownPage: 错误提示被抑制，不显示任何提示');
    return;
  }

  // 不再自动跳转到错误页面，改为显示Toast提示
  console.log('showServerDownPage: 显示Toast提示而不是跳转错误页面');

  // 显示友好的Toast提示
  const toastMessage = errorMessage.includes('服务器') ?
    '学校服务器维护中，请稍后再试' :
    errorMessage;

  uni.showToast({
    title: toastMessage,
    icon: 'none',
    duration: 3000
  });

  // 触发全局事件，通知页面更新状态
  uni.$emit('serverMaintenanceDetected', {
    errorMessage,
    timestamp: Date.now()
  });
}

/**
 * 显示服务器维护提示（新增方法）
 * @param {string} message - 提示消息
 */
export function showMaintenanceToast(message = '学校服务器维护中，功能暂时不可用') {
  console.log('showMaintenanceToast: 显示维护提示', message);

  uni.showToast({
    title: message,
    icon: 'none',
    duration: 3000
  });
}

/**
 * 重置错误页面跳转状态
 * 用于在用户从错误页面返回时重置防重复跳转状态
 */
export function resetErrorPageNavigationState() {
  console.log('resetErrorPageNavigationState: 重置错误页面跳转状态');
  isNavigatingToErrorPage = false;
  lastErrorPageNavigation = 0;
}

/**
 * 获取当前错误页面跳转状态
 * @returns {Object} 当前状态信息
 */
export function getErrorPageNavigationState() {
  return {
    isNavigatingToErrorPage,
    lastErrorPageNavigation,
    timeSinceLastNavigation: Date.now() - lastErrorPageNavigation,
    cooldownRemaining: Math.max(0, ERROR_PAGE_NAVIGATION_COOLDOWN - (Date.now() - lastErrorPageNavigation))
  };
}

/**
 * 检查用户是否主动选择重试
 * @returns {boolean} 是否为用户主动重试
 */
export function isUserRetry() {
  try {
    const retryFlag = uni.getStorageSync('userRetryFlag');
    const retryTimestamp = uni.getStorageSync('userRetryTimestamp');

    if (!retryFlag || !retryTimestamp) {
      return false;
    }

    // 检查重试标志是否在5分钟内设置的（防止过期的标志）
    const now = Date.now();
    const timestamp = parseInt(retryTimestamp);
    const isRecent = (now - timestamp) < 5 * 60 * 1000; // 5分钟

    console.log('isUserRetry: 检查用户重试标志', {
      retryFlag,
      retryTimestamp,
      isRecent,
      timeDiff: now - timestamp
    });

    return retryFlag === 'true' && isRecent;
  } catch (error) {
    console.error('isUserRetry: 检查重试标志失败', error);
    return false;
  }
}

/**
 * 清除用户重试标志
 */
export function clearUserRetryFlag() {
  try {
    uni.removeStorageSync('userRetryFlag');
    uni.removeStorageSync('userRetryTimestamp');
    console.log('clearUserRetryFlag: 已清除用户重试标志');
  } catch (error) {
    console.error('clearUserRetryFlag: 清除重试标志失败', error);
  }
}

/**
 * 检查错误提示是否被抑制
 * @returns {boolean} 是否被抑制
 */
export function isErrorSuppressed() {
  try {
    const suppressUntil = uni.getStorageSync('errorSuppressUntil');
    if (!suppressUntil) {
      return false;
    }

    const suppressTime = parseInt(suppressUntil);
    const now = Date.now();
    const isSuppressed = now < suppressTime;

    console.log('isErrorSuppressed: 检查错误提示抑制状态', {
      suppressUntil: new Date(suppressTime),
      now: new Date(now),
      isSuppressed,
      remainingTime: isSuppressed ? suppressTime - now : 0
    });

    // 如果抑制时间已过，清除抑制标志
    if (!isSuppressed && suppressUntil) {
      uni.removeStorageSync('errorSuppressUntil');
      console.log('isErrorSuppressed: 抑制时间已过，清除抑制标志');
    }

    return isSuppressed;
  } catch (error) {
    console.error('isErrorSuppressed: 检查抑制状态失败', error);
    return false;
  }
}

/**
 * 清除错误抑制标志
 */
export function clearErrorSuppression() {
  try {
    uni.removeStorageSync('errorSuppressUntil');
    console.log('clearErrorSuppression: 已清除错误抑制标志');
  } catch (error) {
    console.error('clearErrorSuppression: 清除抑制标志失败', error);
  }
}

/**
 * 检查页面是否应该自动重新请求数据
 * @returns {boolean} 是否应该自动请求
 */
export function shouldAutoRefreshData() {
  const userRetry = isUserRetry();
  console.log('shouldAutoRefreshData: 检查是否应该自动刷新数据', {
    userRetry,
    result: userRetry
  });

  // 只有在用户主动重试时才自动刷新数据
  if (userRetry) {
    // 清除重试标志，防止重复使用
    clearUserRetryFlag();
    return true;
  }

  return false;
}

/**
 * 处理请求错误的统一入口
 * @param {Object} error - 错误对象
 * @param {Object} options - 处理选项
 */
export function handleRequestError(error, options = {}) {
  console.error('handleRequestError: 统一错误处理入口', error)

  const errorMessage = error?.message || error?.errMsg || '请求失败'
  console.log('handleRequestError: 错误信息', errorMessage)
  console.log('handleRequestError: 错误状态码', error?.statusCode)

  const shouldShow = shouldShowServerDownPage(error, errorMessage)
  console.log('handleRequestError: 是否应该显示错误页面', shouldShow)

  if (shouldShow) {
    console.log('handleRequestError: 准备显示ServerDownPage')
    showServerDownPage({
      errorMessage,
      showDetails: options.showDetails || false,
      retryFunction: options.retryFunction || null
    })
  } else {
    console.log('handleRequestError: 显示Toast提示')
    // 普通错误，使用toast提示
    uni.showToast({
      title: errorMessage,
      icon: 'none',
      duration: 2000
    })
  }
}

/**
 * 创建带有错误处理的请求包装器
 * @param {Function} requestFunction - 原始请求函数
 * @param {Object} options - 错误处理选项
 * @returns {Function} 包装后的请求函数
 */
export function createErrorHandledRequest(requestFunction, options = {}) {
  return async (...args) => {
    try {
      return await requestFunction(...args)
    } catch (error) {
      handleRequestError(error, {
        ...options,
        retryFunction: () => requestFunction(...args)
      })
      throw error
    }
  }
}

export default {
  ERROR_TYPES,
  shouldShowServerDownPage,
  showServerDownPage,
  showMaintenanceToast,
  resetErrorPageNavigationState,
  getErrorPageNavigationState,
  isUserRetry,
  clearUserRetryFlag,
  isErrorSuppressed,
  clearErrorSuppression,
  shouldAutoRefreshData,
  handleRequestError,
  createErrorHandledRequest
}