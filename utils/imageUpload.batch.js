// imageUpload - 批量上传与旧版兼容
// 职责：uploadImages 批量上传、uploadImagesLegacy、uploadSingleImage、删除图片、旧版兼容

import { getImageToken, uploadSingleFileToServer, uploadSingleBlobToServer, API_BASE_URL } from './imageUpload.core.js';

async function uploadBlobToQiniu(blobUrl) {
	// #ifdef H5
	const token = uni.getStorageSync('token') || uni.getStorageSync('accessToken');
	if (!token) throw new Error('未找到认证token，请重新登录');

	const uploadResult = await new Promise((resolve, reject) => {
		const timer = setTimeout(() => reject(new Error('上传超时')), 30000);

		uni.uploadFile({
			url: 'https://go.server.zhongchai.imxiaohu.cn' + '/api/upload/images',
			filePath: blobUrl,
			name: 'files',
			header: { 'X-Access-Token': token },
			success: (res) => { clearTimeout(timer); resolve(res); },
			fail: (err) => { clearTimeout(timer); reject(err); }
		});
	});

	if (uploadResult.statusCode === 200) {
		const result = JSON.parse(uploadResult.data);
		if (result.success) return result.result.url;
		throw new Error(result.message || '上传失败');
	}
	throw new Error(`上传失败，状态码: ${uploadResult.statusCode}`);
	// #endif
	throw new Error('H5 only');
}

async function uploadToGoServer(filePath) {
	return new Promise((resolve, reject) => {
		const token = uni.getStorageSync('token') || uni.getStorageSync('accessToken');
		if (!token) { reject(new Error('未找到认证token，请重新登录')); return; }

		uni.uploadFile({
			url: 'https://go.server.zhongchai.imxiaohu.cn' + '/api/upload/images',
			filePath,
			name: 'files',
			header: { 'X-Access-Token': token },
			success: (res) => {
				if (res.statusCode === 200) {
					try {
						const parsed = JSON.parse(res.data);
						if (parsed.success) {
							const payload = parsed.result;
							const url = payload?.url ||
								payload?.success?.[0]?.url ||
								payload?.success?.[0]?.uploadedUrl ||
								payload?.success?.[0]?.result?.url;
							if (url) resolve(url);
							else reject(new Error(parsed.message || '上传失败'));
						} else {
							reject(new Error(parsed.message || '上传失败'));
						}
					} catch (e) { reject(new Error('响应数据解析失败')); }
				} else {
					reject(new Error(`上传失败，状态码: ${res.statusCode}`));
				}
			},
			fail: (err) => reject(new Error(err.errMsg || '上传失败'))
		});
	});
}

function generateLocalUrl(filePath) {
	return filePath;
}

// ===== 批量上传（新版）=====

export async function uploadImages(files, token = null, onProgress = null) {
	if (!files || files.length === 0) return { success: true, urls: [], errors: [] };

	const uploadToken = token || await getImageToken();
	const uploadedUrls = [];
	const errors = [];
	let loadingShown = false;

	try {
		uni.showLoading({ title: `上传中 0/${files.length}` });
		loadingShown = true;

		for (let i = 0; i < files.length; i++) {
			const filePath = files[i];
			if (onProgress) onProgress(i + 1, files.length);
			uni.showLoading({ title: `上传中 ${i + 1}/${files.length}` });
			console.log(`[imageUpload] 开始上传第${i + 1}张图片:`, filePath);

			try {
				let uploadedUrl;
				const isBlob = filePath.startsWith('blob:');
				if (isBlob) uploadedUrl = await uploadSingleBlobToServer(filePath, i, uploadToken);
				else uploadedUrl = await uploadSingleFileToServer(filePath, i, uploadToken);

				uploadedUrls.push(uploadedUrl);
				console.log(`[imageUpload] 第${i + 1}张图片上传成功:`, uploadedUrl);

				if (i < files.length - 1) await new Promise(r => setTimeout(r, 500));
			} catch (error) {
				console.error(`[imageUpload] 第${i + 1}张图片上传失败:`, error);
				errors.push({ index: i, error: error.message });

				if (error.message.includes('file exists')) {
					console.log(`[imageUpload] 文件名冲突，尝试重新上传第${i + 1}张图片`);
					await new Promise(r => setTimeout(r, 1000));
					try {
						const isBlob = filePath.startsWith('blob:');
						const retryUrl = isBlob
							? await uploadSingleBlobToServer(filePath, i + Date.now(), uploadToken)
							: await uploadSingleFileToServer(filePath, i + Date.now(), uploadToken);
						uploadedUrls.push(retryUrl);
					} catch (retryError) {
						console.error(`[imageUpload] 第${i + 1}张图片重试失败:`, retryError);
					}
				}
			}
		}

		return { success: uploadedUrls.length > 0, urls: uploadedUrls, errors };
	} catch (error) {
		console.error('[imageUpload] 上传图片失败:', error);
		return { success: false, urls: [], errors: [{ index: -1, error: error.message }] };
	} finally {
		if (loadingShown) { try { uni.hideLoading(); } catch (e) { console.warn(e); } }
	}
}

// ===== 旧版批量上传（兼容）=====

export async function uploadImagesLegacy(filePaths, options = {}) {
	const { showLoading = true, compress = true, quality = 0.8, maxRetries = 3 } = options;
	let loadingShown = false;
	const results = [];
	const errors = [];

	try {
		if (showLoading) {
			uni.showLoading({ title: '上传中...', mask: true });
			loadingShown = true;
		}

		console.log('[imageUpload] 开始批量上传图片...');

		for (let i = 0; i < filePaths.length; i++) {
			const filePath = filePaths[i];
			let uploadSuccess = false;
			let retryCount = 0;

			let processedPath = filePath;
			if (compress) {
				try {
					const { compressImage } = await import('./imageUpload.core.js');
					processedPath = await compressImage(filePath, quality);
				} catch (e) { console.warn('[imageUpload] 图片压缩失败:', e); }
			}

			while (!uploadSuccess && retryCount < maxRetries) {
				try {
					let uploadedUrl;
					if (processedPath.startsWith('blob:')) {
						uploadedUrl = await uploadBlobToQiniu(processedPath);
					} else {
						uploadedUrl = await uploadToGoServer(processedPath);
					}

					results.push({
						originalPath: filePath,
						uploadedUrl,
						success: true,
						method: 'server',
						retries: retryCount
					});
					uploadSuccess = true;
				} catch (error) {
					retryCount++;
					const errorMsg = error.message || error.errMsg || '上传失败';
					console.error(`[imageUpload] 图片上传失败 (第${retryCount}/${maxRetries}次尝试):`, errorMsg);

					const isConnectionError = /CONNECTION_REFUSED|uploadFile:fail|网络错误|timeout/.test(errorMsg);

					if (retryCount >= maxRetries) {
						results.push({
							originalPath: filePath,
							uploadedUrl: generateLocalUrl(processedPath),
							success: true,
							isLocal: true,
							method: 'local_fallback',
							retries: retryCount
						});
						uploadSuccess = true;
					} else {
						await new Promise(r => setTimeout(r, Math.min(1000 * Math.pow(2, retryCount - 1), 5000)));
					}
				}
			}
		}

		const localCount = results.filter(r => r.isLocal).length;
		const serverCount = results.filter(r => r.method === 'server').length;

		if (results.length > 0) {
			if (localCount > 0 && serverCount === 0) {
				uni.showToast({ title: `${results.length}张图片已添加（本地模式）`, icon: 'success', duration: 2000 });
			} else if (localCount > 0) {
				uni.showToast({ title: `${results.length}张图片已处理（${localCount}张本地，${serverCount}张云端）`, icon: 'success', duration: 2500 });
			} else {
				uni.showToast({ title: `${results.length}张图片上传成功`, icon: 'success', duration: 1500 });
			}
		}

		return {
			success: results.length > 0,
			results,
			errors,
			urls: results.map(r => r.uploadedUrl),
			summary: {
				total: filePaths.length,
				success: results.length,
				failed: errors.length,
				local: localCount,
				cloud: serverCount
			}
		};
	} catch (error) {
		console.error('[imageUpload] 批量上传失败:', error);
		return { success: false, results: [], errors: [{ error: error.message || '未知错误', errorType: 'system' }], urls: [], summary: { total: filePaths.length, success: 0, failed: filePaths.length, local: 0, cloud: 0 } };
	} finally {
		if (loadingShown) { try { uni.hideLoading(); } catch (e) { console.warn(e); } }
	}
}

// ===== 单图上传兼容 =====

export async function uploadSingleImage(filePath, options = {}) {
	const result = await uploadImagesLegacy([filePath], options);
	return {
		success: result.success && result.urls.length > 0,
		url: result.urls[0] || null,
		error: result.errors[0]?.error || null
	};
}

// ===== 删除图片 =====

export async function deleteUploadedImage(imageUrl, token = null) {
	if (!imageUrl) return { success: false, message: '图片URL不能为空' };
	const uploadToken = token || await getImageToken();

	try {
		const result = await new Promise((resolve, reject) => {
			uni.request({
				url: API_BASE_URL + '/api/upload/delete',
				method: 'POST',
				data: { url: imageUrl },
				header: { 'X-Access-Token': uploadToken, 'Content-Type': 'application/json' },
				success: (res) => {
					if (res.statusCode === 200 && res.data.success) resolve({ success: true, message: '删除成功' });
					else reject(new Error(res.data.message || '删除失败'));
				},
				fail: (error) => reject(new Error('请求失败: ' + (error.errMsg || '未知错误')))
			});
		});
		return result;
	} catch (error) {
		console.error('[imageUpload] 删除图片失败:', error);
		return { success: false, message: error.message };
	}
}

// ===== 旧版兼容函数 =====

export function validateImageFile(filePath, maxSize = 10 * 1024 * 1024) {
	return new Promise((resolve, reject) => {
		uni.getImageInfo({
			src: filePath,
			success: (res) => {
				if (res.size && res.size > maxSize) {
					reject(new Error(`图片大小不能超过${Math.round(maxSize / 1024 / 1024)}MB`));
					return;
				}
				if (res.width > 4096 || res.height > 4096) {
					reject(new Error('图片尺寸过大，请选择较小的图片'));
					return;
				}
				resolve({ width: res.width, height: res.height, size: res.size, type: res.type });
			},
			fail: (error) => {
				console.warn('[imageUpload] 获取图片信息失败:', error);
				reject(new Error('图片格式不支持'));
			}
		});
	});
}

export function previewImages(current, urls = null) {
	console.log('[imageUpload] 预览图片被调用:', { current, urls: urls || [current] });
	try {
		const imageUrls = urls || [current];
		const validUrls = imageUrls.filter(url => url && typeof url === 'string' && url.trim() !== '');
		if (validUrls.length === 0) {
			uni.showToast({ title: '图片加载失败', icon: 'none' });
			return;
		}
		uni.previewImage({ current, urls: validUrls });
	} catch (error) {
		console.error('[imageUpload] 预览图片出错:', error);
		uni.showToast({ title: '图片预览出错', icon: 'none' });
	}
}
