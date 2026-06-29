/**
 * Evaluation Form Composable
 * 评教表单逻辑：数据属性和方法
 */

import { computed, ref } from 'vue'
import {
	EVALUATION_CONSTANTS,
	saveEvaluationCache,
	loadEvaluationCache,
	removeEvaluationCache,
	clearAllEvaluationCache,
	getEvaluationListCache,
	setEvaluationListCache,
	calculateTotalScore,
	calculateCurrentScore,
	formatTimeRange,
	formatSingleDate,
	processEvaluationDetail,
	createDefaultEvaluationData,
	getScoreClass
} from '@/utils/evaluation.js'

export function useEvaluationForm() {
	const evaluationList = ref([])
	const currentIndex = ref(0)
	const evaluationData = ref({})
	const loading = ref(true)
	const swiperHeight = ref(EVALUATION_CONSTANTS.DEFAULT_SWIPER_HEIGHT)
	const quickComments = ref(EVALUATION_CONSTANTS.QUICK_COMMENTS)
	const calculationCache = ref({})
	const statusBarHeight = ref(0)
	const saveTimers = ref({})

	const CONSTANTS = EVALUATION_CONSTANTS

	const hasCompletedEvaluations = computed(() =>
		Object.values(evaluationData.value).some(data => data && data.isCompleted)
	)

	const hasUnevaluatedItems = computed(() =>
		evaluationList.value.some(item => !item.evaluated)
	)

	const currentEvaluationItem = computed(() =>
		evaluationList.value[currentIndex.value] || null
	)

	const currentEvaluationId = computed(() =>
		currentEvaluationItem.value?.id || null
	)

	function getCurrentItemId() {
		return currentEvaluationId.value
	}

	function getNormList(id) {
		if (!evaluationData.value[id] || !evaluationData.value[id].normList) {
			return []
		}
		return evaluationData.value[id].normList
	}

	function getComment(id) {
		if (!evaluationData.value[id]) return ''
		return evaluationData.value[id].comment || ''
	}

	function updateComment(id, comment) {
		if (!evaluationData.value[id]) {
			evaluationData.value[id] = {
				normList: [],
				comment: '',
				isCompleted: false
			}
		}
		evaluationData.value[id].comment = comment
		saveEvaluation(id)
	}

	function selectQuickComment(id, comment) {
		if (!evaluationData.value[id]) {
			evaluationData.value[id] = {
				normList: [],
				comment: '',
				isCompleted: false
			}
		}
		evaluationData.value[id].comment = comment
		saveEvaluation(id)
	}

  function selectScore(id, normIndex, score) {
    const data = evaluationData.value[id]
    if (!data || !data.normList || !data.normList[normIndex]) return

    data.normList[normIndex].score = score
    data.isCompleted = checkIsCompleted(data.normList)
    clearCalculationCache(id)
    debouncedSave(id)
  }

	function setMaxScore(id) {
		const data = evaluationData.value[id]
		if (!data || !data.normList) return

		data.normList.forEach(norm => {
			norm.score = EVALUATION_CONSTANTS.MAX_SCORE
		})

		data.isCompleted = true
		clearCalculationCache(id)
		saveEvaluation(id)

		uni.showToast('已设置满分')
	}

	function checkIsCompleted(normList) {
		if (!normList || normList.length === 0) return false
		return normList.every(norm => norm.score > 0)
	}

	function isCompleted(id) {
		if (!evaluationData.value[id]) return false
		return evaluationData.value[id].isCompleted
	}

	function debouncedSave(id) {
		if (saveTimers.value[id]) {
			clearTimeout(saveTimers.value[id])
		}

		if (!saveTimers.value) saveTimers.value = {}

		saveTimers.value[id] = setTimeout(() => {
			saveEvaluation(id)
		}, 500)
	}

	function saveEvaluation(id) {
		try {
			const data = evaluationData.value[id]
			if (!id || !data) return

			const cacheData = {
				normList: data.normList,
				comment: data.comment,
				isCompleted: data.isCompleted,
				timestamp: Date.now()
			}

			const cacheKey = `${EVALUATION_CONSTANTS.CACHE_PREFIX}${id}`
			uni.setStorageSync(cacheKey, JSON.stringify(cacheData))
			updateListCache(id)
		} catch (e) {
			console.error('保存评教数据失败:', e)
		}
	}

	function clearCalculationCache(id) {
		if (calculationCache.value[id]) {
			delete calculationCache.value[id]
		}
	}

	function saveAllEvaluations() {
		const ids = Object.keys(evaluationData.value)
		if (ids.length === 0) return

		try {
			ids.forEach(id => saveEvaluation(id))
		} catch (error) {
			console.error('批量保存失败:', error)
		}
	}

	function updateListCache(id) {
		try {
			const listCache = getListCache()

			if (id) {
				updateSingleItemCache(listCache, id)
			} else {
				updateAllItemsCache(listCache)
			}

			uni.setStorageSync(EVALUATION_CONSTANTS.LIST_CACHE_KEY, JSON.stringify(listCache))
		} catch (e) {
			console.error('更新列表缓存失败:', e)
		}
	}

	function getListCache() {
		return getEvaluationListCache()
	}

	function updateSingleItemCache(listCache, id) {
		const item = evaluationList.value.find(item => item.id == id)
		const data = evaluationData.value[id]

		if (item && data) {
			listCache[id] = createCacheItem(item, data)
		}
	}

	function updateAllItemsCache(listCache) {
		evaluationList.value.forEach(item => {
			const data = evaluationData.value[item.id]
			if (data) {
				listCache[item.id] = createCacheItem(item, data)
			}
		})
	}

	function createCacheItem(item, data) {
		return {
			id: item.id,
			isCompleted: data.isCompleted,
			timestamp: Date.now(),
			courseName: item.name,
			teacherName: item.teacherName
		}
	}

	function saveCurrentEvaluation() {
		const id = getCurrentItemId()
		if (!id) {
			uni.showToast('无法获取当前评教ID')
			return
		}

		try {
			saveEvaluation(id)
			uni.showToast('已保存', 'success')
		} catch (error) {
			uni.showToast('保存失败')
		}
	}

	 async function loadSingleEvaluationData(item) {
		try {
			if (!item?.id || item.id === 'undefined' || item.id === 'null') {
				console.warn('[useEvaluationForm] skip loadSingleEvaluationData: invalid item.id', item?.id)
				return
			}

			const cached = loadFromCache(item.id)
			if (cached) {
				evaluationData.value[item.id] = cached
				return
			}

			// 如果列表数据已包含评教题目，直接使用
			if (item && Array.isArray(item.courseTeachingEvaluationQuestionsVOS) && item.courseTeachingEvaluationQuestionsVOS.length > 0) {
				const normList = item.courseTeachingEvaluationQuestionsVOS.map(norm => ({ ...norm, score: 0 }))
				evaluationData.value[item.id] = {
					normList,
					comment: '',
					isCompleted: false
				}
				return
			}

			const { getEvaluationDetail } = await import('@/pages/api/evaluation.js')
			const detail = await getEvaluationDetail(item.id)
			evaluationData.value[item.id] = processEvaluationDetail(detail)
		} catch (error) {
			console.error(`加载评教详情失败 ID:${item.id}`, error)
			evaluationData.value[item.id] = createDefaultEvaluationData()
		}
	}

	function loadFromCache(id) {
		try {
			const cacheKey = `${EVALUATION_CONSTANTS.CACHE_PREFIX}${id}`
			const cacheStr = uni.getStorageSync(cacheKey)

			if (cacheStr) {
				const cacheData = JSON.parse(cacheStr)
				return {
					normList: cacheData.normList || [],
					comment: cacheData.comment || '',
					isCompleted: cacheData.isCompleted || false
				}
			}
		} catch (error) {
			console.error('加载缓存失败:', error)
		}
		return null
	}

	function loadCachedData() {
		try {
			const listCacheKey = EVALUATION_CONSTANTS.LIST_CACHE_KEY
			const listCacheStr = uni.getStorageSync(listCacheKey)

			if (listCacheStr) {
				const listCache = JSON.parse(listCacheStr)

				Object.keys(listCache).forEach(id => {
					if (evaluationData.value[id]) {
						evaluationData.value[id].isCompleted = listCache[id].isCompleted || false
					}
				})
			}

			evaluationList.value.forEach(item => {
				const cacheKey = `${EVALUATION_CONSTANTS.CACHE_PREFIX}${item.id}`
				const cacheStr = uni.getStorageSync(cacheKey)

				if (cacheStr) {
					const cacheData = JSON.parse(cacheStr)

					if (cacheData.normList && cacheData.normList.length > 0) {
						evaluationData.value[item.id] = {
							normList: cacheData.normList,
							comment: cacheData.comment || '',
							isCompleted: checkIsCompleted(cacheData.normList)
						}
					}
				}
			})
		} catch (e) {
			console.error('加载缓存数据失败:', e)
		}
	}

	function calculateTotalScoreValue(id) {
		const normList = getNormList(id)
		return calculateTotalScore(normList, EVALUATION_CONSTANTS.MAX_SCORE)
	}

	function calculateCurrentScoreValue(id) {
		const normList = getNormList(id)
		return calculateCurrentScore(normList)
	}

	function calculateScorePercentage(id) {
		const totalScore = calculateTotalScoreValue(id)
		if (totalScore === 0) return 0

		const currentScore = calculateCurrentScoreValue(id)
		return Math.round((currentScore / totalScore) * 100)
	}

	function formatTimeRangeValue(startTime, endTime) {
		return formatTimeRange(startTime, endTime)
	}

	function getScoreClassValue(percentage) {
		return getScoreClass(percentage)
	}

	function cleanup() {
		Object.values(saveTimers.value).forEach(timer => clearTimeout(timer))
		saveTimers.value = {}
	}

	return {
		evaluationList,
		currentIndex,
		evaluationData,
		loading,
		swiperHeight,
		quickComments,
		calculationCache,
		statusBarHeight,
		saveTimers,
		CONSTANTS,
		hasCompletedEvaluations,
		hasUnevaluatedItems,
		currentEvaluationItem,
		currentEvaluationId,
		getCurrentItemId,
		getNormList,
		getComment,
		updateComment,
		selectQuickComment,
		selectScore,
		setMaxScore,
		checkIsCompleted,
		isCompleted,
		debouncedSave,
		saveEvaluation,
		clearCalculationCache,
		saveAllEvaluations,
		updateListCache,
		getListCache,
		updateSingleItemCache,
		updateAllItemsCache,
		createCacheItem,
		saveCurrentEvaluation,
		loadSingleEvaluationData,
		loadFromCache,
		loadCachedData,
		calculateTotalScore: calculateTotalScoreValue,
		calculateCurrentScore: calculateCurrentScoreValue,
		calculateScorePercentage,
		formatTimeRange: formatTimeRangeValue,
		getScoreClass: getScoreClassValue,
		cleanup
	}
}