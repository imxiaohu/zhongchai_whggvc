<template>
	<view class="attendance-page">
		<!-- 自定义导航栏 -->
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">考勤记录</text>
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

		<!-- 学期周次筛选 -->
		<view class="filter-bar">
			<view class="filter-bar__info">
				<text class="filter-bar__text">{{ currentSemester || '加载中...' }} {{ nowWeek || '' }}</text>
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
				<view class="state-icon-wrap state-icon-wrap--clipboard">
					<l-icon name="clipboard-text" style="font-size: 48px; color: #94a3b8;"></l-icon>
				</view>
				<text class="state-text">{{ searchKeyword ? '未找到匹配的考勤记录' : '暂无考勤记录' }}</text>
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
					class="attendance-card"
				>
					<view class="attendance-card__header">
						<text class="attendance-card__course">{{ item.courseName }}</text>
						<view class="attendance-card__status" :class="'attendance-card__status--' + getStatusClass(item.attendance)">
							<text class="attendance-card__status-text">{{ item.attendance }}</text>
						</view>
					</view>
					<view class="attendance-card__body">
						<view class="attendance-card__row">
							<l-icon name="user" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
							<text class="attendance-card__info">{{ item.teacherNames }}</text>
						</view>
						<view class="attendance-card__row">
							<l-icon name="calendar" style="font-size: 14px; color: #94a3b8; margin-right: 6rpx;"></l-icon>
							<text class="attendance-card__info">{{ item.nowWeek }} {{ item.week }} 第{{ item.lessonScope }}节</text>
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
import { getAttendanceList } from '../../pages/api/discover.js';
import { getMCourseSchoolTimetable } from '../../pages/api/schedule.js';

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
			error: '',
			searchKeyword: '',
			currentSemester: '',
			nowWeek: ''
		}
	},
	onLoad() {
		const systemInfo = uni.getSystemInfoSync();
		this.statusBarHeight = systemInfo.statusBarHeight || 20;
		this.fetchSemesterInfo();
		this.fetchData();
	},
	methods: {
		goBack() {
			uni.navigateBack();
		},
		async fetchSemesterInfo() {
			try {
				const res = await getMCourseSchoolTimetable();
				if (res) {
					this.currentSemester = res.currentSemester || '';
					this.nowWeek = res.nowWeek || res.currentWeek ? `第${res.currentWeek}周` : '';
				}
			} catch (e) {
				console.error('获取学期信息失败', e);
			}
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
					pageNo: this.pageNo,
					pageSize: this.pageSize,
					nowWeek: this.nowWeek,
					currentSemester: this.currentSemester,
					courseName: this.searchKeyword
				};
				const res = await getAttendanceList(params);
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
				console.error('获取考勤记录失败', e);
				this.error = e.message || '获取考勤记录失败';
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
		},
		getStatusClass(status) {
			if (!status) return 'default';
			const s = status.trim();
			if (s === '正常') return 'success';
			if (s === '迟到' || s === '早退') return 'warn';
			if (s === '缺勤' || s === '旷课') return 'danger';
			return 'default';
		}
	}
}
</script>

<style lang="scss" scoped>
@import url('../../static/css/home-page.css');

.attendance-page {
	width: 100%;
	height: 100vh;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
}

.nav-bar {
	background: linear-gradient(135deg, #2563eb, #3b82f6);
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
	background: linear-gradient(135deg, #2563eb, #3b82f6);
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
	border: 4rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
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
	background: linear-gradient(135deg, #2563eb, #3b82f6);
	color: #fff;
	font-size: 26rpx;
	font-weight: 600;
	border-radius: 32rpx;
	border: none;
}

.list-container {
	padding: 16rpx 24rpx;
}

.attendance-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 28rpx;
	margin-bottom: 16rpx;
	box-shadow: 0 2rpx 12rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.attendance-card__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 16rpx;
}

.attendance-card__course {
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	flex: 1;
	margin-right: 16rpx;
}

.attendance-card__status {
	padding: 6rpx 16rpx;
	border-radius: 16rpx;
	flex-shrink: 0;
}

.attendance-card__status--success {
	background: rgba(16, 185, 129, 0.12);
}

.attendance-card__status--success .attendance-card__status-text {
	color: #10b981;
}

.attendance-card__status--warn {
	background: rgba(245, 158, 11, 0.12);
}

.attendance-card__status--warn .attendance-card__status-text {
	color: #f59e0b;
}

.attendance-card__status--danger {
	background: rgba(239, 68, 68, 0.12);
}

.attendance-card__status--danger .attendance-card__status-text {
	color: #ef4444;
}

.attendance-card__status--default {
	background: rgba(148, 163, 184, 0.12);
}

.attendance-card__status--default .attendance-card__status-text {
	color: #94a3b8;
}

.attendance-card__status-text {
	font-size: 22rpx;
	font-weight: 600;
}

.attendance-card__body {
	display: flex;
	flex-direction: column;
	gap: 8rpx;
}

.attendance-card__row {
	display: flex;
	align-items: center;
}

.attendance-card__info {
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
	color: var(--primary-500);
	font-weight: 600;
}
</style>
