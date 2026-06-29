<template>
	<view class="profile-page">
		<!-- 自定义导航栏 -->
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">学生档案</text>
				<view class="nav-bar__action" @tap="goToEdit">
					<text class="nav-bar__edit">编辑</text>
				</view>
			</view>
		</view>

		<!-- 内容区 -->
		<scroll-view class="content-scroll" scroll-y>
			<!-- 加载中 -->
			<view v-if="loading" class="state-container">
				<view class="state-spinner"></view>
				<text class="state-text">加载中...</text>
			</view>

			<!-- 错误状态 -->
			<view v-else-if="error" class="state-container state-container--error">
				<view class="state-icon-circle state-icon-circle--error">!</view>
				<text class="state-text">{{ error }}</text>
				<button class="state-btn" @tap="fetchProfile">重试</button>
			</view>

			<!-- 个人信息卡片 -->
			<view v-else class="profile-content">
				<!-- 头像和姓名 -->
				<view class="profile-header">
					<view class="profile-avatar">
						<text class="profile-avatar__text">{{ avatarText }}</text>
					</view>
					<text class="profile-name">{{ userInfo.realname || userInfo.username || '--' }}</text>
					<text class="profile-subtitle">{{ userInfo.adminClass || userInfo.className || '--' }}</text>
				</view>

				<!-- 信息卡片 -->
				<view class="info-card">

					<!-- 基础资料 -->
					<view class="info-section">
						<text class="info-section__title">基础资料</text>
						<view class="info-list">
							<view class="info-item">
								<view class="info-item__left"><l-icon name="user" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">学号</text></view>
								<text class="info-item__value">{{ userInfo.studentNo || userInfo.username || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="book" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">考生号</text></view>
								<text class="info-item__value">{{ userInfo.examNo || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="user" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">姓名</text></view>
								<text class="info-item__value">{{ userInfo.realname || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="edit" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">曾用名</text></view>
								<text class="info-item__value">{{ userInfo.nameUsedBefore || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="user" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">性别</text></view>
								<text class="info-item__value">{{ getSexText(userInfo.sex) }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="cake" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">出生日期</text></view>
								<text class="info-item__value">{{ userInfo.birthday || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="shield" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">身份证号</text></view>
								<text class="info-item__value">{{ userInfo.identityCard || userInfo.idCardNo ? maskIdCard(userInfo.identityCard || userInfo.idCardNo) : '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="star" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">民族</text></view>
								<text class="info-item__value">{{ userInfo.nation || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="flag" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">政治面貌</text></view>
								<text class="info-item__value">{{ userInfo.politicsStatus || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="heart" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">婚姻状况</text></view>
								<text class="info-item__value">{{ getMarriedText(userInfo.isMarried) }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="map-marker" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">籍贯</text></view>
								<text class="info-item__value">{{ userInfo.nativePlace || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="school" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">生源省市</text></view>
								<text class="info-item__value">{{ userInfo.sourceProvince || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="school" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">毕业学校</text></view>
								<text class="info-item__value">{{ userInfo.graduateSchool || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="school" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">毕业学校类型</text></view>
								<text class="info-item__value">{{ userInfo.graduateType || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="school" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">毕业形式</text></view>
								<text class="info-item__value">{{ userInfo.graduateForm || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="clipboard" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">入学总分</text></view>
								<text class="info-item__value">{{ userInfo.entranceScore || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="heart-pulse" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">健康状况</text></view>
								<text class="info-item__value">{{ getHealthText(userInfo.healthCondition) }}</text>
							</view>
						</view>
					</view>

					<!-- 学业信息 -->
					<view class="info-section">
						<text class="info-section__title">学业信息</text>
						<view class="info-list">
							<view class="info-item">
								<view class="info-item__left"><l-icon name="map" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">校区</text></view>
								<text class="info-item__value">{{ userInfo.campus || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="check-circle" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">学籍状态</text></view>
								<text class="info-item__value">{{ userInfo.studentStatus || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="domain" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">院系</text></view>
								<text class="info-item__value">{{ userInfo.facultyName || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="book" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">专业</text></view>
								<text class="info-item__value">{{ userInfo.majorName || userInfo.professionName || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="calendar" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">年级</text></view>
								<text class="info-item__value">{{ userInfo.grade || getGradeName(userInfo.gradeId) || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="account-group" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">行政班</text></view>
								<text class="info-item__value">{{ userInfo.adminClass || userInfo.className || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="calendar-check" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">入学日期</text></view>
								<text class="info-item__value">{{ userInfo.enrollmentDate || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="calendar-clock" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">预计毕业</text></view>
								<text class="info-item__value">{{ userInfo.expectedGradDate || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="school" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">学习形式</text></view>
								<text class="info-item__value">{{ userInfo.studyForm || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="clock-outline" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">学制</text></view>
								<text class="info-item__value">{{ userInfo.educationYears ? userInfo.educationYears + '年' : '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="tag" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">录取类别</text></view>
								<text class="info-item__value">{{ userInfo.enrollmentType || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="time" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">当前学期</text></view>
								<text class="info-item__value">{{ userInfo.currentSemester || '--' }}</text>
							</view>
						</view>
					</view>

					<!-- 联系方式 -->
					<view class="info-section" v-if="hasContactInfo">
						<text class="info-section__title">联系方式</text>
						<view class="info-list">
							<view class="info-item">
								<view class="info-item__left"><l-icon name="cellphone" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">手机</text></view>
								<text class="info-item__value">{{ userInfo.phone || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left info-item__left--email"><l-icon name="email" style="font-size: 16px; color: #fff;"></l-icon><text class="info-item__label">邮箱</text></view>
								<view class="info-item__value info-item__value--email">
									<text class="info-item__text">{{ userInfo.email || '--' }}</text>
								</view>
							</view>
							<view class="info-item" v-if="userInfo.qq">
								<view class="info-item__left"><l-icon name="chat" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">QQ</text></view>
								<text class="info-item__value">{{ userInfo.qq }}</text>
							</view>
							<view class="info-item" v-if="userInfo.personalAddress">
								<view class="info-item__left"><l-icon name="map-marker" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">联系地址</text></view>
								<text class="info-item__value">{{ userInfo.personalAddress }}</text>
							</view>
						</view>
					</view>

					<!-- 户口情况 -->
					<view class="info-section" v-if="hasHouseholdInfo">
						<text class="info-section__title">户口情况</text>
						<view class="info-list">
							<view class="info-item">
								<view class="info-item__left"><l-icon name="home" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">户口类型</text></view>
								<text class="info-item__value">{{ getHouseholdTypeText(userInfo.householdType) }}</text>
							</view>
							<view class="info-item" v-if="userInfo.householdProvince || userInfo.householdCity">
								<view class="info-item__left"><l-icon name="map-marker" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">户籍地</text></view>
								<text class="info-item__value">{{ (userInfo.householdProvince || '') + (userInfo.householdCity || '') }}</text>
							</view>
							<view class="info-item" v-if="userInfo.householdAddress">
								<view class="info-item__left"><l-icon name="map-marker" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">户籍地址</text></view>
								<text class="info-item__value">{{ userInfo.householdAddress }}</text>
							</view>
							<view class="info-item" v-if="userInfo.householdOffice">
								<view class="info-item__left"><l-icon name="office-building" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">户口登记机关</text></view>
								<text class="info-item__value">{{ userInfo.householdOffice }}</text>
							</view>
							<view class="info-item" v-if="userInfo.trainStartStation || userInfo.trainStopStation">
								<view class="info-item__left"><l-icon name="train" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">火车区间</text></view>
								<text class="info-item__value">{{ (userInfo.trainStartStation || '--') + ' ↔ ' + (userInfo.trainStopStation || '--') }}</text>
							</view>
						</view>
					</view>

					<!-- 家庭资料 -->
					<view class="info-section" v-if="userInfo.familyMembers && userInfo.familyMembers.length > 0">
						<text class="info-section__title">家庭资料</text>
						<view class="info-list">
							<view class="info-item" v-if="userInfo.familyAddress">
								<view class="info-item__left"><l-icon name="home" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">家庭地址</text></view>
								<text class="info-item__value">{{ userInfo.familyAddress }}</text>
							</view>
							<view class="info-item" v-if="userInfo.familyPhone">
								<view class="info-item__left"><l-icon name="phone" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">家庭电话</text></view>
								<text class="info-item__value">{{ userInfo.familyPhone }}</text>
							</view>
							<view class="info-item" v-if="userInfo.familyPost">
								<view class="info-item__left"><l-icon name="mail" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">邮政编码</text></view>
								<text class="info-item__value">{{ userInfo.familyPost }}</text>
							</view>
						</view>
						<view class="family-members">
							<view class="family-member-item" v-for="(m, idx) in userInfo.familyMembers" :key="idx">
								<text class="family-member-name">{{ m.name }}</text>
								<text class="family-member-rel">{{ m.relationship }}</text>
								<text class="family-member-guardian" v-if="m.isGuardian === '0'">监护人</text>
								<text class="family-member-contact">{{ m.mobile || '--' }}</text>
							</view>
						</view>
					</view>

					<!-- 学校经历 -->
					<view class="info-section" v-if="userInfo.schoolExperiences && userInfo.schoolExperiences.length > 0">
						<text class="info-section__title">学校经历</text>
						<view class="school-exp-list">
							<view class="school-exp-item" v-for="(exp, idx) in userInfo.schoolExperiences" :key="idx">
								<text class="school-exp-school">{{ exp.school }}</text>
								<text class="school-exp-date">{{ exp.startDate }} ~ {{ exp.endDate }}</text>
								<text class="school-exp-job">{{ exp.job || '无职务' }}</text>
							</view>
						</view>
					</view>

					<!-- 学业调整 -->
					<view class="info-section" v-if="userInfo.academicChanges && userInfo.academicChanges.length > 0">
						<text class="info-section__title">学业调整</text>
						<view class="info-list">
							<view class="info-item" v-for="(ch, idx) in userInfo.academicChanges" :key="idx">
								<view class="info-item__left"><l-icon name="swap-horizontal" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">{{ ch.changeType }}</text></view>
								<text class="info-item__value">{{ ch.changeDate }} {{ ch.operator }}</text>
							</view>
						</view>
					</view>
				</view>

				<view class="bottom-safe-area"></view>
			</view>
		</scroll-view>
	</view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getUserArchive } from '../../pages/api/discover.js'

const statusBarHeight = ref(20)
const userInfo = ref({})
const loading = ref(false)
const error = ref('')

const avatarText = computed(() => {
	const name = userInfo.value.realname || userInfo.value.username || ''
	return name.charAt(0).toUpperCase()
})
const hasContactInfo = computed(() => !!(userInfo.value.phone || userInfo.value.email || userInfo.value.qq || userInfo.value.personalAddress))
const hasIdentityInfo = computed(() => !!(userInfo.value.identityCard || userInfo.value.idCardNo || userInfo.value.birthday))
const hasHouseholdInfo = computed(() => !!(
	userInfo.value.householdType ||
	userInfo.value.householdProvince ||
	userInfo.value.householdCity ||
	userInfo.value.householdAddress ||
	userInfo.value.householdOffice ||
	userInfo.value.trainStartStation ||
	userInfo.value.trainStopStation
))

function goBack() {
	uni.navigateBack()
}

function goToEdit() {
	uni.navigateTo({ url: '/pages/discover/archive-edit' })
}

function loadFromStorage() {
	try {
		const stored = uni.getStorageSync('userInfo')
		if (stored) {
			const info = typeof stored === 'string' ? JSON.parse(stored) : stored
			if (info && Object.keys(info).length > 0) {
				userInfo.value = info
				return true
			}
		}
	} catch {}
	return false
}

async function fetchProfile() {
	loading.value = true
	error.value = ''

	// Step 1: 优先使用缓存的用户信息
	const hasStorage = loadFromStorage()

	// Step 2: 从后端获取完整档案信息（合并本地 + 学校扩展字段）
	try {
		const res = await getUserArchive()
		if (res && res.success && res.result) {
			// 以后端返回的完整数据为准
			userInfo.value = res.result
		} else if (!hasStorage) {
			throw new Error('获取个人信息失败，请重新登录')
		}
	} catch (e) {
		console.error('获取档案信息失败', e)
		if (!hasStorage) {
			error.value = e.message || '获取个人信息失败'
		}
	} finally {
		loading.value = false
	}
}

function getSexText(sex) {
	if (sex === 1 || sex === '男') return '男'
	if (sex === 2 || sex === '女') return '女'
	return '--'
}

function getMarriedText(val) {
	if (val === '0' || val === '已婚') return '已婚'
	if (val === '1' || val === '未婚') return '未婚'
	return '--'
}

function getHealthText(val) {
	if (val === '0' || val === '健康') return '健康'
	if (val === '1' || val === '一般') return '一般'
	if (val === '2' || val === '较差') return '较差'
	return '--'
}

function getHouseholdTypeText(val) {
	if (val === '0' || val === '农村') return '农村'
	if (val === '1' || val === '城市') return '城市'
	return '--'
}

function maskIdCard(id) {
	if (!id || id.length < 8) return id
	return id.substring(0, 4) + '****' + id.substring(id.length - 4)
}

function getFacultyName(facultyId) {
	const map = { 1: '互联网+' }
	return map[facultyId] || '--'
}

function getProfessionName(professionId) {
	const map = { 3: '计算机应用技术' }
	return map[professionId] || '--'
}

function getGradeName(gradeId) {
	const map = { 37: '2024级' }
	return map[gradeId] || '--'
}

onLoad((options) => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 20
	fetchProfile()
})
</script>

<style lang="scss" scoped>
@import url('../../static/css/home-page.css');

.profile-page {
	width: 100%;
	height: 100vh;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
}

.nav-bar {
	background: linear-gradient(135deg, #06b6d4, #22d3ee);
	flex-shrink: 0;
}

.nav-bar__content {
	display: flex;
	align-items: center;
	justify-content: space-between;
	height: 88rpx;
	padding: 0 24rpx;
}

.nav-bar__back,
.nav-bar__action {
	width: 60rpx;
	height: 60rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.nav-bar__action {
	justify-content: flex-end;
}

.nav-bar__edit {
	font-size: 28rpx;
	font-weight: 700;
	color: #fff;
}

.nav-bar__title {
	font-size: 32rpx;
	font-weight: 700;
	color: #fff;
	text-align: center;
}

.content-scroll {
	flex: 1;
	height: 0;
}

.state-container {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 120rpx 0;
	gap: 16rpx;
}

.state-spinner {
	width: 48rpx;
	height: 48rpx;
	border: 4rpx solid rgba(6, 182, 212, 0.15);
	border-top-color: #06b6d4;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

@keyframes spin {
	to { transform: rotate(360deg); }
}

.state-text {
	font-size: 28rpx;
	color: var(--text-secondary);
	font-weight: 600;
}

.state-container--error {
	gap: 12rpx;
}

.state-icon-circle {
	width: 72rpx;
	height: 72rpx;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 36rpx;
	font-weight: 800;
}

.state-icon-circle--error {
	background: rgba(239, 68, 68, 0.1);
	color: #ef4444;
}

.state-btn {
	margin-top: 8rpx;
	padding: 12rpx 40rpx;
	background: linear-gradient(135deg, #06b6d4, #22d3ee);
	color: #fff;
	font-size: 26rpx;
	font-weight: 600;
	border-radius: 32rpx;
	border: none;
}

.profile-content {
	padding: 24rpx;
}

.profile-header {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 48rpx 0 32rpx;
}

.profile-avatar {
	width: 140rpx;
	height: 140rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, #06b6d4, #22d3ee);
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 20rpx;
	box-shadow: 0 8rpx 32rpx rgba(6, 182, 212, 0.3);
}

.profile-avatar__text {
	font-size: 56rpx;
	font-weight: 800;
	color: #fff;
}

.profile-name {
	font-size: 40rpx;
	font-weight: 800;
	color: var(--text-primary);
	margin-bottom: 8rpx;
}

.profile-subtitle {
	font-size: 26rpx;
	color: var(--text-secondary);
}

.info-card {
	background: #fff;
	border-radius: 24rpx;
	overflow: hidden;
	box-shadow: 0 2rpx 12rpx rgba(6, 182, 212, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.info-section {
	padding: 0 0 8rpx;
}

.info-section + .info-section {
	border-top: 1px solid rgba(226, 232, 240, 0.7);
}

.info-section__title {
	display: block;
	padding: 24rpx 32rpx 16rpx;
	font-size: 24rpx;
	font-weight: 700;
	color: var(--text-tertiary);
	letter-spacing: 1px;
}

.info-list {
	display: flex;
	flex-direction: column;
}

.info-item {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 20rpx 32rpx;
	border-bottom: 1px solid rgba(226, 232, 240, 0.5);
}

.info-item:last-child {
	border-bottom: none;
}

.info-item__left {
	display: flex;
	align-items: center;
	gap: 12rpx;
}

.info-item__label {
	font-size: 28rpx;
	color: var(--text-secondary);
}

.info-item__value {
	font-size: 28rpx;
	color: var(--text-primary);
	font-weight: 600;
}

.info-item__left--email {
	background: linear-gradient(135deg, #3b82f6, #60a5fa);
	border-radius: 8rpx;
	padding: 6rpx 12rpx;
	gap: 8rpx;

	l-icon {
		color: #fff !important;
	}

	.info-item__label {
		color: #fff;
	}
}

.info-item__value--email {
	background: rgba(59, 130, 246, 0.06);
	border: 1px solid rgba(59, 130, 246, 0.15);
	border-radius: 10rpx;
	padding: 6rpx 16rpx;

	.info-item__text {
		font-size: 26rpx;
		color: #2563eb;
		font-weight: 600;
	}
}

.bottom-safe-area {
	height: 48rpx;
}

/* 家庭成员 */
.family-members {
	padding: 0 32rpx 24rpx;
	display: flex;
	flex-direction: column;
	gap: 12rpx;
}

.family-member-item {
	display: flex;
	align-items: center;
	gap: 16rpx;
	padding: 16rpx 20rpx;
	background: rgba(6, 182, 212, 0.04);
	border-radius: 12rpx;
	border: 1px solid rgba(6, 182, 212, 0.1);
}

.family-member-name {
	font-size: 28rpx;
	font-weight: 600;
	color: var(--text-primary);
	flex: 1;
}

.family-member-rel {
	font-size: 24rpx;
	color: var(--text-secondary);
	padding: 4rpx 12rpx;
	background: rgba(6, 182, 212, 0.08);
	border-radius: 8rpx;
}

.family-member-guardian {
	font-size: 22rpx;
	color: #06b6d4;
	font-weight: 600;
}

.family-member-contact {
	font-size: 24rpx;
	color: var(--text-secondary);
}

/* 学校经历 */
.school-exp-list {
	padding: 0 32rpx 24rpx;
	display: flex;
	flex-direction: column;
	gap: 12rpx;
}

.school-exp-item {
	padding: 16rpx 20rpx;
	background: rgba(6, 182, 212, 0.04);
	border-radius: 12rpx;
	border: 1px solid rgba(6, 182, 212, 0.1);
	display: flex;
	flex-direction: column;
	gap: 6rpx;
}

.school-exp-school {
	font-size: 28rpx;
	font-weight: 600;
	color: var(--text-primary);
}

.school-exp-date {
	font-size: 24rpx;
	color: var(--text-secondary);
}

.school-exp-job {
	font-size: 24rpx;
	color: var(--text-tertiary);
}
</style>
