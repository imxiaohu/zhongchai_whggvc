/**
 * 学校账号绑定状态管理 (Pinia Store)
 * 统一管理学校账号绑定状态
 */

import { defineStore } from 'pinia'
import { getStorage, setStorage } from '@/utils/storage.js'

export const useSchoolAccountStore = defineStore('schoolAccount', {
  state: () => ({
    // 是否已绑定学校账号
    hasBindSchoolAccount: null,
    // 学校登录状态
    isSchoolLogin: false,
    // 客户端ID
    clientId: null,
    // 是否已初始化
    initialized: false
  }),

  getters: {
    // 是否已绑定学校账号
    isBound: (state) => state.hasBindSchoolAccount === true,
    
    // 是否需要绑定
    needsBinding: (state) => state.hasBindSchoolAccount === false,
    
    // 获取客户端ID
    currentClientId: (state) => state.clientId
  },

  actions: {
    /**
     * 初始化学校账号绑定状态
     */
    initialize() {
      console.log('[SchoolAccountStore] 初始化学校账号绑定状态')
      
      try {
        const hasBindSchoolAccount = getStorage('hasBindSchoolAccount')
        const userInfoStr = getStorage('userInfo')
        const token = getStorage('token')
        const clientId = getStorage('clientId')
        
        let userInfo = null
        if (userInfoStr) {
          try {
            userInfo = JSON.parse(userInfoStr)
          } catch (error) {
            console.error('[SchoolAccountStore] 解析用户信息失败:', error)
          }
        }
        
        const isSchoolLogin = (token && token.startsWith('logged_in_')) || !userInfo
        
        let finalBindStatus = false

        if (userInfo) {
          if (userInfo.hasSchoolAccount !== undefined) {
            finalBindStatus = userInfo.hasSchoolAccount
          } else {
            finalBindStatus = !!(
              userInfo.username &&
              userInfo.password &&
              !userInfo.username.startsWith('wx_') &&
              userInfo.username !== userInfo.wechatOpenID &&
              userInfo.realname
            )
          }
        } else if (hasBindSchoolAccount !== null) {
          finalBindStatus = hasBindSchoolAccount
        }

        if (isSchoolLogin) {
          finalBindStatus = true
        }
        
        this.hasBindSchoolAccount = finalBindStatus
        this.isSchoolLogin = isSchoolLogin
        this.clientId = clientId
        this.initialized = true
        
        console.log('[SchoolAccountStore] 初始化完成:', {
          hasBindSchoolAccount: finalBindStatus,
          isSchoolLogin: isSchoolLogin,
          hasUserInfo: !!userInfo,
          clientId: clientId
        })
      } catch (error) {
        console.error('[SchoolAccountStore] 初始化失败:', error)
        this.hasBindSchoolAccount = false
        this.isSchoolLogin = false
        this.clientId = null
        this.initialized = true
      }
    },

    /**
     * 检查学校账号绑定状态
     * @returns {boolean} 是否已绑定学校账号
     */
    checkBinding() {
      if (!this.initialized) {
        this.initialize()
      }
      console.log('[SchoolAccountStore] 检查绑定状态:', {
        hasBindSchoolAccount: this.hasBindSchoolAccount,
        isSchoolLogin: this.isSchoolLogin,
        hasUserInfo: !!this.userInfo
      })
      return this.hasBindSchoolAccount
    },

    /**
     * 更新绑定状态
     * @param {boolean} isBound - 是否已绑定
     * @param {Object} userInfo - 用户信息（可选）
     */
    updateBinding(isBound, userInfo = null) {
      console.log('[SchoolAccountStore] 更新绑定状态:', {
        isBound: isBound,
        hasUserInfo: !!userInfo
      })
      
      setStorage('hasBindSchoolAccount', isBound)
      this.hasBindSchoolAccount = isBound
      
      if (userInfo) {
        setStorage('userInfo', JSON.stringify(userInfo))
      }
    },

    /**
     * 设置学校登录状态
     * @param {boolean} isSchoolLogin - 是否为学校登录
     */
    setSchoolLogin(isSchoolLogin) {
      console.log('[SchoolAccountStore] 设置学校登录状态:', isSchoolLogin)
      this.isSchoolLogin = isSchoolLogin
      if (isSchoolLogin) {
        this.hasBindSchoolAccount = true
      }
    },

    /**
     * 设置客户端ID
     * @param {string} clientId - 客户端ID
     */
    setClientId(clientId) {
      console.log('[SchoolAccountStore] 设置客户端ID:', clientId)
      if (clientId) {
        setStorage('clientId', clientId)
      }
      this.clientId = clientId
    },

    /**
     * 清除所有状态
     */
    clear() {
      console.log('[SchoolAccountStore] 清除所有状态')
      this.hasBindSchoolAccount = false
      this.isSchoolLogin = false
      this.clientId = null
      this.initialized = false
    }
  }
})

// 保持向后兼容的导出
export default {
  getState: () => {
    const store = useSchoolAccountStore()
    return {
      hasBindSchoolAccount: store.hasBindSchoolAccount,
      userInfo: null,
      isSchoolLogin: store.isSchoolLogin,
      clientId: store.clientId,
      initialized: store.initialized
    }
  },
  
  checkSchoolAccountBinding: () => {
    const store = useSchoolAccountStore()
    return store.checkBinding()
  },
  
  updateSchoolAccountBinding: (isBound, userInfo) => {
    const store = useSchoolAccountStore()
    store.updateBinding(isBound, userInfo)
  },
  
  setSchoolLoginStatus: (isSchoolLogin) => {
    const store = useSchoolAccountStore()
    store.setSchoolLogin(isSchoolLogin)
  },
  
  setClientId: (clientId) => {
    const store = useSchoolAccountStore()
    store.setClientId(clientId)
  },
  
  initializeSchoolAccountState: () => {
    const store = useSchoolAccountStore()
    store.initialize()
  },
  
  clearState: () => {
    const store = useSchoolAccountStore()
    store.clear()
  }
}
