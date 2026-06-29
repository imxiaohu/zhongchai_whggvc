<template>
	<view class="leave-page">
		<!-- 顶部蓝色渐变 Hero 区域 -->
		<view class="leave-hero">
			<view class="leave-hero-bg"></view>
			<view class="leave-hero-overlay"></view>

			<!-- 状态栏占位 -->
			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<!-- 顶部导航栏 -->
			<view class="leave-hero-nav">
				<view class="leave-hero-nav-left">
					<view class="leave-hero-back" @tap="goBack">
						<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
					</view>
					<text class="leave-hero-title">我的请假</text>
				</view>
				<view class="leave-hero-nav-right">
					<view class="leave-hero-icon-btn" @tap="onRefresh">
						<l-icon name="refresh" style="font-size: 20px; color: #fff;"></l-icon>
					</view>
				</view>
			</view>

			<!-- 欢迎语 -->
			<view class="leave-hero-content">
				<text class="leave-hero-subtitle">{{ heroSubtitle }}</text>
			</view>
		</view>

		<!-- 主要内容区域 -->
		<view class="leave-content">

			<!-- 预览模式提示 -->
			<view v-if="isPreviewMode" class="leave-preview-banner">
				<view class="leave-preview-banner-left">
					<l-icon name="info-circle-filled" style="font-size: 18px; color: var(--warning-color);"></l-icon>
					<text class="leave-preview-banner-text">当前为预览模式，显示示例数据</text>
				</view>
				<text class="leave-preview-banner-btn" @tap="goToLogin">立即登录</text>
			</view>

			<view class="leave-page-content">

				<!-- 统计概览卡片 -->
				<view v-if="!loading && !error && records.length > 0" class="leave-summary-card">
					<view class="leave-summary-bg-decor1"></view>
					<view class="leave-summary-bg-decor2"></view>
					<view class="leave-summary-item">
						<text class="leave-summary-value">{{ records.length }}</text>
						<text class="leave-summary-label">全部记录</text>
					</view>
					<view class="leave-summary-divider"></view>
					<view class="leave-summary-item">
						<text class="leave-summary-value" style="color: #86efac;">{{ passedCount }}</text>
						<text class="leave-summary-label">已通过</text>
					</view>
					<view class="leave-summary-divider"></view>
					<view class="leave-summary-item">
						<text class="leave-summary-value" style="color: #fcd34d;">{{ pendingCount }}</text>
						<text class="leave-summary-label">待审核</text>
					</view>
					<view class="leave-summary-divider"></view>
					<view class="leave-summary-item">
						<text class="leave-summary-value" style="color: #fca5a5;">{{ rejectedCount }}</text>
						<text class="leave-summary-label">已拒绝</text>
					</view>
				</view>

				<!-- 加载状态 -->
				<view v-if="loading && records.length === 0" class="leave-status-card">
					<view class="leave-status-spinner"></view>
					<text class="leave-status-title">加载中...</text>
				</view>

				<!-- 错误状态 -->
				<view v-else-if="error" class="leave-status-card">
					<view class="leave-status-icon leave-status-icon--error">
						<l-icon name="error-circle" size="40" color="var(--error-color)"></l-icon>
					</view>
					<text class="leave-status-title">{{ error }}</text>
					<button class="leave-retry-btn" @tap="fetchData(true)">重新加载</button>
				</view>

				<!-- 空状态 -->
				<view v-else-if="records.length === 0" class="leave-status-card">
					<view class="leave-status-icon leave-status-icon--success">
						<l-icon name="check-circle" size="40" color="var(--success-color)"></l-icon>
					</view>
					<text class="leave-status-title">暂无请假记录</text>
					<text class="leave-status-subtitle">您还没有提交过请假申请</text>
				</view>

				<!-- 请假记录列表 -->
				<view v-else class="leave-list">
					<view
						v-for="(item, index) in records"
						:key="index"
						class="leave-card"
						:class="{ 'leave-card--expanded': expandedId === item.studentLeaveID }"
						@tap="toggleExpand(item.studentLeaveID)"
					>
						<!-- 左侧状态条 -->
						<view class="leave-card-status-bar"
							:class="{
								'leave-card-status-bar--passed': getEffectiveStatus(item) === 1,
								'leave-card-status-bar--rejected': getEffectiveStatus(item) === 2,
								'leave-card-status-bar--pending': getEffectiveStatus(item) === 0
							}"
						></view>

						<view class="leave-card-body">
							<!-- Header: 请假类型 + 审核状态 + 箭头 -->
							<view class="leave-card-header">
								<view class="leave-card-tag"
									:class="{
										'leave-card-tag--passed': getEffectiveStatus(item) === 1,
										'leave-card-tag--rejected': getEffectiveStatus(item) === 2,
										'leave-card-tag--pending': getEffectiveStatus(item) === 0
									}">
									<view class="leave-card-tag-dot"
										:class="{
											'leave-card-tag-dot--passed': getEffectiveStatus(item) === 1,
											'leave-card-tag-dot--rejected': getEffectiveStatus(item) === 2,
											'leave-card-tag-dot--pending': getEffectiveStatus(item) === 0
										}">
									</view>
									<text class="leave-card-tag-text">{{ item.type || '请假' }}</text>
								</view>
								<view class="leave-card-status-badge"
									:class="{
										'leave-card-status-badge--passed': getEffectiveStatus(item) === 1,
										'leave-card-status-badge--rejected': getEffectiveStatus(item) === 2,
										'leave-card-status-badge--pending': getEffectiveStatus(item) === 0
									}">
									<text class="leave-card-status-badge-text">{{ getStatusText(item) }}</text>
								</view>
								<l-icon :name="expandedId === item.studentLeaveID ? 'chevron-up' : 'chevron-down'" size="18" color="var(--text-tertiary)"></l-icon>
							</view>

							<!-- 基础信息 -->
							<view class="leave-card-title">
								<text>{{ item.reason || '无请假原因' }}</text>
							</view>

							<view class="leave-card-footer">
								<view class="leave-card-info">
									<view class="leave-card-row">
										<l-icon name="time" size="14" class="leave-card-icon"></l-icon>
										<text>{{ formatTimeRange(item.startTime, item.endTime) }}</text>
									</view>
									<view class="leave-card-row">
										<l-icon name="user-circle-filled" size="14" class="leave-card-icon"></l-icon>
										<text>{{ item.studentName }} · {{ item.className }}</text>
									</view>
								</view>
							</view>

							<!-- 附件指示 -->
							<view v-if="item.studentLeaveAttachments && item.studentLeaveAttachments.length > 0" class="leave-card-attachments">
								<l-icon name="attach" size="14" style="color: var(--text-tertiary);"></l-icon>
								<text class="leave-card-attachments-text">{{ item.studentLeaveAttachments.length }}个附件</text>
								<view class="leave-card-expand-hint" v-if="!isExpanded(item.studentLeaveID)">
									<text>点击查看详情</text>
								</view>
							</view>

							<!-- 展开详情 -->
							<view v-if="expandedId === item.studentLeaveID" class="leave-card-detail">
								<!-- 分割线 -->
								<view class="leave-card-detail-divider"></view>

								<!-- 详细信息网格 -->
								<view class="leave-card-detail-grid">
									<view class="leave-card-detail-item" v-if="item.studyNumber">
										<text class="leave-card-detail-label">学号</text>
										<text class="leave-card-detail-value">{{ item.studyNumber }}</text>
									</view>
									<view class="leave-card-detail-item" v-if="item.contractPhoneNumber">
										<text class="leave-card-detail-label">本人电话</text>
										<text class="leave-card-detail-value">{{ item.contractPhoneNumber }}</text>
									</view>
									<view class="leave-card-detail-item" v-if="item.guardianName">
										<text class="leave-card-detail-label">监护人</text>
										<text class="leave-card-detail-value">{{ item.guardianName }}</text>
									</view>
									<view class="leave-card-detail-item" v-if="item.guardianPhoneNumber">
										<text class="leave-card-detail-label">监护人电话</text>
										<text class="leave-card-detail-value">{{ item.guardianPhoneNumber }}</text>
									</view>
									<view class="leave-card-detail-item leave-card-detail-item--full" v-if="item.provinceName || item.cityName || item.districtPlaceName || item.detailAddress">
										<text class="leave-card-detail-label">去向地点</text>
										<text class="leave-card-detail-value">{{ formatAddress(item) }}</text>
									</view>
									<view class="leave-card-detail-item leave-card-detail-item--full" v-if="item.scope">
										<text class="leave-card-detail-label">离校范围</text>
										<text class="leave-card-detail-value">{{ formatScope(item.scope) }}</text>
									</view>
									<view class="leave-card-detail-item leave-card-detail-item--full" v-if="item.auditMemo">
										<text class="leave-card-detail-label">审核备注</text>
										<text class="leave-card-detail-value">{{ item.auditMemo }}</text>
									</view>
								</view>

								<!-- 审核流程 -->
								<view v-if="item.studentLeaveAudits && item.studentLeaveAudits.length > 0" class="leave-card-audit">
									<text class="leave-card-audit-title">审核进度</text>
									<view
										v-for="(audit, idx) in item.studentLeaveAudits"
										:key="idx"
										class="leave-card-audit-item"
									>
										<view class="leave-card-audit-dot"
											:class="{
												'leave-card-audit-dot--passed': audit.auditstatus === 1,
												'leave-card-audit-dot--rejected': audit.auditstatus === 2,
												'leave-card-audit-dot--pending': audit.auditstatus === 0
											}"
										></view>
										<view class="leave-card-audit-content">
											<text class="leave-card-audit-teacher">{{ audit.teachername || '待审核' }}</text>
											<text class="leave-card-audit-time">{{ audit.audittime || '--' }}</text>
											<text class="leave-card-audit-memo" v-if="audit.auditmemo">{{ audit.auditmemo }}</text>
										</view>
									</view>
								</view>

								<!-- 附件预览 -->
								<view v-if="item.studentLeaveAttachments && item.studentLeaveAttachments.length > 0" class="leave-card-attachments-preview">
									<text class="leave-card-attachments-preview-title">附件</text>
									<scroll-view class="leave-card-attachments-scroll" scroll-x>
										<view
											class="leave-card-attachment"
											v-for="(att, attIdx) in item.studentLeaveAttachments"
											:key="attIdx"
											@tap.stop="previewAttachment(att, item.studentLeaveAttachments)"
										>
											<image
												v-if="isImageFile(att.name)"
												class="leave-card-attachment__image"
												:src="fixAttachmentUrl(att.attachmenturl)"
												mode="aspectFill"
												@error="onImageError($event, item.studentLeaveID, attIdx)"
											></image>
											<view v-else class="leave-card-attachment__file">
												<l-icon name="file-pdf" size="28" color="#ef4444"></l-icon>
											</view>
											<text class="leave-card-attachment__name">{{ att.name }}</text>
											<text class="leave-card-attachment__size">{{ att.size }}</text>
										</view>
									</scroll-view>
								</view>

								<!-- 销假信息 -->
								<view v-if="item.isReturnSchool !== null" class="leave-card-return">
									<view class="leave-card-return-badge"
										:class="{
											'leave-card-return-badge--yes': item.isReturnSchool === 1,
											'leave-card-return-badge--no': item.isReturnSchool === 0
										}"
									>
										<text class="leave-card-return-badge-text">
											{{ item.isReturnSchool === 1 ? '已返校' : '未返校' }}
										</text>
									</view>
									<text v-if="item.reportBackMemo" class="leave-card-return-memo">{{ item.reportBackMemo }}</text>
								</view>
							</view>
						</view>
					</view>

					<!-- 加载更多 -->
					<view class="load-more" v-if="records.length > 0">
						<view v-if="loadingMore" class="state-spinner state-spinner--small"></view>
						<text v-else-if="noMore" class="load-more__text">没有更多了</text>
						<text v-else class="load-more__text" @tap="loadMore">加载更多</text>
					</view>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getLeaveList } from '@/pages/api/discover.js'
import { getFileServerUrl } from '@/pages/api/discover.js'

const statusBarHeight = ref(20)
const records = ref([])
const total = ref(0)
const pageNo = ref(1)
const pageSize = ref(10)
const loading = ref(false)
const loadingMore = ref(false)
const noMore = ref(false)
const error = ref('')
const expandedId = ref(null)
const isPreviewMode = ref(false)
const failedImages = ref(new Set())
const fileServerUrl = ref('')

// 计算属性
const heroSubtitle = computed(() => {
	if (isPreviewMode.value) return '登录后可查看真实请假记录'
	if (records.value.length === 0) return '暂无请假记录'
	return `共 ${records.value.length} 条请假记录`
})

const passedCount = computed(() => records.value.filter(r => r.auditStatus === 1 || r.isReturnSchool === 1).length)
const pendingCount = computed(() => records.value.filter(r => r.auditStatus === 0 && r.isReturnSchool !== 1).length)
const rejectedCount = computed(() => records.value.filter(r => r.auditStatus === 2).length)

function initStatusBarHeight() {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 20
}

function goBack() {
	uni.navigateBack()
}

function goToLogin() {
	uni.vibrateShort()
	uni.navigateTo({ url: '/pages/login/login' })
}

function getStatusClass(status) {
	if (status === 1) return 'passed'
	if (status === 2) return 'rejected'
	return 'pending'
}

function getEffectiveStatus(item) {
	if (item.isReturnSchool === 1) return 1
	return item.auditStatus ?? 0
}

function getStatusText(item) {
	const status = typeof item === 'number' ? item : item.auditStatus
	const isReturned = typeof item === 'object' ? item.isReturnSchool === 1 : false
	if (isReturned || status === 1) return '已通过'
	if (status === 2) return '已拒绝'
	return '待审核'
}

function formatTimeRange(startTime, endTime) {
	if (!startTime) return '--'
	const s = startTime.substring(0, 16)
	if (!endTime) return s
	const e = endTime.substring(0, 16)
	return `${s} - ${e}`
}

function formatAddress(item) {
	const parts = [item.provinceName, item.cityName, item.districtPlaceName, item.detailAddress].filter(Boolean)
	return parts.join('') || '--'
}

function formatScope(scope) {
	if (scope === '0.1') return '市内'
	if (scope === '0.2') return '省内'
	if (scope === '0.3') return '省外'
	if (scope === '1.1') return '校内'
	if (scope === '1.2') return '市内离校'
	if (scope === '1.3') return '省内离校'
	if (scope === '2.1') return '实习离校'
	if (scope === '2.2') return '就医离校'
	if (scope === '2.3') return '集训/比赛离校'
	return scope || '--'
}

function isExpanded(id) {
	return expandedId.value === id
}

function toggleExpand(id) {
	if (expandedId.value === id) {
		expandedId.value = null
	} else {
		expandedId.value = id
	}
}

function isImageFile(name) {
	if (!name) return false
	const ext = name.split('.').pop().toLowerCase()
	return ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg'].includes(ext)
}

function fixAttachmentUrl(url) {
	if (!url) return ''
	const clean = url.split('<!')[0].trim()
	if (clean.startsWith('http')) return clean
	// fileServerUrl 格式如: http://localhost:2333/api/proxy/file
	// attachmentUrl 格式如: group1/M00/00/4D/wKjJiWlBF9GAWPDDAAmaysfVODg529.jpg
	// 拼接后: http://localhost:2333/api/proxy/file/group1/M00/00/4D/...
	if (fileServerUrl.value) {
		const base = fileServerUrl.value.replace(/\/$/, '')
		const path = clean.replace(/^\//, '')
		return `${base}/${path}`
	}
	return ''
}

function onImageError(e, leaveId, attIdx) {
	const key = `${leaveId}_${attIdx}`
	failedImages.value.add(key)
}

function previewAttachment(att, allAttachments) {
	if (isImageFile(att.name)) {
		const urls = allAttachments
			.filter(a => isImageFile(a.name))
			.map(a => fixAttachmentUrl(a.attachmenturl))
		const index = allAttachments.findIndex(a => a.id === att.id)
		uni.previewImage({
			current: index,
			urls
		})
	} else {
		const downloadUrl = fixAttachmentUrl(att.attachmenturl)
		uni.showModal({
			title: '附件信息',
			content: `${att.name}\n大小: ${att.size}\n\n是否在浏览器中打开?`,
			success: (res) => {
				if (res.confirm) {
					// #ifdef H5
					window.open(downloadUrl, '_blank')
					// #endif
					// #ifndef H5
					uni.showToast({ title: '请长按附件保存', icon: 'none' })
					// #endif
				}
			}
		})
	}
}

async function fetchData(reset = false) {
	if (loading.value) return
	if (reset) {
		pageNo.value = 1
		noMore.value = false
		expandedId.value = null
	}
	loading.value = !reset
	error.value = ''
	try {
		const res = await getLeaveList({
			pageNo: pageNo.value,
			pageSize: pageSize.value
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
		console.error('获取请假记录失败', e)
		if (e.statusCode === 401 || e.isTokenInvalid) {
			isPreviewMode.value = true
			loading.value = false
			return
		}
		error.value = e.message || '获取请假记录失败，请重试'
	} finally {
		loading.value = false
	}
}

async function onRefresh() {
	uni.showLoading({ title: '刷新中...', mask: true })
	await fetchData(true)
	uni.hideLoading()
	if (!error.value) {
		uni.showToast({ title: '刷新成功', icon: 'success' })
	}
}

async function loadMore() {
	if (loadingMore.value || noMore.value) return
	loadingMore.value = true
	pageNo.value++
	await fetchData(false)
	loadingMore.value = false
}

onMounted(() => {
	initStatusBarHeight()
	uni.setNavigationBarTitle({ title: '我的请假' })
	const token = uni.getStorageSync('token')
	if (!token) {
		isPreviewMode.value = true
		loading.value = false
		return
	}
	fetchData(true)
	loadFileServerUrl()
})

async function loadFileServerUrl() {
	try {
		const res = await getFileServerUrl()
		if (res && res.result) {
			fileServerUrl.value = res.result
		}
	} catch (e) {
		console.warn('获取文件服务器地址失败，使用默认地址', e)
	}
}
</script>

<style lang="scss" scoped>
/* ============================================
   Leave Page - Hero Style
   ============================================ */

.leave-page {
	flex: 1;
	width: 100%;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
	font-family: -apple-system, "SF Pro Text", "PingFang SC", sans-serif;
	position: relative;
	min-height: 100vh;
}

/* ---- Hero Section ---- */
.leave-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 280rpx;
}

.leave-hero-bg {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		#92400e 0%,
		#b45309 25%,
		#d97706 55%,
		#fbbf24 75%,
		#fde68a 100%);
	z-index: 0;
}

.leave-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(146, 64, 14, 0.65) 0%,
		rgba(180, 83, 9, 0.4) 50%,
		rgba(251, 191, 36, 0.1) 100%);
	z-index: 1;
}

.leave-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.leave-hero-nav-left {
	display: flex;
	align-items: center;
	gap: 16rpx;
	flex: 1;
}

.leave-hero-back {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255, 255, 255, 0.18);
	border-radius: 50%;
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.25);
	flex-shrink: 0;

	&:active {
		background: rgba(255, 255, 255, 0.28);
		transform: scale(0.95);
	}
}

.leave-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
}

.leave-hero-nav-right {
	display: flex;
	align-items: center;
	gap: 16rpx;
	flex-shrink: 0;
}

.leave-hero-icon-btn {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255, 255, 255, 0.18);
	border-radius: 50%;
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.25);
	transition: all 0.2s var(--ease-out);

	&:active {
		background: rgba(255, 255, 255, 0.28);
		transform: scale(0.9);
	}
}

.leave-hero-content {
	position: relative;
	z-index: 2;
	padding: 8rpx 32rpx 0;
}

.leave-hero-subtitle {
	font-size: 28rpx;
	color: rgba(255, 255, 255, 0.8);
	font-weight: 400;
}

/* ---- Content Area ---- */
.leave-content {
	flex: 1;
	padding: 0 24rpx;
	padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
}

/* ---- Preview Banner ---- */
.leave-preview-banner {
	margin-bottom: 20rpx;
	background: rgba(255, 255, 255, 0.95);
	border-radius: 20rpx;
	padding: 20rpx 24rpx;
	display: flex;
	align-items: center;
	justify-content: space-between;
	box-shadow: 0 4rpx 16rpx rgba(146, 64, 14, 0.1);
	border: 1px solid rgba(255, 255, 255, 0.8);
	margin-top: -24rpx;
}

.leave-preview-banner-left {
	display: flex;
	align-items: center;
	gap: 12rpx;
}

.leave-preview-banner-text {
	font-size: 26rpx;
	color: var(--text-secondary);
	font-weight: 500;
}

.leave-preview-banner-btn {
	font-size: 26rpx;
	font-weight: 700;
	color: var(--warning-color);
	padding: 8rpx 20rpx;
	background: rgba(245, 158, 11, 0.1);
	border-radius: 16rpx;
}

/* ---- Page Content ---- */
.leave-page-content {
	padding-top: 20rpx;
}

/* ---- Summary Card ---- */
.leave-summary-card {
	background: linear-gradient(135deg, #d97706, #b45309);
	border-radius: 28rpx;
	padding: 32rpx;
	display: flex;
	justify-content: space-around;
	align-items: center;
	margin-bottom: 24rpx;
	box-shadow: 0 8rpx 32rpx rgba(217, 119, 6, 0.25);
	position: relative;
	overflow: hidden;
}

.leave-summary-bg-decor1 {
	position: absolute;
	top: -60rpx;
	right: -60rpx;
	width: 200rpx;
	height: 200rpx;
	background: rgba(255, 255, 255, 0.08);
	border-radius: 50%;
}

.leave-summary-bg-decor2 {
	position: absolute;
	bottom: -80rpx;
	left: -40rpx;
	width: 240rpx;
	height: 240rpx;
	background: rgba(255, 255, 255, 0.05);
	border-radius: 50%;
}

.leave-summary-item {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 8rpx;
	position: relative;
	z-index: 1;
}

.leave-summary-value {
	font-size: 56rpx;
	font-weight: 800;
	color: #fff;
	line-height: 1;
}

.leave-summary-label {
	font-size: 22rpx;
	color: rgba(255, 255, 255, 0.8);
	font-weight: 500;
}

.leave-summary-divider {
	width: 1px;
	height: 80rpx;
	background: rgba(255, 255, 255, 0.25);
	position: relative;
	z-index: 1;
}

/* ---- Status Card ---- */
.leave-status-card {
	background: #fff;
	border-radius: 28rpx;
	padding: 80rpx 32rpx;
	margin-bottom: 24rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 16rpx;
	box-shadow: 0 4rpx 20rpx rgba(146, 64, 14, 0.08);
	border: 1px solid rgba(148, 163, 184, 0.12);
}

.leave-status-spinner {
	width: 64rpx;
	height: 64rpx;
	border: 5rpx solid rgba(245, 158, 11, 0.15);
	border-top-color: #d97706;
	border-radius: 50%;
	animation: leave-spin 0.8s linear infinite;
}

@keyframes leave-spin {
	to { transform: rotate(360deg); }
}

.leave-status-icon {
	width: 120rpx;
	height: 120rpx;
	border-radius: 32rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 8rpx;

	&--success { background: var(--success-soft); }
	&--error { background: var(--error-soft); }
}

.leave-status-title {
	font-size: 36rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.leave-status-subtitle {
	font-size: 26rpx;
	color: var(--text-secondary);
	text-align: center;
}

.leave-retry-btn {
	margin-top: 8rpx;
	padding: 20rpx 48rpx;
	background: linear-gradient(135deg, #d97706, #b45309);
	color: #fff;
	border-radius: 44rpx;
	font-size: 28rpx;
	font-weight: 700;
	border: none;
	box-shadow: 0 6rpx 24rpx rgba(217, 119, 6, 0.3);

	&:active { transform: translateY(2rpx); opacity: 0.9; }
}

/* ---- List ---- */
.leave-list {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
}

/* ---- Card ---- */
.leave-card {
	background: #fff;
	border-radius: 24rpx;
	display: flex;
	overflow: hidden;
	box-shadow: 0 4rpx 16rpx rgba(146, 64, 14, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	transition: all 0.25s var(--ease-out);

	&--expanded {
		box-shadow: 0 8rpx 32rpx rgba(146, 64, 14, 0.12);
		border-color: rgba(217, 119, 6, 0.2);
	}

	&:active {
		transform: scale(0.99);
		background: #f8fafc;
	}
}

.leave-card-status-bar {
	width: 8rpx;
	flex-shrink: 0;

	&--passed { background: linear-gradient(180deg, var(--success-color), #059669); }
	&--rejected { background: linear-gradient(180deg, var(--error-color), #dc2626); }
	&--pending { background: linear-gradient(180deg, var(--warning-color), #d97706); }
}

.leave-card-body {
	flex: 1;
	padding: 24rpx;
}

.leave-card-header {
	display: flex;
	justify-content: flex-start;
	align-items: center;
	gap: 12rpx;
	margin-bottom: 16rpx;
}

.leave-card-tag {
	display: flex;
	align-items: center;
	gap: 8rpx;
	padding: 6rpx 16rpx;
	border-radius: 100rpx;
	font-size: 22rpx;
	font-weight: 600;

	&--passed { background: var(--success-soft); color: var(--success-color); }
	&--rejected { background: var(--error-soft); color: var(--error-color); }
	&--pending { background: rgba(245, 158, 11, 0.1); color: #d97706; }
}

.leave-card-tag-dot {
	width: 12rpx;
	height: 12rpx;
	border-radius: 50%;

	&--passed { background: var(--success-color); }
	&--rejected { background: var(--error-color); }
	&--pending { background: #d97706; }
}

.leave-card-tag-text {
	font-size: 22rpx;
}

.leave-card-status-badge {
	padding: 4rpx 16rpx;
	border-radius: 100rpx;
	font-size: 22rpx;
	font-weight: 600;
	margin-left: 8rpx;

	&--passed { background: var(--success-soft); color: var(--success-color); }
	&--rejected { background: var(--error-soft); color: var(--error-color); }
	&--pending { background: rgba(245, 158, 11, 0.1); color: #d97706; }
}

.leave-card-status-badge-text {
	font-size: 22rpx;
}

.leave-card-title {
	font-size: 32rpx;
	font-weight: 700;
	color: var(--text-primary);
	line-height: 1.4;
	margin-bottom: 20rpx;
}

.leave-card-footer {
	border-top: 1px solid rgba(226, 232, 240, 0.8);
	padding-top: 16rpx;
}

.leave-card-info {
	display: flex;
	flex-direction: column;
	gap: 8rpx;
}

.leave-card-row {
	display: flex;
	align-items: center;
	gap: 8rpx;
	font-size: 24rpx;
	color: var(--text-secondary);
}

.leave-card-icon {
	color: var(--text-tertiary);
}

.leave-card-attachments {
	display: flex;
	align-items: center;
	gap: 6rpx;
	margin-top: 12rpx;
	padding-top: 12rpx;
	border-top: 1px dashed rgba(226, 232, 240, 0.6);
}

.leave-card-attachments-text {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.leave-card-expand-hint {
	margin-left: auto;
	font-size: 22rpx;
	color: var(--primary-500);
	font-weight: 500;
}

/* ---- Expanded Detail ---- */
.leave-card-detail {
	margin-top: 20rpx;
}

.leave-card-detail-divider {
	height: 1px;
	background: rgba(226, 232, 240, 0.8);
	margin-bottom: 20rpx;
}

.leave-card-detail-grid {
	display: grid;
	grid-template-columns: repeat(2, 1fr);
	gap: 16rpx;
	margin-bottom: 20rpx;
}

.leave-card-detail-item {
	display: flex;
	flex-direction: column;
	gap: 4rpx;

	&--full {
		grid-column: span 2;
	}
}

.leave-card-detail-label {
	font-size: 22rpx;
	color: var(--text-tertiary);
	font-weight: 500;
}

.leave-card-detail-value {
	font-size: 26rpx;
	color: var(--text-primary);
	line-height: 1.4;
}

/* ---- Audit Timeline ---- */
.leave-card-audit {
	margin-bottom: 20rpx;
}

.leave-card-audit-title {
	display: block;
	font-size: 22rpx;
	color: var(--text-tertiary);
	font-weight: 600;
	margin-bottom: 16rpx;
	text-transform: uppercase;
	letter-spacing: 0.5px;
}

.leave-card-audit-item {
	display: flex;
	align-items: flex-start;
	gap: 16rpx;
	margin-bottom: 16rpx;

	&:last-child {
		margin-bottom: 0;
	}
}

.leave-card-audit-dot {
	width: 16rpx;
	height: 16rpx;
	border-radius: 50%;
	margin-top: 6rpx;
	flex-shrink: 0;

	&--passed { background: var(--success-color); }
	&--rejected { background: var(--error-color); }
	&--pending { background: var(--warning-color); }
}

.leave-card-audit-content {
	display: flex;
	flex-direction: column;
	gap: 4rpx;
}

.leave-card-audit-teacher {
	font-size: 26rpx;
	font-weight: 600;
	color: var(--text-primary);
}

.leave-card-audit-time {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.leave-card-audit-memo {
	font-size: 24rpx;
	color: var(--text-secondary);
	margin-top: 4rpx;
}

/* ---- Attachment Preview ---- */
.leave-card-attachments-preview {
	margin-bottom: 16rpx;
}

.leave-card-attachments-preview-title {
	display: block;
	font-size: 22rpx;
	color: var(--text-tertiary);
	font-weight: 600;
	margin-bottom: 12rpx;
	text-transform: uppercase;
	letter-spacing: 0.5px;
}

.leave-card-attachments-scroll {
	white-space: nowrap;
}

.leave-card-attachment {
	display: inline-flex;
	flex-direction: column;
	align-items: center;
	margin-right: 16rpx;
	width: 160rpx;
}

.leave-card-attachment__image {
	width: 160rpx;
	height: 120rpx;
	border-radius: 16rpx;
	background: var(--bg-secondary);
	border: 1px solid rgba(226, 232, 240, 0.8);
}

.leave-card-attachment__file {
	width: 160rpx;
	height: 120rpx;
	border-radius: 16rpx;
	background: rgba(239, 68, 68, 0.06);
	border: 1px solid rgba(239, 68, 68, 0.12);
	display: flex;
	align-items: center;
	justify-content: center;
}

.leave-card-attachment__name {
	font-size: 20rpx;
	color: var(--text-secondary);
	margin-top: 8rpx;
	text-align: center;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
	width: 100%;
}

.leave-card-attachment__size {
	font-size: 18rpx;
	color: var(--text-tertiary);
}

/* ---- Return School Badge ---- */
.leave-card-return {
	display: flex;
	align-items: center;
	gap: 12rpx;
	flex-wrap: wrap;
}

.leave-card-return-badge {
	padding: 4rpx 16rpx;
	border-radius: 100rpx;
	font-size: 22rpx;
	font-weight: 600;

	&--yes {
		background: var(--success-soft);
		color: var(--success-color);
	}
	&--no {
		background: var(--error-soft);
		color: var(--error-color);
	}
}

.leave-card-return-badge-text {
	font-size: 22rpx;
}

.leave-card-return-memo {
	font-size: 24rpx;
	color: var(--text-secondary);
}

/* ---- Load More ---- */
.load-more {
	display: flex;
	justify-content: center;
	align-items: center;
	padding: 32rpx 0;
}

.load-more__text {
	font-size: 26rpx;
	color: #d97706;
	font-weight: 600;
}

.state-spinner {
	width: 48rpx;
	height: 48rpx;
	border: 4rpx solid rgba(245, 158, 11, 0.15);
	border-top-color: #d97706;
	border-radius: 50%;
	animation: leave-spin 0.8s linear infinite;
}

.state-spinner--small {
	width: 32rpx;
	height: 32rpx;
	border-width: 3rpx;
}
</style>
