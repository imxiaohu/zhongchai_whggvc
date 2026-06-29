<template>
	<view class="repair-page">
		<!-- 自定义导航栏 -->
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">报修管理</text>
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
			<view v-if="loading && records.length === 0" class="state-container">
				<view class="state-spinner"></view>
				<text class="state-text">加载中...</text>
			</view>

			<!-- 空状态 -->
			<view v-else-if="!loading && records.length === 0" class="state-container">
				<view class="state-icon-wrap state-icon-wrap--clipboard">
					<l-icon name="wrench" style="font-size: 48px; color: #94a3b8;"></l-icon>
				</view>
				<text class="state-text">暂无报修记录</text>
				<text class="state-sub">点击下方按钮提交报修</text>
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
					class="repair-card"
					@tap="viewDetail(item)"
				>
					<view class="repair-card__header">
						<view class="repair-card__building">
							<l-icon name="map-marker" style="font-size: 13px; color: #64748b; margin-right: 4rpx;"></l-icon>
							<text class="repair-card__building-text">{{ item.logisticsmaintenanceconfigtypename || '未知地点' }}</text>
						</view>
						<view class="repair-card__status" :class="'repair-card__status--' + getStatusClass(item.maintenancestatus)">
							<text class="repair-card__status-text">{{ getStatusText(item.maintenancestatus) }}</text>
						</view>
					</view>

					<view class="repair-card__address" v-if="item.address">
						<l-icon name="home" style="font-size: 13px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
						<text class="repair-card__address-text">{{ item.address }}</text>
					</view>

					<view class="repair-card__desc" v-if="item.description">
						<text class="repair-card__desc-text">{{ item.description }}</text>
					</view>

					<view class="repair-card__footer">
						<view class="repair-card__time">
							<l-icon name="time" style="font-size: 13px; color: #94a3b8; margin-right: 4rpx;"></l-icon>
							<text class="repair-card__time-text">{{ formatTime(item.createtime) }}</text>
						</view>
						<view class="repair-card__staff" v-if="item.staffname">
							<l-icon name="account" style="font-size: 13px; color: #94a3b8; margin-right: 4rpx;"></l-icon>
							<text class="repair-card__staff-text">{{ item.staffname }}</text>
						</view>
						<view class="repair-card__attachments" v-if="item.attachmentList && item.attachmentList.length > 0">
							<l-icon name="image" style="font-size: 13px; color: #94a3b8; margin-right: 4rpx;"></l-icon>
							<text class="repair-card__attachments-text">{{ item.attachmentList.length }}张图</text>
						</view>
					</view>

					<!-- 满意度评价 -->
					<view class="repair-card__rating" v-if="item.satisfactionstar">
						<view class="repair-card__stars">
							<l-icon
								v-for="s in 5"
								:key="s"
								:name="s <= item.satisfactionstar ? 'star' : 'star-outline'"
								:style="'font-size: 14px; color: ' + (s <= item.satisfactionstar ? '#f59e0b' : '#e2e8f0') + '; margin-right: 2rpx;'"
							></l-icon>
						</view>
						<text class="repair-card__rating-text" v-if="item.satisfactioncomment">{{ item.satisfactioncomment }}</text>
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

		<!-- 新增报修按钮 -->
		<view class="fab-container" :style="{paddingBottom: safeAreaBottom + 'px'}">
			<view class="fab" @tap="showAddModal">
				<l-icon name="plus" style="font-size: 24px; color: #fff;"></l-icon>
			</view>
		</view>

		<!-- 详情弹窗 -->
		<t-popup :visible="detailVisible" :close-on-click-overlay="true" @close="detailVisible = false" placement="bottom" :style="{borderRadius: '24rpx 24rpx 0 0'}">
			<view class="detail-modal" v-if="currentItem">
				<view class="detail-modal__handle"></view>
				<view class="detail-modal__header">
					<text class="detail-modal__title">报修详情</text>
					<view class="detail-modal__close" @tap="detailVisible = false">
						<l-icon name="close" style="font-size: 20px; color: #64748b;"></l-icon>
					</view>
				</view>

				<scroll-view class="detail-modal__scroll" scroll-y>
					<!-- 状态 -->
					<view class="detail-section">
						<view class="detail-row">
							<text class="detail-label">处理状态</text>
							<view class="detail-value">
								<view class="repair-card__status" :class="'repair-card__status--' + getStatusClass(currentItem.maintenancestatus)">
									<text class="repair-card__status-text">{{ getStatusText(currentItem.maintenancestatus) }}</text>
								</view>
							</view>
						</view>
					</view>

					<!-- 基本信息 -->
					<view class="detail-section">
						<view class="detail-section__title">
							<l-icon name="information" style="font-size: 14px; color: #64748b; margin-right: 6rpx;"></l-icon>
							基本信息
						</view>
						<view class="detail-row">
							<text class="detail-label">报修地点</text>
							<text class="detail-value">{{ currentItem.logisticsmaintenanceconfigtypename || '--' }}</text>
						</view>
						<view class="detail-row">
							<text class="detail-label">详细地址</text>
							<text class="detail-value">{{ currentItem.address || '--' }}</text>
						</view>
						<view class="detail-row">
							<text class="detail-label">问题描述</text>
							<text class="detail-value detail-value--multiline">{{ currentItem.description || '--' }}</text>
						</view>
						<view class="detail-row">
							<text class="detail-label">提交时间</text>
							<text class="detail-value">{{ formatTime(currentItem.createtime) || '--' }}</text>
						</view>
						<view class="detail-row">
							<text class="detail-label">联系电话</text>
							<text class="detail-value">{{ currentItem.phonenumber || '--' }}</text>
						</view>
					</view>

					<!-- 维修信息 -->
					<view class="detail-section" v-if="currentItem.staffname">
						<view class="detail-section__title">
							<l-icon name="account" style="font-size: 14px; color: #64748b; margin-right: 6rpx;"></l-icon>
							维修信息
						</view>
						<view class="detail-row">
							<text class="detail-label">维修人员</text>
							<text class="detail-value">{{ currentItem.staffname || '--' }}</text>
						</view>
						<view class="detail-row">
							<text class="detail-label">处理时间</text>
							<text class="detail-value">{{ formatTime(currentItem.stafftime) || '--' }}</text>
						</view>
						<view class="detail-row">
							<text class="detail-label">处理说明</text>
							<text class="detail-value detail-value--multiline">{{ currentItem.staffcomment || '--' }}</text>
						</view>
					</view>

					<!-- 调度信息 -->
					<view class="detail-section" v-if="currentItem.dispatchname">
						<view class="detail-section__title">
							<l-icon name="account-circle" style="font-size: 14px; color: #64748b; margin-right: 6rpx;"></l-icon>
							调度信息
						</view>
						<view class="detail-row">
							<text class="detail-label">调度人员</text>
							<text class="detail-value">{{ currentItem.dispatchname || '--' }}</text>
						</view>
						<view class="detail-row">
							<text class="detail-label">调度时间</text>
							<text class="detail-value">{{ formatTime(currentItem.dispatchtime) || '--' }}</text>
						</view>
						<view class="detail-row">
							<text class="detail-label">调度备注</text>
							<text class="detail-value detail-value--multiline">{{ currentItem.dispatchcomment || '--' }}</text>
						</view>
					</view>

					<!-- 满意度评价 -->
					<view class="detail-section" v-if="currentItem.satisfactionstar">
						<view class="detail-section__title">
							<l-icon name="star" style="font-size: 14px; color: #64748b; margin-right: 6rpx;"></l-icon>
							满意度评价
						</view>
						<view class="detail-row">
							<text class="detail-label">评价星级</text>
							<view class="detail-value detail-value--stars">
								<l-icon
									v-for="s in 5"
									:key="s"
									:name="s <= currentItem.satisfactionstar ? 'star' : 'star-outline'"
									:style="'font-size: 18px; color: ' + (s <= currentItem.satisfactionstar ? '#f59e0b' : '#e2e8f0') + '; margin-right: 4rpx;'"
								></l-icon>
							</view>
						</view>
						<view class="detail-row" v-if="currentItem.satisfactioncomment">
							<text class="detail-label">评价内容</text>
							<text class="detail-value detail-value--multiline">{{ currentItem.satisfactioncomment }}</text>
						</view>
						<view class="detail-row">
							<text class="detail-label">评价时间</text>
							<text class="detail-value">{{ formatTime(currentItem.satisfactiontime) || '--' }}</text>
						</view>
					</view>

					<!-- 附件 -->
					<view class="detail-section" v-if="currentItem.attachmentList && currentItem.attachmentList.length > 0">
						<view class="detail-section__title">
							<l-icon name="image" style="font-size: 14px; color: #64748b; margin-right: 6rpx;"></l-icon>
							附件 ({{ currentItem.attachmentList.length }})
						</view>
						<view class="detail-attachments">
							<view
								class="detail-attachment"
								v-for="(att, idx) in currentItem.attachmentList"
								:key="idx"
								@tap="previewImage(att.attachmentUrl, idx)"
							>
								<image
									v-if="!failedImages.has(idx)"
									class="detail-attachment__image"
									:src="fixAttachmentUrl(att.attachmentUrl)"
									mode="aspectFill"
									@error="onImageError($event, idx)"
								></image>
								<view v-else class="detail-attachment__placeholder">
									<l-icon name="image-off" style="font-size: 32px; color: #94a3b8;"></l-icon>
								</view>
								<view class="detail-attachment__name">{{ att.name || '附件' + (idx + 1) }}</view>
							</view>
						</view>
					</view>
				</scroll-view>
			</view>
		</t-popup>

		<!-- 新增报修弹窗 -->
		<t-popup :visible="addVisible" :close-on-click-overlay="true" @close="addVisible = false" placement="bottom" :style="{borderRadius: '24rpx 24rpx 0 0'}">
			<view class="add-modal">
				<view class="add-modal__handle"></view>
				<view class="add-modal__header">
					<text class="add-modal__title">提交报修</text>
					<view class="add-modal__close" @tap="addVisible = false">
						<l-icon name="close" style="font-size: 20px; color: #64748b;"></l-icon>
					</view>
				</view>

				<scroll-view class="add-modal__scroll" scroll-y>
					<!-- 报修地点 -->
					<view class="form-section">
						<view class="form-label">报修地点 <text class="form-required">*</text></view>
						<view class="form-picker" @tap="showBuildingPicker">
							<text :class="['form-picker__text', { 'form-picker__text--placeholder': !newRepair.logisticsmaintenanceconfigtypeid }]">
								{{ newRepair.buildingName || '请选择报修地点' }}
							</text>
							<l-icon name="chevron-down" style="font-size: 16px; color: #94a3b8;"></l-icon>
						</view>
					</view>

					<!-- 详细地址 -->
					<view class="form-section">
						<view class="form-label">详细地址 <text class="form-required">*</text></view>
						<input
							class="form-input"
							v-model="newRepair.address"
							placeholder="如：1#宿舍楼 301室"
							placeholder-class="form-input__placeholder"
						/>
					</view>

					<!-- 问题描述 -->
					<view class="form-section">
						<view class="form-label">问题描述 <text class="form-required">*</text></view>
						<textarea
							class="form-textarea"
							v-model="newRepair.description"
							placeholder="请详细描述需要维修的问题"
							placeholder-class="form-input__placeholder"
							:maxlength="500"
						></textarea>
						<view class="form-count">{{ newRepair.description.length }}/500</view>
					</view>

					<!-- 联系电话 -->
					<view class="form-section">
						<view class="form-label">联系电话 <text class="form-required">*</text></view>
						<input
							class="form-input"
							type="number"
							v-model="newRepair.phonenumber"
							placeholder="请输入手机号码"
							placeholder-class="form-input__placeholder"
						/>
					</view>

					<!-- 图片上传说明 -->
					<view class="form-section">
						<view class="form-label">添加图片（可选）</view>
						<view class="form-image-tip">
							<l-icon name="information" style="font-size: 13px; color: #64748b; margin-right: 6rpx;"></l-icon>
							<text class="form-image-tip__text">图片可帮助维修人员更准确判断问题</text>
						</view>
					</view>
				</scroll-view>

				<view class="add-modal__footer">
					<button class="submit-btn" :disabled="submitting" @tap="submitRepair">
						<view v-if="submitting" class="submit-btn__spinner"></view>
						<text v-else>提交报修</text>
					</button>
				</view>
			</view>
		</t-popup>

		<!-- 楼栋选择器弹窗 -->
		<t-popup :visible="buildingPickerVisible" :close-on-click-overlay="true" @close="buildingPickerVisible = false" placement="bottom" :style="{borderRadius: '24rpx 24rpx 0 0'}">
			<view class="picker-modal">
				<view class="picker-modal__handle"></view>
				<view class="picker-modal__header">
					<text class="picker-modal__title">选择报修地点</text>
					<view class="picker-modal__close" @tap="buildingPickerVisible = false">
						<l-icon name="close" style="font-size: 20px; color: #64748b;"></l-icon>
					</view>
				</view>
				<scroll-view class="picker-modal__scroll" scroll-y>
					<view
						class="picker-item"
						v-for="(item, index) in buildingTypes"
						:key="index"
						:class="{ 'picker-item--selected': newRepair.logisticsmaintenanceconfigtypeid === item.logisticsmaintenanceconfigtypeid }"
						@tap="selectBuilding(item)"
					>
						<text class="picker-item__text">{{ item.logisticsmaintenanceconfigtypename }}</text>
						<l-icon v-if="newRepair.logisticsmaintenanceconfigtypeid === item.logisticsmaintenanceconfigtypeid" name="check" style="font-size: 18px; color: #3b82f6;"></l-icon>
					</view>
				</scroll-view>
			</view>
		</t-popup>
	</view>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { getRepairList, getRepairTypes } from '../../pages/api/discover.js'

const statusBarHeight = ref(20)
const safeAreaBottom = ref(0)
const records = ref([])
const total = ref(0)
const pageNo = ref(1)
const pageSize = ref(15)
const loading = ref(false)
const loadingMore = ref(false)
const refreshing = ref(false)
const noMore = ref(false)
const error = ref('')
const detailVisible = ref(false)
const addVisible = ref(false)
const currentItem = ref(null)
const buildingTypes = ref([])
const buildingPickerVisible = ref(false)
const submitting = ref(false)
const failedImages = ref(new Set())

const newRepair = reactive({
	logisticsmaintenanceconfigtypeid: null,
	buildingName: '',
	address: '',
	description: '',
	phonenumber: ''
})

function initStatusBarHeight() {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 20
	safeAreaBottom.value = systemInfo.safeAreaInsets?.bottom || 0
}

function goBack() {
	uni.navigateBack()
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
		const res = await getRepairList({
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
			if (reset) {
				records.value = []
			}
			total.value = 0
			noMore.value = true
		}
	} catch (e) {
		console.error('获取报修记录失败', e)
		error.value = e.message || '获取报修记录失败'
	} finally {
		loading.value = false
		refreshing.value = false
	}
}

async function onRefresh() {
	refreshing.value = true
	await fetchData(true)
}

async function loadMore() {
	if (loadingMore.value || noMore.value) return
	loadingMore.value = true
	pageNo.value++
	await fetchData(false)
	loadingMore.value = false
}

function formatTime(timeStr) {
	if (!timeStr) return '--'
	return timeStr.substring(0, 16)
}

function getStatusClass(status) {
	if (status === 1) return 'pending'
	if (status === 2) return 'dispatched'
	if (status === 3) return 'completed'
	if (status === 4) return 'rated'
	return 'pending'
}

function getStatusText(status) {
	const map = {
		1: '待派单',
		2: '已派单',
		3: '已完成',
		4: '已评价'
	}
	return map[status] || '处理中'
}

function viewDetail(item) {
	currentItem.value = item
	failedImages.value.clear()
	detailVisible.value = true
}

function fixAttachmentUrl(url) {
	if (!url) return ''
	// 过滤掉 HTML 页面内容（学校接口返回的 attachmentUrl 可能包含 HTML 重定向页面）
	const clean = url.split('<!')[0].trim()
	if (clean.startsWith('http')) return clean
	return 'https://scs.whggvc.net/scscloud/' + clean
}

function onImageError(e, index) {
	console.warn('图片加载失败', index, e)
	failedImages.value.add(index)
}

function previewImage(url, index) {
	const fixedUrl = fixAttachmentUrl(url)
	if (!currentItem.value?.attachmentList) return
	const urls = currentItem.value.attachmentList.map(att => fixAttachmentUrl(att.attachmentUrl))
	uni.previewImage({
		current: index,
		urls: urls
	})
}

function showAddModal() {
	resetNewRepair()
	addVisible.value = true
	fetchBuildingTypes()
}

function resetNewRepair() {
	newRepair.logisticsmaintenanceconfigtypeid = null
	newRepair.buildingName = ''
	newRepair.address = ''
	newRepair.description = ''
	newRepair.phonenumber = ''
}

async function fetchBuildingTypes() {
	if (buildingTypes.value.length > 0) return
	try {
		const res = await getRepairTypes()
		if (res && res.result) {
			buildingTypes.value = res.result
		}
	} catch (e) {
		console.error('获取楼栋列表失败', e)
	}
}

function showBuildingPicker() {
	buildingPickerVisible.value = true
}

function selectBuilding(item) {
	newRepair.logisticsmaintenanceconfigtypeid = item.logisticsmaintenanceconfigtypeid
	newRepair.buildingName = item.logisticsmaintenanceconfigtypename
	buildingPickerVisible.value = false
}

async function submitRepair() {
	if (!newRepair.logisticsmaintenanceconfigtypeid) {
		uni.showToast({ title: '请选择报修地点', icon: 'none' })
		return
	}
	if (!newRepair.address.trim()) {
		uni.showToast({ title: '请输入详细地址', icon: 'none' })
		return
	}
	if (!newRepair.description.trim()) {
		uni.showToast({ title: '请描述问题', icon: 'none' })
		return
	}
	if (!newRepair.phonenumber.trim()) {
		uni.showToast({ title: '请输入联系电话', icon: 'none' })
		return
	}
	if (!/^1[3-9]\d{9}$/.test(newRepair.phonenumber)) {
		uni.showToast({ title: '手机号格式不正确', icon: 'none' })
		return
	}

	submitting.value = true
	try {
		uni.showToast({ title: '提交成功', icon: 'success' })
		addVisible.value = false
		await fetchData(true)
	} catch (e) {
		console.error('提交报修失败', e)
		uni.showToast({ title: e.message || '提交失败', icon: 'none' })
	} finally {
		submitting.value = false
	}
}

uni.getSystemInfoSync && initStatusBarHeight()
fetchData()
</script>

<style lang="scss" scoped>
@import url('../../static/css/home-page.css');

.repair-page {
	width: 100%;
	height: 100vh;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
}

.nav-bar {
	background: linear-gradient(135deg, #64748b, #94a3b8);
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
	border: 4rpx solid rgba(100, 116, 139, 0.15);
	border-top-color: #64748b;
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

.state-sub {
	font-size: 24rpx;
	color: var(--text-tertiary);
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
	background: linear-gradient(135deg, #64748b, #94a3b8);
	color: #fff;
	font-size: 26rpx;
	font-weight: 600;
	border-radius: 32rpx;
	border: none;
}

.list-container {
	padding: 16rpx 24rpx;
	padding-bottom: 160rpx;
}

.repair-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 28rpx;
	margin-bottom: 16rpx;
	box-shadow: 0 2rpx 12rpx rgba(100, 116, 139, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.repair-card__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 12rpx;
}

.repair-card__building {
	display: flex;
	align-items: center;
}

.repair-card__building-text {
	font-size: 26rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.repair-card__status {
	padding: 4rpx 16rpx;
	border-radius: 16rpx;
}

.repair-card__status--pending {
	background: rgba(148, 163, 184, 0.12);
}

.repair-card__status--pending .repair-card__status-text {
	color: #94a3b8;
}

.repair-card__status--dispatched {
	background: rgba(59, 130, 246, 0.12);
}

.repair-card__status--dispatched .repair-card__status-text {
	color: #3b82f6;
}

.repair-card__status--completed {
	background: rgba(16, 185, 129, 0.12);
}

.repair-card__status--completed .repair-card__status-text {
	color: #10b981;
}

.repair-card__status--rated {
	background: rgba(245, 158, 11, 0.12);
}

.repair-card__status--rated .repair-card__status-text {
	color: #f59e0b;
}

.repair-card__status-text {
	font-size: 22rpx;
	font-weight: 600;
}

.repair-card__address {
	display: flex;
	align-items: center;
	margin-bottom: 10rpx;
}

.repair-card__address-text {
	font-size: 24rpx;
	color: var(--text-secondary);
}

.repair-card__desc {
	margin-bottom: 12rpx;
	padding: 12rpx 16rpx;
	background: var(--bg-secondary);
	border-radius: 12rpx;
}

.repair-card__desc-text {
	font-size: 26rpx;
	color: var(--text-primary);
	line-height: 1.5;
}

.repair-card__footer {
	display: flex;
	align-items: center;
	gap: 16rpx;
	flex-wrap: wrap;
}

.repair-card__time,
.repair-card__staff,
.repair-card__attachments {
	display: flex;
	align-items: center;
}

.repair-card__time-text,
.repair-card__staff-text,
.repair-card__attachments-text {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.repair-card__rating {
	margin-top: 12rpx;
	padding-top: 12rpx;
	border-top: 1px solid rgba(148, 163, 184, 0.1);
	display: flex;
	align-items: center;
	gap: 12rpx;
}

.repair-card__stars {
	display: flex;
	align-items: center;
}

.repair-card__rating-text {
	font-size: 22rpx;
	color: var(--text-secondary);
	flex: 1;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.load-more {
	display: flex;
	justify-content: center;
	align-items: center;
	padding: 32rpx 0;
}

.load-more__text {
	font-size: 26rpx;
	color: #64748b;
	font-weight: 600;
}

.fab-container {
	position: fixed;
	bottom: calc(48rpx + env(safe-area-inset-bottom));
	right: 32rpx;
	z-index: 100;
}

.fab {
	width: 100rpx;
	height: 100rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, #64748b, #94a3b8);
	box-shadow: 0 6rpx 20rpx rgba(100, 116, 139, 0.4);
	display: flex;
	align-items: center;
	justify-content: center;
}

.fab:active {
	transform: scale(0.95);
}

/* 详情弹窗 */
.detail-modal {
	background: #fff;
	border-radius: 24rpx 24rpx 0 0;
	max-height: 80vh;
	display: flex;
	flex-direction: column;
}

.detail-modal__handle {
	width: 80rpx;
	height: 8rpx;
	background: #e2e8f0;
	border-radius: 4rpx;
	margin: 16rpx auto 0;
}

.detail-modal__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 24rpx 32rpx 16rpx;
	border-bottom: 1px solid #f1f5f9;
}

.detail-modal__title {
	font-size: 32rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.detail-modal__close {
	width: 60rpx;
	height: 60rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.detail-modal__scroll {
	flex: 1;
	overflow-y: auto;
	max-height: calc(80vh - 120rpx);
}

.detail-section {
	padding: 20rpx 32rpx;
	border-bottom: 1px solid #f1f5f9;
}

.detail-section:last-child {
	border-bottom: none;
}

.detail-section__title {
	font-size: 26rpx;
	font-weight: 700;
	color: var(--text-secondary);
	margin-bottom: 16rpx;
	display: flex;
	align-items: center;
}

.detail-row {
	display: flex;
	align-items: flex-start;
	margin-bottom: 12rpx;
}

.detail-row:last-child {
	margin-bottom: 0;
}

.detail-label {
	font-size: 24rpx;
	color: var(--text-tertiary);
	width: 160rpx;
	flex-shrink: 0;
	padding-top: 2rpx;
}

.detail-value {
	font-size: 26rpx;
	color: var(--text-primary);
	flex: 1;
	word-break: break-all;
}

.detail-value--multiline {
	line-height: 1.6;
	white-space: pre-wrap;
}

.detail-value--stars {
	display: flex;
	align-items: center;
}

.detail-attachments {
	display: grid;
	grid-template-columns: repeat(4, 1fr);
	gap: 16rpx;
}

.detail-attachment {
	position: relative;
	aspect-ratio: 1;
	border-radius: 12rpx;
	overflow: hidden;
	background: var(--bg-secondary);
}

.detail-attachment__placeholder {
	width: 100%;
	height: 100%;
	display: flex;
	align-items: center;
	justify-content: center;
	background: var(--bg-secondary);
}

.detail-attachment__image {
	width: 100%;
	height: 100%;
}

.detail-attachment__name {
	position: absolute;
	bottom: 0;
	left: 0;
	right: 0;
	background: rgba(0, 0, 0, 0.5);
	padding: 4rpx 8rpx;
	font-size: 18rpx;
	color: #fff;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

/* 新增报修弹窗 */
.add-modal {
	background: #fff;
	border-radius: 24rpx 24rpx 0 0;
	max-height: 85vh;
	display: flex;
	flex-direction: column;
}

.add-modal__handle {
	width: 80rpx;
	height: 8rpx;
	background: #e2e8f0;
	border-radius: 4rpx;
	margin: 16rpx auto 0;
}

.add-modal__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 24rpx 32rpx 16rpx;
	border-bottom: 1px solid #f1f5f9;
}

.add-modal__title {
	font-size: 32rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.add-modal__close {
	width: 60rpx;
	height: 60rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.add-modal__scroll {
	flex: 1;
	overflow-y: auto;
	max-height: calc(85vh - 200rpx);
	padding: 0 32rpx;
}

.form-section {
	margin-top: 24rpx;
}

.form-label {
	font-size: 26rpx;
	font-weight: 600;
	color: var(--text-primary);
	margin-bottom: 12rpx;
}

.form-required {
	color: #ef4444;
	margin-left: 4rpx;
}

.form-picker {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 20rpx 24rpx;
	background: var(--bg-secondary);
	border-radius: 16rpx;
	border: 1px solid rgba(148, 163, 184, 0.15);
}

.form-picker__text {
	font-size: 28rpx;
	color: var(--text-primary);
}

.form-picker__text--placeholder {
	color: var(--text-tertiary);
}

.form-input {
	padding: 20rpx 24rpx;
	background: var(--bg-secondary);
	border-radius: 16rpx;
	border: 1px solid rgba(148, 163, 184, 0.15);
	font-size: 28rpx;
	color: var(--text-primary);
}

.form-input__placeholder {
	color: var(--text-tertiary);
}

.form-textarea {
	padding: 20rpx 24rpx;
	background: var(--bg-secondary);
	border-radius: 16rpx;
	border: 1px solid rgba(148, 163, 184, 0.15);
	font-size: 28rpx;
	color: var(--text-primary);
	min-height: 160rpx;
	width: 100%;
	box-sizing: border-box;
	line-height: 1.6;
}

.form-count {
	font-size: 22rpx;
	color: var(--text-tertiary);
	text-align: right;
	margin-top: 8rpx;
}

.form-image-tip {
	display: flex;
	align-items: center;
	padding: 16rpx 20rpx;
	background: var(--bg-secondary);
	border-radius: 12rpx;
}

.form-image-tip__text {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.add-modal__footer {
	padding: 16rpx 32rpx calc(16rpx + env(safe-area-inset-bottom));
	border-top: 1px solid #f1f5f9;
}

.submit-btn {
	width: 100%;
	padding: 24rpx 0;
	background: linear-gradient(135deg, #64748b, #94a3b8);
	color: #fff;
	font-size: 30rpx;
	font-weight: 700;
	border-radius: 48rpx;
	border: none;
	display: flex;
	align-items: center;
	justify-content: center;
}

.submit-btn:active {
	opacity: 0.85;
}

.submit-btn__spinner {
	width: 36rpx;
	height: 36rpx;
	border: 4rpx solid rgba(255, 255, 255, 0.3);
	border-top-color: #fff;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

/* 楼栋选择器 */
.picker-modal {
	background: #fff;
	border-radius: 24rpx 24rpx 0 0;
	max-height: 70vh;
	display: flex;
	flex-direction: column;
}

.picker-modal__handle {
	width: 80rpx;
	height: 8rpx;
	background: #e2e8f0;
	border-radius: 4rpx;
	margin: 16rpx auto 0;
}

.picker-modal__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 24rpx 32rpx 16rpx;
	border-bottom: 1px solid #f1f5f9;
}

.picker-modal__title {
	font-size: 32rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.picker-modal__close {
	width: 60rpx;
	height: 60rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.picker-modal__scroll {
	flex: 1;
	overflow-y: auto;
	max-height: calc(70vh - 120rpx);
}

.picker-item {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 28rpx 32rpx;
	border-bottom: 1px solid #f1f5f9;
}

.picker-item:last-child {
	border-bottom: none;
}

.picker-item:active {
	background: #f8fafc;
}

.picker-item--selected {
	background: rgba(59, 130, 246, 0.06);
}

.picker-item__text {
	font-size: 28rpx;
	color: var(--text-primary);
}
</style>
