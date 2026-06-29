/**
 * 图片裁切工具
 * 支持微信小程序和H5环境
 */

/**
 * 版本号比较
 * @param {string} v1 版本号1
 * @param {string} v2 版本号2
 * @returns {number} 1: v1>v2, 0: v1=v2, -1: v1<v2
 */
function compareVersion(v1, v2) {
	const arr1 = v1.split('.');
	const arr2 = v2.split('.');
	const maxLength = Math.max(arr1.length, arr2.length);

	for (let i = 0; i < maxLength; i++) {
		const num1 = parseInt(arr1[i] || '0');
		const num2 = parseInt(arr2[i] || '0');

		if (num1 > num2) return 1;
		if (num1 < num2) return -1;
	}

	return 0;
}

/**
 * 裁切图片为正方形
 * @param {string} imagePath - 图片路径
 * @param {Object} options - 裁切选项
 * @returns {Promise<string>} 裁切后的图片路径
 */
export function cropImageToSquare(imagePath, options = {}) {
	return new Promise((resolve, reject) => {
		console.log('cropImageToSquare 调用:', { imagePath, options });

		const {
			size = 400,
			quality = 0.8
		} = options;

		// 参数验证
		if (!imagePath) {
			reject(new Error('图片路径不能为空'));
			return;
		}

		// #ifdef MP-WEIXIN
		// 微信小程序环境
		cropImageWechat(imagePath, size, quality, resolve, reject);
		// #endif

		// #ifdef H5
		// H5环境
		cropImageH5(imagePath, size, quality, resolve, reject);
		// #endif

		// #ifndef MP-WEIXIN || H5
		// 其他环境，尝试使用uni.compressImage作为降级方案
		console.log('使用降级方案 - 图片压缩');
		uni.compressImage({
			src: imagePath,
			quality: quality * 100,
			success: (res) => {
				console.log('图片压缩成功:', res.tempFilePath);
				resolve(res.tempFilePath);
			},
			fail: (error) => {
				console.warn('图片压缩失败，返回原图:', error);
				resolve(imagePath);
			}
		});
		// #endif
	});
}

/**
 * 微信小程序图片裁切 - 真正的裁切实现
 */
function cropImageWechat(imagePath, size, quality, resolve, reject) {
	console.log('开始微信小程序图片裁切:', { imagePath, size, quality });

	// 获取图片信息进行真正的裁切
	uni.getImageInfo({
		src: imagePath,
		success: (imageInfo) => {
			console.log('获取图片信息成功:', imageInfo);
			const { width, height } = imageInfo;

			// 检查是否需要裁切
			const isSquare = Math.abs(width - height) < Math.min(width, height) * 0.05;
			if (isSquare) {
				console.log('图片已经是正方形，只进行压缩');
				// 如果已经是正方形，只进行压缩
				uni.compressImage({
					src: imagePath,
					quality: Math.min(quality * 100, 80),
					success: (res) => {
						console.log('正方形图片压缩成功:', res.tempFilePath);
						resolve(res.tempFilePath);
					},
					fail: (error) => {
						console.error('压缩失败，返回原图:', error);
						resolve(imagePath);
					}
				});
				return;
			}

			// 需要裁切，使用Canvas方案
			console.log('图片需要裁切为正方形，使用Canvas方案');
			cropImageWechatLegacy(imagePath, size, quality, resolve, reject, imageInfo);
		},
		fail: (error) => {
			console.error('获取图片信息失败:', error);
			// 降级到压缩
			console.log('获取图片信息失败，降级到压缩');
			fallbackToCompress(imagePath, quality, resolve);
		}
	});
}

/**
 * 微信小程序图片裁切 - 旧版API降级方案
 */
function cropImageWechatLegacy(imagePath, size, quality, resolve, reject, imageInfo) {
	console.log('使用旧版Canvas API进行裁切');

	const { width, height } = imageInfo || {};
	if (!width || !height) {
		console.error('图片尺寸信息缺失，直接使用压缩降级');
		// 直接降级到压缩
		fallbackToCompress(imagePath, quality, resolve);
		return;
	}

	try {
		// 计算裁切参数
		const minSize = Math.min(width, height);
		const x = (width - minSize) / 2;
		const y = (height - minSize) / 2;

		console.log('Canvas裁切参数:', { x, y, minSize, size });

		// 创建canvas进行裁切
		const canvasId = 'cropCanvas';

		// 检查是否能创建Canvas上下文
		let ctx;
		try {
			ctx = uni.createCanvasContext(canvasId);
			if (!ctx) {
				throw new Error('无法创建Canvas上下文');
			}
		} catch (ctxError) {
			console.error('创建Canvas上下文失败:', ctxError);
			fallbackToCompress(imagePath, quality, resolve);
			return;
		}

		// 绘制裁切后的图片
		ctx.drawImage(
			imagePath,
			x, y, minSize, minSize,  // 源图片裁切区域
			0, 0, size, size         // 目标区域
		);

		// 执行绘制
		ctx.draw(false, () => {
			console.log('Canvas绘制完成，开始导出');

			// 添加延迟确保绘制完成
			setTimeout(() => {
				// 导出图片
				uni.canvasToTempFilePath({
					canvasId: canvasId,
					width: size,
					height: size,
					destWidth: size,
					destHeight: size,
					quality: quality,
					fileType: 'jpg',
					success: (res) => {
						console.log('Canvas裁切成功:', res.tempFilePath);
						if (res.tempFilePath) {
							resolve(res.tempFilePath);
						} else {
							console.error('Canvas导出成功但路径为空');
							fallbackToCompress(imagePath, quality, resolve);
						}
					},
					fail: (error) => {
						console.error('Canvas导出失败:', error);
						fallbackToCompress(imagePath, quality, resolve);
					}
				});
			}, 800); // 增加等待时间到800ms
		});

	} catch (error) {
		console.error('Canvas裁切过程出错:', error);
		fallbackToCompress(imagePath, quality, resolve);
	}
}

/**
 * 降级到图片压缩
 */
function fallbackToCompress(imagePath, quality, resolve) {
	console.log('降级到图片压缩方案');

	uni.compressImage({
		src: imagePath,
		quality: Math.min(quality * 100, 80), // 限制最大质量为80
		success: (compressRes) => {
			console.log('图片压缩成功:', compressRes.tempFilePath);
			resolve(compressRes.tempFilePath);
		},
		fail: (compressError) => {
			console.error('图片压缩也失败，返回原图:', compressError);
			// 最后的最后，直接返回原图
			resolve(imagePath);
		}
	});
}

/**
 * H5环境图片裁切
 */
function cropImageH5(imagePath, size, quality, resolve, reject) {
	const canvas = document.createElement('canvas');
	const ctx = canvas.getContext('2d');
	const img = new Image();
	
	img.crossOrigin = 'anonymous';
	
	img.onload = function() {
		const { width, height } = img;
		
		// 计算裁切参数
		const minSize = Math.min(width, height);
		const x = (width - minSize) / 2;
		const y = (height - minSize) / 2;
		
		// 设置canvas尺寸
		canvas.width = size;
		canvas.height = size;
		
		// 绘制裁切后的图片
		ctx.drawImage(
			img,
			x, y, minSize, minSize,  // 源图片裁切区域
			0, 0, size, size         // 目标区域
		);
		
		// 导出为blob
		canvas.toBlob((blob) => {
			if (blob) {
				const url = URL.createObjectURL(blob);
				resolve(url);
			} else {
				reject(new Error('图片裁切失败'));
			}
		}, 'image/jpeg', quality);
	};
	
	img.onerror = function(error) {
		console.error('H5图片加载失败:', error);
		reject(error);
	};
	
	img.src = imagePath;
}

/**
 * 显示图片裁切选择器
 * @param {string} imagePath - 图片路径
 * @param {Object} options - 选项
 * @returns {Promise<string>} 裁切后的图片路径
 */
export function showImageCropper(imagePath, options = {}) {
	return new Promise((resolve, reject) => {
		// #ifdef MP-WEIXIN
		// 微信小程序使用原生裁切
		uni.showActionSheet({
			itemList: ['使用原图', '裁切为正方形'],
			success: (res) => {
				if (res.tapIndex === 0) {
					// 使用原图
					resolve(imagePath);
				} else if (res.tapIndex === 1) {
					// 裁切为正方形
					cropImageToSquare(imagePath, options)
						.then(resolve)
						.catch(reject);
				}
			},
			fail: () => {
				resolve(imagePath);
			}
		});
		// #endif

		// #ifdef H5
		// H5环境直接裁切
		cropImageToSquare(imagePath, options)
			.then(resolve)
			.catch(reject);
		// #endif

		// #ifndef MP-WEIXIN || H5
		// 其他环境直接返回原图
		resolve(imagePath);
		// #endif
	});
}

/**
 * 验证图片是否为正方形
 * @param {string} imagePath - 图片路径
 * @returns {Promise<boolean>} 是否为正方形
 */
export function isSquareImage(imagePath) {
	return new Promise((resolve) => {
		uni.getImageInfo({
			src: imagePath,
			success: (imageInfo) => {
				const { width, height } = imageInfo;
				const ratio = width / height;
				// 允许5%的误差
				resolve(Math.abs(ratio - 1) < 0.05);
			},
			fail: () => {
				resolve(false);
			}
		});
	});
}

/**
 * 获取图片尺寸信息
 * @param {string} imagePath - 图片路径
 * @returns {Promise<Object>} 图片尺寸信息
 */
export function getImageSize(imagePath) {
	return new Promise((resolve, reject) => {
		uni.getImageInfo({
			src: imagePath,
			success: (imageInfo) => {
				resolve({
					width: imageInfo.width,
					height: imageInfo.height,
					ratio: imageInfo.width / imageInfo.height,
					isSquare: Math.abs(imageInfo.width / imageInfo.height - 1) < 0.05
				});
			},
			fail: reject
		});
	});
}

/**
 * 压缩图片
 * @param {string} imagePath - 图片路径
 * @param {Object} options - 压缩选项
 * @returns {Promise<string>} 压缩后的图片路径
 */
export function compressImage(imagePath, options = {}) {
	const {
		quality = 0.8,
		maxWidth = 800,
		maxHeight = 800
	} = options;

	return new Promise((resolve, reject) => {
		uni.compressImage({
			src: imagePath,
			quality: quality * 100,
			compressedWidth: maxWidth,
			compressedHeight: maxHeight,
			success: (res) => {
				resolve(res.tempFilePath);
			},
			fail: (error) => {
				console.warn('图片压缩失败，使用原图:', error);
				resolve(imagePath);
			}
		});
	});
}

// ============================================================================
// Canvas 操作辅助函数
// ============================================================================

/**
 * 绘制图片到 Canvas
 * @param {Object} ctx - Canvas 上下文
 * @param {string} imagePath - 图片路径
 * @param {Object} sourceRect - 源图片裁切区域 {x, y, width, height}
 * @param {Object} destRect - 目标区域 {x, y, width, height}
 * @returns {Promise<void>}
 */
export function drawImageToCanvas(ctx, imagePath, sourceRect, destRect) {
	return new Promise((resolve, reject) => {
		try {
			ctx.drawImage(
				imagePath,
				sourceRect.x,
				sourceRect.y,
				sourceRect.width,
				sourceRect.height,
				destRect.x,
				destRect.y,
				destRect.width,
				destRect.height
			);
			resolve();
		} catch (error) {
			reject(error);
		}
	});
}

/**
 * 旋转 Canvas
 * @param {number} angle - 旋转角度（度数）
 * @param {Object} options - 选项
 * @returns {Object} - { width, height } 旋转后的尺寸
 */
export function getRotatedSize(width, height, angle) {
	const rad = (angle * Math.PI) / 180;
	const cos = Math.abs(Math.cos(rad));
	const sin = Math.abs(Math.sin(rad));
	return {
		width: width * cos + height * sin,
		height: width * sin + height * cos
	};
}

/**
 * 水平翻转 Canvas 尺寸
 * @param {number} width - 原始宽度
 * @returns {number} 翻转后的宽度
 */
export function getFlippedWidth(width) {
	return width;
}

/**
 * 创建变换矩阵
 * @param {Object} transform - 变换参数
 * @returns {string} CSS transform 值
 */
export function createTransformMatrix(transform) {
	const { translateX = 0, translateY = 0, scale = 1, rotation = 0 } = transform;
	return `translate(${translateX}px, ${translateY}px) scale(${scale}) rotate(${rotation}deg)`;
}

/**
 * 计算裁切区域
 * @param {Object} imageSize - 图片尺寸 { width, height }
 * @param {Object} containerSize - 容器尺寸 { width, height }
 * @param {number} cropSize - 裁切框大小
 * @returns {Object} 裁切参数
 */
export function calculateCropArea(imageSize, containerSize, cropSize) {
	const { width: imgW, height: imgH } = imageSize;
	const { width: containerW, height: containerH } = containerSize;

	// 计算图片在容器中的显示尺寸
	let displayW, displayH;
	const imgRatio = imgW / imgH;
	const containerRatio = containerW / containerH;

	if (imgRatio > containerRatio) {
		displayH = containerH;
		displayW = containerH * imgRatio;
	} else {
		displayW = containerW;
		displayH = containerW / imgRatio;
	}

	// 计算裁切框位置（居中）
	const cropX = (displayW - cropSize) / 2;
	const cropY = (displayH - cropSize) / 2;

	// 计算裁切比例
	const scaleX = imgW / displayW;
	const scaleY = imgH / displayH;

	return {
		sourceX: cropX * scaleX,
		sourceY: cropY * scaleY,
		sourceWidth: cropSize * scaleX,
		sourceHeight: cropSize * scaleY,
		destWidth: cropSize,
		destHeight: cropSize,
		displayWidth: displayW,
		displayHeight: displayH,
		offsetX: (containerW - displayW) / 2,
		offsetY: (containerH - displayH) / 2
	};
}

/**
 * 导出 Canvas 为图片
 * @param {string} canvasId - Canvas ID
 * @param {Object} options - 导出选项
 * @returns {Promise<string>} 图片路径
 */
export function exportCanvasToImage(canvasId, options = {}) {
	const {
		width = 400,
		height = 400,
		quality = 0.8,
		fileType = 'jpg'
	} = options;

	return new Promise((resolve, reject) => {
		uni.canvasToTempFilePath(
			{
				canvasId,
				width,
				height,
				destWidth: width,
				destHeight: height,
				quality,
				fileType,
				success: (res) => {
					if (res.tempFilePath) {
						resolve(res.tempFilePath);
					} else {
						reject(new Error('Canvas 导出路径为空'));
					}
				},
				fail: reject
			},
			null
		);
	});
}

/**
 * 计算拖拽后的偏移量
 * @param {Object} current - 当前状态
 * @param {number} deltaX - X 方向变化量
 * @param {number} deltaY - Y 方向变化量
 * @returns {Object} 新的偏移量
 */
export function calculateDragOffset(current, deltaX, deltaY) {
	return {
		translateX: current.translateX + deltaX,
		translateY: current.translateY + deltaY
	};
}

/**
 * 限制缩放范围
 * @param {number} scale - 当前缩放值
 * @param {number} minScale - 最小缩放值
 * @param {number} maxScale - 最大缩放值
 * @returns {number} 限制后的缩放值
 */
export function clampScale(scale, minScale = 0.5, maxScale = 3) {
	return Math.min(Math.max(scale, minScale), maxScale);
}

/**
 * 缩放处理
 * @param {number} currentScale - 当前缩放值
 * @param {number} factor - 缩放因子（大于1放大，小于1缩小）
 * @param {number} minScale - 最小缩放值
 * @param {number} maxScale - 最大缩放值
 * @returns {number} 新的缩放值
 */
export function applyScale(currentScale, factor, minScale = 0.5, maxScale = 3) {
	const newScale = currentScale * factor;
	return clampScale(newScale, minScale, maxScale);
}

/**
 * 旋转角度规范化（0-359）
 * @param {number} angle - 角度
 * @returns {number} 规范化后的角度
 */
export function normalizeAngle(angle) {
	return ((angle % 360) + 360) % 360;
}

/**
 * 初始化裁切器参数
 * @param {Object} imageInfo - 图片信息 { width, height }
 * @param {Object} containerInfo - 容器信息 { width, height }
 * @returns {Object} 初始化参数
 */
export function initCropperParams(imageInfo, containerInfo) {
	const { width: imgW, height: imgH } = imageInfo;
	const { width: containerW, height: containerH } = containerInfo;

	// 计算图片在容器中的显示尺寸
	const containerRatio = containerW / containerH;
	const imageRatio = imgW / imgH;

	let displayW, displayH;
	if (imageRatio > containerRatio) {
		displayH = containerH;
		displayW = containerH * imageRatio;
	} else {
		displayW = containerW;
		displayH = containerW / imageRatio;
	}

	// 计算裁切框大小（容器宽高的80%）
	const cropBoxSize = Math.min(containerW, containerH) * 0.8;

	return {
		imageWidth: displayW,
		imageHeight: displayH,
		cropBoxSize,
		translateX: 0,
		translateY: 0,
		scale: 1,
		rotation: 0
	};
}

/**
 * H5 环境：绘制旋转后的图片
 * @param {CanvasRenderingContext2D} ctx - Canvas 上下文
 * @param {HTMLImageElement} img - 图片元素
 * @param {number} rotation - 旋转角度
 * @param {number} destX - 目标 X 坐标
 * @param {number} destY - 目标 Y 坐标
 * @param {number} destW - 目标宽度
 * @param {number} destH - 目标高度
 */
export function drawRotatedImage(ctx, img, rotation, destX, destY, destW, destH) {
	ctx.save();
	ctx.translate(destX + destW / 2, destY + destH / 2);
	ctx.rotate((rotation * Math.PI) / 180);
	ctx.drawImage(img, -destW / 2, -destH / 2, destW, destH);
	ctx.restore();
}

/**
 * H5 环境：创建旋转后的 Canvas
 * @param {string} imagePath - 图片路径
 * @param {Object} options - 选项
 * @returns {Promise<{ canvas: HTMLCanvasElement, width: number, height: number }>}
 */
export function createRotatedCanvas(imagePath, options = {}) {
	return new Promise((resolve, reject) => {
		const img = new Image();
		img.crossOrigin = 'anonymous';

		img.onload = function () {
			const { rotation = 0 } = options;
			const { width, height } = getRotatedSize(img.width, img.height, rotation);

			const canvas = document.createElement('canvas');
			canvas.width = width;
			canvas.height = height;
			const ctx = canvas.getContext('2d');

			drawRotatedImage(ctx, img, rotation, 0, 0, img.width, img.height);

			resolve({ canvas, width, height });
		};

		img.onerror = reject;
		img.src = imagePath;
	});
}

/**
 * 微信小程序：获取系统信息
 * @returns {Object} 系统信息
 */
export function getSystemInfo() {
	// #ifdef MP-WEIXIN
	try {
		const systemInfo = uni.getSystemInfoSync();
		let statusBarHeight = 0;
		let capsuleInfo = null;

		if (systemInfo) {
			statusBarHeight = systemInfo.statusBarHeight || 0;
		}

		if (uni.getMenuButtonBoundingClientRect) {
			capsuleInfo = uni.getMenuButtonBoundingClientRect();
		}

		return { systemInfo, statusBarHeight, capsuleInfo };
	} catch (error) {
		console.error('获取系统信息失败:', error);
		return { systemInfo: null, statusBarHeight: 0, capsuleInfo: null };
	}
	// #endif

	// #ifndef MP-WEIXIN
	return { systemInfo: null, statusBarHeight: 0, capsuleInfo: null };
	// #endif
}

/**
 * 计算顶部遮罩高度
 * @param {Object} capsuleInfo - 胶囊按钮信息
 * @param {number} statusBarHeight - 状态栏高度
 * @returns {number} 顶部遮罩高度
 */
export function calculateTopMaskHeight(capsuleInfo, statusBarHeight) {
	// #ifdef MP-WEIXIN
	if (capsuleInfo && statusBarHeight) {
		const capsuleBottom = capsuleInfo.bottom || 0;
		const extraSpace = 20;
		return Math.max(capsuleBottom + extraSpace, 200);
	}
	return 200;
	// #endif

	// #ifndef MP-WEIXIN
	return 100;
	// #endif
}
