<template>
	<view class="user-page">
		<!-- 主要内容区域 -->
		<view class="user-page__content" :style="{paddingTop: navPaddingTop}">

			<!-- 顶部渐变背景 -->
			<view class="user-page__header-bg">
				<!-- 用户个人信息卡片 -->
				<view class="user-page__profile" @tap="handleUserInfoClick">
					<view class="user-page__profile-avatar">
						{{ getUserInitial(userInfo) }}
					</view>
					<view class="user-page__profile-info">
						<text class="user-page__profile-name">{{userInfo.name || userInfo.realname || '未登录'}}</text>
						<text class="user-page__profile-id">{{isLoggedIn ? '学号：' + (userInfo.studentId || userInfo.username || '未绑定') : '登录后开启更多功能'}}</text>
					</view>
				</view>

				<!-- 学院与班级标签 -->
				<view class="user-page__tags" v-if="userInfo.college || userInfo.className">
					<view class="user-page__tag" v-if="userInfo.college">
						<l-icon name="education" size="12"></l-icon>
						<text class="user-page__tag-text">{{userInfo.college || '未知学院'}}</text>
					</view>
					<view class="user-page__tag" v-if="userInfo.className">
						<l-icon name="usergroup" size="12"></l-icon>
						<text class="user-page__tag-text">{{userInfo.className || '未知班级'}}</text>
					</view>
				</view>
			</view>

			<!-- 数据统计栏  -->
			<view class="user-page__stats" v-if="isLoggedIn">
				<view class="user-page__stats-item" @tap="handleNavigateToScore">
					<text class="user-page__stats-value">{{userStats.averageScore || '--'}}</text>
					<text class="user-page__stats-label">平均分</text>
				</view>
				<view class="user-page__stats-divider"></view>
				<view class="user-page__stats-item" @tap="handleNavigateToScore">
					<text class="user-page__stats-value">{{userStats.gpa || '--'}}</text>
					<text class="user-page__stats-label">平均绩点</text>
				</view>
				<view class="user-page__stats-divider"></view>
				<view class="user-page__stats-item">
					<text class="user-page__stats-value">{{userStats.pendingEvaluationCount !== null ? userStats.pendingEvaluationCount : '--'}}</text>
					<text class="user-page__stats-label">待评课程</text>
				</view>
				<view class="user-page__stats-divider"></view>
				<view class="user-page__stats-item">
					<text class="user-page__stats-value">{{userStats.creditTotal || '--'}}</text>
					<text class="user-page__stats-label">已获学分</text>
				</view>
			</view>
			
			<!-- 功能菜单组：教务服务 -->
			<view class="user-page__section">
				<view class="user-page__section-title">教务服务</view>
				<view class="user-page__menu">
				<!-- 成绩查询 -->
				<view class="user-page__menu-item" @tap="handleNavigateToScore">
					<view class="user-page__menu-icon user-page__menu-icon--score">
						<l-icon name="book-open" size="20"></l-icon>
					</view>
					<text class="user-page__menu-text">成绩查询</text>
					<l-icon name="chevron-right" size="18" color="var(--text-tertiary)"></l-icon>
				</view>

				<!-- 绑定状态 -->
				<view class="user-page__menu-item" @tap="handleNavigateToBind" v-if="loginType === 'wechat'">
					<view class="user-page__menu-icon user-page__menu-icon--bind">
						<l-icon :name="isSchoolAccountBound ? 'shield-checked' : 'shield'" size="20"></l-icon>
					</view>
					<text class="user-page__menu-text">{{bindMenuText}}</text>
					<view class="user-page__menu-status" v-if="isSchoolAccountBound">已绑定</view>
					<l-icon name="chevron-right" size="18" color="var(--text-tertiary)"></l-icon>
				</view>
				</view>
			</view>

			<!-- 功能菜单组：个人与系统 -->
			<view class="user-page__section">
				<view class="user-page__section-title">个人与系统</view>
				<view class="user-page__menu">
				<!-- 消息中心 -->
				<view class="user-page__menu-item" @tap="handleNavigateToNotificationCenter">
					<view class="user-page__menu-icon user-page__menu-icon--notification">
						<l-icon name="notification" size="20"></l-icon>
					</view>
					<text class="user-page__menu-text">消息中心</text>
					<view class="user-page__menu-badge" v-if="unreadCount > 0">
						{{unreadCount > 99 ? '99+' : unreadCount}}
					</view>
					<l-icon name="chevron-right" size="18" color="var(--text-tertiary)"></l-icon>
				</view>

				<!-- 我的收藏 -->
				<view v-if="isCommunityEnabled" class="user-page__menu-item" @tap="handleNavigateToBookmark">
					<view class="user-page__menu-icon user-page__menu-icon--bookmark">
						<l-icon name="star" size="20"></l-icon>
					</view>
					<text class="user-page__menu-text">我的收藏</text>
					<l-icon name="chevron-right" size="18" color="var(--text-tertiary)"></l-icon>
				</view>

				<!-- 服务器状态 -->
				<view class="user-page__menu-item" @tap="handleNavigateToServerStatus">
					<view class="user-page__menu-icon user-page__menu-icon--server">
						<l-icon name="server" size="20"></l-icon>
					</view>
					<text class="user-page__menu-text">服务器状态</text>
					<view class="user-page__menu-status" :class="serverStatus.isAlive ? 'user-page__menu-status--online' : 'user-page__menu-status--offline'">
						{{serverStatus.isAlive ? '在线' : '离线'}}
					</view>
					<l-icon name="chevron-right" size="18" color="var(--text-tertiary)"></l-icon>
				</view>

				<!-- 设置 -->
				<view class="user-page__menu-item" @tap="handleNavigateToSettings">
					<view class="user-page__menu-icon user-page__menu-icon--setting">
						<l-icon name="setting" size="20"></l-icon>
					</view>
					<text class="user-page__menu-text">系统设置</text>
					<l-icon name="chevron-right" size="18" color="var(--text-tertiary)"></l-icon>
				</view>
				</view>
			</view>

			<!-- 账号操作按钮 -->
			<view class="user-page__footer">
				<view v-if="isLoggedIn" class="user-page__logout-btn" @tap="handleLogout">
					<l-icon name="logout" size="18" color="var(--error-color)"></l-icon>
					<text class="user-page__logout-btn-text">退出当前账号</text>
				</view>
				<view v-else class="user-page__login-btn" @tap="handleGoToLogin">
					<text class="user-page__login-btn-text">立即登录</text>
				</view>
				<view class="user-page__version">Version 1.0.0 · 众柴</view>
			</view>

			<!-- 底部安全区域 -->
			<view class="user-page__safe-area"></view>
		</view>
	</view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow, onShareAppMessage, onShareTimeline } from '@dcloudio/uni-app'
import { useUserHome } from '@/composables/useUserHome.js'
import { getUserInitial } from '@/utils/userHome.js'
import shareManager from '@/utils/shareManager.js'

const navPaddingTop = ref('0px')

const {
	userInfo,
	loginType,
	isLoggedIn,
	unreadCount,
	userStats,
	serverStatus,
	isSchoolAccountBound,
	bindMenuText,
	isCommunityEnabled,
	loadUserInfo,
	refreshUserStatistics,
	handleUserInfoClick,
	handleGoToLogin,
	handleNavigateToScore,
	handleNavigateToBind,
	showBindingInfo,
	confirmRebind,
	confirmUnbind,
	unbindSchoolAccount,
	handleNavigateToSettings,
	handleLogout,
	refreshUnreadMessages,
	handleNavigateToBookmark,
	handleNavigateToNotificationCenter,
	handleNavigateToProfileEdit,
	loadServerStatus,
	handleNavigateToServerStatus,
	onPageShowLogic
} = useUserHome()

function handleNavHeightReady(navInfo) {
	navPaddingTop.value = navInfo.heightPx
}

refreshUnreadMessages()
loadServerStatus()

onShow(() => {
	onPageShowLogic()
})

// #ifdef MP-WEIXIN
onShareAppMessage(() => {
	return shareManager.generateShareConfig({
		title: '个人中心 - 众柴智慧校园',
		path: '/pages/user/index',
		imageUrl: '/static/images/share-user.png',
		trackingData: {
			page: 'user_index',
			feature: 'user_share',
			isLoggedIn: isLoggedIn.value,
			loginType: loginType.value,
			hasSchoolAccount: isSchoolAccountBound.value,
			unreadCount: unreadCount.value,
			serverStatus: serverStatus.value.isAlive
		}
	})
})

onShareTimeline(() => {
	return shareManager.generateTimelineConfig({
		title: '个人中心 - 众柴智慧校园，管理学习生活更便捷',
		query: '',
		imageUrl: '/static/images/share-user.png',
		trackingData: {
			page: 'user_index',
			feature: 'user_timeline_share',
			isLoggedIn: isLoggedIn.value,
			loginType: loginType.value,
			hasSchoolAccount: isSchoolAccountBound.value,
			unreadCount: unreadCount.value,
			serverStatus: serverStatus.value.isAlive
		}
	})
})
// #endif
</script>

<style lang="scss" scoped>
.user-page {
	width: 100%;
	min-height: 100vh;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
	font-family: -apple-system, "SF Pro Text", "SF Pro Icons", sans-serif;

	&__content {
		padding: 0 var(--spacing-md);
		padding-bottom: calc(env(safe-area-inset-bottom) + var(--spacing-lg));
		flex: 1;
	}

	/* 顶部渐变背景 */
	&__header-bg {
		position: relative;
		margin: 0 calc(-1 * var(--spacing-md));
		padding: var(--spacing-xl) var(--spacing-lg);
		background: linear-gradient(180deg, #6366f1 0%, #818cf8 40%, var(--bg-secondary) 100%);
		min-height: 160px;
		
		border-radius: 0 0 24px 24px;
	}

	&__profile {
		display: flex;
		align-items: center;
		position: relative;
		z-index: 1;

		&:active {
			opacity: 0.9;
		}

		&-avatar {
			width: 64px;
			height: 64px;
			border-radius: 50%;
			background: linear-gradient(135deg, rgba(255, 255, 255, 0.4), rgba(255, 255, 255, 0.15));
			border: 3px solid rgba(255, 255, 255, 0.6);
			display: flex;
			align-items: center;
			justify-content: center;
			color: #fff;
			font-size: 24px;
			font-weight: 700;
			box-shadow:
				0 4px 12px rgba(99, 102, 241, 0.3),
				0 2px 4px rgba(0, 0, 0, 0.1),
				inset 0 1px 0 rgba(255, 255, 255, 0.4);
		}

		&-info {
			flex: 1;
			display: flex;
			flex-direction: column;
			margin-left: var(--spacing-md);
		}

		&-name {
			font-size: 22px;
			font-weight: 800;
			color: #fff;
			letter-spacing: -0.3px;
		}

		&-id {
			font-size: 13px;
			color: rgba(255, 255, 255, 0.8);
			margin-top: 2px;
		}
	}

	&__name-row {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		margin-top: 6px;
	}

	&__level-tag {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 4px 10px;
		background: rgba(255, 255, 255, 0.2);
		border-radius: var(--radius-full);
		backdrop-filter: blur(10px);

		.user-page__level-tag-text {
			font-size: 11px;
			font-weight: 600;
			color: #fff;
		}
	}

	&__tags {
		display: flex;
		gap: var(--spacing-sm);
		margin-top: var(--spacing-md);
		flex-wrap: wrap;
	}

	&__tag {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 6px 12px;
		border-radius: var(--radius-full);
		font-size: 12px;
		font-weight: 600;
		background: rgba(255, 255, 255, 0.15);
		color: #fff;
		backdrop-filter: blur(10px);

		.user-page__tag-text {
			font-size: 12px;
			font-weight: 600;
			color: #fff;
		}
	}

	/* 统计卡片 */
	&__stats {
		background: rgba(255, 255, 255, 0.78);
		backdrop-filter: blur(20px) saturate(180%);
		-webkit-backdrop-filter: blur(20px) saturate(180%);
		border-radius: var(--radius-xl);
		padding: var(--spacing-lg) 0;
		margin: var(--spacing-lg) 0;
		display: flex;
		justify-content: space-around;
		align-items: center;
		box-shadow:
			0 4px 16px rgba(99, 102, 241, 0.08),
			0 8px 32px rgba(0, 0, 0, 0.06),
			inset 0 1px 0 rgba(255, 255, 255, 0.9);
		border: 1px solid rgba(255, 255, 255, 0.6);
		top:-130rpx;
		position: relative;
		&-item {
			flex: 1;
			display: flex;
			flex-direction: column;
			align-items: center;
			gap: 4px;

			&:active { opacity: 0.7; }
		}

		&-value {
			font-size: 22px;
			font-weight: 800;
			color: var(--text-primary);
		}

		&-label {
			font-size: 12px;
			color: var(--text-tertiary);
			font-weight: 500;
		}

		&-divider {
			width: 1px;
			height: 32px;
			background-color: var(--border-secondary);
		}
	}

	/* 菜单区块 */
	&__section {
		top:-130rpx;
		position: relative;
		margin-top: var(--spacing-lg);
		margin-bottom: var(--spacing-lg);
	}

	&__section-title {
		font-size: 13px;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--spacing-md);
		padding-left: var(--spacing-xs);
	}

	&__menu {
		background: rgba(255, 255, 255, 0.78);
		backdrop-filter: blur(20px) saturate(180%);
		-webkit-backdrop-filter: blur(20px) saturate(180%);
		border-radius: var(--radius-xl);
		overflow: hidden;
		box-shadow:
			0 4px 16px rgba(99, 102, 241, 0.06),
			0 8px 32px rgba(0, 0, 0, 0.05),
			inset 0 1px 0 rgba(255, 255, 255, 0.9);
		border: 1px solid rgba(255, 255, 255, 0.6);

		&-item {
			display: flex;
			align-items: center;
			padding: var(--spacing-md);
			transition: background-color 0.15s var(--ease-out);
			position: relative;

			&:not(:last-child)::after {
				content: '';
				position: absolute;
				bottom: 0;
				left: calc(40px + var(--spacing-md) * 2);
				right: 0;
				height: 1px;
				background-color: var(--border-secondary);
			}

			&:active {
				background-color: var(--bg-muted);
			}
		}

		&-icon {
			width: 40px;
			height: 40px;
			border-radius: var(--radius-md);
			display: flex;
			align-items: center;
			justify-content: center;
			margin-right: var(--spacing-md);
			
			&--notification { background: var(--warning-soft); color: var(--warning-color); }
			&--bookmark { background: var(--error-soft); color: var(--error-color); }
			&--score { background: var(--info-soft); color: var(--info-color); }
			&--bind { background: var(--success-soft); color: var(--success-color); }
			&--server { background: var(--bg-muted); color: var(--text-secondary); }
			&--setting { background: var(--bg-muted); color: var(--text-secondary); }
		}

		&-text {
			flex: 1;
			font-size: 15px;
			font-weight: 500;
			color: var(--text-primary);
		}

		&-badge {
			padding: 3px 8px;
			background-color: var(--error-color);
			color: #fff;
			font-size: 11px;
			font-weight: 700;
			border-radius: var(--radius-full);
			margin-right: var(--spacing-sm);
		}

		&-status {
			font-size: 13px;
			color: var(--text-tertiary);
			margin-right: var(--spacing-sm);
			
			&--online { color: var(--success-color); }
			&--offline { color: var(--error-color); }
		}
	}

	/* 底部 */
	&__footer {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--spacing-md);
		margin-top: var(--spacing-xl);
	}

	&__logout-btn {
		top:-130rpx;
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--spacing-sm);
		width: 100%;
		padding: 16px;
		background: rgba(255, 255, 255, 0.78);
		backdrop-filter: blur(20px) saturate(180%);
		-webkit-backdrop-filter: blur(20px) saturate(180%);
		border-radius: var(--radius-xl);
		border: 1px solid rgba(255, 255, 255, 0.6);
		box-shadow:
			0 4px 16px rgba(239, 68, 68, 0.06),
			0 8px 32px rgba(0, 0, 0, 0.05),
			inset 0 1px 0 rgba(255, 255, 255, 0.9);

		.user-page__logout-btn-text {
			font-size: 15px;
			font-weight: 600;
			color: var(--error-color);
		}

		&:active {
			background-color: var(--error-soft);
		}
	}

	&__login-btn {
		width: 100%;
		height: 52px;
		background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
		border-radius: var(--radius-xl);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow:
			0 4px 16px rgba(99, 102, 241, 0.25),
			0 8px 32px rgba(79, 70, 229, 0.2),
			inset 0 1px 0 rgba(255, 255, 255, 0.2);

		.user-page__login-btn-text {
			font-size: 16px;
			font-weight: 700;
			color: #fff;
		}

		&:active {
			transform: scale(0.98);
		}
	}

	&__version {
		top:-130rpx;
		position: relative;
		font-size: 12px;
		color: var(--text-tertiary);
		font-weight: 500;
		padding-bottom: var(--spacing-lg);
	}
}
</style>
