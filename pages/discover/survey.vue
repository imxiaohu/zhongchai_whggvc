<template>
	<view class="survey-page">
		<!-- 自定义导航栏 -->
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">调查问卷</text>
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
					<l-icon name="no-result" style="font-size: 48px; color: #94a3b8;"></l-icon>
				</view>
				<text class="state-text">暂无问卷</text>
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
					class="survey-card"
					@tap="goToDetail(item)"
				>
					<view class="survey-card__header">
						<text class="survey-card__title">{{ item.title }}</text>
						<view class="survey-card__badge" :class="item.iswrite === 1 ? 'survey-card__badge--done' : 'survey-card__badge--todo'">
							<text class="survey-card__badge-text">{{ item.iswrite === 1 ? '已填写' : '待填写' }}</text>
						</view>
					</view>
					<view class="survey-card__desc" v-if="item.description && item.description.trim()">
						<text class="survey-card__desc-text">{{ item.description }}</text>
					</view>
					<view class="survey-card__footer">
						<view class="survey-card__time">
							<l-icon name="time" style="font-size: 12px; color: #94a3b8; margin-right: 4rpx;"></l-icon>
							<text class="survey-card__time-text">{{ item.starttime }} 至 {{ item.endtime }}</text>
						</view>
						<view class="survey-card__creator" v-if="item.creatorname">
							<l-icon name="user" style="font-size: 12px; color: #94a3b8; margin-right: 4rpx;"></l-icon>
							<text class="survey-card__creator-text">{{ item.creatorname }}</text>
						</view>
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
	</view>
</template>

<script>
import { getSurveyList } from '../../pages/api/discover.js';
import { navigateTo } from '../../pages/api/page.js';

export default {
	data() {
		return {
			statusBarHeight: 20,
			records: [],
			total: 0,
			page: 1,
			pageSize: 10,
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
		goBack() {
			uni.navigateBack();
		},
		async fetchData(reset = false) {
			if (this.loading) return;
			if (reset) {
				this.page = 1;
				this.noMore = false;
			}
			this.loading = true;
			this.error = '';
			try {
				const res = await getSurveyList({
					current: this.page,
					size: this.pageSize
				});
				const result = res && res.result;
				if (result && result.records) {
					if (reset) {
						this.records = result.records;
					} else {
						this.records = [...this.records, ...result.records];
					}
					this.total = result.total || 0;
					this.noMore = this.records.length >= this.total;
				} else {
					this.records = [];
					this.total = 0;
				}
			} catch (e) {
				console.error('获取问卷列表失败', e);
				this.error = e.message || '获取问卷列表失败';
			} finally {
				this.loading = false;
				this.refreshing = false;
			}
		},
		async onRefresh() {
			this.refreshing = true;
			await this.fetchData(true);
		},
		async loadMore() {
			if (this.loadingMore || this.noMore) return;
			this.loadingMore = true;
			this.page++;
			await this.fetchData(false);
			this.loadingMore = false;
		},
		goToDetail(item) {
			navigateTo({ url: `/pages/discover/survey-detail?id=${item.id}&title=${encodeURIComponent(item.title)}` });
		}
	}
}
</script>

<style lang="scss" scoped>
@import url('../../static/css/home-page.css');

.survey-page {
	width: 100%;
	height: 100vh;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
}

.nav-bar {
	background: linear-gradient(135deg, #8b5cf6, #a78bfa);
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
	border: 4rpx solid rgba(139, 92, 246, 0.15);
	border-top-color: #8b5cf6;
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
	background: linear-gradient(135deg, #8b5cf6, #a78bfa);
	color: #fff;
	font-size: 26rpx;
	font-weight: 600;
	border-radius: 32rpx;
	border: none;
}

.list-container {
	padding: 16rpx 24rpx;
}

.survey-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 28rpx;
	margin-bottom: 16rpx;
	box-shadow: 0 2rpx 12rpx rgba(139, 92, 246, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.survey-card__header {
	display: flex;
	align-items: flex-start;
	justify-content: space-between;
	margin-bottom: 12rpx;
}

.survey-card__title {
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	flex: 1;
	margin-right: 16rpx;
}

.survey-card__badge {
	padding: 4rpx 14rpx;
	border-radius: 16rpx;
	flex-shrink: 0;
}

.survey-card__badge--done {
	background: rgba(16, 185, 129, 0.12);
}

.survey-card__badge--done .survey-card__badge-text {
	color: #10b981;
}

.survey-card__badge--todo {
	background: rgba(245, 158, 11, 0.12);
}

.survey-card__badge--todo .survey-card__badge-text {
	color: #f59e0b;
}

.survey-card__badge-text {
	font-size: 22rpx;
	font-weight: 600;
}

.survey-card__desc {
	margin-bottom: 12rpx;
}

.survey-card__desc-text {
	font-size: 24rpx;
	color: var(--text-secondary);
	line-height: 1.5;
	display: -webkit-box;
	-webkit-line-clamp: 2;
	-webkit-box-orient: vertical;
	overflow: hidden;
}

.survey-card__footer {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.survey-card__time,
.survey-card__creator {
	display: flex;
	align-items: center;
}

.survey-card__time-text,
.survey-card__creator-text {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.load-more {
	display: flex;
	justify-content: center;
	align-items: center;
	padding: 32rpx 0;
}

.load-more__text {
	font-size: 26rpx;
	color: #8b5cf6;
	font-weight: 600;
}
</style>
