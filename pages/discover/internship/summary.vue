<template>
	<view class="summary-page">
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">我的实习总结</text>
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
				<text class="state-text">暂无实习总结</text>
			</view>

			<view v-else-if="error" class="state-container state-container--error">
				<view class="state-icon-circle state-icon-circle--error">!</view>
				<text class="state-text">{{ error }}</text>
				<button class="state-btn" @tap="fetchData">重试</button>
			</view>

			<view v-else class="list-container">
				<view v-for="(item, index) in records" :key="index" class="summary-card">
					<view class="summary-card__header">
						<text class="summary-card__title">{{ item.title || item.summaryTitle || '实习总结' }}</text>
					</view>
					<view class="summary-card__body">
						<view class="summary-card__row" v-if="item.content || item.summaryContent">
							<text class="summary-card__content">{{ item.content || item.summaryContent }}</text>
						</view>
						<view class="summary-card__row" v-if="item.companyName">
							<l-icon name="domain" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
							<text class="summary-card__info">{{ item.companyName }}</text>
						</view>
						<view class="summary-card__row" v-if="item.semester">
							<l-icon name="school" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
							<text class="summary-card__info">{{ item.semester }}</text>
						</view>
						<view class="summary-card__row" v-if="item.createTime || item.submitTime">
							<l-icon name="calendar" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
							<text class="summary-card__info">提交时间: {{ item.createTime || item.submitTime }}</text>
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
import { getMyInternshipSummaryList } from '../../api/discover.js';

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
				const res = await getMyInternshipSummaryList({ pageNo: this.pageNo, pageSize: this.pageSize });
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
				console.error('获取实习总结失败', e);
				this.error = e.message || '获取实习总结失败';
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
		}
	}
}
</script>

<style lang="scss" scoped>
@import url('../../../static/css/home-page.css');

.summary-page { width: 100%; height: 100vh; display: flex; flex-direction: column; background-color: var(--bg-secondary); }
.nav-bar { background: linear-gradient(135deg, #06b6d4, #22d3ee); flex-shrink: 0; }
.nav-bar__content { display: flex; align-items: center; justify-content: space-between; height: 88rpx; padding: 0 24rpx; }
.nav-bar__back, .nav-bar__placeholder { width: 60rpx; height: 60rpx; display: flex; align-items: center; justify-content: flex-start; }
.nav-bar__title { font-size: 32rpx; font-weight: 700; color: #fff; text-align: center; }
.content-scroll { flex: 1; height: 0; }
.state-container { display: flex; flex-direction: column; align-items: center; padding: 120rpx 0; gap: 16rpx; }
.state-spinner { width: 48rpx; height: 48rpx; border: 4rpx solid rgba(6, 182, 212, 0.15); border-top-color: #06b6d4; border-radius: 50%; animation: spin 0.8s linear infinite; }
.state-spinner--small { width: 32rpx; height: 32rpx; border-width: 3rpx; }
@keyframes spin { to { transform: rotate(360deg); } }
.state-text { font-size: 28rpx; color: var(--text-secondary); font-weight: 600; }
.state-container--error { gap: 12rpx; }
.state-icon-circle { width: 72rpx; height: 72rpx; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 36rpx; font-weight: 800; }
.state-icon-circle--error { background: rgba(6, 182, 212, 0.1); color: #06b6d4; }
.state-btn { margin-top: 8rpx; padding: 12rpx 40rpx; background: linear-gradient(135deg, #06b6d4, #22d3ee); color: #fff; font-size: 26rpx; font-weight: 600; border-radius: 32rpx; border: none; }
.list-container { padding: 16rpx 24rpx; }
.summary-card { background: #fff; border-radius: 24rpx; padding: 28rpx; margin-bottom: 16rpx; box-shadow: 0 2rpx 12rpx rgba(6, 182, 212, 0.06); border: 1px solid rgba(148, 163, 184, 0.1); }
.summary-card__header { margin-bottom: 12rpx; }
.summary-card__title { font-size: 30rpx; font-weight: 700; color: var(--text-primary); }
.summary-card__body { display: flex; flex-direction: column; gap: 8rpx; }
.summary-card__row { display: flex; align-items: flex-start; }
.summary-card__content { font-size: 24rpx; color: var(--text-secondary); line-height: 1.6; }
.summary-card__info { font-size: 24rpx; color: var(--text-secondary); }
.load-more { display: flex; justify-content: center; align-items: center; padding: 32rpx 0; }
.load-more__text { font-size: 26rpx; color: #06b6d4; font-weight: 600; }
</style>
