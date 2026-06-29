import { request } from '../../utils/request.js';
import { useCourseCache } from '../../store/courseCache.js';

/**
 * 获取学期周数信息
 * @returns {Promise<Object>} 返回学期周数信息
 */
export const getTermWeekNum = async () => {
    try {
      const res = await request({
        url: '/scloud/courseTimetable/getTermWeekNum',
        method: 'GET'
      });
      return res;
    } catch (error) {
      console.error('获取学期周数信息失败:', error);
      throw error;
    }
};

/**
 * 获取课程时间段信息（使用 Pinia Store）
 * @returns {Promise<Array>} 返回课程时间段信息数组
 */
export const getCourseLessonTime = async () => {
    try {
      const token = uni.getStorageSync('token');
      if (!token) {
        throw new Error('未登录，请先登录');
      }
      const courseCacheStore = useCourseCache();
      return await courseCacheStore.getCourseLessonTime();
    } catch (error) {
      console.error('获取课程时间段失败:', error);
      throw error;
    }
};

/**
 * 获取完整课表列表
 * @param {Object} options 可选参数
 * @param {number} options.week 指定周次，默认为当前周
 * @param {string} options.currentSemester 当前学期，如"2024-2025学年第二学期"
 * @returns {Promise<Array>} 返回课表列表数组
 */
export const getTimetableList = async (options = {}) => {
    try {
      const loginType = uni.getStorageSync('loginType');
      if (loginType === 'mobile') {
        try {
          const result = await getMCourseTimeTableByWeek({
            currentSemester: options.currentSemester || '2024-2025学年第二学期',
            nowWeek: options.week || undefined
          });
          if (result && typeof result === 'object' && result.courses) {
            return result.courses;
          } else if (Array.isArray(result)) {
            return result;
          } else {
            return [];
          }
        } catch (mobileError) {
          console.error('移动端接口调用失败:', mobileError);
          throw mobileError;
        }
      }

      const params = {
        week: options.week || undefined,
        currentSemester: options.currentSemester || '2024-2025学年第二学期'
      };

      let queryString = '';
      if (params.week) {
        queryString += `week=${params.week}`;
      }
      if (params.currentSemester) {
        if (queryString) queryString += '&';
        queryString += `currentSemester=${encodeURIComponent(params.currentSemester)}`;
      }

      const res = await request({
        url: '/scloud/courseTimetableDetail/getTimetableList',
        method: 'POST',
        header: {
          'content-type': 'application/x-www-form-urlencoded'
        },
        data: queryString
      });

      if (Array.isArray(res)) {
        return res;
      } else if (res && res.result && typeof res.result === 'object') {
        const weekCourses = res.result;
        let allCourses = [];
        const weekdays = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];
        weekdays.forEach(weekday => {
          if (weekCourses[weekday] && Array.isArray(weekCourses[weekday])) {
            allCourses = allCourses.concat(weekCourses[weekday]);
          }
        });
        return allCourses;
      } else if (res && res.data && Array.isArray(res.data)) {
        return res.data;
      } else {
        return [];
      }
    } catch (error) {
      console.error('获取完整课表列表失败:', error);
      throw error;
    }
};

/**
 * 格式化时间范围
 * @param {string} startTime 开始时间
 * @param {string} endTime 结束时间
 * @returns {string} 格式化后的时间范围
 */
const formatTimeRange = (startTime, endTime) => {
    if (!startTime || !endTime) return '';
    const formatTime = (timeStr) => {
      if (!timeStr) return '';
      const parts = timeStr.split(':');
      return parts.length >= 2 ? `${parts[0]}:${parts[1]}` : timeStr;
    };
    return `${formatTime(startTime)}-${formatTime(endTime)}`;
};

/**
 * 根据节次和长度计算时间范围
 * @param {number} startScope 开始节次
 * @param {number} length 课程长度
 * @returns {string} 格式化后的时间范围
 */
const formatTime = (startScope, length) => {
    if (!startScope || !length) return '';
    const timeTable = [
      { start: '08:00', end: '08:45' }, { start: '08:50', end: '09:35' },
      { start: '09:50', end: '10:35' }, { start: '10:40', end: '11:25' },
      { start: '11:30', end: '12:15' }, { start: '14:00', end: '14:45' },
      { start: '14:50', end: '15:35' }, { start: '15:50', end: '16:35' },
      { start: '16:40', end: '17:25' }, { start: '18:30', end: '19:15' },
      { start: '19:20', end: '20:05' }, { start: '20:10', end: '20:55' }
    ];
    const start = parseInt(startScope);
    const end = start + parseInt(length) - 1;
    if (start < 1 || start > timeTable.length) return '';
    if (end < 1 || end > timeTable.length) return '';
    return `${timeTable[start - 1].start}-${timeTable[end - 1].end}`;
};

/**
 * 格式化节数
 * @param {string|Array} sections 节数信息
 * @returns {string} 格式化后的节数
 */
const formatSection = (sections) => {
    if (!sections) return '';
    if (Array.isArray(sections)) {
      if (sections.length === 0) return '';
      if (sections.length === 1) return sections[0].toString();
      return `${sections[0]}-${sections[sections.length - 1]}`;
    }
    return sections.toString();
};

/**
 * 移动端获取学期和当前周信息
 * @returns {Promise<Object>} 返回学期和当前周信息
 */
export const getMCourseSchoolTimetable = async () => {
  try {
    const token = uni.getStorageSync('token');
    if (!token) {
      throw new Error('未登录，请先登录');
    }
    const res = await request({
      url: '/scloudoa/userQuery/tSysUser/getCourseSchoolTimetable',
      method: 'GET'
    });
    if (res && res.result) {
      return res.result;
    } else if (res && res.data) {
      return res.data;
    } else {
      return {};
    }
  } catch (error) {
    console.error('获取学期和当前周信息失败:', error);
    throw error;
  }
};

/**
 * 移动端获取课表时间配置（使用 Pinia Store 单例模式）
 * @returns {Promise<Array>} 返回课表时间配置数组
 */
export const getMCourseLessonTime = async () => {
  try {
    const token = uni.getStorageSync('token');
    if (!token) {
      throw new Error('未登录，请先登录');
    }
      const courseCacheStore = useCourseCache();
    return await courseCacheStore.getCourseLessonTime();
  } catch (error) {
    console.error('获取课表时间配置失败:', error);
    throw error;
  }
};

/**
 * 移动端获取今日课表
 * @returns {Promise<Array>} 返回今日课表数组
 */
export const getMCourseTimeTableByDay = async () => {
  try {
    const token = uni.getStorageSync('token');
    if (!token) {
      throw new Error('未登录，请先登录');
    }
    const res = await request({
      url: '/scloudoa/scs/course/tCourseTimetableDetail/getCourseTimeTableByDay',
      method: 'GET',
      data: { current: 1, size: -1 }
    });

    console.log('[getMCourseTimeTableByDay] 原始响应成功，records数量:', res?.result?.records?.length);
    let courseArray = null;
    // 优先级：标准化后单层 > 标准化后三层嵌套 > 数据库缓存 > 直接数组
    if (res?.result?.records && Array.isArray(res.result.records)) {
      courseArray = res.result.records;
    } else if (res?.result?.result?.records && Array.isArray(res.result.result.records)) {
      courseArray = res.result.result.records;
      console.log('[getMCourseTimeTableByDay] 使用 res.result.result.records (标准化三层)');
    } else if (res?.data?.records && Array.isArray(res.data.records)) {
      courseArray = res.data.records;
    } else if (Array.isArray(res)) {
      courseArray = res;
    } else if (res?.data && Array.isArray(res.data)) {
      courseArray = res.data;
    }

    if (courseArray && courseArray.length >= 0) {
      return courseArray;
    } else {
      return [];
    }
  } catch (error) {
    console.error('获取今日课表失败:', error);
    throw error;
  }
};

/**
 * 移动端按周获取课表
 * @param {Object} options 查询参数
 * @param {string} options.currentSemester 当前学期
 * @param {number} options.nowWeek 当前周次
 * @returns {Promise<Object>} 返回课表对象
 */
export const getMCourseTimeTableByWeek = async (options = {}) => {
  try {
    const token = uni.getStorageSync('token');
    if (!token) {
      throw new Error('未登录，请先登录');
    }

    const res = await request({
      url: '/scloudoa/scs/course/tCourseTimetableDetail/getCourseTimeTableByWeek',
      method: 'GET',
      data: {
        current: 1,
        size: 84,
        currentSemester: options.currentSemester || '',
        nowWeek: options.nowWeek || ''
      }
    });

    const fromCache = !!(res && res.dataSourceType === 'database');
    const cacheUpdatedAt = (res && res.cacheUpdatedAt) || '';

    let courseArray = null;
    if (res?.result?.records && Array.isArray(res.result.records)) {
      courseArray = res.result.records;
    } else if (res?.result && Array.isArray(res.result)) {
      courseArray = res.result;
    } else if (Array.isArray(res)) {
      courseArray = res;
    }

    if (courseArray && courseArray.length >= 0) {
      const weekDayMap = { 1: '一', 2: '二', 3: '三', 4: '四', 5: '五', 6: '六', 7: '日' };
      courseArray = courseArray.map(course => {
        if (course.weekday !== undefined && course.week === undefined) {
          return { ...course, week: weekDayMap[course.weekday] || course.weekday.toString() };
        }
        return course;
      });
      return { courses: courseArray, fromCache, cacheUpdatedAt };
    } else {
      return { courses: [], fromCache, cacheUpdatedAt };
    }
  } catch (error) {
    console.error('按周获取课表失败:', error);
    throw error;
  }
};