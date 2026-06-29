/**
 * Schedule Composable
 * Schedule data state, week navigation methods, loading methods, and handlers
 */

import { computed, ref, watch } from 'vue'
import { request } from '@/utils/request.js'
import { useCourseCache } from '@/store/courseCache.js'
import { getMCourseTimeTableByWeek } from '@/pages/api/schedule.js'
import {
	DEFAULT_TIME_SLOTS,
	generateWeekArray,
	createSemesterInfo,
	generatePreviewTimePeriods,
	generatePreviewCourseData,
	isServerMaintenanceError
} from '@/utils/scheduleDisplay.js'

export function useSchedule() {
	const courses = ref([])
	const timePeriods = ref([])

	const currentWeek = ref(1)
	const totalWeeks = ref(20)
	const displayWeek = ref(1)

	const isDataLoaded = ref(false)
	const isWeeksLoaded = ref(false)
	const isLoading = ref(false)
	const isWeekChanging = ref(false)

	const currentSemester = ref('2024-2025学年第二学期')

	const showSetCurrentWeekBtn = ref(true)
	const showSetCurrentWeekPopup = ref(false)
	const tempCurrentWeek = ref(1)

	const lastSwipeDirection = ref('')

	const isPreviewMode = ref(false)

	const showCacheBanner = ref(false)
	const cacheUpdatedAt = ref('')

	const isCurrentTerm = computed(() =>
		currentWeek.value > 0 && currentWeek.value <= totalWeeks.value
	)

	const weekArray = computed(() => generateWeekArray(totalWeeks.value))

	const semesterInfo = computed(() =>
		createSemesterInfo(currentSemester.value, currentWeek.value, totalWeeks.value)
	)

	watch(currentWeek, (newVal) => {
		tempCurrentWeek.value = newVal
	})

	async function fetchTermWeekInfo() {
		try {
			const res = await request({
				url: '/scloudoa/scs/course/tCourseTimetableDetail/getCurrentTime',
				method: 'GET'
			})

			let apiCurrentWeek = null
			if (res && res.result) {
				let result = res.result

				if (result.result && typeof result.result === 'object') {
					result = result.result
				}

				if (result.nowweek) {
					apiCurrentWeek = parseInt(result.nowweek)
				}
				if (result.currentSemester) {
					currentSemester.value = result.currentSemester
				}
				if (result.weekCount) {
					totalWeeks.value = parseInt(result.weekCount)
				} else {
					totalWeeks.value = 20
				}
			}

			let savedCurrentWeek = null
			let savedTimestamp = null
			try {
				const saved = uni.getStorageSync('currentWeek')
				const timestamp = uni.getStorageSync('currentWeekTimestamp')
				if (saved) {
					savedCurrentWeek = parseInt(saved)
					savedTimestamp = timestamp ? parseInt(timestamp) : 0
				}
			} catch (storageErr) {
				console.error('读取本地存储失败', storageErr)
			}

			const now = Date.now()
			const isRecentManualSetting = savedTimestamp && (now - savedTimestamp) < 24 * 60 * 60 * 1000

			if (apiCurrentWeek) {
				if (isRecentManualSetting && savedCurrentWeek !== apiCurrentWeek) {
					currentWeek.value = savedCurrentWeek
					displayWeek.value = savedCurrentWeek
					tempCurrentWeek.value = savedCurrentWeek
				} else {
					currentWeek.value = apiCurrentWeek
					displayWeek.value = apiCurrentWeek
					tempCurrentWeek.value = apiCurrentWeek

					uni.setStorageSync('currentWeek', apiCurrentWeek.toString())
					uni.setStorageSync('currentWeekTimestamp', now.toString())
				}
			} else if (savedCurrentWeek) {
				currentWeek.value = savedCurrentWeek
				displayWeek.value = savedCurrentWeek
				tempCurrentWeek.value = savedCurrentWeek
			} else {
				currentWeek.value = 1
				displayWeek.value = 1
				tempCurrentWeek.value = 1
			}

			isWeeksLoaded.value = true
		} catch (error) {
			console.error('获取学期周数信息失败', error)
			isWeeksLoaded.value = true

			try {
				const saved = uni.getStorageSync('currentWeek')
				if (saved) {
					const savedCurrentWeek = parseInt(saved)
					currentWeek.value = savedCurrentWeek
					displayWeek.value = savedCurrentWeek
					tempCurrentWeek.value = savedCurrentWeek
				}
			} catch (storageErr) {
				currentWeek.value = 1
				displayWeek.value = 1
				tempCurrentWeek.value = 1
			}

			uni.showToast({
				title: error.message || '获取学期周数信息失败',
				icon: 'none',
				duration: 3000
			})
		}
	}

	async function fetchTimePeriods() {
		try {
			const courseCacheStore = useCourseCache()
			const result = await courseCacheStore.getCourseLessonTime()

			if (Array.isArray(result) && result.length > 0) {
				timePeriods.value = result
			} else {
				timePeriods.value = [...DEFAULT_TIME_SLOTS]
			}
		} catch (error) {
			console.error('获取课程时间段失败', error)
			timePeriods.value = [...DEFAULT_TIME_SLOTS]

			if (!error.message || !error.message.includes('API返回数据格式错误')) {
				const errorMessage = error.message || '获取课程时间段失败'
				uni.showToast({
					title: errorMessage,
					icon: 'none',
					duration: 3000
				})
			}
		}
	}

		async function fetchCourseData(week = null) {
		try {
			isLoading.value = true
			isDataLoaded.value = false

			const options = {
				currentSemester: currentSemester.value,
				nowWeek: week || displayWeek.value
			}

			const result = await getMCourseTimeTableByWeek(options)

			if (result && typeof result === 'object' && result.courses) {
				courses.value = Array.isArray(result.courses) ? result.courses : []
				isDataLoaded.value = true

				if (result.fromCache) {
					showCacheBanner.value = true
					cacheUpdatedAt.value = result.cacheUpdatedAt || ''
				}
			} else if (Array.isArray(result)) {
				courses.value = result
				isDataLoaded.value = true
			} else {
				courses.value = []
			}
		} catch (error) {
			console.error('获取课表数据失败', error)

			if (error.statusCode === 401 || error.isTokenInvalid ||
				(error.message && error.message.includes('会话已失效'))) {
				handleSessionExpired()
				return
			}

			if (isServerMaintenanceError(error)) {
				uni.showToast({
					title: '学校服务器维护中，请稍后再试',
					icon: 'none',
					duration: 3000
				})
			} else {
				uni.showToast({
					title: error.message || '获取课表数据失败',
					icon: 'none',
					duration: 3000
				})
			}
		} finally {
			isLoading.value = false
		}
	}

	async function initData() {
		console.log('DEBUG initData: 开始执行')
		const token = uni.getStorageSync('token')
		console.log('DEBUG initData: token存在=', !!token)
		if (!token) {
			showPreviewData()
			return
		}

		try {
			console.log('DEBUG initData: 开始加载数据')
			isLoading.value = true
			isPreviewMode.value = false
			await fetchTermWeekInfo()
			await fetchTimePeriods()
			await fetchCourseData(displayWeek.value)
		} catch (error) {
			console.error('初始化数据失败', error)

			if (error.statusCode === 401 || error.isTokenInvalid ||
				(error.message && error.message.includes('会话已失效'))) {
				handleSessionExpired()
				return
			}

			if (isServerMaintenanceError(error)) {
				uni.showToast({
					title: '学校服务器维护中，请稍后再试',
					icon: 'none',
					duration: 3000
				})
			}

			// API 失败时回退到预览模式，确保用户能看到示例数据
			console.log('API 请求失败，回退到预览模式')
			showPreviewData()
		} finally {
			isLoading.value = false
		}
	}

	function showPreviewData() {
		isLoading.value = false
		isDataLoaded.value = true
		isPreviewMode.value = true

		currentWeek.value = 1
		displayWeek.value = 1
		totalWeeks.value = 20
		currentSemester.value = '2024-2025学年第二学期（预览版）'
		timePeriods.value = generatePreviewTimePeriods()
		courses.value = generatePreviewCourseData()
	}

	function goToLogin() {
		uni.vibrateShort()
		uni.navigateTo({
			url: '/pages/login/login'
		})
	}

	function handleSessionExpired() {
		const userInfoStr = uni.getStorageSync('userInfo')
		const hasWechatLogin = !!userInfoStr

		if (hasWechatLogin) {
			uni.showModal({
				title: '会话失效',
				content: '学校账号会话已失效，是否前往重新绑定？',
				confirmText: '去绑定',
				cancelText: '稍后再说',
				success: (res) => {
					if (res.confirm) {
						uni.navigateTo({
							url: '/pages/user/bind'
						})
					}
				}
			})
		} else {
			uni.showModal({
				title: '登录过期',
				content: '登录已过期，是否前往重新登录？',
				confirmText: '去登录',
				cancelText: '稍后再说',
				success: (res) => {
					if (res.confirm) {
						uni.reLaunch({
							url: '/pages/login/login'
						})
					}
				}
			})
		}
	}

	function prevWeek() {
		if (displayWeek.value > 1 && !isWeekChanging.value && !isLoading.value) {
			isWeekChanging.value = true
			lastSwipeDirection.value = 'right'
			displayWeek.value--
			fetchCourseData(displayWeek.value).finally(() => {
				setTimeout(() => {
					isWeekChanging.value = false
				}, 300)
			})
		}
	}

	function nextWeek() {
		if (displayWeek.value < totalWeeks.value && !isWeekChanging.value && !isLoading.value) {
			isWeekChanging.value = true
			lastSwipeDirection.value = 'left'
			displayWeek.value++
			fetchCourseData(displayWeek.value).finally(() => {
				setTimeout(() => {
					isWeekChanging.value = false
				}, 300)
			})
		}
	}

	function changeWeek(weekNum) {
		if (displayWeek.value === weekNum || isLoading.value) return
		displayWeek.value = weekNum
		fetchCourseData(displayWeek.value)
	}

	function onWeekChange(weekNum) {
		if (displayWeek.value === weekNum || isLoading.value) return

		isWeekChanging.value = true
		displayWeek.value = weekNum
		fetchCourseData(displayWeek.value).finally(() => {
			setTimeout(() => {
				isWeekChanging.value = false
			}, 300)
		})
	}

	function onWeekPickerChange(e) {
		tempCurrentWeek.value = parseInt(e.detail.value) + 1
	}

	function setCurrentWeek() {
		currentWeek.value = tempCurrentWeek.value
		displayWeek.value = currentWeek.value

		try {
			const now = Date.now()
			uni.setStorageSync('currentWeek', currentWeek.value.toString())
			uni.setStorageSync('currentWeekTimestamp', now.toString())
		} catch (storageErr) {
			console.error('保存到本地存储失败', storageErr)
		}

		fetchCourseData(displayWeek.value)
		showSetCurrentWeekPopup.value = false

		uni.showToast({
			title: '设置当前周成功',
			icon: 'success'
		})
	}

	function resetToApiCurrentWeek() {
		try {
			uni.removeStorageSync('currentWeek')
			uni.removeStorageSync('currentWeekTimestamp')
			initData()
		} catch (error) {
			console.error('清除本地存储失败', error)
		}
	}

	return {
		courses,
		timePeriods,
		currentWeek,
		totalWeeks,
		displayWeek,
		isDataLoaded,
		isWeeksLoaded,
		isLoading,
		isWeekChanging,
		currentSemester,
		showSetCurrentWeekBtn,
		showSetCurrentWeekPopup,
		tempCurrentWeek,
		lastSwipeDirection,
		isPreviewMode,
		showCacheBanner,
		cacheUpdatedAt,
		isCurrentTerm,
		weekArray,
		semesterInfo,
		fetchTermWeekInfo,
		fetchTimePeriods,
		fetchCourseData,
		initData,
		showPreviewData,
		goToLogin,
		handleSessionExpired,
		prevWeek,
		nextWeek,
		changeWeek,
		onWeekChange,
		onWeekPickerChange,
		setCurrentWeek,
		resetToApiCurrentWeek
	}
}