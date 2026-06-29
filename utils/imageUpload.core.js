// imageUpload - 核心上传逻辑
// 职责：配置、Token 验证、图片处理工具、上传单个文件/Blob

import { isDevelopment } from '@/config/development.js';

const isDevEnv = isDevelopment();

const UPLOAD_CONFIG = {
	goServer: {
		url: 'https://go.server.zhongchai.imxiaohu.cn',
		enabled: true
	},
	local: { enabled: false }
};

const BASE_URL = 'https://go.server.zhongchai.imxiaohu.cn';
const isDevelopmentEnv = process.env.NODE_ENV === 'development' ||
	(typeof window !== 'undefined' && window.location.hostname === 'localhost');
const API_BASE_URL = isDevelopmentEnv ? 'http://localhost:2333' : BASE_URL;

console.log(`[imageUpload] 当前环境: ${isDevEnv ? '开发' : '生产'}`);
console.log('[imageUpload] API Base URL:', API_BASE_URL);

// ===== Token =====

export async function getImageToken() {
	let token = uni.getStorageSync('token') || uni.getStorageSync('accessToken');
	if (!token) throw new Error('未找到认证token，请重新登录');
	return token;
}

export function validateToken(token) {
	if (!token || typeof token !== 'string') return false;
	const parts = token.split('.');
	if (parts.length === 3) return true;
	if (token.startsWith('logged_in_')) return true;
	return true;
}

// ===== 图片验证 =====

export function validateImage(file, maxSizeMB = 10) {
	return new Promise((resolve, reject) => {
		if (!file) { reject(new Error('请选择图片')); return; }
		const maxSize = maxSizeMB * 1024 * 1024;
		if (file.size && file.size > maxSize) {
			reject(new Error(`图片大小不能超过${maxSizeMB}MB`));
			return;
		}
		const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'];
		if (file.type && !allowedTypes.includes(file.type)) {
			reject(new Error('不支持的图片格式'));
			return;
		}
		resolve({ size: file.size, type: file.type, name: file.name });
	});
}

// ===== 图片处理 =====

export async function compressImage(filePath, quality = 0.8) {
	try {
		if (typeof uni.compressImage === 'function') {
			const result = await new Promise((resolve, reject) => {
				uni.compressImage({ src: filePath, quality, success: resolve, fail: reject });
			});
			return result.tempFilePath || filePath;
		}
		return filePath;
	} catch (error) {
		console.warn('[imageUpload] 图片压缩失败，使用原图:', error);
		return filePath;
	}
}

export async function processBlobUrl(blobUrl) {
	if (!blobUrl.startsWith('blob:')) return blobUrl;
	return blobUrl;
}

export function generateImageUrl(file) {
	if (!file) return '';
	if (/^https?:\/\/|\/\//.test(file)) return file;
	if (file.startsWith('/tmp') || file.startsWith('file://') || file.startsWith('blob:')) return file;
	return file;
}

export function isSvgImage(imageUrl) {
	if (!imageUrl || typeof imageUrl !== 'string') return false;
	const url = imageUrl.toLowerCase();
	return url.includes('.svg') || url.includes('image/svg+xml') || url.includes('svg+xml');
}

export function getImageGridClass(count) {
	if (count === 1) return 'grid-1';
	if (count === 2) return 'grid-2';
	if (count <= 4) return 'grid-4';
	return 'grid-9';
}

// ===== 选择 & 预览 =====

export function chooseImages(count = 9, sourceType = ['album', 'camera']) {
	return new Promise((resolve, reject) => {
		uni.chooseImage({
			count: Math.min(count, 9),
			sizeType: ['compressed'],
			sourceType,
			success: (res) => resolve(res.tempFilePaths),
			fail: reject
		});
	});
}

export function previewImage(url, urls = []) {
	uni.previewImage({ current: url, urls: urls.length > 0 ? urls : [url] });
}

// ===== 上传单个文件 =====

function generateUniqueId(index) {
	return `${Date.now()}_${index}_${Math.random().toString(36).substring(2, 11)}`;
}

async function uploadSingleFileToServer(filePath, index, token) {
	return new Promise(async (resolve, reject) => {
		try {
			const uniqueId = generateUniqueId(index);
			console.log('[imageUpload] 上传文件，唯一标识:', uniqueId);

			uni.uploadFile({
				url: API_BASE_URL + '/api/upload/image',
				filePath,
				name: 'file',
				formData: { uniqueId },
				header: { 'X-Access-Token': token },
				success: (res) => {
					try {
						if (res.statusCode !== 200) {
							reject(new Error(`HTTP错误: ${res.statusCode}`));
							return;
						}
						const result = JSON.parse(res.data);
						if (result.success) {
							resolve(result.result.url);
						} else {
							reject(new Error(result.message || '上传失败'));
						}
					} catch (e) {
						reject(new Error('解析响应失败: ' + e.message));
					}
				},
				fail: (error) => reject(new Error('网络请求失败: ' + (error.errMsg || error.message || '未知错误')))
			});
		} catch (tokenError) {
			reject(new Error('获取认证token失败: ' + tokenError.message));
		}
	});
}

async function uploadSingleBlobToServer(blobUrl, index, token) {
	return new Promise(async (resolve, reject) => {
		try {
			const uniqueId = generateUniqueId(index);
			console.log('[imageUpload] 上传blob，唯一标识:', uniqueId);

			uni.uploadFile({
				url: API_BASE_URL + '/api/upload/image',
				filePath: blobUrl,
				name: 'file',
				formData: { uniqueId },
				header: { 'X-Access-Token': token },
				success: (res) => {
					console.log('[imageUpload] blob上传响应:', res);
					if (res.statusCode === 200) {
						try {
							const result = JSON.parse(res.data);
							if (result.success) resolve(result.result.url);
							else reject(new Error(result.message || '上传失败'));
						} catch (parseError) {
							reject(new Error('响应数据解析失败'));
						}
					} else {
						reject(new Error(`上传失败，状态码: ${res.statusCode}`));
					}
				},
				fail: reject
			});
		} catch (error) {
			reject(new Error('上传失败: ' + error.message));
		}
	});
}

// ===== 公开 API =====

export async function uploadImage(file, token = null) {
	const uploadToken = token || await getImageToken();
	const isBlob = typeof file === 'string' && file.startsWith('blob:');
	if (isBlob) return uploadSingleBlobToServer(file, 0, uploadToken);
	return uploadSingleFileToServer(file, 0, uploadToken);
}

export { API_BASE_URL, UPLOAD_CONFIG, generateUniqueId, uploadSingleFileToServer, uploadSingleBlobToServer };
