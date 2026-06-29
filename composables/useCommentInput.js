/**
 * Comment Input Composable
 * 评论输入框的逻辑和状态管理
 */

import { computed, ref } from 'vue'
import { getUserDisplayName, searchMentionUsers } from '@/utils/userMention.js'
import { addToRecent } from '@/utils/emoji.js'

export function useCommentInput(options = {}) {
	const postId = options.postId
	const maxImages = options.maxImages ?? 9
	const emojiPickerRef = options.emojiPickerRef || ref(null)

	const commentText = ref('')
	const images = ref([])
	const submitting = ref(false)
	const focused = ref(false)
	const replyTarget = ref(null)
	const parentComment = ref(null)
	const mentionSuggestions = ref([])
	const showMentionSuggestions = ref(false)
	const mentionedUsers = ref([])
	const cursorPosition = ref(0)
	const visible = ref(false)

	const canSubmit = computed(() =>
		(commentText.value.trim().length > 0 || images.value.length > 0) && !submitting.value
	)

	function show(replyTargetParam = null, parentCommentParam = null) {
		replyTarget.value = replyTargetParam
		parentComment.value = parentCommentParam

		if (replyTargetParam) {
			const userName = getUserDisplayName(replyTargetParam.user)
			commentText.value = `@${userName} `
			addMentionedUser(replyTargetParam.user)
		}

		visible.value = true
		focused.value = true
	}

	function close() {
		visible.value = false
	}

	function handlePopupChange(e) {
		if (!e.visible) {
			reset()
		}
	}

	function reset() {
		commentText.value = ''
		images.value = []
		replyTarget.value = null
		parentComment.value = null
		mentionedUsers.value = []
		showMentionSuggestions.value = false
		focused.value = false
		cursorPosition.value = 0
	}

	function getReplyTargetName() {
		if (!replyTarget.value) return ''
		return getUserDisplayName(replyTarget.value.user)
	}

	function clearReplyTarget() {
		replyTarget.value = null
		parentComment.value = null
		commentText.value = commentText.value.replace(/@\w+\s/, '')
		mentionedUsers.value = []
	}

	function detectMention(text) {
		const mentionMatch = text.match(/@(\w*)$/)
		if (mentionMatch) {
			const query = mentionMatch[1]
			if (query.length >= 1) {
				searchUsers(query)
			} else {
				showMentionSuggestions.value = false
			}
		} else {
			showMentionSuggestions.value = false
		}
	}

	async function searchUsers(query) {
		if (!query || query.length < 1) {
			showMentionSuggestions.value = false
			return
		}

		try {
			mentionSuggestions.value = await searchMentionUsers(query)
			showMentionSuggestions.value = mentionSuggestions.value.length > 0
		} catch (error) {
			console.error('搜索用户失败:', error)
			showMentionSuggestions.value = false
		}
	}

	function selectMentionUser(user) {
		const userName = getUserDisplayName(user)
		commentText.value = commentText.value.replace(/@\w*$/, `@${userName} `)
		addMentionedUser(user)
		showMentionSuggestions.value = false
	}

	function addMentionedUser(user) {
		if (!mentionedUsers.value.find(u => u.id === user.id)) {
			mentionedUsers.value.push(user)
		}
	}

	async function chooseImage() {
		try {
			const res = await new Promise((resolve, reject) => {
				uni.chooseImage({
					count: maxImages - images.value.length,
					sizeType: ['compressed'],
					sourceType: ['album', 'camera'],
					success: resolve,
					fail: reject
				})
			})

			if (res.tempFilePaths && res.tempFilePaths.length > 0) {
				const validPaths = []
				for (const path of res.tempFilePaths) {
					try {
						const { validateImage } = await import('@/utils/imageUpload.js')
						await validateImage(path)
						validPaths.push(path)
					} catch (error) {
						uni.showToast({ title: error.message, icon: 'none', duration: 2000 })
					}
				}

				if (validPaths.length > 0) {
					await handleImageUpload(validPaths)
				}
			}
		} catch (err) {
			console.error('选择图片失败:', err)
			uni.showToast({ title: '选择图片失败', icon: 'none' })
		}
	}

	async function handleImageUpload(filePaths) {
		try {
			const { uploadImages } = await import('@/utils/imageUpload.js')
			const result = await uploadImages(filePaths, {
				showLoading: true,
				compress: true,
				quality: 0.8,
				maxRetries: 3
			})

			if (result.success && result.urls.length > 0) {
				images.value.push(...result.urls)

				if (result.summary) {
					const { local, failed } = result.summary
					if (local > 0) {
						console.log(`${local}张图片已添加（开发模式）`)
					} else if (failed > 0) {
						console.warn(`${result.urls.length}张图片成功，${failed}张失败`)
					}
				}
			} else if (result.errors && result.errors.length > 0) {
				const connectionErrors = result.errors.filter(e => e.errorType === 'connection')
				if (connectionErrors.length > 0) {
					uni.showToast({ title: '网络连接失败，请检查服务器状态', icon: 'none', duration: 3000 })
				}
			}
		} catch (error) {
			console.error('图片上传处理失败:', error)
			uni.showToast({ title: '上传失败', icon: 'none', duration: 3000 })
		}
	}

	function removeImage(index) {
		images.value.splice(index, 1)
	}

	function previewImage(current) {
		uni.previewImage({ current, urls: images.value })
	}

	function showEmojiPanel() {
		uni.vibrateShort({ type: 'light' })
		emojiPickerRef.value?.show?.()
	}

	function handleEmojiSelect(emoji) {
		insertEmojiAtCursor(emoji.char)
		addToRecent(emoji.char)
		emojiPickerRef.value?.close?.()
	}

	function handleEmojiClose() {
		focused.value = true
	}

	function insertEmojiAtCursor(emojiChar) {
		const text = commentText.value
		const position = cursorPosition.value
		const before = text.substring(0, position)
		const after = text.substring(position)
		commentText.value = before + emojiChar + after
		cursorPosition.value = position + emojiChar.length
		focused.value = true
	}

	function updateCursorPosition(e) {
		if (e && e.detail && e.detail.cursor !== undefined) {
			cursorPosition.value = e.detail.cursor
			detectMention(commentText.value)
		}
	}

	async function submit(emitFn) {
		const trimmedContent = commentText.value.trim()
		if (!trimmedContent && images.value.length === 0) {
			uni.showToast({ title: '请输入评论内容', icon: 'none' })
			return
		}

		submitting.value = true

		try {
			const { createComment } = await import('@/pages/api/community.js')
			const result = await createComment({
				postId: postId,
				content: commentText.value,
				images: images.value,
				parentId: parentComment.value ? parentComment.value.id : null,
				replyId: replyTarget.value ? replyTarget.value.id : null
			})

			if (result.success) {
				uni.showToast({ title: '评论成功', icon: 'success' })
				close()
				emitFn?.()
			} else {
				uni.showToast({ title: result.message || '评论失败', icon: 'none' })
			}
		} catch (error) {
			console.error('发表评论失败:', error)
			uni.showToast({ title: '评论失败', icon: 'none' })
		} finally {
			submitting.value = false
		}
	}

	return {
		visible,
		commentText,
		images,
		submitting,
		focused,
		replyTarget,
		parentComment,
		mentionSuggestions,
		showMentionSuggestions,
		mentionedUsers,
		cursorPosition,
		canSubmit,
		show,
		close,
		handlePopupChange,
		reset,
		getReplyTargetName,
		clearReplyTarget,
		detectMention,
		searchUsers,
		selectMentionUser,
		addMentionedUser,
		chooseImage,
		handleImageUpload,
		removeImage,
		previewImage,
		showEmojiPanel,
		handleEmojiSelect,
		handleEmojiClose,
		insertEmojiAtCursor,
		updateCursorPosition,
		submit
	}
}