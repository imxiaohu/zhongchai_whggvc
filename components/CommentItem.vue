<template>
	<view class="ci-root" :class="rootClasses" :style="rootStyle">
		<view class="ci-main">
			<UserAvatar
				class="ci-avatar"
				:src="comment.user?.avatar"
				:name="getUserDisplayName(comment.user)"
				:size="level > 0 ? 56 : 72"
				@click="$emit('clickAuthor', comment.user)"
			></UserAvatar>

			<view class="ci-body">
				<view class="ci-header">
					<view class="ci-author-row">
						<text class="ci-author-name">{{ getUserDisplayName(comment.user) }}</text>
						<view v-if="isHot" class="ci-hot-badge">
							<l-icon name="local" size="11" color="#fff"></l-icon>
							<text class="ci-hot-text">热</text>
						</view>
					</view>
					<text class="ci-time">{{ formatTime(comment.createdAt) }}</text>
				</view>

				<view class="ci-content">
					<view v-if="level > 1 && comment.replyToUser" class="ci-reply-hint">
						<l-icon name="corner-down-left" size="12" color="var(--primary-500)"></l-icon>
						<text class="ci-reply-hint-text">@{{ getUserDisplayName(comment.replyToUser) }}</text>
					</view>
					<rich-text :nodes="processContent(comment.content)"></rich-text>
				</view>

				<view v-if="comment.images && comment.images.length > 0" class="ci-images">
					<view class="ci-images-grid" :class="imgGridClass">
						<image
							v-for="(img, idx) in comment.images"
							:key="idx"
							class="ci-image"
							:src="img"
							mode="aspectFill"
							lazy-load
							@tap.stop="previewImage(img, comment.images)"
						></image>
					</view>
				</view>

				<view class="ci-footer">
					<view class="ci-action ci-action--reply" @tap="handleReply">
						<l-icon name="chat" size="13" color="var(--text-tertiary)"></l-icon>
						<text class="ci-action-text">回复</text>
						<text v-if="comment.repliesCount > 0" class="ci-action-badge">{{ comment.repliesCount }}</text>
					</view>
					<view class="ci-action" :class="{ 'ci-action--liked': comment.isLiked }" @tap="handleLike">
						<l-icon
							:name="comment.isLiked ? 'heart-filled' : 'heart'"
							size="13"
							:color="comment.isLiked ? 'var(--error-color)' : 'var(--text-tertiary)'"
						></l-icon>
						<text class="ci-action-text" :class="{ 'ci-action-text--liked': comment.isLiked }">{{ formatNum(comment.likesCount) }}</text>
					</view>
				</view>

				<view v-if="comment.repliesCount > 0" class="ci-replies">
					<view v-if="!comment.showReplies" class="ci-replies-toggle" @tap="toggleReplies">
						<view class="ci-replies-toggle-line"></view>
						<text class="ci-replies-toggle-text">展开 {{ comment.repliesCount }} 条回复</text>
						<l-icon name="chevron-down" size="14" color="var(--primary-500)"></l-icon>
					</view>

					<view v-if="comment.showReplies && comment.replies" class="ci-replies-list">
						<comment-item
							v-for="reply in comment.replies"
							:key="reply.id"
							:comment="reply"
							:level="nextLevel"
							@reply="handleReplyToReply"
							@like="c => emit('like', c)"
							@load-replies="c => emit('loadReplies', c)"
							@show-more="c => emit('showMore', c)"
							@click-author="author => emit('clickAuthor', author)"
						/>
						<view class="ci-replies-collapse" @tap="toggleReplies">
							<l-icon name="chevron-up" size="14" color="var(--text-tertiary)"></l-icon>
							<text class="ci-replies-collapse-text">收起回复</text>
						</view>
						<view v-if="comment.repliesCount > comment.replies.length" class="ci-replies-more" @tap="showMoreReplies">
							<text class="ci-replies-more-text">查看更多 {{ comment.repliesCount - comment.replies.length }} 条回复</text>
						</view>
					</view>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup>
import UserAvatar from './UserAvatar.vue'
import { computed } from 'vue'
import { formatNumber } from '@/pages/api/community.js'
import { useTimeFormat } from '@/composables/useTimeFormat.js'

const props = defineProps({
	comment: { type: Object, required: true },
	level: { type: Number, default: 0 },
	isHot: { type: Boolean, default: false }
})

const emit = defineEmits(['reply', 'like', 'loadReplies', 'showMore', 'clickAuthor'])

const { formatTime } = useTimeFormat()

const rootClasses = computed(() => ({
	'ci-root--hot': props.isHot,
	'ci-root--reply': props.level > 0,
	'ci-root--deep': props.level > 1
}))

const rootStyle = computed(() => props.level > 0 ? { animationDelay: '0ms' } : {})

const nextLevel = computed(() => Math.min(props.level + 1, 2))

function getUserDisplayName(user) {
	if (!user) return '匿名用户'
	return user.realname || user.nickname || user.username || '匿名用户'
}

function processContent(content) {
	if (!content) return ''
	let processed = content
		.replace(/@([^\s@]+)/g, '<span class="ci-mention">@$1</span>')
		.replace(/\n/g, '<br/>')
		.replace(/(https?:\/\/[^\s]+)/g, '<span class="ci-link">$1</span>')
	return processed
}

const imgGridClass = computed(() => {
	const n = props.comment.images?.length || 0
	if (n === 1) return 'ci-grid-1'
	if (n === 2) return 'ci-grid-2'
	if (n <= 4) return 'ci-grid-4'
	return 'ci-grid-9'
})

function formatNum(num) { return formatNumber(num) }

function previewImage(current, urls) {
	uni.previewImage({ current, urls: urls || [current] })
}

function handleReply() { emit('reply', props.comment) }
function handleReplyToReply(reply) { emit('reply', reply, props.comment) }
function handleLike() { emit('like', props.comment) }

function toggleReplies() {
	if (!props.comment.repliesLoaded) emit('loadReplies', props.comment)
	else props.comment.showReplies = !props.comment.showReplies
}

function showMoreReplies() { emit('showMore', props.comment) }
</script>

<style lang="scss" scoped>
/* Rich text styles injected via rich-text nodes */
:deep(.ci-mention) {
	color: var(--primary-500);
	font-weight: 600;
	background: rgba(59, 102, 241, 0.08);
	padding: 2rpx 8rpx;
	border-radius: 8rpx;
}

:deep(.ci-link) {
	color: var(--primary-500);
	text-decoration: underline;
}

/* Root */
.ci-root {
	padding: 28rpx 24rpx;
	border-bottom: 1px solid rgba(226, 232, 240, 0.6);
	background: #fff;
	animation: ci-fadeIn 0.35s ease-out both;

	&--hot {
		background: linear-gradient(135deg, rgba(255, 107, 53, 0.04) 0%, rgba(255, 107, 53, 0.01) 100%);
		border-left: 4rpx solid #f59e0b;
		padding-left: 20rpx;
	}

	&--reply {
		background: #fafbfc;
		padding: 20rpx 20rpx 20rpx 16rpx;
		border-left: none;
		border-bottom-color: rgba(226, 232, 240, 0.4);
		border-radius: 16rpx;
		margin-top: 12rpx;
		margin-bottom: 0;
	}

	&--deep {
		background: rgba(59, 102, 241, 0.02);
		border-left: 2rpx dashed rgba(59, 102, 241, 0.2);
		border-radius: 12rpx;
	}
}

@keyframes ci-fadeIn {
	from { opacity: 0; transform: translateY(6rpx); }
	to { opacity: 1; transform: translateY(0); }
}

/* Main Layout */
.ci-main {
	display: flex;
	gap: 20rpx;
}

.ci-avatar {
	width: 72rpx;
	height: 72rpx;
	border-radius: 50%;
	flex-shrink: 0;
	background: var(--bg-muted);
	border: 2rpx solid rgba(148, 163, 184, 0.15);
	transition: transform 0.15s ease;

	&:active { transform: scale(0.95); }
}

.ci-root--reply .ci-avatar {
	width: 56rpx;
	height: 56rpx;
}

/* Body */
.ci-body {
	flex: 1;
	min-width: 0;
}

/* Header */
.ci-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 10rpx;
}

.ci-author-row {
	display: flex;
	align-items: center;
	gap: 10rpx;
}

.ci-author-name {
	font-size: 28rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.ci-hot-badge {
	display: flex;
	align-items: center;
	gap: 4rpx;
	padding: 3rpx 10rpx;
	border-radius: 100rpx;
	background: linear-gradient(135deg, #f59e0b, #f97316);
	flex-shrink: 0;
	animation: ci-pulse 2.5s ease-in-out infinite;

	.ci-hot-text {
		font-size: 18rpx;
		color: #fff;
		font-weight: 700;
	}
}

@keyframes ci-pulse {
	0%, 100% { opacity: 1; transform: scale(1); }
	50% { opacity: 0.85; transform: scale(0.97); }
}

.ci-time {
	font-size: 22rpx;
	color: var(--text-tertiary);
	white-space: nowrap;
	flex-shrink: 0;
}

/* Content */
.ci-content {
	font-size: 28rpx;
	line-height: 1.7;
	color: var(--text-primary);
	margin-bottom: 14rpx;
	word-break: break-all;
}

.ci-reply-hint {
	display: inline-flex;
	align-items: center;
	gap: 6rpx;
	margin-bottom: 10rpx;
	padding: 4rpx 14rpx;
	border-radius: 100rpx;
	background: rgba(59, 102, 241, 0.06);
	border: 1px solid rgba(59, 102, 241, 0.12);

	.ci-reply-hint-text {
		font-size: 22rpx;
		color: var(--primary-500);
		font-weight: 600;
	}
}

/* Images */
.ci-images {
	margin-bottom: 14rpx;
}

.ci-images-grid {
	display: grid;
	gap: 8rpx;
	border-radius: 16rpx;
	overflow: hidden;

	&.ci-grid-1 {
		grid-template-columns: 1fr;
		.ci-image { height: 320rpx; max-width: 480rpx; }
	}
	&.ci-grid-2 { grid-template-columns: repeat(2, 1fr); .ci-image { height: 200rpx; } }
	&.ci-grid-4 { grid-template-columns: repeat(2, 1fr); .ci-image { height: 180rpx; } }
	&.ci-grid-9 { grid-template-columns: repeat(3, 1fr); .ci-image { height: 150rpx; } }
}

.ci-image {
	width: 100%;
	object-fit: cover;
	transition: transform 0.15s ease;

	&:active { transform: scale(0.97); }
}

/* Footer Actions */
.ci-footer {
	display: flex;
	align-items: center;
	gap: 8rpx;
}

.ci-action {
	display: flex;
	align-items: center;
	gap: 6rpx;
	padding: 8rpx 16rpx;
	border-radius: 100rpx;
	transition: all 0.15s ease;

	&:active { background: var(--bg-muted); transform: scale(0.96); }

	&--liked {
		background: rgba(239, 68, 68, 0.06);

		&:active { background: rgba(239, 68, 68, 0.1); }
	}

	&--reply { margin-right: 4rpx; }
}

.ci-action-text {
	font-size: 22rpx;
	color: var(--text-tertiary);
	font-weight: 500;

	&--liked { color: var(--error-color); font-weight: 700; }
}

.ci-action-badge {
	font-size: 20rpx;
	color: var(--primary-500);
	font-weight: 700;
	background: rgba(59, 102, 241, 0.08);
	padding: 2rpx 8rpx;
	border-radius: 100rpx;
}

/* Replies */
.ci-replies {
	margin-top: 16rpx;
	padding-top: 16rpx;
}

.ci-replies-toggle {
	display: flex;
	align-items: center;
	gap: 10rpx;
	transition: all 0.15s ease;

	&:active { opacity: 0.7; }
}

.ci-replies-toggle-line {
	flex: 1;
	height: 1rpx;
	background: rgba(226, 232, 240, 0.8);
	max-width: 60rpx;
}

.ci-replies-toggle-text {
	font-size: 24rpx;
	color: var(--primary-500);
	font-weight: 600;
}

.ci-replies-list {
	margin-top: 12rpx;
	padding-left: 8rpx;
	border-left: 2rpx solid rgba(59, 102, 241, 0.15);
}

.ci-replies-collapse {
	display: flex;
	align-items: center;
	gap: 8rpx;
	padding: 12rpx 0 8rpx;

	&:active { opacity: 0.7; }

	.ci-replies-collapse-text {
		font-size: 24rpx;
		color: var(--text-tertiary);
		font-weight: 500;
	}
}

.ci-replies-more {
	padding: 12rpx 0;

	&:active { opacity: 0.7; }

	.ci-replies-more-text {
		font-size: 24rpx;
		color: var(--primary-500);
		font-weight: 600;
	}
}
</style>
