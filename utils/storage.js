/**
 * 本地存储工具类
 * 提供统一的本地存储管理功能
 */

/**
 * 获取本地存储数据
 * @param {string} key - 存储键名
 * @param {any} defaultValue - 默认值
 * @returns {any} 存储的数据或默认值
 */
export function getStorage(key, defaultValue = null) {
  try {
    const value = uni.getStorageSync(key);
    return value !== '' ? value : defaultValue;
  } catch (error) {
    console.error('获取本地存储失败:', error);
    return defaultValue;
  }
}

/**
 * 设置本地存储数据
 * @param {string} key - 存储键名
 * @param {any} value - 要存储的数据
 * @returns {boolean} 是否存储成功
 */
export function setStorage(key, value) {
  try {
    uni.setStorageSync(key, value);
    return true;
  } catch (error) {
    console.error('设置本地存储失败:', error);
    return false;
  }
}

/**
 * 删除指定的本地存储数据
 * @param {string} key - 存储键名
 * @returns {boolean} 是否删除成功
 */
export function removeStorage(key) {
  try {
    uni.removeStorageSync(key);
    return true;
  } catch (error) {
    console.error('删除本地存储失败:', error);
    return false;
  }
}

/**
 * 清空所有本地存储数据
 * @returns {boolean} 是否清空成功
 */
export function clearStorage() {
  try {
    uni.clearStorageSync();
    return true;
  } catch (error) {
    console.error('清空本地存储失败:', error);
    return false;
  }
}

/**
 * 获取本地存储信息
 * @returns {object} 存储信息对象
 */
export function getStorageInfo() {
  try {
    return uni.getStorageInfoSync();
  } catch (error) {
    console.error('获取存储信息失败:', error);
    return {
      keys: [],
      currentSize: 0,
      limitSize: 0
    };
  }
}

/**
 * 检查指定键是否存在
 * @param {string} key - 存储键名
 * @returns {boolean} 是否存在
 */
export function hasStorage(key) {
  try {
    const info = getStorageInfo();
    return info.keys.includes(key);
  } catch (error) {
    console.error('检查存储键失败:', error);
    return false;
  }
}

/**
 * 批量设置存储数据
 * @param {object} data - 要存储的数据对象
 * @returns {boolean} 是否全部设置成功
 */
export function setBatchStorage(data) {
  try {
    for (const [key, value] of Object.entries(data)) {
      uni.setStorageSync(key, value);
    }
    return true;
  } catch (error) {
    console.error('批量设置存储失败:', error);
    return false;
  }
}

/**
 * 批量获取存储数据
 * @param {array} keys - 要获取的键名数组
 * @returns {object} 包含所有键值对的对象
 */
export function getBatchStorage(keys) {
  const result = {};
  try {
    keys.forEach(key => {
      result[key] = uni.getStorageSync(key);
    });
    return result;
  } catch (error) {
    console.error('批量获取存储失败:', error);
    return result;
  }
}

// 默认导出所有方法
export default {
  getStorage,
  setStorage,
  removeStorage,
  clearStorage,
  getStorageInfo,
  hasStorage,
  setBatchStorage,
  getBatchStorage
};