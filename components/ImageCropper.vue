<template>
	<view class="image-cropper-modal" v-if="visible" @tap="handleMaskTap">
		<view class="cropper-container" @tap.stop>
			<!-- 顶部遮罩区域 -->
			<view class="top-mask" @tap="handleMaskTap" :style="{ minHeight: topMaskHeight + 'rpx' }"></view>

			<!-- 底部裁切面板 -->
			<view class="cropper-panel">
				<!-- 标题栏 -->
				<view class="panel-header">
					<text class="header-title">裁切头像</text>
					<view class="header-actions">
						<button class="action-btn cancel-btn" @tap="handleCancel">取消</button>
						<button class="action-btn confirm-btn" @tap="handleConfirm">确定</button>
					</view>
				</view>

				<!-- 图片容器 -->
				<view class="image-container" :style="containerStyle">
					<image
						:src="imageSrc"
						class="crop-image"
						:style="imageStyle"
						mode="aspectFit"
						@load="onImageLoad"
						@error="onImageError"
						@touchstart="onTouchStart"
						@touchmove="onTouchMove"
						@touchend="onTouchEnd"
					/>

					<!-- 裁切框 - 简化版本 -->
					<view class="crop-mask">
						<!-- 圆形裁切区域 -->
						<view class="crop-circle" :style="cropCircleStyle"></view>
					</view>

					<!-- 隐藏的Canvas用于图片裁切 -->
					<canvas
						id="cropCanvas"
						canvas-id="cropCanvas"
						class="crop-canvas"
						width="400"
						height="400"
						:style="{ width: '400px', height: '400px' }"
					></canvas>
				</view>

				<!-- 操作按钮 -->
				<view class="cropper-controls">
					<button class="control-btn" @tap="handleZoomOut" hover-class="control-btn-hover">
						<!-- #ifdef MP-WEIXIN -->
						<text class="btn-icon">-</text>
						<!-- #endif -->
						<!-- #ifndef MP-WEIXIN -->
						<text class="mdi mdi-minus"></text>
						<!-- #endif -->
						<text class="btn-text">缩小</text>
					</button>
					<button class="control-btn" @tap="handleZoomIn" hover-class="control-btn-hover">
						<!-- #ifdef MP-WEIXIN -->
						<text class="btn-icon">+</text>
						<!-- #endif -->
						<!-- #ifndef MP-WEIXIN -->
						<text class="mdi mdi-plus"></text>
						<!-- #endif -->
						<text class="btn-text">放大</text>
					</button>
					<button class="control-btn" @tap="handleRotate" hover-class="control-btn-hover">
						<!-- #ifdef MP-WEIXIN -->
						<text class="btn-icon">↻</text>
						<!-- #endif -->
						<!-- #ifndef MP-WEIXIN -->
						<text class="mdi mdi-rotate-right"></text>
						<!-- #endif -->
						<text class="btn-text">旋转</text>
					</button>
					<button class="control-btn" @tap="handleReset" hover-class="control-btn-hover">
						<!-- #ifdef MP-WEIXIN -->
						<text class="btn-icon">⟲</text>
						<!-- #endif -->
						<!-- #ifndef MP-WEIXIN -->
						<text class="mdi mdi-refresh"></text>
						<!-- #endif -->
						<text class="btn-text">重置</text>
					</button>
				</view>
			</view>
		</view>
	</view>
</template>

<script>
import {
	cropImageToSquare,
	getSystemInfo,
	calculateTopMaskHeight,
	initCropperParams
} from '@/utils/imageCrop.js';

export default {
	name: 'ImageCropper',
	props: {
		visible: {
			type: Boolean,
			default: false
		},
		imageSrc: {
			type: String,
			default: ''
		},
		cropSize: {
			type: Number,
			default: 300
		}
	},
	data() {
		return {
			containerWidth: 0,
			containerHeight: 0,
			imageWidth: 0,
			imageHeight: 0,
			imageNaturalWidth: 0,
			imageNaturalHeight: 0,
			scale: 1,
			rotation: 0,
			translateX: 0,
			translateY: 0,
			cropBoxSize: 300,
			isDragging: false,
			lastTouchX: 0,
			lastTouchY: 0,
			statusBarHeight: 0,
			capsuleInfo: null
		};
	},
	computed: {
		topMaskHeight() {
			return calculateTopMaskHeight(this.capsuleInfo, this.statusBarHeight);
		},

		containerStyle() {
			return {
				width: '100%',
				height: '500rpx',
				position: 'relative',
				overflow: 'hidden',
				background: '#000'
			};
		},
		imageStyle() {
			return {
				width: this.imageWidth + 'px',
				height: this.imageHeight + 'px',
				transform: `translate(${this.translateX}px, ${this.translateY}px) scale(${this.scale}) rotate(${this.rotation}deg)`,
				transformOrigin: 'center center',
				transition: this.isDragging ? 'none' : 'transform 0.3s ease'
			};
		},
		cropCircleStyle() {
			const size = this.cropBoxSize;
			const left = (this.containerWidth - size) / 2;
			const top = (this.containerHeight - size) / 2;

			return {
				width: size + 'px',
				height: size + 'px',
				left: left + 'px',
				top: top + 'px'
			};
		}
	},
	watch: {
		visible(newVal) {
			if (newVal) {
				this.$nextTick(() => {
					this.initSystemInfo();
					this.initCropper();
				});
			}
		}
	},
	methods: {
		initSystemInfo() {
			const info = getSystemInfo();
			this.statusBarHeight = info.statusBarHeight;
			this.capsuleInfo = info.capsuleInfo;
		},

		initCropper() {
			uni.createSelectorQuery().in(this).select('.image-container').boundingClientRect((rect) => {
				if (rect) {
					this.containerWidth = rect.width;
					this.containerHeight = rect.height;
					this.cropBoxSize = Math.min(rect.width, rect.height) * 0.8;
				}
			}).exec();
		},
		
		onImageLoad(e) {
			uni.getImageInfo({
				src: this.imageSrc,
				success: (res) => {
					this.imageNaturalWidth = res.width;
					this.imageNaturalHeight = res.height;
					this.calculateImageSize();
				}
			});
		},
		
		onImageError(e) {
			uni.showToast({
				title: '图片加载失败',
				icon: 'none'
			});
		},
		
		calculateImageSize() {
			if (!this.containerWidth || !this.imageNaturalWidth) return;
			
			const containerRatio = this.containerWidth / this.containerHeight;
			const imageRatio = this.imageNaturalWidth / this.imageNaturalHeight;
			
			if (imageRatio > containerRatio) {
				this.imageHeight = this.containerHeight;
				this.imageWidth = this.containerHeight * imageRatio;
			} else {
				this.imageWidth = this.containerWidth;
				this.imageHeight = this.containerWidth / imageRatio;
			}
			
			this.translateX = 0;
			this.translateY = 0;
			this.scale = 1;
		},
		
		handleZoomIn() {
			this.scale = Math.min(this.scale * 1.2, 3);
		},

		handleZoomOut() {
			this.scale = Math.max(this.scale / 1.2, 0.5);
		},

		handleRotate() {
			this.rotation = (this.rotation + 90) % 360;
		},

		handleReset() {
			this.scale = 1;
			this.rotation = 0;
			this.translateX = 0;
			this.translateY = 0;
		},
		
		handleCancel() {
			this.$emit('cancel');
		},
		
		async handleConfirm() {
			try {
				const croppedImage = await this.getCroppedImage();
				this.$emit('confirm', croppedImage);
			} catch (error) {
				uni.showToast({
					title: '裁切失败',
					icon: 'none'
				});
			}
		},
		
		handleMaskTap() {
			this.handleCancel();
		},

		onTouchStart(e) {
			if (e.touches.length === 1) {
				this.isDragging = true;
				this.lastTouchX = e.touches[0].clientX;
				this.lastTouchY = e.touches[0].clientY;
			}
		},

		onTouchMove(e) {
			if (this.isDragging && e.touches.length === 1) {
				const deltaX = e.touches[0].clientX - this.lastTouchX;
				const deltaY = e.touches[0].clientY - this.lastTouchY;

				this.translateX += deltaX;
				this.translateY += deltaY;

				this.lastTouchX = e.touches[0].clientX;
				this.lastTouchY = e.touches[0].clientY;
			}
		},

		onTouchEnd() {
			this.isDragging = false;
		},
		
		async getCroppedImage() {
			try {
				const croppedImage = await cropImageToSquare(this.imageSrc, {
					size: 400,
					quality: 0.8
				});
				return croppedImage;
			} catch (error) {
				throw error;
			}
		}
	}
};
</script>

<style lang="scss" scoped>
.image-cropper-modal {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.8);
	z-index: 9999;
	display: flex;
	flex-direction: column;
	width: 100vw;
	height: 100vh;
	box-sizing: border-box;
	/* #ifdef MP-WEIXIN */
	position: fixed !important;
	top: 0 !important;
	left: 0 !important;
	right: 0 !important;
	bottom: 0 !important;
	width: 100vw !important;
	height: 100vh !important;
	z-index: 999999 !important;
	/* #endif */
}

.top-mask {
	flex: 1;
	background: transparent;
	min-height: 0;
	/* #ifdef MP-WEIXIN */
	min-height: 200rpx;
	/* #endif */
}

.cropper-container {
	display: flex;
	flex-direction: column;
	background: #000;
	flex-shrink: 0;
}

.cropper-panel {
	background: #000;
	border-radius: 32rpx 32rpx 0 0;
	display: flex;
	flex-direction: column;
	box-sizing: border-box;
	padding-bottom: env(safe-area-inset-bottom);
	/* #ifdef MP-WEIXIN */
	height: 90vh;
	min-height: 70vh;
	flex-shrink: 0;
	/* #endif */
	/* #ifndef MP-WEIXIN */
	max-height: 80vh;
	min-height: 400rpx;
	/* #endif */
}

.panel-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 32rpx;
	background: rgba(0, 0, 0, 0.9);
	border-radius: 32rpx 32rpx 0 0;
	position: relative;
	flex-shrink: 0;
	box-sizing: border-box;
	/* #ifdef MP-WEIXIN */
	padding: 24rpx 32rpx;
	min-height: 120rpx;
	/* #endif */
}

.panel-header::after {
	content: '';
	position: absolute;
	top: 16rpx;
	left: 50%;
	transform: translateX(-50%);
	width: 80rpx;
	height: 8rpx;
	background: rgba(255, 255, 255, 0.3);
	border-radius: 4rpx;
	/* #ifdef MP-WEIXIN */
	top: 12rpx;
	width: 60rpx;
	height: 6rpx;
	/* #endif */
}

.header-title {
	color: #fff;
	font-size: 32rpx;
	font-weight: 600;
	/* #ifdef MP-WEIXIN */
	font-size: 30rpx;
	/* #endif */
}

.header-actions {
	display: flex;
	gap: 24rpx;
	/* #ifdef MP-WEIXIN */
	gap: 16rpx;
	/* #endif */
}

.action-btn {
	padding: 16rpx 32rpx;
	border-radius: 24rpx;
	font-size: 28rpx;
	border: none;
	box-sizing: border-box;
	/* #ifdef MP-WEIXIN */
	padding: 12rpx 24rpx;
	font-size: 26rpx;
	min-width: 120rpx;
	text-align: center;
	/* #endif */
}

.cancel-btn {
	background: rgba(255, 255, 255, 0.1);
	color: #fff;
	/* #ifdef MP-WEIXIN */
	background: rgba(255, 255, 255, 0.15);
	/* #endif */
}

.confirm-btn {
	background: #6366f1;
	color: #fff;
	/* #ifdef MP-WEIXIN */
	background: #5b5bd6;
	/* #endif */
}

.image-container {
	position: relative;
	display: flex;
	align-items: center;
	justify-content: center;
	margin: 32rpx;
	border-radius: 16rpx;
	overflow: hidden;
	box-sizing: border-box;
	flex: 1;
	/* #ifdef MP-WEIXIN */
	margin: 24rpx;
	border-radius: 12rpx;
	min-height: 400rpx;
	background: rgba(255, 255, 255, 0.05);
	/* #endif */
	/* #ifndef MP-WEIXIN */
	height: 600rpx;
	flex-shrink: 0;
	/* #endif */
}

.crop-image {
	position: absolute;
	/* #ifdef MP-WEIXIN */
	max-width: 100%;
	max-height: 100%;
	/* #endif */
}

.crop-mask {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	pointer-events: none;
}

.crop-circle {
	position: absolute;
	border: 4px solid #fff;
	border-radius: 50%;
	box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.6);
	background: transparent;
	/* #ifdef MP-WEIXIN */
	border: 5px solid #fff;
	box-shadow:
		0 0 0 9999px rgba(0, 0, 0, 0.7),
		0 0 20px rgba(255, 255, 255, 0.3);
	/* #endif */
}

.crop-canvas {
	position: absolute;
	top: -9999px;
	left: -9999px;
	width: 400px;
	height: 400px;
	opacity: 0;
	pointer-events: none;
}

.cropper-controls {
	display: flex;
	justify-content: center;
	gap: 40rpx;
	padding: 40rpx;
	background: rgba(0, 0, 0, 0.8);
	flex-shrink: 0;
	box-sizing: border-box;
	/* #ifdef MP-WEIXIN */
	gap: 32rpx;
	padding: 32rpx 24rpx;
	/* #endif */
}

.control-btn {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 8rpx;
	background: rgba(255, 255, 255, 0.1);
	border: none;
	border-radius: 16rpx;
	color: #fff;
	font-size: 40rpx;
	padding: 16rpx 12rpx;
	min-width: 80rpx;
	transition: all 0.3s ease;
	box-sizing: border-box;
	/* #ifdef MP-WEIXIN */
	font-size: 36rpx;
	padding: 12rpx 8rpx;
	min-width: 72rpx;
	border-radius: 12rpx;
	-webkit-tap-highlight-color: transparent;
	/* #endif */
}

.control-btn:active {
	background: rgba(255, 255, 255, 0.2);
	transform: scale(0.95);
	/* #ifdef MP-WEIXIN */
	background: rgba(255, 255, 255, 0.25);
	/* #endif */
}

.btn-text {
	font-size: 20rpx;
	color: #fff;
	text-align: center;
	/* #ifdef MP-WEIXIN */
	font-size: 18rpx;
	/* #endif */
}

/* #ifdef MP-WEIXIN */
.btn-icon {
	font-size: 32rpx;
	color: #fff;
	font-weight: bold;
	line-height: 1;
}

.control-btn-hover {
	background: rgba(255, 255, 255, 0.3) !important;
	transform: scale(0.95) !important;
}
/* #endif */

/* #ifdef MP-WEIXIN */
.image-cropper-modal {
	position: fixed !important;
	top: 0 !important;
	left: 0 !important;
	right: 0 !important;
	bottom: 0 !important;
	width: 100vw !important;
	height: 100vh !important;
	z-index: 999999 !important;
	background: rgba(0, 0, 0, 0.9) !important;
}

.top-mask {
	min-height: 200rpx !important;
	background: transparent;
}

.cropper-panel {
	flex-shrink: 0 !important;
	height: 90vh !important;
	min-height: 70vh !important;
	position: relative;
	z-index: 1000000;
}

.panel-header {
	flex-shrink: 0 !important;
	padding: 32rpx !important;
	min-height: 120rpx !important;
}

.image-container {
	background: rgba(255, 255, 255, 0.05) !important;
	touch-action: manipulation;
	flex: 1 !important;
	min-height: 400rpx !important;
	flex-shrink: 0 !important;
}

.crop-image {
	object-fit: contain !important;
	max-width: 100% !important;
	max-height: 100% !important;
}

.crop-mask {
	background: rgba(0, 0, 0, 0.7) !important;
}

.crop-circle {
	border: 5px solid #fff !important;
	box-shadow:
		0 0 0 3px rgba(0, 0, 0, 0.5),
		0 0 20px rgba(255, 255, 255, 0.3) !important;
}

.cropper-controls {
	flex-shrink: 0 !important;
	padding: 40rpx 24rpx !important;
	background: rgba(0, 0, 0, 0.9) !important;
}

.control-btn {
	user-select: none;
	-webkit-user-select: none;
	pointer-events: auto;
	min-width: 80rpx !important;
	min-height: 80rpx !important;
	flex-shrink: 0;
}

.cropper-panel {
	padding-bottom: calc(env(safe-area-inset-bottom) + 40rpx) !important;
}

.panel-header,
.image-container,
.cropper-controls {
	position: relative;
	z-index: 1000001;
}
/* #endif */
</style>
