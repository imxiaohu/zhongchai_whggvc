<template>
	<view class="container">
		<CustomNavBar
			center-title="图片上传预览测试"
			:show-back="true"
			:transparent="false"
			:fixed="true"
		/>
		
		<view class="content-area" style="padding-top: 88px;">
			<view class="test-section">
				<text class="section-title">图片上传预览测试</text>
				<text class="section-desc">测试多图上传的布局效果</text>
				
				<!-- 图片上传区域 -->
				<view class="images-upload" :class="getImageGridClassFn(testImages.length)">
					<view
						v-for="(image, index) in testImages"
						:key="index"
						class="image-item"
						:class="{ 'svg-image': isSvgImage(image) }"
					>
						<image
							class="uploaded-image"
							:src="image"
							mode="aspectFill"
							@tap="previewImage(image, testImages)"
						></image>
						<view v-if="isSvgImage(image)" class="svg-overlay">
							<text class="svg-label">SVG</text>
						</view>
						<view class="image-actions">
							<view class="action-btn preview-btn" @tap="previewImage(image, testImages)">
								<l-icon name="browse" size="14" color="#fff"></l-icon>
							</view>
							<view class="action-btn remove-btn" @tap="removeImage(index)">
								<l-icon name="close" size="14" color="#fff"></l-icon>
							</view>
						</view>
					</view>
					<view
						v-if="testImages.length < 9"
						class="add-image-btn"
						@tap="selectImages"
					>
						<view class="add-image-content">
							<l-icon name="photo" size="24" color="var(--text-secondary)"></l-icon>
							<text class="add-text">{{ testImages.length }}/9</text>
							<text class="add-hint">点击添加图片</text>
						</view>
					</view>
				</view>
				
				<!-- 测试按钮 -->
				<view class="test-buttons">
					<button class="test-btn" @tap="addTestImage">添加测试图片</button>
					<button class="test-btn" @tap="clearImages">清空图片</button>
					<button class="test-btn" @tap="addMultipleImages">添加多张图片</button>
				</view>
				
				<!-- 当前状态 -->
				<view class="status-info">
					<text class="status-text">当前图片数量: {{ testImages.length }}</text>
					<text class="status-text">网格类名: {{ getImageGridClass(testImages.length) }}</text>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref } from 'vue'
import { previewImages, isSvgImage as checkSvgImage, getImageGridClass } from '../../utils/imageUpload.js'
import CustomNavBar from '../../components/CustomNavBar.vue'

// 状态
const testImages = ref([])

// 选择图片
function selectImages() {
	const maxCount = 9 - testImages.value.length
	uni.chooseImage({
		count: maxCount,
		sizeType: ['compressed'],
		sourceType: ['album', 'camera'],
		success: (res) => {
			testImages.value.push(...res.tempFilePaths)
		},
		fail: (error) => {
			console.error('选择图片失败:', error)
		}
	})
}

// 移除图片
function removeImage(index) {
	testImages.value.splice(index, 1)
}

// 预览图片
function previewImage(current, urls = null) {
	previewImages(current, urls || testImages.value)
}

// 检测SVG图片
function isSvgImage(imageUrl) {
	return checkSvgImage(imageUrl)
}

// 获取图片网格类名
function getImageGridClassFn(count) {
	return getImageGridClass(count)
}

// 添加测试图片
function addTestImage() {
	const testUrls = [
		'https://picsum.photos/300/300?random=1',
		'https://picsum.photos/300/300?random=2',
		'https://picsum.photos/300/300?random=3',
		'https://picsum.photos/300/300?random=4',
		'https://picsum.photos/300/300?random=5'
	]
	
	if (testImages.value.length < 9) {
		const randomUrl = testUrls[Math.floor(Math.random() * testUrls.length)]
		testImages.value.push(randomUrl)
	}
}

// 清空图片
function clearImages() {
	testImages.value = []
}

// 添加多张图片
function addMultipleImages() {
	const testUrls = [
		'https://picsum.photos/300/300?random=10',
		'https://picsum.photos/300/300?random=11',
		'https://picsum.photos/300/300?random=12'
	]
	
	const remainingSlots = 9 - testImages.value.length
	const imagesToAdd = Math.min(testUrls.length, remainingSlots)
	
	for (let i = 0; i < imagesToAdd; i++) {
		testImages.value.push(testUrls[i])
	}
}
</script>

<style lang="scss" scoped>
	.container {
		height: 100vh;
		background: linear-gradient(135deg, var(--bg-primary) 0%, var(--bg-secondary) 100%);
	}
	
	.content-area {
		height: 100%;
		padding: 32rpx;
	}
	
	.test-section {
		display: flex;
		flex-direction: column;
		gap: 32rpx;
	}
	
	.section-title {
		font-size: 32rpx;
		font-weight: 600;
		color: var(--text-primary);
	}
	
	.section-desc {
		font-size: 24rpx;
		color: var(--text-secondary);
		margin-bottom: 16rpx;
	}
	
	/* 复制 create-post.vue 中的图片上传样式 */
	.images-upload {
		display: flex;
		flex-wrap: wrap;
		margin: -8rpx;
	}

	.images-upload .image-item,
	.images-upload .add-image-btn {
		flex: 0 0 calc(100% / 3);
		max-width: calc(100% / 3);
		padding: 8rpx;
		box-sizing: border-box;
	}

	.image-item {
		position: relative;
		width: 100%;
		height: 0;
		padding-bottom: 100%;
		border-radius: 12rpx;
		overflow: hidden;
		box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
		transition: all 0.3s ease;
	}

	.uploaded-image {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform 0.3s ease;
	}

	.image-actions {
		position: absolute;
		top: 8rpx;
		right: 8rpx;
		display: flex;
		gap: 8rpx;
		opacity: 0;
		transition: opacity 0.3s ease;
	}

	.image-item:hover .image-actions {
		opacity: 1;
	}

	.action-btn {
		width: 28rpx;
		height: 28rpx;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.3s ease;
	}

	.preview-btn {
		background: rgba(0, 122, 255, 0.8);
	}

	.remove-btn {
		background: rgba(255, 59, 48, 0.8);
	}

	.add-image-btn {
		width: 100%;
		height: 0;
		padding-bottom: 100%;
		border: 2rpx dashed var(--border-color);
		border-radius: 12rpx;
		background: linear-gradient(135deg, var(--bg-secondary), var(--bg-tertiary));
		transition: all 0.3s ease;
		position: relative;
		overflow: hidden;
	}

	.add-image-content {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 8rpx;
	}

	.add-text {
		font-size: 20rpx;
		color: var(--text-secondary);
		font-weight: 500;
	}

	.add-hint {
		font-size: 18rpx;
		color: var(--text-tertiary);
	}
	
	.test-buttons {
		display: flex;
		flex-wrap: wrap;
		gap: 16rpx;
	}
	
	.test-btn {
		flex: 1;
		min-width: 120rpx;
		padding: 16rpx 24rpx;
		border-radius: 12rpx;
		background: var(--primary-color);
		color: #fff;
		font-size: 24rpx;
		border: none;
	}
	
	.status-info {
		padding: 24rpx;
		border-radius: 12rpx;
		background: var(--bg-secondary);
		border: 1px solid var(--border-color);
	}
	
	.status-text {
		display: block;
		font-size: 24rpx;
		color: var(--text-primary);
		margin-bottom: 8rpx;
	}
</style>
