import { request } from '../../utils/request.js';

/**
 * 成绩查询相关API
 */

/**
 * 获取学期列表
 * @returns {Promise<Array>} 学期列表
 */
export const getSemesterList = async () => {
  console.log('获取学期列表');

  try {
    const res = await request({
      url: '/api/m/scs/course/tCourseScore/getSemester',
      method: 'GET'
    });

    console.log('学期列表响应:', res);

    let semesterData = [];

    // 处理后端返回的嵌套格式，支持多种数据结构
    if (res && res.result && res.result.result && Array.isArray(res.result.result)) {
      // 三层嵌套格式：result.result.result
      semesterData = res.result.result;
      console.log('getSemesterList: 使用三层嵌套格式 - result.result.result，数据条数:', semesterData.length);
    } else if (res && Array.isArray(res.result)) {
      // 二层嵌套格式：result
      semesterData = res.result;
      console.log('getSemesterList: 使用二层嵌套格式 - result，数据条数:', semesterData.length);
    } else if (res && Array.isArray(res)) {
      // 直接数组格式
      semesterData = res;
      console.log('getSemesterList: 使用直接数组格式，数据条数:', semesterData.length);
    } else {
      console.warn('学期列表响应格式异常:', res);
      console.warn('期望的格式: result.result.result 或 result 或直接数组');
      return [];
    }

    // 处理学期数据，提取唯一的学期名称
    const processedSemesters = processSemesterData(semesterData);
    console.log('处理后的学期数据:', processedSemesters);

    return processedSemesters;
  } catch (error) {
    console.error('获取学期列表失败:', error);
    throw new Error('获取学期列表失败: ' + (error.message || '网络错误'));
  }
};

/**
 * 获取课程成绩列表
 * @param {string} semesterName 学期名称，默认为"全部"获取所有学期成绩
 * @returns {Promise<Array>} 课程成绩列表
 */
export const getCourseScores = async (semesterName = '全部') => {
  console.log('获取课程成绩，学期:', semesterName);

  try {
    const res = await request({
      url: '/api/m/scs/course/tCourseScore/getScoreList',
      method: 'GET',
      data: {
        semesterName: semesterName
      }
    });

    console.log('课程成绩响应:', res);

    const meta = {
		fromCache: !!(res && (res.fromCache || res.dataSourceType === 'database')),
		cacheUpdatedAt: (res && res.cacheUpdatedAt) || '',
		dataSourceType: (res && res.dataSourceType) || ''
	}

    if (res && res.success && res.result) {
      // 处理后端返回的嵌套格式，支持多种数据结构
      if (res.result.result && res.result.result.records && Array.isArray(res.result.result.records)) {
        // 三层嵌套格式：result.result.records
        console.log('getCourseScores: 使用三层嵌套格式 - result.result.records，数据条数:', res.result.result.records.length);
        return { data: res.result.result.records, meta };
      } else if (res.result.result && Array.isArray(res.result.result)) {
        // 三层嵌套格式：result.result
        console.log('getCourseScores: 使用三层嵌套格式 - result.result，数据条数:', res.result.result.length);
        return { data: res.result.result, meta };
      } else if (res.result.records && Array.isArray(res.result.records)) {
        // 二层嵌套格式：result.records
        console.log('getCourseScores: 使用二层嵌套格式 - result.records，数据条数:', res.result.records.length);
        return { data: res.result.records, meta };
      } else if (Array.isArray(res.result)) {
        // 二层嵌套格式：result
        console.log('getCourseScores: 使用二层嵌套格式 - result数组，数据条数:', res.result.length);
        return { data: res.result, meta };
      } else {
        console.warn('课程成绩数据格式异常:', res.result);
        console.warn('期望的格式: result.result.records 或 result.result 或 result.records 或 result');
        return { data: [], meta };
      }
    } else {
      console.warn('课程成绩响应格式异常:', res);
      return { data: [], meta };
    }
  } catch (error) {
    console.error('获取课程成绩失败:', error);
    throw new Error('获取课程成绩失败: ' + (error.message || '网络错误'));
  }
};

/**
 * 获取全部学期的课程成绩列表
 * @returns {Promise<Array>} 全部学期的课程成绩列表
 */
export const getAllCourseScores = async () => {
  console.log('获取全部学期的课程成绩');
  return await getCourseScores('全部');
};

/**
 * 获取学期统计信息
 * @param {string} semesterName 学期名称
 * @returns {Promise<Object>} 学期统计信息
 */
export const getSemesterStatistics = async (semesterName) => {
  console.log('获取学期统计信息，学期:', semesterName);
  
  if (!semesterName) {
    throw new Error('学期名称不能为空');
  }
  
  try {
    const res = await request({
      url: '/api/m/scs/course/tCourseScore/getSemesterScore',
      method: 'GET',
      data: {
        semesterName: semesterName
      }
    });
    
    console.log('学期统计响应:', res);

    const meta = {
		fromCache: !!(res && (res.fromCache || res.dataSourceType === 'database')),
		cacheUpdatedAt: (res && res.cacheUpdatedAt) || '',
		dataSourceType: (res && res.dataSourceType) || ''
	}

    if (res && res.success && res.result) {
      // 处理后端返回的嵌套格式，支持多种数据结构
      if (res.result.result) {
        // 三层嵌套格式：result.result
        console.log('getSemesterStatistics: 使用三层嵌套格式 - result.result');
        return { data: res.result.result, meta };
      } else {
        // 二层嵌套格式：result
        console.log('getSemesterStatistics: 使用二层嵌套格式 - result');
        return { data: res.result, meta };
      }
    } else {
      console.warn('学期统计响应格式异常:', res);
      return { data: null, meta };
    }
  } catch (error) {
    console.error('获取学期统计失败:', error);
    // 统计信息获取失败不抛出错误，让调用方自行计算
    return { data: null, meta: { fromCache: false, cacheUpdatedAt: '', dataSourceType: '' } };
  }
};

/**
 * 获取学习数据（可能包含更多统计信息）
 * @param {string} semesterName 学期名称
 * @returns {Promise<Object>} 学习数据
 */
export const getLearningData = async (semesterName) => {
  console.log('获取学习数据，学期:', semesterName);
  
  try {
    const res = await request({
      url: '/api/m/scs/course/tCourseScore/getLearningData',
      method: 'GET',
      data: semesterName ? { semesterName } : {}
    });
    
    console.log('学习数据响应:', res);

    if (res && res.success && res.result) {
      // 处理后端返回的嵌套格式，支持多种数据结构
      if (res.result.result) {
        // 三层嵌套格式：result.result
        console.log('getLearningData: 使用三层嵌套格式 - result.result');
        return res.result.result;
      } else {
        // 二层嵌套格式：result
        console.log('getLearningData: 使用二层嵌套格式 - result');
        return res.result;
      }
    } else {
      console.warn('学习数据响应格式异常:', res);
      return null;
    }
  } catch (error) {
    console.error('获取学习数据失败:', error);
    return null;
  }
};

/**
 * 处理学期数据，提取唯一的学期名称
 * @param {Array} rawSemesterData 原始学期数据
 * @returns {Array} 处理后的学期数据
 */
export const processSemesterData = (rawSemesterData) => {
  if (!Array.isArray(rawSemesterData)) {
    console.warn('学期数据不是数组格式:', rawSemesterData);
    return [];
  }

  // 使用 Set 来去重学期名称
  const semesterSet = new Set();
  const processedSemesters = [];

  rawSemesterData.forEach((item, index) => {
    console.log(`处理学期项目 ${index}:`, item);

    let semesterName = '';

    // 处理不同的数据格式
    if (typeof item === 'string') {
      semesterName = item;
    } else if (typeof item === 'object' && item !== null) {
      // 优先使用 currentSemester 字段
      semesterName = item.currentSemester ||
                    item.semesterName ||
                    item.semester ||
                    item.name ||
                    '';
    }

    // 过滤掉空值和重复值
    if (semesterName && !semesterSet.has(semesterName)) {
      semesterSet.add(semesterName);
      processedSemesters.push({
        currentSemester: semesterName,
        semesterName: semesterName,
        name: semesterName
      });
      console.log(`添加学期: ${semesterName}`);
    }
  });

  // 按学期名称排序（可选）
  processedSemesters.sort((a, b) => {
    // 简单的字符串排序，可以根据需要调整排序逻辑
    return b.currentSemester.localeCompare(a.currentSemester);
  });

  console.log('最终处理的学期列表:', processedSemesters);
  return processedSemesters;
};

/**
 * 计算GPA
 * @param {number} score 分数
 * @returns {string} GPA值
 */
export const calculateGPA = (score) => {
  const gpaScale = [
    { min: 90, gpa: '4.0' },
    { min: 85, gpa: '3.7' },
    { min: 80, gpa: '3.3' },
    { min: 75, gpa: '3.0' },
    { min: 70, gpa: '2.7' },
    { min: 65, gpa: '2.3' },
    { min: 60, gpa: '2.0' },
    { min: 0, gpa: '0.0' }
  ];
  
  const gpaItem = gpaScale.find(item => score >= item.min);
  return gpaItem ? gpaItem.gpa : '0.0';
};

/**
 * 计算学期统计信息
 * @param {Array} courseScores 课程成绩列表
 * @returns {Object} 统计信息
 */
export const calculateSemesterStats = (courseScores) => {
  if (!Array.isArray(courseScores) || courseScores.length === 0) {
    return {
      gpa: '0.00',
      averageScore: '0.00',
      creditTotal: '0',
      passRate: '0'
    };
  }

  const stats = courseScores.reduce((acc, course) => {
    const credit = parseFloat(course.credit || course.courseCredit || 0);

    // 只统计有学分的课程
    if (credit > 0) {
      acc.totalCredit += credit;
      acc.courseCount++;

      // 获取用于计算的分数（处理字母等级）
      let scoreForCalculation = 0;
      const finalScore = course.finalScore;
      if (typeof finalScore === 'string' && isNaN(parseFloat(finalScore))) {
        // 字母等级，使用 courseScore
        scoreForCalculation = parseFloat(course.courseScore || 0);
      } else {
        // 数字分数
        scoreForCalculation = parseFloat(finalScore || course.courseScore || 0);
      }

      // 累计分数（用于平均分计算）
      if (scoreForCalculation > 0) {
        acc.totalScore += scoreForCalculation;
        acc.validScoreCount++;
      }

      // 计算GPA权重，优先使用getPoint字段
      const gpa = parseFloat(course.getPoint || course.gpa || calculateGPA(scoreForCalculation));
      acc.totalGpaWeight += gpa * credit;

      // 判断是否通过：优先使用绩点判断，绩点大于0说明通过
      const getPoint = parseFloat(course.getPoint || course.gpa || 0);
      if (getPoint > 0) {
        acc.passCount++;
      } else if (getPoint === 0 && scoreForCalculation >= 60) {
        // 如果绩点为0但分数>=60，也算通过（兼容处理）
        acc.passCount++;
      }
    }
    return acc;
  }, {
    totalScore: 0,
    totalCredit: 0,
    totalGpaWeight: 0,
    passCount: 0,
    courseCount: 0,
    validScoreCount: 0
  });

  return {
    gpa: stats.courseCount > 0 && stats.totalCredit > 0
      ? (stats.totalGpaWeight / stats.totalCredit).toFixed(2)
      : '0.00',
    averageScore: stats.validScoreCount > 0
      ? (stats.totalScore / stats.validScoreCount).toFixed(2)
      : '0.00',
    creditTotal: stats.totalCredit.toFixed(1),
    passRate: stats.courseCount > 0
      ? Math.round((stats.passCount / stats.courseCount) * 100).toString()
      : '0'
  };
};

/**
 * 格式化课程成绩数据
 * @param {Array} rawScores 原始成绩数据
 * @returns {Array} 格式化后的成绩数据
 */
export const formatCourseScores = (rawScores) => {
  if (!Array.isArray(rawScores)) {
    return [];
  }

  return rawScores.map(course => {
    // 正确的数据结构：
    // dailyScore: 平时分
    // courseScore: 考试分
    // finalScore: 最终分

    // 确保分数为数字类型，便于统计计算
    const dailyScore = parseFloat(course.dailyScore) || 0;
    const courseScore = parseFloat(course.courseScore || course.score) || 0; // 考试分

    // 处理 finalScore，保持原始值（可能是字母等级）
    let finalScore = course.finalScore;
    if (finalScore === undefined || finalScore === null || finalScore === '') {
      // 如果没有 finalScore，使用 courseScore
      finalScore = courseScore || 0;
    }

    const credit = parseFloat(course.credit || course.courseCredit) || 0;

    // 计算绩点时，如果 finalScore 是字母等级，使用 courseScore
    const scoreForGPA = parseFloat(finalScore) || courseScore || 0;
    const getPoint = parseFloat(course.getPoint || course.gpa) || calculateGPA(scoreForGPA);

    // 调试日志已移除，保持性能

    return {
      courseName: course.courseName || course.name || '未知课程',
      courseScore: courseScore, // 考试分
      finalScore: finalScore,   // 最终分（保持原始值，可能是字母等级）
      dailyScore: dailyScore,   // 平时分
      credit: credit,
      gpa: getPoint,
      getPoint: getPoint,
      examType: course.examType || course.type || '正常考试',
      courseProperty: course.courseProperty || course.examType || course.type || '必修课',
      courseCode: course.courseCode || course.code || '',
      courseNumber: course.courseNumber || '',
      teacher: course.teacher || course.teacherName || '',
      teacherNames: course.teacherNames || course.teacher || course.teacherName || '',
      semester: course.semester || course.semesterName || course.currentSemester || '',
      testNote: course.testNote || '正常',
      supplementaryScore: course.supplementaryScore || '0',
      courseTime: course.courseTime || 0,
      electiveCredits: course.electiveCredits || 0,
      id: course.id || ''
    };
  });
};

/**
 * 获取成绩等级样式类名
 * @param {string|number} score 分数
 * @returns {string} 样式类名
 */
export const getScoreClass = (score) => {
  const numScore = parseFloat(score);
  if (isNaN(numScore) || numScore === 0) return '';
  if (numScore >= 90) return 'score-excellent';
  if (numScore >= 80) return 'score-good';
  if (numScore >= 70) return 'score-medium';
  if (numScore >= 60) return 'score-pass';
  return 'score-fail';
};
