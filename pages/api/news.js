import { request } from '../../utils/request.js';

// 本地缓存管理
const cache = {
  newsTypes: null,
  newsTypesExpiry: 0,
  topNews: null,
  topNewsExpiry: 0,
  CACHE_DURATION: 5 * 60 * 1000, // 5分钟缓存

  // 获取缓存的新闻类型
  getCachedNewsTypes() {
    if (this.newsTypes && Date.now() < this.newsTypesExpiry) {
      return this.newsTypes;
    }
    return null;
  },

  // 设置新闻类型缓存
  setCachedNewsTypes(data) {
    this.newsTypes = data;
    this.newsTypesExpiry = Date.now() + this.CACHE_DURATION;
  },

  // 获取缓存的置顶新闻
  getCachedTopNews() {
    if (this.topNews && Date.now() < this.topNewsExpiry) {
      return this.topNews;
    }
    return null;
  },

  // 设置置顶新闻缓存
  setCachedTopNews(data) {
    this.topNews = data;
    this.topNewsExpiry = Date.now() + this.CACHE_DURATION;
  },

  // 清除所有缓存
  clearAll() {
    this.newsTypes = null;
    this.newsTypesExpiry = 0;
    this.topNews = null;
    this.topNewsExpiry = 0;
  }
};

/**
 * 获取新闻类型列表
 * @returns {Promise<Array>} 返回新闻类型列表
 */
export const getNewsTypeList = async () => {
  try {
    // 先检查缓存
    const cachedTypes = cache.getCachedNewsTypes();
    if (cachedTypes) {
      console.log('getNewsTypeList: 使用缓存数据');
			return { data: cachedTypes, meta: { fromCache: true, cacheUpdatedAt: '', dataSourceType: 'database' } };
    }

    // 检查登录状态
    const token = uni.getStorageSync('token');

    if (!token) {
      throw new Error('未登录，请先登录');
    }

    console.log('getNewsTypeList: 从服务器获取数据');
    // 使用本地后端代理接口，避免CORS问题
    const res = await request({
      url: '/scloudoa/scs/news/eoaNewsType/getEoaNewsTypeList',
      method: 'GET',
      data: {
        pageNo: 1,
        pageSize: -1
      },
      timeout: 15000 // 增加超时时间到15秒
    });

    let result = [];
		const meta = {
			fromCache: !!(res && (res.fromCache || res.dataSourceType === 'database')),
			cacheUpdatedAt: (res && res.cacheUpdatedAt) || '',
			dataSourceType: (res && res.dataSourceType) || ''
		};

    // 添加详细的响应结构日志
    console.log('getNewsTypeList: API响应结构:', {
      hasRes: !!res,
      resType: typeof res,
      hasResult: !!(res && res.result),
      resultType: res && res.result ? typeof res.result : null,
      hasData: !!(res && res.data),
      dataType: res && res.data ? typeof res.data : null,
      isArray: Array.isArray(res),
      keys: res ? Object.keys(res) : [],
      isEmpty: res === '' || res === null || res === undefined,
      resLength: typeof res === 'string' ? res.length : null
    });

    // 处理空响应或错误响应
    if (!res || res === '' || res === null || res === undefined) {
      console.warn('getNewsTypeList: 收到空响应或无效响应');
      return { data: [], meta: { fromCache: false, cacheUpdatedAt: '', dataSourceType: '' } };
    }

    // 处理标准的学校服务器响应格式
    if (res && res.result && Array.isArray(res.result)) {
      result = res.result;
      console.log('getNewsTypeList: 使用 res.result 格式，数量:', result.length);
    } else if (res && res.data && Array.isArray(res.data)) {
      result = res.data;
      console.log('getNewsTypeList: 使用 res.data 格式，数量:', result.length);
    } else if (res && Array.isArray(res)) {
      result = res;
      console.log('getNewsTypeList: 使用直接数组格式，数量:', result.length);
    } else if (res && res.result && res.result.result && Array.isArray(res.result.result)) {
      // 处理嵌套的 result.result 格式
      result = res.result.result;
      console.log('getNewsTypeList: 使用 res.result.result 格式，数量:', result.length);
    } else if (res && res.success && res.result && Array.isArray(res.result)) {
      // 处理带 success 标志的格式
      result = res.result;
      console.log('getNewsTypeList: 使用带 success 的 res.result 格式，数量:', result.length);
    } else {
      console.warn('getNewsTypeList: 无法识别的新闻类型数据结构:', res);
      console.warn('getNewsTypeList: 响应详情:', JSON.stringify(res, null, 2));
      result = [];
    }

    // 缓存结果
    if (result.length > 0) {
      cache.setCachedNewsTypes(result);
    }

		return { data: result, meta };
  } catch (error) {
    console.error('getNewsTypeList: 获取新闻类型列表失败:', error);
    console.error('getNewsTypeList: 错误详情:', {
      message: error.message,
      stack: error.stack,
      url: '/scloudoa/scs/news/eoaNewsType/getEoaNewsTypeList'
    });

    // 如果是网络错误、服务器错误或CORS错误，返回空数组而不是抛出错误
    // 这样可以避免影响首页的正常加载
    if (error.message && (
      error.message.includes('网络') ||
      error.message.includes('timeout') ||
      error.message.includes('服务器') ||
      error.message.includes('500') ||
      error.message.includes('502') ||
      error.message.includes('503') ||
      error.message.includes('CORS') ||
      error.message.includes('超时') ||
      error.statusCode === 408
    )) {
      console.warn('getNewsTypeList: 网络、服务器或CORS错误，返回空数组');
			return { data: [], meta: { fromCache: false, cacheUpdatedAt: '', dataSourceType: '' } };
    }

    // 对于其他类型的错误，也返回空数组，确保应用不会崩溃
    console.warn('getNewsTypeList: 未知错误，返回空数组以保证应用稳定性');
		return { data: [], meta: { fromCache: false, cacheUpdatedAt: '', dataSourceType: '' } };
  }
};

/**
 * 根据类型ID获取新闻列表
 * @param {Object} options 查询参数
 * @param {string|number} options.typeId 新闻类型ID
 * @param {number} options.pageNo 页码，默认为1
 * @param {number} options.pageSize 每页数量，默认为6
 * @returns {Promise<Object>} 返回新闻列表数据
 */
export const getNewsListByTypeId = async (options = {}) => {
  try {
    // 检查登录状态
    const token = uni.getStorageSync('token');

    if (!token) {
      throw new Error('未登录，请先登录');
    }

    const { typeId, pageNo = 1, pageSize = 6 } = options;

    // 使用本地后端代理接口，避免CORS问题
    const res = await request({
      url: '/scloudoa/scs/news/eoaNews/getEoaNewsListByTypeId',
      method: 'GET',
      data: {
        pageNo,
        pageSize,
        ids: typeId
      },
      timeout: 15000 // 增加超时时间到15秒
    });

    // 添加详细的响应结构日志
    console.log('getNewsListByTypeId: API响应结构:', {
      hasRes: !!res,
      resType: typeof res,
      hasResult: !!(res && res.result),
      resultType: res && res.result ? typeof res.result : null,
      hasData: !!(res && res.data),
      dataType: res && res.data ? typeof res.data : null,
      isArray: Array.isArray(res),
      keys: res ? Object.keys(res) : [],
      isEmpty: res === '' || res === null || res === undefined,
      resLength: typeof res === 'string' ? res.length : null
    });

    // 处理空响应或错误响应
    if (!res || res === '' || res === null || res === undefined) {
      console.warn('getNewsListByTypeId: 收到空响应或无效响应');
      return {
        records: [],
        total: 0,
        pageNo,
        pageSize,
			meta: { fromCache: false, cacheUpdatedAt: '', dataSourceType: '' }
      };
    }

		const meta = {
			fromCache: !!(res && (res.fromCache || res.dataSourceType === 'database')),
			cacheUpdatedAt: (res && res.cacheUpdatedAt) || '',
			dataSourceType: (res && res.dataSourceType) || ''
		};

    // 处理标准的学校服务器响应格式
    if (res && res.result && res.result.records && Array.isArray(res.result.records)) {
      console.log('getNewsListByTypeId: 使用 res.result.records 格式，数量:', res.result.records.length);
      return {
        records: res.result.records,
        total: res.result.total || 0,
        pageNo: res.result.current || pageNo,
			pageSize: res.result.size || pageSize,
			meta
      };
    } else if (res && res.data && res.data.records && Array.isArray(res.data.records)) {
      console.log('getNewsListByTypeId: 使用 res.data.records 格式，数量:', res.data.records.length);
      return {
        records: res.data.records,
        total: res.data.total || 0,
        pageNo: res.data.current || pageNo,
			pageSize: res.data.size || pageSize,
			meta
      };
    } else if (res && Array.isArray(res)) {
      console.log('getNewsListByTypeId: 使用直接数组格式，数量:', res.length);
      return {
        records: res,
        total: res.length,
        pageNo,
			pageSize,
			meta
      };
    } else if (res && res.result && res.result.result && Array.isArray(res.result.result)) {
      // 处理嵌套的 result.result 格式
      console.log('getNewsListByTypeId: 使用 res.result.result 格式，数量:', res.result.result.length);
      return {
        records: res.result.result,
        total: res.result.result.length,
        pageNo,
			pageSize,
			meta
      };
    } else if (res && res.success && res.result && Array.isArray(res.result)) {
      // 处理带 success 标志的直接数组格式
      console.log('getNewsListByTypeId: 使用带 success 的 res.result 格式，数量:', res.result.length);
      return {
        records: res.result,
        total: res.result.length,
        pageNo,
			pageSize,
			meta
      };
    } else if (res && res.result && Array.isArray(res.result)) {
      // 处理 result 直接是数组的格式
      console.log('getNewsListByTypeId: 使用 res.result 数组格式，数量:', res.result.length);
      return {
        records: res.result,
        total: res.result.length,
        pageNo,
			pageSize,
			meta
      };
    } else {
      console.warn('getNewsListByTypeId: 无法识别的新闻列表数据结构:', res);
      console.warn('getNewsListByTypeId: 响应详情:', JSON.stringify(res, null, 2));
      return {
        records: [],
        total: 0,
        pageNo,
			pageSize,
			meta
      };
    }
  } catch (error) {
    console.error('getNewsListByTypeId: 获取新闻列表失败:', error);
    console.error('getNewsListByTypeId: 错误详情:', {
      message: error.message,
      stack: error.stack,
      typeId: options.typeId,
      url: '/scloudoa/scs/news/eoaNews/getEoaNewsListByTypeId'
    });

    // 如果是网络错误、服务器错误或CORS错误，返回空结果而不是抛出错误
    if (error.message && (
      error.message.includes('网络') ||
      error.message.includes('timeout') ||
      error.message.includes('服务器') ||
      error.message.includes('500') ||
      error.message.includes('502') ||
      error.message.includes('503') ||
      error.message.includes('CORS') ||
      error.message.includes('超时') ||
      error.statusCode === 408
    )) {
      console.warn('getNewsListByTypeId: 网络、服务器或CORS错误，返回空结果');
      return {
        records: [],
        total: 0,
        pageNo: options.pageNo || 1,
			pageSize: options.pageSize || 6,
			meta: { fromCache: false, cacheUpdatedAt: '', dataSourceType: '' }
      };
    }

    // 对于其他类型的错误，也返回空结果，确保应用不会崩溃
    console.warn('getNewsListByTypeId: 未知错误，返回空结果以保证应用稳定性');
    return {
      records: [],
      total: 0,
      pageNo: options.pageNo || 1,
		pageSize: options.pageSize || 6,
		meta: { fromCache: false, cacheUpdatedAt: '', dataSourceType: '' }
    };
  }
};

/**
 * 获取置顶新闻列表
 * @param {number} limit 获取数量限制，默认为5
 * @param {boolean} forceRefresh 是否强制刷新，默认为false
 * @returns {Promise<Array>} 返回置顶新闻列表
 */
export const getTopNews = async (limit = 5, forceRefresh = false) => {
  try {
    // 先检查缓存（除非强制刷新）
    if (!forceRefresh) {
      const cachedNews = cache.getCachedTopNews();
      if (cachedNews && cachedNews.length > 0) {
        console.log('getTopNews: 使用缓存数据');
        return cachedNews.slice(0, limit);
      }
    }

    // 检查登录状态
    const token = uni.getStorageSync('token');

    if (!token) {
      throw new Error('未登录，请先登录');
    }

    console.log('getTopNews: 从服务器获取数据');

    // 直接使用通知公告的类型ID（18），避免额外的类型查询请求
    // 这样可以减少网络请求次数，提高加载速度
    let typeId = 18; // 通知公告的默认ID
    let result = [];

    try {
      // 优先尝试直接获取通知公告
      console.log('getTopNews: 尝试直接获取通知公告，typeId:', typeId);
      const newsData = await getNewsListByTypeId({
        typeId,
        pageNo: 1,
        pageSize: limit
      });

      result = newsData.records || [];
      console.log('getTopNews: 直接获取结果数量:', result.length);

      if (result.length > 0) {
        // 缓存结果
        cache.setCachedTopNews(result);
        console.log('getTopNews: 直接获取成功，已缓存');
        return result;
      } else {
        console.log('getTopNews: 直接获取返回空结果，尝试降级方案');
      }
    } catch (directError) {
      console.warn('getTopNews: 直接获取通知公告失败，尝试查询类型列表:', directError.message);
      console.warn('getTopNews: 直接获取错误详情:', directError);
    }

    // 如果直接获取失败，再尝试查询类型列表（降级方案）
    try {
      console.log('getTopNews: 开始降级方案，查询新闻类型列表');
      const newsTypeResult = await getNewsTypeList();
      const newsTypes = newsTypeResult.data || [];
      console.log('getTopNews: 获取到新闻类型数量:', newsTypes ? newsTypes.length : 0);

      if (!newsTypes || newsTypes.length === 0) {
        console.warn('getTopNews: 没有找到新闻类型');
        return [];
      }

      // 打印所有新闻类型，便于调试
      console.log('getTopNews: 可用的新闻类型:', newsTypes.map(type => ({
        id: type.id,
        name: type.name
      })));

      // 获取通知公告类型的新闻
      const noticeType = newsTypes.find(type =>
        type.name && (type.name.includes('通知') || type.name.includes('公告'))
      );

      typeId = noticeType ? noticeType.id : newsTypes[0].id;
      console.log('getTopNews: 选择的类型ID:', typeId, '类型名称:', noticeType ? noticeType.name : newsTypes[0].name);

      // 获取该类型的新闻列表
      const newsData = await getNewsListByTypeId({
        typeId,
        pageNo: 1,
        pageSize: limit
      });

      result = newsData.records || [];
      console.log('getTopNews: 降级方案获取结果数量:', result.length);

      // 缓存结果
      if (result.length > 0) {
        cache.setCachedTopNews(result);
        console.log('getTopNews: 降级方案成功，已缓存');
      }

      return result;
    } catch (fallbackError) {
      console.error('getTopNews: 降级方案也失败:', fallbackError);
      console.error('getTopNews: 降级方案错误详情:', {
        message: fallbackError.message,
        stack: fallbackError.stack
      });
      return [];
    }

  } catch (error) {
    console.error('getTopNews: 获取置顶新闻失败:', error);
    console.error('getTopNews: 错误详情:', {
      message: error.message,
      stack: error.stack
    });

    // 如果是登录相关错误，抛出错误让首页处理
    if (error.message && (error.message.includes('未登录') || error.message.includes('登录'))) {
      throw error;
    }

    // 其他错误返回空数组，避免影响首页加载
    return [];
  }
};

/**
 * 获取新闻详情
 * @param {string|number} newsId 新闻ID
 * @param {string|number} typeId 新闻类型ID（可选，用于优化查询）
 * @returns {Promise<Object>} 返回新闻详情
 */
export const getNewsDetail = async (newsId, typeId = null) => {
  try {
    // 检查登录状态
    const token = uni.getStorageSync('token');

    if (!token) {
      throw new Error('未登录，请先登录');
    }

    // 由于学校服务器的新闻列表接口已经包含完整内容，
    // 我们通过获取新闻列表来找到对应的新闻详情

    // 如果没有提供类型ID，先获取所有新闻类型
    let searchTypeIds = [];
    if (typeId) {
      searchTypeIds = [typeId];
    } else {
      const newsTypeResult = await getNewsTypeList();
      const newsTypes = newsTypeResult.data || [];
      searchTypeIds = newsTypes.map(type => type.id);
    }

    // 在各个类型中搜索该新闻
    for (const currentTypeId of searchTypeIds) {
      try {
        // 获取该类型的新闻列表（获取更多数据以确保找到目标新闻）
        const newsData = await getNewsListByTypeId({
          typeId: currentTypeId,
          pageNo: 1,
          pageSize: 50 // 增加页面大小以提高找到目标新闻的概率
        });

        // 在列表中查找目标新闻
        const targetNews = newsData.records.find(news =>
          news.id == newsId || news.id === newsId
        );

        if (targetNews) {
          // 处理新闻数据格式，确保包含所有必要字段
          const newsDetail = {
            id: targetNews.id,
            title: targetNews.title,
            content: targetNews.content,
            publishTime: targetNews.pubdate || targetNews.createtime,
            createTime: targetNews.createtime,
            source: targetNews.departname || targetNews.author,
            author: targetNews.creatorid_dictText || targetNews.author,
            viewCount: targetNews.readcount,
            summary: extractSummary(targetNews.content),
            attachments: targetNews.eoaNewsAttachments || [],
            isTop: targetNews.isTop || false,
            type: targetNews.eoanewstypename,
            typeId: targetNews.eoanewstypeid
          };

          return newsDetail;
        }
      } catch (error) {
        console.warn('getNewsDetail: 在类型', currentTypeId, '中搜索失败:', error.message);
        // 继续搜索下一个类型
      }
    }

    // 如果在所有类型中都没找到，抛出错误
    throw new Error('新闻不存在或已被删除');

  } catch (error) {
    console.error('getNewsDetail: 获取新闻详情失败:', error);
    throw error;
  }
};

/**
 * 从HTML内容中提取摘要
 * @param {string} content HTML内容
 * @returns {string} 摘要文本
 */
function extractSummary(content) {
  if (!content) return '';

  try {
    // 移除HTML标签，提取纯文本
    const textContent = content
      .replace(/<[^>]*>/g, '') // 移除HTML标签
      .replace(/&nbsp;/g, ' ') // 替换HTML实体
      .replace(/&lt;/g, '<')
      .replace(/&gt;/g, '>')
      .replace(/&amp;/g, '&')
      .replace(/\s+/g, ' ') // 合并多个空白字符
      .trim();

    // 截取前150个字符作为摘要
    return textContent.length > 150 ? textContent.substring(0, 150) + '...' : textContent;
  } catch (error) {
    console.error('提取摘要失败:', error);
    return '';
  }
}

/**
 * 搜索新闻
 * @param {Object} options 搜索参数
 * @param {string} options.keyword 搜索关键词
 * @param {string|number} options.categoryId 分类ID（可选）
 * @param {number} options.pageNo 页码，默认为1
 * @param {number} options.pageSize 每页数量，默认为15
 * @returns {Promise<Object>} 返回搜索结果
 */
export const searchNews = async (options = {}) => {
  try {
    // 检查登录状态
    const token = uni.getStorageSync('token');
    
    if (!token) {
      throw new Error('未登录，请先登录');
    }
    
    const { keyword, categoryId, pageNo = 1, pageSize = 15 } = options;
    
    if (!keyword || keyword.trim() === '') {
      throw new Error('搜索关键词不能为空');
    }
    
    // 使用本地后端接口
    const res = await request({
      url: '/api/news/search/',
      method: 'GET',
      data: {
        keyword: keyword.trim(),
        category_id: categoryId,
        page: pageNo,
        size: pageSize
      }
    });
    
    if (res && res.data) {
      return {
        records: res.data.records || [],
        total: res.data.total || 0,
        pageNo: res.data.current || pageNo,
        pageSize: res.data.size || pageSize
      };
    } else {
      return {
        records: [],
        total: 0,
        pageNo,
        pageSize
      };
    }
  } catch (error) {
    console.error('搜索新闻失败:', error);
    throw error;
  }
};

/**
 * 下载附件（通过代理）
 * @param {string} attachmentUrl 附件URL
 * @param {string} fileName 文件名
 * @returns {Promise<Blob>} 返回文件数据
 */
export const downloadAttachment = async (attachmentUrl, fileName) => {
  try {
    // 检查登录状态
    const token = uni.getStorageSync('token');

    if (!token) {
      throw new Error('未登录，请先登录');
    }

    if (!attachmentUrl) {
      throw new Error('附件链接无效');
    }

    // 使用本地后端代理下载
    const response = await request({
      url: '/api/news/download-attachment',
      method: 'POST',
      data: {
        attachmentUrl,
        fileName: fileName || 'attachment'
      },
      responseType: 'blob' // 期望返回二进制数据
    });

    if (response && response.data) {
      return response.data;
    } else {
      throw new Error('下载失败，服务器无响应');
    }

  } catch (error) {
    console.error('downloadAttachment: 下载附件失败:', error);
    throw error;
  }
};

/**
 * 清除新闻相关的本地缓存
 * @description 在用户登出或需要强制刷新数据时调用
 */
export const clearNewsCache = () => {
  console.log('clearNewsCache: 清除新闻缓存');
  cache.clearAll();
};

/**
 * 预加载新闻数据
 * @description 在应用启动时预加载常用数据
 */
export const preloadNewsData = async () => {
  try {
    console.log('preloadNewsData: 开始预加载新闻数据');

    // 预加载新闻类型列表
    await getNewsTypeList();

    // 预加载置顶新闻
    await getTopNews(5);

    console.log('preloadNewsData: 预加载完成');
  } catch (error) {
    console.warn('preloadNewsData: 预加载失败，但不影响正常使用:', error.message);
  }
};
