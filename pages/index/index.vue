<template>
	<view class="home-page">
		<!-- 顶部蓝色渐变 Hero 区域（含背景图） -->
		<view class="hero-section">
			<!-- 背景渐变 + 图片叠加 -->
			<view class="hero-bg"></view>
			<image
				class="hero-bg-image"
				src="/static/images/index_bg.png"
				mode="aspectFill"
			></image>
			<view class="hero-bg-overlay"></view>

			<!-- 状态栏占位 -->
			<view class="hero-statusbar" :style="{ height: statusBarHeight + 'px' }"></view>

			<!-- 顶部导航栏 -->
			<view class="hero-nav">
				<view class="hero-nav-left" @tap="handleTitleClick">
					<text class="hero-brand">众柴智慧校园</text>
					<text class="hero-brand-en">YOUR LEARNING COMPANION</text>
				</view>
				<view class="hero-nav-right">
					<view class="hero-icon-btn" @tap="goToNotifications">
						<l-icon name="notification" style="font-size: 22px; color: #fff;"></l-icon>
						<view v-if="unreadCount > 0" class="hero-badge">{{ unreadCount > 99 ? '99+' : unreadCount }}</view>
					</view>
					<view class="hero-avatar" @tap="goToUser">
						<text class="hero-avatar-text">{{ avatarText }}</text>
					</view>
				</view>
			</view>

			<!-- 欢迎语与搜索框 -->
			<view class="hero-content">
				<view class="hero-greeting">
					<text class="hero-title">{{ greetingText }}, {{ userNickname }}</text>
					<text class="hero-subtitle">{{ greetingSubtitle }}</text>
				</view>

				<view class="hero-search" @tap="handleSearch">
					<l-icon name="search" style="font-size: 18px; color: rgba(255,255,255,0.7); margin-right: 12rpx;"></l-icon>
					<text class="hero-search-placeholder">搜索课程、成绩、通知…</text>
				</view>
			</view>
		</view>

		<!-- 主体内容区域 -->
		<view class="content-area" :style="{paddingTop: navPaddingTop}">

			<!-- 学校限制横幅 -->
			<view v-if="showLoginPrompt" class="school-banner">
				<view class="banner-content">
					<l-icon name="info-circle-filled" style="font-size: 20px; color: var(--warning-color); margin-right: 8rpx;"></l-icon>
					<text class="banner-text">仅限武汉光谷职业学院内部使用，拥有教务账号才可登录使用！</text>
				</view>
			</view>

			<!-- 登录提示 -->
			<view v-if="showLoginPrompt" class="prompt-card">
				<view class="prompt-icon prompt-icon--login">
					<l-icon name="user" style="font-size: 32px; color: #fff;"></l-icon>
				</view>
				<text class="prompt-title">欢迎使用评教系统</text>
				<text class="prompt-desc">登录后可获取真实的课表和成绩数据</text>
				<button class="prompt-btn prompt-btn--primary" @tap="handleLogin">立即登录</button>
			</view>

			<!-- 绑定提示 -->
			<view v-else-if="showBindingPrompt" class="prompt-card">
				<view class="prompt-icon prompt-icon--warn">
					<l-icon name="link" style="font-size: 32px; color: #fff;"></l-icon>
				</view>
				<text class="prompt-title">绑定学校账号</text>
				<text class="prompt-desc">绑定学校账号后可获取课表和成绩数据</text>
				<button class="prompt-btn prompt-btn--warn" @tap="handleBind">立即绑定</button>
			</view>

			<!-- 今日课表模块 -->
			<view v-else-if="loading" class="state-card">
				<view class="state-spinner"></view>
				<text class="state-text">正在加载今日课程...</text>
			</view>
			<view v-else-if="error" class="state-card state-card--error">
				<view class="state-icon-circle state-icon-circle--error">!</view>
				<text class="state-text">{{error}}</text>
				<view v-if="serverMaintenance" class="state-subtext">学校服务器维护中，请稍后再试</view>
			</view>
			<view v-else-if="originalClassData.length === 0" class="state-card">
				<view class="state-emoji">🌿</view>
				<text class="state-text">今天没有课程</text>
				<text class="state-subtext">享受难得的休息时光吧~</text>
			</view>
			<view v-else class="section-card">
				<view class="section-header">
					<view class="section-header-left">
						<text class="section-title-main">今日课程</text>
						<text class="section-title-sub">{{ todayDateText }}</text>
					</view>
					<text class="section-action" @tap="goToShedule">查看全部 ›</text>
				</view>
			<TodayClass
				:classes="originalClassData"
				:loading="false"
				:hideHeader="true"
				:layout="todayClassLayout"
				:hidePast="hidePastClasses"
			/>
			</view>

			<!-- 快捷功能区 -->
			<view class="section-card">
				<view class="section-header">
					<view class="section-header-left">
						<text class="section-title-main">快捷功能</text>
						<text class="section-title-sub">QUICK ACCESS</text>
					</view>
				</view>
				<view class="quick-grid">
					<view class="quick-item"
						:class="{'disabled': !featureAvailable}"
						hover-class="quick-item-hover"
						@tap="goToEvaluation">
						<view class="quick-icon quick-icon--evaluation">
							<l-icon name="edit" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<text class="quick-label">评教</text>
						<text class="quick-desc">{{ featureAvailable ? '对课程进行评价' : (!isLoggedIn ? '需要登录' : '维护中') }}</text>
						<view v-if="!featureAvailable" class="quick-lock">
							<l-icon name="lock-on" style="font-size: 14px; color: #fff;"></l-icon>
						</view>
					</view>
					<view class="quick-item"
						:class="{'disabled': !featureAvailable}"
						hover-class="quick-item-hover"
						@tap="goToShedule">
						<view class="quick-icon quick-icon--schedule">
							<l-icon name="calendar" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<text class="quick-label">课表</text>
						<text class="quick-desc">{{ featureAvailable ? '查看课程安排' : (!isLoggedIn ? '需要登录' : '维护中') }}</text>
						<view v-if="!featureAvailable" class="quick-lock">
							<l-icon name="lock-on" style="font-size: 14px; color: #fff;"></l-icon>
						</view>
					</view>
					<view class="quick-item"
						:class="{'disabled': !featureAvailable}"
						hover-class="quick-item-hover"
						@tap="goToScore">
						<view class="quick-icon quick-icon--grade">
							<l-icon name="book" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<text class="quick-label">成绩</text>
						<text class="quick-desc">{{ featureAvailable ? '查询课程成绩' : (!isLoggedIn ? '需要登录' : '维护中') }}</text>
						<view v-if="!featureAvailable" class="quick-lock">
							<l-icon name="lock-on" style="font-size: 14px; color: #fff;"></l-icon>
						</view>
					</view>
					<view class="quick-item"
						hover-class="quick-item-hover"
						@tap="goToCommunity">
						<view class="quick-icon quick-icon--community">
							<l-icon name="usergroup" style="font-size: 22px; color: #fff;"></l-icon>
						</view>
						<text class="quick-label">社区</text>
						<text class="quick-desc">学习交流园地</text>
					</view>
				</view>
			</view>

			<!-- 通知公告 -->
			<view class="section-card">
				<view class="section-header">
					<view class="section-header-left">
						<text class="section-title-main">最新动态</text>
						<text class="section-title-sub">LATEST NEWS</text>
					</view>
					<text class="section-action" @tap="goToNotices">查看全部 ›</text>
				</view>

				<view v-if="loadingNotices" class="state-card state-card--compact">
					<view class="state-spinner"></view>
					<text class="state-text">加载中…</text>
				</view>
				<view v-else-if="noticeError" class="state-card state-card--compact state-card--error">
					<text class="state-text">{{noticeError}}</text>
				</view>
				<view v-else-if="notices.length === 0" class="state-card state-card--compact">
					<text class="state-text">暂无通知</text>
				</view>
				<view v-else class="notice-list">
					<view
						v-for="notice in notices"
						:key="notice.id"
						class="notice-item"
						hover-class="notice-item-hover"
						@tap="viewNoticeDetail(notice.id)"
					>
						<view class="notice-bar" :class="{'notice-bar--top': notice.isTop}"></view>
						<view class="notice-body">
							<view class="notice-title-row">
								<text class="notice-title" :class="{'notice-title--unread': !notice.isRead}">{{notice.title}}</text>
								<view v-if="notice.isTop" class="notice-pin">
									<l-icon name="pin" style="font-size: 11px; color: #fff;"></l-icon>
									<text class="notice-pin-text">置顶</text>
								</view>
							</view>
							<view class="notice-meta">
								<text class="notice-source">{{notice.source || '系统'}}</text>
								<text class="notice-dot">·</text>
								<text class="notice-date">{{formatNoticeDate(notice.publishTime || notice.createTime)}}</text>
							</view>
						</view>
					</view>
				</view>
			</view>

			<!-- 底部安全区 -->
			<view class="bottom-safe-area"></view>
		</view>
	</view>
</template>

<script>
import { showToast, switchTab, navigateTo } from '../../pages/api/page.js';
import { getTodayTimetable } from '../../pages/api/course.js';
import { getTopNews, clearNewsCache } from '../../pages/api/news.js';
import { shouldAutoRefreshData } from '@/utils/errorHandler.js';
import TodayClass from '../../components/TodayClass.vue';
import { useCourseCache } from '../../store/courseCache.js';
import unreadMessageManager from '@/utils/unreadMessageManager.js';
import shareManager from '@/utils/shareManager.js';
import {
	isServerMaintenanceError,
	getGreetingText,
	getGreetingSubtitle,
	getAvatarText,
	getTodayDateText,
	formatNoticeDate,
	initStatusBarHeight,
	refreshUserInfo,
	isFeatureAvailable,
	getFeatureUnavailableMessage,
	getCurrentWeek,
	onClearAllLoadingStates
} from '@/utils/homePage.js';
import {
	formatNewsItem,
	formatNewsList,
	createPreviewNotices
} from '@/utils/newsUtils.js';
import {
	formatTodayClasses,
	createPreviewClasses,
	checkFromSettingsPage,
	clearSettingsPageMark,
	setErrorSuppress,
	clearErrorSuppress
} from '@/utils/scheduleUtils.js';

export default {
	components: {
		TodayClass
	},
	data() {
		return {
			loading: false,
			error: '',
			todayClasses: [],
			originalClassData: [],
			hasInitialized: false,
			isLoggedIn: false,
			showLoginPrompt: false,
			showBindingPrompt: false,
			serverMaintenance: false,
			loadingNotices: false,
			noticeError: '',
			notices: [],
			navPaddingTop: '0px',
			statusBarHeight: 20,
			unreadCount: 0,
			userNickname: '同学',
			userInfo: null,
			todayClassLayout: 'vertical',
			hidePastClasses: false
		}
	},
	computed: {
		avatarText() {
			return getAvatarText(this.userNickname);
		},
		todayDateText() {
			return getTodayDateText();
		},
		greetingText() {
			return getGreetingText(this.isLoggedIn);
		},
		greetingSubtitle() {
			return getGreetingSubtitle(this.isLoggedIn, this.showBindingPrompt);
		},
		featureAvailable() {
			return isFeatureAvailable(this.isLoggedIn, this.serverMaintenance);
		}
	},
	onLoad() {
		this.initStatusBarHeight();
		this.initPage();
		uni.$on('clearAllLoadingStates', onClearAllLoadingStates);
	},
	onUnload() {
		uni.$off('clearAllLoadingStates', onClearAllLoadingStates);
	},
	onShow() {
		this.refreshUnreadMessages();
		this.refreshUserInfo();
		this.loadDisplaySettings();

		try {
			const shouldRefresh = shouldAutoRefreshData();
			if (this.hasInitialized && this.isLoggedIn && shouldRefresh) {
				this.fetchTodayClasses();
				this.fetchNotices();
			}
		} catch (error) {
			console.error('调用shouldAutoRefreshData失败:', error);
		}
	},
	methods: {
		formatNoticeDate,
		async initPage() {
			this.loading = true;
			try {
				const token = uni.getStorageSync('token');
				const clientId = uni.getStorageSync('clientId');
				const userInfoStr = uni.getStorageSync('userInfo');
				console.log('[index] initPage 开始，token:', !!token, 'clientId:', !!clientId, 'userInfo:', !!userInfoStr);
				await this.checkLoginStatus();
			} finally {
				console.log('[index] initPage finally，当前 loading:', this.loading, 'isLoggedIn:', this.isLoggedIn);
				this.loading = false;
			}
		},
		initStatusBarHeight() {
			this.statusBarHeight = initStatusBarHeight();
		},
		refreshUserInfo() {
			const result = refreshUserInfo();
			this.userInfo = result.userInfo;
			this.userNickname = result.nickname;
		},
		loadDisplaySettings() {
			const layout = uni.getStorageSync('todayClassLayout');
			this.todayClassLayout = layout || 'vertical';
			const hidePast = uni.getStorageSync('hidePastClasses');
			this.hidePastClasses = hidePast === 'true';
		},
		async refreshUnreadMessages() {
			try {
				const count = await unreadMessageManager.fetchUnreadCount();
				this.unreadCount = count || unreadMessageManager.getUnreadCount() || 0;
			} catch (error) {
				console.error('首页: 刷新未读消息数量失败:', error);
			}
		},
		handleNavHeightReady(navInfo) {
			this.navPaddingTop = navInfo.heightPx;
		},
		goToNotifications() {
			uni.vibrateShort();
			navigateTo({ url: '/pages/user/notification-center' });
		},
		goToUser() {
			switchTab({ url: '/pages/user/index' });
		},
		goToCommunity() {
			switchTab({ url: '/pages/community/index' });
		},
		handleSearch() {
			uni.vibrateShort();
			showToast({ title: '搜索功能即将上线', icon: 'none' });
		},
		async getCurrentWeek() {
			return await getCurrentWeek();
		},
		checkLoginStatus() {
			const token = uni.getStorageSync('token');
			const userInfoStr = uni.getStorageSync('userInfo');
			const clientId = uni.getStorageSync('clientId');
			const hasBindSchoolAccount = uni.getStorageSync('hasBindSchoolAccount');
			const isSchoolLogin = token && (token.startsWith('logged_in_') || !userInfoStr);
			console.log('[index] checkLoginStatus:', {
				tokenPrefix: token ? token.substring(0, 20) : null,
				hasClientId: !!clientId,
				hasUserInfo: !!userInfoStr,
				hasBindSchoolAccount,
				isSchoolLogin
			});

			if (!token) {
				this.hasInitialized = true;
				this.loading = false;
				this.isLoggedIn = false;
				this.showPreviewData();
				console.log('[index] 分支A: 无token，显示预览数据');
				return;
			}

			this.isLoggedIn = true;
			this.getCurrentWeek().catch(error => {
				console.error('获取当前教学周失败:', error);
			});

			if (!clientId) {
				const hasUserInfo = !!userInfoStr;

				if (hasUserInfo) {
					this.hasInitialized = true;
					this.loading = false;
					this.showBindingPromptState();
					console.log('[index] 分支B: 无clientId有userInfo，显示绑定提示');
					return;
				} else {
					uni.removeStorageSync('token');
					this.isLoggedIn = false;
					this.loading = false;
					this.showPreviewData();
					console.log('[index] 分支C: 无clientId无userInfo，清除token，显示预览');
					return;
				}
			}

			if (isSchoolLogin) {
				uni.setStorageSync('hasBindSchoolAccount', true);
				this.hasInitialized = true;
				this.loading = false;
				this.fetchTodayClasses();
				this.fetchNotices();
				console.log('[index] 分支D: 学校登录，调用fetchTodayClasses');
				return;
			}

			if (!hasBindSchoolAccount) {
				this.hasInitialized = true;
				this.loading = false;
				this.showBindingPromptState();
				console.log('[index] 分支E: 未绑定学校账号，显示绑定提示');
			} else {
				this.hasInitialized = true;
				this.loading = false;
				this.fetchTodayClasses();
				this.fetchNotices();
				console.log('[index] 分支F: 已绑定，调用fetchTodayClasses');
			}
		},
		showPreviewData() {
			this.showLoginPrompt = true;
			this.showBindingPrompt = false;
			this.todayClasses = createPreviewClasses();
			this.notices = createPreviewNotices();
		},
		showBindingPromptState() {
			this.showLoginPrompt = false;
			this.showBindingPrompt = true;
			this.todayClasses = [];
			this.notices = [];
		},
		handleLogin() {
			uni.vibrateShort();
			uni.navigateTo({ url: '/pages/login/login' });
		},
		handleBind() {
			uni.vibrateShort();
			uni.navigateTo({ url: '/pages/user/bind' });
		},
		async fetchTodayClasses() {
			if (this.loading) return;

			this.loading = true;
			this.error = '';

			try {
				const isFromSettings = checkFromSettingsPage();

				if (!isFromSettings) {
					setErrorSuppress(15000);
				} else {
					clearSettingsPageMark();
				}

				const courseCacheStore = useCourseCache();
				const [lessonTimeResult, currentTimeResult] = await Promise.allSettled([
					courseCacheStore.getCourseLessonTime(),
					courseCacheStore.getCurrentTime()
				]);
				console.log('[index] fetchTodayClasses Promise.allSettled 结果:', {
					lessonTimeStatus: lessonTimeResult.status,
					lessonTimeValue: lessonTimeResult.status === 'fulfilled' ? lessonTimeResult.value : lessonTimeResult.reason?.message,
					currentTimeStatus: currentTimeResult.status,
					currentTimeValue: currentTimeResult.status === 'fulfilled' ? currentTimeResult.value : currentTimeResult.reason?.message
				});

				console.log('[index] fetchTodayClasses 准备调用 getTodayTimetable');
				const result = await getTodayTimetable();

				if (result && result.success === false) {
					if (result.message && result.message.includes('会话已失效')) {
						this.handleSessionExpired();
						return;
					}
					throw new Error(result.message || '获取数据失败');
				}

				const { originalData, formattedData } = formatTodayClasses(result);
				this.originalClassData = originalData;
				this.todayClasses = formattedData;

				this.loading = false;
				clearErrorSuppress();
			} catch (error) {
				console.error('获取今日课表失败:', error);
				clearErrorSuppress();

				if ((error.message && error.message.includes('会话已失效')) ||
					error.statusCode === 401 || error.isTokenInvalid) {
					this.handleSessionExpiredOnHome();
					return;
				}

				if (isServerMaintenanceError(error)) {
					this.serverMaintenance = true;
					this.error = '学校服务器正在维护中，暂时无法获取课表数据';
				} else {
					this.error = '获取课表失败，请稍后再试';
				}

				this.loading = false;

				const errorMessage = this.serverMaintenance ?
					'学校服务器维护中，请稍后再试' :
					(error.message || '获取课表失败');
				showToast({ title: errorMessage, icon: 'none', duration: 3000 });
			}
		},
		handleSessionExpiredOnHome() {
			uni.removeStorageSync('token');
			uni.removeStorageSync('clientId');

			this.isLoggedIn = false;
			this.hasInitialized = true;
			this.loading = false;

			this.showPreviewData();

			uni.showToast({
				title: '登录已过期，请重新登录获取真实数据',
				icon: 'none',
				duration: 3000
			});
		},
		handleSessionExpired() {
			const userInfoStr = uni.getStorageSync('userInfo');
			const isWechatLogin = !!userInfoStr;

			if (isWechatLogin) {
				this.error = '请先绑定学校账号';
				showToast({ title: '请先绑定学校账号以获取课表', icon: 'none', duration: 2000 });

				setTimeout(() => {
					uni.navigateTo({ url: '/pages/user/bind' });
				}, 2000);

				this.loading = false;
			} else {
				uni.removeStorageSync('token');
				uni.removeStorageSync('clientId');
				uni.reLaunch({ url: '/pages/login/login' });
			}
		},
		goToEvaluation() {
			uni.vibrateShort();
			navigateTo({ url: '/pages/evaluation/list' });
		},
		goToShedule() {
			uni.vibrateShort();
			if (!this.featureAvailable) {
				showToast({
					title: getFeatureUnavailableMessage(this.isLoggedIn, this.serverMaintenance),
					icon: 'none',
					duration: 3000
				});
				return;
			}
			switchTab({ url: '/pages/schedule/index' });
		},
		goToScore() {
			uni.vibrateShort();
			if (!this.featureAvailable) {
				showToast({
					title: getFeatureUnavailableMessage(this.isLoggedIn, this.serverMaintenance),
					icon: 'none',
					duration: 3000
				});
				return;
			}

			const hasBindSchoolAccount = uni.getStorageSync('hasBindSchoolAccount');
			const userInfoStr = uni.getStorageSync('userInfo');

			if (!hasBindSchoolAccount && !userInfoStr) {
				showToast({ title: '请先登录获取真实数据', icon: 'none' });
				return;
			}

			if (!hasBindSchoolAccount) {
				showToast({ title: '请先绑定学校账号', icon: 'none' });
				setTimeout(() => {
					navigateTo({ url: '/pages/user/bind' });
				}, 1500);
				return;
			}

			navigateTo({ url: '/pages/user/score' });
		},
		handleTitleClick() {
			showToast({ title: '版本信息', icon: 'none', duration: 1500 });
		},
		async fetchNotices() {
			this.loadingNotices = true;
			this.noticeError = '';

			try {
				const isFromSettings = checkFromSettingsPage();

				if (!isFromSettings) {
					setErrorSuppress(15000);
				}

				const forceRefresh = isFromSettings || !this.notices || this.notices.length === 0;

				if (forceRefresh) {
					clearNewsCache();
				}

				const notices = await getTopNews(5, forceRefresh);

				if (!notices || notices.length === 0) {
					this.notices = [];
					this.loadingNotices = false;
					clearErrorSuppress();
					return;
				}

				this.notices = formatNewsList(notices);
				this.loadingNotices = false;
				clearErrorSuppress();
			} catch (error) {
				console.error('fetchNotices: 获取通知公告失败:', error);

				clearErrorSuppress();

				if ((error.message && error.message.includes('会话已失效')) ||
					error.statusCode === 401 || error.isTokenInvalid) {
					this.handleSessionExpiredOnHome();
					return;
				}

				this.noticeError = '';
				this.loadingNotices = false;
				this.notices = [];
			}
		},
		viewNoticeDetail(id) {
			navigateTo({ url: `/pages/news/detail?id=${id}` });
		},
		goToNotices() {
			navigateTo({ url: '/pages/news/list' });
		}
	},

	// #ifdef MP-WEIXIN
	onShareAppMessage() {
		return shareManager.generateShareConfig({
			title: '众柴智慧校园 - 让评教更简单',
			path: '/pages/index/index',
			imageUrl: '/static/images/share-home.png',
			trackingData: {
				page: 'home_index',
				feature: 'home_share',
				hasLogin: this.isLoggedIn,
				hasClasses: this.todayClasses.length > 0,
				hasNotices: this.notices.length > 0
			}
		});
	},
	onShareTimeline() {
		return shareManager.generateTimelineConfig({
			title: '众柴智慧校园 - 让评教更简单，课表查询更便捷',
			query: '',
			imageUrl: '/static/images/share-home.png',
			trackingData: {
				page: 'home_index',
				feature: 'home_timeline_share',
				hasLogin: this.isLoggedIn,
				hasClasses: this.todayClasses.length > 0,
				hasNotices: this.notices.length > 0
			}
		});
	}
	// #endif
}
</script>

<style lang="scss" scoped>
/* Styles moved to static/css/home-page.css */
</style>
