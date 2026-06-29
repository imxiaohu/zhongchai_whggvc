<template>
	<view class="teachers-page">
		<!-- 自定义导航栏 -->
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">我的老师</text>
				<view class="nav-bar__placeholder"></view>
			</view>
		</view>

		<!-- 内容区 -->
		<scroll-view
			class="content-scroll"
			scroll-y
			@scrolltolower="loadMore"
			:refresher-enabled="true"
			@refresherrefresh="onRefresh"
			:refresher-triggered="refreshing"
		>
			<!-- 加载中 -->
			<view v-if="loading && list.length === 0" class="state-container">
				<view class="state-spinner"></view>
				<text class="state-text">加载中...</text>
			</view>

			<!-- 错误状态 -->
			<view v-else-if="error" class="state-container state-container--error">
				<view class="state-icon-circle state-icon-circle--error">!</view>
				<text class="state-text">{{ error }}</text>
				<button class="state-btn" @tap="fetchData">重试</button>
			</view>

			<!-- 列表 -->
			<view v-else class="list-container">
				<!-- 筛选标签 -->
				<view class="filter-bar">
					<view
						class="filter-tag"
						:class="{ 'filter-tag--active': currentGrade === '' }"
						@tap="filterByGrade('')"
					>
						<text class="filter-tag__text">全部</text>
					</view>
					<view
						v-for="g in availableGrades"
						:key="g"
						class="filter-tag"
						:class="{ 'filter-tag--active': currentGrade === g }"
						@tap="filterByGrade(g)"
					>
						<text class="filter-tag__text">{{ g }}级</text>
					</view>
				</view>

				<!-- 空状态 -->
				<view v-if="list.length === 0" class="state-container">
					<view class="state-icon-wrap state-icon-wrap--clipboard">
						<l-icon name="no-result" style="font-size: 48px; color: #94a3b8;"></l-icon>
					</view>
					<text class="state-text">暂无老师数据</text>
				</view>

				<!-- 老师列表 -->
				<view v-else>
					<view
						v-for="(item, index) in list"
						:key="index"
						class="teacher-card"
					>
						<view class="teacher-card__left">
							<view class="teacher-avatar">
								<text class="teacher-avatar__text">{{ getAvatarText(item.name) }}</text>
							</view>
						</view>
						<view class="teacher-card__right">
							<view class="teacher-card__header">
								<text class="teacher-name">{{ item.name || '--' }}</text>
								<view class="teacher-badge" v-if="item.mobileNumber">
									<l-icon name="cellphone" style="font-size: 12px;"></l-icon>
									<text class="teacher-badge__text">有电话</text>
								</view>
							</view>
							<text class="teacher-course">{{ item.courseName || '--' }}</text>
							<text class="teacher-semester">{{ item.currentSemester || '--' }}</text>
						</view>
						<view class="teacher-card__action" v-if="item.mobileNumber">
							<view class="action-btn" @tap="callTeacher(item)">
								<l-icon name="call" style="font-size: 20px; color: #fff;"></l-icon>
							</view>
						</view>
					</view>

					<!-- 加载更多 -->
					<view v-if="loadingMore" class="load-more">
						<view class="load-more-spinner"></view>
						<text class="load-more-text">加载中...</text>
					</view>
					<view v-else-if="noMore" class="no-more">
						<text class="no-more-text">没有更多了</text>
					</view>
				</view>
			</view>

			<view class="bottom-safe-area"></view>
		</scroll-view>

		<!-- PC验证码弹窗 -->
		<PCCaptchaModal
			:visible="captchaModalVisible"
			:sessionId="captchaSessionId"
			:captchaImage="captchaImage"
			:tips="'登录已过期，请输入验证码'"
			@close="onCaptchaClose"
			@success="onCaptchaSuccess"
			@refresh-captcha="onCaptchaRefresh"
		/>
	</view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { pcGetTeachers } from '../../pages/api/discover.js'
import PCCaptchaModal from '@/components/PCCaptchaModal.vue'

const statusBarHeight = ref(20)
const list = ref([])
const loading = ref(false)
const error = ref('')
const refreshing = ref(false)
const loadingMore = ref(false)
const noMore = ref(false)
const currentPage = ref(1)
const totalPages = ref(1)
const currentGrade = ref('')
const availableGrades = ref([])
const captchaModalVisible = ref(false)
const captchaSessionId = ref('')
const captchaImage = ref('')

const getAvatarText = (name) => {
	if (!name) return '?'
	return name.charAt(0).toUpperCase()
}

function goBack() {
	uni.navigateBack()
}

async function fetchData() {
	loading.value = true
	error.value = ''
	noMore.value = false
	currentPage.value = 1
	try {
		const res = await pcGetTeachers({
			pageNum: 1,
			pageSize: 15,
			gradeName: currentGrade.value
		})
		if (res && res.success && res.result) {
			const data = res.result
			if (data.needManual) {
				captchaModalVisible.value = true
				captchaSessionId.value = data.sessionId || ''
				captchaImage.value = data.captcha || ''
				list.value = []
				return
			}
			list.value = data.list || []
			totalPages.value = data.pages || 1
			currentPage.value = 1
			extractGrades(list.value)
		} else {
			throw new Error(res?.message || '获取老师列表失败')
		}
	} catch (e) {
		error.value = e?.message || e?.errMsg || '获取失败，请重试'
	} finally {
		loading.value = false
	}
}

function onCaptchaClose() {
	captchaModalVisible.value = false
}
async function onCaptchaSuccess() {
	captchaModalVisible.value = false
	await fetchData()
}
function onCaptchaRefresh() {
	captchaModalVisible.value = false
	uni.navigateTo({ url: '/pages/discover/pc-login?redirect=/pages/discover/my-teachers' })
}

function extractGrades(items) {
	const grades = new Set()
	items.forEach(item => {
		const match = (item.currentSemester || '').match(/(\d{4})-(\d{4})/)
		if (match) {
			grades.add(match[1])
		}
	})
	availableGrades.value = Array.from(grades).sort()
}

async function loadMore() {
	if (loadingMore.value || noMore.value || currentPage.value >= totalPages.value) return
	loadingMore.value = true
	currentPage.value++
	try {
		const res = await pcGetTeachers({
			pageNum: currentPage.value,
			pageSize: 15,
			gradeName: currentGrade.value
		})
		if (res && res.success && res.result) {
			const data = res.result
			if (data.needManual) {
				captchaModalVisible.value = true
				captchaSessionId.value = data.sessionId || ''
				captchaImage.value = data.captcha || ''
				currentPage.value--
				noMore.value = true
				return
			}
			list.value.push(...(data.list || []))
			if (currentPage.value >= data.pages) {
				noMore.value = true
			}
		}
	} catch (e) {
		currentPage.value--
	} finally {
		loadingMore.value = false
	}
}

async function onRefresh() {
	refreshing.value = true
	await fetchData()
	refreshing.value = false
}

function filterByGrade(grade) {
	currentGrade.value = grade
	fetchData()
}

function callTeacher(item) {
	if (!item.mobileNumber) return
	uni.showModal({
		title: item.name,
		content: `是否拨打 ${item.mobileNumber}？`,
		confirmText: '拨打',
		success: (res) => {
			if (res.confirm) {
				uni.makePhoneCall({
					phoneNumber: item.mobileNumber
				})
			}
		}
	})
}

onLoad(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 20
	fetchData()
})
</script>

<style lang="scss" scoped>
@import url('../../static/css/home-page.css');

.teachers-page {
	width: 100%;
	height: 100vh;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
}

.nav-bar {
	background: linear-gradient(135deg, #f59e0b, #fbbf24);
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
.nav-bar__placeholder {
	width: 60rpx;
	height: 60rpx;
	display: flex;
	align-items: center;
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

.list-container {
	padding: 24rpx;
}

/* 筛选栏 */
.filter-bar {
	display: flex;
	gap: 12rpx;
	flex-wrap: wrap;
	margin-bottom: 20rpx;
}

.filter-tag {
	padding: 10rpx 24rpx;
	border-radius: 32rpx;
	background: rgba(245, 158, 11, 0.08);
	border: 1px solid rgba(245, 158, 11, 0.2);
}

.filter-tag--active {
	background: linear-gradient(135deg, #f59e0b, #fbbf24);
	border-color: transparent;
}

.filter-tag__text {
	font-size: 26rpx;
	font-weight: 600;
	color: #f59e0b;
}

.filter-tag--active .filter-tag__text {
	color: #fff;
}

/* 老师卡片 */
.teacher-card {
	display: flex;
	align-items: center;
	background: #fff;
	border-radius: 20rpx;
	padding: 24rpx;
	margin-bottom: 16rpx;
	box-shadow: 0 2rpx 12rpx rgba(245, 158, 11, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	gap: 20rpx;
}

.teacher-avatar {
	width: 96rpx;
	height: 96rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, #f59e0b, #fbbf24);
	display: flex;
	align-items: center;
	justify-content: center;
}

.teacher-avatar__text {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
}

.teacher-card__right {
	flex: 1;
}

.teacher-card__header {
	display: flex;
	align-items: center;
	gap: 12rpx;
	margin-bottom: 8rpx;
}

.teacher-name {
	font-size: 32rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.teacher-badge {
	display: flex;
	align-items: center;
	gap: 4rpx;
	background: rgba(16, 185, 129, 0.08);
	border-radius: 8rpx;
	padding: 4rpx 10rpx;
	color: #10b981;
}

.teacher-badge__text {
	font-size: 22rpx;
}

.teacher-course {
	display: block;
	font-size: 28rpx;
	color: var(--text-primary);
	font-weight: 600;
	margin-bottom: 6rpx;
}

.teacher-semester {
	display: block;
	font-size: 24rpx;
	color: var(--text-tertiary);
}

.teacher-card__action {
	flex-shrink: 0;
}

.action-btn {
	width: 72rpx;
	height: 72rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, #10b981, #34d399);
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 6rpx 20rpx rgba(16, 185, 129, 0.3);
}

/* 加载更多 */
.load-more {
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 32rpx 0;
	gap: 12rpx;
}

.load-more-spinner {
	width: 36rpx;
	height: 36rpx;
	border: 4rpx solid rgba(245, 158, 11, 0.15);
	border-top-color: #f59e0b;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

.load-more-text {
	font-size: 26rpx;
	color: var(--text-secondary);
}

.no-more {
	text-align: center;
	padding: 32rpx 0;
}

.no-more-text {
	font-size: 26rpx;
	color: var(--text-tertiary);
}

/* 状态 */
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
	border: 4rpx solid rgba(245, 158, 11, 0.15);
	border-top-color: #f59e0b;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

.state-text {
	font-size: 28rpx;
	color: var(--text-secondary);
	font-weight: 600;
}

.state-emoji {
	font-size: 80rpx;
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
	background: linear-gradient(135deg, #f59e0b, #fbbf24);
	color: #fff;
	font-size: 26rpx;
	font-weight: 600;
	border-radius: 32rpx;
	border: none;
}

.bottom-safe-area {
	height: 48rpx;
}

@keyframes spin {
	to { transform: rotate(360deg); }
}
</style>
