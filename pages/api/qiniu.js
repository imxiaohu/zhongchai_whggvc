/**
 * 七牛云相关API
 */

import { request } from '../../utils/request.js';

/**
 * 获取七牛云上传token
 * @returns {Promise<Object>} 上传token信息
 */
export async function getQiniuUploadToken() {
	try {
		const result = await request({
			url: '/api/upload/token', // 修正为实际存在的接口
			method: 'GET'
		});

		return result;
	} catch (error) {
		console.error('获取七牛云上传token失败:', error);
		throw error;
	}
}

/**
 * 获取七牛云配置信息
 * @returns {Promise<Object>} 配置信息
 */
export async function getQiniuConfig() {
	try {
		const result = await request({
			url: '/api/upload/stats', // 修正为实际存在的接口
			method: 'GET'
		});

		return result;
	} catch (error) {
		console.error('获取七牛云配置失败:', error);
		throw error;
	}
}

/**
 * 上传图片到七牛云
 * @param {string} filePath 本地文件路径
 * @param {Object} options 上传选项
 * @returns {Promise<string>} 上传后的URL
 */
export async function uploadImageToQiniu(filePath, options = {}) {
	try {
		console.log('开始上传图片到七牛云:', filePath);

		// 检查是否是H5环境的blob URL，如果是则使用统一的上传工具
		if (filePath.startsWith('blob:')) {
			console.log('检测到blob URL，使用统一上传工具');
			// 导入统一上传工具
			const { uploadSingleImage } = await import('../../utils/imageUpload.js');
			const result = await uploadSingleImage(filePath, {
				showLoading: false, // 外层已经显示loading
				compress: true,
				quality: 0.8
			});

			if (result.success) {
				return result.url;
			} else {
				throw new Error(result.error || '上传失败');
			}
		}

		// 获取上传token
		const tokenResult = await getQiniuUploadToken();
		console.log('获取token结果:', tokenResult);

		if (!tokenResult.success) {
			throw new Error(tokenResult.message || '获取上传token失败');
		}

		const { token, keyPrefix, domain, uploadUrl } = tokenResult.result;

		// 验证必要参数
		if (!token) {
			throw new Error('上传token为空');
		}
		if (!domain) {
			throw new Error('域名配置为空');
		}

		// 生成完整的文件key
		const timestamp = Date.now();
		const random = Math.random().toString(36).substring(2);

		// 处理文件扩展名，特别是blob URL的情况
		let ext = 'jpg'; // 默认扩展名
		if (filePath.startsWith('blob:')) {
			// 对于blob URL，使用默认扩展名
			ext = 'jpg';
		} else {
			// 从文件路径中提取扩展名
			const pathParts = filePath.split('.');
			if (pathParts.length > 1) {
				ext = pathParts.pop().toLowerCase();
				// 确保扩展名是有效的图片格式
				if (!['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg'].includes(ext)) {
					ext = 'jpg';
				}
			}
		}

		const key = `${keyPrefix || 'avatars/'}${timestamp}_${random}.${ext}`;

		console.log('上传参数:', {
			url: uploadUrl || 'https://up-z2.qiniup.com',
			key: key,
			token: token.substring(0, 20) + '...' // 只显示token前20位用于调试
		});

		// 上传文件
		const uploadResult = await new Promise((resolve, reject) => {
			uni.uploadFile({
				url: uploadUrl || 'https://up-z2.qiniup.com',
				filePath: filePath,
				name: 'file',
				formData: {
					token: token,
					key: key
				},
				success: (res) => {
					console.log('上传成功响应:', res);
					resolve(res);
				},
				fail: (err) => {
					console.error('上传失败:', err);
					reject(err);
				}
			});
		});

		console.log('上传结果状态码:', uploadResult.statusCode);
		console.log('上传结果数据:', uploadResult.data);

		if (uploadResult.statusCode === 200) {
			// 解析响应
			let response;
			try {
				response = JSON.parse(uploadResult.data);
			} catch (parseError) {
				console.error('解析响应数据失败:', parseError);
				throw new Error('服务器响应格式错误');
			}

			if (response.key) {
				// 构建完整URL
				let fullUrl;
				if (domain.startsWith('http://') || domain.startsWith('https://')) {
					fullUrl = `${domain}/${response.key}`;
				} else {
					fullUrl = `https://${domain}/${response.key}`;
				}
				console.log('上传成功，文件URL:', fullUrl);
				return fullUrl;
			} else {
				console.error('响应中缺少key字段:', response);
				throw new Error('上传响应格式错误：缺少文件key');
			}
		} else {
			const errorMsg = `上传失败，HTTP状态码: ${uploadResult.statusCode}`;
			console.error(errorMsg, uploadResult);
			throw new Error(errorMsg);
		}
	} catch (error) {
		console.error('七牛云上传失败:', error);
		// 提供更友好的错误信息
		if (error.message.includes('token')) {
			throw new Error('上传凭证获取失败，请重试');
		} else if (error.message.includes('400')) {
			throw new Error('上传参数错误，请检查文件格式');
		} else if (error.message.includes('401')) {
			throw new Error('上传权限验证失败');
		} else if (error.message.includes('413')) {
			throw new Error('文件大小超过限制（最大10MB）');
		} else {
			throw new Error(error.message || '上传失败，请重试');
		}
	}
}

/**
 * 批量上传图片到七牛云
 * @param {Array<string>} filePaths 文件路径数组
 * @param {Object} options 上传选项
 * @returns {Promise<Array<string>>} 上传后的URL数组
 */
export async function batchUploadToQiniu(filePaths, options = {}) {
	const {
		maxConcurrent = 3, // 最大并发数
		onProgress = null   // 进度回调
	} = options;
	
	const results = [];
	const errors = [];
	
	// 分批上传，控制并发数
	for (let i = 0; i < filePaths.length; i += maxConcurrent) {
		const batch = filePaths.slice(i, i + maxConcurrent);
		const batchPromises = batch.map(async (filePath) => {
			try {
				const url = await uploadImageToQiniu(filePath);
				results.push({ filePath, url, success: true });
				
				// 调用进度回调
				if (onProgress) {
					onProgress({
						completed: results.length,
						total: filePaths.length,
						current: filePath,
						url: url
					});
				}
				
				return url;
			} catch (error) {
				errors.push({ filePath, error: error.message });
				
				// 调用进度回调
				if (onProgress) {
					onProgress({
						completed: results.length,
						total: filePaths.length,
						current: filePath,
						error: error.message
					});
				}
				
				throw error;
			}
		});
		
		// 等待当前批次完成
		await Promise.allSettled(batchPromises);
	}
	
	return {
		success: results.length > 0,
		results: results,
		errors: errors,
		urls: results.map(r => r.url)
	};
}

/**
 * 删除七牛云文件
 * @param {string} key 文件key
 * @returns {Promise<Object>} 删除结果
 */
export async function deleteQiniuFile(key) {
	try {
		const result = await request({
			url: '/api/upload/file', // 修正为实际存在的接口
			method: 'DELETE',
			data: { key }
		});

		return result;
	} catch (error) {
		console.error('删除七牛云文件失败:', error);
		throw error;
	}
}

/**
 * 获取文件信息
 * @param {string} key 文件key
 * @returns {Promise<Object>} 文件信息
 */
export async function getQiniuFileInfo(key) {
	try {
		// 注意：后端暂未实现此接口，这里仅作为占位
		console.warn('getQiniuFileInfo接口暂未实现, key:', key);
		return {
			success: false,
			message: '接口暂未实现'
		};
	} catch (error) {
		console.error('获取文件信息失败:', error);
		throw error;
	}
}

export default {
	getQiniuUploadToken,
	getQiniuConfig,
	uploadImageToQiniu,
	batchUploadToQiniu,
	deleteQiniuFile,
	getQiniuFileInfo
};
