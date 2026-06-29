<template>
	<view class="miss-class-page">
		<!-- 自定义导航栏 -->
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">缺课统计</text>
				<view class="nav-bar__placeholder"></view>
			</view>
		</view>

		<!-- 搜索栏 -->
		<view class="search-bar">
			<view class="search-bar__input">
				<l-icon name="search" style="font-size: 18px; color: #94a3b8; margin-right: 8rpx;"></l-icon>
				<input
					class="search-bar__field"
					placeholder="搜索课程名称"
					v-model="searchKeyword"
					@confirm="handleSearch"
					confirm-type="search"
					placeholder-class="search-bar__placeholder"
				/>
			</view>
			<button class="search-bar__btn" @tap="handleSearch">搜索</button>
		</view>

		<!-- 筛选信息栏 -->
		<view class="filter-bar">
			<view class="filter-bar__info">
				<text class="filter-bar__text">{{ semesterFilter || '全部学期' }}</text>
			</view>
			<view class="filter-bar__total" v-if="!loading && records.length > 0">
				<text class="filter-bar__count">共 {{ total }} 条记录</text>
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
				<view class="state-icon-wrap">
					<l-icon name="no-result" style="font-size: 48px; color: #94a3b8;"></l-icon>
				</view>
				<text class="state-text">{{ searchKeyword ? '未找到匹配的缺课记录' : '暂无缺课记录' }}</text>
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
					class="miss-card"
				>
					<view class="miss-card__header">
						<text class="miss-card__course">{{ item.courseName }}</text>
						<view class="miss-card__ratio">
							<text class="miss-card__ratio-text">{{ item.missClass }} 节</text>
						</view>
					</view>
					<view class="miss-card__body">
						<view class="miss-card__row">
							<l-icon name="user" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
							<text class="miss-card__info">{{ item.teacherNames }}</text>
						</view>
						<view class="miss-card__row">
							<l-icon name="school" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
							<text class="miss-card__info">{{ item.teachingClassName }} | 总课时 {{ item.totalLessonScope }} 节</text>
						</view>
						<view class="miss-card__row">
							<l-icon name="calendar" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
							<text class="miss-card__info">{{ item.currentSemester }}</text>
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

		<!-- PC验证码弹窗 -->
		<PCCaptchaModal
			:visible="captchaModalVisible"
			:sessionId="captchaSessionId"
			:captchaImage="captchaImage"
			:tips="'登录已过期，请输入验证码'"
			@close="onCaptchaClose"
			@success="onCaptchaSuccess"
			@refresh-captcha="onCaptchaRefresh"
		/>
	</view>
</template>

<script>
import { pcGetMissClassList } from '../../pages/api/discover.js';
import PCCaptchaModal from '@/components/PCCaptchaModal.vue';

export default {
	components: { PCCaptchaModal },
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
			error: '',
			searchKeyword: '',
			semesterFilter: '',
			captchaModalVisible: false,
			captchaSessionId: '',
			captchaImage: '',
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
		onCaptchaClose() {
			this.captchaModalVisible = false;
		},
		async onCaptchaSuccess() {
			this.captchaModalVisible = false;
			await this.fetchData(true);
		},
		async onCaptchaRefresh() {
			this.captchaModalVisible = false;
			uni.navigateTo({ url: '/pages/discover/pc-login?redirect=/pages/discover/miss-class' });
		},
		async fetchData(reset = false) {
			if (this.loading) return;
			if (reset) {
				this.pageNo = 1;
				this.noMore = false;
			}
			this.loading = true;
			this.error = '';
			try {
				const params = {
					pageNum: this.pageNo,
					pageSize: this.pageSize,
					semester: this.semesterFilter,
					gradeName: ''
				};
			const res = await pcGetMissClassList(params);
			const result = res && res.result;
			if (result && result.needManual) {
				this.captchaModalVisible = true;
				this.captchaSessionId = result.sessionId || '';
				this.captchaImage = result.captcha || '';
				this.error = '';
				this.records = [];
				this.total = 0;
				return;
			}
			if (result && result.list) {
					const list = result.list;
					// 前端搜索过滤
					const filtered = this.searchKeyword
						? list.filter(item => (item.courseName || '').includes(this.searchKeyword))
						: list;
					if (reset) {
						this.records = filtered;
					} else {
						this.records = [...this.records, ...filtered];
					}
					this.total = result.total || 0;
					this.noMore = this.records.length >= this.total || (result.pages && this.pageNo >= result.pages);
				} else {
					this.records = [];
					this.total = 0;
				}
			} catch (e) {
				console.error('获取缺课统计失败', e);
				this.error = e.message || '获取缺课统计失败';
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
			this.pageNo++;
			await this.fetchData(false);
			this.loadingMore = false;
		},
		handleSearch() {
			this.fetchData(true);
		}
	}
}
</script>

<style lang="scss" scoped>
@import url('../../static/css/home-page.css');

.miss-class-page {
	width: 100%;
	height: 100vh;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
}

.nav-bar {
	background: linear-gradient(135deg, #ef4444, #f97316);
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
	justify-content: flex-start;
}

.nav-bar__title {
	font-size: 32rpx;
	font-weight: 700;
	color: #fff;
	text-align: center;
}

.search-bar {
	display: flex;
	align-items: center;
	padding: 16rpx 24rpx;
	background: #fff;
	gap: 16rpx;
}

.search-bar__input {
	flex: 1;
	display: flex;
	align-items: center;
	height: 72rpx;
	padding: 0 24rpx;
	background: var(--bg-secondary);
	border-radius: 36rpx;
}

.search-bar__field {
	flex: 1;
	font-size: 28rpx;
	height: 72rpx;
}

.search-bar__placeholder {
	color: #94a3b8;
}

.search-bar__btn {
	height: 64rpx;
	padding: 0 32rpx;
	background: linear-gradient(135deg, #ef4444, #f97316);
	color: #fff;
	font-size: 26rpx;
	font-weight: 600;
	border-radius: 32rpx;
	border: none;
	display: flex;
	align-items: center;
	justify-content: center;
}

.filter-bar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 16rpx 24rpx;
	background: #fff;
	border-bottom: 1px solid rgba(226, 232, 240, 0.7);
}

.filter-bar__text {
	font-size: 26rpx;
	color: var(--text-secondary);
}

.filter-bar__count {
	font-size: 24rpx;
	color: var(--text-tertiary);
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
	border: 4rpx solid rgba(239, 68, 68, 0.15);
	border-top-color: #ef4444;
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
	background: linear-gradient(135deg, #ef4444, #f97316);
	color: #fff;
	font-size: 26rpx;
	font-weight: 600;
	border-radius: 32rpx;
	border: none;
}

.list-container {
	padding: 16rpx 24rpx;
}

.miss-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 28rpx;
	margin-bottom: 16rpx;
	box-shadow: 0 2rpx 12rpx rgba(239, 68, 68, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.miss-card__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 16rpx;
}

.miss-card__course {
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	flex: 1;
	margin-right: 16rpx;
}

.miss-card__ratio {
	padding: 6rpx 16rpx;
	border-radius: 16rpx;
	background: rgba(239, 68, 68, 0.1);
	flex-shrink: 0;
}

.miss-card__ratio-text {
	font-size: 22rpx;
	font-weight: 600;
	color: #ef4444;
}

.miss-card__body {
	display: flex;
	flex-direction: column;
	gap: 8rpx;
}

.miss-card__row {
	display: flex;
	align-items: center;
}

.miss-card__info {
	font-size: 24rpx;
	color: var(--text-secondary);
}

.load-more {
	display: flex;
	justify-content: center;
	align-items: center;
	padding: 32rpx 0;
}

.load-more__text {
	font-size: 26rpx;
	color: #ef4444;
	font-weight: 600;
}
</style>
