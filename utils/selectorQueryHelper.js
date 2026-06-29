/**
 * SelectorQuery 辅助工具
 * 解决微信小程序中 SelectorQuery 需要指定组件上下文的问题
 */

/**
 * 创建安全的 SelectorQuery 实例
 * @param {Object} component - Vue组件实例或页面实例
 * @returns {Object} SelectorQuery 实例
 */
export function createSafeSelectorQuery(component = null) {
  try {
    const query = uni.createSelectorQuery();
    
    // 如果提供了组件实例，使用 .in() 方法指定上下文
    if (component) {
      return query.in(component);
    }
    
    // 如果没有提供组件实例，尝试获取当前页面实例
    const pages = getCurrentPages();
    if (pages.length > 0) {
      const currentPage = pages[pages.length - 1];
      return query.in(currentPage);
    }
    
    // 如果都没有，返回默认的 query（可能会有警告）
    console.warn('SelectorQuery: 无法找到合适的组件上下文，可能会出现警告');
    return query;
  } catch (error) {
    console.error('创建 SelectorQuery 失败:', error);
    return null;
  }
}

/**
 * 安全地执行 SelectorQuery
 * @param {Object} component - Vue组件实例或页面实例
 * @param {string} selector - CSS选择器
 * @param {Function} callback - 回调函数
 * @param {Object} options - 选项
 */
export function safeSelectorQuery(component, selector, callback, options = {}) {
  try {
    const query = createSafeSelectorQuery(component);
    if (!query) {
      console.error('无法创建 SelectorQuery 实例');
      return;
    }
    
    // 根据选项配置查询
    let queryMethod = query.select(selector);
    
    if (options.boundingClientRect) {
      queryMethod = queryMethod.boundingClientRect();
    }
    
    if (options.scrollOffset) {
      queryMethod = queryMethod.scrollOffset();
    }
    
    if (options.fields) {
      queryMethod = queryMethod.fields(options.fields);
    }
    
    // 执行查询
    query.exec((res) => {
      if (typeof callback === 'function') {
        callback(res);
      }
    });
  } catch (error) {
    console.error('执行 SelectorQuery 失败:', error);
    if (typeof callback === 'function') {
      callback(null);
    }
  }
}

/**
 * 获取元素的边界信息
 * @param {Object} component - Vue组件实例或页面实例
 * @param {string} selector - CSS选择器
 * @returns {Promise} 返回边界信息的 Promise
 */
export function getBoundingClientRect(component, selector) {
  return new Promise((resolve, reject) => {
    safeSelectorQuery(component, selector, (res) => {
      if (res && res[0]) {
        resolve(res[0]);
      } else {
        reject(new Error(`未找到元素: ${selector}`));
      }
    }, { boundingClientRect: true });
  });
}

/**
 * 获取元素的滚动信息
 * @param {Object} component - Vue组件实例或页面实例
 * @param {string} selector - CSS选择器
 * @returns {Promise} 返回滚动信息的 Promise
 */
export function getScrollOffset(component, selector) {
  return new Promise((resolve, reject) => {
    safeSelectorQuery(component, selector, (res) => {
      if (res && res[0]) {
        resolve(res[0]);
      } else {
        reject(new Error(`未找到元素: ${selector}`));
      }
    }, { scrollOffset: true });
  });
}
