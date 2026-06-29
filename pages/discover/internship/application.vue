<template>
	<view class="application-page">
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">我的申请</text>
				<view class="nav-bar__placeholder"></view>
			</view>
		</view>

		<scroll-view class="content-scroll" scroll-y @scrolltolower="loadMore"
			:refresher-enabled="true" @refresherrefresh="onRefresh" :refresher-triggered="refreshing">
			<view v-if="loading && records.length === 0" class="state-container">
				<view class="state-spinner"></view>
				<text class="state-text">加载中...</text>
			</view>

			<view v-else-if="!loading && records.length === 0" class="state-container">
				<view class="state-icon-wrap">
					<l-icon name="no-result" style="font-size: 48px; color: #94a3b8;"></l-icon>
				</view>
				<text class="state-text">暂无实习申请</text>
			</view>

			<view v-else-if="error" class="state-container state-container--error">
				<view class="state-icon-circle state-icon-circle--error">!</view>
				<text class="state-text">{{ error }}</text>
				<button class="state-btn" @tap="fetchData">重试</button>
			</view>

			<view v-else class="list-container">
				<view v-for="(item, index) in records" :key="index" class="app-card">
					<view class="app-card__header">
						<text class="app-card__title">{{ item.positionName || item.jobTitle || '实习申请' }}</text>
						<view class="app-card__status" :class="getStatusClass(item.status)">
							<text class="app-card__status-text">{{ item.statusText || item.status || '待审核' }}</text>
						</view>
					</view>
					<view class="app-card__body">
						<view class="app-card__row" v-if="item.companyName">
							<l-icon name="domain" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
							<text class="app-card__info">{{ item.companyName }}</text>
						</view>
						<view class="app-card__row" v-if="item.applyDate || item.createTime">
							<l-icon name="calendar" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
							<text class="app-card__info">申请时间: {{ item.applyDate || item.createTime }}</text>
						</view>
						<view class="app-card__row" v-if="item.startDate || item.internshipStartDate">
							<l-icon name="clock" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
							<text class="app-card__info">实习时间: {{ item.startDate || item.internshipStartDate }} ~ {{ item.endDate || item.internshipEndDate }}</text>
						</view>
					</view>
				</view>

				<view class="load-more" v-if="records.length > 0">
					<view v-if="loadingMore" class="state-spinner state-spinner--small"></view>
					<text v-else-if="noMore" class="load-more__text">没有更多了</text>
					<text v-else class="load-more__text" @tap="loadMore">加载更多</text>
				</view>
			</view>
		</scroll-view>
	</view>
</template>

<script>
import { getInternshipApplicationList } from '../../api/discover.js';

export default {
	data() {
		return {
			statusBarHeight: 20,
			records: [],
			total: 0,
			pageNo: 1,
			pageSize: 15,
			loading: false,
			loadingMore: false,
			refreshing: false,
			noMore: false,
			error: ''
		}
	},
	onLoad() {
		const systemInfo = uni.getSystemInfoSync();
		this.statusBarHeight = systemInfo.statusBarHeight || 20;
		this.fetchData();
	},
	methods: {
		goBack() { uni.navigateBack(); },
		async fetchData(reset = false) {
			if (this.loading) return;
			if (reset) { this.pageNo = 1; this.noMore = false; }
			this.loading = true;
			this.error = '';
			try {
				const res = await getInternshipApplicationList({ pageNo: this.pageNo, pageSize: this.pageSize });
				const result = res && res.result;
				if (result && result.records) {
					const list = result.records;
					if (reset) { this.records = list; } else { this.records = [...this.records, ...list]; }
					this.total = result.total || 0;
					this.noMore = this.pageNo >= (result.pages || 1);
				} else {
					this.records = [];
					this.total = 0;
				}
			} catch (e) {
				console.error('获取实习申请失败', e);
				this.error = e.message || '获取实习申请失败';
			} finally {
				this.loading = false;
				this.refreshing = false;
			}
		},
		async onRefresh() { this.refreshing = true; await this.fetchData(true); },
		async loadMore() {
			if (this.loadingMore || this.noMore) return;
			this.loadingMore = true;
			this.pageNo++;
			await this.fetchData(false);
			this.loadingMore = false;
		},
		getStatusClass(status) {
			const s = status || '';
			if (s.includes('通过') || s.includes('成功') || s.includes('录用')) return 'app-card__status--pass';
			if (s.includes('拒绝') || s.includes('失败')) return 'app-card__status--fail';
			if (s.includes('待') || s.includes('审核') || s.includes('中')) return 'app-card__status--pending';
			return 'app-card__status--default';
		}
	}
}
</script>

<style lang="scss" scoped>
@import url('../../../static/css/home-page.css');

.application-page { width: 100%; height: 100vh; display: flex; flex-direction: column; background-color: var(--bg-secondary); }
.nav-bar { background: linear-gradient(135deg, #f59e0b, #fbbf24); flex-shrink: 0; }
.nav-bar__content { display: flex; align-items: center; justify-content: space-between; height: 88rpx; padding: 0 24rpx; }
.nav-bar__back, .nav-bar__placeholder { width: 60rpx; height: 60rpx; display: flex; align-items: center; justify-content: flex-start; }
.nav-bar__title { font-size: 32rpx; font-weight: 700; color: #fff; text-align: center; }
.content-scroll { flex: 1; height: 0; }
.state-container { display: flex; flex-direction: column; align-items: center; padding: 120rpx 0; gap: 16rpx; }
.state-spinner { width: 48rpx; height: 48rpx; border: 4rpx solid rgba(245, 158, 11, 0.15); border-top-color: #f59e0b; border-radius: 50%; animation: spin 0.8s linear infinite; }
.state-spinner--small { width: 32rpx; height: 32rpx; border-width: 3rpx; }
@keyframes spin { to { transform: rotate(360deg); } }
.state-text { font-size: 28rpx; color: var(--text-secondary); font-weight: 600; }
.state-container--error { gap: 12rpx; }
.state-icon-circle { width: 72rpx; height: 72rpx; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 36rpx; font-weight: 800; }
.state-icon-circle--error { background: rgba(245, 158, 11, 0.1); color: #f59e0b; }
.state-btn { margin-top: 8rpx; padding: 12rpx 40rpx; background: linear-gradient(135deg, #f59e0b, #fbbf24); color: #fff; font-size: 26rpx; font-weight: 600; border-radius: 32rpx; border: none; }
.list-container { padding: 16rpx 24rpx; }
.app-card { background: #fff; border-radius: 24rpx; padding: 28rpx; margin-bottom: 16rpx; box-shadow: 0 2rpx 12rpx rgba(245, 158, 11, 0.06); border: 1px solid rgba(148, 163, 184, 0.1); }
.app-card__header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16rpx; }
.app-card__title { font-size: 30rpx; font-weight: 700; color: var(--text-primary); flex: 1; margin-right: 16rpx; }
.app-card__status { padding: 6rpx 16rpx; border-radius: 16rpx; flex-shrink: 0; }
.app-card__status--pass { background: rgba(16, 185, 129, 0.1); }
.app-card__status--pass .app-card__status-text { color: #10b981; }
.app-card__status--fail { background: rgba(239, 68, 68, 0.1); }
.app-card__status--fail .app-card__status-text { color: #ef4444; }
.app-card__status--pending { background: rgba(245, 158, 11, 0.1); }
.app-card__status--pending .app-card__status-text { color: #f59e0b; }
.app-card__status--default { background: rgba(148, 163, 184, 0.1); }
.app-card__status--default .app-card__status-text { color: #94a3b8; }
.app-card__status-text { font-size: 22rpx; font-weight: 600; }
.app-card__body { display: flex; flex-direction: column; gap: 8rpx; }
.app-card__row { display: flex; align-items: center; }
.app-card__info { font-size: 24rpx; color: var(--text-secondary); }
.load-more { display: flex; justify-content: center; align-items: center; padding: 32rpx 0; }
.load-more__text { font-size: 26rpx; color: #f59e0b; font-weight: 600; }
</style>
