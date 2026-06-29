<template>
	<view class="post-content">
		<!-- 筛选栏 -->
		<view class="filter-bar">
			<scroll-view class="filter-tabs-scroll" scroll-x show-scrollbar="false">
				<view class="filter-tabs">
					<view
						v-for="(filter, index) in filters"
						:key="index"
						class="filter-tab"
						:class="{ active: currentFilter === index }"
						@tap="switchFilter(index)"
					>
				<l-icon
					:name="filter.icon"
					size="14"
					:color="currentFilter === index ? '#fff' : 'var(--text-secondary)'"
					class="filter-icon"
				></l-icon>
						<text class="filter-text">{{ filter.name }}</text>
					</view>
				</view>
			</scroll-view>
			<view class="create-btn" @tap="$emit('createPost')">
				<l-icon name="add" size="16" color="var(--primary-color)"></l-icon>
				<text class="create-text">发布动态</text>
			</view>
		</view>

		<!-- 帖子列表 -->
		<view v-if="posts.length > 0" class="posts-list">
			<view
				v-for="post in posts"
				:key="post.id"
				class="post-item"
				@tap="$emit('postClick', post)"
			>
				<!-- 帖子头部 -->
				<view class="post-header">
				<UserAvatar
					class="author-avatar"
					:src="post.author?.avatar"
					:name="post.author?.realname || post.author?.nickname"
					:size="80"
					@click="$emit('clickAuthor', post.author)"
				></UserAvatar>
			<view class="author-info">
				<view class="author-name-row">
					<text class="author-name" @tap.stop="$emit('clickAuthor', post.author)">{{ post.author?.realname || post.author?.nickname }}</text>
							<view v-if="post.isOfficial" class="official-badge">
								<l-icon name="star-filled" size="12" color="#fff"></l-icon>
								<text class="badge-text">官方</text>
							</view>
						</view>
						<view class="post-meta">
							<text class="post-time">{{ formatTime(post.publishedAt) }}</text>
							<text v-if="post.club" class="club-name">{{ post.club.name }}</text>
						</view>
					</view>
					<view class="header-actions">
						<view v-if="post.isTop" class="top-badge">
							<l-icon name="jump" size="16" color="var(--warning-color)"></l-icon>
						</view>
						<view class="action-menu">
							<view class="action-item" @tap.stop="toggleFavorite(post)">
								<l-icon
									:name="post.isFavorited ? 'star-filled' : 'star'"
									size="18"
									:color="post.isFavorited ? '#fbbf24' : 'var(--text-secondary)'"
								></l-icon>
							</view>
							<view class="action-item" @tap.stop="showManageMenu(post)">
								<l-icon name="more" size="18" color="var(--text-secondary)"></l-icon>
							</view>
						</view>
					</view>
				</view>

				<!-- 帖子内容 -->
				<view class="post-main">
					<text class="post-title">{{ post.title }}</text>
					<text v-if="post.summary" class="post-summary">{{ post.summary }}</text>

					<!-- 帖子图片 -->
					<view v-if="post.images && post.images.length > 0" class="post-images">
						<view class="images-grid" :class="getImageGridClass(post.images.length)">
							<image
								v-for="(image, index) in post.images.slice(0, 9)"
								:key="index"
								class="post-image"
								:src="image"
								mode="aspectFill"
								@tap.stop="previewImage(post.images, index)"
							></image>
						</view>
					</view>

					<!-- 帖子类型标签 -->
					<view class="post-type">
						<text class="type-tag" :class="getTypeClass(post.type)">
							{{ getTypeText(post.type) }}
						</text>
					</view>
				</view>

				<!-- 帖子底部 -->
				<view class="post-footer">
					<view class="post-stats">
						<view class="stat-item">
							<l-icon name="browse" size="14" color="var(--text-secondary)"></l-icon>
							<text class="stat-text">{{ formatNumber(post.viewsCount) }}</text>
						</view>
						<view class="stat-item">
							<l-icon name="chat" size="14" color="var(--text-secondary)"></l-icon>
							<text class="stat-text">{{ formatNumber(post.commentsCount) }}</text>
						</view>
					</view>
					<view class="post-actions">
						<view
							class="action-btn like-btn"
							:class="{ liked: post.isLiked }"
							@tap.stop="$emit('likePost', post)"
						>
<l-icon
							:name="post.isLiked ? 'heart-filled' : 'heart'"
							size="16"
							:color="post.isLiked ? '#ff4757' : 'var(--text-secondary)'"
						></l-icon>
							<text class="action-text">{{ formatNumber(post.likesCount) }}</text>
						</view>
						<view class="action-btn share-btn" @tap.stop="sharePost(post)">
							<l-icon name="refresh" size="16" color="var(--text-secondary)"></l-icon>
						</view>
					</view>
				</view>
			</view>
		</view>
	</view>

	<!-- 加载状态 -->
	<view v-if="loading" class="loading-container">
		<t-loading v-if="loading" theme="circular" size="40rpx" color="var(--primary-color)"></t-loading>
	</view>

	<!-- 空状态 -->
	<view v-if="!loading && posts.length === 0" class="empty-state">
		<image class="empty-image" src="/static/images/empty-posts.png" mode="aspectFit"></image>
		<text class="empty-text">暂无动态</text>
		<button class="create-btn-large" @tap="$emit('createPost')">发布动态</button>
	</view>
</template>

<script>
import UserAvatar from '@/components/UserAvatar.vue'
import lIcon from '@/uni_modules/lime-icon/components/l-icon/l-icon.vue'
import { usePostContent } from '@/composables/usePostContent.js';
import { useTimeFormat } from '@/composables/useTimeFormat.js';
import { getTypeText, getTypeClass, getImageGridClass } from '@/utils/postContent.js';

export default {
	components: {
		lIcon
	},
	props: {
		loading: {
			type: Boolean,
			default: false
		},
		posts: {
			type: Array,
			default: () => []
		}
	},
	emits: ['postClick', 'likePost', 'createPost', 'filterChange', 'edit', 'delete', 'report', 'block', 'clickAuthor'],
	setup(props, { emit }) {
		const { formatTime, formatNumber } = useTimeFormat();

		const {
			currentFilter,
			filters,
			switchFilter: switchFilterBase,
			previewImage,
			sharePost,
			toggleFavorite,
			showManageMenu: showManageMenuBase,
			confirmDelete: confirmDeleteBase,
			confirmBlock: confirmBlockBase
		} = usePostContent();

		function switchFilter(index) {
			switchFilterBase(index);
			emit('filterChange', filters.value[index]);
		}

		function showManageMenu(post) {
			const handlers = {
				onEdit: (p) => emit('edit', p),
				onDelete: (p) => emit('delete', p),
				onReport: (p) => emit('report', p),
				onBlock: (author) => emit('block', author)
			};
			showManageMenuBase(post, handlers);
		}

		return {
			currentFilter,
			filters,
			formatTime,
			formatNumber,
			switchFilter,
			previewImage,
			sharePost,
			toggleFavorite,
			showManageMenu,
			confirmDelete: confirmDeleteBase,
			confirmBlock: confirmBlockBase,
			getTypeText,
			getTypeClass,
			getImageGridClass
		};
	}
};
</script>

<style lang="scss" scoped>
.author-avatar {
	width: 80rpx;
	height: 80rpx;
	border-radius: 50%;
}

/* 微信小程序特定优化 */
::v-deep .filter-bar {
	padding: 8rpx;
	margin-bottom: 24rpx;
}

::v-deep .filter-tab {
	min-height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 6rpx;
	transform: translateZ(0);
}

::v-deep .filter-tab:active {
	transform: scale(0.95) translateZ(0) !important;
}

::v-deep .filter-tab.active {
	transform: scale(1.02) translateZ(0) !important;
}

::v-deep .filter-tabs-scroll {
	-webkit-overflow-scrolling: touch;
	scroll-behavior: smooth;
}

::v-deep .create-btn {
	min-height: 56rpx;
	max-width: 140rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 6rpx;
}

::v-deep .create-btn:active {
	transform: scale(0.95) !important;
	background: rgba(99, 102, 241, 0.2) !important;
}

::v-deep .create-text {
	position: relative;
	z-index: 2;
}

/* H5平台优化 */
::v-deep .filter-tab:hover {
	border-color: var(--primary-color);
	box-shadow: 0 2rpx 8rpx rgba(99, 102, 241, 0.15);
}

::v-deep .filter-tab.active:hover {
	box-shadow: 0 6rpx 20rpx rgba(99, 102, 241, 0.4);
}

::v-deep .create-btn:hover {
	background: rgba(99, 102, 241, 0.05);
	box-shadow: 0 4rpx 8rpx rgba(99, 102, 241, 0.15);
	transform: translateY(-1rpx);
}

/* App平台优化 */
::v-deep .filter-tab {
	min-height: 60rpx;
}

::v-deep .create-btn {
	min-height: 56rpx;
	max-width: 140rpx;
}

::v-deep .filter-tab:active,
::v-deep .create-btn:active {
	transition: transform 0.1s ease;
}
</style>
