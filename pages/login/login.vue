<!-- webpackChunkName: "login" -->
<template>
	<view class="login-page">

		<!-- 背景装饰 -->
		<view class="login-bg">
			<view class="login-bg__orb login-bg__orb--1"></view>
			<view class="login-bg__orb login-bg__orb--2"></view>
			<view class="login-bg__orb login-bg__orb--3"></view>
			<view class="login-bg__grid"></view>
		</view>
		<view class="hero">
			<image src="../../static/images/keke.png" mode="widthFix"></image>
			<view>让教务管理更简单、更高效</view>
		</view>

		<view class="login-page__content" :style="{paddingTop: navPaddingTop}">
			<!-- Hero Brand -->
			<view class="login-page__hero">
				<view class="login-page__logo-wrap">
					<view class="login-page__logo-glow"></view>
					<view class="login-page__logo">
						<!-- <l-icon name="education" size="36" color="#fff"></l-icon> -->
						<image src="../../static/images/keke.png" mode="widthFix"></image>

					</view>
				</view>
				<view class="login-page__brand">
					<text class="login-page__brand-name">众柴智慧校园</text>
					<text class="login-page__brand-slogan">让教务管理更简单、更高效</text>
				</view>
			</view>

			<!-- Tabs -->
			<view class="login-page__tabs">
				<view
					class="login-page__tab"
					:class="{'login-page__tab--active': loginType === 'school'}"
					@tap="switchLoginType('school')"
				>
					<l-icon name="education" size="14" :color="loginType === 'school' ? '#fff' : 'var(--text-tertiary)'"></l-icon>
					<text>学校账号</text>
				</view>
				<!-- #ifdef MP-WEIXIN -->
				<view
					class="login-page__tab"
					:class="{'login-page__tab--active': loginType === 'wechat'}"
					@tap="switchLoginType('wechat')"
				>
					<l-icon name="logo-wechat-stroke" size="14" :color="loginType === 'wechat' ? '#fff' : 'var(--text-tertiary)'"></l-icon>
					<text>微信登录</text>
				</view>
				<!-- #endif -->
			</view>

			<!-- School Login Card -->
			<view class="login-card login-card--school" v-if="loginType === 'school'">
				<view class="login-card__header">
					<text class="login-card__title">欢迎回来</text>
					<text class="login-card__subtitle">请使用教务系统账号登录</text>
				</view>

				<view class="login-card__form">
					<!-- Username -->
					<view class="login-field" :class="{'login-field--focus': usernameFocused}">
						<view class="login-field__icon">
							<l-icon name="user" size="18" :color="usernameFocused ? 'var(--primary-500)' : 'var(--text-tertiary)'"></l-icon>
						</view>
						<input
							class="login-field__input"
							type="text"
							v-model="username"
							placeholder="请输入学号"
							placeholder-class="login-field__placeholder"
							@focus="usernameFocused = true"
							@blur="usernameFocused = false"
						/>
						<view v-if="username" class="login-field__clear" @tap="username = ''">
							<l-icon name="close-circle-filled" size="14" color="var(--text-tertiary)"></l-icon>
						</view>
					</view>

					<!-- Password -->
					<view class="login-field" :class="{'login-field--focus': passwordFocused}">
						<view class="login-field__icon">
							<l-icon name="lock-on" size="18" :color="passwordFocused ? 'var(--primary-500)' : 'var(--text-tertiary)'"></l-icon>
						</view>
						<input
							class="login-field__input"
							:type="showPassword ? 'text' : 'password'"
							v-model="password"
							placeholder="请输入密码"
							placeholder-class="login-field__placeholder"
							@focus="passwordFocused = true"
							@blur="passwordFocused = false"
						/>
						<view class="login-field__action" @tap="togglePasswordVisibility">
							<l-icon
								:name="showPassword ? 'browse-filled' : 'browse-off'"
								size="18"
								color="var(--text-tertiary)"
							></l-icon>
						</view>
					</view>

					<!-- Options Row -->
		<view class="login-options">
			<view class="login-check" @tap="toggleRememberPassword">
				<view class="login-check__box" :class="{'login-check__box--checked': rememberPassword}">
					<l-icon v-if="rememberPassword" name="check" size="10" color="#fff"></l-icon>
				</view>
				<text class="login-check__label">记住密码</text>
			</view>
		</view>

					<!-- Submit Button -->
					<button
						class="login-btn login-btn--primary"
						:class="{'login-btn--loading': loading}"
						@tap="handleSchoolLogin"
						:disabled="loading"
					>
						<template v-if="!loading">
							<text>立即登录</text>
						</template>
						<template v-else>
							<view class="login-btn__spinner"></view>
							<text>登录中...</text>
						</template>
					</button>
				</view>
			</view>

			<!-- WeChat Login Card -->
			<!-- #ifdef MP-WEIXIN -->
			<view class="login-card login-card--wechat" v-else-if="loginType === 'wechat'">
				<view class="login-wechat__icon-wrap">
					<view class="login-wechat__icon-ring"></view>
					<view class="login-wechat__icon">
						<l-icon name="logo-wechat-fill" size="52" color="#07c160"></l-icon>
					</view>
				</view>
				<text class="login-wechat__title">微信一键登录</text>
				<text class="login-wechat__subtitle">更安全、更便捷的登录方式</text>

				<button
					class="login-btn login-btn--wechat"
					:class="{'login-btn--loading': wxLoading}"
					@tap="handleWechatLogin"
					:disabled="wxLoading"
				>
					<template v-if="!wxLoading">
						<l-icon name="logo-wechat-fill" size="16" color="#fff"></l-icon>
						<text>微信授权登录</text>
					</template>
					<template v-else>
						<view class="login-btn__spinner login-btn__spinner--light"></view>
						<text>授权中...</text>
					</template>
				</button>

				<view class="login-wechat__tip">
					<l-icon name="shield-check" size="13" color="var(--text-tertiary)"></l-icon>
					<text>首次登录将自动创建账号</text>
				</view>
			</view>
			<!-- #endif -->

			<!-- Security Badge -->
			<view class="login-security">
				<view class="login-security__icon">
					<l-icon name="lock-on" size="12" color="var(--primary-500)"></l-icon>
				</view>
				<text class="login-security__text">账号信息通过加密通道传输，不保存明文密码</text>
			</view>

			<view class="login-page__safe-area"></view>
		</view>
	</view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { showToast, relaunch } from '../../pages/api/page.js'
import { login, initLogin } from '../api/login.js'
import { getUserArchive } from '../api/discover.js'
import { request } from '../../utils/request.js'
import { useUserStore } from '@/store/user.js'
import { useSchoolAccountStore } from '@/store/schoolAccount.js'

const userStore = useUserStore()
const schoolAccountStore = useSchoolAccountStore()

const loginType = ref('school')
const username = ref('')
const password = ref('')
const loading = ref(false)
const wxLoading = ref(false)
const rememberPassword = ref(true)
const showPassword = ref(false)
const navPaddingTop = ref('0px')
const usernameFocused = ref(false)
const passwordFocused = ref(false)

onLoad((options) => {
	const fromBind = options && options.fromBind === 'true'
	const token = uni.getStorageSync('token')
	if (token && !fromBind) {
		const pages = getCurrentPages()
		const isFromIndex = pages.length > 1 && pages[pages.length - 2]?.route?.includes('index')
		if (isFromIndex) return
		relaunch({ url: '/pages/index/index' })
		return
	}
	clearLoginInfo(fromBind)
	checkSavedCredentials()
	fetchLoginClientId()
})

onShow(() => {
	const token = uni.getStorageSync('token')
	if (token) {
		const pages = getCurrentPages()
		const isFromIndex = pages.length > 1 && pages[pages.length - 2]?.route?.includes('index')
		if (isFromIndex) return
		relaunch({ url: '/pages/index/index' })
	}
})

async function fetchLoginClientId() {
	if (uni.getStorageSync('clientId')) return
	try {
		await initLogin()
	} catch (err) {
		console.warn('clientId init failed', err)
	}
}

async function handleWechatLogin() {
	uni.vibrateShort()
	wxLoading.value = true
	uni.login({
		provider: 'weixin',
		success: async (loginRes) => {
			if (!loginRes.code) { wxLoading.value = false; return }
			try {
				const res = await request({ url: '/api/user/wx/login', method: 'POST', data: { code: loginRes.code } })
				if (res.success) {
					const { token, userInfo } = res.result || {}
					if (token) uni.setStorageSync('token', token)
					uni.setStorageSync('loginType', 'wechat')
					if (userInfo) {
						uni.setStorageSync('userInfo', JSON.stringify(userInfo))
						uni.setStorageSync('hasBindSchoolAccount', userInfo.hasSchoolAccount || false)
					}
					await showToast({ title: '登录成功', icon: 'success' })
					const currentPage = getCurrentPages().at(-1)
					const fromBind = currentPage?.options?.fromBind === 'true'
					if (fromBind || !userInfo?.hasSchoolAccount) {
						uni.navigateTo({ url: '/pages/user/bind' })
					} else {
						await relaunch({ url: '/pages/index/index' })
					}
				} else {
					throw new Error(res.message || '微信登录失败')
				}
			} catch (error) {
				await showToast({ title: '微信登录失败', icon: 'none' })
			} finally {
				wxLoading.value = false
			}
		},
		fail: () => {
			wxLoading.value = false
			showToast({ title: '微信登录失败', icon: 'none' })
		}
	})
}

function clearLoginInfo(preserveUserInfo = false) {
	try {
		if (!preserveUserInfo) {
			uni.removeStorageSync('token')
			uni.removeStorageSync('userInfo')
			uni.removeStorageSync('loginType')
			uni.removeStorageSync('currentWeek')
			uni.removeStorageSync('todayClasses')
			uni.removeStorageSync('weekClasses')
			uni.removeStorageSync('termInfo')
		}
	} catch {}
}

function checkSavedCredentials() {
	try {
		const savedUsername = uni.getStorageSync('saved_username')
		const savedPassword = uni.getStorageSync('saved_password')
		const remember = uni.getStorageSync('remember_password')
		if (remember && savedUsername && savedPassword) {
			username.value = savedUsername
			password.value = savedPassword
			rememberPassword.value = true
		} else {
			rememberPassword.value = true
		}
	} catch {
		rememberPassword.value = true
	}
}

function togglePasswordVisibility() {
	uni.vibrateShort()
	showPassword.value = !showPassword.value
}

function toggleRememberPassword() {
	uni.vibrateShort()
	rememberPassword.value = !rememberPassword.value
}

function switchLoginType(type) {
	uni.vibrateShort()
	loginType.value = type
}

function handleForgotPassword() {
	uni.vibrateShort()
	uni.showModal({
		title: '找回密码',
		content: '请前往学校教务系统官网（jwgl.xxx.edu.cn）通过"忘记密码"功能重置密码后，再回到本应用登录。',
		showCancel: false,
		confirmText: '我知道了',
		confirmColor: '#2563eb'
	})
}

async function handleSchoolLogin() {
	uni.vibrateShort()
	if (loading.value) return
	if (!username.value?.trim()) { await showToast({ title: '请输入学号', icon: 'none' }); return }
	if (!password.value?.trim()) { await showToast({ title: '请输入密码', icon: 'none' }); return }

	loading.value = true
	try {
		const res = await login({ username: username.value.trim(), password: password.value.trim() })
		if (res?.success) {
			if (rememberPassword.value) {
				uni.setStorageSync('saved_username', username.value)
				uni.setStorageSync('saved_password', password.value)
				uni.setStorageSync('remember_password', true)
			} else {
				uni.removeStorageSync('saved_username')
				uni.removeStorageSync('saved_password')
				uni.removeStorageSync('remember_password')
			}
			const token = res.result?.token
			if (!token) throw new Error('no token')
			uni.setStorageSync('token', token)
			uni.setStorageSync('loginType', 'school')
			if (res.result?.userInfo) uni.setStorageSync('userInfo', JSON.stringify(res.result.userInfo))
			schoolAccountStore.setSchoolLogin(true)
			schoolAccountStore.updateBinding(true)
			await showToast({ title: '登录成功', icon: 'success' })

			// 登录成功后立即获取完整档案信息，确保学院等信息显示正确
			try {
				const archiveRes = await getUserArchive()
				if (archiveRes && archiveRes.success && archiveRes.result) {
					const archive = archiveRes.result
					const stored = uni.getStorageSync('userInfo')
					const base = stored ? JSON.parse(stored) : {}
					const merged = {
						...base,
						...archive,
						// 档案中的字段优先级更高
						college: archive.facultyName || archive.college || base.college || '',
						className: archive.adminClass || archive.className || base.className || ''
					}
					uni.setStorageSync('userInfo', JSON.stringify(merged))
				}
			} catch (e) {
				console.warn('获取档案信息失败，不影响登录流程', e)
			}

			await relaunch({ url: '/pages/index/index' })
		} else {
			await showToast({ title: res.message || '登录失败', icon: 'none' })
		}
	} catch (err) {
		console.error('login error:', err)
		await showToast({ title: '网络异常，请稍后重试', icon: 'none' })
	} finally {
		loading.value = false
	}
}
</script>

<style lang="scss" scoped>
/* ── Background ── */
.login-page {
	flex: 1;
	width: 100%;
	min-height: 100vh;
	display: flex;
	flex-direction: column;
	background: var(--bg-secondary);
	position: relative;
	overflow: hidden;
	

}
.hero{
	background-size: 100% 100%;
	width: 100%;
	color: #2563eb;
	font-size: 32rpx;
	font-weight: 600;
	text-align: center;
	padding: 20rpx 0;
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	justify-content: flex-start;
	padding-left: 40rpx;
	gap: 20rpx;
	image{
		width: 100rpx;
		height: 100rpx;
		border-radius: 50%;
		overflow: hidden;
		// box-shadow: 0 0 10rpx 0 rgba(0, 0, 0, 0.1);
		transition: all 0.3s ease;
		&:hover {
			transform: scale(1.05);
		}
	}
}
.login-bg {
	position: absolute;
	inset: 0;
	pointer-events: none;
	z-index: 0;
	background: url('../../static/images/hero_bg.png') no-repeat,radial-gradient(circle, rgba(99, 102, 241, 0.25) 0%, transparent 70%)  ;
	background-size: contain;
	width: 100%;
	height: 100%;
	position: absolute;
	top: -160rpx;
	left: 0;
	z-index: 0;
}

.login-bg__orb {
	position: absolute;
	border-radius: 50%;
	filter: blur(80rpx);
	opacity: 0.5;

	&--1 {
		width: 400rpx;
		height: 400rpx;
		background: radial-gradient(circle, rgba(99, 102, 241, 0.25) 0%, transparent 70%);
		top: -100rpx;
		right: -80rpx;
		animation: orbFloat 8s ease-in-out infinite;
	}

	&--2 {
		width: 300rpx;
		height: 300rpx;
		background: radial-gradient(circle, rgba(139, 92, 246, 0.2) 0%, transparent 70%);
		bottom: 200rpx;
		left: -60rpx;
		animation: orbFloat 10s ease-in-out infinite reverse;
	}

	&--3 {
		width: 200rpx;
		height: 200rpx;
		background: radial-gradient(circle, rgba(16, 185, 129, 0.15) 0%, transparent 70%);
		bottom: 80rpx;
		right: 40rpx;
		animation: orbFloat 7s ease-in-out infinite 2s;
	}
}

@keyframes orbFloat {
	0%, 100% { transform: translateY(0) scale(1); }
	50% { transform: translateY(-20rpx) scale(1.05); }
}

.login-bg__grid {
	position: absolute;
	inset: 0;
	background-image: radial-gradient(circle, rgba(99, 102, 241, 0.06) 1px, transparent 1px);
	background-size: 32rpx 32rpx;
}

/* ── Content ── */
.login-page__content {
	position: relative;
	z-index: 1;
	padding: 0 40rpx;
	padding-bottom: calc(env(safe-area-inset-bottom) + 32rpx);
}

/* ── Hero ── */
.login-page__hero {
	display: flex;
	align-items: center;
	gap: 28rpx;
	padding: 56rpx 0 48rpx;
}

.login-page__logo-wrap {
	position: relative;
	flex-shrink: 0;
}

.login-page__logo-glow {
	position: absolute;
	inset: -16rpx;
	background: radial-gradient(circle, rgba(99, 102, 241, 0.4) 0%, transparent 70%);
	border-radius: 32rpx;
	animation: logoGlow 3s ease-in-out infinite;
}

@keyframes logoGlow {
	0%, 100% { opacity: 0.6; transform: scale(1); }
	50% { opacity: 1; transform: scale(1.08); }
}

.login-page__logo {
	top: -100rpx;
	position: relative;
	width: 100rpx;
	height: 100rpx;
	border-radius: 28rpx;
	background: linear-gradient(145deg, var(--primary-500), var(--primary-700));
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 12rpx 40rpx rgba(37, 99, 235, 0.35);
	position: relative;
	z-index: 1;
}

.login-page__brand {
	top: -100rpx;
	position: relative;
	display: flex;
	flex-direction: column;
	gap: 6rpx;
}

.login-page__brand-name {
	font-size: 40rpx;
	font-weight: 800;
	color: var(--text-primary);
	letter-spacing: -0.5px;
	line-height: 1.2;
}

.login-page__brand-slogan {
	font-size: 24rpx;
	color: var(--text-tertiary);
	letter-spacing: 0.3px;
}

/* ── Tabs ── */
.login-page__tabs {
	display: flex;
	gap: 12rpx;
	margin-bottom: 28rpx;
	background: rgba(255, 255, 255, 0.7);
	backdrop-filter: blur(20px) saturate(180%);
	-webkit-backdrop-filter: blur(20px) saturate(180%);
	border-radius: 20rpx;
	padding: 6rpx;
	border: 1px solid rgba(255, 255, 255, 0.6);
	box-shadow: 0 4rpx 16rpx rgba(99, 102, 241, 0.06);
}

.login-page__tab {
	top: -100rpx;
	position: relative;
	flex: 1;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 8rpx;
	padding: 18rpx 0;
	border-radius: 16rpx;
	font-size: 28rpx;
	color: var(--text-tertiary);
	font-weight: 600;
	transition: all 0.2s var(--ease-spring);

	&:active { transform: scale(0.97); }

	&--active {
		background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
		color: #fff;
		box-shadow: 0 6rpx 20rpx rgba(37, 99, 235, 0.35);
	}
}

/* ── Cards ── */
.login-card {
	top: -100rpx;
	position: relative;
	background: rgba(255, 255, 255, 0.82);
	backdrop-filter: blur(24px) saturate(180%);
	-webkit-backdrop-filter: blur(24px) saturate(180%);
	border-radius: 28rpx;
	padding: 36rpx;
	border: 1px solid rgba(255, 255, 255, 0.7);
	box-shadow:
		0 8rpx 32rpx rgba(99, 102, 241, 0.08),
		0 2rpx 8rpx rgba(0, 0, 0, 0.04),
		inset 0 1px 0 rgba(255, 255, 255, 0.9);
	animation: cardIn 0.4s ease-out both;
}

@keyframes cardIn {
	from { opacity: 0; transform: translateY(12rpx); }
	to { opacity: 1; transform: translateY(0); }
}

.login-card__header {
	margin-bottom: 36rpx;
}

.login-card__title {
	display: block;
	font-size: 36rpx;
	font-weight: 800;
	color: var(--text-primary);
	letter-spacing: -0.3px;
}

.login-card__subtitle {
	display: block;
	font-size: 24rpx;
	color: var(--text-tertiary);
	margin-top: 6rpx;
}

/* ── Fields ── */
.login-card__form {
	display: flex;
	flex-direction: column;
	gap: 0;
}

.login-field {
	display: flex;
	align-items: center;
	height: 100rpx;
	padding: 0 24rpx;
	border-radius: 20rpx;
	background: rgba(255, 255, 255, 0.6);
	border: 2rpx solid rgba(148, 163, 184, 0.15);
	transition: all 0.2s var(--ease-out);
	margin-bottom: 20rpx;

	&--focus {
		border-color: var(--primary-400);
		background: rgba(255, 255, 255, 0.95);
		box-shadow: 0 0 0 4rpx rgba(99, 102, 241, 0.1);
	}
}

.login-field__icon {
	width: 40rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
	transition: color 0.2s ease;
}

.login-field__input {
	flex: 1;
	height: 100%;
	font-size: 30rpx;
	color: var(--text-primary);
	background: transparent;
}

.login-field__placeholder {
	color: #cbd5e1;
	font-size: 28rpx;
}

.login-field__clear,
.login-field__action {
	width: 44rpx;
	height: 44rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
	transition: transform 0.15s ease;
	&:active { transform: scale(0.9); }
}

/* ── Options ── */
.login-options {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 32rpx;
}

.login-check {
	display: flex;
	align-items: center;
	gap: 12rpx;

	&:active { opacity: 0.7; }
}

.login-check__box {
	width: 36rpx;
	height: 36rpx;
	border-radius: 10rpx;
	border: 2rpx solid rgba(148, 163, 184, 0.3);
	display: flex;
	align-items: center;
	justify-content: center;
	transition: all 0.2s var(--ease-spring);
	background: #fff;
	flex-shrink: 0;

	&--checked {
		background: var(--primary-500);
		border-color: var(--primary-500);
		transform: scale(1.05);
	}
}

.login-check__label {
	font-size: 26rpx;
	color: var(--text-secondary);
}

.login-forgot {
	font-size: 26rpx;
	color: var(--primary-500);
	font-weight: 600;

	&:active { opacity: 0.6; }
}

/* ── Buttons ── */
.login-btn {
	width: 100%;
	height: 96rpx;
	border-radius: 24rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 12rpx;
	font-size: 32rpx;
	font-weight: 700;
	border: none;
	transition: all 0.2s var(--ease-spring);

	&:active:not([disabled]) { transform: scale(0.97); }

	&[disabled] { opacity: 0.7; transform: none; }

	&--primary {
		background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
		color: #fff;
		box-shadow: 0 8rpx 28rpx rgba(37, 99, 235, 0.35);
		position: relative;
		overflow: hidden;

		&::after {
			content: '';
			position: absolute;
			top: 0;
			left: -100%;
			width: 100%;
			height: 100%;
			background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
			animation: btnShimmer 3s ease-in-out infinite;
		}
	}

	&--wechat {
		background: linear-gradient(135deg, #06a054, #07c160);
		color: #fff;
		box-shadow: 0 8rpx 28rpx rgba(7, 193, 96, 0.35);
	}

	&__spinner {
		width: 32rpx;
		height: 32rpx;
		border: 3rpx solid rgba(255, 255, 255, 0.3);
		border-top-color: #fff;
		border-radius: 50%;
		animation: btnSpin 0.8s linear infinite;

		&--light {
			border-color: rgba(255, 255, 255, 0.3);
			border-top-color: #fff;
		}
	}
}

@keyframes btnShimmer {
	0% { left: -100%; }
	50%, 100% { left: 100%; }
}

@keyframes btnSpin { to { transform: rotate(360deg); } }

/* ── WeChat Card ── */
.login-card--wechat {
	display: flex;
	flex-direction: column;
	align-items: center;
	text-align: center;
	padding: 48rpx 36rpx 40rpx;
}

.login-wechat__icon-wrap {
	position: relative;
	width: 140rpx;
	height: 140rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 28rpx;
}

.login-wechat__icon-ring {
	position: absolute;
	inset: 0;
	border-radius: 50%;
	border: 3rpx solid rgba(7, 193, 96, 0.2);
	animation: ringPulse 2.5s ease-in-out infinite;
}

@keyframes ringPulse {
	0%, 100% { transform: scale(1); opacity: 1; }
	50% { transform: scale(1.1); opacity: 0.6; }
}

.login-wechat__icon {
	width: 120rpx;
	height: 120rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, rgba(7, 193, 96, 0.08), rgba(7, 193, 96, 0.03));
	border: 2rpx solid rgba(7, 193, 96, 0.15);
	display: flex;
	align-items: center;
	justify-content: center;
}

.login-wechat__title {
	font-size: 36rpx;
	font-weight: 800;
	color: var(--text-primary);
	letter-spacing: -0.3px;
	margin-bottom: 8rpx;
}

.login-wechat__subtitle {
	font-size: 26rpx;
	color: var(--text-tertiary);
	margin-bottom: 40rpx;
}

.login-wechat__tip {
	display: flex;
	align-items: center;
	gap: 8rpx;
	margin-top: 28rpx;
	font-size: 24rpx;
	color: var(--text-tertiary);
}

/* ── Security Badge ── */
.login-security {
	top: -100rpx;
	position: relative;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 10rpx;
	margin-top: 36rpx;
	padding: 16rpx 28rpx;
	background: rgba(255, 255, 255, 0.6);
	backdrop-filter: blur(12px);
	border-radius: 100rpx;
	border: 1rpx solid rgba(99, 102, 241, 0.12);

	.login-security__text {
		font-size: 22rpx;
		color: var(--text-tertiary);
		line-height: 1.4;
	}
}

.login-security__icon {
	width: 36rpx;
	height: 36rpx;
	border-radius: 50%;
	background: rgba(99, 102, 241, 0.1);
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
}

.login-page__safe-area {
	height: 32rpx;
}
</style>
