<template>
	<view class="cc-root">
		<!-- 我的社团 -->
		<view class="cc-section" v-if="myClubs.length > 0">
			<view class="cc-section-header">
				<view class="cc-section-header-left">
					<view class="cc-section-title-group">
						<view class="cc-section-dot"></view>
						<text class="cc-section-title">我的社团</text>
					</view>
					<text class="cc-section-sub">MY CLUBS</text>
				</view>
				<view class="cc-section-more" @tap="goToMyClubs">
					<text>查看全部</text>
					<l-icon name="chevron-right" size="16" color="var(--primary-500)"></l-icon>
				</view>
			</view>
			<scroll-view class="cc-my-clubs-scroll" scroll-x show-scrollbar="false" enhanced :bounces="true">
				<view class="cc-my-clubs-container">
					<view
						v-for="(club, index) in myClubs"
						:key="club.id"
						class="cc-my-club-card"
						:style="{ animationDelay: (index * 0.06) + 's' }"
						@tap="$emit('clubClick', club)"
					>
						<view class="cc-my-club-logo-wrap">
							<ClubAvatar
								class="cc-my-club-logo"
								:src="club.logoUrl || club.avatar"
								:name="club.name"
								:size="100"
							></ClubAvatar>
						</view>
						<text class="cc-my-club-name">{{ club.name }}</text>
						<view class="cc-my-club-badge" :class="getRoleClass(club.role)">
							{{ getRoleText(club.role) }}
						</view>
					</view>
				</view>
			</scroll-view>
		</view>

		<!-- 发现社团 -->
		<view class="cc-section">
			<!-- 筛选栏 -->
			<view class="cc-filter-bar">
				<scroll-view class="cc-filter-tabs-scroll" scroll-x show-scrollbar="false">
					<view class="cc-filter-tabs">
						<view
							v-for="(filter, index) in filters"
							:key="index"
							class="cc-filter-tab"
							:class="{ 'cc-filter-tab--active': currentFilter === index }"
							@tap="switchFilter(index)"
						>
							<text class="cc-filter-text">{{ filter.name }}</text>
						</view>
					</view>
				</scroll-view>
				<view class="cc-search-box" @tap="goToSearch">
					<l-icon name="search" size="18" color="var(--text-tertiary)"></l-icon>
					<text class="cc-search-placeholder">搜索社团...</text>
				</view>
			</view>

			<!-- 社团列表 -->
			<view class="cc-clubs-list">
				<view
					v-for="(club, index) in clubs"
					:key="club.id"
					class="cc-club-item"
					:style="{ animationDelay: (index * 0.06) + 's' }"
					@tap="$emit('clubClick', club)"
				>
					<view class="cc-club-avatar-wrap">
						<ClubAvatar
							class="cc-club-avatar"
							:src="club.logoUrl || club.avatar"
							:name="club.name"
							:size="96"
						></ClubAvatar>
						<view v-if="club.isOfficial" class="cc-club-avatar-badge">
							<l-icon name="check-circle-filled" size="12" color="#fff"></l-icon>
						</view>
					</view>
					<view class="cc-club-main">
						<view class="cc-club-header">
							<text class="cc-club-name">{{ club.name }}</text>
							<view v-if="club.isOfficial" class="cc-official-tag">
								<text class="cc-official-tag-text">官方</text>
							</view>
						</view>
						<text class="cc-club-desc">{{ club.description || '暂无介绍' }}</text>
						<view class="cc-club-footer">
							<view class="cc-club-stats">
								<view class="cc-stat-item">
									<l-icon name="usergroup" size="12" color="var(--text-tertiary)"></l-icon>
									<text class="cc-stat-item-text">{{ club.memberCount || 0 }} 成员</text>
								</view>
								<view class="cc-stat-item">
									<l-icon name="file-text" size="12" color="var(--text-tertiary)"></l-icon>
									<text class="cc-stat-item-text">{{ club.postCount || 0 }} 动态</text>
								</view>
							</view>
							<view
								class="cc-join-btn"
								:class="{ 'cc-join-btn--joined': club.isMember }"
								@tap.stop="handleJoin(club)"
							>
								<text>{{ club.isMember ? '已加入' : '加入' }}</text>
							</view>
						</view>
					</view>
				</view>
			</view>
		</view>

		<!-- Loading -->
		<view v-if="loading" class="cc-loading">
			<view class="cc-loading-spinner"></view>
			<text class="cc-loading-text">加载中...</text>
		</view>

		<!-- 空状态 -->
		<view v-if="!loading && clubs.length === 0" class="cc-empty-state">
			<view class="cc-empty-illustration">
				<view class="cc-empty-circle cc-empty-circle--1"></view>
				<view class="cc-empty-circle cc-empty-circle--2"></view>
				<view class="cc-empty-icon-center">
					<l-icon name="usergroup" size="56" color="var(--primary-400)"></l-icon>
				</view>
			</view>
			<text class="cc-empty-title">暂时没有发现更多社团</text>
			<text class="cc-empty-subtitle">创建你的第一个社团吧</text>
			<view class="cc-empty-btn" @tap="$emit('createClub')">
				<l-icon name="add" size="16" color="#fff"></l-icon>
				<text>创建社团</text>
			</view>
		</view>

		<view class="cc-bottom-safe"></view>
	</view>
</template>

<script>
import ClubAvatar from '@/components/ClubAvatar.vue'
import lIcon from '@/uni_modules/lime-icon/components/l-icon/l-icon.vue'

export default {
	components: { ClubAvatar, lIcon },
	props: {
		loading: { type: Boolean, default: false },
		clubs: { type: Array, default: () => [] },
		myClubs: { type: Array, default: () => [] }
	},
	data() {
		return {
			currentFilter: 0,
			filters: [
				{ name: '全部', key: 'all' },
				{ name: '热门', key: 'hot' },
				{ name: '最新', key: 'latest' },
				{ name: '官方', key: 'official' }
			]
		}
	},
	emits: ['filterChange', 'clubClick', 'joinClub', 'search', 'createClub'],
	methods: {
		switchFilter(index) {
			this.currentFilter = index
			this.$emit('filterChange', this.filters[index])
		},
		getRoleText(role) {
			const roleMap = { 'admin': '管理员', 'creator': '创建者', 'member': '成员' }
			return roleMap[role] || '成员'
		},
		getRoleClass(role) {
			if (role === 'admin' || role === 'creator') return 'cc-role-admin'
			return 'cc-role-member'
		},
		handleJoin(club) {
			this.$emit('joinClub', club)
		},
		goToMyClubs() {
			uni.navigateTo({ url: '/pages/community/club-list?type=my' })
		},
		goToSearch() {
			this.$emit('search')
		}
	}
}
</script>

<style lang="scss" scoped>
.cc-root { padding-bottom: env(safe-area-inset-bottom); }
.cc-bottom-safe { height: env(safe-area-inset-bottom); }

/* Section */
.cc-section { margin-bottom: 24rpx; }

.cc-section-header {
	display: flex;
	justify-content: space-between;
	align-items: flex-end;
	padding: 24rpx 24rpx 20rpx;
}

.cc-section-header-left { display: flex; flex-direction: column; }

.cc-section-title-group {
	display: flex;
	align-items: center;
	gap: 10rpx;
}

.cc-section-dot {
	width: 8rpx;
	height: 32rpx;
	border-radius: 4rpx;
	background: linear-gradient(180deg, var(--primary-500), var(--primary-300));
}

.cc-section-title {
	font-size: 34rpx;
	font-weight: 800;
	color: var(--text-primary);
	letter-spacing: -0.3px;
}

.cc-section-sub {
	font-size: 18rpx;
	font-weight: 600;
	color: var(--text-tertiary);
	letter-spacing: 1.5px;
	margin-top: 4rpx;
}

.cc-section-more {
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

/* My Clubs */
.cc-my-clubs-scroll { width: 100%; white-space: nowrap; }
.cc-my-clubs-container { display: inline-flex; padding: 0 24rpx 24rpx; gap: 20rpx; }

.cc-my-club-card {
	width: 156rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	background: #fff;
	border-radius: 24rpx;
	padding: 20rpx 12rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	animation: cc-slideIn 0.4s ease-out both;
	flex-shrink: 0;
	transition: all 0.2s var(--ease-spring);
	&:active { transform: scale(0.97); background: #f8fafc; }
}

@keyframes cc-slideIn {
	from { opacity: 0; transform: translateX(-12rpx); }
	to { opacity: 1; transform: translateX(0); }
}

.cc-my-club-logo-wrap { margin-bottom: 12rpx; }
.cc-my-club-logo {
	width: 96rpx;
	height: 96rpx;
	border-radius: 50%;
	background: var(--bg-muted);
	border: 4rpx solid rgba(148, 163, 184, 0.15);
}
.cc-my-club-name {
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
.cc-my-club-badge {
	font-size: 20rpx;
	padding: 4rpx 14rpx;
	border-radius: 100rpx;
	font-weight: 600;
	&.cc-role-admin, &.cc-role-creator { background: rgba(59, 102, 241, 0.1); color: var(--primary-500); }
	&.cc-role-member { background: rgba(148, 163, 184, 0.1); color: var(--text-tertiary); }
}

/* Filter Bar */
.cc-filter-bar {
	position: sticky;
	top: -2rpx;
	z-index: 100;
	background: var(--bg-secondary);
	padding: 16rpx 24rpx 12rpx;
}

.cc-filter-tabs-scroll { white-space: nowrap; margin-bottom: 12rpx; }
.cc-filter-tabs { display: inline-flex; gap: 12rpx; }

.cc-filter-tab {
	position: relative;
	padding: 10rpx 24rpx;
	border-radius: 100rpx;
	background: rgba(148, 163, 184, 0.08);
	transition: all 0.2s var(--ease-spring);

	&--active {
		background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
		box-shadow: 0 4rpx 16rpx rgba(37, 99, 235, 0.3);
		.cc-filter-text { color: #fff; font-weight: 700; }
	}

	&:active:not(.cc-filter-tab--active) { background: rgba(148, 163, 184, 0.16); }
}

.cc-filter-text {
	font-size: 26rpx;
	color: var(--text-secondary);
	font-weight: 500;
	transition: color 0.2s ease;
}

.cc-search-box {
	background: #fff;
	height: 72rpx;
	border-radius: 36rpx;
	display: flex;
	align-items: center;
	padding: 0 24rpx;
	gap: 12rpx;
	border: 1px solid rgba(148, 163, 184, 0.15);
	box-shadow: 0 2rpx 8rpx rgba(30, 64, 175, 0.04);
	transition: all 0.2s ease;
	&:active { background: #f8fafc; border-color: rgba(59, 102, 241, 0.3); }
}
.cc-search-placeholder { font-size: 26rpx; color: var(--text-tertiary); }

/* Club List */
.cc-clubs-list { padding: 0 24rpx; }

.cc-club-item {
	background: #fff;
	border-radius: 24rpx;
	padding: 24rpx;
	margin-bottom: 16rpx;
	display: flex;
	gap: 20rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	transition: all 0.2s var(--ease-out);
	animation: cc-fadeIn 0.4s ease-out both;
	&:active { transform: scale(0.99); background: #f8fafc; }
}

@keyframes cc-fadeIn {
	from { opacity: 0; transform: translateY(10rpx); }
	to { opacity: 1; transform: translateY(0); }
}

.cc-club-avatar-wrap { position: relative; flex-shrink: 0; }
.cc-club-avatar {
	width: 120rpx;
	height: 120rpx;
	border-radius: 24rpx;
	background: var(--bg-muted);
}
.cc-club-avatar-badge {
	position: absolute;
	bottom: -6rpx;
	right: -6rpx;
	width: 36rpx;
	height: 36rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	border: 3rpx solid #fff;
}

.cc-club-main {
	flex: 1;
	display: flex;
	flex-direction: column;
	justify-content: space-between;
	overflow: hidden;
}

.cc-club-header {
	display: flex;
	align-items: center;
	gap: 10rpx;
	margin-bottom: 6rpx;
}
.cc-club-name {
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}
.cc-official-tag {
	padding: 3rpx 10rpx;
	border-radius: 8rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	flex-shrink: 0;
	.cc-official-tag-text { font-size: 18rpx; color: #fff; font-weight: 600; }
}

.cc-club-desc {
	font-size: 24rpx;
	color: var(--text-secondary);
	display: -webkit-box;
	-webkit-box-orient: vertical;
	-webkit-line-clamp: 2;
	line-clamp: 2;
	overflow: hidden;
	margin-bottom: 8rpx;
}

.cc-club-footer {
	display: flex;
	justify-content: space-between;
	align-items: center;
}
.cc-club-stats { display: flex; gap: 20rpx; }
.cc-stat-item {
	display: flex;
	align-items: center;
	gap: 6rpx;
	.cc-stat-item-text { font-size: 22rpx; color: var(--text-tertiary); }
}

.cc-join-btn {
	padding: 8rpx 28rpx;
	border-radius: 100rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	color: #fff;
	font-size: 24rpx;
	font-weight: 700;
	box-shadow: 0 4rpx 12rpx rgba(37, 99, 235, 0.25);
	transition: all 0.15s var(--ease-spring);
	&:active { transform: scale(0.95); }
	&--joined {
		background: rgba(148, 163, 184, 0.12);
		color: var(--text-secondary);
		box-shadow: none;
	}
}

/* Loading */
.cc-loading {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 48rpx 0;
	gap: 12rpx;
}
.cc-loading-spinner {
	width: 48rpx;
	height: 48rpx;
	border: 4rpx solid rgba(99, 102, 241, 0.12);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: cc-spin 0.8s linear infinite;
}
@keyframes cc-spin { to { transform: rotate(360deg); } }
.cc-loading-text { font-size: 24rpx; color: var(--text-tertiary); }

/* Empty State */
.cc-empty-state {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 80rpx 40rpx 40rpx;
	gap: 16rpx;
}
.cc-empty-illustration {
	position: relative;
	width: 200rpx;
	height: 200rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 16rpx;
}
.cc-empty-circle {
	position: absolute;
	border-radius: 50%;
	&--1 { width: 200rpx; height: 200rpx; background: rgba(59, 102, 241, 0.05); animation: cc-emp-pulse 3s ease-in-out infinite; }
	&--2 { width: 140rpx; height: 140rpx; background: rgba(59, 102, 241, 0.08); animation: cc-emp-pulse 3s ease-in-out infinite 0.6s; }
}
@keyframes cc-emp-pulse {
	0%, 100% { transform: scale(1); opacity: 1; }
	50% { transform: scale(1.06); opacity: 0.7; }
}
.cc-empty-icon-center {
	position: relative;
	z-index: 1;
	width: 120rpx;
	height: 120rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, rgba(59, 102, 241, 0.1), rgba(59, 102, 241, 0.04));
	border: 1px solid rgba(59, 102, 241, 0.15);
	display: flex;
	align-items: center;
	justify-content: center;
}
.cc-empty-title { font-size: 34rpx; font-weight: 800; color: var(--text-primary); }
.cc-empty-subtitle { font-size: 26rpx; color: var(--text-tertiary); }

.cc-empty-btn {
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
</style>
