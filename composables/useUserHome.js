/**
 * User Home Composable
 * 用户主页的状态和方法
 */

import { computed, ref } from 'vue'
import { shouldAutoRefreshData } from '@/utils/errorHandler.js'
import { checkSchoolServerStatus } from '@/utils/schoolServerStatus.js'
import { useUserStore } from '@/store/user.js'
import { useSchoolAccountStore } from '@/store/schoolAccount.js'
import unreadMessageManager from '@/utils/unreadMessageManager.js'
import shareManager from '@/utils/shareManager.js'
import { COMMUNITY_MODULE_ENABLE } from '@/config/modules.js'
import { getUserStatistics } from '@/pages/api/user.js'
import { getAllCourseScores } from '@/pages/api/score.js'
import { calculateSemesterStats } from '@/pages/api/score.js'
import { getUserArchive } from '@/pages/api/discover.js'
import {
	getDefaultUserInfo,
	getDefaultUserStats,
	getDefaultServerStatus,
	formatUserInfoDisplay,
	formatUserStats,
	shouldFetchUserStats,
	getFacultyName,
	getBindMenuText
} from '@/utils/userHome.js'

export function useUserHome() {
	// 使用 Pinia Store 管理用户状态
	const userStore = useUserStore()
	const schoolAccountStore = useSchoolAccountStore()

	// 确保 store 状态与本地存储同步
	userStore.initFromStorage()

	const userInfo = ref(getDefaultUserInfo())
	const loginType = ref('')
	const isLoggedIn = computed(() => userStore.isLoggedIn)
	const unreadCount = ref(0)
	const userStats = ref(getDefaultUserStats())
	const serverStatus = ref(getDefaultServerStatus())

	const isSchoolAccountBound = computed(() =>
		schoolAccountStore.checkBinding()
	)

	const bindMenuText = computed(() =>
		getBindMenuText(isSchoolAccountBound.value)
	)

	const isCommunityEnabled = computed(() =>
		COMMUNITY_MODULE_ENABLE
	)

	async function loadUserInfo() {
		try {
			const lt = uni.getStorageSync('loginType')
			loginType.value = lt

			const token = uni.getStorageSync('token')
			if (!token) {
				userStore.logout()
				userInfo.value = getDefaultUserInfo()
				return
			}

			let baseInfo = null

			try {
				const response = await uni.request({
					url: '/api/user/info',
					method: 'GET',
					header: {
						'Authorization': `Bearer ${token}`,
						'Content-Type': 'application/json'
					}
				})

				if (response.statusCode === 200 && response.data.success) {
					baseInfo = response.data.result

					schoolAccountStore.updateBinding(
						baseInfo.hasSchoolAccount || false,
						baseInfo
					)
				} else {
					loadUserInfoFromStorage()
					return
				}
			} catch (apiError) {
				loadUserInfoFromStorage()
				return
			}

			// 学校账号登录后，优先从 archive 获取完整档案信息（包含 facultyName 等字段）
			if (lt === 'school' && baseInfo && baseInfo.hasSchoolAccount) {
				try {
					const archiveRes = await getUserArchive()
					if (archiveRes && archiveRes.success && archiveRes.result) {
						const archive = archiveRes.result
						baseInfo = {
							...baseInfo,
							...archive,
							college: archive.facultyName || archive.college || baseInfo.college || '',
							className: archive.adminClass || archive.className || baseInfo.className || ''
						}
					}
				} catch (e) {
					console.warn('获取档案信息失败，不影响页面展示', e)
				}
			}

			updateUserInfoDisplay(baseInfo)
			uni.setStorageSync('userInfo', JSON.stringify(baseInfo))
		} catch (error) {
			userInfo.value = getDefaultUserInfo()
		}
	}

	async function refreshUserStatistics() {
		try {
			if (!shouldFetchUserStats(isLoggedIn.value, loginType.value, isSchoolAccountBound.value)) {
				userStats.value = getDefaultUserStats()
				return
			}

			const [statsRes, scoreRes] = await Promise.all([
				getUserStatistics().catch(() => null),
				getAllCourseScores().catch(() => null)
			])

			const stats = statsRes?.result || statsRes || {}
			const scoreData = Array.isArray(scoreRes?.data) ? scoreRes.data
				: Array.isArray(scoreRes) ? scoreRes : []

			if (scoreData.length > 0) {
				const allScores = scoreData.map(item => ({
					finalScore: item.finalScore ?? item.courseScore ?? 0,
					courseScore: item.courseScore ?? 0,
					credit: item.credit ?? item.courseCredit ?? 0,
					getPoint: item.getPoint ?? item.gpa ?? 0
				}))
				const computed = calculateSemesterStats(allScores)
				stats.averageScore = computed.averageScore
				stats.creditTotal = computed.creditTotal
				stats.gpa = computed.gpa
			}

			userStats.value = formatUserStats(stats)
		} catch (error) {
			userStats.value = getDefaultUserStats()
		}
	}

	function loadUserInfoFromStorage() {
		const token = uni.getStorageSync('token')
		userStore.initFromStorage()

		loginType.value = uni.getStorageSync('loginType') || ''

		const userInfoStr = uni.getStorageSync('userInfo')
		if (userInfoStr && token) {
			try {
				const info = JSON.parse(userInfoStr)
				updateUserInfoDisplay(info)
				schoolAccountStore.initialize()
			} catch (e) {
				userInfo.value = getDefaultUserInfo()
			}
		} else {
			userInfo.value = getDefaultUserInfo()
		}
	}

	function updateUserInfoDisplay(info) {
		userInfo.value = formatUserInfoDisplay(info, loginType.value)
	}

	function handleUserInfoClick() {
		if (!isLoggedIn.value) {
			handleGoToLogin()
		}
	}

	function handleGoToLogin() {
		uni.vibrateShort()
		uni.navigateTo({ url: '/pages/login/login' })
	}

	function handleNavigateToScore() {
		uni.vibrateShort()
		if (!isLoggedIn.value) {
			handleGoToLogin()
			return
		}
		uni.navigateTo({ url: '/pages/user/score' })
	}

	function handleNavigateToBind() {
		uni.vibrateShort()
		if (!isLoggedIn.value) {
			handleGoToLogin()
			return
		}

		if (isSchoolAccountBound.value) {
			uni.showActionSheet({
				itemList: ['查看绑定信息', '重新绑定', '解除绑定'],
				success: (res) => {
					switch (res.tapIndex) {
						case 0: showBindingInfo(); break
						case 1: confirmRebind(); break
						case 2: confirmUnbind(); break
					}
				}
			})
		} else {
			uni.navigateTo({ url: '/pages/user/bind' })
		}
	}

	function showBindingInfo() {
		try {
			const userInfoStr = uni.getStorageSync('userInfo')
			if (userInfoStr) {
				const info = JSON.parse(userInfoStr)
				uni.showModal({
					title: '绑定信息',
					content: `姓名：${info.realname || '未知'}\n学号：${info.username || '未知'}\n学院：${getFacultyName(info.facultyId) || info.college || '未知'}`,
					showCancel: false,
					confirmText: '确定'
				})
			}
		} catch (error) {
			uni.showToast({ title: '获取绑定信息失败', icon: 'none' })
		}
	}

	function confirmRebind() {
		uni.showModal({
			title: '重新绑定',
			content: '确定要重新绑定学校账号吗？重新绑定后需要重新验证。',
			success: (res) => {
				if (res.confirm) {
					schoolAccountStore.updateBinding(false)
					uni.navigateTo({ url: '/pages/user/bind' })
				}
			}
		})
	}

	function confirmUnbind() {
		uni.showModal({
			title: '解除绑定',
			content: '确定要解除绑定学校账号吗？解除绑定后将无法使用需要学校账号的功能。',
			success: (res) => {
				if (res.confirm) {
					unbindSchoolAccount()
				}
			}
		})
	}

	function unbindSchoolAccount() {
		try {
			const userInfoStr = uni.getStorageSync('userInfo')
			let updatedUserInfo = null

			if (userInfoStr) {
				const info = JSON.parse(userInfoStr)
				updatedUserInfo = {
					id: info.id,
					nickname: info.nickname,
					avatarUrl: info.avatarUrl,
					hasSchoolAccount: false
				}
			}

			schoolAccountStore.updateBinding(false, updatedUserInfo)
			loadUserInfo()

			uni.showToast({ title: '解除绑定成功', icon: 'success' })
		} catch (error) {
			uni.showToast({ title: '解除绑定失败', icon: 'none' })
		}
	}

	function handleNavigateToSettings() {
		uni.vibrateShort()
		if (!isLoggedIn.value) {
			handleGoToLogin()
			return
		}
		uni.navigateTo({ url: '/pages/user/setting' })
	}

	function handleLogout() {
		uni.vibrateShort()
		uni.showModal({
			title: '提示',
			content: '确定要退出登录吗？',
			success: (res) => {
				if (res.confirm) {
					const savedUsername = uni.getStorageSync('saved_username')
					const savedPassword = uni.getStorageSync('saved_password')
					const rememberPassword = uni.getStorageSync('remember_password')

					uni.clearStorageSync()

					if (rememberPassword) {
						uni.setStorageSync('saved_username', savedUsername)
						uni.setStorageSync('saved_password', savedPassword)
						uni.setStorageSync('remember_password', rememberPassword)
					}

					uni.reLaunch({ url: '/pages/login/login' })
				}
			}
		})
	}

	async function refreshUnreadMessages() {
		try {
			await unreadMessageManager.fetchUnreadCount()
			unreadCount.value = unreadMessageManager.getUnreadCount()
			unreadMessageManager.updateTabBarBadge()
		} catch (error) {
			console.error('刷新未读消息数量失败:', error)
		}
	}

	function handleNavigateToBookmark() {
		uni.navigateTo({ url: '/pages/community/bookmark-list' })
	}

	function handleNavigateToNotificationCenter() {
		uni.vibrateShort()
		if (!isLoggedIn.value) {
			handleGoToLogin()
			return
		}
		uni.navigateTo({ url: '/pages/user/notification-center' })
	}

	function handleNavigateToProfileEdit() {
		uni.vibrateShort()
		if (!isLoggedIn.value) {
			handleGoToLogin()
			return
		}
		uni.navigateTo({ url: '/pages/user/profile-edit' })
	}

	async function loadServerStatus() {
		try {
			const status = await checkSchoolServerStatus()
			serverStatus.value = status
		} catch (error) {
			serverStatus.value = getDefaultServerStatus()
			serverStatus.value.errorMsg = '获取状态失败'
		}
	}

	function handleNavigateToServerStatus() {
		uni.vibrateShort()
		uni.navigateTo({ url: '/pages/user/server-status' })
	}

	function onPageShowLogic() {
		refreshUnreadMessages()

		try {
			const shouldRefresh = shouldAutoRefreshData()
			loadUserInfo().finally(() => {
				refreshUserStatistics()
			})
		} catch (error) {
			loadUserInfoFromStorage()
			refreshUserStatistics()
		}
	}

	return {
		userInfo,
		loginType,
		isLoggedIn,
		unreadCount,
		userStats,
		serverStatus,
		isSchoolAccountBound,
		bindMenuText,
		isCommunityEnabled,
		loadUserInfo,
		refreshUserStatistics,
		loadUserInfoFromStorage,
		updateUserInfoDisplay,
		handleUserInfoClick,
		handleGoToLogin,
		handleNavigateToScore,
		handleNavigateToBind,
		showBindingInfo,
		confirmRebind,
		confirmUnbind,
		unbindSchoolAccount,
		handleNavigateToSettings,
		handleLogout,
		refreshUnreadMessages,
		handleNavigateToBookmark,
		handleNavigateToNotificationCenter,
		handleNavigateToProfileEdit,
		loadServerStatus,
		handleNavigateToServerStatus,
		onPageShowLogic
	}
}