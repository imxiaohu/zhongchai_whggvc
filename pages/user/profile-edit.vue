<template>
	<view class="container">
		<!-- 自定义导航栏 -->
		<CustomNavBar
			:center-title="'编辑资料'"
			:show-back="true"
			@leftClick="goBack"
			@navHeightReady="handleNavHeightReady"
		/>

		<view class="content" :style="{paddingTop: navPaddingTop}">
			<!-- 头像上传区域 -->
			<view class="avatar-section">
				<view class="avatar-wrapper" @tap="selectAvatar">
					<UserAvatar
						class="avatar-image"
						:src="userInfo.avatar"
						:name="userInfo.nickname || userInfo.name"
						:size="160"
					></UserAvatar>
					<view class="avatar-overlay">
						<l-icon name="photo" size="20" color="#fff"></l-icon>
					</view>
				</view>
				<text class="avatar-hint">点击更换头像</text>
			</view>

			<!-- 表单区域 -->
			<view class="form-group">
				<!-- 姓名（只读） -->
				<view class="form-item">
					<view class="form-item__label">真实姓名</view>
					<view class="form-item__content readonly">
						<text class="form-item__content-text">{{ userInfo.name || '未设置' }}</text>
						<view class="lock-icon">
							<l-icon name="lock-on" size="14" color="var(--text-tertiary)"></l-icon>
						</view>
					</view>
				</view>

				<!-- 昵称 -->
				<view class="form-item">
					<view class="form-item__label">昵称</view>
					<view class="form-item__content">
						<input 
							class="form-input" 
							type="text" 
							v-model="userInfo.nickname" 
							placeholder="请输入昵称"
							placeholder-class="placeholder"
							maxlength="20"
						/>
					</view>
				</view>

				<!-- 邮箱 -->
				<view class="form-item form-item--email">
					<view class="form-item__label">
						<view class="email-label-badge">
							<l-icon name="mail" style="font-size: 12px; color: #fff;"></l-icon>
						</view>
						邮箱
					</view>
					<view class="form-item__content form-item__content--email">
						<input 
							class="form-input form-input--email" 
							type="text" 
							v-model="userInfo.email" 
							placeholder="请输入邮箱地址"
							placeholder-class="placeholder"
						/>
					</view>
				</view>
			</view>

			<!-- 保存按钮 -->
			<view class="footer-actions">
				<view class="save-btn" :class="{ 'save-btn--loading': isSaving }" @tap="saveProfile">
					<text class="save-btn-text" v-if="!isSaving">保存修改</text>
					<view v-else class="loading-spinner"></view>
				</view>
			</view>
		</view>
	</view>

	<!-- 隐藏的Canvas，用于图片裁切 -->
	<!-- #ifdef MP-WEIXIN -->
	<canvas
		canvas-id="cropCanvas"
		style="position: fixed; top: -9999px; left: -9999px; width: 400px; height: 400px;"
	></canvas>
	<!-- #endif -->

	<!-- 图片裁切组件 -->
	<ImageCropper
		:visible="showCropper"
		:imageSrc="selectedImagePath"
		@confirm="handleCropConfirm"
		@cancel="handleCropCancel"
	/>
</template>

<script>
	import { showToast, navigateBack } from '../../pages/api/page.js';
	import { getUserInfo, updateUserInfo } from '../api/user.js';
	import { uploadSingleImage, validateImage } from '../../utils/imageUpload.js';
	import { showImageCropper, cropImageToSquare, getImageSize } from '../../utils/imageCrop.js';
	import CustomNavBar from '../../components/CustomNavBar.vue';
	import ImageCropper from '../../components/ImageCropper.vue';
	import UserAvatar from '../../components/UserAvatar.vue';

	export default {
		components: {
			CustomNavBar,
			ImageCropper,
			UserAvatar
		},
		data() {
			return {
				navPaddingTop: '0px',
				userInfo: {
					avatar: "",
					name: "",
					nickname: "",
					email: ""
				},
				isSaving: false,
				// 图片裁切相关
				showCropper: false,
				selectedImagePath: ''
			};
		},
		onLoad() {
			this.loadUserInfo();
		},
		methods: {
			// 处理导航栏高度
			handleNavHeightReady(navInfo) {
				this.navPaddingTop = navInfo.heightPx;
			},
			
			// 返回上一页
			goBack() {
				navigateBack();
			},
			
			// 加载用户信息
			async loadUserInfo() {
				try {
					uni.showLoading({
						title: '加载中...'
					});
					
					const result = await getUserInfo();
					if (result.success && result.result) {
						this.userInfo = {
							avatar: result.result.avatar || "",
							name: result.result.name || result.result.realname || "",
							nickname: result.result.nickname || "",
							email: result.result.email || ""
						};
					} else {
						showToast({
							title: result.message || '加载失败',
							icon: 'none'
						});
					}
				} catch (error) {
					console.error('加载用户信息失败:', error);
					showToast({
						title: '加载失败',
						icon: 'none'
					});
				} finally {
					uni.hideLoading();
				}
			},
			
			// 选择头像
			async selectAvatar() {
				try {
					// 使用统一的图片选择
					const filePaths = await new Promise((resolve, reject) => {
						uni.chooseImage({
							count: 1,
							sizeType: ['compressed'],
							sourceType: ['album', 'camera'],
							success: (res) => resolve(res.tempFilePaths),
							fail: reject
						});
					});

					if (filePaths && filePaths.length > 0) {
						await this.processAndUploadAvatar(filePaths[0]);
					}
				} catch (error) {
					console.error('选择图片失败:', error);
					showToast({
						title: '选择图片失败',
						icon: 'none'
					});
				}
			},

			// 处理并上传头像
			async processAndUploadAvatar(filePath) {
				try {
					// 获取图片尺寸信息
					const imageInfo = await getImageSize(filePath);
					console.log('图片信息:', imageInfo);

					// 如果不是正方形，显示裁切界面
					if (!imageInfo.isSquare) {
						this.selectedImagePath = filePath;
						this.showCropper = true;
					} else {
						// 如果已经是正方形，直接上传
						await this.uploadAvatar(filePath);
					}

				} catch (error) {
					console.error('处理图片失败:', error);
					showToast({
						title: '处理图片失败',
						icon: 'none'
					});
				}
			},

			// 处理裁切确认
			async handleCropConfirm(croppedImagePath) {
				try {
					this.showCropper = false;

					if (croppedImagePath) {
						// 上传裁切后的图片
						await this.uploadAvatar(croppedImagePath);

						showToast({
							title: '图片裁切完成',
							icon: 'success'
						});
					}
				} catch (error) {
					console.error('处理裁切图片失败:', error);
					showToast({
						title: '处理失败',
						icon: 'none'
					});
				}
			},

			// 处理裁切取消
			handleCropCancel() {
				this.showCropper = false;
				// 可以选择直接上传原图或者重新选择
				uni.showModal({
					title: '提示',
					content: '是否使用原图作为头像？',
					confirmText: '使用原图',
					cancelText: '重新选择',
					success: async (res) => {
						if (res.confirm) {
							// 使用原图
							await this.uploadAvatar(this.selectedImagePath);
						}
						// 如果取消，什么都不做，用户可以重新选择图片
					}
				});
			},



			// 上传头像
			async uploadAvatar(filePath) {
				try {
					// 验证图片
					await validateImage(filePath, 5 * 1024 * 1024); // 5MB限制

					// 使用统一的图片上传工具
					const result = await uploadSingleImage(filePath, {
						showLoading: true,
						compress: true,
						quality: 0.8
					});

					if (result.success && result.url) {
						this.userInfo.avatar = result.url;
						showToast({
							title: '头像上传成功',
							icon: 'success'
						});
					} else {
						throw new Error(result.error || '上传失败');
					}

				} catch (error) {
					console.error('上传头像失败:', error);
					showToast({
						title: error.message || '上传失败',
						icon: 'none'
					});
				}
			},
			
			// 保存个人信息
			async saveProfile() {
				if (this.isSaving) return;
				
				// 表单验证
				if (this.userInfo.email && !this.validateEmail(this.userInfo.email)) {
					return showToast({
						title: '邮箱格式不正确',
						icon: 'none'
					});
				}
				
				try {
					this.isSaving = true;
					
					const result = await updateUserInfo({
						avatar: this.userInfo.avatar,
						nickname: this.userInfo.nickname,
						email: this.userInfo.email
					});
					
					if (result.success) {
						showToast({
							title: '保存成功',
							icon: 'success'
						});
						
						// 返回上一页
						setTimeout(() => {
							navigateBack();
						}, 1500);
					} else {
						showToast({
							title: result.message || '保存失败',
							icon: 'none'
						});
					}
				} catch (error) {
					console.error('保存个人信息失败:', error);
					showToast({
						title: '保存失败',
						icon: 'none'
					});
				} finally {
					this.isSaving = false;
				}
			},
			
			// 验证邮箱格式
			validateEmail(email) {
				const emailRegex = /^[\w-]+(\.[\w-]+)*@[\w-]+(\.[\w-]+)+$/;
				return emailRegex.test(email);
			}
		}
	}
</script>

<style lang="scss" scoped>
.container {
	min-height: 100vh;
	background: var(--bg-secondary);
}

.content {
	padding: var(--spacing-md);
}

/* 头像上传区域 */
.avatar-section {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: var(--spacing-xl) 0;
	gap: var(--spacing-sm);
}

.avatar-wrapper {
	position: relative;
	width: 90px;
	height: 90px;
	border-radius: 50%;
	background-color: var(--bg-card);
	box-shadow: var(--shadow-md);
	border: 3px solid var(--bg-card);
}

.avatar-image {
	width: 100%;
	height: 100%;
	border-radius: 50%;
}

.avatar-overlay {
	position: absolute;
	right: 0;
	bottom: 0;
	width: 28px;
	height: 28px;
	background-color: var(--primary-color);
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	border: 2px solid var(--bg-card);
	box-shadow: var(--shadow-sm);
}

.avatar-hint {
	font-size: var(--font-size-xs);
	color: var(--text-tertiary);
	font-weight: 500;
}

/* 表单区域 */
.form-group {
	background-color: var(--bg-card);
	border-radius: var(--radius-xl);
	padding: 0 var(--spacing-md);
	box-shadow: var(--shadow-sm);
	border: 0.5px solid var(--border-primary);
	margin-bottom: var(--spacing-xl);
}

.form-item {
	display: flex;
	align-items: center;
	padding: var(--spacing-md) 0;
	min-height: 56px;
	box-sizing: border-box;

	&:not(:last-child) {
		border-bottom: 0.5px solid var(--border-secondary);
	}

	&__label {
		width: 80px;
		font-size: 15px;
		font-weight: 600;
		color: var(--text-primary);
		display: flex;
		align-items: center;
		gap: 6px;
	}
}

.form-item--email .form-item__label {
	color: #3b82f6;
}

.email-label-badge {
	width: 22px;
	height: 22px;
	border-radius: 6px;
	background: linear-gradient(135deg, #3b82f6, #60a5fa);
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
	box-shadow: 0 2rpx 6rpx rgba(59, 130, 246, 0.3);
}

.form-item__content {
	flex: 1;
	display: flex;
	align-items: center;
	justify-content: flex-end;
	text-align: right;

	&.readonly {
		color: var(--text-tertiary);
		gap: 4px;

		.form-item__content-text {
			font-size: 15px;
		}
	}
}

.form-item__content--email {
	background: rgba(59, 130, 246, 0.04);
	border: 1px solid rgba(59, 130, 246, 0.15);
	border-radius: 12rpx;
	padding: 4px 12rpx;
}

.form-input {
	width: 100%;
	font-size: 15px;
	color: var(--text-primary);
	text-align: right;
}

.form-input--email {
	color: #2563eb;
	font-weight: 600;
}

.placeholder {
	color: var(--text-tertiary);
}

/* 保存按钮 */
.footer-actions {
	padding: 0 var(--spacing-sm);
}

.save-btn {
	height: 52px;
	background-color: var(--primary-color);
	border-radius: var(--radius-lg);
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 8px 24px rgba(22, 93, 255, 0.2);
	transition: all 0.2s var(--ease-in-out);

	.save-btn-text {
		font-size: 16px;
		font-weight: 700;
		color: #fff;
	}

	&:active {
		transform: scale(0.99);
		opacity: 0.9;
	}

	&--loading {
		opacity: 0.7;
		pointer-events: none;
	}
}

.loading-spinner {
	width: 20px;
	height: 20px;
	border: 2px solid rgba(255, 255, 255, 0.3);
	border-top-color: #fff;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

@keyframes spin {
	to { transform: rotate(360deg); }
}
</style>
