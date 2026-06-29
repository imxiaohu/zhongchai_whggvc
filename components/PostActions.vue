<template>
	<view class="post-actions">
		<view class="main-actions">
			<button
				class="action-btn like-btn"
				:class="{ 'liked': isLiked, 'loading': likeLoading }"
				@click="toggleLike"
				:disabled="likeLoading"
			>
				<l-icon :name="isLiked ? 'heart-filled' : 'heart'" size="16" :color="isLiked ? 'var(--error-color)' : 'var(--text-secondary)'"></l-icon>
				<text class="count">{{ formatNum(likesCount) }}</text>
			</button>

			<button class="action-btn comment-btn" @click="goToComments">
				<l-icon name="chat" size="16" color="var(--text-secondary)"></l-icon>
				<text class="count">{{ formatNum(commentsCount) }}</text>
			</button>

			<BookmarkButton
				:post-id="postId"
				:initial-bookmarked="initialBookmarked"
				:author-id="authorId"
				:post-title="postTitle"
				@bookmark-changed="onBookmarkChanged"
			/>

			<button class="action-btn more-btn" @click="showMoreActions">
				<l-icon name="more" size="16" color="var(--text-secondary)"></l-icon>
			</button>
		</view>

		<t-popup v-model:visible="moreActionsVisible" placement="bottom" :overlay="true" :close-on-overlay-click="true">
			<view class="more-actions-modal">
				<view class="modal-header">
					<text class="modal-title">更多操作</text>
					<button class="close-btn" @click="hideMoreActions">
						<text class="close-icon">×</text>
					</button>
				</view>

				<view class="actions-list">
					<button class="action-item" @click="sharePost">
						<l-icon name="share" size="18" color="var(--text-primary)" class="action-icon"></l-icon>
						<text class="action-text">分享</text>
					</button>

					<button class="action-item" @click="copyLink">
						<l-icon name="link" size="18" color="var(--text-primary)" class="action-icon"></l-icon>
						<text class="action-text">复制链接</text>
					</button>

					<button class="action-item report-item" @click="showReportModal">
						<l-icon name="help-circle-filled" size="18" color="#ff4757" class="action-icon"></l-icon>
						<text class="action-text">举报</text>
					</button>

					<button
						v-if="!isOwnPost"
						class="action-item block-item"
						@click="blockAuthor"
					>
						<l-icon name="view-off" size="18" color="#ff4757" class="action-icon"></l-icon>
						<text class="action-text">屏蔽作者</text>
					</button>

					<button
						v-if="isOwnPost"
						class="action-item edit-item"
						@click="editPost"
					>
						<l-icon name="edit-1" size="18" color="var(--text-primary)" class="action-icon"></l-icon>
						<text class="action-text">编辑</text>
					</button>

					<button
						v-if="isOwnPost"
						class="action-item delete-item"
						@click="deletePost"
					>
						<l-icon name="delete-1" size="18" color="#ff4757" class="action-icon"></l-icon>
						<text class="action-text">删除</text>
					</button>
				</view>
			</view>
		</t-popup>

		<ReportModal
			:visible="reportModalVisible"
			target-type="post"
			:target-id="postId"
			@close="hideReportModal"
			@success="onReportSubmitted"
		/>
	</view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { likePost, unlikePost, deletePost as deletePostAPI, blockUser, formatNumber } from '@/pages/api/community.js'
import BookmarkButton from './BookmarkButton.vue'
import ReportModal from './ReportModal.vue'

const props = defineProps({
	postId: {
		type: [Number, String],
		required: true
	},
	initialLiked: {
		type: Boolean,
		default: false
	},
	initialLikesCount: {
		type: Number,
		default: 0
	},
	commentsCount: {
		type: Number,
		default: 0
	},
	initialBookmarked: {
		type: Boolean,
		default: false
	},
	authorId: {
		type: [Number, String],
		default: null
	},
	postUrl: {
		type: String,
		default: ''
	},
	postTitle: {
		type: String,
		default: ''
	}
})

const emit = defineEmits([
	'like-changed', 'show-comments', 'bookmark-changed',
	'share', 'report-submitted', 'author-blocked',
	'edit-post', 'post-deleted'
])

const moreActionsVisible = ref(false)
const isLiked = ref(false)
const likesCount = ref(0)
const likeLoading = ref(false)
const reportModalVisible = ref(false)

const isOwnPost = computed(() => {
	const currentUserId = uni.getStorageSync('userId')
	return currentUserId && props.authorId && currentUserId.toString() === props.authorId.toString()
})

onMounted(() => {
	isLiked.value = props.initialLiked
	likesCount.value = props.initialLikesCount
})

async function toggleLike() {
	if (likeLoading.value) return

	const token = uni.getStorageSync('token') || uni.getStorageSync('accessToken')
	if (!token) {
		uni.showToast({ title: '请先登录', icon: 'none' })
		return
	}

	likeLoading.value = true

	try {
		let response
		if (isLiked.value) {
			response = await unlikePost(props.postId)
		} else {
			response = await likePost(props.postId)
		}

		if (response.success) {
			isLiked.value = !isLiked.value
			likesCount.value += isLiked.value ? 1 : -1
			emit('like-changed', {
				postId: props.postId,
				isLiked: isLiked.value,
				likesCount: likesCount.value
			})
		} else {
			throw new Error(response.message || '操作失败')
		}
	} catch (error) {
		console.error('点赞操作失败:', error)
		uni.showToast({ title: error.message || '操作失败', icon: 'none' })
	} finally {
		likeLoading.value = false
	}
}

function goToComments() {
	emit('show-comments')
}

function onBookmarkChanged(data) {
	emit('bookmark-changed', data)
}

function showMoreActions() {
	moreActionsVisible.value = true
}

function hideMoreActions() {
	moreActionsVisible.value = false
}

function sharePost() {
	hideMoreActions()

	// #ifdef MP-WEIXIN
	uni.showShareMenu({
		withShareTicket: true,
		success: () => {
			console.log('分享菜单显示成功')
		}
	})
	// #endif

	// #ifdef H5
	if (typeof navigator !== 'undefined' && navigator.share) {
		navigator.share({
			title: '分享帖子',
			url: props.postUrl || window.location.href
		})
	} else {
		copyLink()
	}
	// #endif

	emit('share', { postId: props.postId })
}

function copyLink() {
	hideMoreActions()

	const url = props.postUrl || `${window.location.origin}/pages/community/post-detail?id=${props.postId}`

	// #ifdef MP-WEIXIN
	uni.setClipboardData({
		data: url,
		success: () => {
			uni.showToast({ title: '链接已复制', icon: 'success' })
		}
	})
	// #endif

	// #ifdef H5
	if (typeof navigator !== 'undefined' && navigator.clipboard) {
		navigator.clipboard.writeText(url).then(() => {
			uni.showToast({ title: '链接已复制', icon: 'success' })
		})
	}
	// #endif
}

function showReportModal() {
	hideMoreActions()
	reportModalVisible.value = true
}

function hideReportModal() {
	reportModalVisible.value = false
}

function onReportSubmitted() {
	uni.showToast({ title: '举报提交成功', icon: 'success' })
	emit('report-submitted')
}

async function blockAuthor() {
	hideMoreActions()

	if (!props.authorId) {
		uni.showToast({ title: '操作失败', icon: 'none' })
		return
	}

	uni.showModal({
		title: '屏蔽作者',
		content: '确定要屏蔽该作者吗？',
		success: async (res) => {
			if (res.confirm) {
				try {
					const response = await blockUser(props.authorId)
					if (response.success) {
						uni.showToast({ title: '屏蔽成功', icon: 'success' })
						emit('author-blocked', { authorId: props.authorId })
					}
				} catch (error) {
					uni.showToast({ title: error.message || '操作失败', icon: 'none' })
				}
			}
		}
	})
}

function editPost() {
	hideMoreActions()
	emit('edit-post', { postId: props.postId })
}

function deletePost() {
	hideMoreActions()

	uni.showModal({
		title: '删除动态',
		content: '确定要删除这条动态吗？',
		success: async (res) => {
			if (res.confirm) {
				try {
					const response = await deletePostAPI(props.postId)
					if (response.success) {
						uni.showToast({ title: '删除成功', icon: 'success' })
						emit('post-deleted', { postId: props.postId })
					}
				} catch (error) {
					uni.showToast({ title: error.message || '操作失败', icon: 'none' })
				}
			}
		}
	})
}

function formatNum(num) {
	return formatNumber(num)
}
</script>

<style lang="scss" scoped>
.post-actions {
	padding: 24rpx 0;
}

.main-actions {
	display: flex;
	align-items: center;
	gap: 32rpx;
}

.action-btn {
	display: flex;
	align-items: center;
	gap: 8rpx;
	padding: 12rpx 20rpx;
	border: none;
	border-radius: 20rpx;
	background-color: transparent;
	transition: all 0.3s ease;
	min-height: 60rpx;

	&:not(:disabled):active {
		transform: scale(0.95);
	}

	&.like-btn {
		&.liked {
			background-color: rgba(255, 107, 53, 0.1);
		}

		&.loading {
			opacity: 0.7;
		}
	}

	&.comment-btn,
	&.more-btn {
		&:hover {
			background-color: rgba(0, 0, 0, 0.05);
		}
	}
}

.icon {
	font-size: 32rpx;
}

.count {
	font-size: 28rpx;
	color: #666;
	font-weight: 500;
}

.more-actions-modal {
	background-color: white;
	border-radius: 32rpx 32rpx 0 0;
	padding: 0;
	margin: 0;
	max-height: 80vh;
	overflow: hidden;
}

.modal-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 32rpx;
	border-bottom: 2rpx solid #f0f0f0;
}

.modal-title {
	font-size: 36rpx;
	font-weight: 600;
	color: #333;
}

.close-btn {
	width: 60rpx;
	height: 60rpx;
	border-radius: 50%;
	background-color: #f5f5f5;
	border: none;
	display: flex;
	align-items: center;
	justify-content: center;
}

.close-icon {
	font-size: 40rpx;
	color: #666;
	line-height: 1;
}

.actions-list {
	padding: 16rpx 0 32rpx;
}

.action-item {
	display: flex;
	align-items: center;
	width: 100%;
	padding: 24rpx 32rpx;
	border: none;
	background-color: transparent;
	text-align: left;
	transition: background-color 0.3s ease;

	&:hover {
		background-color: #f8f9fa;
	}

	&.report-item,
	&.block-item,
	&.delete-item {
		.action-text {
			color: #ff4757;
		}
	}
}

.action-icon {
	font-size: 40rpx;
	margin-right: 24rpx;
	width: 40rpx;
	text-align: center;
}

.action-text {
	font-size: 32rpx;
	color: #333;
	font-weight: 500;
}

/* 主题适配 */
.theme-dark {
	.action-btn {
		&.comment-btn,
		&.more-btn {
			&:hover {
				background-color: rgba(255, 255, 255, 0.1);
			}
		}
	}

	.count {
		color: #999;
	}

	.more-actions-modal {
		background-color: #2a2a2a;
	}

	.modal-header {
		border-bottom-color: #444;
	}

	.modal-title {
		color: #e0e0e0;
	}

	.close-btn {
		background-color: #444;
	}

	.close-icon {
		color: #ccc;
	}

	.action-item {
		&:hover {
			background-color: #333;
		}
	}

	.action-text {
		color: #e0e0e0;

		&.report-item,
		&.block-item,
		&.delete-item {
			color: #ff6b6b;
		}
	}
}

/* 响应式设计 */
@media screen and (max-width: 480px) {
	.main-actions {
		gap: 24rpx;
	}

	.action-btn {
		padding: 10rpx 16rpx;
		min-height: 56rpx;
	}

	.icon {
		font-size: 28rpx;
	}

	.count {
		font-size: 26rpx;
	}

	.modal-header,
	.action-item {
		padding-left: 24rpx;
		padding-right: 24rpx;
	}

	.action-icon {
		font-size: 36rpx;
		margin-right: 20rpx;
	}

	.action-text {
		font-size: 30rpx;
	}
}
</style>
