// imageUpload - 统一导出
// 所有 API 从子模块重新导出，向后兼容

// 只从 imageUpload.core.js 导入它实际导出的内容
import {
	getImageToken,
	validateToken,
	validateImage,
	compressImage,
	chooseImages,
	previewImage,
	uploadImage,
	processBlobUrl,
	generateImageUrl,
	isSvgImage,
	getImageGridClass,
	API_BASE_URL,
	UPLOAD_CONFIG,
	generateUniqueId
} from './imageUpload.core.js';

// 从 imageUpload.batch.js 导入
import {
	uploadImages,
	uploadImagesLegacy,
	uploadSingleImage,
	deleteUploadedImage,
	validateImageFile,
	previewImages
} from './imageUpload.batch.js';

export {
	getImageToken,
	validateToken,
	validateImage,
	compressImage,
	processBlobUrl,
	generateImageUrl,
	isSvgImage,
	getImageGridClass,
	chooseImages,
	previewImage,
	uploadImage,
	API_BASE_URL,
	UPLOAD_CONFIG,
	generateUniqueId
};

export {
	uploadImages,
	uploadImagesLegacy,
	uploadSingleImage,
	deleteUploadedImage,
	validateImageFile,
	previewImages
};

export default {
	getImageToken,
	validateToken,
	validateImage,
	compressImage,
	chooseImages,
	previewImage,
	uploadImage,
	uploadImages,
	deleteUploadedImage,
	uploadSingleImage,
	uploadImagesLegacy,
	validateImageFile,
	previewImages
};
