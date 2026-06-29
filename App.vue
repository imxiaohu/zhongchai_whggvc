<script setup>
import { ref } from 'vue'
import { onLaunch, onShow, onHide, onError, onUnload } from '@dcloudio/uni-app'
import './utils/request.js'
import unreadMessageManager from '@/utils/unreadMessageManager.js'

const refreshDebounceTimer = ref(null)

onLaunch(() => {
  console.log('App Launch')

  initUnreadMessageManager()
  checkAppUpdate()

  const token = uni.getStorageSync('token')
  const clientId = uni.getStorageSync('clientId')
  console.log('应用启动，检查登录状态:', {
    hasToken: !!token,
    hasClientId: !!clientId
  })

  const appUrl = uni.getStorageSync('app_url')
  console.log('存储的服务器地址:', appUrl)

  const systemInfo = uni.getSystemInfoSync()
  console.log('系统信息:', systemInfo.platform)

})

onShow(() => {
  console.log('App Show')
  debouncedRefreshUnreadMessages()
})

onHide(() => {
  console.log('App Hide')
})

onUnload(() => {
  if (refreshDebounceTimer.value) {
    clearTimeout(refreshDebounceTimer.value)
    refreshDebounceTimer.value = null
  }
})



function debouncedRefreshUnreadMessages() {
  if (refreshDebounceTimer.value) {
    clearTimeout(refreshDebounceTimer.value)
  }

  refreshDebounceTimer.value = setTimeout(() => {
    refreshUnreadMessages()
    refreshDebounceTimer.value = null
  }, 300)
}

async function initUnreadMessageManager() {
  console.log('App.vue 初始化未读消息管理器')
  await unreadMessageManager.init()
}

async function refreshUnreadMessages() {
  console.log('App.vue 刷新未读消息数量')
  const token = uni.getStorageSync('token')
  if (!token) {
    console.log('用户未登录，跳过未读消息刷新')
    return
  }

  await unreadMessageManager.fetchUnreadCount()
}

async function checkAppUpdate() {
  // #ifdef APP-PLUS
  console.log('检查应用更新...')

  const appInfo = {
    appId: plus.runtime.appid,
    version: plus.runtime.version,
    versionCode: plus.runtime.versionCode,
    platform: plus.os.name
  }

  console.log('当前应用信息:', appInfo)

  const response = await uni.request({
    url: '/api/app/check-update',
    method: 'POST',
    header: {
      'Content-Type': 'application/json'
    },
    data: appInfo
  })

  console.log('更新检查响应:', response)

  if (response.statusCode === 200 && response.data.success) {
    const { hasUpdate, updateInfo } = response.data.result

    if (hasUpdate && updateInfo) {
      console.log('发现新版本:', updateInfo)

      if (updateInfo.isForced) {
        showForceUpdateDialog(updateInfo)
      } else {
        setTimeout(() => {
          showOptionalUpdateDialog(updateInfo)
        }, 3000)
      }
    }
  }
  // #endif
}

function showForceUpdateDialog(updateInfo) {
  const content = `发现新版本 ${updateInfo.version}\n\n更新内容:\n${updateInfo.releaseNotes}`

  uni.showModal({
    title: '发现新版本',
    content: content,
    showCancel: false,
    confirmText: '立即更新',
    success: (res) => {
      if (res.confirm) {
        startAppUpdate(updateInfo)
      }
    }
  })
}

function showOptionalUpdateDialog(updateInfo) {
  const content = `发现新版本 ${updateInfo.version}\n\n更新内容:\n${updateInfo.releaseNotes}`

  uni.showModal({
    title: '发现新版本',
    content: content,
    confirmText: '立即更新',
    cancelText: '稍后提醒',
    success: (res) => {
      if (res.confirm) {
        startAppUpdate(updateInfo)
      }
    }
  })
}

async function startAppUpdate(updateInfo) {
  // #ifdef APP-PLUS
  console.log('开始下载更新:', updateInfo.downloadUrl)

  uni.showLoading({
    title: '正在下载更新...',
    mask: true
  })

  const downloadTask = uni.downloadFile({
    url: updateInfo.downloadUrl,
    success: (res) => {
      uni.hideLoading()
      if (res.statusCode === 200) {
        console.log('下载完成，开始安装:', res.tempFilePath)
        installUpdate(res.tempFilePath)
      } else {
        console.error('下载失败:', res.statusCode)
        uni.showToast({
          title: '下载失败',
          icon: 'none'
        })
      }
    },
    fail: (error) => {
      uni.hideLoading()
      console.error('下载失败:', error)
      uni.showToast({
        title: '下载失败',
        icon: 'none'
      })
    }
  })

  downloadTask.onProgressUpdate((res) => {
    console.log('下载进度:', res.progress + '%')
  })
  // #endif
}

function installUpdate(filePath) {
  // #ifdef APP-PLUS
  plus.runtime.install(filePath, {
    force: false
  }, () => {
    console.log('安装成功')
    uni.showToast({
      title: '安装成功',
      icon: 'success'
    })

    setTimeout(() => {
      plus.runtime.restart()
    }, 1500)
  }, (error) => {
    console.error('安装失败:', error)
    uni.showToast({
      title: '安装失败',
      icon: 'none'
    })
  })
  // #endif
}
</script>

<style lang="scss">
/* TDesign 图标样式 */
@import './static/css/tdesign-icons.css';

/* iconfont.ttf 在 iconfont.css 中引用，构建时无法解析，图标已迁移至 TDesign */
/* @import './static/fonts/iconfont.css'; */

/* 引入 icon 图标样式（cuIcon 字体） */
@import './static/css/icon.css';

/* 引入统一视觉规范体系 */
@import './static/css/design-system.css';

/* 引入社区页面增强样式 */
@import './static/css/community-enhanced.css';

/* 引入首页专用样式 */
@import './static/css/home-page.css';

/* 每个页面公共css */
/* 重置样式 */
page,
uni-page-body {
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  font-size: var(--font-size-md);
  color: var(--text-primary);
  background-color: var(--bg-secondary);
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Text', 'PingFang SC', 'Helvetica Neue', Helvetica, Arial, sans-serif;
}

/* 通用卡片样式 */
.card {
  background-color: var(--bg-card);
  border-radius: var(--radius-lg);
  padding: var(--spacing-md);
  box-shadow: var(--shadow-sm);
  margin-bottom: var(--spacing-md);
  border: 1px solid var(--border-light);
}

/* 标准按钮样式 */
button {
  background-color: var(--primary-600);
  color: #fff;
  border-radius: var(--radius-md);
  font-size: var(--font-size-md);
  font-weight: 600;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  transition: all 0.2s var(--ease-out);

  &:active {
    opacity: 0.85;
    transform: scale(0.98);
  }
}

button[disabled] {
  background-color: var(--bg-muted) !important;
  color: var(--text-tertiary) !important;
}

/* 输入框样式 */
input, textarea {
  background-color: var(--bg-muted);
  border-radius: var(--radius-md);
  padding: 12px 16px;
  font-size: var(--font-size-md);
  box-sizing: border-box;
  border: 1.5px solid transparent;

  &:focus {
    border-color: var(--primary-400);
    background-color: var(--bg-primary);
  }
}

/* 工具类 */
.flex-center {
  display: flex;
  align-items: center;
  justify-content: center;
}

.text-ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.safe-area-bottom {
  height: env(safe-area-inset-bottom);
  width: 100%;
}
</style>
