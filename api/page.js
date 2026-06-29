/**
 * 页面导航相关的方法
 */

/**
 * 显示提示信息
 * @param {Object} options 配置项
 * @param {string} options.title 提示内容
 * @param {string} [options.icon='none'] 图标类型
 * @param {number} [options.duration=1500] 显示时长
 * @returns {Promise} 返回Promise
 */
export const showToast = ({ title, icon = 'none', duration = 1500 }) => {
  return new Promise((resolve) => {
    uni.showToast({
      title,
      icon,
      duration
    });
    setTimeout(resolve, duration);
  });
};

/**
 * 显示加载提示
 * @param {Object} options 配置项
 * @param {string} [options.title='加载中...'] 提示内容
 * @returns {Promise} 返回Promise
 */
export const showLoading = ({ title = '加载中...' } = {}) => {
  return uni.showLoading({
    title,
    mask: true
  });
};

/**
 * 隐藏加载提示
 */
export const hideLoading = () => {
  uni.hideLoading();
};

/**
 * 显示模态对话框
 * @param {Object} options 配置项
 * @param {string} options.title 标题
 * @param {string} [options.content] 内容
 * @param {boolean} [options.showCancel=true] 是否显示取消按钮
 * @param {string} [options.cancelText='取消'] 取消按钮文字
 * @param {string} [options.confirmText='确定'] 确认按钮文字
 * @returns {Promise} 返回Promise
 */
export const showModal = ({ title, content = '', showCancel = true, cancelText = '取消', confirmText = '确定' }) => {
  return uni.showModal({
    title,
    content,
    showCancel,
    cancelText,
    confirmText
  });
};

/**
 * 重新加载页面
 * @param {Object} options 配置项
 * @param {string} options.url 页面路径
 * @param {number} [options.delay=0] 延迟时间
 * @returns {Promise} 返回Promise
 */
export const relaunch = ({ url, delay = 0 }) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      uni.reLaunch({ url });
      resolve();
    }, delay);
  });
};

/**
 * 导航到页面
 * @param {Object} options 配置项
 * @param {string} options.url 页面路径
 * @returns {Promise} 返回Promise
 */
export const navigateTo = ({ url }) => {
  return uni.navigateTo({ url });
};

/**
 * 返回上一页
 * @param {Object} options 配置项
 * @param {number} [options.delta=1] 返回的页面数
 * @returns {Promise} 返回Promise
 */
export const navigateBack = ({ delta = 1 } = {}) => {
  return uni.navigateBack({ delta });
};

/**
 * 切换选项卡
 * @param {Object} options 配置项
 * @param {string} options.url 页面路径
 * @returns {Promise} 返回Promise
 */
export const switchTab = ({ url }) => {
  return uni.switchTab({ url });
};

/**
 * 清除本地存储
 */
export const clearStorage = () => {
  uni.clearStorageSync();
};