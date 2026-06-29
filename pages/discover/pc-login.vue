<template>
	<view class="pc-login-page">
		<!-- 自定义导航栏 -->
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">PC端授权</text>
				<view class="nav-bar__placeholder"></view>
			</view>
		</view>

		<!-- 内容区 -->
		<scroll-view class="content-scroll" scroll-y>
			<view class="page-content">

				<!-- 状态展示卡片 -->
			<view class="status-card">
				<view class="status-icon" :class="isLoggedIn ? 'status-icon--success' : 'status-icon--warn'">
					<l-icon :name="isLoggedIn ? 'check-circle' : 'shield-alert'" style="font-size: 48px;"></l-icon>
				</view>
				<text class="status-title">{{ isLoggedIn ? '已授权' : '未授权' }}</text>
				<text class="status-desc">{{ isLoggedIn ? 'PC端会话有效，可使用专属功能' : '自动识别中...' }}</text>
				<view v-if="isLoggedIn && sessionInfo" class="status-detail">
					<view class="detail-row">
						<text class="detail-label">登录时间</text>
						<text class="detail-value">{{ sessionInfo.loginAt }}</text>
					</view>
					<view class="detail-row">
						<text class="detail-label">有效期至</text>
						<text class="detail-value">{{ sessionInfo.expiresAt }}</text>
					</view>
					<view class="detail-row">
						<text class="detail-label">剩余时间</text>
						<text class="detail-value detail-value--highlight">{{ formatTTL(sessionInfo.ttlSeconds) }}</text>
					</view>
				</view>
			</view>

			<!-- 自动登录进度（未登录时显示） -->
			<view v-if="!isLoggedIn && autoLogging" class="auto-progress-card">
				<view class="progress-header">
					<l-icon name="robot" style="font-size: 20px; color: #6366f1; margin-right: 8rpx;"></l-icon>
					<text class="progress-title">自动识别中</text>
					<view class="progress-spinner"></view>
				</view>
				<view class="progress-steps">
					<view class="progress-step" :class="stepClass(1)">
						<view class="step-dot"></view>
						<text class="step-text">获取会话</text>
					</view>
					<view class="progress-line" :class="stepLineClass(1)"></view>
					<view class="progress-step" :class="stepClass(2)">
						<view class="step-dot"></view>
						<text class="step-text">OCR识别</text>
					</view>
					<view class="progress-line" :class="stepLineClass(2)"></view>
					<view class="progress-step" :class="stepClass(3)">
						<view class="step-dot"></view>
						<text class="step-text">提交验证</text>
					</view>
				</view>
				<text class="progress-tip">{{ progressTip }}</text>
			</view>

			<!-- 手动输入（OCR失败或首次需要） -->
			<view v-if="!isLoggedIn && showManualInput" class="login-card">
				<view class="form-title">自动识别失败，请手动输入</view>

				<view class="captcha-area">
					<view class="captcha-left">
						<view class="captcha-label">验证码</view>
						<input
							class="captcha-input"
							v-model="captcha"
							type="text"
							maxlength="4"
							placeholder="请输入4位验证码"
							placeholder-class="input-placeholder"
						/>
					</view>
					<view class="captcha-right" @tap="reloadCaptcha">
						<image
							v-if="captchaImageUrl"
							class="captcha-image"
							:src="captchaImageUrl"
							mode="aspectFill"
						></image>
						<view v-else class="captcha-placeholder">
							<view class="captcha-loading"></view>
							<text class="captcha-loading-text">加载中</text>
						</view>
						<l-icon name="refresh" style="font-size: 18px; color: #94a3b8; margin-top: 8rpx;"></l-icon>
					</view>
				</view>

				<view class="btn-area">
					<button
						class="btn btn-primary"
						:disabled="submitting || !captcha || captcha.length !== 4"
						@tap="submitLogin"
					>
						<view v-if="submitting" class="btn-spinner"></view>
						<text v-else>确认授权</text>
					</button>
					<button class="btn btn-secondary" @tap="startAutoLogin">
						<l-icon name="robot" style="font-size: 16px; margin-right: 8rpx;"></l-icon>
						<text>重新自动识别</text>
					</button>
				</view>
			</view>

			<!-- 已登录操作 -->
			<view v-if="isLoggedIn" class="action-card">
				<button class="btn btn-danger" @tap="logout">
					<l-icon name="logout" style="font-size: 18px; margin-right: 8rpx;"></l-icon>
					<text>解除授权</text>
				</button>
			</view>

			<!-- 错误提示 -->
			<view v-if="errorMsg" class="error-tip">
				<l-icon name="alert-circle" style="font-size: 16px; color: #ef4444; margin-right: 8rpx;"></l-icon>
				<text class="error-tip__text">{{ errorMsg }}</text>
			</view>

			<!-- 功能列表 -->
			<view class="feature-card">
				<view class="feature-title">PC端专属功能</view>
				<view class="feature-list">
					<view class="feature-item">
						<view class="feature-icon feature-icon--teachers">
							<l-icon name="account-multiple" style="font-size: 20px; color: #fff;"></l-icon>
						</view>
						<view class="feature-info">
							<text class="feature-name">我的老师</text>
							<text class="feature-desc">查看各学期任课老师及联系方式</text>
						</view>
					</view>
					<view class="feature-item">
						<view class="feature-icon feature-icon--archive">
							<l-icon name="file-document-edit" style="font-size: 20px; color: #fff;"></l-icon>
						</view>
						<view class="feature-info">
							<text class="feature-name">完善档案</text>
							<text class="feature-desc">编辑联系方式、家庭成员、学校经历</text>
						</view>
					</view>
					<view class="feature-item">
						<view class="feature-icon feature-icon--stuinfo">
							<l-icon name="badge-account" style="font-size: 20px; color: #fff;"></l-icon>
						</view>
						<view class="feature-info">
							<text class="feature-name">学籍信息</text>
							<text class="feature-desc">查看学院、辅导员等学籍信息</text>
						</view>
					</view>
				</view>
			</view>

			<view class="bottom-safe-area"></view>
		</view>
	</scroll-view>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import {
	pcAutoLogin,
	pcLoginSubmit,
	pcGetSessionStatus,
	pcLogout,
	pcGetSessionCredentials,
	pcSetSessionCredentials
} from '../../pages/api/discover.js'

const STATUS_KEY = 'pc_session'

const statusBarHeight = ref(20)
const isLoggedIn = ref(false)
const sessionInfo = ref(null)
const captcha = ref('')
const captchaImageUrl = ref('')
const submitting = ref(false)
const showManualInput = ref(false)
const errorMsg = ref('')
const autoLogging = ref(false)
const currentStep = ref(0)
const progressTip = ref('正在初始化...')
const redirectPath = ref('')

let autoTimer = null

function goBack() {
	uni.navigateBack()
}

function formatTTL(seconds) {
	if (!seconds || seconds <= 0) return '已过期'
	const mins = Math.floor(seconds / 60)
	const secs = seconds % 60
	if (mins > 0) return `${mins}分${secs}秒`
	return `${secs}秒`
}

function stepClass(step) {
	if (currentStep.value > step) return 'step--done'
	if (currentStep.value === step) return 'step--active'
	return ''
}

function stepLineClass(afterStep) {
	if (currentStep.value > afterStep) return 'line--done'
	return ''
}

async function checkStoredSession() {
	try {
		const res = await pcGetSessionCredentials()
		if (res && res.success && res.result && res.result.stored) {
			const cred = res.result
			if (cred.sessionId && cred.ttlSeconds > 0) {
				// 尝试恢复会话到后端
				const setRes = await pcSetSessionCredentials({
					sessionId: cred.sessionId,
					expireAt: cred.expireAt
				})
				if (setRes && setRes.success && setRes.result && setRes.result.valid) {
					isLoggedIn.value = true
					sessionInfo.value = {
						loginAt: formatTime(cred.loginAt),
						expiresAt: formatTime(cred.expireAt),
						ttlSeconds: cred.ttlSeconds
					}
					return
				}
			}
		}
	} catch (e) {
		console.log('恢复存储会话失败', e)
	}
	isLoggedIn.value = false
}

async function checkStatus() {
	try {
		const res = await pcGetSessionStatus()
		if (res && res.success && res.result) {
			const result = res.result
			isLoggedIn.value = result.loggedIn === true
			if (result.loggedIn) {
				sessionInfo.value = {
					loginAt: result.loginAt || '',
					expiresAt: result.expiresAt || '',
					ttlSeconds: result.ttlSeconds || 0
				}
				// 同步到 Storage
				if (result.sessionId) {
					await saveToStorage(result)
				}
			}
		}
	} catch (e) {
		console.error('检查会话状态失败', e)
	}
}

async function saveToStorage(result) {
	try {
		const cred = {
			sessionId: result.sessionId,
			loginAt: result.loginAt,
			expireAt: result.expiresAt,
			ttlSeconds: result.ttlSeconds,
			updatedAt: Date.now()
		}
		uni.setStorageSync(STATUS_KEY, cred)
	} catch (e) {
		console.error('保存会话到Storage失败', e)
	}
}

function formatTime(isoStr) {
	if (!isoStr) return '--'
	try {
		const d = new Date(isoStr.replace(' ', 'T'))
		const pad = n => String(n).padStart(2, '0')
		return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
	} catch {
		return isoStr
	}
}

async function startAutoLogin() {
	if (autoLogging.value) return
	autoLogging.value = true
	showManualInput.value = false
	errorMsg.value = ''
	captcha.value = ''
	captchaImageUrl.value = ''
	currentStep.value = 0
	progressTip.value = '正在连接学校服务器...'

	try {
		// Step 1: 获取会话
		currentStep.value = 1
		progressTip.value = '正在获取会话...'
		await delay(300)

		// Step 2: OCR识别 + 提交
		currentStep.value = 2
		progressTip.value = '正在OCR识别验证码...'
		const res = await pcAutoLogin()

		if (!res || !res.success) {
			throw new Error(res?.message || '授权失败')
		}

		const result = res.result

		if (result.success) {
			// 自动登录成功
			currentStep.value = 3
			progressTip.value = '授权成功！'
			await delay(300)

			isLoggedIn.value = true
			sessionInfo.value = {
				loginAt: formatTime(result.expireTime),
				expiresAt: formatTime(result.expireTime),
				ttlSeconds: result.ttlSeconds || 1800
			}

			await saveToStorage({
				sessionId: result.sessionId,
				loginAt: new Date().toISOString(),
				expireAt: result.expireTime,
				ttlSeconds: result.ttlSeconds || 1800
			})

			uni.showToast({ title: '授权成功', icon: 'success' })
			if (redirectPath.value) {
				setTimeout(() => uni.redirectTo({ url: redirectPath.value }), 1000)
			}
		} else if (result.needManual) {
			// 需要手动输入
			currentStep.value = 0
			autoLogging.value = false
			showManualInput.value = true
			captchaImageUrl.value = result.captcha || ''
			if (result.sessionId) {
				const sess = uni.getStorageSync(STATUS_KEY) || {}
				await saveToStorage({ ...sess, sessionId: result.sessionId })
			}
			if (result.message && result.message !== 'OCR识别失败，请重试') {
				uni.showToast({ title: '请手动输入', icon: 'none', duration: 2000 })
			}
		} else {
			throw new Error(result.message || '未知错误')
		}
	} catch (e) {
		currentStep.value = 0
		autoLogging.value = false
		errorMsg.value = e?.message || e?.errMsg || '授权失败，请重试'
		showManualInput.value = true
		// 获取新的验证码图片
		reloadCaptcha()
	}
}

async function reloadCaptcha() {
	showManualInput.value = true
	captchaImageUrl.value = ''
	// 重新触发自动流程以获取新的会话和验证码
	await startAutoLogin()
}

async function submitLogin() {
	if (submitting.value || captcha.value.length !== 4) return
	submitting.value = true
	errorMsg.value = ''
	try {
		const res = await pcLoginSubmit(captcha.value)
		if (res && res.success) {
			isLoggedIn.value = true
			captcha.value = ''
			await checkStatus()
			uni.showToast({ title: '授权成功', icon: 'success' })
			if (redirectPath.value) {
				setTimeout(() => uni.redirectTo({ url: redirectPath.value }), 1000)
			}
		} else {
			errorMsg.value = res?.message || '授权失败，请重试'
			await reloadCaptcha()
		}
	} catch (e) {
		errorMsg.value = e?.message || e?.errMsg || '授权失败'
		await reloadCaptcha()
	} finally {
		submitting.value = false
	}
}

async function logout() {
	uni.showModal({
		title: '解除授权',
		content: '确定要解除PC端授权吗？',
		confirmColor: '#ef4444',
		success: async (res) => {
			if (res.confirm) {
				try {
					await pcLogout()
					isLoggedIn.value = false
					sessionInfo.value = null
					captchaImageUrl.value = ''
					captcha.value = ''
					uni.removeStorageSync(STATUS_KEY)
					uni.showToast({ title: '已解除授权', icon: 'success' })
				} catch (e) {
					uni.showToast({ title: '解除失败', icon: 'none' })
				}
			}
		}
	})
}

function delay(ms) {
	return new Promise(resolve => setTimeout(resolve, ms))
}

onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 20
})

onUnmounted(() => {
	if (autoTimer) clearTimeout(autoTimer)
})

onLoad((options) => {
	if (options.redirect) {
		redirectPath.value = options.redirect
	}
	// 优先从 Storage 恢复会话
	checkStoredSession()
	// 如果 Storage 恢复成功不再自动登录，失败则尝试完整登录
	setTimeout(async () => {
		if (!isLoggedIn.value && !showManualInput.value) {
			await startAutoLogin()
		}
	}, 500)
})
</script>

<style lang="scss" scoped>
@import url('../../static/css/home-page.css');

.pc-login-page {
	width: 100%;
	height: 100vh;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
}

.nav-bar {
	background: linear-gradient(135deg, #6366f1, #818cf8);
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

.page-content {
	padding: 24rpx;
}

.bottom-safe-area {
	height: 48rpx;
}

/* 状态卡片 */
.status-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 48rpx 32rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	margin-bottom: 24rpx;
	box-shadow: 0 2rpx 12rpx rgba(99, 102, 241, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.status-icon {
	width: 96rpx;
	height: 96rpx;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 24rpx;
}

.status-icon--success {
	background: rgba(16, 185, 129, 0.1);
	color: #10b981;
}

.status-icon--warn {
	background: rgba(245, 158, 11, 0.1);
	color: #f59e0b;
}

.status-title {
	font-size: 36rpx;
	font-weight: 800;
	color: var(--text-primary);
	margin-bottom: 8rpx;
}

.status-desc {
	font-size: 26rpx;
	color: var(--text-secondary);
	text-align: center;
}

.status-detail {
	width: 100%;
	margin-top: 24rpx;
	padding-top: 24rpx;
	border-top: 1px solid rgba(226, 232, 240, 0.7);
}

.detail-row {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 12rpx 0;
}

.detail-label {
	font-size: 26rpx;
	color: var(--text-secondary);
}

.detail-value {
	font-size: 26rpx;
	color: var(--text-primary);
	font-weight: 600;
}

.detail-value--highlight {
	color: #6366f1;
}

/* 自动进度卡片 */
.auto-progress-card {
	background: rgba(99, 102, 241, 0.04);
	border: 1px solid rgba(99, 102, 241, 0.15);
	border-radius: 20rpx;
	padding: 32rpx 24rpx;
	margin-bottom: 24rpx;
}

.progress-header {
	display: flex;
	align-items: center;
	margin-bottom: 24rpx;
}

.progress-title {
	font-size: 28rpx;
	font-weight: 700;
	color: #6366f1;
	flex: 1;
}

.progress-spinner {
	width: 32rpx;
	height: 32rpx;
	border: 4rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: #6366f1;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

.progress-steps {
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 20rpx;
}

.progress-step {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 8rpx;
}

.step-dot {
	width: 32rpx;
	height: 32rpx;
	border-radius: 50%;
	background: rgba(148, 163, 184, 0.3);
	border: 3rpx solid rgba(148, 163, 184, 0.3);
	transition: all 0.3s;
}

.step--active .step-dot {
	background: rgba(99, 102, 241, 0.15);
	border-color: #6366f1;
}

.step--done .step-dot {
	background: #10b981;
	border-color: #10b981;
}

.step-text {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.step--active .step-text {
	color: #6366f1;
	font-weight: 600;
}

.step--done .step-text {
	color: #10b981;
}

.progress-line {
	width: 80rpx;
	height: 3rpx;
	background: rgba(148, 163, 184, 0.3);
	margin: 0 8rpx;
	margin-bottom: 20rpx;
	transition: background 0.3s;
}

.line--done {
	background: #10b981;
}

.progress-tip {
	display: block;
	text-align: center;
	font-size: 24rpx;
	color: var(--text-secondary);
}

/* 授权表单 */
.login-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 32rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 2rpx 12rpx rgba(99, 102, 241, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.form-title {
	font-size: 28rpx;
	font-weight: 700;
	color: var(--text-primary);
	margin-bottom: 24rpx;
}

.captcha-area {
	display: flex;
	gap: 20rpx;
	align-items: flex-end;
	margin-bottom: 24rpx;
}

.captcha-left {
	flex: 1;
}

.captcha-label {
	font-size: 26rpx;
	color: var(--text-secondary);
	margin-bottom: 12rpx;
}

.captcha-input {
	height: 88rpx;
	background: var(--bg-secondary);
	border-radius: 16rpx;
	padding: 0 24rpx;
	font-size: 32rpx;
	font-weight: 600;
	letter-spacing: 4px;
	border: 1px solid rgba(148, 163, 184, 0.2);
}

.input-placeholder {
	color: #94a3b8;
	font-size: 28rpx;
	font-weight: 400;
	letter-spacing: 0;
}

.captcha-right {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 8rpx;
}

.captcha-image {
	width: 160rpx;
	height: 72rpx;
	border-radius: 12rpx;
	border: 1px solid rgba(148, 163, 184, 0.2);
	background: #fff;
}

.captcha-placeholder {
	width: 160rpx;
	height: 72rpx;
	border-radius: 12rpx;
	background: rgba(148, 163, 184, 0.1);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 8rpx;
}

.captcha-loading {
	width: 32rpx;
	height: 32rpx;
	border: 3rpx solid rgba(148, 163, 184, 0.2);
	border-top-color: #6366f1;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

.captcha-loading-text {
	font-size: 20rpx;
	color: #94a3b8;
}

/* 按钮 */
.btn-area {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
}

.btn {
	height: 88rpx;
	border-radius: 44rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 30rpx;
	font-weight: 700;
	border: none;
	margin: 0;
}

.btn-primary {
	background: linear-gradient(135deg, #6366f1, #818cf8);
	color: #fff;
	box-shadow: 0 6rpx 20rpx rgba(99, 102, 241, 0.3);
}

.btn-primary[disabled] {
	background: rgba(99, 102, 241, 0.4);
	box-shadow: none;
}

.btn-secondary {
	background: rgba(99, 102, 241, 0.08);
	color: #6366f1;
}

.btn-spinner {
	width: 36rpx;
	height: 36rpx;
	border: 4rpx solid rgba(255, 255, 255, 0.3);
	border-top-color: #fff;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

.btn-danger {
	background: rgba(239, 68, 68, 0.08);
	color: #ef4444;
	border: 1px solid rgba(239, 68, 68, 0.2);
}

/* 操作卡片 */
.action-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 24rpx;
	margin-bottom: 24rpx;
}

/* 错误提示 */
.error-tip {
	display: flex;
	align-items: center;
	background: rgba(239, 68, 68, 0.06);
	border: 1px solid rgba(239, 68, 68, 0.2);
	border-radius: 12rpx;
	padding: 16rpx 20rpx;
	margin-bottom: 24rpx;
}

.error-tip__text {
	font-size: 26rpx;
	color: #ef4444;
}

/* 功能卡片 */
.feature-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 24rpx;
}

.feature-title {
	font-size: 28rpx;
	font-weight: 700;
	color: var(--text-primary);
	margin-bottom: 20rpx;
}

.feature-list {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
}

.feature-item {
	display: flex;
	align-items: center;
	gap: 16rpx;
	padding: 16rpx;
	border-radius: 16rpx;
	background: var(--bg-secondary);
}

.feature-icon {
	width: 72rpx;
	height: 72rpx;
	border-radius: 18rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
}

.feature-icon--teachers {
	background: linear-gradient(135deg, #f59e0b, #fbbf24);
}

.feature-icon--archive {
	background: linear-gradient(135deg, #8b5cf6, #a78bfa);
}

.feature-icon--stuinfo {
	background: linear-gradient(135deg, #14b8a6, #2dd4bf);
}

.feature-info {
	flex: 1;
}

.feature-name {
	display: block;
	font-size: 28rpx;
	font-weight: 700;
	color: var(--text-primary);
	margin-bottom: 4rpx;
}

.feature-desc {
	display: block;
	font-size: 24rpx;
	color: var(--text-secondary);
}

@keyframes spin {
	to { transform: rotate(360deg); }
}
</style>
