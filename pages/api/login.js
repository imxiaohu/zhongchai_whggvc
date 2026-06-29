import { request } from '../../utils/request.js';

/**
 * 初始化登录页面
 * @returns {Promise} 返回初始化结果，包含clientId
 */
export const initLogin = () => {
  return request({
    url: '/scloud/init',
    method: 'GET'
  }).then(result => {
    const existingClientId = uni.getStorageSync('clientId');
    if (existingClientId) {
      return result;
    }
    
    let clientId = null;
    
    if (result) {
      if (result.clientId) {
        clientId = result.clientId;
      } else if (result.result?.clientId) {
        clientId = result.result.clientId;
      }
      
      if (clientId) {
        uni.setStorageSync('clientId', clientId);
        result.clientId = clientId;
      }
    }
    
    return result;
  });
};

/**
 * 获取验证码
 * @param {string} queryParam 验证码查询参数
 * @returns {Promise} 返回验证码图片的arraybuffer
 */
export const getValidateCode = (queryParam) => {
  const clientId = uni.getStorageSync('clientId');
  if (!clientId) {
    return Promise.reject(new Error('请先初始化登录页面'));
  }

  return request({
    url: `/scloud/validateCode${queryParam ? `?${queryParam}` : ''}`,
    method: 'GET',
    responseType: 'arraybuffer',
    header: {
      'x-client-id': clientId
    }
  });
};

/**
 * 登录
 * @param {Object} data 登录参数
 * @param {string} data.username 用户名
 * @param {string} data.password 密码
 * @param {string} data.randomcode 验证码
 * @returns {Promise} 返回登录结果
 */
export const login = (data) => {
  const clientId = uni.getStorageSync('clientId');
  if (!clientId) {
    return Promise.reject(new Error('请先初始化登录页面'));
  }

  const payload = {
    username: data.username,
    password: data.password,
    captcha: data.captcha || data.randomcode || ''
  }

  return request({
    url: '/scloud/login',
    method: 'POST',
    data: payload,
    header: {
      'x-client-id': clientId
    }
  });
};

/**
 * 移动端登录
 * @param {Object} data 登录参数
 * @param {string} data.username 用户名
 * @param {string} data.password 密码
 * @returns {Promise} 返回登录结果，包含token信息
 */
export const mLogin = (data) => {
  return request({
    url: '/api/m/sys/mLogin',
    method: 'POST',
    data: {
      username: data.username,
      password: data.password
    },
    header: {
      'content-type': 'application/json'
    }
  }).then(result => {
    if (result && result.success && result.result && result.result.token) {
      uni.setStorageSync('token', result.result.token);
      uni.setStorageSync('loginType', 'school');
      if (data.rememberPassword) {
        uni.setStorageSync('saved_username', data.username);
        uni.setStorageSync('saved_password', data.password);
        uni.setStorageSync('remember_password', true);
      }
    }
    return result;
  });
};
