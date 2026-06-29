/**
 * User Settings Composable
 * 用户设置相关的状态和方法
 */

import { ref } from 'vue'
import {
	THEMES,
	FONT_SIZE_OPTIONS,
	FACULTY_MAP,
	SYNC_STATUS_MAP,
	ABOUT_INFO,
	DEFAULT_SYNC_SETTINGS,
	DEFAULT_SYNC_STATUS,
	DEFAULT_SETTINGS,
	getAppInfo,
	getFacultyName,
	getSyncStatusText,
	calculateCacheSize,
	PRESERVED_DATA_KEYS,
	serializeUserDetail,
	updateUserInfoDisplay,
	getUserInitial
} from '@/utils/userSettings.js'
import { showToast, navigateBack, relaunch } from '@/pages/api/page.js'
import { getSyncSettings, getSyncStatus } from '@/api/sync.js'
import { getUserSettings, updateUserSettings } from '@/pages/api/user.js'
import { useSchoolAccountStore } from '@/store/schoolAccount.js'

export function useUserSettings() {
	const schoolAccountStore = useSchoolAccountStore()
	
	const username = ref('')
	const studentId = ref('')
	const isBound = ref(false)
	const userTags = ref([])
	const userDetail = ref({})

	const appVersion = ref('1.0.0')
	const cacheSize = ref('0.0')
	const hasUpdate = ref(false)
	const updateInfo = ref(null)
	const updateProgress = ref(0)
	const isUpdating = ref(false)

	const syncSettings = ref({
		enabled: false,
		frequency: 'daily',
		timeRange: '08:30-22:20',
		autoRetryEnabled: true,
		maxRetryCount: 3
	})

	const syncStatus = ref({
		syncStatus: 'idle',
		lastSyncAt: null,
		nextSyncAt: null,
		lastSyncMessage: '',
		coursesCount: 0
	})

	function handleNavHeightReady(navInfo) {
		return navInfo.heightPx
	}

	async function loadUserInfo() {
		try {
			const loginType = uni.getStorageSync('loginType')
			const token = uni.getStorageSync('token')
			if (!token) {
				loadUserInfoFromStorage()
				return
			}

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
					const userInfo = response.data.result
					updateUserInfoDisplayMixin(userInfo)

					schoolAccountStore.updateBinding(
						userInfo.hasSchoolAccount || false,
						userInfo
					)
				} else {
					loadUserInfoFromStorage()
				}
			} catch (apiError) {
				loadUserInfoFromStorage()
			}
		} catch (error) {
			setDefaultUserInfo()
		}
	}

	function loadUserInfoFromStorage() {
		try {
			const loginType = uni.getStorageSync('loginType')
			const hasBindSchoolAccount = uni.getStorageSync('hasBindSchoolAccount')
			const userInfoStr = uni.getStorageSync('userInfo')
			if (userInfoStr) {
				try {
					const userInfo = JSON.parse(userInfoStr)
					updateUserInfoDisplayMixin(userInfo)
				} catch (parseError) {
					setDefaultUserInfo()
				}
			} else {
				setDefaultUserInfo()
			}
		} catch (error) {
			setDefaultUserInfo()
		}
	}

	function updateUserInfoDisplayMixin(userInfo) {
		const loginType = uni.getStorageSync('loginType')
		const hasBindSchoolAccount = uni.getStorageSync('hasBindSchoolAccount')
		const info = updateUserInfoDisplay(userInfo, loginType, hasBindSchoolAccount)

		username.value = info.username
		studentId.value = info.studentId
		isBound.value = info.isBound
		userTags.value = info.userTags
		userDetail.value = info.userDetail
	}

	function setDefaultUserInfo() {
		username.value = '未登录'
		studentId.value = '未登录'
		isBound.value = false
	}

	function getUserInitialMixin() {
		return getUserInitial(username.value)
	}

	function showUserDetail() {
		if (!userDetail.value || Object.keys(userDetail.value).length === 0) {
			showToast({ title: '用户信息', icon: 'none' })
			return
		}

		const content = serializeUserDetail(userDetail.value)
		uni.showModal({
			title: '用户信息',
			content: content || '用户信息',
			showCancel: false,
			confirmText: '确定'
		})
	}

	function navigateToBindAccount() {
		if (isBound.value) {
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
				const userInfo = JSON.parse(userInfoStr)
				const studentIdVal = userInfo.username || '未知'
				const realname = userInfo.realname || '未知'
				const college = getFacultyName(userInfo.facultyId) || userInfo.college || '未知'
				const className = userInfo.className || '未知'

				uni.showModal({
					title: '绑定信息',
					content: `真实姓名：${realname}\n学号：${studentIdVal}\n学院：${college}\n班级：${className}`,
					showCancel: false,
					confirmText: '确定'
				})
			}
		} catch (error) {
			showToast({ title: '获取绑定信息失败', icon: 'none' })
		}
	}

	function confirmRebind() {
		uni.showModal({
			title: '重新绑定',
			content: '确定要重新绑定学校账号吗？',
			success: (res) => {
				if (res.confirm) {
					uni.removeStorageSync('hasBindSchoolAccount')
					uni.navigateTo({ url: '/pages/user/bind' })
				}
			}
		})
	}

	function confirmUnbind() {
		uni.showModal({
			title: '解除绑定',
			content: '确定要解除绑定学校账号吗？',
			success: (res) => {
				if (res.confirm) {
					unbindSchoolAccount()
				}
			}
		})
	}

	function unbindSchoolAccount() {
		try {
			uni.removeStorageSync('hasBindSchoolAccount')

			const userInfoStr = uni.getStorageSync('userInfo')
			if (userInfoStr) {
				const userInfo = JSON.parse(userInfoStr)
				const updatedUserInfo = {
					id: userInfo.id,
					nickname: userInfo.nickname,
					avatarUrl: userInfo.avatarUrl,
					hasSchoolAccount: false
				}
				uni.setStorageSync('userInfo', JSON.stringify(updatedUserInfo))
			}

			loadUserInfo()
			showToast({ title: '解除绑定成功', icon: 'success' })
		} catch (error) {
			showToast({ title: '解除绑定失败', icon: 'none' })
		}
	}

	function getFacultyNameMixin(facultyId) {
		return getFacultyName(facultyId)
	}

	function calculateCacheSizeMixin() {
		cacheSize.value = calculateCacheSize()
	}

	function getAppVersion() {
		appVersion.value = '1.0.0'
	}

	async function checkHasUpdate() {
		try {
			const response = await checkForUpdates()
			if (response && response.success) {
				hasUpdate.value = response.result.hasUpdate
				if (response.result.hasUpdate) {
					updateInfo.value = response.result.updateInfo
				}
			}
		} catch (error) {
			hasUpdate.value = false
		}
	}

	function clearCache() {
		uni.showModal({
			title: '确定',
			content: '确定要清空缓存吗？',
			success: (res) => {
				if (res.confirm) {
					const preservedData = {}
					PRESERVED_DATA_KEYS.forEach(key => {
						preservedData[key] = uni.getStorageSync(key)
					})

					uni.clearStorageSync()

					Object.keys(preservedData).forEach(key => {
						if (preservedData[key] !== null && preservedData[key] !== undefined && preservedData[key] !== '') {
							uni.setStorageSync(key, preservedData[key])
						}
					})

					calculateCacheSizeMixin()
					showToast({ title: '缓存已清空', icon: 'success' })
					loadUserInfo()
				}
			}
		})
	}

	async function checkUpdate() {
		try {
			uni.showLoading({ title: '检查更新中...', mask: true })

			const response = await checkForUpdates()
			uni.hideLoading()

			if (response && response.success) {
				const { hasUpdate: hasUpdateVal, updateInfo: updateInfoVal } = response.result

				if (hasUpdateVal) {
					hasUpdate.value = true
					updateInfo.value = updateInfoVal
					showUpdateDialog(updateInfoVal)
				} else {
					showToast({ title: '已是最新版本', icon: 'none' })
				}
			} else {
				throw new Error(response?.message || 'Update check failed')
			}
		} catch (error) {
			uni.hideLoading()
			showToast({ title: '检查更新失败', icon: 'none' })
		}
	}

	function showAbout() {
		uni.showModal({
			title: '关于我们',
			content: `${ABOUT_INFO.description}\n\n版本: v${appVersion.value}\n开发者: ${ABOUT_INFO.developer}`,
			showCancel: false,
			confirmText: '我知道了'
		})
	}

	function logout() {
		uni.showModal({
			title: '确定',
			content: '确定要退出登录吗？',
			success: (res) => {
				if (res.confirm) {
					const savedUsername = uni.getStorageSync('saved_username')
					const savedPassword = uni.getStorageSync('saved_password')
					const rememberPassword = uni.getStorageSync('remember_password')
					const clientId = uni.getStorageSync('clientId')

					uni.clearStorageSync()

					if (rememberPassword) {
						uni.setStorageSync('saved_username', savedUsername)
						uni.setStorageSync('saved_password', savedPassword)
						uni.setStorageSync('remember_password', rememberPassword)
					}

					if (clientId) {
						uni.setStorageSync('clientId', clientId)
					}

					showToast({ title: '退出登录成功', icon: 'success' })

					setTimeout(() => {
						relaunch({ url: '/pages/login/login' })
					}, 1500)
				}
			}
		})
	}

	function goBack() {
		navigateBack()
	}

	async function loadSyncSettings() {
		if (!isBound.value) return

		try {
			const token = uni.getStorageSync('token') || uni.getStorageSync('accessToken')
			if (!token) return

			const [settingsRes, statusRes] = await Promise.allSettled([
				getSyncSettings(),
				getSyncStatus()
			])

			if (settingsRes.status === 'fulfilled') {
				syncSettings.value = settingsRes.value.result
			}

			if (statusRes.status === 'fulfilled') {
				syncStatus.value = statusRes.value.result
			}
		} catch (error) {
			// silent fail
		}
	}

	function getSyncStatusTextMixin() {
		return getSyncStatusText(syncStatus.value, syncSettings.value, isBound.value)
	}

	function showSyncSettings() {
		if (!isBound.value) {
			uni.showModal({
				title: '确定',
				content: '需要绑定账号',
				showCancel: false,
				confirmText: '确定'
			})
			return
		}

		uni.navigateTo({ url: '/pages/user/sync-settings' })
	}

	async function loadUserSettings() {
		try {
			const token = uni.getStorageSync('token') || uni.getStorageSync('accessToken')
			if (!token) return

			const response = await getUserSettings()
			if (response && response.success) {
				// settings loaded
			}
		} catch (error) {
			// silent fail
		}
	}

	async function saveUserSettings(settings) {
		try {
			const token = uni.getStorageSync('token') || uni.getStorageSync('accessToken')
			if (!token) return false

			const response = await updateUserSettings(settings)
			if (response && response.success) return true
			return false
		} catch (error) {
			return false
		}
	}

	async function checkForUpdates() {
		try {
			const token = uni.getStorageSync('token') || uni.getStorageSync('accessToken')
			const appInfo = getAppInfo()

			const response = await uni.request({
				url: '/api/app/check-update',
				method: 'POST',
				header: {
					'Authorization': token ? `Bearer ${token}` : '',
					'Content-Type': 'application/json'
				},
				data: {
					appId: appInfo.appId,
					version: appInfo.version,
					versionCode: appInfo.versionCode,
					platform: appInfo.platform
				}
			})

			if (response.statusCode === 200) {
				return response.data
			} else {
				throw new Error(`HTTP ${response.statusCode}`)
			}
		} catch (error) {
			throw error
		}
	}

	function showUpdateDialog(updateInfo) {
		const content = `发现新版本\n\n版本: ${updateInfo.version}\n大小: ${updateInfo.size}\n\n更新说明:\n${updateInfo.releaseNotes}`

		uni.showModal({
			title: '发现新版本',
			content: content,
			confirmText: updateInfo.isForced ? '立即更新' : '立即更新',
			cancelText: updateInfo.isForced ? '' : '取消',
			showCancel: !updateInfo.isForced,
			success: (res) => {
				if (res.confirm) {
					startUpdate(updateInfo)
				} else if (updateInfo.isForced) {
					startUpdate(updateInfo)
				}
			}
		})
	}

	function startUpdate(updateInfo) {
		// #ifdef APP-PLUS
		startAppUpdate(updateInfo)
		// #endif

		// #ifdef H5
		startH5Update(updateInfo)
		// #endif

		// #ifdef MP-WEIXIN
		startWechatUpdate(updateInfo)
		// #endif
	}

	async function startAppUpdate(updateInfo) {
		try {
			isUpdating.value = true
			updateProgress.value = 0

			uni.showModal({
				title: '下载中',
				content: `进度: 0%`,
				showCancel: false,
				confirmText: '后台运行'
			})

			const downloadTask = uni.downloadFile({
				url: updateInfo.downloadUrl,
				success: (res) => {
					if (res.statusCode === 200) {
						installUpdate(res.tempFilePath, updateInfo)
					} else {
						throw new Error(`Download failed: ${res.statusCode}`)
					}
				},
				fail: (error) => {
					handleUpdateError(error)
				}
			})

			downloadTask.onProgressUpdate((res) => {
				updateProgress.value = res.progress
			})
		} catch (error) {
			handleUpdateError(error)
		}
	}

	function installUpdate(filePath, updateInfo) {
		// #ifdef APP-PLUS
		plus.runtime.install(filePath, { force: false }, () => {
			isUpdating.value = false
			showToast({ title: '安装成功', icon: 'success' })

			setTimeout(() => {
				plus.runtime.restart()
			}, 1500)
		}, (error) => {
			handleUpdateError(error)
		})
		// #endif
	}

	function startH5Update(updateInfo) {
		showToast({ title: 'H5版本请刷新页面', icon: 'none', duration: 3000 })
		setTimeout(() => {
			location.reload()
		}, 3000)
	}

	function startWechatUpdate(updateInfo) {
		// #ifdef MP-WEIXIN
		const updateManager = uni.getUpdateManager()

		updateManager.onCheckForUpdate(() => {})

		updateManager.onUpdateReady(() => {
			uni.showModal({
				title: '发现新版本',
				content: '微信小程序更新已就绪',
				success: (res) => {
					if (res.confirm) {
						updateManager.applyUpdate()
					}
				}
			})
		})

		updateManager.onUpdateFailed(() => {
			showToast({ title: '微信小程序更新失败', icon: 'none' })
		})
		// #endif
	}

	function handleUpdateError(error) {
		isUpdating.value = false
		updateProgress.value = 0

		uni.showModal({
			title: '更新失败',
			content: '更新错误' + ': ' + (error.message || error),
			showCancel: false,
			confirmText: '确定'
		})
	}

	return {
		username,
		studentId,
		isBound,
		userTags,
		userDetail,
		appVersion,
		cacheSize,
		hasUpdate,
		updateInfo,
		updateProgress,
		isUpdating,
		syncSettings,
		syncStatus,
		loadUserInfo,
		loadUserInfoFromStorage,
		updateUserInfoDisplayMixin,
		setDefaultUserInfo,
		getUserInitialMixin,
		showUserDetail,
		navigateToBindAccount,
		showBindingInfo,
		confirmRebind,
		confirmUnbind,
		unbindSchoolAccount,
		getFacultyNameMixin,
		calculateCacheSizeMixin,
		getAppVersion,
		checkHasUpdate,
		clearCache,
		checkUpdate,
		showAbout,
		logout,
		goBack,
		loadSyncSettings,
		getSyncStatusTextMixin,
		showSyncSettings,
		loadUserSettings,
		saveUserSettings,
		checkForUpdates,
		showUpdateDialog,
		startUpdate,
		startAppUpdate,
		installUpdate,
		startH5Update,
		startWechatUpdate,
		handleUpdateError
	}
}