<template>
	<view class="hc-root">
		<!-- 排序选项 -->
		<view class="hc-sort-bar" v-if="showSortOptions">
			<scroll-view class="hc-sort-tabs" scroll-x show-scrollbar="false">
				<view class="hc-sort-tabs-inner">
					<view
						v-for="option in sortOptions"
						:key="option.value"
						class="hc-sort-tab"
						:class="{ 'hc-sort-tab--active': currentSort === option.value }"
						@tap="changeSortOrder(option.value)"
					>
						<l-icon v-if="option.icon" :name="option.icon" size="12" :color="currentSort === option.value ? '#fff' : 'var(--text-tertiary)'"></l-icon>
						<text class="hc-sort-text">{{ option.label }}</text>
					</view>
				</view>
			</scroll-view>
		</view>

		<!-- 热门评论 -->
		<view v-if="hotComments.length > 0 && showHotComments" class="hc-hot-section">
			<view class="hc-hot-header">
				<view class="hc-hot-title-group">
					<view class="hc-hot-dot"></view>
					<text class="hc-hot-title">热门评论</text>
					<view class="hc-hot-count">{{ hotComments.length }}</view>
				</view>
			</view>
			<view class="hc-hot-list">
				<comment-item
					v-for="(comment, index) in hotComments"
					:key="`hot-${comment.id}`"
					:comment="comment"
					:level="0"
					:is-hot="true"
					:style="{ animationDelay: (index * 0.06) + 's' }"
					@reply="handleReply"
					@like="handleLike"
					@load-replies="loadReplies"
					@show-more="showMoreReplies"
					@click-author="author => $emit('clickAuthor', author)"
				/>
			</view>
		</view>

		<!-- 全部评论 -->
		<view class="hc-all-section">
			<!-- Skeleton Loading -->
			<template v-if="loading && comments.length === 0">
				<view v-for="i in 3" :key="'sk-'+i" class="hc-skeleton">
					<view class="hc-skeleton-avatar"></view>
					<view class="hc-skeleton-body">
						<view class="hc-skeleton-line hc-skeleton-line--w40"></view>
						<view class="hc-skeleton-line hc-skeleton-line--w80 hc-skeleton-line--mt8"></view>
						<view class="hc-skeleton-line hc-skeleton-line--w60 hc-skeleton-line--mt6"></view>
					</view>
				</view>
			</template>

			<!-- 评论列表 -->
			<template v-else>
				<view v-if="comments.length > 0" class="hc-comments-header">
					<view class="hc-comments-title-group">
						<view class="hc-comments-dot"></view>
						<text class="hc-comments-title">全部评论</text>
					</view>
					<text class="hc-comments-count">{{ comments.length }} 条</text>
				</view>

				<view class="hc-comments-list">
				<comment-item
					v-for="(comment, index) in comments"
					:key="comment.id"
					:comment="comment"
					:level="0"
					:style="{ animationDelay: (index * 0.06) + 's' }"
					@reply="handleReply"
					@like="handleLike"
					@load-replies="loadReplies"
					@show-more="showMoreReplies"
					@click-author="author => $emit('clickAuthor', author)"
				/>
				</view>

				<!-- 加载更多 -->
				<view v-if="hasMore" class="hc-load-more" @tap="loadMore">
					<view class="hc-load-more-inner">
						<view class="hc-load-more-spinner"></view>
						<text class="hc-load-more-text">加载更多</text>
					</view>
				</view>

				<!-- 空状态 -->
				<view v-if="comments.length === 0 && !loading" class="hc-empty-state">
					<view class="hc-empty-illustration">
						<view class="hc-empty-circle hc-empty-circle--1"></view>
						<view class="hc-empty-circle hc-empty-circle--2"></view>
						<view class="hc-empty-icon-center">
							<l-icon name="chat" size="52" color="var(--primary-400)"></l-icon>
						</view>
					</view>
					<text class="hc-empty-title">暂无评论</text>
					<text class="hc-empty-subtitle">快来抢沙发吧</text>
				</view>
			</template>
		</view>

		<!-- 全局加载中 -->
		<view v-if="loading && comments.length > 0" class="hc-loading-more">
			<view class="hc-loading-spinner"></view>
			<text class="hc-loading-text">加载中...</text>
		</view>

		<!-- 错误状态 -->
		<view v-if="error && !loading" class="hc-error-state">
			<view class="hc-error-illustration">
				<l-icon name="info-circle-filled" size="48" color="var(--error-color)"></l-icon>
			</view>
			<text class="hc-error-title">加载失败</text>
			<text class="hc-error-subtitle">{{ error }}</text>
			<view class="hc-error-btn" @tap="retry">
				<l-icon name="refresh" size="14" color="#fff"></l-icon>
				<text>重新加载</text>
			</view>
		</view>
	</view>
</template>

<script>
import {
	getCommentsList,
	getHotComments,
	getCommentReplies,
	likeComment,
	unlikeComment
} from '@/pages/api/community.js';
import { showToast } from '@/pages/api/page.js';
import CommentItem from './CommentItem.vue';

export default {
	name: 'HierarchicalComments',
	components: { CommentItem },
	props: {
		postId: { type: [String, Number], required: true },
		showSortOptions: { type: Boolean, default: true },
		showHotComments: { type: Boolean, default: true },
		pageSize: { type: Number, default: 20 }
	},
	data() {
		return {
			comments: [],
			hotComments: [],
			loading: false,
			hasMore: true,
			currentPage: 1,
			currentSort: 'time',
			error: null,
			retryCount: 0,
			maxRetries: 3,
			sortOptions: [
				{ label: '时间正序', value: 'time', icon: 'clock' },
				{ label: '最新优先', value: 'time_desc', icon: 'new' },
				{ label: '热门优先', value: 'hot', icon: 'local-filled' },
				{ label: '点赞优先', value: 'likes', icon: 'heart-filled' }
			]
		}
	},
	mounted() {
		this.loadComments();
		if (this.showHotComments) this.loadHotComments();
	},
	methods: {
		async loadComments(refresh = false) {
			if (this.loading) return;
			this.loading = true;
			this.error = null;

			if (refresh) {
				this.currentPage = 1;
				this.comments = [];
				this.hasMore = true;
				this.retryCount = 0;
			}

			try {
				const result = await getCommentsList(this.postId, {
					page: this.currentPage,
					pageSize: this.pageSize,
					sortBy: this.currentSort
				});

				if (result.success) {
					const newComments = (result.result.records || []).filter(c => c.status === 1);
					newComments.forEach(c => this.processCommentData(c));
					if (refresh) this.comments = newComments;
					else this.comments.push(...newComments);
					this.hasMore = newComments.length >= this.pageSize;
					this.currentPage++;
					this.retryCount = 0;
				} else {
					throw new Error(result.message || '加载评论失败');
				}
			} catch (error) {
				console.error('加载评论失败:', error);
				this.error = error.message || '加载评论失败';
				if (!refresh && this.retryCount < this.maxRetries) {
					this.retryCount++;
					setTimeout(() => this.loadComments(false), 2000 * this.retryCount);
				} else {
					showToast({ title: this.error, icon: 'none', duration: 3000 });
				}
			} finally {
				this.loading = false;
			}
		},

		async loadHotComments() {
			try {
				const result = await getHotComments(this.postId, { limit: 3 });
				if (result.success) {
					this.hotComments = (result.result || []).map(comment => {
						this.processCommentData(comment);
						return comment;
					});
				}
			} catch (error) {
				console.error('加载热门评论失败:', error);
			}
		},

		processCommentData(comment) {
			if (comment.images && typeof comment.images === 'string') {
				try { comment.images = JSON.parse(comment.images); } catch { comment.images = []; }
			} else if (!comment.images) { comment.images = []; }

			if (comment.mentionedUsers && typeof comment.mentionedUsers === 'string') {
				try { comment.mentionedUsers = JSON.parse(comment.mentionedUsers); } catch { comment.mentionedUsers = []; }
			} else if (!comment.mentionedUsers) { comment.mentionedUsers = []; }

			if (!comment.replies) comment.replies = [];
			comment.showReplies = false;
			comment.repliesLoaded = false;
		},

		async loadReplies(comment) {
			if (comment.repliesLoaded || comment.repliesCount === 0) {
				comment.showReplies = !comment.showReplies;
				return;
			}
			try {
				const result = await getCommentReplies(comment.id, { page: 1, pageSize: 10, sortBy: 'time' });
				if (result.success) {
					const replies = result.result.records || [];
					replies.forEach(r => this.processCommentData(r));
					comment.replies = replies;
					comment.repliesLoaded = true;
					comment.showReplies = true;
				}
			} catch (error) {
				console.error('加载回复失败:', error);
				showToast({ title: '加载失败', icon: 'none' });
			}
		},

		showMoreReplies(comment) { this.$emit('showMoreReplies', comment); },

		handleReply(comment, parentComment = null) {
			this.$emit('reply', { comment, parentComment, postId: this.postId });
		},

		async handleLike(comment) {
			try {
				const result = comment.isLiked ? await unlikeComment(comment.id) : await likeComment(comment.id);
				if (result && result.success) {
					comment.isLiked = !comment.isLiked;
					comment.likesCount = (comment.likesCount || 0) + (comment.isLiked ? 1 : -1);
				}
			} catch (error) {
				console.error('点赞失败:', error);
				uni.showToast({ title: '操作失败', icon: 'none' });
			}
		},

		changeSortOrder(sortBy) {
			if (this.currentSort === sortBy) return;
			this.currentSort = sortBy;
			this.loadComments(true);
		},

		loadMore() {
			if (this.hasMore && !this.loading) this.loadComments();
		},

		refresh() {
			this.loadComments(true);
			if (this.showHotComments) this.loadHotComments();
		},

		retry() {
			this.error = null;
			this.retryCount = 0;
			this.loadComments(true);
		}
	}
}
</script>

<style lang="scss" scoped>
.hc-root {
	background: var(--bg-primary);
}

/* Sort Bar */
.hc-sort-bar {
	background: var(--bg-primary);
	border-bottom: 1px solid rgba(226, 232, 240, 0.6);
}

.hc-sort-tabs { white-space: nowrap; }
.hc-sort-tabs-inner { display: inline-flex; gap: 12rpx; }

.hc-sort-tab {
	display: inline-flex;
	align-items: center;
	gap: 6rpx;
	padding: 10rpx 20rpx;
	border-radius: 100rpx;
	background: rgba(148, 163, 184, 0.08);
	transition: all 0.2s var(--ease-spring);
	flex-shrink: 0;

	&:active { transform: scale(0.97); }

	&--active {
		background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
		box-shadow: 0 4rpx 16rpx rgba(37, 99, 235, 0.3);

		.hc-sort-text { color: #fff; font-weight: 700; }
	}
}

.hc-sort-text {
	font-size: 24rpx;
	color: var(--text-secondary);
	font-weight: 500;
	transition: color 0.2s ease;
}

/* Hot Section */
.hc-hot-section {
	margin-bottom: 8rpx;
	background: linear-gradient(180deg, rgba(255, 107, 53, 0.03) 0%, transparent 100%);
	border-bottom: 1px solid rgba(226, 232, 240, 0.6);
}

.hc-hot-header {
	padding: 20rpx 24rpx 16rpx;
}

.hc-hot-title-group {
	display: flex;
	align-items: center;
	gap: 10rpx;
}

.hc-hot-dot {
	width: 8rpx;
	height: 32rpx;
	border-radius: 4rpx;
	background: linear-gradient(180deg, #f59e0b, #fbbf24);
}

.hc-hot-title {
	font-size: 32rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.hc-hot-count {
	font-size: 22rpx;
	color: #f59e0b;
	font-weight: 700;
	background: rgba(245, 158, 11, 0.1);
	padding: 2rpx 12rpx;
	border-radius: 100rpx;
}

.hc-hot-list {
	padding-bottom: 4rpx;
}

/* All Section */
.hc-comments-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 20rpx 24rpx 16rpx;
	border-bottom: 1px solid rgba(226, 232, 240, 0.6);
}

.hc-comments-title-group {
	display: flex;
	align-items: center;
	gap: 10rpx;
}

.hc-comments-dot {
	width: 8rpx;
	height: 32rpx;
	border-radius: 4rpx;
	background: linear-gradient(180deg, var(--primary-500), var(--primary-300));
}

.hc-comments-title {
	font-size: 32rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.hc-comments-count {
	font-size: 24rpx;
	color: var(--text-tertiary);
	font-weight: 600;
}

/* Skeleton */
.hc-skeleton {
	display: flex;
	gap: 20rpx;
	padding: 32rpx 24rpx;
	border-bottom: 1px solid rgba(226, 232, 240, 0.5);
}

.hc-skeleton-avatar {
	width: 80rpx;
	height: 80rpx;
	border-radius: 50%;
	flex-shrink: 0;
	background: linear-gradient(90deg, var(--bg-muted) 25%, rgba(148, 163, 184, 0.12) 50%, var(--bg-muted) 75%);
	background-size: 200% 100%;
	animation: hc-shimmer 1.5s infinite;
}

.hc-skeleton-body {
	flex: 1;
	display: flex;
	flex-direction: column;
	justify-content: center;
	gap: 8rpx;
}

.hc-skeleton-line {
	height: 22rpx;
	border-radius: 8rpx;
	background: linear-gradient(90deg, var(--bg-muted) 25%, rgba(148, 163, 184, 0.12) 50%, var(--bg-muted) 75%);
	background-size: 200% 100%;
	animation: hc-shimmer 1.5s infinite;

	&--w40 { width: 40%; }
	&--w60 { width: 60%; }
	&--w80 { width: 80%; }
	&--mt6 { margin-top: 6rpx; }
	&--mt8 { margin-top: 8rpx; }
}

@keyframes hc-shimmer {
	0% { background-position: 200% 0; }
	100% { background-position: -200% 0; }
}

/* Load More */
.hc-load-more {
	padding: 32rpx 24rpx;
	display: flex;
	justify-content: center;
}

.hc-load-more-inner {
	display: flex;
	align-items: center;
	gap: 10rpx;
	padding: 14rpx 32rpx;
	border-radius: 100rpx;
	background: rgba(59, 102, 241, 0.08);
	transition: all 0.2s ease;

	&:active {
		transform: scale(0.98);
		background: rgba(59, 102, 241, 0.12);
	}
}

.hc-load-more-spinner {
	width: 32rpx;
	height: 32rpx;
	border: 3rpx solid rgba(59, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: hc-spin 0.8s linear infinite;
}

@keyframes hc-spin { to { transform: rotate(360deg); } }

.hc-load-more-text {
	font-size: 26rpx;
	color: var(--primary-500);
	font-weight: 600;
}

.hc-loading-more {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 12rpx;
	padding: 32rpx;
}

.hc-loading-spinner {
	width: 40rpx;
	height: 40rpx;
	border: 3rpx solid rgba(59, 102, 241, 0.12);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: hc-spin 0.8s linear infinite;
}

.hc-loading-text {
	font-size: 24rpx;
	color: var(--text-tertiary);
}

/* Empty State */
.hc-empty-state {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 80rpx 40rpx 40rpx;
	gap: 16rpx;
}

.hc-empty-illustration {
	position: relative;
	width: 200rpx;
	height: 200rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 8rpx;
}

.hc-empty-circle {
	position: absolute;
	border-radius: 50%;
	&--1 { width: 200rpx; height: 200rpx; background: rgba(59, 102, 241, 0.05); animation: hc-emp-pulse 3s ease-in-out infinite; }
	&--2 { width: 140rpx; height: 140rpx; background: rgba(59, 102, 241, 0.08); animation: hc-emp-pulse 3s ease-in-out infinite 0.6s; }
}

@keyframes hc-emp-pulse {
	0%, 100% { transform: scale(1); opacity: 1; }
	50% { transform: scale(1.06); opacity: 0.7; }
}

.hc-empty-icon-center {
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

.hc-empty-title {
	font-size: 34rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.hc-empty-subtitle {
	font-size: 26rpx;
	color: var(--text-tertiary);
}

/* Error State */
.hc-error-state {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 60rpx 40rpx;
	gap: 16rpx;
}

.hc-error-illustration {
	width: 100rpx;
	height: 100rpx;
	border-radius: 50%;
	background: rgba(239, 68, 68, 0.08);
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 8rpx;
}

.hc-error-title {
	font-size: 32rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.hc-error-subtitle {
	font-size: 26rpx;
	color: var(--text-tertiary);
	text-align: center;
}

.hc-error-btn {
	display: flex;
	align-items: center;
	gap: 10rpx;
	margin-top: 8rpx;
	padding: 16rpx 40rpx;
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
