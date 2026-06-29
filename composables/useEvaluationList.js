/**
 * Evaluation List Composable
 * 评教列表相关的状态和方法
 */

import { ref } from 'vue'
import {
	generatePreviewEvaluationList,
	formatTimeRange,
	getPendingEvaluationItems
} from '@/utils/evaluationList.js'
import { getEvaluationList, submitEvaluation, getEvaluationDetail } from '@/pages/api/evaluation.js'
import { showToast, navigateTo, navigateBack } from '@/pages/api/page.js'
import { shouldAutoRefreshData } from '@/utils/errorHandler.js'
import unreadMessageManager from '@/utils/unreadMessageManager.js'
import shareManager from '@/utils/shareManager.js'
import {
	loadCacheData,
	getCacheStatus,
	updateListCache,
	clearEvaluationCache,
	fetchEvaluationDetail,
	processEvaluationItemData,
	formatEvaluationData,
	createDefaultEvaluationData
} from '@/utils/evaluationList.js'

export function useEvaluationList() {
	const loading = ref(true)
	const error = ref('')
	const evaluationList = ref([])
	const evaluationData = ref({})
	const evaluationCache = ref({})
	const showClearCacheConfirm = ref(false)
	const lastRefreshTime = ref(0)
	const isPreviewMode = ref(false)
	const navPaddingTop = ref('0px')

	 async function refreshUnreadMessages() {
		try {
			await unreadMessageManager.fetchUnreadCount()
		} catch (err) {
			console.error('评教页面: 刷新未读消息数量失败:', err)
		}
	}

	function showPreviewData() {
		loading.value = false
		error.value = ''
		isPreviewMode.value = true
		evaluationList.value = generatePreviewEvaluationList()
	}

	function goToLogin() {
		uni.vibrateShort()
		uni.navigateTo({ url: '/pages/login/login' })
	}

	function handleNavHeightReady(navInfo) {
		navPaddingTop.value = navInfo.heightPx
	}

	function handleBackClick() {
		navigateBack()
	}

	function handleTitleClick() {
		refreshEvaluationData()
	}

	async function refreshEvaluationData() {
		if (isPreviewMode.value) {
			uni.stopPullDownRefresh()
			showToast({ title: '预览模式无需刷新', icon: 'none' })
			return
		}

		try {
			await loadEvaluationList()
			await loadAllEvaluationData()
			lastRefreshTime.value = Date.now()
		} catch (err) {
			console.error('刷新数据失败:', err)
			showToast({ title: '刷新失败，请重试', icon: 'none' })
		} finally {
			uni.stopPullDownRefresh()
		}
	}

	async function loadEvaluationList() {
		loading.value = true
		error.value = ''

		try {
			const res = await getEvaluationList()
			const listData = Array.isArray(res) ? res : Array.isArray(res?.result) ? res.result : null
			if (!listData) {
				evaluationList.value = []
				return
			}

			evaluationList.value = listData.map(item => ({
				id: item.id || item.courseTeachingEvaluationStudentConfigID || item.courseTeachingEvaluationStudentConfigId,
				courseTeachingClassTaskID: item.courseTeachingClassTaskID,
				teacherID: item.teacherID,
				name: item.courseName || item.name,
				teacherName: item.teacherName,
				startTime: item.startTime,
				endTime: item.endTime,
				evaluated: item.evaluated || false,
				courseTeachingEvaluationQuestionsVOS: item.courseTeachingEvaluationQuestionsVOS
			}))
		} catch (err) {
			if (err.statusCode === 401 || err.isTokenInvalid ||
				(err.message && err.message.includes('会话已失效'))) {
				handleSessionExpired()
				return
			}

			error.value = err.message || '获取评教列表失败，请重试'
		} finally {
			loading.value = false
		}
	}

	function handleSessionExpired() {
		const userInfoStr = uni.getStorageSync('userInfo')
		const hasWechatLogin = !!userInfoStr

		if (hasWechatLogin) {
			error.value = '学校账号会话失效，请重新绑定'
			uni.showModal({
				title: '会话失效',
				content: '学校账号会话已失效，是否前往重新绑定？',
				confirmText: '去绑定',
				cancelText: '稍后再说',
				success: (res) => {
					if (res.confirm) {
						uni.navigateTo({ url: '/pages/user/bind' })
					}
				}
			})
		} else {
			error.value = '登录已过期，请重新登录'
			uni.showModal({
				title: '登录过期',
				content: '登录已过期，是否前往重新登录？',
				confirmText: '去登录',
				cancelText: '稍后再说',
				success: (res) => {
					if (res.confirm) {
						uni.reLaunch({ url: '/pages/login/login' })
					}
				}
			})
		}
	}

	 async function loadAllEvaluationData() {
		const promises = evaluationList.value.map(item => loadSingleEvaluationData(item))
		await Promise.all(promises)
	}

	 async function loadSingleEvaluationData(item) {
		try {
			if (!item?.id || item.id === 'undefined' || item.id === 'null') {
				console.warn('[useEvaluationList] skip loadSingleEvaluationData: invalid item.id', item?.id)
				return
			}

			if (isPreviewMode.value || (item && item.isPreview)) {
				evaluationData.value[item.id] = createDefaultEvaluationData()
				return
			}

			if (item && Array.isArray(item.courseTeachingEvaluationQuestionsVOS)) {
				evaluationData.value[item.id] = {
					normList: item.courseTeachingEvaluationQuestionsVOS.map(norm => ({ ...norm, score: 0 })),
					comment: '',
					isCompleted: false
				}
				return
			}

			const cachedData = evaluationCache.value[item.id]
			if (cachedData) {
				evaluationData.value[item.id] = cachedData
				return
			}

			const detail = await getEvaluationDetail(item.id)
			evaluationData.value[item.id] = processEvaluationDetail(detail)
		} catch (err) {
			console.error(`加载评教详情失败 ID:${item.id}`, err)
			evaluationData.value[item.id] = createDefaultEvaluationData()
		}
	}

	function loadCacheDataMixin() {
		const cache = loadCacheData()
		evaluationCache.value = cache
	}

	function getCacheStatusMixin(id) {
		return getCacheStatus(evaluationCache.value, id)
	}

	function submitAllEvaluations() {
		const completed = Object.values(evaluationCache.value).filter(item => item.isCompleted)

		if (completed.length === 0) {
			uni.showToast({ title: '没有可提交的评教', icon: 'none' })
			return
		}

		uni.showModal({
			title: '批量提交',
			content: `确定要提交${completed.length}个已完成的评教吗？`,
			success: (res) => {
				if (res.confirm) {
					processBatchSubmit(completed)
				}
			}
		})
	}

	 async function processBatchSubmit(evaluations) {
		uni.showLoading({ title: '提交中...', mask: true })

		let successCount = 0
		let failCount = 0

		for (const evaluation of evaluations) {
			try {
				const cacheKey = `evaluation_cache_${evaluation.id}`
				const cacheStr = uni.getStorageSync(cacheKey)
				if (!cacheStr) continue

				const cacheData = JSON.parse(cacheStr)
				const submitData = {
					id: evaluation.id,
					normList: cacheData.normList.map(norm => ({
						normId: norm.id,
						score: norm.score
					})),
					comment: cacheData.comment || ''
				}

				await submitEvaluation(submitData)
				uni.removeStorageSync(cacheKey)
				successCount++
			} catch (err) {
				failCount++
			}
		}

		updateListCacheMixin()
		await loadEvaluationList()

		uni.hideLoading()

		if (failCount === 0) {
			uni.showToast({ title: `成功提交${successCount}个评教`, icon: 'success' })
		} else {
			uni.showModal({
				title: '提交结果',
				content: `成功: ${successCount}个, 失败: ${failCount}个`,
				showCancel: false
			})
		}
	}

	function updateListCacheMixin() {
		loadCacheDataMixin()
		updateListCache(evaluationCache.value)
	}

	function clearAllCache() {
		showClearCacheConfirm.value = true
	}

	function cancelClearCache() {
		showClearCacheConfirm.value = false
	}

	function confirmClearCache() {
		try {
			const items = Object.keys(evaluationCache.value)
			items.forEach(id => {
				const cacheKey = `evaluation_cache_${id}`
				uni.removeStorageSync(cacheKey)
			})
			uni.removeStorageSync('evaluation_list_cache')
			evaluationCache.value = {}
			showClearCacheConfirm.value = false
			uni.showToast({ title: '所有草稿已清除', icon: 'success' })
		} catch (e) {
			uni.showToast({ title: '清除失败', icon: 'none' })
		}
	}

	function goToSwipeEvaluation() {
		if (isPreviewMode.value) {
			uni.showModal({
				title: '预览模式',
				content: '这是预览数据，登录后可以使用滑动评教功能',
				confirmText: '立即登录',
				cancelText: '取消',
				success: (res) => {
					if (res.confirm) uni.navigateTo({ url: '/pages/login/login' })
				}
			})
			return
		}
		navigateTo({ url: '/pages/evaluation/swipe' })
	}

	 async function quickEvaluateAll() {
		if (isPreviewMode.value) {
			uni.showModal({
				title: '预览模式',
				content: '这是预览数据，登录后可以进行真实的评教操作',
				confirmText: '立即登录',
				cancelText: '取消',
				success: (res) => {
					if (res.confirm) uni.navigateTo({ url: '/pages/login/login' })
				}
			})
			return
		}

		const pending = getPendingEvaluationItems(evaluationList.value)
		if (!validatePendingItems(pending)) {
			showToast({ title: '没有待评教的课程', icon: 'none' })
			return
		}

		try {
			uni.showLoading({ title: '数据准备中...', mask: true })
			const evaluationData = await prepareEvaluationDataMixin(pending)
			uni.hideLoading()

		if (!evaluationData || evaluationData.length === 0) {
			showToast({ title: '获取评教数据失败', icon: 'none' })
			return
		}

		mergeUserScoresMixin(evaluationData)
		showSubmitConfirmationMixin(evaluationData)
	} catch (err) {
		uni.hideLoading()
		showToast({ title: '准备数据失败，请重试', icon: 'none' })
	}
	}

	function mergeUserScoresMixin(evaluationDataItems) {
		evaluationDataItems.forEach(item => {
			const userScores = evaluationData.value[item.id]?.normList
			if (userScores && item.detail && Array.isArray(item.detail)) {
				item.detail.forEach((detailItem, index) => {
					if (userScores[index]?.score > 0) {
						detailItem.score = userScores[index].score
					}
				})
			}
		})
	}

	function showSubmitConfirmationMixin(evaluationDataItems) {
		uni.showModal({
			title: '提交评教',
			content: `已准备${evaluationDataItems.length}个课程的评教数据，确定要提交吗？（将保留您已评过的分数）`,
			success: (res) => {
				if (res.confirm) {
					processQuickEvaluateAndSubmitMixin(evaluationDataItems)
				}
			}
		})
	}

	 async function prepareEvaluationDataMixin(pendingItems) {
		const data = []

		for (const item of pendingItems) {
			try {
				const detail = await fetchEvaluationDetail(
					item.id,
					isPreviewMode.value,
					evaluationList.value,
					evaluationData.value
				)
				const processed = processEvaluationItemData(item, detail, evaluationData.value)
				if (processed) data.push(processed)
			} catch (err) {
				console.error(`获取评教详情失败(${item.id}):`, err)
			}
		}

		return data
	}

	 async function processQuickEvaluateAndSubmitMixin(evaluationDataItems) {
		if (!evaluationDataItems || evaluationDataItems.length === 0) {
			showToast({ title: '没有待评教的数据', icon: 'none' })
			return
		}

		try {
			loading.value = true
			uni.showLoading({ title: '正在提交所有评教数据...', mask: true })
			const formattedData = formatEvaluationData(evaluationDataItems)

			await submitEvaluation(formattedData)

			evaluationDataItems.forEach(item => {
				clearEvaluationCache(item.id)
				updateEvaluationStatusMixin(item.id)
			})

			showToast({ title: `成功评教${evaluationDataItems.length}个课程`, icon: 'success' })
		} catch (err) {
			uni.showModal({
				title: '提交失败',
				content: err.message || '批量提交评教失败，请稍后再试',
				showCancel: false
			})
		} finally {
			loading.value = false
			uni.hideLoading()
		}
	}

	function updateEvaluationStatusMixin(id) {
		const idx = evaluationList.value.findIndex(item => item.id == id)
		if (idx >= 0) {
			evaluationList.value[idx].evaluated = true
		}
	}

	function goToEvaluate(item) {
		if (!item?.id || item.id === 'undefined' || item.id === 'null') {
			console.warn('[goToEvaluate] skip: invalid item.id', item?.id);
			return;
		}
		if (isPreviewMode.value || (item && item.isPreview)) {
			uni.showModal({
				title: '预览模式',
				content: '这是预览数据，登录后可以进行真实的评教操作',
				confirmText: '立即登录',
				cancelText: '取消',
				success: (res) => {
					if (res.confirm) uni.navigateTo({ url: '/pages/login/login' })
				}
			})
			return
		}
		navigateTo({
			url: `/pages/evaluation/swipe?id=${item.id}&name=${item.name}&teacherName=${item.teacherName}&startTime=${item.startTime}&endTime=${item.endTime}`
		})
	}

	function formatTimeRangeMixin(startTime, endTime) {
		return formatTimeRange(startTime, endTime)
	}

	return {
		loading,
		error,
		evaluationList,
		evaluationData,
		evaluationCache,
		showClearCacheConfirm,
		lastRefreshTime,
		isPreviewMode,
		navPaddingTop,
		refreshUnreadMessages,
		showPreviewData,
		goToLogin,
		handleNavHeightReady,
		handleBackClick,
		handleTitleClick,
		refreshEvaluationData,
		loadEvaluationList,
		handleSessionExpired,
		loadAllEvaluationData,
		loadSingleEvaluationData,
		loadCacheDataMixin,
		getCacheStatusMixin,
		submitAllEvaluations,
		processBatchSubmit,
		updateListCacheMixin,
		clearAllCache,
		cancelClearCache,
		confirmClearCache,
		goToSwipeEvaluation,
		quickEvaluateAll,
		mergeUserScoresMixin,
		showSubmitConfirmationMixin,
		prepareEvaluationDataMixin,
		processQuickEvaluateAndSubmitMixin,
		updateEvaluationStatusMixin,
		goToEvaluate,
		formatTimeRangeMixin
	}
}