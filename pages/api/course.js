import { getMCourseTimeTableByDay } from './schedule.js';
import { useCourseCache } from '../../store/courseCache.js';

/**
 * 课程相关API
 */

/**
 * 格式化时间范围
 * @param {string} startTime 开始时间，格式为HH:mm:ss
 * @param {string} endTime 结束时间，格式为HH:mm:ss
 * @returns {string} 格式化后的时间范围，例如：08:00-09:40
 */
const formatTimeRange = (startTime, endTime) => {
  if (!startTime || !endTime) return '';
  const start = startTime.substring(0, 5);
  const end = endTime.substring(0, 5);
  return `${start}-${end}`;
};

/**
 * 根据节次和长度格式化时间
 * @param {number} startLesson 开始节次
 * @param {number} length 课程长度
 * @returns {string} 格式化后的节次信息
 */
const formatTime = (startLesson, length) => {
  if (!startLesson) return '';
  const endLesson = startLesson + (length || 1) - 1;
  return `第${startLesson}-${endLesson}节`;
};

/**
 * 获取今日课表
 * @param {Object} options 可选参数
 * @param {string} options.date 指定日期，格式为YYYY-MM-DD，默认为今天
 * @returns {Promise<Object>} 包含原始数据和格式化数据的对象
 */
export const getTodayTimetable = async (options = {}) => {
  const token = uni.getStorageSync('token');
  console.log('[getTodayTimetable] 调用，loginType:', uni.getStorageSync('loginType'), 'hasToken:', !!token);
  try {
    return await getMobileTimeTable();
  } catch (error) {
    console.error('[getTodayTimetable] 失败:', error);
    return {
      success: false,
      message: error.message || '获取课表失败',
      originalData: [],
      formattedData: []
    };
  }
};

/**
 * 获取今日课表（移动端接口）
 * @returns {Promise<Object>} 包含原始数据和格式化数据的对象
 */
async function getMobileTimeTable() {
  const response = await getMCourseTimeTableByDay();

  let originalData = [];
  if (Array.isArray(response)) {
    originalData = response;
  } else if (response?.result?.records) {
    originalData = response.result.records;
  } else if (response?.data && Array.isArray(response.data)) {
    originalData = response.data;
  }

  if (originalData.length === 0) {
    return { success: true, originalData: [], formattedData: [] };
  }

  let currentWeekInfo = null;
  try {
    const courseCacheStore = useCourseCache();
    currentWeekInfo = await courseCacheStore.getCurrentTime();
  } catch (error) {}

  if (currentWeekInfo) {
    originalData = originalData.map(item => ({
      ...item,
      nowWeek: currentWeekInfo.nowweek,
      currentWeek: currentWeekInfo.week,
      currentSemester: currentWeekInfo.currentSemester,
      weekCount: currentWeekInfo.weekCount
    }));
  }

  const formattedData = originalData.map(item => {
    const name = item.name || item.courseName || '未知课程';
    const teacher = item.teacher || item.teacherNames || '未知教师';
    const classroom = item.classroom || item.classroomName || '未知教室';
    const section = item.scope || item.lessonScope || '';

    let time = '';
    if (item.time) {
      time = item.time;
    } else if (item.startTime && item.endTime) {
      time = formatTimeRange(item.startTime, item.endTime);
    } else if (item.startLessonScope) {
      time = formatTime(item.startLessonScope, item.lessonScopeLenght || 1);
    }

    return {
      name,
      teacher,
      classroom,
      time,
      section: typeof section === 'string' ? section : (section ? JSON.stringify(section) : '')
    };
  });

  return { success: true, originalData, formattedData };
}
