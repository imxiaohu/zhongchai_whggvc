<template>
	<view class="signin-page">
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">实习签到</text>
				<view class="nav-bar__placeholder"></view>
			</view>
		</view>

		<scroll-view class="content-scroll" scroll-y
			:refresher-enabled="true" @refresherrefresh="onRefresh" :refresher-triggered="refreshing">
			<!-- 加载中 -->
			<view v-if="loading" class="state-container">
				<view class="state-spinner"></view>
				<text class="state-text">加载中...</text>
			</view>

			<!-- 错误状态 -->
			<view v-else-if="error" class="state-container state-container--error">
				<view class="state-icon-circle state-icon-circle--error">!</view>
				<text class="state-text">{{ error }}</text>
				<button class="state-btn" @tap="fetchData">重试</button>
			</view>

			<!-- 无签到计划 -->
			<view v-else-if="!loading && !planData" class="state-container">
				<view class="state-icon-wrap">
					<l-icon name="no-result" style="font-size: 48px; color: #94a3b8;"></l-icon>
				</view>
				<text class="state-text">暂无签到计划</text>
				<text class="state-sub">当前没有需要签到的实习安排</text>
			</view>

			<!-- 签到计划内容 -->
			<view v-else class="plan-container">
				<view class="plan-card">
					<view class="plan-card__title">签到计划</view>
					<view class="plan-card__row" v-if="planData.startDate">
						<l-icon name="calendar" style="font-size: 16px; color: #94a3b8; margin-right: 8rpx;"></l-icon>
						<text class="plan-card__label">开始日期：</text>
						<text class="plan-card__value">{{ planData.startDate }}</text>
					</view>
					<view class="plan-card__row" v-if="planData.endDate">
						<l-icon name="calendar-end" style="font-size: 16px; color: #94a3b8; margin-right: 8rpx;"></l-icon>
						<text class="plan-card__label">结束日期：</text>
						<text class="plan-card__value">{{ planData.endDate }}</text>
					</view>
					<view class="plan-card__row" v-if="planData.companyName">
						<l-icon name="domain" style="font-size: 16px; color: #94a3b8; margin-right: 8rpx;"></l-icon>
						<text class="plan-card__label">实习单位：</text>
						<text class="plan-card__value">{{ planData.companyName }}</text>
					</view>
					<view class="plan-card__row" v-if="planData.position">
						<l-icon name="briefcase" style="font-size: 16px; color: #94a3b8; margin-right: 8rpx;"></l-icon>
						<text class="plan-card__label">实习岗位：</text>
						<text class="plan-card__value">{{ planData.position }}</text>
					</view>
					<view class="plan-card__row" v-if="planData.signInTimes || planData.signInTime">
						<l-icon name="clock" style="font-size: 16px; color: #94a3b8; margin-right: 8rpx;"></l-icon>
						<text class="plan-card__label">签到次数：</text>
						<text class="plan-card__value">{{ planData.signInTimes || planData.signInTime }}</text>
					</view>
				</view>

				<!-- 签到记录 -->
				<view class="records-section">
					<view class="records-section__title">签到记录</view>
					<view v-if="planData.signInRecords && planData.signInRecords.length > 0">
						<view v-for="(record, index) in planData.signInRecords" :key="index" class="record-card">
							<view class="record-card__date">
								<text class="record-card__date-text">{{ record.date || record.signInDate }}</text>
								<view class="record-card__status" :class="record.status ? 'record-card__status--' + record.status : ''">
									<text class="record-card__status-text">{{ record.statusText || record.status || '已签到' }}</text>
								</view>
							</view>
							<view class="record-card__info" v-if="record.time || record.signInTime">
								<l-icon name="clock" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
								<text class="record-card__info-text">{{ record.time || record.signInTime }}</text>
							</view>
							<view class="record-card__info" v-if="record.location || record.signInLocation">
								<l-icon name="map-marker" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
								<text class="record-card__info-text">{{ record.location || record.signInLocation }}</text>
							</view>
						</view>
					</view>
					<view v-else class="records-empty">
						<l-icon name="calendar-blank" style="font-size: 32px; color: #94a3b8;"></l-icon>
						<text class="records-empty__text">暂无签到记录</text>
					</view>
				</view>
			</view>
		</scroll-view>
	</view>
</template>

<script>
import { getInternshipSignInPlan } from '../../api/discover.js';

export default {
	data() {
		return {
			statusBarHeight: 20,
			loading: false,
			refreshing: false,
			error: '',
			planData: null
		}
	},
	onLoad() {
		const systemInfo = uni.getSystemInfoSync();
		this.statusBarHeight = systemInfo.statusBarHeight || 20;
		this.fetchData();
	},
	methods: {
		goBack() { uni.navigateBack(); },
		async fetchData() {
			this.loading = true;
			this.error = '';
			try {
				const res = await getInternshipSignInPlan();
				const result = res && res.result;
				if (result) {
					this.planData = result;
				} else {
					this.planData = null;
				}
			} catch (e) {
				console.error('获取实习签到计划失败', e);
				this.error = e.message || '获取实习签到计划失败';
			} finally {
				this.loading = false;
				this.refreshing = false;
			}
		},
		async onRefresh() {
			this.refreshing = true;
			await this.fetchData();
		}
	}
}
</script>

<style lang="scss" scoped>
@import url('../../../static/css/home-page.css');

.signin-page { width: 100%; height: 100vh; display: flex; flex-direction: column; background-color: var(--bg-secondary); }
.nav-bar { background: linear-gradient(135deg, #ef4444, #f97316); flex-shrink: 0; }
.nav-bar__content { display: flex; align-items: center; justify-content: space-between; height: 88rpx; padding: 0 24rpx; }
.nav-bar__back, .nav-bar__placeholder { width: 60rpx; height: 60rpx; display: flex; align-items: center; justify-content: flex-start; }
.nav-bar__title { font-size: 32rpx; font-weight: 700; color: #fff; text-align: center; }
.content-scroll { flex: 1; height: 0; }
.state-container { display: flex; flex-direction: column; align-items: center; padding: 120rpx 0; gap: 16rpx; }
.state-spinner { width: 48rpx; height: 48rpx; border: 4rpx solid rgba(239, 68, 68, 0.15); border-top-color: #ef4444; border-radius: 50%; animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.state-text { font-size: 28rpx; color: var(--text-secondary); font-weight: 600; }
.state-sub { font-size: 24rpx; color: var(--text-tertiary); }
.state-container--error { gap: 12rpx; }
.state-icon-circle { width: 72rpx; height: 72rpx; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 36rpx; font-weight: 800; }
.state-icon-circle--error { background: rgba(239, 68, 68, 0.1); color: #ef4444; }
.state-btn { margin-top: 8rpx; padding: 12rpx 40rpx; background: linear-gradient(135deg, #ef4444, #f97316); color: #fff; font-size: 26rpx; font-weight: 600; border-radius: 32rpx; border: none; }
.plan-container { padding: 16rpx 24rpx; }
.plan-card { background: #fff; border-radius: 24rpx; padding: 28rpx; margin-bottom: 16rpx; box-shadow: 0 2rpx 12rpx rgba(239, 68, 68, 0.06); border: 1px solid rgba(148, 163, 184, 0.1); }
.plan-card__title { font-size: 30rpx; font-weight: 700; color: var(--text-primary); margin-bottom: 20rpx; }
.plan-card__row { display: flex; align-items: center; margin-bottom: 12rpx; }
.plan-card__label { font-size: 26rpx; color: var(--text-secondary); flex-shrink: 0; }
.plan-card__value { font-size: 26rpx; color: var(--text-primary); font-weight: 500; }
.records-section { margin-top: 8rpx; }
.records-section__title { font-size: 28rpx; font-weight: 700; color: var(--text-primary); padding: 8rpx 0 16rpx; }
.record-card { background: #fff; border-radius: 20rpx; padding: 24rpx; margin-bottom: 12rpx; box-shadow: 0 2rpx 8rpx rgba(239, 68, 68, 0.04); border: 1px solid rgba(148, 163, 184, 0.1); }
.record-card__date { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12rpx; }
.record-card__date-text { font-size: 28rpx; font-weight: 700; color: var(--text-primary); }
.record-card__status { padding: 4rpx 14rpx; border-radius: 14rpx; background: rgba(16, 185, 129, 0.1); }
.record-card__status-text { font-size: 22rpx; font-weight: 600; color: #10b981; }
.record-card__info { display: flex; align-items: center; margin-top: 8rpx; }
.record-card__info-text { font-size: 24rpx; color: var(--text-secondary); }
.records-empty { display: flex; flex-direction: column; align-items: center; padding: 60rpx 0; gap: 12rpx; }
.records-empty__text { font-size: 26rpx; color: var(--text-tertiary); }
</style>
