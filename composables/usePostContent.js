/**
 * Post Content Composable
 * Post content state and interaction methods
 */

import { computed, ref } from 'vue'
import { POST_FILTERS } from '@/utils/postContent.js'

export function usePostContent() {
	const currentFilter = ref(0)
	const filters = ref(POST_FILTERS)

	function switchFilter(index) {
		currentFilter.value = index
	}

	function previewImage(images, index) {
		uni.previewImage({
			urls: images,
			current: index
		})
	}

	function sharePost(post) {
		console.log('分享帖子:', post.title)

		// #ifdef H5
		uni.showModal({
			title: '分享',
			content: '分享功能开发中',
			showCancel: false
		})
		// #endif
	}

	function toggleFavorite(post) {
		post.isFavorited = !post.isFavorited
		uni.showToast({
			title: post.isFavorited ? '已收藏' : '已取消收藏',
			icon: 'none'
		})
	}

	function showManageMenu(post, handlers) {
		const isMine = post.isOwner
		const actions = ['举报', '屏蔽作者']

		if (isMine) {
			actions.unshift('编辑', '删除')
		}

		uni.showActionSheet({
			itemList: actions,
			success: (res) => {
				const action = actions[res.tapIndex]
				handleMenuAction(action, post, handlers)
			}
		})
	}

	function handleMenuAction(action, post, handlers) {
		switch (action) {
			case '编辑':
				handlers?.onEdit?.(post)
				break
			case '删除':
				confirmDelete(post, handlers)
				break
			case '举报':
				handlers?.onReport?.(post)
				break
			case '屏蔽作者':
				confirmBlock(post, handlers)
				break
		}
	}

	function confirmDelete(post, handlers) {
		uni.showModal({
			title: '确认删除',
			content: '删除后无法恢复，确定要删除这篇文章吗？',
			success: (modalRes) => {
				if (modalRes.confirm) {
					handlers?.onDelete?.(post)
				}
			}
		})
	}

	function confirmBlock(post, handlers) {
		uni.showModal({
			title: '确认屏蔽',
			content: `确定要屏蔽用户 ${post.author?.nickname || '该用户'} 吗？`,
			success: (modalRes) => {
				if (modalRes.confirm) {
					handlers?.onBlock?.(post.author)
				}
			}
		})
	}

	return {
		currentFilter,
		filters,
		switchFilter,
		previewImage,
		sharePost,
		toggleFavorite,
		showManageMenu,
		handleMenuAction,
		confirmDelete,
		confirmBlock
	}
}