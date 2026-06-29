<template>
	<view class="internship-page">
		<!-- 自定义导航栏 -->
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">实习管理</text>
				<view class="nav-bar__placeholder"></view>
			</view>
		</view>

		<!-- 内容区 -->
		<scroll-view class="content-scroll" scroll-y>
			<view class="module-grid">
				<!-- 第一行 -->
				<view class="module-item" hover-class="module-item-hover" @tap="goToRequirements">
					<view class="module-icon module-icon--requirements">
						<l-icon name="tips" style="font-size: 24px; color: #fff;"></l-icon>
					</view>
					<text class="module-label">实习要求</text>
					<text class="module-desc">查看实习相关规定</text>
				</view>
				<view class="module-item" hover-class="module-item-hover" @tap="goToPosition">
					<view class="module-icon module-icon--position">
						<l-icon name="certificate-1" style="font-size: 24px; color: #fff;"></l-icon>
					</view>
					<text class="module-label">实习岗位</text>
					<text class="module-desc">搜索和浏览实习岗位</text>
				</view>
				<!-- 第二行 -->
				<view class="module-item" hover-class="module-item-hover" @tap="goToApplication">
					<view class="module-icon module-icon--application">
						<l-icon name="file-edit" style="font-size: 24px; color: #fff;"></l-icon>
					</view>
					<text class="module-label">我的申请</text>
					<text class="module-desc">管理我的实习申请</text>
				</view>
				<view class="module-item" hover-class="module-item-hover" @tap="goToSignin">
					<view class="module-icon module-icon--signin">
						<l-icon name="check-rectangle" style="font-size: 24px; color: #fff;"></l-icon>
					</view>
					<text class="module-label">实习签到</text>
					<text class="module-desc">每日实习打卡记录</text>
				</view>
				<!-- 第三行 -->
				<view class="module-item" hover-class="module-item-hover" @tap="goToPlan">
					<view class="module-icon module-icon--plan">
						<l-icon name="address-book" style="font-size: 24px; color: #fff;"></l-icon>
					</view>
					<text class="module-label">实习计划</text>
					<text class="module-desc">我的实习计划安排</text>
				</view>
				<view class="module-item" hover-class="module-item-hover" @tap="goToSummary">
					<view class="module-icon module-icon--summary">
						<l-icon name="summary" style="font-size: 24px; color: #fff;"></l-icon>
					</view>
					<text class="module-label">实习总结</text>
					<text class="module-desc">撰写和查看实习总结</text>
				</view>
				<!-- 第四行 -->
				<view class="module-item" hover-class="module-item-hover" @tap="goToAppraisal">
					<view class="module-icon module-icon--appraisal">
						<l-icon name="certificate" style="font-size: 24px; color: #fff;"></l-icon>
					</view>
					<text class="module-label">实习鉴定表</text>
					<text class="module-desc">查看实习鉴定信息</text>
				</view>
				<view class="module-item" hover-class="module-item-hover" @tap="goToScore">
					<view class="module-icon module-icon--score">
						<l-icon name="chart-line" style="font-size: 24px; color: #fff;"></l-icon>
					</view>
					<text class="module-label">实习成绩</text>
					<text class="module-desc">查询实习成绩信息</text>
				</view>
			</view>

			<view class="bottom-safe-area"></view>
		</scroll-view>
	</view>
</template>

<script>
export default {
	data() {
		return {
			statusBarHeight: 20
		}
	},
	onLoad() {
		const systemInfo = uni.getSystemInfoSync()
		this.statusBarHeight = systemInfo.statusBarHeight || 20
	},
	methods: {
		goBack() { uni.navigateBack() },
		goToRequirements() {
			uni.vibrateShort()
			uni.navigateTo({ url: '/pages/discover/internship/requirements' })
		},
		goToPosition() {
			uni.vibrateShort()
			uni.navigateTo({ url: '/pages/discover/internship/position' })
		},
		goToApplication() {
			uni.vibrateShort()
			uni.navigateTo({ url: '/pages/discover/internship/application' })
		},
		goToSignin() {
			uni.vibrateShort()
			uni.navigateTo({ url: '/pages/discover/internship/signin' })
		},
		goToPlan() {
			uni.vibrateShort()
			uni.navigateTo({ url: '/pages/discover/internship/plan' })
		},
		goToSummary() {
			uni.vibrateShort()
			uni.navigateTo({ url: '/pages/discover/internship/summary' })
		},
		goToAppraisal() {
			uni.vibrateShort()
			uni.navigateTo({ url: '/pages/discover/internship/appraisal' })
		},
		goToScore() {
			uni.vibrateShort()
			uni.navigateTo({ url: '/pages/discover/internship/score' })
		}
	}
}
</script>

<style lang="scss" scoped>
.internship-page {
	flex: 1;
	width: 100%;
	display: flex;
	flex-direction: column;
	background-color: #f0f2f5;
	min-height: 100vh;
}

.nav-bar {
	background: linear-gradient(135deg, #1e40af, #3b82f6);
}

.nav-bar__content {
	display: flex;
	align-items: center;
	justify-content: space-between;
	height: 88rpx;
	padding: 0 24rpx;
}

.nav-bar__back {
	width: 60rpx;
	height: 60rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.nav-bar__title {
	font-size: 34rpx;
	font-weight: 700;
	color: #fff;
}

.nav-bar__placeholder {
	width: 60rpx;
}

.content-scroll {
	flex: 1;
	padding: 24rpx;
	box-sizing: border-box;
}

.module-grid {
	display: grid;
	grid-template-columns: repeat(2, 1fr);
	gap: 20rpx;
}

.module-item {
	background: #fff;
	border-radius: 24rpx;
	padding: 32rpx 24rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	box-shadow: 0 2rpx 12rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	transition: all 0.18s ease;
}

.module-item:active {
	transform: scale(0.96);
	box-shadow: 0 1rpx 6rpx rgba(30, 64, 175, 0.05);
}

.module-icon {
	width: 96rpx;
	height: 96rpx;
	border-radius: 24rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 20rpx;
}

.module-label {
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	margin-bottom: 8rpx;
}

.module-desc {
	font-size: 22rpx;
	color: var(--text-tertiary);
	text-align: center;
	line-height: 1.4;
}

.module-icon--requirements {
	background: linear-gradient(135deg, #3b82f6, #60a5fa);
	box-shadow: 0 6rpx 20rpx rgba(59, 130, 246, 0.3);
}

.module-icon--position {
	background: linear-gradient(135deg, #10b981, #34d399);
	box-shadow: 0 6rpx 20rpx rgba(16, 185, 129, 0.3);
}

.module-icon--application {
	background: linear-gradient(135deg, #f59e0b, #fbbf24);
	box-shadow: 0 6rpx 20rpx rgba(245, 158, 11, 0.3);
}

.module-icon--signin {
	background: linear-gradient(135deg, #ef4444, #f97316);
	box-shadow: 0 6rpx 20rpx rgba(239, 68, 68, 0.3);
}

.module-icon--plan {
	background: linear-gradient(135deg, #8b5cf6, #a78bfa);
	box-shadow: 0 6rpx 20rpx rgba(139, 92, 246, 0.3);
}

.module-icon--summary {
	background: linear-gradient(135deg, #06b6d4, #22d3ee);
	box-shadow: 0 6rpx 20rpx rgba(6, 182, 212, 0.3);
}

.module-icon--appraisal {
	background: linear-gradient(135deg, #ec4899, #f472b6);
	box-shadow: 0 6rpx 20rpx rgba(236, 72, 153, 0.3);
}

.module-icon--score {
	background: linear-gradient(135deg, #f59e0b, #fbbf24);
	box-shadow: 0 6rpx 20rpx rgba(245, 158, 11, 0.3);
}

.bottom-safe-area {
	height: 48rpx;
}
</style>
