import { request } from '../../utils/request.js';

/**
 * 获取评教列表
 * @returns {Promise} 返回评教列表数据
 */
export const getEvaluationList = () => {
  return request({
    url: '/scloud/educational/evaluation/getEvaluationList',
    method: 'GET'
  });
};

/**
 * 获取评教详情
 * @param {string} id 评教ID
 * @returns {Promise} 返回评教详情数据
 */
export const getEvaluationDetail = (id) => {
  if (!id || id === 'undefined' || id === 'null') {
    console.warn('[getEvaluationDetail] invalid id:', id);
    return Promise.reject(new Error('评教ID无效'));
  }
  return request({
    url: `/scloud/educational/evaluation/getEvaluationNorm/${id}`,
    method: 'GET'
  });
};

/**
 * 提交评教
 * @param {Object} data 评教数据
 * @returns {Promise} 返回提交结果
 */
export const submitEvaluation = (data) => {
  return request({
    // 后端代理学校新接口，路径为 /scloud/educational/evaluation/submit
    url: '/scloud/educational/evaluation/submit',
    method: 'POST',
    data
  });
}; 