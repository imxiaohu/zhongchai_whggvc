<template>
	<view v-if="visible" class="report-modal" :class="themeClass">
		<view class="modal-overlay" @click="closeModal"></view>
		<view class="modal-content" :class="{ 'modal-show': visible }">
			<view class="modal-header">
				<text class="modal-title">举报内容</text>
				<button class="close-btn" @click="closeModal">
					<text class="close-icon">×</text>
				</button>
			</view>

			<view class="modal-body">
				<view class="form-group">
					<text class="label">举报原因</text>
					<view class="reason-list">
						<view
							v-for="reason in reportReasons"
							:key="reason.value"
							class="reason-item"
							:class="{ 'selected': selectedReason === reason.value }"
							@click="selectReason(reason.value)"
						>
							<view class="reason-radio">
								<text v-if="selectedReason === reason.value" class="radio-checked">●</text>
								<text v-else class="radio-unchecked">○</text>
							</view>
							<text class="reason-text">{{ reason.label }}</text>
						</view>
					</view>
				</view>

				<view class="form-group">
					<text class="label">详细描述</text>
					<textarea
						v-model="description"
						class="description-input"
						placeholder="请详细描述举报原因..."
						:maxlength="500"
					></textarea>
					<text class="char-count">{{ description.length }}/500</text>
				</view>
			</view>

			<view class="modal-footer">
				<button class="cancel-btn" @click="closeModal">取消</button>
				<button
					class="submit-btn"
					:class="{ 'loading': loading }"
					@click="submitReport"
					:disabled="!selectedReason || loading"
				>
					<text v-if="loading">提交中...</text>
					<text v-else>提交</text>
				</button>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { submitReport as submitReportApi } from '@/pages/api/community.js'

const props = defineProps({
	visible: {
		type: Boolean,
		default: false
	},
	targetType: {
		type: String,
		required: true
	},
	targetId: {
		type: [Number, String, null],
		default: null
	}
})

const emit = defineEmits(['close', 'success'])

const selectedReason = ref('')
const description = ref('')
const loading = ref(false)
const isDarkMode = ref(false)
let themeWxCallback = null
let themeH5Callback = null
let themeMediaQuery = null
let themeAppCallback = null

const reportReasons = [
	{ value: 'spam', label: '垃圾广告' },
	{ value: 'inappropriate', label: '言语不当' },
	{ value: 'harassment', label: '骚扰辱骂' },
	{ value: 'infringement', label: '侵权盗用' },
	{ value: 'other', label: '其他' }
]

const themeClass = computed(() =>
	isDarkMode.value ? 'theme-dark' : 'theme-light'
)

onMounted(() => {
	initTheme()
})

function initTheme() {
	try {
		const systemInfo = uni.getSystemInfoSync()
		if (systemInfo.theme) {
			isDarkMode.value = systemInfo.theme === 'dark'
		}
		listenThemeChange()
	} catch (error) {
		console.error('ReportModal: 初始化主题失败:', error)
		isDarkMode.value = false
	}
}

function listenThemeChange() {
	// #ifdef MP-WEIXIN
	themeWxCallback = (res) => {
		isDarkMode.value = res.theme === 'dark'
	}
	uni.onThemeChange(themeWxCallback)
	// #endif

	// #ifdef H5
	if (typeof window !== 'undefined' && window.matchMedia) {
		themeH5Callback = (e) => {
			isDarkMode.value = e.matches
		}
		const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
		if (mediaQuery.addEventListener) {
			mediaQuery.addEventListener('change', themeH5Callback)
		}
		themeMediaQuery = mediaQuery
		isDarkMode.value = mediaQuery.matches
	}
	// #endif

	// #ifdef APP-PLUS
	themeAppCallback = (res) => {
		isDarkMode.value = res.theme === 'dark'
	}
	uni.onThemeChange(themeAppCallback)
	// #endif
}

function cleanupThemeListeners() {
	// #ifdef MP-WEIXIN
	if (themeWxCallback) {
		uni.offThemeChange(themeWxCallback)
		themeWxCallback = null
	}
	// #endif
	// #ifdef H5
	if (themeMediaQuery && themeH5Callback) {
		themeMediaQuery.removeEventListener('change', themeH5Callback)
		themeMediaQuery = null
		themeH5Callback = null
	}
	// #endif
	// #ifdef APP-PLUS
	if (themeAppCallback) {
		uni.offThemeChange(themeAppCallback)
		themeAppCallback = null
	}
	// #endif
}

onBeforeUnmount(() => {
	cleanupThemeListeners()
})

function selectReason(reason) {
	selectedReason.value = reason
}

function closeModal() {
	resetForm()
	emit('close')
}

function resetForm() {
	selectedReason.value = ''
	description.value = ''
	loading.value = false
}

async function submitReport() {
	if (!props.targetId) {
		uni.showToast({ title: '帖子信息加载中，请稍后重试', icon: 'none' })
		return
	}
	if (!selectedReason.value) {
		uni.showToast({ title: '请选择举报原因', icon: 'none' })
		return
	}

	loading.value = true
	try {
		const response = await submitReportApi({
			targetType: props.targetType,
			targetId: props.targetId,
			reason: selectedReason.value,
			description: description.value
		})

		if (response.success) {
			uni.showToast({ title: '举报提交成功', icon: 'success' })
			closeModal()
			emit('success')
		} else {
			throw new Error(response.message || '举报失败')
		}
	} catch (error) {
		console.error('提交举报失败:', error)
		uni.showToast({
			title: error.message || '举报失败',
			icon: 'none'
		})
	} finally {
		loading.value = false
	}
}
</script>

<style lang="scss" scoped>
.report-modal {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	z-index: 9999;
	pointer-events: auto;
}

.modal-overlay {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background-color: rgba(0, 0, 0, 0.5);
	backdrop-filter: blur(4px);
}

.modal-content {
	position: absolute;
	top: 50%;
	left: 50%;
	transform: translate(-50%, -50%) scale(0.9);
	width: 90%;
	max-width: 500rpx;
	max-height: 80vh;
	background-color: white;
	border-radius: 24rpx;
	box-shadow: 0 20rpx 60rpx rgba(0, 0, 0, 0.3);
	opacity: 0;
	transition: all 0.3s ease;
	overflow: hidden;

	&.modal-show {
		transform: translate(-50%, -50%) scale(1);
		opacity: 1;
	}
}

.modal-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 32rpx;
	border-bottom: 2rpx solid #f0f0f0;
}

.modal-title {
	font-size: 36rpx;
	font-weight: 600;
	color: #333;
}

.close-btn {
	width: 60rpx;
	height: 60rpx;
	border-radius: 50%;
	background-color: #f5f5f5;
	border: none;
	display: flex;
	align-items: center;
	justify-content: center;
	transition: background-color 0.3s ease;

	&:hover {
		background-color: #e0e0e0;
	}
}

.close-icon {
	font-size: 40rpx;
	color: #666;
	line-height: 1;
}

.modal-body {
	padding: 32rpx;
	max-height: 60vh;
	overflow-y: auto;
}

.form-group {
	margin-bottom: 32rpx;

	&:last-child {
		margin-bottom: 0;
	}
}

.label {
	display: block;
	font-size: 32rpx;
	font-weight: 500;
	color: #333;
	margin-bottom: 16rpx;
}

.reason-list {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
}

.reason-item {
	display: flex;
	align-items: center;
	padding: 20rpx;
	border: 2rpx solid #e0e0e0;
	border-radius: 16rpx;
	cursor: pointer;
	transition: all 0.3s ease;

	&.selected {
		border-color: #ff6b35;
		background-color: rgba(255, 107, 53, 0.1);
	}

	&:hover {
		border-color: #ff6b35;
	}
}

.reason-radio {
	margin-right: 16rpx;
}

.radio-checked,
.radio-unchecked {
	font-size: 32rpx;
	color: #ff6b35;
}

.radio-unchecked {
	color: #ccc;
}

.reason-text {
	font-size: 30rpx;
	color: #333;
}

.description-input {
	width: 100%;
	min-height: 200rpx;
	padding: 20rpx;
	border: 2rpx solid #e0e0e0;
	border-radius: 16rpx;
	font-size: 30rpx;
	color: #333;
	background-color: #fafafa;
	resize: none;
	box-sizing: border-box;

	&:focus {
		border-color: #ff6b35;
		background-color: white;
	}
}

.char-count {
	display: block;
	text-align: right;
	font-size: 24rpx;
	color: #999;
	margin-top: 8rpx;
}

.modal-footer {
	display: flex;
	gap: 16rpx;
	padding: 32rpx;
	border-top: 2rpx solid #f0f0f0;
}

.cancel-btn,
.submit-btn {
	flex: 1;
	height: 80rpx;
	border-radius: 20rpx;
	font-size: 32rpx;
	font-weight: 500;
	border: none;
	transition: all 0.3s ease;

	&:active {
		transform: scale(0.95);
	}
}

.cancel-btn {
	background-color: #f5f5f5;
	color: #666;

	&:hover {
		background-color: #e0e0e0;
	}
}

.submit-btn {
	background-color: #ff6b35;
	color: white;

	&:hover {
		background-color: #e55a2b;
	}

	&:disabled {
		background-color: #ccc;
		cursor: not-allowed;
	}

	&.loading {
		opacity: 0.7;
	}
}

/* 主题适配 */
.theme-dark {
	.modal-content {
		background-color: #2a2a2a;
	}

	.modal-header {
		border-bottom-color: #444;
	}

	.modal-title {
		color: #e0e0e0;
	}

	.close-btn {
		background-color: #444;

		&:hover {
			background-color: #555;
		}
	}

	.close-icon {
		color: #ccc;
	}

	.label {
		color: #e0e0e0;
	}

	.reason-item {
		border-color: #444;
		background-color: #333;

		&.selected {
			background-color: rgba(255, 107, 53, 0.2);
		}

		&:hover {
			border-color: #ff6b35;
		}
	}

	.reason-text {
		color: #e0e0e0;
	}

	.description-input {
		border-color: #444;
		background-color: #333;
		color: #e0e0e0;

		&:focus {
			background-color: #2a2a2a;
		}
	}

	.modal-footer {
		border-top-color: #444;
	}

	.cancel-btn {
		background-color: #444;
		color: #ccc;

		&:hover {
			background-color: #555;
		}
	}
}

/* 响应式设计 */
@media screen and (max-width: 480px) {
	.modal-content {
		width: 95%;
		max-width: none;
	}

	.modal-header,
	.modal-body,
	.modal-footer {
		padding: 24rpx;
	}

	.modal-title {
		font-size: 32rpx;
	}

	.reason-item {
		padding: 16rpx;
	}

	.reason-text {
		font-size: 28rpx;
	}

	.description-input {
		min-height: 160rpx;
		font-size: 28rpx;
	}
}
</style>
