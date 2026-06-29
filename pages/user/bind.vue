<template>
	<view class="container">
		<!-- 自定义导航栏 -->
		<CustomNavBar
			:center-title="'绑定学校账号'"
			:show-back="true"
			:fixed="true"
			@navHeightReady="handleNavHeightReady"
		/>

		<view class="content" :style="{paddingTop: navPaddingTop}">
			<!-- 品牌区域 -->
			<view class="brand-section">
				<view class="brand-icon">
					<l-icon name="education" size="48" color="var(--primary-color)"></l-icon>
				</view>
				<text class="brand-title">学号绑定</text>
				<text class="brand-subtitle">绑定教务系统学号以同步成绩与课程</text>
			</view>

			<!-- 绑定表单 -->
			<view class="form-group">
				<!-- 账号输入 -->
				<t-input
					v-model="username"
					label="学号"
					placeholder="请输入教务系统学号"
					:clearable="true"
				/>

				<!-- 密码输入 -->
				<t-input
					v-model="password"
					label="密码"
					placeholder="请输入教务系统密码"
					type="password"
					:clearable="true"
					suffix-icon="eye-open"
				/>
			</view>

			<!-- 绑定按钮 -->
			<view class="footer-actions">
				<view
					class="bind-btn"
					:class="{ 'bind-btn--loading': loading }"
					@tap="handleBind"
				>
					<text class="bind-btn-text" v-if="!loading">立即绑定</text>
					<view v-else class="loading-spinner"></view>
				</view>

				<view class="disclaimer">
					<l-icon name="lock-on" size="14" color="var(--text-tertiary)"></l-icon>
					<text class="disclaimer-text">您的账号密码将仅用于教务系统身份验证</text>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref } from 'vue'
import { initLogin } from '../../pages/api/login.js'
import { showToast } from '../../pages/api/page.js'
import { request } from '../../utils/request.js'
import { useSchoolAccountStore } from '@/store/schoolAccount.js'
import CustomNavBar from '@/components/CustomNavBar.vue'

const schoolAccountStore = useSchoolAccountStore()

const navPaddingTop = ref('0px')
const username = ref('')
const password = ref('')
const loading = ref(false)
const clientId = ref('')

function handleNavHeightReady(navInfo) {
	navPaddingTop.value = navInfo.heightPx
}

async function initClientId() {
	try {
		const existingClientId = uni.getStorageSync('clientId')
		if (!existingClientId) {
			const result = await initLogin()
			console.log('初始化登录成功:', result)
		}
		clientId.value = uni.getStorageSync('clientId') || ''
	} catch (e) {
		console.error('初始化客户端ID失败:', e)
	}
}

async function handleBind() {
	if (!username.value) {
		showToast({ title: '请输入学号', icon: 'none' })
		return
	}
	if (!password.value) {
		showToast({ title: '请输入密码', icon: 'none' })
		return
	}
	loading.value = true
	try {
		const res = await request({
			url: '/scloud/bind',
			method: 'POST',
			data: {
				username: username.value,
				password: password.value
			},
			header: {
				'x-client-id': uni.getStorageSync('clientId') || ''
			}
		})
		if (res && res.success) {
			schoolAccountStore.updateBinding(true, res.result)
			showToast({ title: '绑定成功', icon: 'success' })
			setTimeout(() => { uni.navigateBack() }, 1500)
		} else {
			showToast({ title: res.message || '绑定失败', icon: 'none' })
		}
	} catch (e) {
		console.error('绑定失败:', e)
		showToast({ title: '网络错误，请重试', icon: 'none' })
	} finally {
		loading.value = false
	}
}

initClientId()
</script>

<style lang="scss" scoped>
.container {
	min-height: 100vh;
	background-color: var(--bg-secondary);
}

.content {
	padding: var(--spacing-md);
}

/* 品牌区域 */
.brand-section {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 40px 0 30px;
	text-align: center;
}

.brand-icon {
	width: 80px;
	height: 80px;
	background-color: var(--bg-card);
	border-radius: 24px;
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: var(--spacing-md);
	box-shadow: var(--shadow-md);
	border: 0.5px solid var(--border-primary);
}

.brand-title {
	font-size: 24px;
	font-weight: 800;
	color: var(--text-primary);
	margin-bottom: 8px;
	letter-spacing: -0.5px;
}

.brand-subtitle {
	font-size: 14px;
	color: var(--text-tertiary);
	font-weight: 500;
	padding: 0 40px;
	line-height: 1.5;
}

/* 表单区域 */
.form-group {
	background-color: var(--bg-card);
	border-radius: var(--radius-xl);
	padding: 0 var(--spacing-md);
	box-shadow: var(--shadow-sm);
	border: 0.5px solid var(--border-primary);
	margin-bottom: var(--spacing-xl);
}

/* 底部操作 */
.footer-actions {
	padding: 0 var(--spacing-sm);
}

.bind-btn {
	height: 52px;
	background-color: var(--primary-color);
	border-radius: var(--radius-lg);
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 8px 24px rgba(22, 93, 255, 0.2);
	transition: all 0.2s var(--ease-in-out);

	.bind-btn-text {
		font-size: 16px;
		font-weight: 700;
		color: #fff;
	}

	&:active {
		transform: scale(0.99);
		opacity: 0.9;
	}

	&--loading {
		opacity: 0.7;
		pointer-events: none;
		box-shadow: none;
	}
}

.loading-spinner {
	width: 20px;
	height: 20px;
	border: 2px solid rgba(255, 255, 255, 0.3);
	border-top-color: #fff;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

.disclaimer {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 4px;
	margin-top: var(--spacing-lg);

	.disclaimer-text {
		font-size: 11px;
		color: var(--text-tertiary);
		font-weight: 500;
	}
}

@keyframes spin {
	to { transform: rotate(360deg); }
}
</style>
