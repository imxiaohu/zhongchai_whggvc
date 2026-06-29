/**
 * Sync Settings Composable
 * Sync settings form state and methods
 */

import { ref, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { showToast, navigateBack } from '@/pages/api/page.js'
import { getSyncSettings, getSyncStatus, getSyncLogs, updateSyncSettings, manualSync } from '@/api/sync.js'
import {
	SYNC_FREQUENCY_OPTIONS,
	TIME_RANGE_PRESETS,
	getFrequencyText,
	getFrequencyDescription,
	getAutoRetryDescription
} from '@/utils/syncSettings.js'

export function useSyncSettings() {
	const syncSettings = ref({
		enabled: false,
		frequency: 'daily',
		timeRange: '08:30-22:20',
		autoRetryEnabled: true,
		maxRetryCount: 3,
		personalInfoSyncEnabled: false,
		personalInfoCacheStatus: 'active'
	})

	const syncStatus = ref({
		syncStatus: 'idle',
		lastSyncAt: null,
		nextSyncAt: null,
		lastSyncMessage: '',
		coursesCount: 0
	})

	const syncLogs = ref([])
	const loading = ref(false)

	async function loadSyncSettings() {
		try {
			const token = uni.getStorageSync('token') || uni.getStorageSync('accessToken')
			if (!token) {
				showToast({ title: '请先登录', icon: 'none' })
				return
			}

			const [settingsRes, statusRes] = await Promise.allSettled([
				getSyncSettings(),
				getSyncStatus()
			])

			if (settingsRes.status === 'fulfilled') {
				syncSettings.value = settingsRes.value.result
			} else {
				console.error('获取同步设置失败:', settingsRes.reason)
			}

			if (statusRes.status === 'fulfilled') {
				syncStatus.value = statusRes.value.result
			} else {
				console.error('获取同步状态失败:', statusRes.reason)
			}
		} catch (error) {
			console.error('加载同步设置失败:', error)
			showToast({ title: '加载设置失败', icon: 'none' })
		}
	}

	async function loadSyncLogs() {
		try {
			const token = uni.getStorageSync('token') || uni.getStorageSync('accessToken')
			if (!token) return

			const result = await getSyncLogs(10)
			syncLogs.value = result.result || []
		} catch (error) {
			console.error('加载同步日志失败:', error)
		}
	}

	async function saveSyncSettings() {
		try {
			loading.value = true

			const result = await updateSyncSettings(syncSettings.value)
			syncSettings.value = result.result
			showToast({ title: '设置保存成功', icon: 'success' })
			await loadSyncSettings()
		} catch (error) {
			console.error('保存同步设置失败:', error)
			showToast({ title: error.message || '保存失败', icon: 'none' })
		} finally {
			loading.value = false
		}
	}

	async function manualSyncAction() {
		if (loading.value) return

		try {
			loading.value = true

			await manualSync()
			showToast({ title: '同步已开始', icon: 'success' })

			setTimeout(() => {
				loadSyncSettings()
				loadSyncLogs()
			}, 2000)
		} catch (error) {
			console.error('手动同步失败:', error)
			showToast({ title: error.message || '同步失败', icon: 'none' })
		} finally {
			loading.value = false
		}
	}

	function onSyncEnabledChange(e) {
		syncSettings.value.enabled = e.detail.value
		saveSyncSettings()
	}

	function onAutoRetryChange(e) {
		syncSettings.value.autoRetryEnabled = e.detail.value
		saveSyncSettings()
	}

	function onPersonalInfoSyncChange(e) {
		syncSettings.value.personalInfoSyncEnabled = e.detail.value
		saveSyncSettings()
		if (e.detail.value) {
			showToast({ title: '已开启个人信息缓存', icon: 'success' })
		} else {
			showToast({ title: '已关闭个人信息缓存', icon: 'success' })
		}
	}

	function showFrequencyPicker() {
		uni.showActionSheet({
			itemList: SYNC_FREQUENCY_OPTIONS.map(f => f.text),
			success: (res) => {
				syncSettings.value.frequency = SYNC_FREQUENCY_OPTIONS[res.tapIndex].value
				saveSyncSettings()
			}
		})
	}

	function showTimeRangePicker() {
		showPresetTimes()
	}

	function showPresetTimes() {
		uni.showActionSheet({
			itemList: TIME_RANGE_PRESETS.map(p => p.text),
			success: (res) => {
				const selectedPreset = TIME_RANGE_PRESETS[res.tapIndex]
				syncSettings.value.timeRange = selectedPreset.value
				saveSyncSettings()

				showToast({
					title: `已设置同步时间范围：${selectedPreset.text}`,
					icon: 'success'
				})
			}
		})
	}

	function getFrequencyTextValue() {
		return getFrequencyText(syncSettings.value.frequency)
	}

	function getFrequencyDescriptionValue() {
		return getFrequencyDescription(syncSettings.value.frequency)
	}

	function getAutoRetryDescriptionValue() {
		return getAutoRetryDescription(syncSettings.value.maxRetryCount)
	}

	function goBack() {
		navigateBack()
	}

	function bindLifecycle() {
		onMounted(() => {
			loadSyncSettings()
			loadSyncLogs()
		})

		onShow(() => {
			// B1 fix: onShow 不重复加载 settings，只刷新状态（避免与 onMounted 重复请求）
			// 日志数量少且变动频繁，适合每次进入页面刷新
			loadSyncLogs()
		})
	}

	return {
		syncSettings,
		syncStatus,
		syncLogs,
		loading,
		loadSyncSettings,
		loadSyncLogs,
		saveSyncSettings,
		manualSyncAction,
		onSyncEnabledChange,
		onAutoRetryChange,
		onPersonalInfoSyncChange,
		showFrequencyPicker,
		showTimeRangePicker,
		showPresetTimes,
		getFrequencyTextValue,
		getFrequencyDescriptionValue,
		getAutoRetryDescriptionValue,
		getAutoRetryDescription,
		goBack,
		bindLifecycle
	}
}