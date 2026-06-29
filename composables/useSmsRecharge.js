/**
 * SMS Recharge Composable
 * Recharge form state, payment processing methods, and order management
 */

import { onUnmounted, ref } from 'vue'
import { showToast, showLoading, hideLoading, navigateBack } from '@/pages/api/page.js'
import {
	PAYMENT_TIMEOUT,
	QR_CODE_CONFIG,
	createDefaultBalance,
	updateBalanceFromResponse
} from '@/utils/smsPricing.js'

function getAuthToken() {
	return uni.getStorageSync('token')
}

async function authRequest(options) {
	const token = getAuthToken()
	if (!token) {
		showToast({ title: '请先登录' })
		throw new Error('未登录')
	}

	const defaultOptions = {
		header: {
			'Authorization': `Bearer ${token}`,
			'Content-Type': 'application/json'
		}
	}

	return uni.request({
		...defaultOptions,
		...options
	})
}

export function useSmsRecharge() {
	const smsBalance = ref(createDefaultBalance())
	const rechargePackages = ref([])
	const selectedPackage = ref(null)

	const paying = ref(false)
	const showPayModal = ref(false)
	const paymentStatus = ref(null)
	const currentOrderId = ref(null)

	const qrCodeSize = ref(QR_CODE_CONFIG.defaultSize)
	const qrCodeDataURL = ref(null)
	const qrCodeLoading = ref(false)
	const paymentCheckInterval = ref(null)
	const currentPaymentUrl = ref(null)

	async function loadSMSBalance() {
		try {
			const response = await authRequest({
				url: '/api/sms/balance',
				method: 'GET'
			})

			if (response.statusCode === 200 && response.data.success) {
				smsBalance.value = updateBalanceFromResponse(response.data.result)
			}
		} catch (error) {
			console.error('加载短信余额出错:', error)
		}
	}

	async function loadRechargePackages() {
		try {
			const response = await authRequest({
				url: '/api/sms/recharge-packages',
				method: 'GET'
			})

			if (response.statusCode === 200 && response.data.success) {
				rechargePackages.value = response.data.result || []
				const recommendedPackage = rechargePackages.value.find(pkg => pkg.id === 3)
				if (recommendedPackage) {
					selectedPackage.value = recommendedPackage
				}
			}
		} catch (error) {
			console.error('加载充值套餐出错:', error)
		}
	}

	function selectPackage(pkg) {
		selectedPackage.value = pkg
	}

	async function createOrder() {
		if (!selectedPackage.value || paying.value) return

		try {
			paying.value = true
			showLoading({ title: '创建订单中...' })

			const response = await authRequest({
				url: '/api/sms/recharge',
				method: 'POST',
				data: {
					packageId: selectedPackage.value.id
				}
			})

			if (response.statusCode === 200 && response.data.success) {
				const result = response.data.result
				currentOrderId.value = result.orderId
				currentPaymentUrl.value = result.codeUrl

				showPayModal.value = true
				generateQRCode(result.codeUrl)

				startPaymentCheck()
			} else {
				const errorMsg = (response.data && response.data.message) ? response.data.message : '创建订单失败'
				showToast({ title: errorMsg })
			}
		} catch (error) {
			console.error('创建订单出错:', error)
			showToast({ title: '网络错误，请稍后重试' })
		} finally {
			paying.value = false
			hideLoading()
		}
	}

	function generateQRCode(codeURL) {
		qrCodeLoading.value = true
		qrCodeDataURL.value = null

		// #ifdef H5
		generateQRCodeH5(codeURL)
		// #endif

		// #ifdef MP-WEIXIN
		generateQRCodeMP(codeURL)
		// #endif
	}

	function generateQRCodeH5(codeURL) {
		console.log('H5环境生成二维码:', codeURL)

		import('qrcode').then(QRCode => {
			const query = uni.createSelectorQuery()
			query.select('#qrcode').node().exec((res) => {
				if (res && res[0] && res[0].node) {
					const canvas = res[0].node

					canvas.width = qrCodeSize.value
					canvas.height = qrCodeSize.value

					QRCode.toCanvas(canvas, codeURL, {
						width: qrCodeSize.value,
						height: qrCodeSize.value,
						margin: QR_CODE_CONFIG.margin,
						color: QR_CODE_CONFIG.colors
					}, (error) => {
						if (error) {
							console.error('生成二维码失败:', error)
							showToast({ title: '生成二维码失败' })
						} else {
							console.log('二维码生成成功')
						}
					})
				} else {
					QRCode.toDataURL(codeURL, {
						width: qrCodeSize.value,
						height: qrCodeSize.value,
						margin: QR_CODE_CONFIG.margin,
						color: QR_CODE_CONFIG.colors
					}, (error, url) => {
						if (error) {
							console.error('生成二维码失败:', error)
							showToast({ title: '生成二维码失败' })
						} else {
							console.log('二维码生成成功(DataURL)')
							qrCodeDataURL.value = url
						}
					})
				}
			})
		}).catch(error => {
			console.error('加载QRCode库失败:', error)
			showToast({ title: '加载二维码组件失败' })
		})
	}

	function generateQRCodeMP(codeURL) {
		console.log('微信小程序环境生成二维码:', codeURL)
		generateQRCodeDataURL(codeURL)
	}

	function generateQRCodeDataURL(codeURL) {
		try {
			const qrApiUrl = `https://api.qrserver.com/v1/create-qr-code/?size=${qrCodeSize.value}x${qrCodeSize.value}&data=${encodeURIComponent(codeURL)}`
			qrCodeDataURL.value = qrApiUrl
			qrCodeLoading.value = false
			console.log('二维码生成成功(在线服务)')
		} catch (error) {
			console.error('生成二维码失败:', error)
			qrCodeLoading.value = false
			showToast({ title: '生成二维码失败' })
		}
	}

	function copyPaymentLink() {
		if (!currentPaymentUrl.value) {
			showToast({ title: '支付链接不存在' })
			return
		}

		// #ifdef MP-WEIXIN
		uni.setClipboardData({
			data: currentPaymentUrl.value,
			success: () => {
				showToast({
					title: '链接已复制到剪贴板',
					icon: 'success'
				})
			},
			fail: () => {
				showToast({ title: '复制失败' })
			}
		})
		// #endif
	}

	function saveQRCode() {
		if (!qrCodeDataURL.value) {
			showToast({ title: '二维码不存在' })
			return
		}

		// #ifdef MP-WEIXIN
		uni.downloadFile({
			url: qrCodeDataURL.value,
			success: (res) => {
				if (res.statusCode === 200) {
					uni.saveImageToPhotosAlbum({
						filePath: res.tempFilePath,
						success: () => {
							showToast({
								title: '二维码已保存到相册',
								icon: 'success'
							})
						},
						fail: (error) => {
							console.error('保存图片失败:', error)
							if (error.errMsg.includes('auth')) {
								uni.showModal({
									title: '需要授权',
									content: '需要授权访问相册才能保存二维码',
									confirmText: '去设置',
									success: (modalRes) => {
										if (modalRes.confirm) {
											uni.openSetting()
										}
									}
								})
							} else {
								showToast({ title: '保存失败' })
							}
						}
					})
				} else {
					showToast({ title: '下载二维码失败' })
				}
			},
			fail: () => {
				showToast({ title: '下载二维码失败' })
			}
		})
		// #endif
	}

	function startPaymentCheck() {
		if (paymentCheckInterval.value) {
			clearInterval(paymentCheckInterval.value)
		}

		paymentCheckInterval.value = setInterval(async () => {
			await checkPaymentStatus()
		}, PAYMENT_TIMEOUT.checkInterval)

		setTimeout(() => {
			if (paymentCheckInterval.value) {
				clearInterval(paymentCheckInterval.value)
				paymentCheckInterval.value = null
			}
		}, PAYMENT_TIMEOUT.maxDuration)
	}

	async function checkPaymentStatus() {
		if (!currentOrderId.value) return

		try {
			const response = await authRequest({
				url: `/api/payment/status/${currentOrderId.value}`,
				method: 'GET'
			})

			if (response.statusCode === 200 && response.data.success) {
				const status = response.data.result.status

				if (status === 'SUCCESS') {
					paymentStatus.value = 'success'
					clearInterval(paymentCheckInterval.value)
					paymentCheckInterval.value = null

					await loadSMSBalance()

					setTimeout(() => {
						closePayModal()
						showToast({
							title: '充值成功！',
							icon: 'success'
						})
						setTimeout(() => {
							navigateBack()
						}, 1500)
					}, PAYMENT_TIMEOUT.successDelay)
				} else if (status === 'FAILED') {
					paymentStatus.value = 'error'
					clearInterval(paymentCheckInterval.value)
					paymentCheckInterval.value = null
				}
			}
		} catch (error) {
			console.error('检查支付状态出错:', error)
		}
	}

	function closePayModal() {
		showPayModal.value = false
		paymentStatus.value = null
		currentOrderId.value = null
		qrCodeDataURL.value = null
		qrCodeLoading.value = false
		currentPaymentUrl.value = null

		if (paymentCheckInterval.value) {
			clearInterval(paymentCheckInterval.value)
			paymentCheckInterval.value = null
		}
	}

	function cleanupOfficialTheme() {
		// 预留扩展点
	}

	async function loadData() {
		await Promise.all([
			loadSMSBalance(),
			loadRechargePackages()
		])
	}

	function bindLifecycle() {
		onUnmounted(() => {
			if (paymentCheckInterval.value) {
				clearInterval(paymentCheckInterval.value)
				paymentCheckInterval.value = null
			}
		})
	}

	return {
		smsBalance,
		rechargePackages,
		selectedPackage,
		paying,
		showPayModal,
		paymentStatus,
		currentOrderId,
		qrCodeSize,
		qrCodeDataURL,
		qrCodeLoading,
		paymentCheckInterval,
		currentPaymentUrl,
		loadSMSBalance,
		loadRechargePackages,
		selectPackage,
		createOrder,
		generateQRCode,
		generateQRCodeH5,
		generateQRCodeMP,
		generateQRCodeDataURL,
		copyPaymentLink,
		saveQRCode,
		startPaymentCheck,
		checkPaymentStatus,
		closePayModal,
		cleanupOfficialTheme,
		loadData,
		bindLifecycle
	}
}