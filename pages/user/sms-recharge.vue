<template>
	<view class="sms-page">
		<!-- 背景装饰 -->
		<view class="sms-bg">
			<view class="sms-bg__orb sms-bg__orb--1"></view>
			<view class="sms-bg__orb sms-bg__orb--2"></view>
			<view class="sms-bg__grid"></view>
		</view>

		<!-- 导航栏 -->
		<CustomNavBar
			:center-title="'短信充值'"
			:transparent="true"
			:show-back="true"
			@back="goBack"
			@nav-height-ready="handleNavHeightReady"
		/>

		<!-- 内容 -->
		<view class="sms-content" :style="{ paddingTop: navPaddingTop }">

			<!-- 余额卡片 -->
			<view class="sms-balance-card">
				<view class="sms-balance-card__glow"></view>
				<view class="sms-balance-card__inner">
					<view class="sms-balance-card__left">
						<view class="sms-balance-card__icon-wrap">
							<l-icon name="mail" size="28" color="var(--primary-500)"></l-icon>
						</view>
						<view class="sms-balance-card__label">当前余额</view>
					</view>
					<view class="sms-balance-card__right">
						<view class="sms-balance-card__amount">
							<text class="sms-balance-card__amount-value">{{ smsBalance.balanceYuan || '0.00' }}</text>
							<text class="sms-balance-card__amount-unit">元</text>
						</view>
						<view class="sms-balance-card__sms-count">
							<l-icon name="chat" size="12" color="var(--text-tertiary)"></l-icon>
							<text>约可发送 {{ Math.floor((smsBalance.balance || 0) / (smsBalance.smsCost || 10)) }} 条短信</text>
						</view>
					</view>
				</view>
				<view class="sms-balance-card__bar">
					<view class="sms-balance-card__bar-fill" :style="{ width: balanceBarWidth }"></view>
				</view>
			</view>

			<!-- 充值套餐 -->
			<view class="sms-packages-card">
				<view class="sms-packages-card__header">
					<text class="sms-packages-card__title">充值套餐</text>
					<text class="sms-packages-card__subtitle">选择合适的套餐进行充值</text>
				</view>

				<view class="sms-packages-list">
					<view
						v-for="pkg in rechargePackages"
						:key="pkg.id"
						class="sms-pkg"
						:class="{ 'sms-pkg--selected': selectedPackage?.id === pkg.id }"
						@tap="selectPackage(pkg)"
					>
						<view class="sms-pkg__body">
							<view class="sms-pkg__check">
								<view class="sms-pkg__check-ring"></view>
								<l-icon v-if="selectedPackage?.id === pkg.id" name="check" size="12" color="#fff"></l-icon>
							</view>
							<view class="sms-pkg__info">
								<text class="sms-pkg__name">{{ pkg.name }}</text>
								<text class="sms-pkg__desc">{{ pkg.description }}</text>
							</view>
						</view>
						<view class="sms-pkg__footer">
							<text class="sms-pkg__price">¥{{ formatPrice(pkg.amount) }}</text>
							<view v-if="pkg.id === 3" class="sms-pkg__badge">推荐</view>
						</view>
					</view>
				</view>
			</view>

			<!-- 底部占位 -->
			<view class="sms-bottom-placeholder"></view>
		</view>

		<!-- 固定支付按钮 -->
		<view class="sms-pay-bar">
			<button
				class="sms-pay-btn"
				:class="{ 'sms-pay-btn--disabled': !selectedPackage || paying }"
				@tap="createOrder"
				:disabled="!selectedPackage || paying"
			>
				<template v-if="paying">
					<view class="sms-pay-btn__spinner"></view>
					<text>支付中...</text>
				</template>
				<template v-else-if="selectedPackage">
					<text>立即支付</text>
					<text class="sms-pay-btn__price">¥{{ formatPrice(selectedPackage.amount) }}</text>
				</template>
				<template v-else>
					<l-icon name="hand-up" size="14" color="rgba(255,255,255,0.6)"></l-icon>
					<text>请选择充值套餐</text>
				</template>
			</button>
		</view>

		<!-- 支付弹窗 -->
		<view v-if="showPayModal" class="sms-pay-modal" @tap="closePayModal">
			<view class="sms-pay-modal__sheet" @tap.stop>
				<!-- 拖拽指示条 -->
				<view class="sms-pay-modal__handle"></view>

				<view class="sms-pay-modal__header">
					<text class="sms-pay-modal__title">微信扫码支付</text>
					<view class="sms-pay-modal__close" @tap="closePayModal">
						<l-icon name="close" size="16" color="var(--text-secondary)"></l-icon>
					</view>
				</view>

				<!-- 二维码区域 -->
				<view class="sms-pay-modal__qr-wrap">
					<view class="sms-pay-modal__qr-frame">
						<!-- #ifdef MP-WEIXIN -->
						<canvas
							v-if="!qrCodeDataURL"
							canvas-id="qrcode"
							id="qrcode"
							class="sms-pay-modal__qr"
							:style="{ width: qrCodeSize + 'px', height: qrCodeSize + 'px' }"
						></canvas>
						<!-- #endif -->
						<image
							v-if="qrCodeDataURL"
							:src="qrCodeDataURL"
							class="sms-pay-modal__qr"
							:style="{ width: qrCodeSize + 'px', height: qrCodeSize + 'px' }"
							mode="aspectFit"
						></image>
						<view v-if="!qrCodeDataURL && qrCodeLoading" class="sms-pay-modal__qr-loading">
							<view class="sms-pay-modal__qr-spinner"></view>
						</view>
					</view>
				</view>

				<!-- 金额信息 -->
				<view class="sms-pay-modal__amount-row">
					<text class="sms-pay-modal__amount-label">支付金额</text>
					<text class="sms-pay-modal__amount-value">¥{{ selectedPackage ? formatPrice(selectedPackage.amount) : '0.00' }}</text>
				</view>
				<view class="sms-pay-modal__tip">
					<l-icon name="scan" size="13" color="var(--text-tertiary)"></l-icon>
					<text>请使用微信扫描上方二维码完成支付</text>
				</view>

				<!-- 支付状态 -->
				<view v-if="paymentStatus" class="sms-pay-modal__status" :class="`sms-pay-modal__status--${paymentStatus}`">
					<l-icon
						:name="paymentStatus === 'success' ? 'check-circle' : 'error'"
						size="18"
						:color="paymentStatus === 'success' ? '#10b981' : '#ef4444'"
					></l-icon>
					<text>{{ paymentStatus === 'success' ? '支付成功！' : '支付失败，请重试' }}</text>
				</view>

				<!-- #ifdef MP-WEIXIN -->
				<view class="sms-pay-modal__mp-actions">
					<view class="sms-pay-modal__mp-tip">
						<text class="sms-pay-modal__mp-tip-text">小程序内无法直接调起支付，请复制链接到浏览器打开，或保存二维码使用微信扫一扫</text>
					</view>
					<view class="sms-pay-modal__mp-btns">
						<button class="sms-pay-modal__mp-btn sms-pay-modal__mp-btn--copy" @tap="copyPaymentLink">
							<l-icon name="copy" size="14" color="#fff"></l-icon>
							<text>复制支付链接</text>
						</button>
						<button class="sms-pay-modal__mp-btn sms-pay-modal__mp-btn--save" @tap="saveQRCode" v-if="qrCodeDataURL">
							<l-icon name="download" size="14" color="#fff"></l-icon>
							<text>保存二维码</text>
						</button>
					</view>
				</view>
				<!-- #endif -->
			</view>
		</view>
	</view>
</template>

<script>
import { ref, computed } from 'vue';
import { onUnload as onUniUnload } from '@dcloudio/uni-app';
import { useSmsRecharge } from '@/composables/useSmsRecharge.js';
import { formatPrice } from '@/utils/smsPricing.js';
import CustomNavBar from '@/components/CustomNavBar.vue';

export default {
	components: { CustomNavBar },
	setup() {
		const navPaddingTop = ref('0px');

		const {
			smsBalance,
			rechargePackages,
			selectedPackage,
			paying,
			showPayModal,
			paymentStatus,
			currentOrderId,
			qrCodeSize,
			qrCodeDataURL,
			qrCodeLoading,
			currentPaymentUrl,
			loadSMSBalance,
			loadRechargePackages,
			selectPackage,
			createOrder,
			generateQRCode,
			generateQRCodeH5,
			generateQRCodeMP,
			generateQRCodeDataURL,
			copyPaymentLink,
			saveQRCode,
			startPaymentCheck,
			checkPaymentStatus,
			handlePaymentSuccess,
			handlePaymentFail,
			stopPaymentCheck,
			closePayModal,
			cleanupOfficialTheme
		} = useSmsRecharge();

		loadSMSBalance();
		loadRechargePackages();

		const handleNavHeightReady = (navInfo) => {
			navPaddingTop.value = navInfo.heightPx;
		};

		const goBack = () => uni.navigateBack();

		onUniUnload(() => {
			cleanupOfficialTheme();
		});

		const balanceBarWidth = computed(() => {
			const max = 100;
			const val = parseFloat(smsBalance.value?.balanceYuan || 0);
			const pct = Math.min((val / max) * 100, 100);
			return pct + '%';
		});

		return {
			navPaddingTop,
			smsBalance,
			rechargePackages,
			selectedPackage,
			paying,
			showPayModal,
			paymentStatus,
			currentOrderId,
			qrCodeSize,
			qrCodeDataURL,
			qrCodeLoading,
			currentPaymentUrl,
			selectPackage,
			createOrder,
			generateQRCode,
			generateQRCodeH5,
			generateQRCodeMP,
			generateQRCodeDataURL,
			copyPaymentLink,
			saveQRCode,
			startPaymentCheck,
			checkPaymentStatus,
			handlePaymentSuccess,
			handlePaymentFail,
			stopPaymentCheck,
			closePayModal,
			cleanupOfficialTheme,
			formatPrice,
			handleNavHeightReady,
			goBack,
			balanceBarWidth
		};
	},

	// #ifndef MP-WEIXIN
	data() {
		return {
			mpThemeClass: undefined,
			mpCurrentTheme: undefined,
			mpIsDarkMode: false,
			mpThemeColors: {}
		};
	}
	// #endif
};
</script>

<style lang="scss" scoped>
/* ── Page ── */
.sms-page {
	position: relative;
	width: 100%;
	min-height: 100vh;
	background: var(--bg-primary);
	overflow: hidden;
}

/* ── Background ── */
.sms-bg {
	position: fixed;
	inset: 0;
	pointer-events: none;
	z-index: 0;
}

.sms-bg__orb {
	position: absolute;
	border-radius: 50%;
	filter: blur(80rpx);
	opacity: 0.4;

	&--1 {
		width: 360rpx;
		height: 360rpx;
		background: radial-gradient(circle, rgba(37, 99, 235, 0.2) 0%, transparent 70%);
		top: -80rpx;
		right: -60rpx;
		animation: orbFloat 9s ease-in-out infinite;
	}

	&--2 {
		width: 280rpx;
		height: 280rpx;
		background: radial-gradient(circle, rgba(16, 185, 129, 0.15) 0%, transparent 70%);
		bottom: 300rpx;
		left: -40rpx;
		animation: orbFloat 11s ease-in-out infinite reverse;
	}
}

@keyframes orbFloat {
	0%, 100% { transform: translateY(0); }
	50% { transform: translateY(-20rpx); }
}

.sms-bg__grid {
	position: absolute;
	inset: 0;
	background-image: radial-gradient(circle, rgba(99, 102, 241, 0.05) 1px, transparent 1px);
	background-size: 32rpx 32rpx;
}

/* ── Content ── */
.sms-content {
	position: relative;
	z-index: 1;
	padding: 0 32rpx 220rpx;
}

/* ── Balance Card ── */
.sms-balance-card {
	position: relative;
	margin: 16rpx 0 28rpx;
	background: rgba(255, 255, 255, 0.85);
	backdrop-filter: blur(24px) saturate(180%);
	-webkit-backdrop-filter: blur(24px) saturate(180%);
	border-radius: 28rpx;
	padding: 36rpx;
	border: 1rpx solid rgba(255, 255, 255, 0.6);
	box-shadow:
		0 8rpx 32rpx rgba(37, 99, 235, 0.1),
		0 2rpx 8rpx rgba(0, 0, 0, 0.04),
		inset 0 1px 0 rgba(255, 255, 255, 0.9);
	overflow: hidden;
}

.sms-balance-card__glow {
	position: absolute;
	top: -40rpx;
	right: -40rpx;
	width: 200rpx;
	height: 200rpx;
	border-radius: 50%;
	background: radial-gradient(circle, rgba(37, 99, 235, 0.12) 0%, transparent 70%);
	pointer-events: none;
}

.sms-balance-card__inner {
	display: flex;
	align-items: center;
	gap: 28rpx;
}

.sms-balance-card__left {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 10rpx;
}

.sms-balance-card__icon-wrap {
	width: 80rpx;
	height: 80rpx;
	border-radius: 20rpx;
	background: linear-gradient(135deg, rgba(37, 99, 235, 0.1), rgba(37, 99, 235, 0.04));
	border: 1rpx solid rgba(37, 99, 235, 0.12);
	display: flex;
	align-items: center;
	justify-content: center;
}

.sms-balance-card__label {
	font-size: 22rpx;
	color: var(--text-tertiary);
	font-weight: 500;
}

.sms-balance-card__right {
	flex: 1;
}

.sms-balance-card__amount {
	display: flex;
	align-items: baseline;
	gap: 6rpx;
	margin-bottom: 8rpx;
}

.sms-balance-card__amount-value {
	font-size: 64rpx;
	font-weight: 800;
	color: var(--text-primary);
	letter-spacing: -1px;
	line-height: 1;
}

.sms-balance-card__amount-unit {
	font-size: 32rpx;
	color: var(--text-tertiary);
	font-weight: 500;
}

.sms-balance-card__sms-count {
	display: flex;
	align-items: center;
	gap: 8rpx;
	font-size: 24rpx;
	color: var(--text-tertiary);
}

.sms-balance-card__bar {
	height: 6rpx;
	background: rgba(226, 232, 240, 0.6);
	border-radius: 100rpx;
	margin-top: 24rpx;
	overflow: hidden;
}

.sms-balance-card__bar-fill {
	height: 100%;
	background: linear-gradient(90deg, var(--primary-500), var(--primary-300));
	border-radius: 100rpx;
	transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}

/* ── Packages Card ── */
.sms-packages-card {
	background: rgba(255, 255, 255, 0.82);
	backdrop-filter: blur(20px) saturate(180%);
	-webkit-backdrop-filter: blur(20px) saturate(180%);
	border-radius: 28rpx;
	padding: 32rpx;
	border: 1rpx solid rgba(255, 255, 255, 0.6);
	box-shadow:
		0 4rpx 20rpx rgba(37, 99, 235, 0.06),
		0 1rpx 4rpx rgba(0, 0, 0, 0.04),
		inset 0 1px 0 rgba(255, 255, 255, 0.9);
}

.sms-packages-card__header {
	margin-bottom: 28rpx;
}

.sms-packages-card__title {
	display: block;
	font-size: 34rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.sms-packages-card__subtitle {
	display: block;
	font-size: 24rpx;
	color: var(--text-tertiary);
	margin-top: 4rpx;
}

.sms-packages-list {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
}

/* ── Package Item ── */
.sms-pkg {
	background: rgba(255, 255, 255, 0.6);
	border: 2rpx solid rgba(226, 232, 240, 0.6);
	border-radius: 20rpx;
	padding: 24rpx;
	transition: all 0.2s var(--ease-spring);
	cursor: pointer;

	&:active { transform: scale(0.98); }

	&--selected {
		border-color: var(--primary-500);
		background: rgba(59, 130, 246, 0.06);
		box-shadow: 0 0 0 1rpx var(--primary-500),
			0 0 0 4rpx rgba(59, 130, 246, 0.15),
			0 8rpx 24rpx rgba(37, 99, 235, 0.12);
	}
}

.sms-pkg__body {
	display: flex;
	align-items: center;
	gap: 20rpx;
	margin-bottom: 16rpx;
}

.sms-pkg__check {
	position: relative;
	width: 40rpx;
	height: 40rpx;
	flex-shrink: 0;
	overflow: hidden;
	border-radius: 50%;

	.sms-pkg__check-ring {
		position: absolute;
		top: 0;
		left: 0;
		width: 40rpx;
		height: 40rpx;
		border-radius: 50%;
		border: 2rpx solid rgba(148, 163, 184, 0.3);
		transition: all 0.2s var(--ease-spring);
		box-sizing: border-box;
	}

	.sms-pkg--selected & .sms-pkg__check-ring {
		background: var(--primary-500);
		border-color: var(--primary-500);
		transform: scale(1.1);
	}
}

.sms-pkg__info {
	flex: 1;
	min-width: 0;
}

.sms-pkg__name {
	display: block;
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	margin-bottom: 4rpx;
}

.sms-pkg__desc {
	display: block;
	font-size: 24rpx;
	color: var(--text-tertiary);
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.sms-pkg__footer {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding-left: 60rpx;
}

.sms-pkg__price {
	font-size: 36rpx;
	font-weight: 800;
	color: var(--primary-600);
}

.sms-pkg__badge {
	font-size: 20rpx;
	font-weight: 700;
	color: #fff;
	background: linear-gradient(135deg, #f59e0b, #f97316);
	padding: 4rpx 14rpx;
	border-radius: 100rpx;
	box-shadow: 0 4rpx 12rpx rgba(245, 158, 11, 0.3);
}

.sms-pkg--selected .sms-pkg__badge {
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	box-shadow: 0 4rpx 12rpx rgba(37, 99, 235, 0.3);
}

.sms-bottom-placeholder {
	height: 200rpx;
}

/* ── Pay Bar ── */
.sms-pay-bar {
	position: fixed;
	bottom: 0;
	left: 0;
	right: 0;
	z-index: 100;
	padding: 0 32rpx calc(env(safe-area-inset-bottom) + 20rpx);
	background: linear-gradient(to top, var(--bg-primary) 60%, transparent);
}

.sms-pay-btn {
	width: 100%;
	height: 100rpx;
	border-radius: 50rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	color: #fff;
	font-size: 32rpx;
	font-weight: 700;
	border: none;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 10rpx;
	box-shadow: 0 8rpx 32rpx rgba(37, 99, 235, 0.35);
	transition: all 0.2s var(--ease-spring);
	position: relative;
	overflow: hidden;

	&::before {
		content: '';
		position: absolute;
		top: 0;
		left: -100%;
		width: 100%;
		height: 100%;
		background: linear-gradient(90deg, transparent, rgba(255,255,255,0.15), transparent);
		animation: payShimmer 2.5s ease-in-out infinite;
	}

	&:active:not(.sms-pay-btn--disabled) { transform: scale(0.97); }

	&--disabled {
		background: linear-gradient(135deg, #9ca3af, #6b7280);
		box-shadow: none;
		&::before { display: none; }
	}
}

@keyframes payShimmer {
	0% { left: -100%; }
	50%, 100% { left: 100%; }
}

.sms-pay-btn__price {
	font-size: 36rpx;
	font-weight: 800;
}

.sms-pay-btn__spinner {
	width: 32rpx;
	height: 32rpx;
	border: 3rpx solid rgba(255,255,255,0.3);
	border-top-color: #fff;
	border-radius: 50%;
	animation: paySpin 0.8s linear infinite;
}

@keyframes paySpin { to { transform: rotate(360deg); } }

/* ── Pay Modal ── */
.sms-pay-modal {
	position: fixed;
	inset: 0;
	z-index: 1000;
	background: rgba(0, 0, 0, 0.5);
	backdrop-filter: blur(4px);
	display: flex;
	align-items: flex-end;
	justify-content: center;
	animation: modalFadeIn 0.25s ease-out;
}

@keyframes modalFadeIn {
	from { opacity: 0; }
	to { opacity: 1; }
}

.sms-pay-modal__sheet {
	width: 100%;
	max-width: 750rpx;
	background: var(--bg-primary);
	border-radius: 32rpx 32rpx 0 0;
	padding: 0 40rpx calc(env(safe-area-inset-bottom) + 40rpx);
	animation: sheetSlideUp 0.3s cubic-bezier(0.32, 0.72, 0, 1);
	position: relative;
}

@keyframes sheetSlideUp {
	from { transform: translateY(100%); }
	to { transform: translateY(0); }
}

.sms-pay-modal__handle {
	width: 80rpx;
	height: 6rpx;
	background: rgba(148, 163, 184, 0.3);
	border-radius: 100rpx;
	margin: 20rpx auto 0;
}

.sms-pay-modal__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 32rpx 0 28rpx;
}

.sms-pay-modal__title {
	font-size: 36rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.sms-pay-modal__close {
	width: 64rpx;
	height: 64rpx;
	border-radius: 50%;
	background: var(--bg-secondary);
	display: flex;
	align-items: center;
	justify-content: center;
	transition: all 0.15s ease;
	&:active { transform: scale(0.9); background: var(--bg-muted); }
}

/* QR Area */
.sms-pay-modal__qr-wrap {
	display: flex;
	justify-content: center;
	margin-bottom: 32rpx;
}

.sms-pay-modal__qr-frame {
	position: relative;
	padding: 24rpx;
	background: #fff;
	border-radius: 24rpx;
	border: 1rpx solid rgba(226, 232, 240, 0.8);
	box-shadow: 0 8rpx 32rpx rgba(0, 0, 0, 0.06);

	&::before {
		content: '';
		position: absolute;
		inset: -1rpx;
		border-radius: 25rpx;
		background: linear-gradient(135deg, var(--primary-400), rgba(16, 185, 129, 0.5));
		z-index: -1;
	}
}

.sms-pay-modal__qr {
	display: block;
	border-radius: 8rpx;
}

.sms-pay-modal__qr-loading {
	position: absolute;
	inset: 24rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255,255,255,0.9);
	border-radius: 8rpx;
}

.sms-pay-modal__qr-spinner {
	width: 56rpx;
	height: 56rpx;
	border: 4rpx solid rgba(37, 99, 235, 0.12);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: paySpin 0.8s linear infinite;
}

/* Amount & Tip */
.sms-pay-modal__amount-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 10rpx;
}

.sms-pay-modal__amount-label {
	font-size: 28rpx;
	color: var(--text-secondary);
}

.sms-pay-modal__amount-value {
	font-size: 40rpx;
	font-weight: 800;
	color: var(--primary-600);
}

.sms-pay-modal__tip {
	display: flex;
	align-items: center;
	gap: 8rpx;
	font-size: 24rpx;
	color: var(--text-tertiary);
	margin-bottom: 24rpx;
}

/* Status */
.sms-pay-modal__status {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 10rpx;
	padding: 20rpx;
	border-radius: 16rpx;
	margin-bottom: 24rpx;
	font-size: 28rpx;
	font-weight: 600;

	&--success {
		background: rgba(16, 185, 129, 0.08);
		color: #10b981;
	}

	&--error {
		background: rgba(239, 68, 68, 0.08);
		color: #ef4444;
	}
}

/* MP Actions */
.sms-pay-modal__mp-actions {
	padding-top: 4rpx;
}

.sms-pay-modal__mp-tip {
	background: rgba(245, 158, 11, 0.08);
	border: 1rpx solid rgba(245, 158, 11, 0.2);
	border-radius: 16rpx;
	padding: 20rpx;
	margin-bottom: 20rpx;

	.sms-pay-modal__mp-tip-text {
		font-size: 24rpx;
		color: #b45309;
		line-height: 1.6;
	}
}

.sms-pay-modal__mp-btns {
	display: flex;
	gap: 16rpx;
}

.sms-pay-modal__mp-btn {
	flex: 1;
	height: 88rpx;
	border-radius: 20rpx;
	border: none;
	font-size: 28rpx;
	font-weight: 700;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 8rpx;
	transition: all 0.15s ease;

	&:active { transform: scale(0.97); }

	&--copy {
		background: linear-gradient(135deg, #6366f1, #8b5cf6);
		color: #fff;
		box-shadow: 0 6rpx 20rpx rgba(99, 102, 241, 0.35);
	}

	&--save {
		background: linear-gradient(135deg, #06a054, #10b981);
		color: #fff;
		box-shadow: 0 6rpx 20rpx rgba(7, 193, 96, 0.35);
	}
}
</style>
