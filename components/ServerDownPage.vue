<!-- 服务器关闭状态页面组件 -->
<template>
	<view class="server-down-container">
		<!-- 主要内容区域 -->
		<view class="content-wrapper">
			<!-- 图标区域 -->
			<view class="icon-container">
				<view class="server-icon">
					<!-- 服务器图标 -->
					<view class="server-body">
						<view class="server-light" :class="{active: isBlinking}"></view>
						<view class="server-screen">
							<view class="error-symbol">×</view>
						</view>
					</view>
					<!-- 连接线 -->
					<view class="connection-lines">
						<view class="line line-1" :class="{disconnected: true}"></view>
						<view class="line line-2" :class="{disconnected: true}"></view>
						<view class="line line-3" :class="{disconnected: true}"></view>
					</view>
				</view>
			</view>
			
			<!-- 标题和描述 -->
			<view class="text-content">
				<text class="main-title">服务器暂时无法连接</text>
				<text class="sub-title">学校服务器可能正在维护中</text>
				<text class="description">请稍后再试，或联系管理员获取帮助</text>
			</view>
			
			<!-- 错误信息展示（可选） -->
			<view class="error-details" v-if="showDetails && errorMessage">
				<view class="error-header" @click="toggleDetails">
					<text class="error-title">错误详情</text>
					<text class="toggle-icon" :class="{expanded: detailsExpanded}">▼</text>
				</view>
				<view class="error-content" v-if="detailsExpanded">
					<text class="error-message">{{formatErrorMessage(errorMessage)}}</text>
				</view>
			</view>
			
			<!-- 操作按钮 -->
			<view class="action-buttons">
				<button class="retry-btn" @click="handleRetry" :disabled="isRetrying">
					<view class="btn-content">
						<view class="retry-icon" :class="{spinning: isRetrying}">↻</view>
						<text class="btn-text">{{isRetrying ? '重试中...' : '重新尝试'}}</text>
					</view>
				</button>

				<button class="back-btn" @click="handleGoBack" v-if="showBackButton">
					<text class="btn-text">返回首页</text>
				</button>

				<button class="suppress-btn" @click="handleSuppressFor24h" :disabled="isSuppressing">
					<view class="btn-content">
						<text class="suppress-icon">🔕</text>
						<text class="btn-text">{{isSuppressing ? '设置中...' : '24小时内不再提示'}}</text>
					</view>
				</button>
			</view>
			
			<!-- 提示信息 -->
			<view class="tips">
				<view class="tip-item">
					<text class="tip-icon">💡</text>
					<text class="tip-text">请检查网络连接是否正常</text>
				</view>
				<view class="tip-item">
					<text class="tip-icon">⏰</text>
					<text class="tip-text">服务器可能正在维护，请稍后再试</text>
				</view>
			</view>
		</view>
		
		<!-- 背景装饰 -->
		<view class="background-decoration">
			<view class="floating-dot dot-1"></view>
			<view class="floating-dot dot-2"></view>
			<view class="floating-dot dot-3"></view>
			<view class="floating-dot dot-4"></view>
		</view>
	</view>
</template>

<script>
export default {
	name: 'ServerDownPage',
	props: {
		// 错误信息
		errorMessage: {
			type: String,
			default: ''
		},
		// 是否显示错误详情
		showDetails: {
			type: Boolean,
			default: false
		},
		// 是否显示返回按钮
		showBackButton: {
			type: Boolean,
			default: true
		},
		// 自定义重试函数
		retryFunction: {
			type: Function,
			default: null
		}
	},
	data() {
		return {
			isBlinking: false,
			detailsExpanded: false,
			isRetrying: false,
			isSuppressing: false,
			blinkTimer: null
		};
	},
	mounted() {
		// 启动服务器指示灯闪烁动画
		this.startBlinking();
	},
	beforeDestroy() {
		// 清理定时器
		if (this.blinkTimer) {
			clearInterval(this.blinkTimer);
		}
	},
	methods: {
		// 启动闪烁动画
		startBlinking() {
			this.blinkTimer = setInterval(() => {
				this.isBlinking = !this.isBlinking;
			}, 1500);
		},
		
		// 切换错误详情显示
		toggleDetails() {
			this.detailsExpanded = !this.detailsExpanded;
		},
		
		// 格式化错误信息
		formatErrorMessage(message) {
			if (!message) return '';
			
			// 尝试解析JSON格式的错误信息
			try {
				const errorObj = JSON.parse(message);
				if (errorObj.message) {
					return errorObj.message;
				}
			} catch (e) {
				// 如果不是JSON格式，直接返回原始信息
			}
			
			// 简化超长的错误信息
			if (message.length > 200) {
				return message.substring(0, 200) + '...';
			}
			
			return message;
		},
		
		// 处理重试
		async handleRetry() {
			if (this.isRetrying) return;

			this.isRetrying = true;

			try {
				// 设置重试标志，告诉页面这是用户主动重试
				uni.setStorageSync('userRetryFlag', 'true')
				uni.setStorageSync('userRetryTimestamp', Date.now().toString())

				if (this.retryFunction && typeof this.retryFunction === 'function') {
					// 使用自定义重试函数
					await this.retryFunction();
				} else {
					// 默认重试行为：触发事件让父组件处理
					this.$emit('retry');
				}
			} catch (error) {
				console.error('重试失败:', error);
				uni.showToast({
					title: '重试失败，请稍后再试',
					icon: 'none'
				});
			} finally {
				// 延迟重置状态，给用户反馈
				setTimeout(() => {
					this.isRetrying = false;
				}, 1000);
			}
		},
		
		// 返回首页
		handleGoBack() {
			try {
				// 清除重试标志，表示用户选择返回而不是重试
				uni.removeStorageSync('userRetryFlag')
				uni.removeStorageSync('userRetryTimestamp')

				uni.switchTab({
					url: '/pages/index/index',
					fail: () => {
						// 如果switchTab失败，尝试navigateTo
						uni.navigateTo({
							url: '/pages/index/index',
							fail: () => {
								// 如果都失败，使用reLaunch
								uni.reLaunch({
									url: '/pages/index/index'
								});
							}
						});
					}
				});
			} catch (error) {
				console.error('导航失败:', error);
				this.$emit('goBack');
			}
		},

		// 处理24小时内不再提示
		async handleSuppressFor24h() {
			if (this.isSuppressing) return;

			this.isSuppressing = true;

			try {
				// 设置24小时抑制标志
				const suppressUntil = Date.now() + 24 * 60 * 60 * 1000; // 24小时后
				uni.setStorageSync('errorSuppressUntil', suppressUntil.toString());

				console.log('设置24小时内不再显示错误提示，抑制到:', new Date(suppressUntil));

				// 显示确认提示
				uni.showToast({
					title: '已设置24小时内不再提示',
					icon: 'success',
					duration: 2000
				});

				// 延迟后返回首页
				setTimeout(() => {
					this.handleGoBack();
				}, 2000);

			} catch (error) {
				console.error('设置抑制标志失败:', error);
				uni.showToast({
					title: '设置失败，请重试',
					icon: 'none',
					duration: 2000
				});
			} finally {
				setTimeout(() => {
					this.isSuppressing = false;
				}, 2000);
			}
		}
	}
};
</script>

<style lang="scss" scoped>
.server-down-container {
	width: 100%;
	min-height: 100vh;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 40rpx 30rpx;
	background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	position: relative;
	overflow: hidden;
	box-sizing: border-box;
}

.content-wrapper {
	display: flex;
	flex-direction: column;
	align-items: center;
	width: 100%;
	max-width: 90%;
	background: rgba(255, 255, 255, 0.95);
	border-radius: 24rpx;
	padding: 60rpx 40rpx;
	box-shadow: 0 20rpx 60rpx rgba(0, 0, 0, 0.1);
	backdrop-filter: blur(10rpx);
	z-index: 10;
	box-sizing: border-box;
}

/* 图标区域 */
.icon-container {
	margin-bottom: 40rpx;
	position: relative;
}

.server-icon {
	display: flex;
	flex-direction: column;
	align-items: center;
}

.server-body {
	width: 120rpx;
	height: 80rpx;
	background: linear-gradient(145deg, #f0f0f0, #d1d1d1);
	border-radius: 12rpx;
	position: relative;
	box-shadow: 
		0 8rpx 16rpx rgba(0, 0, 0, 0.1),
		inset 0 2rpx 4rpx rgba(255, 255, 255, 0.8);
	display: flex;
	align-items: center;
	justify-content: center;
}

.server-light {
	position: absolute;
	top: 12rpx;
	right: 12rpx;
	width: 16rpx;
	height: 16rpx;
	border-radius: 50%;
	background: #ff4757;
	box-shadow: 0 0 10rpx rgba(255, 71, 87, 0.6);
	transition: all 0.3s ease;
	
	&.active {
		background: #ff6b7a;
		box-shadow: 0 0 20rpx rgba(255, 71, 87, 0.8);
		transform: scale(1.2);
	}
}

.server-screen {
	width: 60rpx;
	height: 40rpx;
	background: #2c3e50;
	border-radius: 4rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	position: relative;
	overflow: hidden;
}

.error-symbol {
	font-size: 24rpx;
	color: #ff4757;
	font-weight: bold;
	animation: pulse 2s infinite;
}

.connection-lines {
	display: flex;
	flex-direction: column;
	gap: 8rpx;
	margin-top: 20rpx;
	align-items: center;
}

.line {
	height: 4rpx;
	background: #3498db;
	border-radius: 2rpx;
	transition: all 0.3s ease;
	
	&.line-1 { width: 60rpx; }
	&.line-2 { width: 40rpx; }
	&.line-3 { width: 20rpx; }
	
	&.disconnected {
		background: #95a5a6;
		animation: disconnect 1.5s infinite;
	}
}

/* 文本内容 */
.text-content {
	display: flex;
	flex-direction: column;
	align-items: center;
	text-align: center;
	margin-bottom: 40rpx;
}

.main-title {
	font-size: 48rpx;
	font-weight: bold;
	color: #2c3e50;
	margin-bottom: 16rpx;
	line-height: 1.2;
}

.sub-title {
	font-size: 32rpx;
	color: #7f8c8d;
	margin-bottom: 12rpx;
	font-weight: 500;
}

.description {
	font-size: 28rpx;
	color: #95a5a6;
	line-height: 1.4;
	max-width: 400rpx;
}

/* 错误详情 */
.error-details {
	width: 100%;
	margin-bottom: 30rpx;
	border: 2rpx solid #e74c3c;
	border-radius: 12rpx;
	overflow: hidden;
}

.error-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 20rpx;
	background: rgba(231, 76, 60, 0.1);
	cursor: pointer;
}

.error-title {
	font-size: 28rpx;
	color: #e74c3c;
	font-weight: 500;
}

.toggle-icon {
	font-size: 24rpx;
	color: #e74c3c;
	transition: transform 0.3s ease;
	
	&.expanded {
		transform: rotate(180deg);
	}
}

.error-content {
	padding: 20rpx;
	background: rgba(231, 76, 60, 0.05);
}

.error-message {
	font-size: 24rpx;
	color: #7f8c8d;
	line-height: 1.4;
	word-break: break-all;
}

/* 操作按钮 */
.action-buttons {
	display: flex;
	flex-direction: column;
	gap: 20rpx;
	width: 100%;
	margin-bottom: 30rpx;
}

.retry-btn, .back-btn, .suppress-btn {
	width: 100%;
	height: 88rpx;
	border-radius: 44rpx;
	border: none;
	font-size: 32rpx;
	font-weight: 500;
	transition: all 0.3s ease;
	position: relative;
	overflow: hidden;
}

.retry-btn {
	background: linear-gradient(135deg, #3498db, #2980b9);
	color: white;
	box-shadow: 0 8rpx 20rpx rgba(52, 152, 219, 0.3);
	
	&:active {
		transform: translateY(2rpx);
		box-shadow: 0 4rpx 12rpx rgba(52, 152, 219, 0.4);
	}
	
	&:disabled {
		opacity: 0.7;
		transform: none;
	}
}

.back-btn {
	background: linear-gradient(135deg, #95a5a6, #7f8c8d);
	color: white;
	box-shadow: 0 8rpx 20rpx rgba(149, 165, 166, 0.3);

	&:active {
		transform: translateY(2rpx);
		box-shadow: 0 4rpx 12rpx rgba(149, 165, 166, 0.4);
	}
}

.suppress-btn {
	background: linear-gradient(135deg, #f39c12, #e67e22);
	color: white;
	box-shadow: 0 8rpx 20rpx rgba(243, 156, 18, 0.3);

	&:active {
		transform: translateY(2rpx);
		box-shadow: 0 4rpx 12rpx rgba(243, 156, 18, 0.4);
	}

	&:disabled {
		opacity: 0.7;
		transform: none;
	}
}

.btn-content {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 12rpx;
}

.retry-icon {
	font-size: 32rpx;
	transition: transform 0.3s ease;

	&.spinning {
		animation: spin 1s linear infinite;
	}
}

.suppress-icon {
	font-size: 28rpx;
	flex-shrink: 0;
}

.btn-text {
	font-size: 32rpx;
}

/* 提示信息 */
.tips {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
	width: 100%;
}

.tip-item {
	display: flex;
	align-items: center;
	gap: 12rpx;
	padding: 16rpx 20rpx;
	background: rgba(52, 152, 219, 0.1);
	border-radius: 12rpx;
	border-left: 4rpx solid #3498db;
}

.tip-icon {
	font-size: 28rpx;
	flex-shrink: 0;
}

.tip-text {
	font-size: 24rpx;
	color: #7f8c8d;
	line-height: 1.3;
	flex: 1;
}

/* 背景装饰 */
.background-decoration {
	position: absolute;
	top: 0;
	left: 0;
	width: 100%;
	height: 100%;
	pointer-events: none;
	z-index: 1;
}

.floating-dot {
	position: absolute;
	width: 20rpx;
	height: 20rpx;
	background: rgba(255, 255, 255, 0.2);
	border-radius: 50%;
	animation: float 6s ease-in-out infinite;
}

.dot-1 {
	top: 20%;
	left: 10%;
	animation-delay: 0s;
}

.dot-2 {
	top: 60%;
	right: 15%;
	animation-delay: 2s;
}

.dot-3 {
	bottom: 30%;
	left: 20%;
	animation-delay: 4s;
}

.dot-4 {
	top: 40%;
	right: 30%;
	animation-delay: 1s;
}

/* 动画定义 */
@keyframes pulse {
	0%, 100% { opacity: 1; }
	50% { opacity: 0.3; }
}

@keyframes disconnect {
	0%, 100% { opacity: 0.3; }
	50% { opacity: 0.8; }
}

@keyframes spin {
	0% { transform: rotate(0deg); }
	100% { transform: rotate(360deg); }
}

@keyframes float {
	0%, 100% {
		transform: translateY(0px) rotate(0deg);
		opacity: 0.2;
	}
	50% {
		transform: translateY(-20rpx) rotate(180deg);
		opacity: 0.5;
	}
}

/* 响应式设计 */
@media screen and (max-width: 600rpx) {
	.content-wrapper {
		padding: 40rpx 30rpx;
		margin: 20rpx;
	}
	
	.main-title {
		font-size: 40rpx;
	}
	
	.sub-title {
		font-size: 28rpx;
	}
	
	.description {
		font-size: 24rpx;
	}
	
	.retry-btn, .back-btn {
		height: 80rpx;
		font-size: 28rpx;
	}
}

/* 深色模式适配 */
@media (prefers-color-scheme: dark) {
	.server-down-container {
		background: linear-gradient(135deg, #2c3e50 0%, #34495e 100%);
	}
	
	.content-wrapper {
		background: rgba(44, 62, 80, 0.95);
		color: #ecf0f1;
	}
	
	.main-title {
		color: #ecf0f1;
	}
	
	.sub-title {
		color: #bdc3c7;
	}
	
	.description {
		color: #95a5a6;
	}
	
	.server-body {
		background: linear-gradient(145deg, #34495e, #2c3e50);
	}
	
	.server-screen {
		background: #1a252f;
	}
}
</style>