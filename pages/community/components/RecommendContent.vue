<template>
	<view class="rc-root">
		<!-- 快捷入口区 - 改为左右全宽卡片 -->
		<view class="rc-quick-actions">
			<view class="rc-action-card" :class="'rc-action-card--' + (index === 0 ? 'post' : 'club')" v-for="(action, index) in quickActions" :key="index" @tap="handleQuickAction(action.type)">
				<view class="rc-action-card-bg"></view>
				<view class="rc-action-card-content">
					<view class="rc-action-icon-wrap">
						<l-icon :name="action.icon" size="28" color="#fff"></l-icon>
					</view>
					<view class="rc-action-text-wrap">
						<text class="rc-action-title">{{ action.title }}</text>
						<text class="rc-action-desc">{{ action.desc }}</text>
					</view>
				</view>
				<view class="rc-action-card-decoration">
					<text class="rc-action-decoration-text">{{ action.decoration }}</text>
				</view>
			</view>
		</view>

		<!-- 活跃社团 -->
		<view class="rc-section" v-if="clubs.length > 0">
			<view class="rc-section-header">
				<view class="rc-section-header-left">
					<view class="rc-section-title-group">
						<view class="rc-section-dot"></view>
						<text class="rc-section-title">活跃社团</text>
					</view>
					<text class="rc-section-sub">DISCOVER CLUBS</text>
				</view>
				<view class="rc-section-more" @tap="goToClubList">
					<text>查看全部</text>
					<l-icon name="chevron-right" size="16" color="var(--primary-500)"></l-icon>
				</view>
			</view>

			<scroll-view class="rc-clubs-scroll" scroll-x show-scrollbar="false" enhanced :bounces="true">
				<view class="rc-clubs-container">
					<view
						v-for="(club, index) in clubs"
						:key="club.id"
						class="rc-club-card"
						:style="{ animationDelay: (index * 0.05) + 's' }"
						@tap="$emit('clubClick', club)"
					>
						<view class="rc-club-logo-wrap">
							<ClubAvatar
								class="rc-club-logo"
								:src="club.logoUrl || club.avatar"
								:name="club.name"
								:size="80"
							></ClubAvatar>
							<view v-if="club.isOfficial" class="rc-club-verified">
								<l-icon name="check-circle-filled" size="12" color="#fff"></l-icon>
							</view>
						</view>
						<text class="rc-club-name">{{ club.name }}</text>
						<view class="rc-club-meta">
							<l-icon name="usergroup" size="10" color="var(--text-tertiary)"></l-icon>
							<text class="rc-club-meta-text">{{ club.memberCount || 0 }}</text>
						</view>
					</view>

					<!-- "查看全部"占位卡 -->
					<view class="rc-club-card rc-club-card--more" @tap="goToClubList">
						<view class="rc-club-logo-wrap rc-club-logo-wrap--more">
							<l-icon name="add" size="32" color="var(--text-tertiary)"></l-icon>
						</view>
						<text class="rc-club-name">更多社团</text>
					</view>
				</view>
			</scroll-view>
		</view>

		<!-- 校园动态 -->
		<view class="rc-section">
			<view class="rc-section-header">
				<view class="rc-section-header-left">
					<view class="rc-section-title-group">
						<view class="rc-section-dot rc-section-dot--orange"></view>
						<text class="rc-section-title">校园动态</text>
					</view>
					<text class="rc-section-sub">CAMPUS FEED</text>
				</view>
			</view>
			<PostList
				:posts="posts"
				:loading="loading"
				@postClick="$emit('postClick', $event)"
				@likePost="$emit('likePost', $event)"
				@clickAuthor="$emit('clickAuthor', $event)"
			/>
		</view>

		<!-- 空状态 -->
		<view v-if="!loading && clubs.length === 0 && posts.length === 0" class="rc-empty-state">
			<view class="rc-empty-illustration">
				<view class="rc-empty-circle rc-empty-circle--1"></view>
				<view class="rc-empty-circle rc-empty-circle--2"></view>
				<view class="rc-empty-icon-center">
					<l-icon name="file-text" size="56" color="var(--primary-400)"></l-icon>
				</view>
			</view>
			<text class="rc-empty-title">这里还空空如也</text>
			<text class="rc-empty-subtitle">去发布第一条动态吧</text>
			<view class="rc-empty-btn" @tap="$emit('createPost')">
				<l-icon name="edit" size="16" color="#fff"></l-icon>
				<text>发布动态</text>
			</view>
		</view>

		<view class="rc-bottom-safe"></view>
	</view>
</template>

<script setup>
import ClubAvatar from '@/components/ClubAvatar.vue'
import PostList from './PostList.vue'
import lIcon from '@/uni_modules/lime-icon/components/l-icon/l-icon.vue'

defineProps({
	loading: { type: Boolean, default: false },
	clubs: { type: Array, default: () => [] },
	posts: { type: Array, default: () => [] }
})

const emit = defineEmits(['createClub', 'createPost', 'clubClick', 'postClick', 'likePost', 'clickAuthor'])

const quickActions = [
	{ type: 'post', title: '发动态', desc: '分享校园生活', icon: 'edit', decoration: 'POST' },
	{ type: 'club', title: '创社团', desc: '创建精彩组织', icon: 'usergroup', decoration: 'CLUB' }
]

function handleQuickAction(type) {
	emit(type === 'post' ? 'createPost' : 'createClub')
}

function goToClubList() {
	uni.navigateTo({ url: '/pages/community/club-list' })
}
</script>

<style lang="scss" scoped>
.rc-root {
	padding-bottom: env(safe-area-inset-bottom);
}

/* Quick Actions - 全宽渐变卡片 */
.rc-quick-actions {
	display: flex;
	gap: 20rpx;
	padding: 24rpx;
	padding-top: 28rpx;
}

.rc-action-card {
	flex: 1;
	border-radius: 28rpx;
	overflow: hidden;
	position: relative;
	min-height: 150rpx;
	transition: transform 0.2s var(--ease-spring);

	&:active { transform: scale(0.97); }

	&--post {
		background: linear-gradient(145deg, #10b981 0%, #059669 100%);
		box-shadow: 0 8rpx 32rpx rgba(16, 185, 129, 0.35);

		.rc-action-card-bg {
			background: radial-gradient(circle at 80% 20%, rgba(255,255,255,0.15) 0%, transparent 60%);
		}
	}

	&--club {
		background: linear-gradient(145deg, var(--primary-600) 0%, #1d4ed8 100%);
		box-shadow: 0 8rpx 32rpx rgba(37, 99, 235, 0.35);

		.rc-action-card-bg {
			background: radial-gradient(circle at 80% 20%, rgba(255,255,255,0.15) 0%, transparent 60%);
		}
	}
}

.rc-action-card-bg {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	pointer-events: none;
}

.rc-action-card-content {
	position: relative;
	z-index: 1;
	padding: 24rpx;
	display: flex;
	flex-direction: column;
	height: 100%;
	justify-content: space-between;
	box-sizing: border-box;
}

.rc-action-icon-wrap {
	width: 72rpx;
	height: 72rpx;
	border-radius: 20rpx;
	background: rgba(255, 255, 255, 0.2);
	backdrop-filter: blur(8px);
	display: flex;
	align-items: center;
	justify-content: center;
}

.rc-action-text-wrap {
	display: flex;
	flex-direction: column;
	gap: 2rpx;
}

.rc-action-title {
	display: block;
	font-size: 32rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
}

.rc-action-desc {
	display: block;
	font-size: 22rpx;
	color: rgba(255, 255, 255, 0.75);
}

.rc-action-card-decoration {
	position: absolute;
	top: 16rpx;
	right: 16rpx;
	z-index: 1;
}

.rc-action-decoration-text {
	font-size: 14rpx;
	font-weight: 800;
	color: rgba(255, 255, 255, 0.25);
	letter-spacing: 2px;
	writing-mode: vertical-rl;
	text-orientation: mixed;
}

/* Section */
.rc-section {
	margin-top: 8rpx;
}

.rc-section-header {
	display: flex;
	justify-content: space-between;
	align-items: flex-end;
	padding: 0 24rpx 20rpx;
}

.rc-section-header-left {
	display: flex;
	flex-direction: column;
}

.rc-section-title-group {
	display: flex;
	align-items: center;
	gap: 10rpx;
}

.rc-section-dot {
	width: 8rpx;
	height: 32rpx;
	border-radius: 4rpx;
	background: linear-gradient(180deg, var(--primary-500), var(--primary-300));

	&--orange {
		background: linear-gradient(180deg, #f59e0b, #fbbf24);
	}
}

.rc-section-title {
	font-size: 34rpx;
	font-weight: 800;
	color: var(--text-primary);
	letter-spacing: -0.3px;
}

.rc-section-sub {
	font-size: 18rpx;
	font-weight: 600;
	color: var(--text-tertiary);
	letter-spacing: 1.5px;
	margin-top: 4rpx;
}

.rc-section-more {
	display: flex;
	align-items: center;
	gap: 4rpx;
	font-size: 24rpx;
	color: var(--primary-500);
	font-weight: 600;
	padding: 8rpx 16rpx;
	border-radius: 100rpx;
	transition: all 0.2s ease;

	&:active { background: rgba(59, 102, 241, 0.08); }
}

/* Clubs Scroll */
.rc-clubs-scroll {
	white-space: nowrap;
	padding: 0 24rpx;
}

.rc-clubs-container {
	display: inline-flex;
	gap: 20rpx;
}

.rc-club-card {
	width: 156rpx;
	background: #fff;
	border-radius: 24rpx;
	padding: 20rpx 12rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	margin-right: 20rpx;
	transition: all 0.2s var(--ease-spring);
	animation: rc-slideIn 0.4s ease-out both;
	flex-shrink: 0;

	&:active {
		transform: scale(0.95);
		background: #f8fafc;
	}

	&--more {
		background: linear-gradient(135deg, #f8fafc, #f1f5f9);
		border-style: dashed;
		border-color: rgba(148, 163, 184, 0.3);

		&:active { background: rgba(59, 102, 241, 0.05); }
	}
}

@keyframes rc-slideIn {
	from { opacity: 0; transform: translateX(-12rpx); }
	to { opacity: 1; transform: translateX(0); }
}

.rc-club-logo-wrap {
	position: relative;
	margin-bottom: 12rpx;

	&--more {
		width: 88rpx;
		height: 88rpx;
		border-radius: 24rpx;
		background: rgba(148, 163, 184, 0.1);
		border: 2px dashed rgba(148, 163, 184, 0.3);
		display: flex;
		align-items: center;
		justify-content: center;
	}
}

.rc-club-logo {
	width: 88rpx;
	height: 88rpx;
	border-radius: 24rpx;
	background: var(--bg-muted);
}

.rc-club-verified {
	position: absolute;
	bottom: -4rpx;
	right: -4rpx;
	width: 32rpx;
	height: 32rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	border: 3rpx solid #fff;
}

.rc-club-name {
	font-size: 26rpx;
	font-weight: 700;
	color: var(--text-primary);
	width: 100%;
	text-align: center;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
	margin-bottom: 8rpx;
}

.rc-club-meta {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 6rpx;
}

.rc-club-meta-text {
	font-size: 20rpx;
	color: var(--text-tertiary);
}

/* Empty State */
.rc-empty-state {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 80rpx 40rpx 40rpx;
	gap: 16rpx;
}

.rc-empty-illustration {
	position: relative;
	width: 200rpx;
	height: 200rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 16rpx;
}

.rc-empty-circle {
	position: absolute;
	border-radius: 50%;

	&--1 {
		width: 200rpx;
		height: 200rpx;
		background: linear-gradient(135deg, rgba(59, 102, 241, 0.06), rgba(59, 102, 241, 0.02));
		animation: rc-pulse 3s ease-in-out infinite;
	}

	&--2 {
		width: 140rpx;
		height: 140rpx;
		background: linear-gradient(135deg, rgba(59, 102, 241, 0.08), rgba(59, 102, 241, 0.03));
		animation: rc-pulse 3s ease-in-out infinite 0.5s;
	}
}

@keyframes rc-pulse {
	0%, 100% { transform: scale(1); opacity: 1; }
	50% { transform: scale(1.05); opacity: 0.8; }
}

.rc-empty-icon-center {
	position: relative;
	z-index: 1;
	width: 120rpx;
	height: 120rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, rgba(59, 102, 241, 0.1), rgba(59, 102, 241, 0.05));
	border: 1px solid rgba(59, 102, 241, 0.15);
	display: flex;
	align-items: center;
	justify-content: center;
}

.rc-empty-title {
	font-size: 34rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.rc-empty-subtitle {
	font-size: 26rpx;
	color: var(--text-tertiary);
}

.rc-empty-btn {
	display: flex;
	align-items: center;
	gap: 10rpx;
	margin-top: 8rpx;
	padding: 20rpx 48rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	color: #fff;
	border-radius: 44rpx;
	font-size: 28rpx;
	font-weight: 700;
	box-shadow: 0 8rpx 28rpx rgba(37, 99, 235, 0.3);
	transition: all 0.2s var(--ease-spring);

	&:active {
		transform: translateY(2rpx) scale(0.98);
		box-shadow: 0 4rpx 16rpx rgba(37, 99, 235, 0.25);
	}
}

.rc-bottom-safe {
	height: env(safe-area-inset-bottom);
}
</style>
