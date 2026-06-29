<template>
	<view class="discover-page">
		<!-- 顶部蓝色渐变 Hero 区域 -->
		<view class="hero-section">
			<view class="hero-bg"></view>
			<image
				class="hero-bg-image"
				src="/static/images/index_bg.png"
				mode="aspectFill"
			></image>
			<view class="hero-bg-overlay"></view>

			<!-- 状态栏占位 -->
			<view class="hero-statusbar" :style="{ height: statusBarHeight + 'px' }"></view>

			<!-- 顶部导航栏 -->
			<view class="hero-nav">
				<view class="hero-nav-left">
					<text class="hero-brand">发现</text>
					<text class="hero-brand-en">DISCOVER</text>
				</view>
			</view>

			<!-- 欢迎语 -->
			<view class="hero-content">
				<view class="hero-greeting">
					<text class="hero-title">{{ greetingText }}</text>
					<text class="hero-subtitle">探索校园服务的每个角落</text>
				</view>
			</view>
		</view>

		<!-- 主体内容区域 -->
		<view class="content-area" :style="{paddingTop: '0'}">

			<!-- 功能卡片区 -->
			<view class="section-card">
				<view class="section-header">
					<view class="section-header-left">
						<text class="section-title-main">校园服务</text>
						<text class="section-title-sub">CAMPUS SERVICES</text>
					</view>
				</view>

				<!-- 功能卡片 grid -->
				<view class="service-grid">
					<view class="service-item"
						hover-class="service-item-hover"
						@tap="goToAttendance">
						<view class="service-icon quick-icon--attendance">
							<l-icon name="alarm" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<view class="service-text">
							<text class="service-label">考勤记录</text>
							<text class="service-desc">查看考勤情况</text>
						</view>
					</view>
					<view class="service-item"
						hover-class="service-item-hover"
						@tap="goToOptionalCourse">
						<view class="service-icon quick-icon--course">
							<l-icon name="bookmark" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<view class="service-text">
							<text class="service-label">选课记录</text>
							<text class="service-desc">选修课详情</text>
						</view>
					</view>
					<view class="service-item"
						hover-class="service-item-hover"
						@tap="goToLeave">
						<view class="service-icon quick-icon--leave">
							<l-icon name="calendar" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<view class="service-text">
							<text class="service-label">我的请假</text>
							<text class="service-desc">请假申请记录</text>
						</view>
					</view>
					<view class="service-item"
						hover-class="service-item-hover"
						@tap="goToSurvey">
						<view class="service-icon quick-icon--survey">
							<l-icon name="edit" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<view class="service-text">
							<text class="service-label">调查问卷</text>
							<text class="service-desc">填写校园问卷</text>
						</view>
					</view>
					<view class="service-item"
						hover-class="service-item-hover"
						@tap="goToProfile">
						<view class="service-icon quick-icon--profile">
							<l-icon name="user" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<view class="service-text">
							<text class="service-label">个人信息</text>
							<text class="service-desc">查看详细档案</text>
						</view>
					</view>
					<view class="service-item"
						hover-class="service-item-hover"
						@tap="goToCoursePlan">
						<view class="service-icon quick-icon--courseplan">
							<l-icon name="address-book" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<view class="service-text">
							<text class="service-label">课程计划</text>
							<text class="service-desc">培养方案查看</text>
						</view>
					</view>
					<view class="service-item"
						hover-class="service-item-hover"
						@tap="goToEvaluation">
						<view class="service-icon quick-icon--evaluation">
							<l-icon name="chart-radial" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<view class="service-text">
							<text class="service-label">学生评教</text>
							<text class="service-desc">评教待办提醒</text>
						</view>
					</view>
					<view class="service-item"
						hover-class="service-item-hover"
						@tap="goToBankCard">
						<view class="service-icon quick-icon--bank">
							<l-icon name="creditcard" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<view class="service-text">
							<text class="service-label">我的银行卡</text>
							<text class="service-desc">账户信息查看</text>
						</view>
					</view>
					<view class="service-item"
						hover-class="service-item-hover"
						@tap="goToRepair">
						<view class="service-icon quick-icon--repair">
							<l-icon name="brush" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<view class="service-text">
							<text class="service-label">报修管理</text>
							<text class="service-desc">设施报修记录</text>
						</view>
					</view>
					<view class="service-item"
						hover-class="service-item-hover"
						@tap="goToScoreNotification">
						<view class="service-icon quick-icon--scorenotify">
							<l-icon name="chart-line" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<view class="service-text">
							<text class="service-label">成绩通知</text>
							<text class="service-desc">配置成绩主动查询与通知渠道</text>
						</view>
					</view>
					<view class="service-item"
						hover-class="service-item-hover"
						@tap="goToMyTeachers">
						<view class="service-icon quick-icon--teachers">
							<l-icon name="user-business" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<view class="service-text">
							<text class="service-label">我的老师</text>
							<text class="service-desc">各学期任课老师及联系方式</text>
						</view>
					</view>
					<view class="service-item"
						hover-class="service-item-hover"
						@tap="goToMissClass">
						<view class="service-icon quick-icon--missclass">
							<l-icon name="alarm-off" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<view class="service-text">
							<text class="service-label">缺课统计</text>
							<text class="service-desc">各课程缺课节数统计</text>
						</view>
					</view>
					<view class="service-item"
						hover-class="service-item-hover"
						@tap="goToMakeupExam">
						<view class="service-icon quick-icon--makeup">
							<l-icon name="collection-1" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<view class="service-text">
							<text class="service-label">补考查询</text>
							<text class="service-desc">不及格课程补考成绩查询</text>
						</view>
					</view>
					<view class="service-item"
						hover-class="service-item-hover"
						@tap="goToDisciplinary">
						<view class="service-icon quick-icon--disciplinary">
							<l-icon name="shield-error" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<view class="service-text">
							<text class="service-label">违纪查询</text>
							<text class="service-desc">违纪处分记录查询</text>
						</view>
					</view>
					<view class="service-item"
						hover-class="service-item-hover"
						@tap="goToInternship">
						<view class="service-icon quick-icon--repair">
							<l-icon name="work" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<view class="service-text">
							<text class="service-label">实习管理</text>
							<text class="service-desc">实习岗位、申请与成绩</text>
						</view>
					</view>
				</view>
			</view>

			<!-- 底部安全区 -->
			<view class="bottom-safe-area"></view>
		</view>
	</view>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const statusBarHeight = ref(20)

function getGreetingText() {
	const hour = new Date().getHours()
	if (hour < 6) return '凌晨好'
	if (hour < 9) return '早上好'
	if (hour < 12) return '上午好'
	if (hour < 14) return '中午好'
	if (hour < 18) return '下午好'
	if (hour < 22) return '晚上好'
	return '夜深了'
}

const greetingText = getGreetingText()

function initStatusBarHeight() {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 20
}

function goToAttendance() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/discover/attendance' })
}

function goToOptionalCourse() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/discover/optional-course' })
}

function goToLeave() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/discover/leave' })
}

function goToSurvey() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/discover/survey' })
}

function goToProfile() {
	uni.vibrateShort()
	const stored = uni.getStorageSync('userInfo')
	let url = '/pages/discover/profile'
	if (stored) {
		const info = typeof stored === 'string' ? JSON.parse(stored) : stored
		if (info?.id) url += `?userId=${info.id}`
	}
	uni.navigateTo({ url })
}

function goToCoursePlan() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/discover/course-plan' })
}

function goToEvaluation() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/evaluation/list' })
}

function goToRepair() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/discover/repair' })
}

function goToBankCard() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/discover/bank-card' })
}

function goToMyTeachers() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/discover/my-teachers' })
}

function goToScoreNotification() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/discover/score-notification' })
}

function goToArchiveEdit() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/discover/archive-edit' })
}

function goToMissClass() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/discover/miss-class' })
}

function goToMakeupExam() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/discover/makeup-exam' })
}

function goToDisciplinary() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/discover/disciplinary' })
}

function goToInternship() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/discover/internship/index' })
}

initStatusBarHeight()
</script>

<style lang="scss" scoped>
@import url('../../static/css/home-page.css');

.discover-page {
	flex: 1;
	width: 100%;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
	position: relative;
}

.hero-section {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 320rpx;
}

.hero-bg {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: linear-gradient(180deg,
		#1e3a8a 0%,
		#1e40af 25%,
		#2563eb 55%,
		#3b82f6 78%,
		#93c5fd 100%);
	z-index: 0;
}

.hero-bg-image {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	width: 100%;
	height: 100%;
	z-index: 1;
	opacity: 0.6;
}

.hero-bg-overlay {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.7) 0%,
		rgba(37, 99, 235, 0.45) 50%,
		rgba(147, 197, 253, 0.15) 100%);
	z-index: 2;
}

.hero-statusbar {
	width: 100%;
	flex-shrink: 0;
}

.hero-nav {
	position: relative;
	z-index: 10;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.hero-nav-left {
	flex: 1;
}

.hero-brand {
	display: block;
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 1px;
	line-height: 1.2;
}

.hero-brand-en {
	display: block;
	font-size: 20rpx;
	font-weight: 500;
	color: rgba(255, 255, 255, 0.65);
	letter-spacing: 2px;
	margin-top: 2rpx;
}

.hero-content {
	position: relative;
	z-index: 10;
	padding: 8rpx 32rpx 24rpx;
}

.hero-greeting {
	margin-bottom: 0;
}

.hero-title {
	display: block;
	font-size: 44rpx;
	font-weight: 800;
	color: #fff;
	line-height: 1.3;
}

.hero-subtitle {
	display: block;
	font-size: 26rpx;
	color: rgba(255, 255, 255, 0.75);
	margin-top: 6rpx;
	font-weight: 400;
}

.content-area {
	position: relative;
	z-index: 1;
	padding: 0 24rpx;
	padding-bottom: calc(48rpx + env(safe-area-inset-bottom));
	margin-top: -40rpx;
}

.service-grid {
	display: grid;
	grid-template-columns: repeat(2, 1fr);
	gap: 20rpx;
}

.service-item {
	background: #fff;
	border-radius: 24rpx;
	padding: 28rpx 24rpx;
	display: flex;
	flex-direction: row;
	align-items: center;
	box-shadow: 0 2rpx 12rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	transition: all 0.18s ease;
	overflow: hidden;
}

.service-item:active {
	transform: scale(0.96);
	box-shadow: 0 1rpx 6rpx rgba(30, 64, 175, 0.05);
}

.service-icon {
	width: 80rpx;
	height: 80rpx;
	border-radius: 22rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	margin-right: 20rpx;
	flex-shrink: 0;
}

.service-text {
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	flex: 1;
}

.service-label {
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	margin-bottom: 6rpx;
}

.service-desc {
	font-size: 22rpx;
	color: var(--text-tertiary);
	line-height: 1.4;
}

.service-item-hover {
	opacity: 0.85;
}

.quick-icon--attendance {
	background: linear-gradient(135deg, #3b82f6, #60a5fa);
	box-shadow: 0 6rpx 20rpx rgba(59, 130, 246, 0.3);
}

.quick-icon--course {
	background: linear-gradient(135deg, #10b981, #34d399);
	box-shadow: 0 6rpx 20rpx rgba(16, 185, 129, 0.3);
}

.quick-icon--leave {
	background: linear-gradient(135deg, #f59e0b, #fbbf24);
	box-shadow: 0 6rpx 20rpx rgba(245, 158, 11, 0.3);
}

.quick-icon--survey {
	background: linear-gradient(135deg, #8b5cf6, #a78bfa);
	box-shadow: 0 6rpx 20rpx rgba(139, 92, 246, 0.3);
}

.quick-icon--profile {
	background: linear-gradient(135deg, #06b6d4, #22d3ee);
	box-shadow: 0 6rpx 20rpx rgba(6, 182, 212, 0.3);
}

.quick-icon--courseplan {
	background: linear-gradient(135deg, #ec4899, #f472b6);
	box-shadow: 0 6rpx 20rpx rgba(236, 72, 153, 0.3);
}

.quick-icon--evaluation {
	background: linear-gradient(135deg, #f59e0b, #fbbf24);
	box-shadow: 0 6rpx 20rpx rgba(245, 158, 11, 0.3);
}

.quick-icon--repair {
	background: linear-gradient(135deg, #64748b, #94a3b8);
	box-shadow: 0 6rpx 20rpx rgba(100, 116, 139, 0.3);
}

.quick-icon--bank {
	background: linear-gradient(135deg, #10b981, #34d399);
	box-shadow: 0 6rpx 20rpx rgba(16, 185, 129, 0.3);
}

.quick-icon--teachers {
	background: linear-gradient(135deg, #f59e0b, #fbbf24);
	box-shadow: 0 6rpx 20rpx rgba(245, 158, 11, 0.3);
}

.quick-icon--scorenotify {
	background: linear-gradient(135deg, #8b5cf6, #a78bfa);
	box-shadow: 0 6rpx 20rpx rgba(139, 92, 246, 0.3);
}

.quick-icon--missclass {
	background: linear-gradient(135deg, #ef4444, #f97316);
	box-shadow: 0 6rpx 20rpx rgba(239, 68, 68, 0.3);
}

.quick-icon--makeup {
	background: linear-gradient(135deg, #8b5cf6, #a78bfa);
	box-shadow: 0 6rpx 20rpx rgba(139, 92, 246, 0.3);
}

.quick-icon--disciplinary {
	background: linear-gradient(135deg, #64748b, #94a3b8);
	box-shadow: 0 6rpx 20rpx rgba(100, 116, 139, 0.3);
}

.section-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 0 20rpx;
}

.section-header-right {
	display: flex;
	align-items: center;
}

.section-header-action {
	font-size: 24rpx;
	font-weight: 600;
	padding: 4rpx 16rpx;
	border-radius: 20rpx;
}

.action--success {
	background: rgba(16, 185, 129, 0.1);
	color: #10b981;
}

.action--warn {
	background: rgba(245, 158, 11, 0.1);
	color: #f59e0b;
}

.bottom-safe-area {
	height: env(safe-area-inset-bottom);
}
</style>
