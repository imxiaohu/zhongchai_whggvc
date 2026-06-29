/**
 * 用户认证状态管理 (Pinia Store)
 * 统一管理用户登录状态、Token 和用户信息
 */

import { defineStore } from 'pinia'
import { getStorage } from '@/utils/storage.js'

export const useUserStore = defineStore('user', {
  state: () => ({
    // 认证状态
    token: null,
    userInfo: null,
    loginType: '',
    isLoggedIn: false,
    
    // 初始化状态
    initialized: false
  }),

  getters: {
    // 是否已认证
    isAuthenticated: (state) => !!state.token,
    
    // 是否有用户信息
    hasUserInfo: (state) => !!state.userInfo,
    
    // 获取用户昵称
    nickname: (state) => {
      if (!state.userInfo) return '同学'
      return state.userInfo.nickname || 
             state.userInfo.nickName || 
             state.userInfo.realname || 
             state.userInfo.username || 
             '同学'
    },
    
    // 获取用户头像文字
    avatarText: (state) => {
      const name = state.userInfo?.nickname || 
                   state.userInfo?.nickName || 
                   state.userInfo?.realname || 
                   state.userInfo?.username || 
                   '同'
      return String(name).slice(-1) || '同'
    }
  },

  actions: {
    /**
     * 从本地存储初始化状态
     */
    initFromStorage() {
      console.log('[UserStore] 从本地存储初始化')
      
      try {
        const token = getStorage('token')
        const userInfoStr = getStorage('userInfo')
        const loginType = getStorage('loginType')
        
        this.token = token || null
        this.loginType = loginType || ''
        this.isLoggedIn = !!token
        
        if (userInfoStr) {
          try {
            this.userInfo = JSON.parse(userInfoStr)
          } catch (e) {
            console.error('[UserStore] 解析用户信息失败:', e)
            this.userInfo = null
          }
        }
        
        this.initialized = true
        console.log('[UserStore] 初始化完成:', {
          hasToken: !!this.token,
          hasUserInfo: !!this.userInfo,
          loginType: this.loginType
        })
      } catch (error) {
        console.error('[UserStore] 初始化失败:', error)
        this.initialized = true
      }
    },

    /**
     * 设置 Token
     * @param {string} token - 用户 Token
     */
    setToken(token) {
      this.token = token
      this.isLoggedIn = !!token
      if (token) {
        uni.setStorageSync('token', token)
      } else {
        uni.removeStorageSync('token')
      }
    },

    /**
     * 设置用户信息
     * @param {Object} userInfo - 用户信息对象
     */
    setUserInfo(userInfo) {
      this.userInfo = userInfo
      if (userInfo) {
        uni.setStorageSync('userInfo', JSON.stringify(userInfo))
      } else {
        uni.removeStorageSync('userInfo')
      }
    },

    /**
     * 设置登录类型
     * @param {string} type - 登录类型 ('school' | 'wechat')
     */
    setLoginType(type) {
      this.loginType = type
      if (type) {
        uni.setStorageSync('loginType', type)
      } else {
        uni.removeStorageSync('loginType')
      }
    },

    /**
     * 登录成功
     * @param {Object} params - 登录参数 { token, userInfo, loginType }
     */
    login({ token, userInfo, loginType }) {
      console.log('[UserStore] 登录成功')
      
      this.setToken(token)
      if (userInfo) this.setUserInfo(userInfo)
      if (loginType) this.setLoginType(loginType)
      this.isLoggedIn = true
    },

    /**
     * 退出登录
     */
    logout() {
      console.log('[UserStore] 退出登录')
      
      // 保留记住密码的信息
      const savedUsername = uni.getStorageSync('saved_username')
      const savedPassword = uni.getStorageSync('saved_password')
      const rememberPassword = uni.getStorageSync('remember_password')
      const clientId = uni.getStorageSync('clientId')
      
      // 清除所有存储
      uni.clearStorageSync()
      
      // 恢复需要保留的信息
      if (rememberPassword) {
        uni.setStorageSync('saved_username', savedUsername)
        uni.setStorageSync('saved_password', savedPassword)
        uni.setStorageSync('remember_password', rememberPassword)
      }
      
      if (clientId) {
        uni.setStorageSync('clientId', clientId)
      }
      
      // 重置状态
      this.token = null
      this.userInfo = null
      this.loginType = ''
      this.isLoggedIn = false
    },

    /**
     * 更新用户信息字段
     * @param {Object} updates - 要更新的字段
     */
    updateUserInfo(updates) {
      if (this.userInfo) {
        this.userInfo = { ...this.userInfo, ...updates }
        uni.setStorageSync('userInfo', JSON.stringify(this.userInfo))
      }
    }
  }
})
