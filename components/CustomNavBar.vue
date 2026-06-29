<template>
  <view class="custom-nav-bar" :class="navBarClasses" :style="navBarStyle">
    <!-- 状态栏占位 - 使用CSS安全区域适配 -->
    <view class="safe-area-top" :style="{ height: statusBarHeight + 'px' }"></view>

    <!-- Navigation content -->
    <view class="nav-content" :style="navContentStyle">
      <!-- Left section -->
      <view class="nav-left" @tap="handleLeftClick">
        <view v-if="showBack" class="back-button">
          <text class="back-icon">←</text>
          <text v-if="leftText" class="left-text">{{ leftText }}</text>
        </view>
        <view v-else-if="leftIcon || leftText" class="custom-left">
          <text v-if="leftIcon" class="left-icon">{{ leftIcon }}</text>
          <text v-if="leftText" class="left-text">{{ leftText }}</text>
        </view>
      </view>

      <!-- Center section -->
      <view class="nav-center" @tap="handleCenterClick" :style="centerTitleStyle">
        <text class="center-title">{{ centerTitle }}</text>
      </view>

      <!-- Right section - 微信小程序中使用可拖动悬浮按钮 -->
      <!-- #ifdef MP-WEIXIN -->
      <view v-if="$slots.right" class="nav-right-fab-container">
        <movable-area class="movable-area">
          <movable-view
            class="nav-right-fab"
            direction="all"
            :inertia="true"
            :out-of-bounds="false"
            :damping="20"
            :friction="2"
            :x="fabPosition.x"
            :y="fabPosition.y"
            @change="onFabMove"
          >
            <view class="fab-content">
              <slot name="right"></slot>
            </view>
          </movable-view>
        </movable-area>
      </view>
      <!-- #endif -->

      <!-- #ifndef MP-WEIXIN -->
      <view class="nav-right">
        <slot name="right"></slot>
      </view>
      <!-- #endif -->
    </view>
  </view>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import systemUtils from '@/utils/system.js'

const props = defineProps({
  leftIcon: {
    type: String,
    default: ''
  },
  leftText: {
    type: String,
    default: ''
  },
  centerTitle: {
    type: String,
    required: true,
    default: ''
  },
  showBack: {
    type: Boolean,
    default: true
  },
  transparent: {
    type: Boolean,
    default: false
  },
  fixed: {
    type: Boolean,
    default: true
  },
  autoHide: {
    type: Boolean,
    default: false
  },
  backgroundColor: {
    type: String,
    default: ''
  },
  textColor: {
    type: String,
    default: '#2c3e50'
  }
})

const emit = defineEmits(['navHeightReady', 'heightReady', 'leftClick', 'centerClick', 'scroll'])

const statusBarHeight = ref(0)
const navBarHeight = ref(44)
const menuButtonInfo = ref(null)
const isVisible = ref(true)
const lastScrollTop = ref(0)
const scrollDirection = ref('up')
const fabPosition = ref({ x: 0, y: 0 })

// 内部状态（非响应式缓存）
let _systemInfo = null
let _windowResizeHandler = null
let _fabMoveTimer = null
let _scrollHandlerRef = null

const totalNavHeight = computed(() => statusBarHeight.value + navBarHeight.value)

const navBarClasses = computed(() => ({
  'nav-transparent': props.transparent && !props.backgroundColor,
  'nav-fixed': props.fixed,
  'nav-hidden': props.autoHide && !isVisible.value
}))

const navBarStyle = computed(() => {
  const style = {
    height: totalNavHeight.value + 'px',
    zIndex: props.fixed ? 1000 : 'auto'
  }

  if (props.backgroundColor) {
    style.backgroundColor = props.backgroundColor
  }

  return style
})

const navContentStyle = computed(() => ({
  height: navBarHeight.value + 'px',
  color: props.textColor
}))

const centerTitleStyle = computed(() => {
  const style = {}

  // #ifdef MP-WEIXIN
  if (menuButtonInfo.value) {
    const systemInfo = uni.getSystemInfoSync()
    const leftSpace = 120
    const rightSpace = systemInfo.windowWidth - menuButtonInfo.value.left + 20
    const maxSpace = Math.max(leftSpace, rightSpace)
    style.maxWidth = `calc(100% - ${maxSpace * 2}rpx)`
  }
  // #endif

  return style
})

function emitNavHeight() {
  const payload = {
    height: totalNavHeight.value,
    heightPx: totalNavHeight.value + 'px',
    heightRpx: (totalNavHeight.value * 2) + 'rpx'
  }
  emit('navHeightReady', payload)
  emit('heightReady', payload)
}

function getSystemInfo() {
  try {
    // #ifdef MP-WEIXIN
    const systemInfo = uni.getSystemInfoSync()
    _systemInfo = systemInfo

    if (systemInfo.safeArea && systemInfo.safeArea.top > 0) {
      statusBarHeight.value = systemInfo.safeArea.top
    } else {
      statusBarHeight.value = systemInfo.statusBarHeight || 20
    }

    getMenuButtonInfo()
    // #endif

    // #ifdef APP-PLUS
    if (systemUtils && typeof systemUtils.getStatusBarHeight === 'function') {
      statusBarHeight.value = systemUtils.getStatusBarHeight()
    } else {
      console.warn('systemUtils 未正确导入，使用默认状态栏高度')
      statusBarHeight.value = 44
    }
    navBarHeight.value = 44
    // #endif

    // #ifdef H5
    statusBarHeight.value = 0
    navBarHeight.value = 44
    // #endif
  } catch (error) {
    console.warn('Failed to get system info:', error)
    statusBarHeight.value = 20
    navBarHeight.value = 44
  }
}

function getMenuButtonInfo() {
  // #ifdef MP-WEIXIN
  try {
    let info
    try {
      info = uni.getMenuButtonBoundingClientRect()
    } catch (e) {
      console.warn('获取胶囊按钮信息失败，使用默认值:', e)
      const systemInfo = uni.getSystemInfoSync()
      info = {
        width: 87,
        height: 32,
        left: systemInfo.windowWidth - 7 - 87,
        right: systemInfo.windowWidth - 7,
        top: statusBarHeight.value + 4,
        bottom: statusBarHeight.value + 4 + 32
      }
    }

    menuButtonInfo.value = info

    const menuGap = info.top - statusBarHeight.value
    navBarHeight.value = Math.max(44, menuGap + info.height + menuGap)

  } catch (error) {
    console.warn('获取胶囊按钮信息失败:', error)
    navBarHeight.value = 44
    menuButtonInfo.value = null
  }
  // #endif
}

function handleLeftClick() {
  if (props.showBack) {
    goBack()
  }
  emit('leftClick')
}

function handleCenterClick() {
  emit('centerClick')
}

function goBack() {
  try {
    uni.navigateBack({
      delta: 1,
      fail: () => {
        uni.switchTab({
          url: '/pages/index/index',
          fail: () => {
            console.warn('Failed to navigate back or to home')
          }
        })
      }
    })
  } catch (error) {
    console.error('Navigation error:', error)
  }
}

function handleScroll(scrollTop) {
  if (!props.autoHide) return

  const threshold = 10
  const diff = scrollTop - lastScrollTop.value

  if (Math.abs(diff) > threshold) {
    if (diff > 0 && scrollTop > 100) {
      isVisible.value = false
      scrollDirection.value = 'down'
    } else if (diff < 0) {
      isVisible.value = true
      scrollDirection.value = 'up'
    }

    lastScrollTop.value = scrollTop
  }
}

function initFabPosition() {
  // #ifdef MP-WEIXIN
  try {
    if (!_systemInfo) {
      _systemInfo = uni.getSystemInfoSync()
    }

    const systemInfo = _systemInfo
    const fabSize = 64
    const margin = 30

    const rpxToPx = systemInfo.windowWidth / 750
    const fabSizePx = fabSize * rpxToPx
    const marginPx = margin * rpxToPx

    const safeAreaRight = menuButtonInfo.value
      ? menuButtonInfo.value.left - 10
      : systemInfo.windowWidth - marginPx
    const safeAreaBottom = systemInfo.windowHeight - 200 * rpxToPx

    fabPosition.value.x = Math.min(
      safeAreaRight - fabSizePx,
      systemInfo.windowWidth - fabSizePx - marginPx
    )
    fabPosition.value.y = Math.max(safeAreaBottom, totalNavHeight.value + 50)

  } catch (error) {
    console.warn('初始化悬浮按钮位置失败:', error)
    fabPosition.value.x = 250
    fabPosition.value.y = 500
  }
  // #endif
}

function onFabMove(e) {
  if (_fabMoveTimer) {
    clearTimeout(_fabMoveTimer)
  }

  _fabMoveTimer = setTimeout(() => {
    fabPosition.value.x = e.detail.x
    fabPosition.value.y = e.detail.y
    constrainFabPosition()
    saveFabPosition()
  }, 16)
}

function constrainFabPosition() {
  // #ifdef MP-WEIXIN
  try {
    if (!_systemInfo) {
      _systemInfo = uni.getSystemInfoSync()
    }

    const systemInfo = _systemInfo
    const fabSize = 64 * (systemInfo.windowWidth / 750)
    const margin = 10

    fabPosition.value.x = Math.max(
      margin,
      Math.min(fabPosition.value.x, systemInfo.windowWidth - fabSize - margin)
    )

    const minY = totalNavHeight.value + 10
    const maxY = systemInfo.windowHeight - fabSize - margin
    fabPosition.value.y = Math.max(minY, Math.min(fabPosition.value.y, maxY))

  } catch (error) {
    console.warn('约束FAB位置失败:', error)
  }
  // #endif
}

function saveFabPosition() {
  try {
    uni.setStorageSync('fabPosition', {
      x: fabPosition.value.x,
      y: fabPosition.value.y,
      timestamp: Date.now()
    })
  } catch (error) {
    console.warn('保存FAB位置失败:', error)
  }
}

function restoreFabPosition() {
  try {
    const savedPosition = uni.getStorageSync('fabPosition')
    if (savedPosition && savedPosition.timestamp) {
      const isExpired = (Date.now() - savedPosition.timestamp) > 7 * 24 * 60 * 60 * 1000

      if (!isExpired) {
        fabPosition.value.x = savedPosition.x
        fabPosition.value.y = savedPosition.y
        constrainFabPosition()
        return true
      }
    }
  } catch (error) {
    console.warn('恢复FAB位置失败:', error)
  }
  return false
}

function show() {
  isVisible.value = true
}

function hide() {
  isVisible.value = false
}

function toggle() {
  isVisible.value = !isVisible.value
}

onMounted(() => {
  getSystemInfo()

  nextTick(() => {
    if (!restoreFabPosition()) {
      initFabPosition()
    }
    emitNavHeight()
  })

  _windowResizeHandler = () => {
    getSystemInfo()
    nextTick(() => {
      emitNavHeight()
    })
  }
  uni.onWindowResize(_windowResizeHandler)
})

onBeforeUnmount(() => {
  if (_windowResizeHandler) {
    uni.offWindowResize(_windowResizeHandler)
    _windowResizeHandler = null
  }

  if (_fabMoveTimer) {
    clearTimeout(_fabMoveTimer)
    _fabMoveTimer = null
  }
})
</script>

<style scoped>
/* Import theme styles */

.custom-nav-bar {
  width: 100%;
  background-color: var(--bg-card);
  border-bottom: 1px solid var(--border-light);
  box-shadow: var(--shadow-xs);
  transition: transform 0.2s var(--ease-out), opacity 0.2s var(--ease-out);
}

.nav-transparent {
  background-color: var(--bg-overlay);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}

.nav-fixed {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
}

.nav-hidden {
  transform: translateY(-100%);
  opacity: 0;
}

.safe-area-top {
  width: 100%;
  background-color: inherit;
}

.nav-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--spacing-md);
  position: relative;
}

.nav-left {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  min-width: 60px;
}

.back-button {
  display: flex;
  align-items: center;
  padding: var(--spacing-xs);
  border-radius: var(--radius-sm);
  transition: background-color 0.15s ease;
}

.back-button:active {
  background-color: var(--primary-soft);
}

.back-icon {
  font-size: 18px;
  font-weight: 600;
  margin-right: 4px;
  color: var(--primary-600);
}

.custom-left {
  display: flex;
  align-items: center;
}

.left-icon {
  font-size: 16px;
  margin-right: 4px;
  color: var(--primary-600);
}

.left-text {
  font-size: 14px;
  color: var(--primary-600);
  font-weight: 500;
}

.nav-center {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  max-width: calc(100% - 120px);
}

.center-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.nav-right {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  min-width: 60px;
}

/* 微信小程序可拖动悬浮按钮 */
/* #ifdef MP-WEIXIN */
.nav-right-fab-container {
  position: fixed;
  top: 0;
  right: 0;
  width: 100vw;
  height: 100vh;
  z-index: 1001;
  pointer-events: none;
}

.movable-area {
  width: 100%;
  height: 100%;
  position: relative;
}

.nav-right-fab {
  position: absolute;
  bottom: 50px;
  right: 15px;
  width: 32px;
  height: 32px;
  pointer-events: auto;
  z-index: 1002;
}

.fab-content {
  width: 100%;
  height: 100%;
  background: var(--bg-card);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-md);
  border: 1px solid var(--border-light);
  transition: all 0.2s var(--ease-out);
}

.fab-content:active {
  transform: scale(0.92);
  box-shadow: var(--shadow-sm);
}
/* #endif */

/* Platform-specific adjustments */
/* #ifdef MP-WEIXIN */
.custom-nav-bar {
  box-shadow: var(--shadow-xs);
}

.nav-right {
  max-width: 100px;
}
/* #endif */

/* #ifdef APP-PLUS */
.nav-transparent {
  background-color: var(--bg-overlay);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}
/* #endif */

/* Active states */
.back-button:active {
  background-color: var(--primary-soft);
  transform: scale(0.97);
}

.nav-center:active {
  opacity: 0.7;
}

/* Responsive design */
@media screen and (max-width: 375px) {
  .center-title {
    font-size: 16px;
  }

  .back-icon {
    font-size: 16px;
  }

  .left-text {
    font-size: 13px;
  }
}

/* Animation for auto-hide */
@keyframes slideUp {
  from {
    transform: translateY(0);
    opacity: 1;
  }
  to {
    transform: translateY(-100%);
    opacity: 0;
  }
}

@keyframes slideDown {
  from {
    transform: translateY(-100%);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

/* Accessibility improvements */
.back-button,
.nav-center {
  min-height: 22px;
  min-width: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Loading state */
.nav-loading .center-title {
  opacity: 0.6;
}

.nav-loading .center-title::after {
  content: '...';
  animation: loading-dots 1.5s infinite;
}

@keyframes loading-dots {
  0%, 20% { opacity: 0; }
  50% { opacity: 1; }
  80%, 100% { opacity: 0; }
}
</style>
