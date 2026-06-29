/**
 * 课程数据缓存状态管理 (Pinia Store)
 * 统一管理 getCurrentTime 和 getCourseLessonTime 的缓存
 */

import { defineStore } from 'pinia'
import { request } from '../utils/request.js'

export const useCourseCache = defineStore('courseCache', {
  state: () => ({
    // 当前时间信息缓存
    currentTime: {
      data: null,
      timestamp: null,
      expireTime: null
    },
    
    // 课程时间配置缓存
    lessonTime: {
      data: null,
      timestamp: null,
      expireTime: null
    }
  }),

  getters: {
    // 获取当前时间缓存状态
    currentTimeStatus: (state) => {
      const now = Date.now()
      return {
        cached: !!state.currentTime.data,
        expired: state.currentTime.expireTime ? now > state.currentTime.expireTime : true,
        data: state.currentTime.data,
        nextUpdate: state.currentTime.expireTime ? new Date(state.currentTime.expireTime) : null
      }
    },

    // 获取课程时间配置缓存状态
    lessonTimeStatus: (state) => {
      const now = Date.now()
      return {
        cached: !!state.lessonTime.data,
        expired: state.lessonTime.expireTime ? now > state.lessonTime.expireTime : true,
        data: state.lessonTime.data,
        nextUpdate: state.lessonTime.expireTime ? new Date(state.lessonTime.expireTime) : null
      }
    },

    // 获取完整缓存状态
    cacheStatus: (state) => {
      const now = Date.now()
      return {
        currentTime: {
          cached: !!state.currentTime.data,
          expired: state.currentTime.expireTime ? now > state.currentTime.expireTime : true,
          data: state.currentTime.data,
          nextUpdate: state.currentTime.expireTime ? new Date(state.currentTime.expireTime) : null
        },
        lessonTime: {
          cached: !!state.lessonTime.data,
          expired: state.lessonTime.expireTime ? now > state.lessonTime.expireTime : true,
          data: state.lessonTime.data,
          nextUpdate: state.lessonTime.expireTime ? new Date(state.lessonTime.expireTime) : null
        }
      }
    }
  },

  actions: {
    /**
     * 计算下个周日晚上23:59的时间戳
     * @returns {number} 时间戳
     */
    getNextSundayNight() {
      const now = new Date()
      const currentDay = now.getDay() // 0是周日，1是周一
      
      // 计算到下个周日的天数
      let daysToSunday
      if (currentDay === 0) {
        // 如果今天是周日，则到下个周日是7天
        daysToSunday = 7
      } else {
        // 否则计算到本周日的天数
        daysToSunday = 7 - currentDay
      }

      const nextSunday = new Date(now)
      nextSunday.setDate(now.getDate() + daysToSunday)
      nextSunday.setHours(23, 59, 59, 999) // 设置为23:59:59.999

      return nextSunday.getTime()
    },

    /**
     * 计算15天后的时间戳
     * @returns {number} 时间戳
     */
    getFifteenDaysLater() {
      const now = new Date()
      const fifteenDaysLater = new Date(now.getTime() + 15 * 24 * 60 * 60 * 1000)
      return fifteenDaysLater.getTime()
    },

    /**
     * 检查缓存是否有效
     * @param {string} cacheType 缓存类型 ('currentTime' | 'lessonTime')
     * @returns {boolean} 是否有效
     */
    isCacheValid(cacheType) {
      const cache = this[cacheType]

      if (!cache.data || !cache.expireTime) {
        return false
      }

      const now = Date.now()
      return now <= cache.expireTime
    },

    /**
     * 设置缓存数据
     * @param {string} cacheType 缓存类型 ('currentTime' | 'lessonTime')
     * @param {any} data 要缓存的数据
     * @param {number} expireTime 过期时间戳
     */
    setCache(cacheType, data, expireTime) {
      this[cacheType] = {
        data,
        timestamp: Date.now(),
        expireTime
      }
    },

    /**
     * 清除指定缓存
     * @param {string} cacheType 缓存类型 ('currentTime' | 'lessonTime')
     */
    clearCache(cacheType) {
      this[cacheType] = {
        data: null,
        timestamp: null,
        expireTime: null
      }
    },

    /**
     * 清除所有缓存
     */
    clearAllCache() {
      this.clearCache('currentTime')
      this.clearCache('lessonTime')
    },

    /**
     * 获取当前时间信息（带缓存）
     * @param {boolean} forceRefresh 是否强制刷新缓存
     * @returns {Promise<Object>} 当前时间信息
     */
    async getCurrentTime(forceRefresh = false) {
      // 检查缓存
      if (!forceRefresh && this.isCacheValid('currentTime')) {
        return this.currentTime.data
      }

      try {
        const res = await request({
          url: '/scloudoa/scs/course/tCourseTimetableDetail/getCurrentTime',
          method: 'GET'
        })

        if (res && res.result) {
          const currentTimeData = {
            nowweek: parseInt(res.result.nowweek),
            week: res.result.week,
            currentSemester: res.result.currentSemester,
            weekCount: res.result.weekCount,
            shortDate: res.result.shortDate,
            startDate: res.result.startDate
          }

          // 缓存到下个周日晚上
          const expireTime = this.getNextSundayNight()
          this.setCache('currentTime', currentTimeData, expireTime)

          // 同时保存到旧的存储位置以保持兼容性
          if (currentTimeData.nowweek >= 1 && currentTimeData.nowweek <= 30) {
            uni.setStorageSync('currentWeek', currentTimeData.nowweek.toString())
            uni.setStorageSync('currentWeekTimestamp', Date.now().toString())
          }

          return currentTimeData
        } else {
          throw new Error('API返回数据格式错误')
        }
      } catch (error) {
        console.error('获取当前时间信息失败:', error)

        // 尝试使用旧的本地存储作为降级方案
        const savedWeek = uni.getStorageSync('currentWeek')
        if (savedWeek) {
          const weekNum = parseInt(savedWeek)
          if (weekNum >= 1 && weekNum <= 30) {
            return {
              nowweek: weekNum,
              week: null,
              currentSemester: null,
              weekCount: null
            }
          }
        }

        throw error
      }
    },

    /**
     * 获取课程时间配置（带缓存）
     * @param {boolean} forceRefresh 是否强制刷新缓存
     * @returns {Promise<Array>} 课程时间配置
     */
    async getCourseLessonTime(forceRefresh = false) {
      // 检查缓存
      if (!forceRefresh && this.isCacheValid('lessonTime')) {
        return this.lessonTime.data
      }

      try {
        const res = await request({
          url: '/scloudoa/scs/course/tCourseTimetableDetail/getCourseLessonTime',
          method: 'GET'
        })

        // 处理多种可能的响应格式
        let lessonTimeArray = null

        console.log('Pinia CourseCache: getCourseLessonTime 原始API响应:', res)
        console.log('Pinia CourseCache: API响应类型:', typeof res)
        if (res && typeof res === 'object') {
          console.log('Pinia CourseCache: API响应键:', Object.keys(res))
        }

        // 格式1: 直接是数组
        if (Array.isArray(res) && res.length > 0) {
          lessonTimeArray = res
          console.log('Pinia CourseCache: 使用格式1 - 直接数组')
        }
        // 格式2: 标准API响应格式 {code, message, result}
        else if (res && res.result && Array.isArray(res.result) && res.result.length > 0) {
          lessonTimeArray = res.result
          console.log('Pinia CourseCache: 使用格式2 - result数组')
        }
        // 格式3: 其他嵌套格式
        else if (res && res.data && Array.isArray(res.data) && res.data.length > 0) {
          lessonTimeArray = res.data
          console.log('Pinia CourseCache: 使用格式3 - data数组')
        }
        // 格式4: 带success字段的响应格式，result可能是对象包含数组
        else if (res && res.success && res.result) {
          // 检查result是否直接是数组
          if (Array.isArray(res.result) && res.result.length > 0) {
            lessonTimeArray = res.result
            console.log('Pinia CourseCache: 使用格式4a - success.result数组')
          }
          // 检查result是否是对象，包含数组字段
          else if (typeof res.result === 'object' && res.result !== null) {
            // 特殊处理：检查是否是嵌套的result.result.records结构
            if (res.result.result && res.result.result.records && Array.isArray(res.result.result.records)) {
              lessonTimeArray = res.result.result.records
              console.log('Pinia CourseCache: 使用格式4-nested - success.result.result.records数组')
            }
            // 尝试常见的数组字段名
            else {
              const possibleArrayFields = ['data', 'list', 'items', 'records', 'lessonTimes', 'timeTable']
              for (const field of possibleArrayFields) {
                if (res.result[field] && Array.isArray(res.result[field]) && res.result[field].length > 0) {
                  lessonTimeArray = res.result[field]
                  console.log(`Pinia CourseCache: 使用格式4b - success.result.${field}数组`)
                  break
                }
              }

              // 如果result对象的所有值都检查过了，尝试直接使用result的属性值
              if (!lessonTimeArray) {
                const resultValues = Object.values(res.result)
                for (const value of resultValues) {
                  if (Array.isArray(value) && value.length > 0) {
                    lessonTimeArray = value
                    console.log('Pinia CourseCache: 使用格式4c - success.result对象中的数组值')
                    break
                  }
                }
              }
            }
          }
        }

        if (lessonTimeArray && lessonTimeArray.length > 0) {
          console.log('Pinia CourseCache: 成功解析到课程时间数组:', lessonTimeArray)

          // 验证数据是否包含必要的时间字段
          const hasValidTimeData = lessonTimeArray.some(item =>
            item && (item.name || item.startTime || item.endTime)
          )

          if (!hasValidTimeData) {
            console.warn('Pinia CourseCache: 解析到的数组不包含有效的时间数据，可能是课程数据而非时间配置')
            console.warn('数组第一项示例:', lessonTimeArray[0])

            // 如果不是时间配置数据，使用默认配置
            const defaultTimeTable = [
              { name: '1', startTime: '08:30', endTime: '09:15' },
              { name: '2', startTime: '09:20', endTime: '10:05' },
              { name: '3', startTime: '10:20', endTime: '11:05' },
              { name: '4', startTime: '11:10', endTime: '11:55' },
              { name: '5', startTime: '13:30', endTime: '14:15' },
              { name: '6', startTime: '14:20', endTime: '15:05' },
              { name: '7', startTime: '15:20', endTime: '16:05' },
              { name: '8', startTime: '16:10', endTime: '16:55' },
              { name: '9', startTime: '18:00', endTime: '18:45' },
              { name: '10', startTime: '18:50', endTime: '19:35' },
              { name: '11', startTime: '19:40', endTime: '20:25' },
              { name: '12', startTime: '20:30', endTime: '21:15' }
            ]

            console.log('Pinia CourseCache: 使用默认时间配置')
            const expireTime = this.getFifteenDaysLater()
            this.setCache('lessonTime', defaultTimeTable, expireTime)
            return defaultTimeTable
          }

          const lessonTimeData = lessonTimeArray.map(item => ({
            name: item.name || item.period || item.lessonNumber || '',
            startTime: item.startTime || '',
            endTime: item.endTime || ''
          })).filter(item => item.name && item.startTime && item.endTime)

          console.log('Pinia CourseCache: 格式化后的课程时间数据:', lessonTimeData)

          if (lessonTimeData.length === 0) {
            console.warn('Pinia CourseCache: 格式化后没有有效的时间数据，使用默认配置')
            const defaultTimeTable = [
              { name: '1', startTime: '08:30', endTime: '09:15' },
              { name: '2', startTime: '09:20', endTime: '10:05' },
              { name: '3', startTime: '10:20', endTime: '11:05' },
              { name: '4', startTime: '11:10', endTime: '11:55' },
              { name: '5', startTime: '13:30', endTime: '14:15' },
              { name: '6', startTime: '14:20', endTime: '15:05' },
              { name: '7', startTime: '15:20', endTime: '16:05' },
              { name: '8', startTime: '16:10', endTime: '16:55' },
              { name: '9', startTime: '18:00', endTime: '18:45' },
              { name: '10', startTime: '18:50', endTime: '19:35' },
              { name: '11', startTime: '19:40', endTime: '20:25' },
              { name: '12', startTime: '20:30', endTime: '21:15' }
            ]

            const expireTime = this.getFifteenDaysLater()
            this.setCache('lessonTime', defaultTimeTable, expireTime)
            return defaultTimeTable
          }

          // 缓存15天
          const expireTime = this.getFifteenDaysLater()
          this.setCache('lessonTime', lessonTimeData, expireTime)

          return lessonTimeData
        } else {
          console.warn('Pinia CourseCache: 无法解析API响应格式，将使用默认时间配置')
          console.warn('响应结构:', JSON.stringify(res, null, 2))
          console.warn('响应类型:', typeof res)
          console.warn('响应键:', res ? Object.keys(res) : 'null')

          // 尝试提供更详细的调试信息
          if (res && res.result) {
            console.warn('result类型:', typeof res.result)
            console.warn('result内容:', res.result)
            if (typeof res.result === 'object' && res.result !== null) {
              console.warn('result键:', Object.keys(res.result))
            }
          }

          // 不抛出错误，直接返回默认时间配置
          console.log('Pinia CourseCache: 使用默认时间配置作为降级方案')
          const defaultTimeTable = [
            { name: '1', startTime: '08:30', endTime: '09:15' },
            { name: '2', startTime: '09:20', endTime: '10:05' },
            { name: '3', startTime: '10:20', endTime: '11:05' },
            { name: '4', startTime: '11:10', endTime: '11:55' },
            { name: '5', startTime: '13:30', endTime: '14:15' },
            { name: '6', startTime: '14:20', endTime: '15:05' },
            { name: '7', startTime: '15:20', endTime: '16:05' },
            { name: '8', startTime: '16:10', endTime: '16:55' },
            { name: '9', startTime: '18:00', endTime: '18:45' },
            { name: '10', startTime: '18:50', endTime: '19:35' },
            { name: '11', startTime: '19:40', endTime: '20:25' },
            { name: '12', startTime: '20:30', endTime: '21:15' }
          ]

          // 缓存默认时间配置
          const expireTime = this.getFifteenDaysLater()
          this.setCache('lessonTime', defaultTimeTable, expireTime)

          return defaultTimeTable
        }
      } catch (error) {
        console.error('获取课程时间配置失败:', error)

        // 检查是否是网络错误或认证错误，这些情况下抛出错误
        if (error.statusCode === 401 || error.isTokenInvalid ||
            (error.message && error.message.includes('会话已失效'))) {
          console.error('Pinia CourseCache: 认证错误，抛出异常')
          throw error
        }

        // 其他错误（如网络错误、服务器错误）使用默认配置
        console.log('Pinia CourseCache: 网络或服务器错误，使用默认时间配置作为降级方案')

        const defaultTimeTable = [
          { name: '1', startTime: '08:30', endTime: '09:15' },
          { name: '2', startTime: '09:20', endTime: '10:05' },
          { name: '3', startTime: '10:20', endTime: '11:05' },
          { name: '4', startTime: '11:10', endTime: '11:55' },
          { name: '5', startTime: '13:30', endTime: '14:15' },
          { name: '6', startTime: '14:20', endTime: '15:05' },
          { name: '7', startTime: '15:20', endTime: '16:05' },
          { name: '8', startTime: '16:10', endTime: '16:55' },
          { name: '9', startTime: '18:00', endTime: '18:45' },
          { name: '10', startTime: '18:50', endTime: '19:35' },
          { name: '11', startTime: '19:40', endTime: '20:25' },
          { name: '12', startTime: '20:30', endTime: '21:15' }
        ]

        // 缓存默认时间配置，避免重复请求
        const expireTime = this.getFifteenDaysLater()
        this.setCache('lessonTime', defaultTimeTable, expireTime)

        return defaultTimeTable
      }
    }
  }
})
