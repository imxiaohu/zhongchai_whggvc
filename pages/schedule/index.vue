	<template>
	<view class="schedule-page">
		<!-- 缓存提示横幅 -->
		<cache-banner
			:visible="showCacheBanner"
			:cache-updated-at="cacheUpdatedAt"
			@close="showCacheBanner = false"
		></cache-banner>


		<!-- 主要内容区域 -->
		<view class="schedule-page__content" :style="{paddingTop: navPaddingTop, flex: 1}">

			<!-- 顶部蓝渐变头部 -->
			<view class="schedule-page__header">
				<!-- 标题已在 CustomNavBar 中显示 -->
			</view>

			<!-- 预览模式横幅 -->
			<view v-if="isPreviewMode" class="schedule-page__preview">
				<view class="schedule-page__preview-icon">
					<l-icon name="browse-filled" size="16" color="#f59e0b"></l-icon>
				</view>
				<text class="schedule-page__preview-text">当前为预览模式，显示示例数据</text>
				<text class="schedule-page__preview-btn" @tap="goToLogin">立即登录</text>
			</view>

			<!-- 课程表容器 -->
			<view class="schedule-wrapper">
				<!-- 课程表组件 -->
				<view class="class-schedule-container" :class="{'fade-in': isDataLoaded && !isLoading}">
					<class-schedule
						v-if="isDataLoaded"
						:courses="courses"
						:timePeriods="timePeriods"
						:currentWeek="currentWeek"
						:displayWeek="displayWeek"
						:totalWeeks="totalWeeks"
						:semesterInfo="semesterInfo"
						ref="classSchedule"
						@prev-week="prevWeek"
						@next-week="nextWeek"
						@week-change="onWeekChange"
					/>
				</view>

				<!-- 加载提示 -->
				<view v-if="!isDataLoaded || isLoading" class="loading-container">
					<view class="loading-spinner"></view>
					<text class="loading-text">{{ isWeekChanging ? '切换中...' : '加载中...' }}</text>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup>
	import { ref, onMounted } from 'vue';
	import { onShow, onPullDownRefresh, onShareAppMessage, onShareTimeline, onUnload } from '@dcloudio/uni-app';

	import { shouldAutoRefreshData } from '@/utils/errorHandler.js';
	import ClassSchedule from '../../components/ClassSchedule.vue';
	import CacheBanner from '../../components/CacheBanner.vue';
	import CustomNavBar from '../../components/CustomNavBar.vue';
	import shareManager from '@/utils/shareManager.js';
	import unreadMessageManager from '@/utils/unreadMessageManager.js';
	import { useSchedule } from '@/composables/useSchedule.js';

	const classSchedule = ref(null);
	const navPaddingTop = ref('0px');
	const currentTheme = ref('light');
	const isDarkMode = ref(false);
	const themeClass = ref('theme-light');

	const {
		courses,
		timePeriods,
		currentWeek,
		totalWeeks,
		displayWeek,
		isDataLoaded,
		isWeeksLoaded,
		isLoading,
		isWeekChanging,
		currentSemester,
		showSetCurrentWeekBtn,
		showSetCurrentWeekPopup,
		tempCurrentWeek,
		lastSwipeDirection,
		isPreviewMode,
		showCacheBanner,
		cacheUpdatedAt,
		semesterInfo,
		fetchTermWeekInfo,
		fetchTimePeriods,
		fetchCourseData,
		initData,
		showPreviewData,
		goToLogin,
		handleSessionExpired,
		prevWeek,
		nextWeek,
		onWeekChange,
		setCurrentWeek,
		resetToApiCurrentWeek
	} = useSchedule();

	// 更新主题状态函数
	const updateThemeState = () => {
		try {
			const systemInfo = uni.getSystemInfoSync();
			if (systemInfo.theme) {
				currentTheme.value = systemInfo.theme;
				isDarkMode.value = systemInfo.theme === 'dark';
				themeClass.value = systemInfo.theme === 'dark' ? 'theme-dark' : 'theme-light';
			}
		} catch (error) {
			currentTheme.value = 'light';
			isDarkMode.value = false;
			themeClass.value = 'theme-light';
		}
	};

	// 处理导航栏高度就绪事件
	const handleNavHeightReady = (navInfo) => {
		navPaddingTop.value = navInfo.heightPx;
	};

	// 刷新未读消息
	const refreshUnreadMessages = async () => {
		try {
			await unreadMessageManager.fetchUnreadCount();
		} catch (e) {}
	};

	const themeChangeCallback = () => {
		updateThemeState()
	}

	// 组件挂载
	onMounted(() => {
		updateThemeState()
		uni.onThemeChange(themeChangeCallback)
		initData()
	});

	// 监听页面显示
	onShow(() => {
		refreshUnreadMessages()
		updateThemeState()
		if (!isDataLoaded.value) {
			initData()
		} else if (shouldAutoRefreshData()) {
			initData()
		}
	});

	// 页面卸载时清理监听器
	onUnload(() => {
		uni.offThemeChange(themeChangeCallback)
	})

	// 下拉刷新
	onPullDownRefresh(() => {
		initData().finally(() => uni.stopPullDownRefresh());
	});

	// #ifdef MP-WEIXIN
	onShareAppMessage(() => {
		return shareManager.generateShareConfig({
			title: '我的课程表 - 众柴智慧校园',
			path: '/pages/schedule/index',
			imageUrl: '/static/images/share-schedule.png',
			trackingData: {
				page: 'schedule_index',
				feature: 'schedule_share',
				currentWeek: currentWeek.value,
				displayWeek: displayWeek.value,
				totalWeeks: totalWeeks.value,
				hasData: isDataLoaded.value,
				isPreview: isPreviewMode.value,
				courseCount: courses.value ? courses.value.length : 0
			}
		});
	});

	onShareTimeline(() => {
		return shareManager.generateTimelineConfig({
			title: '我的课程表 - 众柴智慧校园，轻松查看课程安排',
			query: '',
			imageUrl: '/static/images/share-schedule.png',
			trackingData: {
				page: 'schedule_index',
				feature: 'schedule_timeline_share',
				currentWeek: currentWeek.value,
				displayWeek: displayWeek.value,
				totalWeeks: totalWeeks.value,
				hasData: isDataLoaded.value,
				isPreview: isPreviewMode.value,
				courseCount: courses.value ? courses.value.length : 0
			}
		});
	});
	// #endif
</script>


<style lang="scss" scoped>
.schedule-page {
	flex: 1;
	width: 100%;
	min-height: 100vh;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
	font-family: -apple-system, "SF Pro Text", "SF Pro Icons", sans-serif;



}

.safe-area-top { display: none; }
.nav-bar { display: none; }
.page-title { display: none; }
.week-switcher, .current-week-setting { display: none; }
.preview-banner { display: none; }

.schedule-wrapper {
	width: 100%;
	position: relative;
	height: 100%;
}

.class-schedule-container {
	opacity: 0;
	transition: opacity 0.3s ease-in-out;
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;

	&.fade-in { opacity: 1; }
}

.loading-container {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 100px 0;
	gap: var(--spacing-md);
}

.loading-spinner {
	width: 36px;
	height: 36px;
	border: 3px solid rgba(99, 102, 241, 0.15);
	border-top-color: #6366f1;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

.loading-text {
	font-size: 14px;
	color: var(--text-tertiary);
}

@keyframes spin {
	to { transform: rotate(360deg); }
}

.theme-dark .schedule-page {
	&__preview {
		background: rgba(30, 41, 59, 0.85);
		border-color: rgba(255, 255, 255, 0.1);
	}
}
</style>
