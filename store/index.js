/**
 * Pinia Store 统一导出入口
 * 提供所有 Store 的统一导出，便于模块化管理
 */

// 用户认证状态
export { useUserStore } from './user.js'

// 学校账号绑定状态
export { useSchoolAccountStore } from './schoolAccount.js'

// 课程数据缓存状态
export { useCourseCache } from './courseCache.js'
