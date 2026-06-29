<template>
	<view class="makeup-exam-page">
		<!-- 自定义导航栏 -->
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">补考查询</text>
				<view class="nav-bar__placeholder"></view>
			</view>
		</view>

		<!-- 筛选信息栏 -->
		<view class="filter-bar">
			<view class="filter-bar__info">
				<text class="filter-bar__text">{{ semesterFilter || '当前学期' }}</text>
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
				<text class="state-text">暂无补考记录</text>
				<text class="state-sub-text">说明：补考信息通常在学期末考试后更新</text>
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
					class="exam-card"
				>
					<view class="exam-card__header">
						<text class="exam-card__course">{{ item.courseName || item.courseiname || '未知课程' }}</text>
						<view class="exam-card__status" :class="'exam-card__status--' + getStatusClass(item)">
							<text class="exam-card__status-text">{{ getStatusText(item) }}</text>
						</view>
					</view>
					<view class="exam-card__body">
						<view class="exam-card__row" v-if="item.courseScore !== undefined">
							<text class="exam-card__label">原成绩</text>
							<text class="exam-card__value exam-card__value--danger">{{ item.courseScore }}</text>
						</view>
						<view class="exam-card__row" v-if="item.examScore !== undefined">
							<text class="exam-card__label">补考成绩</text>
							<text class="exam-card__value" :class="getScoreClass(item.examScore)">{{ item.examScore }}</text>
						</view>
						<view class="exam-card__row" v-if="item.totalScore !== undefined">
							<text class="exam-card__label">综合成绩</text>
							<text class="exam-card__value" :class="getScoreClass(item.totalScore)">{{ item.totalScore }}</text>
						</view>
						<view class="exam-card__row" v-if="item.examTime">
							<text class="exam-card__label">考试时间</text>
							<text class="exam-card__value">{{ item.examTime }}</text>
						</view>
						<view class="exam-card__row" v-if="item.examAddress">
							<text class="exam-card__label">考试地点</text>
							<text class="exam-card__value">{{ item.examAddress }}</text>
						</view>
						<view class="exam-card__row" v-if="item.teacherNames">
							<text class="exam-card__label">授课老师</text>
							<text class="exam-card__value">{{ item.teacherNames }}</text>
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
import { pcGetMakeupExamList } from '../../pages/api/discover.js';
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
			uni.navigateTo({ url: '/pages/discover/pc-login?redirect=/pages/discover/makeup-exam' });
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
					semester: this.semesterFilter
				};
				const res = await pcGetMakeupExamList(params);
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
					if (reset) {
						this.records = result.list;
					} else {
						this.records = [...this.records, ...result.list];
					}
					this.total = result.total || 0;
					this.noMore = this.records.length >= this.total || (result.pages && this.pageNo >= result.pages);
				} else {
					this.records = [];
					this.total = 0;
				}
			} catch (e) {
				console.error('获取补考记录失败', e);
				this.error = e.message || '获取补考记录失败';
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
		getStatusClass(item) {
			if (item.examScore !== undefined) {
				const score = parseFloat(item.examScore);
				if (isNaN(score)) return 'default';
				return score >= 60 ? 'success' : 'danger';
			}
			if (item.courseScore !== undefined) return 'warn';
			return 'default';
		},
		getStatusText(item) {
			if (item.examScore !== undefined) {
				const score = parseFloat(item.examScore);
				if (isNaN(score)) return '待补考';
				return score >= 60 ? '已通过' : '未通过';
			}
			return '待补考';
		},
		getScoreClass(score) {
			const num = parseFloat(score);
			if (isNaN(num)) return '';
			if (num >= 90) return 'exam-card__value--excellent';
			if (num >= 80) return 'exam-card__value--good';
			if (num >= 60) return 'exam-card__value--pass';
			return 'exam-card__value--fail';
		}
	}
}
</script>

<style lang="scss" scoped>
@import url('../../static/css/home-page.css');

.makeup-exam-page {
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
	justify-content: flex-start;
}

.nav-bar__title {
	font-size: 32rpx;
	font-weight: 700;
	color: #fff;
	text-align: center;
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

.state-text {
	font-size: 28rpx;
	color: var(--text-secondary);
	font-weight: 600;
}

.state-sub-text {
	font-size: 24rpx;
	color: var(--text-tertiary);
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

.exam-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 28rpx;
	margin-bottom: 16rpx;
	box-shadow: 0 2rpx 12rpx rgba(139, 92, 246, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.exam-card__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 20rpx;
	padding-bottom: 16rpx;
	border-bottom: 1px solid rgba(226, 232, 240, 0.5);
}

.exam-card__course {
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	flex: 1;
	margin-right: 16rpx;
}

.exam-card__status {
	padding: 6rpx 16rpx;
	border-radius: 16rpx;
	flex-shrink: 0;
}

.exam-card__status--success {
	background: rgba(16, 185, 129, 0.12);
}

.exam-card__status--success .exam-card__status-text {
	color: #10b981;
}

.exam-card__status--warn {
	background: rgba(245, 158, 11, 0.12);
}

.exam-card__status--warn .exam-card__status-text {
	color: #f59e0b;
}

.exam-card__status--danger {
	background: rgba(239, 68, 68, 0.12);
}

.exam-card__status--danger .exam-card__status-text {
	color: #ef4444;
}

.exam-card__status--default {
	background: rgba(148, 163, 184, 0.12);
}

.exam-card__status--default .exam-card__status-text {
	color: #94a3b8;
}

.exam-card__status-text {
	font-size: 22rpx;
	font-weight: 600;
}

.exam-card__body {
	display: flex;
	flex-direction: column;
	gap: 10rpx;
}

.exam-card__row {
	display: flex;
	align-items: center;
}

.exam-card__label {
	font-size: 24rpx;
	color: var(--text-tertiary);
	min-width: 140rpx;
	margin-right: 16rpx;
}

.exam-card__value {
	font-size: 26rpx;
	color: var(--text-primary);
	font-weight: 600;
}

.exam-card__value--danger {
	color: #ef4444;
}

.exam-card__value--excellent {
	color: #10b981;
}

.exam-card__value--good {
	color: #10b981;
}

.exam-card__value--pass {
	color: #f59e0b;
}

.exam-card__value--fail {
	color: #ef4444;
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
