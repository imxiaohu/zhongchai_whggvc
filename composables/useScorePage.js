/**
 * Score Page Composable
 * 成绩页面的状态和方法
 */

import { computed, ref } from 'vue'
import {
	LOADING_SEMESTER,
	generatePreviewSemesterList,
	generatePreviewScoreData,
	generatePreviewStats,
	formatSemesterList,
	getInitialStats,
	formatStatsData,
	isValidSemesterIndex,
	findPreviousSemester,
	findNextSemester,
	isServerMaintenanceError,
	getCurrentUserSemester
} from '@/utils/scoreDisplay.js'
import {
	getSemesterList,
	getCourseScores,
	getSemesterStatistics,
	calculateSemesterStats,
	formatCourseScores,
	getScoreClass as getScoreClassUtil
} from '@/pages/api/score.js'

export function useScorePage() {
	const loading = ref(false)
	const semesterList = ref([LOADING_SEMESTER])
	const semesterIndex = ref(0)
	const courseScores = ref([])
	const semesterStats = ref(getInitialStats())
	const hasData = ref(false)
	const error = ref(null)
	const showCacheBanner = ref(false)
	const cacheUpdatedAt = ref('')
	const isPreviewMode = ref(false)

	const currentSemesterName = computed(() =>
		semesterList.value[semesterIndex.value]?.name || '请选择学期'
	)
	const currentSemesterValue = computed(() =>
		semesterList.value[semesterIndex.value]?.value || ''
	)
	const isCurrentSemesterSelected = computed(() => {
		const currentUserSem = getCurrentUserSemester()
		if (!currentUserSem) return false
		const currentValue = currentSemesterValue.value
		return currentValue === currentUserSem ||
			currentValue.includes(currentUserSem) ||
			currentUserSem.includes(currentValue)
	})

	function applyCacheMeta(meta) {
		if (!meta) return
		if (meta.fromCache) {
			showCacheBanner.value = true
			cacheUpdatedAt.value = meta.cacheUpdatedAt || ''
		}
	}

	function checkLoginStatus() {
		const token = uni.getStorageSync('token')
		if (!token) {
			showPreviewData()
			return false
		}
		return true
	}

	function showPreviewData() {
		loading.value = false
		isPreviewMode.value = true
		hasData.value = true

		semesterList.value = generatePreviewSemesterList()
		courseScores.value = generatePreviewScoreData()
		semesterStats.value = generatePreviewStats()
		semesterIndex.value = 0

		uni.showModal({
			title: '预览模式',
			content: '当前显示的是成绩预览数据，登录后可获取真实的成绩信息',
			confirmText: '立即登录',
			cancelText: '稍后再说',
			success: (res) => {
				if (res.confirm) {
					uni.navigateTo({ url: '/pages/login/login' })
				}
			}
		})
	}

	function filterPreviewDataBySemester(semester) {
		const allPreviewData = generatePreviewScoreData()

		if (semester === '全部') {
			courseScores.value = allPreviewData
		} else {
			courseScores.value = allPreviewData.filter(course =>
				course.semester === semester
			)
		}

		calculateStatistics()
	}

	function goToLogin() {
		uni.vibrateShort()
		uni.navigateTo({ url: '/pages/login/login' })
	}

	async function initializeData() {
		try {
			await loadAllScoreData()
			await loadSemesterList()
		} catch (error) {
			handleError('加载失败，请稍后重试', error)
		}
	}

	async function loadAllScoreData() {
		try {
			await loadCourseScores('全部')
		} catch (error) {
			console.error('加载全部成绩数据失败:', error)
		}
	}

	async function loadSemesterList() {
		try {
			const semesterData = await getSemesterList()

			if (Array.isArray(semesterData) && semesterData.length > 0) {
				semesterList.value = formatSemesterList(semesterData)
			} else {
				semesterList.value = [{ name: '加载失败', value: '全部' }]
			}

			semesterIndex.value = 0
		} catch (error) {
			semesterList.value = [{ name: '全部学期', value: '全部' }]
			semesterIndex.value = 0
		}
	}

	async function handleSemesterChange(e) {
		const index = parseInt(e.detail.value)
		if (!isValidSemesterIndex(index, semesterList.value)) return

		semesterIndex.value = index
		const semester = currentSemesterValue.value

		if (isPreviewMode.value) {
			filterPreviewDataBySemester(semester)
			return
		}

		if (semester) {
			await loadScoreData(semester)
		}
	}

	function openSemesterPicker() {
		try {
			if (Array.isArray(semesterList.value) && semesterList.value.length > 0) {
				uni.showActionSheet({
					itemList: semesterList.value.map(s => s.name || s.value || ''),
					success: (res) => {
						const idx = res.tapIndex
						handleSemesterChange({ detail: { value: idx } })
					}
				})
			}
		} catch (error) {
			console.error('openSemesterPicker failed:', error)
		}
	}

	async function loadScoreData(semester) {
		if (!semester) {
			showToast('请选择有效的学期')
			return
		}

		try {
			await Promise.all([
				loadCourseScores(semester),
				loadSemesterStatistics(semester)
			])
		} catch (error) {
			handleError('加载成绩数据失败', error)
		}
	}

	async function loadCourseScores(semester) {
		loading.value = true
		courseScores.value = []
		hasData.value = false

		try {
			const resp = await getCourseScores(semester)
			applyCacheMeta(resp.meta)
			const scoreData = resp.data

			if (Array.isArray(scoreData)) {
				courseScores.value = formatCourseScores(scoreData)
				calculateStatistics()
			} else {
				showToast('暂无成绩数据')
			}
		} catch (error) {
			handleError('获取成绩失败', error)
		} finally {
			loading.value = false
		}
	}

	async function loadSemesterStatistics(semester) {
		try {
			const resp = await getSemesterStatistics(semester)
			applyCacheMeta(resp.meta)
			const statsData = resp.data

			if (statsData && typeof statsData === 'object') {
				semesterStats.value = formatStatsData(statsData)
				hasData.value = true
			} else {
				calculateStatistics()
			}
		} catch (error) {
			calculateStatistics()
		}
	}

	function calculateStatistics() {
		if (courseScores.value.length === 0) {
			hasData.value = false
			semesterStats.value = getInitialStats()
			return
		}

		semesterStats.value = calculateSemesterStats(courseScores.value)
		hasData.value = true
	}

	function getScoreClass(course) {
		const finalScore = parseFloat(course.finalScore)
		if (!isNaN(finalScore)) {
			return getScoreClassUtil(finalScore)
		}

		const courseScore = parseFloat(course.courseScore)
		if (!isNaN(courseScore)) {
			return getScoreClassUtil(courseScore)
		}

		const letterGrade = course.finalScore
		if (letterGrade === 'A' || letterGrade === 'A+') return 'score-excellent'
		if (letterGrade === 'B' || letterGrade === 'B+') return 'score-good'
		if (letterGrade === 'C' || letterGrade === 'C+') return 'score-medium'
		if (letterGrade === 'D' || letterGrade === 'D+') return 'score-pass'
		if (letterGrade === 'F' || letterGrade === 'E') return 'score-fail'

		return 'score-medium'
	}

	function handleError(message, err) {
		error.value = err
		if (isServerMaintenanceError(err)) {
			showToast('学校服务器维护中，请稍后再试')
		} else {
			showToast(message)
		}
	}

	function showToast(title, icon = 'none') {
		uni.showToast({ title, icon })
	}

	async function quickQueryCurrentSemester() {
		if (isPreviewMode.value) {
			uni.showModal({
				title: '预览模式',
				content: '这是预览数据，登录后可以查看真实的本学期成绩',
				confirmText: '立即登录',
				cancelText: '取消',
				success: (res) => {
					if (res.confirm) {
						uni.navigateTo({ url: '/pages/login/login' })
					}
				}
			})
			return
		}

		const currentUserSemester = getCurrentUserSemester()
		if (!currentUserSemester) {
			showToast('无法获取当前学期信息')
			return
		}

		const currentIndex = semesterList.value.findIndex(semester =>
			semester.value === currentUserSemester ||
			semester.value.includes(currentUserSemester) ||
			currentUserSemester.includes(semester.value)
		)

		if (currentIndex !== -1) {
			semesterIndex.value = currentIndex
			await loadScoreData(currentSemesterValue.value)
		} else {
			await loadScoreData(currentUserSemester)
		}
	}

	async function quickQueryPreviousSemester() {
		if (isPreviewMode.value) {
			uni.showModal({
				title: '预览模式',
				content: '这是预览数据，登录后可以查看真实的上学期成绩',
				confirmText: '立即登录',
				cancelText: '取消',
				success: (res) => {
					if (res.confirm) {
						uni.navigateTo({ url: '/pages/login/login' })
					}
				}
			})
			return
		}

		const previousSemester = findPreviousSemester(semesterList.value)
		if (!previousSemester) {
			showToast('未找到上学期数据')
			return
		}

		const previousIndex = semesterList.value.findIndex(semester =>
			semester.value === previousSemester
		)

		if (previousIndex !== -1) {
			semesterIndex.value = previousIndex
			await loadScoreData(currentSemesterValue.value)
		} else {
			await loadScoreData(previousSemester)
		}
	}

	async function quickQueryNextSemester() {
		if (isPreviewMode.value) {
			uni.showModal({
				title: '预览模式',
				content: '这是预览数据，登录后可以查看真实的下学期成绩',
				confirmText: '立即登录',
				cancelText: '取消',
				success: (res) => {
					if (res.confirm) {
						uni.navigateTo({ url: '/pages/login/login' })
					}
				}
			})
			return
		}

		const nextSemester = findNextSemester(semesterList.value)
		if (!nextSemester) {
			showToast('未找到下学期数据')
			return
		}

		const nextIndex = semesterList.value.findIndex(semester =>
			semester.value === nextSemester
		)

		if (nextIndex !== -1) {
			semesterIndex.value = nextIndex
			await loadScoreData(currentSemesterValue.value)
		} else {
			await loadScoreData(nextSemester)
		}
	}

	return {
		loading,
		semesterList,
		semesterIndex,
		courseScores,
		semesterStats,
		hasData,
		error,
		showCacheBanner,
		cacheUpdatedAt,
		isPreviewMode,
		currentSemesterName,
		currentSemesterValue,
		isCurrentSemesterSelected,
		applyCacheMeta,
		checkLoginStatus,
		showPreviewData,
		filterPreviewDataBySemester,
		goToLogin,
		initializeData,
		loadAllScoreData,
		loadSemesterList,
		handleSemesterChange,
		openSemesterPicker,
		loadScoreData,
		loadCourseScores,
		loadSemesterStatistics,
		calculateStatistics,
		getScoreClass,
		handleError,
		showToast,
		quickQueryCurrentSemester,
		quickQueryPreviousSemester,
		quickQueryNextSemester
	}
}