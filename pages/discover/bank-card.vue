<template>
	<view class="bank-page">
		<!-- 自定义导航栏 -->
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">我的银行卡</text>
				<view class="nav-bar__placeholder"></view>
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
				<button class="state-btn" @tap="fetchBankInfo">重试</button>
			</view>

			<!-- 无银行卡：引导添加 -->
			<view v-else-if="!hasBankCard" class="empty-content">
				<view class="empty-card">
					<view class="empty-icon">
						<l-icon name="credit-card-outline" style="font-size: 56px; color: #10b981;"></l-icon>
					</view>
					<text class="empty-title">暂未绑定银行卡</text>
					<text class="empty-desc">添加您的银行卡信息，方便学校发放奖学金、助学金等</text>
					<button class="empty-btn" @tap="openEdit">立即添加</button>
				</view>
			</view>

			<!-- 银行卡信息 -->
			<view v-else class="bank-content">
				<!-- 银行卡卡片 -->
				<view class="bank-card" @tap="openEdit">
					<view class="bank-card__bg"></view>
					<view class="bank-card__content">
						<view class="bank-card__header">
							<view class="bank-card__icon">
								<l-icon name="credit-card" style="font-size: 24px; color: #fff;"></l-icon>
							</view>
							<text class="bank-card__bank-name">{{ bankInfo.bankname || '未知银行' }}</text>
							<view class="bank-card__edit-tag">
								<l-icon name="pencil" style="font-size: 12px; color: #fff;"></l-icon>
								<text style="font-size: 20rpx; color: #fff; margin-left: 4rpx;">编辑</text>
							</view>
						</view>
						<view class="bank-card__number">
							{{ formatCardNumber(bankInfo.bankcardnumber) }}
						</view>
						<view class="bank-card__footer">
							<view class="bank-card__info-item">
								<text class="bank-card__label">持卡人</text>
								<text class="bank-card__value">{{ bankInfo.cardholder || bankInfo.name || '--' }}</text>
							</view>
							<view class="bank-card__info-item">
								<text class="bank-card__label">卡类型</text>
								<text class="bank-card__value">{{ bankInfo.bankcardtype || '--' }}</text>
							</view>
						</view>
					</view>
				</view>

				<!-- 账户信息 -->
				<view class="info-card">
					<view class="info-section">
						<text class="info-section__title">账户信息</text>
						<view class="info-list">
							<view class="info-item">
								<view class="info-item__left"><l-icon name="wallet" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">开户银行</text></view>
								<text class="info-item__value">{{ bankInfo.bankname || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="map-marker" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">开户省市</text></view>
								<text class="info-item__value">{{ bankInfo.bankprovincecity || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="domain" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">开户支行</text></view>
								<text class="info-item__value">{{ bankInfo.banksubbranch || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="book" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">银行卡号</text></view>
								<text class="info-item__value">{{ maskCardNumber(bankInfo.bankcardnumber) }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="tag" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">卡类型</text></view>
								<text class="info-item__value">{{ bankInfo.bankcardtype || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="account-circle" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">持卡人</text></view>
								<text class="info-item__value">{{ bankInfo.cardholder || '--' }}</text>
							</view>
						</view>
					</view>

					<!-- 学生档案（从银行卡接口获取的档案补充字段） -->
					<view class="info-section" v-if="hasStudentInfo">
						<text class="info-section__title">学生档案</text>
						<view class="info-list">
							<view class="info-item">
								<view class="info-item__left"><l-icon name="user" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">姓名</text></view>
								<text class="info-item__value">{{ bankInfo.name || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="numeric" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">学号</text></view>
								<text class="info-item__value">{{ bankInfo.studentnumber || bankInfo.studynumber || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="user" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">性别</text></view>
								<text class="info-item__value">{{ getGenderText(bankInfo.gender) }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="cake" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">出生日期</text></view>
								<text class="info-item__value">{{ formatBirthday(bankInfo.birthday) }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="shield" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">身份证号</text></view>
								<text class="info-item__value">{{ maskIdCard(bankInfo.identitycard) }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="star" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">民族</text></view>
								<text class="info-item__value">{{ bankInfo.nation || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="flag" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">政治面貌</text></view>
								<text class="info-item__value">{{ bankInfo.politicsstatus || '--' }}</text>
							</view>
						</view>
					</view>

					<!-- 学业信息 -->
					<view class="info-section" v-if="hasAcademicInfo">
						<text class="info-section__title">学业信息</text>
						<view class="info-list">
							<view class="info-item">
								<view class="info-item__left"><l-icon name="school" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">院系</text></view>
								<text class="info-item__value">{{ bankInfo.facultystation || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="book" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">专业</text></view>
								<text class="info-item__value">{{ bankInfo.professionname || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="account-group" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">行政班</text></view>
								<text class="info-item__value">{{ bankInfo.classnumber || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="map" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">校区</text></view>
								<text class="info-item__value">{{ bankInfo.branchcourts || '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="certificate" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">学历层次</text></view>
								<text class="info-item__value">{{ getHierarchyText(bankInfo.hierarchy) }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="clock-outline" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">学制</text></view>
								<text class="info-item__value">{{ bankInfo.studysystem ? bankInfo.studysystem + '年' : '--' }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="calendar-check" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">入学日期</text></view>
								<text class="info-item__value">{{ formatDate(bankInfo.entrancedate) }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="calendar-clock" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">预计毕业</text></view>
								<text class="info-item__value">{{ formatDate(bankInfo.expectedgraduatedate) }}</text>
							</view>
							<view class="info-item">
								<view class="info-item__left"><l-icon name="cellphone" style="font-size: 16px; color: #94a3b8;"></l-icon><text class="info-item__label">手机号</text></view>
								<text class="info-item__value">{{ bankInfo.phoneNumber || '--' }}</text>
							</view>
						</view>
					</view>
				</view>

				<view class="bottom-safe-area"></view>
			</view>
		</scroll-view>

		<!-- 编辑弹窗 -->
		<t-popup
			:visible="editVisible"
			placement="bottom"
			@close="closeEdit"
			:close-on-overlay-click="true"
		>
			<view class="edit-popup">
				<!-- 弹窗头部 -->
				<view class="edit-popup__header">
					<text class="edit-popup__title">{{ hasBankCard ? '编辑银行卡' : '添加银行卡' }}</text>
					<view class="edit-popup__close" @tap="closeEdit">
						<l-icon name="close" style="font-size: 18px; color: #94a3b8;"></l-icon>
					</view>
				</view>

				<!-- 表单 -->
				<view class="edit-form">
					<view class="form-item">
						<text class="form-label">银行卡号</text>
						<input
							class="form-input"
							v-model="formData.bankcardnumber"
							placeholder="请输入银行卡号"
							type="number"
							maxlength="25"
						/>
					</view>
					<view class="form-item">
						<text class="form-label">手机号码</text>
						<input
							class="form-input"
							v-model="formData.phoneNumber"
							placeholder="请输入手机号码"
							type="number"
							maxlength="11"
						/>
					</view>
					<view class="form-item">
						<text class="form-label">开户银行</text>
						<input
							class="form-input"
							v-model="formData.bankname"
							placeholder="如：中国建设银行"
						/>
					</view>
					<view class="form-item">
						<text class="form-label">开户省市</text>
						<input
							class="form-input"
							v-model="formData.bankprovincecity"
							placeholder="如：河南省漯河市"
						/>
					</view>
					<view class="form-item">
						<text class="form-label">开户支行</text>
						<input
							class="form-input"
							v-model="formData.banksubbranch"
							placeholder="请输入开户支行全称"
						/>
					</view>
				</view>

				<!-- 提交按钮 -->
				<view class="edit-popup__footer">
					<button class="edit-submit-btn" :disabled="submitting" @tap="submitEdit">
						<view v-if="submitting" class="btn-spinner"></view>
						<text v-else>{{ submitting ? '保存中...' : '保存' }}</text>
					</button>
				</view>
			</view>
		</t-popup>

		<!-- Toast 提示 -->
		<t-popup
			:visible="toastVisible"
			:overlay-attributes="{timeout: 0}"
			placement="center"
		>
			<view class="toast-popup">
				<l-icon v-if="toastType === 'success'" name="check-circle" style="font-size: 48px; color: #10b981;"></l-icon>
				<l-icon v-else name="alert-circle" style="font-size: 48px; color: #ef4444;"></l-icon>
				<text class="toast-text">{{ toastMessage }}</text>
			</view>
		</t-popup>
	</view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getStudentBank, editStudentBank } from '../../pages/api/discover.js'

const statusBarHeight = ref(20)
const bankInfo = ref({})
const loading = ref(false)
const error = ref('')

const editVisible = ref(false)
const submitting = ref(false)

const toastVisible = ref(false)
const toastMessage = ref('')
const toastType = ref('success')

const formData = ref({
	bankcardnumber: '',
	phoneNumber: '',
	bankname: '',
	bankprovincecity: '',
	banksubbranch: ''
})

const hasBankCard = computed(() => !!(
	bankInfo.value.bankcardnumber ||
	bankInfo.value.bankname ||
	bankInfo.value.banksubbranch
))

const hasStudentInfo = computed(() => !!(
	bankInfo.value.name ||
	bankInfo.value.studentnumber ||
	bankInfo.value.studynumber ||
	bankInfo.value.gender ||
	bankInfo.value.birthday ||
	bankInfo.value.identitycard ||
	bankInfo.value.nation ||
	bankInfo.value.politicsstatus
))

const hasAcademicInfo = computed(() => !!(
	bankInfo.value.facultystation ||
	bankInfo.value.professionname ||
	bankInfo.value.classnumber ||
	bankInfo.value.branchcourts ||
	bankInfo.value.hierarchy ||
	bankInfo.value.studysystem ||
	bankInfo.value.entrancedate ||
	bankInfo.value.expectedgraduatedate ||
	bankInfo.value.phoneNumber
))

function goBack() {
	uni.navigateBack()
}

function showToast(message, type = 'success') {
	toastMessage.value = message
	toastType.value = type
	toastVisible.value = true
	setTimeout(() => {
		toastVisible.value = false
	}, 1500)
}

function openEdit() {
	// 用已有数据填充表单
	formData.value = {
		bankcardnumber: bankInfo.value.bankcardnumber || '',
		phoneNumber: bankInfo.value.phoneNumber || '',
		bankname: bankInfo.value.bankname || '',
		bankprovincecity: bankInfo.value.bankprovincecity || '',
		banksubbranch: bankInfo.value.banksubbranch || ''
	}
	editVisible.value = true
}

function closeEdit() {
	editVisible.value = false
}

async function submitEdit() {
	// 基础校验
	if (!formData.value.bankcardnumber) {
		showToast('请输入银行卡号', 'error')
		return
	}
	if (!formData.value.phoneNumber) {
		showToast('请输入手机号码', 'error')
		return
	}
	if (!/^1[3-9]\d{9}$/.test(formData.value.phoneNumber)) {
		showToast('手机号格式不正确', 'error')
		return
	}
	if (!formData.value.bankname) {
		showToast('请输入开户银行', 'error')
		return
	}

	submitting.value = true
	try {
		const res = await editStudentBank(formData.value)
		if (res && res.success) {
			showToast(res.message || '保存成功')
			closeEdit()
			// 重新拉取最新数据
			await fetchBankInfo()
		} else {
			showToast(res?.message || '保存失败', 'error')
		}
	} catch (e) {
		console.error('保存银行卡失败', e)
		showToast(e.message || '保存失败，请稍后重试', 'error')
	} finally {
		submitting.value = false
	}
}

async function fetchBankInfo() {
	loading.value = true
	error.value = ''

	try {
		const res = await getStudentBank()
		if (res && res.success && res.result) {
			bankInfo.value = res.result
		} else {
			throw new Error(res?.message || '获取银行卡信息失败')
		}
	} catch (e) {
		console.error('获取银行卡信息失败', e)
		error.value = e.message || '获取银行卡信息失败，请检查网络'
	} finally {
		loading.value = false
	}
}

function formatCardNumber(cardNumber) {
	if (!cardNumber) return '**** **** **** ****'
	const cleaned = cardNumber.replace(/\s/g, '')
	if (cleaned.length < 4) return cardNumber
	const last4 = cleaned.slice(-4)
	return '**** **** **** ' + last4
}

function maskCardNumber(cardNumber) {
	if (!cardNumber) return '--'
	const cleaned = cardNumber.replace(/\s/g, '')
	if (cleaned.length < 8) return cleaned
	return cleaned.slice(0, 4) + ' **** **** ' + cleaned.slice(-4)
}

function maskIdCard(id) {
	if (!id || id.length < 8) return id || '--'
	return id.substring(0, 4) + '****' + id.substring(id.length - 4)
}

function getGenderText(val) {
	if (val === '0' || val === 0 || val === '男') return '男'
	if (val === '1' || val === 1 || val === '女') return '女'
	return '--'
}

function getHierarchyText(val) {
	if (!val) return '--'
	const map = { '本科': '本科', '专科': '专科', '硕士研究生': '硕士研究生', '博士研究生': '博士研究生' }
	return map[val] || val
}

function formatBirthday(val) {
	if (!val) return '--'
	if (val.length === 8) {
		return val.substring(0, 4) + '-' + val.substring(4, 6) + '-' + val.substring(6, 8)
	}
	return val
}

function formatDate(val) {
	if (!val) return '--'
	const cleaned = String(val).replace(/-/g, '')
	if (cleaned.length === 8) {
		return cleaned.substring(0, 4) + '-' + cleaned.substring(4, 6) + '-' + cleaned.substring(6, 8)
	}
	return val
}

onLoad(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 20
	fetchBankInfo()
})
</script>

<style lang="scss" scoped>
@import url('../../static/css/home-page.css');

.bank-page {
	width: 100%;
	height: 100vh;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
}

.nav-bar {
	background: linear-gradient(135deg, #10b981, #34d399);
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
	border: 4rpx solid rgba(16, 185, 129, 0.15);
	border-top-color: #10b981;
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
	background: linear-gradient(135deg, #10b981, #34d399);
	color: #fff;
	font-size: 26rpx;
	font-weight: 600;
	border-radius: 32rpx;
	border: none;
}

/* 空状态 */
.empty-content {
	padding: 80rpx 32rpx;
}

.empty-card {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 80rpx 40rpx;
	background: #fff;
	border-radius: 32rpx;
	box-shadow: 0 2rpx 12rpx rgba(16, 185, 129, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.empty-icon {
	width: 160rpx;
	height: 160rpx;
	border-radius: 80rpx;
	background: rgba(16, 185, 129, 0.08);
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 32rpx;
}

.empty-title {
	font-size: 36rpx;
	font-weight: 700;
	color: var(--text-primary);
	margin-bottom: 16rpx;
}

.empty-desc {
	font-size: 26rpx;
	color: var(--text-secondary);
	text-align: center;
	line-height: 1.6;
	margin-bottom: 48rpx;
}

.empty-btn {
	padding: 20rpx 64rpx;
	background: linear-gradient(135deg, #10b981, #34d399);
	color: #fff;
	font-size: 28rpx;
	font-weight: 700;
	border-radius: 40rpx;
	border: none;
	box-shadow: 0 8rpx 32rpx rgba(16, 185, 129, 0.3);
}

.bank-content {
	padding: 24rpx;
	padding-top: 0;
	margin-top: -30rpx;
	position: relative;
	z-index: 1;
}

/* 银行卡卡片 */
.bank-card {
	position: relative;
	border-radius: 32rpx;
	overflow: hidden;
	margin-bottom: 24rpx;
	box-shadow: 0 8rpx 40rpx rgba(16, 185, 129, 0.3);
	cursor: pointer;
}

.bank-card__bg {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: linear-gradient(135deg, #10b981 0%, #059669 30%, #047857 60%, #065f46 100%);
}

.bank-card__content {
	position: relative;
	z-index: 1;
	padding: 40rpx 36rpx;
}

.bank-card__header {
	display: flex;
	align-items: center;
	gap: 16rpx;
	margin-bottom: 32rpx;
}

.bank-card__icon {
	width: 72rpx;
	height: 72rpx;
	border-radius: 20rpx;
	background: rgba(255, 255, 255, 0.15);
	display: flex;
	align-items: center;
	justify-content: center;
}

.bank-card__bank-name {
	font-size: 32rpx;
	font-weight: 700;
	color: #fff;
	letter-spacing: 1px;
	flex: 1;
}

.bank-card__edit-tag {
	display: flex;
	align-items: center;
	padding: 8rpx 16rpx;
	background: rgba(255, 255, 255, 0.15);
	border-radius: 20rpx;
	border: 1px solid rgba(255, 255, 255, 0.2);
}

.bank-card__number {
	font-size: 36rpx;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.95);
	letter-spacing: 4rpx;
	font-family: 'Courier New', monospace;
	margin-bottom: 32rpx;
}

.bank-card__footer {
	display: flex;
	justify-content: space-between;
}

.bank-card__info-item {
	display: flex;
	flex-direction: column;
	gap: 8rpx;
}

.bank-card__label {
	font-size: 20rpx;
	color: rgba(255, 255, 255, 0.6);
	text-transform: uppercase;
	letter-spacing: 1px;
}

.bank-card__value {
	font-size: 26rpx;
	font-weight: 600;
	color: #fff;
}

/* 信息卡片 */
.info-card {
	background: #fff;
	border-radius: 24rpx;
	overflow: hidden;
	box-shadow: 0 2rpx 12rpx rgba(16, 185, 129, 0.06);
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

.bottom-safe-area {
	height: 48rpx;
}

/* 编辑弹窗 */
.edit-popup {
	background: #fff;
	border-radius: 32rpx 32rpx 0 0;
	padding-bottom: calc(env(safe-area-inset-bottom) + 32rpx);
}

.edit-popup__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 32rpx 32rpx 24rpx;
	border-bottom: 1px solid rgba(226, 232, 240, 0.7);
}

.edit-popup__title {
	font-size: 32rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.edit-popup__close {
	width: 56rpx;
	height: 56rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.edit-form {
	padding: 24rpx 32rpx;
}

.form-item {
	margin-bottom: 24rpx;
}

.form-label {
	display: block;
	font-size: 26rpx;
	font-weight: 600;
	color: var(--text-secondary);
	margin-bottom: 12rpx;
}

.form-input {
	width: 100%;
	height: 88rpx;
	padding: 0 24rpx;
	background: var(--bg-secondary);
	border: 1px solid rgba(226, 232, 240, 0.8);
	border-radius: 16rpx;
	font-size: 28rpx;
	color: var(--text-primary);
	box-sizing: border-box;
}

.form-input:focus {
	border-color: #10b981;
}

.edit-popup__footer {
	padding: 0 32rpx 24rpx;
}

.edit-submit-btn {
	width: 100%;
	height: 88rpx;
	background: linear-gradient(135deg, #10b981, #34d399);
	color: #fff;
	font-size: 30rpx;
	font-weight: 700;
	border-radius: 44rpx;
	border: none;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 8rpx 32rpx rgba(16, 185, 129, 0.3);
}

.edit-submit-btn[disabled] {
	opacity: 0.6;
}

.btn-spinner {
	width: 36rpx;
	height: 36rpx;
	border: 3rpx solid rgba(255, 255, 255, 0.3);
	border-top-color: #fff;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

/* Toast */
.toast-popup {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 48rpx;
	background: #fff;
	border-radius: 24rpx;
	min-width: 360rpx;
}

.toast-text {
	font-size: 28rpx;
	font-weight: 600;
	color: var(--text-primary);
	margin-top: 16rpx;
	text-align: center;
}
</style>
