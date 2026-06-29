<template>
	<view class="news-list-page">
		<cache-banner
			:visible="showCacheBanner"
			:cache-updated-at="cacheUpdatedAt"
			@close="showCacheBanner = false"
		></cache-banner>

		<!-- 顶部蓝色渐变 Hero -->
		<view class="news-hero">
			<view class="news-hero-bg"></view>
			<view class="news-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<view class="news-hero-nav">
				<view class="news-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="news-hero-title">新闻资讯</text>
				<view class="news-refresh-btn" @tap="refreshData">
					<l-icon name="refresh" style="font-size: 20px; color: #fff;"></l-icon>
				</view>
			</view>

			<view class="news-hero-content">
				<text class="news-hero-sub">NEWS CENTER</text>
			</view>
		</view>

		<!-- 内容区域 -->
		<view class="news-content">
			<!-- 新闻类型选择 -->
			<view class="news-type-bar" v-if="newsTypes.length > 0">
				<scroll-view class="news-type-scroll" scroll-x show-scrollbar="false">
					<view
						v-for="type in newsTypes"
						:key="type.id"
						class="news-type-tab"
					:class="{ 'news-type-tab--active': selectedTypeId === type.id }"
					@tap="selectType(type.id)"
				>
					<text class="news-type-tab-text">{{ type.name }}</text>
				</view>
				</scroll-view>
			</view>

			<!-- 列表 -->
			<scroll-view class="news-scroll" scroll-y @scrolltolower="loadMore" @scrolltoupper="onScrollTop">
				<view class="news-list-container">
					<!-- 加载状态 -->
					<view v-if="loading && currentPage === 1" class="news-state-card">
						<view class="news-state-spinner"></view>
						<text class="news-state-text">加载中...</text>
					</view>

					<!-- 错误状态 -->
					<view v-else-if="error" class="news-state-card news-state-card--error">
						<l-icon name="error-circle" style="font-size: 48px; color: var(--error-color);"></l-icon>
						<text class="news-state-title">{{ error }}</text>
						<view class="news-retry-btn" @tap="loadNewsList">重试</view>
					</view>

					<!-- 空状态 -->
					<view v-else-if="newsList.length === 0" class="news-empty">
						<view class="news-empty-icon">
							<l-icon name="file-text" style="font-size: 64px; color: var(--text-tertiary);"></l-icon>
						</view>
						<text class="news-empty-title">暂无新闻</text>
						<text class="news-empty-sub">当前分类下没有新闻内容</text>
					</view>

					<!-- 新闻列表 -->
					<view v-else class="news-card-list">
						<view
							v-for="news in newsList"
							:key="news.id"
							class="news-card"
							@tap="viewNewsDetail(news.id)"
						>
							<view class="news-card-indicator" :class="{ 'news-card-indicator--top': news.isTop }"></view>
							<view class="news-card-body">
								<view class="news-card-header">
									<view class="news-card-top-tag" v-if="news.isTop">
										<text class="news-card-top-tag-text">置顶</text>
									</view>
									<text class="news-card-title">{{ news.title }}</text>
								</view>
								<text v-if="news.summary" class="news-card-summary">{{ news.summary }}</text>
								<view class="news-card-footer">
									<text class="news-card-date">{{ formatDate(news.publishTime || news.createTime) }}</text>
									<text class="news-card-source">{{ news.source || news.author || '系统资讯' }}</text>
									<text class="news-card-views" v-if="news.viewCount">{{ news.viewCount }} 阅</text>
								</view>
							</view>
						</view>
					</view>

					<!-- 加载更多 -->
					<view v-if="hasMore && !loading && newsList.length > 0" class="news-load-more" @tap="loadMore">
						<text class="news-load-more-text">加载更多</text>
					</view>
					<view v-if="hasMore && loading && currentPage > 1" class="news-load-more">
						<view class="news-load-more-spinner"></view>
						<text class="news-load-more-text">加载中...</text>
					</view>

					<!-- 到底了 -->
					<view v-if="!hasMore && newsList.length > 0" class="news-load-more">
						<text class="news-load-more-text">已显示全部资讯</text>
					</view>
				</view>
			</scroll-view>
		</view>
	</view>
</template>

<script>
import { getNewsTypeList, getNewsListByTypeId } from '../api/news.js';
import { showToast, navigateTo, navigateBack } from '../../pages/api/page.js';
import CacheBanner from '../../components/CacheBanner.vue';

export default {
	components: {
		CacheBanner
	},
	data() {
		return {
			newsTypes: [],
			selectedTypeId: null,
			newsList: [],
			loading: false,
			error: '',
			currentPage: 1,
			pageSize: 10,
			hasMore: true,
			showCacheBanner: false,
			cacheUpdatedAt: '',
			statusBarHeight: 20,
			navPaddingTop: '0px'
		};
	},
	onLoad() {
		this.initPage();
		this.initStatusBar();
	},
	onPullDownRefresh() {
		this.refreshData();
	},
	onReachBottom() {
		if (this.hasMore && !this.loading) {
			this.loadMore();
		}
	},
	methods: {
		initStatusBar() {
			try {
				const systemInfo = uni.getSystemInfoSync();
				this.statusBarHeight = systemInfo.statusBarHeight || 20;
			} catch (e) {
				this.statusBarHeight = 20;
			}
		},
		applyCacheMeta(meta) {
			if (!meta) return;
			if (meta.fromCache) {
				this.showCacheBanner = true;
				this.cacheUpdatedAt = meta.cacheUpdatedAt || '';
			}
		},
		handleNavHeightReady(navInfo) {
			this.navPaddingTop = navInfo.heightPx;
		},
		async initPage() {
			try {
				await this.loadNewsTypes();
				await this.loadNewsList();
			} catch (error) {
				console.error('初始化页面失败:', error);
				this.error = '页面加载失败，请稍后重试';
			}
		},
		async loadNewsTypes() {
			try {
				const resp = await getNewsTypeList();
				this.applyCacheMeta(resp.meta);
				const types = resp.data || [];
				this.newsTypes = types;
				if (types.length > 0) {
					const noticeType = types.find(type =>
						type.name && (type.name.includes('通知') || type.name.includes('公告'))
					);
					this.selectedTypeId = noticeType ? noticeType.id : types[0].id;
				}
			} catch (error) {
				console.error('获取新闻类型失败:', error);
			}
		},
		selectType(typeId) {
			if (this.selectedTypeId === typeId) return;
			this.selectedTypeId = typeId;
			this.currentPage = 1;
			this.hasMore = true;
			this.loadNewsList();
		},
		async loadNewsList() {
			if (this.loading) return;
			this.loading = true;
			this.error = '';
			try {
				const result = await getNewsListByTypeId({
					typeId: this.selectedTypeId,
					pageNo: this.currentPage,
					pageSize: this.pageSize
				});
				this.applyCacheMeta(result.meta);
				if (this.currentPage === 1) {
					this.newsList = result.records;
				} else {
					this.newsList = [...this.newsList, ...result.records];
				}
				this.hasMore = result.records.length === this.pageSize;
			} catch (error) {
				console.error('获取新闻列表失败:', error);
				this.error = '获取新闻失败，请稍后重试';
				showToast({ title: '获取新闻失败', icon: 'none' });
			} finally {
				this.loading = false;
			}
		},
		loadMore() {
			this.currentPage++;
			this.loadNewsList();
		},
		onScrollTop() {
			uni.pageScrollTo({ scrollTop: 0, duration: 300 });
		},
		async refreshData() {
			this.currentPage = 1;
			this.hasMore = true;
			await this.loadNewsList();
			uni.stopPullDownRefresh();
		},
		viewNewsDetail(newsId) {
			navigateTo({ url: `/pages/news/detail?id=${newsId}` });
		},
		formatDate(dateStr) {
			if (!dateStr) return '';
			try {
				const date = new Date(dateStr);
				const now = new Date();
				const diffTime = now - date;
				const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));
				if (diffDays === 0) return '今天';
				else if (diffDays === 1) return '昨天';
				else if (diffDays < 7) return `${diffDays}天前`;
				return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
			} catch (error) {
				return dateStr;
			}
		},
		goBack() {
			navigateBack();
		}
	}
};
</script>

<style lang="scss" scoped>
/* ============================================
   News List - Hero Style
   ============================================ */

.news-list-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
}

/* ---- Hero Section ---- */
.news-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 200rpx;
}

.news-hero-bg {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		#1e3a8a 0%,
		#1e40af 25%,
		#2563eb 55%,
		#3b82f6 75%,
		#93c5fd 100%);
	z-index: 0;
}

.news-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.65) 0%,
		rgba(37, 99, 235, 0.4) 50%,
		rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.news-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.news-back-btn {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255, 255, 255, 0.18);
	border-radius: 50%;
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.25);

	&:active {
		background: rgba(255, 255, 255, 0.28);
		transform: scale(0.95);
	}
}

.news-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
}

.news-refresh-btn {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255, 255, 255, 0.18);
	border-radius: 50%;
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.25);

	&:active {
		background: rgba(255, 255, 255, 0.28);
		transform: scale(0.9);
	}
}

.news-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 8rpx;
}

.news-hero-sub {
	font-size: 20rpx;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.6);
	letter-spacing: 2px;
}

/* ---- Content ---- */
.news-content {
	flex: 1;
	min-height: 0;
	display: flex;
	flex-direction: column;
}

/* ---- Type Bar ---- */
.news-type-bar {
	background: #fff;
	margin: -24rpx 24rpx 0;
	border-radius: 24rpx;
	padding: 8rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.12);
	position: relative;
	z-index: 2;
}

.news-type-scroll {
	white-space: nowrap;
}

.news-type-tab {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	height: 72rpx;
	padding: 0 28rpx;
	border-radius: 16rpx;
	margin: 0 4rpx;
	transition: all 0.2s ease;

	.news-type-tab-text {
		font-size: 26rpx;
		font-weight: 600;
		color: var(--text-secondary);
		white-space: nowrap;
	}

	&--active {
		background: linear-gradient(135deg, var(--primary-600), var(--primary-700));

		.news-type-tab-text { color: #fff; }
	}

	&:active:not(.news-type-tab--active) {
		background: var(--bg-muted);
	}
}

/* ---- Scroll ---- */
.news-scroll {
	flex: 1;
	min-height: 0;
}

.news-list-container {
	padding: 24rpx;
}

/* ---- State Card ---- */
.news-state-card {
	background: #fff;
	border-radius: 28rpx;
	padding: 80rpx 32rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 16rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.08);
	border: 1px solid rgba(148, 163, 184, 0.12);
}

.news-state-card--error {
	gap: 20rpx;
}

.news-state-spinner {
	width: 64rpx;
	height: 64rpx;
	border: 5rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: news-spin 0.8s linear infinite;
}

@keyframes news-spin {
	to { transform: rotate(360deg); }
}

.news-state-text {
	font-size: 28rpx;
	color: var(--text-secondary);
}

.news-state-title {
	font-size: 28rpx;
	color: var(--text-secondary);
	text-align: center;
}

.news-retry-btn {
	height: 72rpx;
	padding: 0 48rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 36rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 8rpx 24rpx rgba(37, 99, 235, 0.3);

	.news-load-more-text {
		font-size: 28rpx;
		font-weight: 700;
		color: #fff;
	}

	&:active { transform: scale(0.95); }
}

/* ---- Empty State ---- */
.news-empty {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 80rpx 40rpx;
	gap: 16rpx;
}

.news-empty-icon {
	width: 160rpx;
	height: 160rpx;
	border-radius: 50%;
	background: var(--bg-muted);
	display: flex;
	align-items: center;
	justify-content: center;
}

.news-empty-title {
	font-size: 32rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.news-empty-sub {
	font-size: 26rpx;
	color: var(--text-tertiary);
}

/* ---- News Card ---- */
.news-card-list {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
}

.news-card {
	display: flex;
	background: #fff;
	border-radius: 24rpx;
	overflow: hidden;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	transition: all 0.2s var(--ease-out);

	&:active {
		transform: scale(0.99);
		background: #f8fafc;
	}
}

.news-card-indicator {
	width: 8rpx;
	flex-shrink: 0;
	background: var(--primary-300);

	&--top {
		background: linear-gradient(180deg, var(--error-color), #f87171);
	}
}

.news-card-body {
	flex: 1;
	padding: 24rpx;
}

.news-card-header {
	display: flex;
	align-items: flex-start;
	gap: 12rpx;
	margin-bottom: 12rpx;
}

.news-card-top-tag {
	flex-shrink: 0;
	padding: 4rpx 12rpx;
	border-radius: 8rpx;
	background: rgba(239, 68, 68, 0.1);
	border: 1px solid rgba(239, 68, 68, 0.3);

	.news-card-top-tag-text {
		font-size: 18rpx;
		font-weight: 700;
		color: var(--error-color);
	}
}

.news-card-title {
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	line-height: 1.4;
	flex: 1;
}

.news-card-summary {
	display: block;
	font-size: 26rpx;
	color: var(--text-secondary);
	line-height: 1.6;
	display: -webkit-box;
	-webkit-line-clamp: 2;
	-webkit-box-orient: vertical;
	overflow: hidden;
	margin-bottom: 16rpx;
	line-clamp: 2;
}

.news-card-footer {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.news-card-date,
.news-card-source,
.news-card-views {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

/* ---- Load More ---- */
.news-load-more {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 12rpx;
	padding: 32rpx 0;

	.news-load-more-text {
		font-size: 26rpx;
		color: var(--primary-500);
		font-weight: 600;
	}
}

.news-load-more-spinner {
	width: 32rpx;
	height: 32rpx;
	border: 3rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: news-spin 0.8s linear infinite;
}
</style>
