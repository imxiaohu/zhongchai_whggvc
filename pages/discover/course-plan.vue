<template>
	<view class="course-plan-page">
		<!-- 自定义导航栏 -->
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">课程计划</text>
				<view class="nav-bar__placeholder"></view>
			</view>
		</view>

		<!-- 学期筛选栏 -->
		<view class="filter-bar">
			<view class="filter-bar__semester" @tap="openSemesterPicker">
				<l-icon name="calendar" style="font-size: 14px; color: #ec4899; margin-right: 8rpx;"></l-icon>
				<text class="filter-bar__semester-text">{{ currentSemester || '加载中...' }}</text>
				<l-icon name="chevron-down" style="font-size: 14px; color: #94a3b8; margin-left: 4rpx;"></l-icon>
			</view>
			<view class="filter-bar__count" v-if="total > 0">
				<text class="filter-bar__count-text">共 {{ total }} 门</text>
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
			<view v-if="loading && records.length === 0" class="state-container">
				<view class="state-spinner"></view>
				<text class="state-text">加载中...</text>
			</view>

			<!-- 空状态 -->
			<view v-else-if="!loading && records.length === 0" class="state-container">
				<view class="state-icon-wrap state-icon-wrap--clipboard">
					<l-icon name="no-result" style="font-size: 48px; color: #94a3b8;"></l-icon>
				</view>
				<text class="state-text">{{ semesterLoading ? '学期加载中...' : '暂无课程计划' }}</text>
			</view>

			<!-- 错误状态 -->
			<view v-else-if="error" class="state-container state-container--error">
				<view class="state-icon-circle state-icon-circle--error">!</view>
				<text class="state-text">{{ error }}</text>
				<button class="state-btn" @tap="fetchData">重试</button>
			</view>

			<!-- 列表 -->
			<view v-else class="list-container">
				<view
					v-for="(item, index) in records"
					:key="index"
					class="course-card"
					@tap="viewDetail(item)"
				>
					<view class="course-card__header">
						<text class="course-card__name">{{ item.courseName }}</text>
						<view class="course-card__badge" :class="getPropertyClass(item.courseProperty)">
							<text class="course-card__badge-text">{{ item.courseProperty }}</text>
						</view>
					</view>

					<view class="course-card__info">
						<view class="course-card__info-item">
							<l-icon name="school" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
							<text class="course-card__info-text">{{ item.teacherNames ? item.teacherNames.replace(/#/g, '').split('(')[0].trim() : '--' }}</text>
						</view>
						<view class="course-card__info-item">
							<l-icon name="certificate" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
							<text class="course-card__info-text">{{ item.credit }} 学分</text>
						</view>
						<view class="course-card__info-item" v-if="item.getPoint">
							<l-icon name="star" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
							<text class="course-card__info-text">{{ item.getPoint }}绩点</text>
						</view>
					</view>

					<view class="course-card__score" v-if="item.finalScore">
						<text class="course-card__score-label">成绩</text>
						<text class="course-card__score-value" :class="getScoreClass(item.finalScore)">{{ item.finalScore }}</text>
					</view>

					<view class="course-card__footer">
						<text class="course-card__test-note">{{ item.testNote }}</text>
					</view>
				</view>

				<!-- 加载更多 -->
				<view class="load-more" v-if="records.length > 0">
					<view v-if="loadingMore" class="state-spinner state-spinner--small"></view>
					<text v-else-if="noMore" class="load-more__text">没有更多了</text>
					<text v-else class="load-more__text" @tap="loadMore">加载更多</text>
				</view>
			</view>
		</scroll-view>
	</view>
</template>

<script setup>
import { ref } from 'vue'
import { getCoursePlan, getSemesterList, getCurrentSemester } from '../../pages/api/discover.js'

const statusBarHeight = ref(20)
const records = ref([])
const total = ref(0)
const pageNo = ref(1)
const pageSize = ref(15)
const loading = ref(false)
const loadingMore = ref(false)
const refreshing = ref(false)
const noMore = ref(false)
const error = ref('')

const semesterList = ref([])
const semesterIndex = ref(0)
const semesterLoading = ref(true)
const currentSemester = ref('')

function initStatusBarHeight() {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 20
}

function goBack() {
	uni.navigateBack()
}

function openSemesterPicker() {
	if (semesterLoading.value || !semesterList.value.length) return
	try {
		uni.showActionSheet({
			itemList: semesterList.value.map(s => s.name || s.value || ''),
			success: (res) => {
				const idx = res.tapIndex
				semesterIndex.value = idx
				currentSemester.value = semesterList.value[idx]?.value || ''
				fetchData(true)
			}
		})
	} catch (e) {
		console.error('openSemesterPicker failed:', e)
	}
}

async function loadSemesterInfo() {
	semesterLoading.value = true
	try {
		const [semList, currentSem] = await Promise.all([
			getSemesterList(),
			getCurrentSemester()
		])
		if (Array.isArray(semList) && semList.length > 0) {
			semesterList.value = semList
			const currentIdx = semList.findIndex(s => s.value === currentSem)
			semesterIndex.value = currentIdx !== -1 ? currentIdx : 0
			currentSemester.value = semList[semesterIndex.value]?.value || currentSem || ''
		} else {
			currentSemester.value = currentSem || ''
		}
	} catch (e) {
		currentSemester.value = ''
		console.error('获取学期信息失败', e)
	} finally {
		semesterLoading.value = false
	}
}

async function fetchData(reset = false) {
	if (loading.value) return
	if (reset) {
		pageNo.value = 1
		noMore.value = false
	}
	loading.value = true
	error.value = ''
	try {
		const res = await getCoursePlan({
			current: pageNo.value,
			size: pageSize.value,
			currentSemester: currentSemester.value
		})
		const result = res && res.result
		if (result && result.records) {
			if (reset) {
				records.value = result.records
			} else {
				records.value = [...records.value, ...result.records]
			}
			total.value = result.total || 0
			noMore.value = records.value.length >= total.value
		} else {
			records.value = []
			total.value = 0
		}
	} catch (e) {
		console.error('获取课程计划失败', e)
		error.value = e.message || '获取课程计划失败'
	} finally {
		loading.value = false
		refreshing.value = false
	}
}

async function onRefresh() {
	refreshing.value = true
	await loadSemesterInfo()
	await fetchData(true)
}

async function loadMore() {
	if (loadingMore.value || noMore.value) return
	loadingMore.value = true
	pageNo.value++
	await fetchData(false)
	loadingMore.value = false
}

function getPropertyClass(property) {
	if (property === '必修课') return 'course-card__badge--required'
	if (property === '选修课') return 'course-card__badge--elective'
	return 'course-card__badge--other'
}

function getScoreClass(score) {
	if (!score) return ''
	const s = String(score).toUpperCase()
	if (s === 'A' || s === 'A+') return 'course-card__score-value--a'
	if (s === 'B') return 'course-card__score-value--b'
	if (s === 'C') return 'course-card__score-value--c'
	if (s === 'D') return 'course-card__score-value--d'
	return 'course-card__score-value--fail'
}

function viewDetail(item) {
	const info = []
	info.push(`学期：${item.currentSemester || '--'}`)
	info.push(`课程性质：${item.courseProperty || '--'}`)
	info.push(`学分：${item.credit || '--'}`)
	info.push(`获得学分：${item.getCredit || '--'}`)
	info.push(`获得绩点：${item.getPoint || '--'}`)
	info.push(`成绩：${item.finalScore || '--'}`)
	info.push(`考试状态：${item.testNote || '--'}`)
	if (item.courseTime) {
		info.push(`课时：${item.courseTime}学时`)
	}
	if (item.teacherNames) {
		info.push(`教师：${item.teacherNames.replace(/#/g, '').trim()}`)
	}
	uni.showModal({
		title: item.courseName,
		content: info.join('\n'),
		showCancel: false
	})
}

uni.getSystemInfoSync && initStatusBarHeight()
loadSemesterInfo().then(() => fetchData(true))
</script>

<style lang="scss" scoped>
@import url('../../static/css/home-page.css');

.course-plan-page {
	width: 100%;
	height: 100vh;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
}

.nav-bar {
	background: linear-gradient(135deg, #ec4899, #f472b6);
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

.filter-bar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 16rpx 24rpx;
	background: #fff;
	border-bottom: 1rpx solid rgba(148, 163, 184, 0.1);
	flex-shrink: 0;
}

.filter-bar__semester {
	display: flex;
	align-items: center;
	padding: 10rpx 20rpx;
	background: rgba(236, 72, 153, 0.06);
	border-radius: 24rpx;
}

.filter-bar__semester-text {
	font-size: 26rpx;
	font-weight: 600;
	color: #ec4899;
	max-width: 380rpx;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.filter-bar__count {
	flex-shrink: 0;
}

.filter-bar__count-text {
	font-size: 24rpx;
	color: var(--text-tertiary);
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
	border: 4rpx solid rgba(236, 72, 153, 0.15);
	border-top-color: #ec4899;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

.state-spinner--small {
	width: 32rpx;
	height: 32rpx;
	border-width: 3rpx;
}

@keyframes spin {
	to { transform: rotate(360deg); }
}

.state-emoji {
	font-size: 80rpx;
	line-height: 1;
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
	background: linear-gradient(135deg, #ec4899, #f472b6);
	color: #fff;
	font-size: 26rpx;
	font-weight: 600;
	border-radius: 32rpx;
	border: none;
}

.list-container {
	padding: 16rpx 24rpx;
}

.course-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 28rpx;
	margin-bottom: 16rpx;
	box-shadow: 0 2rpx 12rpx rgba(236, 72, 153, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.course-card__header {
	display: flex;
	align-items: flex-start;
	justify-content: space-between;
	margin-bottom: 16rpx;
}

.course-card__name {
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	flex: 1;
	padding-right: 16rpx;
	line-height: 1.4;
}

.course-card__badge {
	padding: 4rpx 16rpx;
	border-radius: 16rpx;
	flex-shrink: 0;
}

.course-card__badge--required {
	background: rgba(239, 68, 68, 0.12);
}

.course-card__badge--required .course-card__badge-text {
	color: #ef4444;
}

.course-card__badge--elective {
	background: rgba(59, 130, 246, 0.12);
}

.course-card__badge--elective .course-card__badge-text {
	color: #3b82f6;
}

.course-card__badge--other {
	background: rgba(148, 163, 184, 0.12);
}

.course-card__badge--other .course-card__badge-text {
	color: #94a3b8;
}

.course-card__badge-text {
	font-size: 22rpx;
	font-weight: 600;
}

.course-card__info {
	display: flex;
	flex-wrap: wrap;
	gap: 16rpx;
	margin-bottom: 12rpx;
}

.course-card__info-item {
	display: flex;
	align-items: center;
}

.course-card__info-text {
	font-size: 24rpx;
	color: var(--text-secondary);
}

.course-card__score {
	display: flex;
	align-items: center;
	gap: 8rpx;
	margin-bottom: 12rpx;
}

.course-card__score-label {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.course-card__score-value {
	font-size: 28rpx;
	font-weight: 800;
}

.course-card__score-value--a { color: #10b981; }
.course-card__score-value--b { color: #3b82f6; }
.course-card__score-value--c { color: #f59e0b; }
.course-card__score-value--d { color: #f97316; }
.course-card__score-value--fail { color: #ef4444; }

.course-card__footer {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding-top: 12rpx;
	border-top: 1rpx solid rgba(148, 163, 184, 0.1);
}

.course-card__test-note {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.load-more {
	display: flex;
	justify-content: center;
	align-items: center;
	padding: 32rpx 0;
}

.load-more__text {
	font-size: 26rpx;
	color: #ec4899;
	font-weight: 600;
}
</style>
