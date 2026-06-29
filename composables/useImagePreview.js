/**
 * 图片预览 Composable
 * 从 mixins/imagePreviewMixin.js 迁移而来
 * 支持 H5 / 微信小程序 / 其他小程序平台
 */

import { ref } from 'vue'

export function useImagePreview() {
  const showImagePreview = ref(false)
  const previewImageUrl = ref('')
  const contentImages = ref([])
  const imageScale = ref(1)
  const imageTranslateX = ref(0)
  const imageTranslateY = ref(0)
  const isDragging = ref(false)
  const lastTouchX = ref(0)
  const lastTouchY = ref(0)
  const lastTouchDistance = ref(0)
  const lastTapTime = ref(0)

  function previewImage(imageUrl) {
    previewImageUrl.value = imageUrl
    showImagePreview.value = true
    resetImageTransform()

    if (typeof document !== 'undefined') {
      document.body.style.overflow = 'hidden'
    }

    setTimeout(() => {
      fitImageToScreen(imageUrl)
    }, 100)
  }

  function closeImagePreview() {
    showImagePreview.value = false
    previewImageUrl.value = ''
    resetImageTransform()

    if (typeof document !== 'undefined') {
      document.body.style.overflow = 'auto'
    }
  }

  function resetImageTransform() {
    imageScale.value = 1
    imageTranslateX.value = 0
    imageTranslateY.value = 0
    isDragging.value = false
  }

  function resetZoom() {
    resetImageTransform()
  }

  function zoomIn() {
    imageScale.value = Math.min(imageScale.value * 1.5, 50)
  }

  function zoomOut() {
    imageScale.value = Math.max(imageScale.value / 1.5, 0.5)
    constrainTranslation()
  }

  function zoomToPoint(scale, event) {
    const rect = event.currentTarget.getBoundingClientRect()
    const centerX = rect.width / 2
    const centerY = rect.height / 2
    const clickX = event.clientX - rect.left
    const clickY = event.clientY - rect.top

    const deltaX = (clickX - centerX) * (scale - imageScale.value)
    const deltaY = (clickY - centerY) * (scale - imageScale.value)

    imageScale.value = scale
    imageTranslateX.value -= deltaX / imageScale.value
    imageTranslateY.value -= deltaY / imageScale.value
    constrainTranslation()
  }

  function constrainTranslation() {
    if (imageScale.value <= 1) {
      imageTranslateX.value = 0
      imageTranslateY.value = 0
      return
    }

    const maxTranslate = 100 * (imageScale.value - 1)
    imageTranslateX.value = Math.max(-maxTranslate, Math.min(maxTranslate, imageTranslateX.value))
    imageTranslateY.value = Math.max(-maxTranslate, Math.min(maxTranslate, imageTranslateY.value))
  }

  function handleTouchStart(event) {
    event.preventDefault()
    const touches = event.touches

    if (touches.length === 1) {
      isDragging.value = true
      lastTouchX.value = touches[0].clientX
      lastTouchY.value = touches[0].clientY
    } else if (touches.length === 2) {
      const distance = getTouchDistance(touches[0], touches[1])
      lastTouchDistance.value = distance
    }
  }

  function handleTouchMove(event) {
    event.preventDefault()
    const touches = event.touches

    if (touches.length === 1 && isDragging.value && imageScale.value > 1) {
      const deltaX = touches[0].clientX - lastTouchX.value
      const deltaY = touches[0].clientY - lastTouchY.value

      imageTranslateX.value += deltaX / imageScale.value
      imageTranslateY.value += deltaY / imageScale.value
      constrainTranslation()

      lastTouchX.value = touches[0].clientX
      lastTouchY.value = touches[0].clientY
    } else if (touches.length === 2) {
      const distance = getTouchDistance(touches[0], touches[1])
      const scale = distance / lastTouchDistance.value

      imageScale.value = Math.max(0.5, Math.min(50, imageScale.value * scale))
      constrainTranslation()
      lastTouchDistance.value = distance
    }
  }

  function handleTouchEnd() {
    isDragging.value = false
  }

  function handleMouseDown(event) {
    if (imageScale.value > 1) {
      event.preventDefault()
      isDragging.value = true
      lastTouchX.value = event.clientX
      lastTouchY.value = event.clientY
    }
  }

  function handleMouseMove(event) {
    if (isDragging.value && imageScale.value > 1) {
      event.preventDefault()
      const deltaX = event.clientX - lastTouchX.value
      const deltaY = event.clientY - lastTouchY.value

      imageTranslateX.value += deltaX / imageScale.value
      imageTranslateY.value += deltaY / imageScale.value
      constrainTranslation()

      lastTouchX.value = event.clientX
      lastTouchY.value = event.clientY
    }
  }

  function handleMouseUp() {
    isDragging.value = false
  }

  function handleWheel(event) {
    event.preventDefault()
    const delta = event.deltaY > 0 ? 0.9 : 1.1
    imageScale.value = Math.max(0.5, Math.min(50, imageScale.value * delta))
    constrainTranslation()
  }

  function handleImageTap(event) {
    event.stopPropagation()

    const now = Date.now()
    if (lastTapTime.value && (now - lastTapTime.value) < 300) {
      if (imageScale.value === 1) {
        zoomToPoint(2, event)
      } else {
        resetZoom()
      }
    }
    lastTapTime.value = now
  }

  function handleMpImageTap() {
    const now = Date.now()
    if (lastTapTime.value && (now - lastTapTime.value) < 300) {
      if (imageScale.value === 1) {
        imageScale.value = 2
      } else {
        resetZoom()
      }
    }
    lastTapTime.value = now
  }

  function handleCloseButtonTouch(event) {
    event.stopPropagation()
    closeImagePreview()
  }

  function getTouchDistance(touch1, touch2) {
    const dx = touch1.clientX - touch2.clientX
    const dy = touch1.clientY - touch2.clientY
    return Math.sqrt(dx * dx + dy * dy)
  }

  function fitImageToScreen(imageUrl) {
    // #ifdef H5
    const img = new Image()
    img.onload = () => {
      try {
        const container = document.querySelector('.image-preview-container')
        if (!container) return

        const containerRect = container.getBoundingClientRect()
        const containerWidth = containerRect.width - 20
        const containerHeight = containerRect.height - 80

        const imageWidth = img.naturalWidth
        const imageHeight = img.naturalHeight

        const scaleX = containerWidth / imageWidth
        const scaleY = containerHeight / imageHeight
        const isLongImage = imageHeight > imageWidth * 1.5

        let minScale
        if (isLongImage) {
          minScale = Math.max(scaleX * 1.5, 3.0)
        } else {
          minScale = Math.max(scaleX, 2.2)
        }

        imageScale.value = minScale
        imageTranslateX.value = 0
        imageTranslateY.value = 0
      } catch (error) {
        imageScale.value = 2.2
      }
    }
    img.onerror = () => {}
    img.src = imageUrl
    // #endif

    // #ifdef MP-WEIXIN
    try {
      const systemInfo = uni.getSystemInfoSync()
      const screenWidth = systemInfo.screenWidth
      const screenHeight = systemInfo.screenHeight

      const containerWidth = screenWidth - 20
      const containerHeight = screenHeight - 150

      uni.getImageInfo({
        src: imageUrl,
        success: (res) => {
          const imageWidth = res.width
          const imageHeight = res.height

          const scaleX = containerWidth / imageWidth
          const scaleY = containerHeight / imageHeight
          const isLongImage = imageHeight > imageWidth * 1.5

          let minScale
          if (isLongImage) {
            minScale = Math.max(scaleX * 1.5, 3.0)
          } else {
            minScale = Math.max(scaleX, 2.2)
          }

          imageScale.value = minScale
          imageTranslateX.value = 0
          imageTranslateY.value = 0
        },
        fail: () => {
          imageScale.value = 2.2
        }
      })
    } catch (error) {
      imageScale.value = 2.2
    }
    // #endif

    // #ifdef MP-ALIPAY || MP-BAIDU || MP-TOUTIAO || MP-QQ
    try {
      uni.getImageInfo({
        src: imageUrl,
        success: (res) => {
          const imageWidth = res.width
          const imageHeight = res.height
          const isLongImage = imageHeight > imageWidth * 1.5

          if (isLongImage) {
            imageScale.value = 3.0
          } else {
            imageScale.value = 2.2
          }
          imageTranslateX.value = 0
          imageTranslateY.value = 0
        },
        fail: () => {
          imageScale.value = 2.2
        }
      })
    } catch (error) {
      imageScale.value = 2.2
    }
    // #endif
  }

  function saveImageToAlbum() {
    // #ifdef MP-WEIXIN
    if (!previewImageUrl.value) {
      uni.showToast({ title: '图片地址无效', icon: 'none' })
      return
    }

    uni.showActionSheet({
      itemList: ['保存到相册', '分享给好友'],
      success: (res) => {
        if (res.tapIndex === 0) {
          doSaveToAlbum()
        } else if (res.tapIndex === 1) {
          shareToFriend()
        }
      }
    })
    // #endif

    // #ifndef MP-WEIXIN
    uni.showToast({ title: '该功能仅在微信小程序中支持', icon: 'none' })
    // #endif
  }

  function doSaveToAlbum() {
    uni.getSetting({
      success: (res) => {
        if (res.authSetting['scope.writePhotosAlbum']) {
          doSaveImage()
        } else if (res.authSetting['scope.writePhotosAlbum'] === false) {
          uni.showModal({
            title: '权限申请',
            content: '需要访问您的相册来保存图片，请在设置中开启相册权限',
            confirmText: '去设置',
            success: (modalRes) => {
              if (modalRes.confirm) {
                uni.openSetting({
                  success: (settingRes) => {
                    if (settingRes.authSetting['scope.writePhotosAlbum']) {
                      doSaveImage()
                    }
                  }
                })
              }
            }
          })
        } else {
          uni.authorize({
            scope: 'scope.writePhotosAlbum',
            success: () => {
              doSaveImage()
            },
            fail: () => {
              uni.showModal({
                title: '权限申请',
                content: '需要访问您的相册来保存图片，请允许相册权限',
                showCancel: false
              })
            }
          })
        }
      },
      fail: () => {
        uni.showToast({ title: '获取权限信息失败', icon: 'none' })
      }
    })
  }

  function shareToFriend() {
    uni.showLoading({ title: '准备分享...' })

    if (previewImageUrl.value.startsWith('http')) {
      uni.downloadFile({
        url: previewImageUrl.value,
        success: (res) => {
          if (res.statusCode === 200) {
            uni.hideLoading()
            wx.shareImageMessage({
              imagePath: res.tempFilePath,
              success: () => {
                uni.showToast({ title: '分享成功', icon: 'success' })
              },
              fail: () => {
                uni.showToast({ title: '分享失败', icon: 'none' })
              }
            })
          } else {
            uni.hideLoading()
            uni.showToast({ title: '下载图片失败', icon: 'none' })
          }
        },
        fail: () => {
          uni.hideLoading()
          uni.showToast({ title: '下载图片失败', icon: 'none' })
        }
      })
    } else {
      uni.hideLoading()
      wx.shareImageMessage({
        imagePath: previewImageUrl.value,
        success: () => {
          uni.showToast({ title: '分享成功', icon: 'success' })
        },
        fail: () => {
          uni.showToast({ title: '分享失败', icon: 'none' })
        }
      })
    }
  }

  function doSaveImage() {
    uni.showLoading({ title: '保存中...' })

    if (previewImageUrl.value.startsWith('http')) {
      uni.downloadFile({
        url: previewImageUrl.value,
        success: (res) => {
          if (res.statusCode === 200) {
            uni.saveImageToPhotosAlbum({
              filePath: res.tempFilePath,
              success: () => {
                uni.hideLoading()
                uni.showToast({ title: '保存成功', icon: 'success' })
              },
              fail: () => {
                uni.hideLoading()
                uni.showToast({ title: '保存失败', icon: 'none' })
              }
            })
          } else {
            uni.hideLoading()
            uni.showToast({ title: '下载图片失败', icon: 'none' })
          }
        },
        fail: () => {
          uni.hideLoading()
          uni.showToast({ title: '下载图片失败', icon: 'none' })
        }
      })
    } else {
      uni.saveImageToPhotosAlbum({
        filePath: previewImageUrl.value,
        success: () => {
          uni.hideLoading()
          uni.showToast({ title: '保存成功', icon: 'success' })
        },
        fail: () => {
          uni.hideLoading()
          uni.showToast({ title: '保存失败', icon: 'none' })
        }
      })
    }
  }

  return {
    showImagePreview,
    previewImageUrl,
    contentImages,
    imageScale,
    imageTranslateX,
    imageTranslateY,
    isDragging,
    previewImage,
    closeImagePreview,
    resetImageTransform,
    resetZoom,
    zoomIn,
    zoomOut,
    zoomToPoint,
    constrainTranslation,
    handleTouchStart,
    handleTouchMove,
    handleTouchEnd,
    handleMouseDown,
    handleMouseMove,
    handleMouseUp,
    handleWheel,
    handleImageTap,
    handleMpImageTap,
    handleCloseButtonTouch,
    fitImageToScreen,
    saveImageToAlbum,
    doSaveToAlbum,
    shareToFriend,
    doSaveImage
  }
}
