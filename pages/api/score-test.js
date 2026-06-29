// 测试学期数据解析
import { processSemesterData } from './score.js';

// 模拟后端返回的数据格式
const mockBackendResponse = {
  "cacheUpdatedAt": "2025-06-20 16:30:31",
  "code": 200,
  "dataSourceNote": "智能缓存数据",
  "dataSourceType": "database",
  "message": "获取成功",
  "result": {
    "code": 0,
    "message": "",
    "result": [
      {
        "courseName": null,
        "courseNumber": null,
        "courseProperty": null,
        "courseScore": null,
        "courseTime": null,
        "credit": null,
        "currentSemester": "2017-2018学年第二学期",
        "dailyScore": null,
        "electiveCredits": 0,
        "failCourseCount": null,
        "finalScore": null,
        "getCredit": null,
        "getPoint": null,
        "id": null,
        "nowweek": null,
        "practicalScore": null,
        "shortDate": null,
        "startDate": null,
        "sumCredits": null,
        "supplementaryScore": null,
        "teacherNames": null,
        "testNote": null,
        "week": null,
        "weekCount": null
      },
      {
        "courseName": null,
        "courseNumber": null,
        "courseProperty": null,
        "courseScore": null,
        "courseTime": null,
        "credit": null,
        "currentSemester": "2018-2019学年第一学期",
        "dailyScore": null,
        "electiveCredits": 0,
        "failCourseCount": null,
        "finalScore": null,
        "getCredit": null,
        "getPoint": null,
        "id": null,
        "nowweek": null,
        "practicalScore": null,
        "shortDate": null,
        "startDate": null,
        "sumCredits": null,
        "supplementaryScore": null,
        "teacherNames": null,
        "testNote": null,
        "week": null,
        "weekCount": null
      },
      {
        "courseName": null,
        "courseNumber": null,
        "courseProperty": null,
        "courseScore": null,
        "courseTime": null,
        "credit": null,
        "currentSemester": "2024-2025学年第二学期",
        "dailyScore": null,
        "electiveCredits": 0,
        "failCourseCount": null,
        "finalScore": null,
        "getCredit": null,
        "getPoint": null,
        "id": null,
        "nowweek": null,
        "practicalScore": null,
        "shortDate": null,
        "startDate": null,
        "sumCredits": null,
        "supplementaryScore": null,
        "teacherNames": null,
        "testNote": null,
        "week": null,
        "weekCount": null
      }
    ],
    "success": true,
    "timestamp": 1750376848628
  },
  "success": true,
  "timestamp": "2025-06-20 16:31:38"
};

// 测试数据解析
console.log('=== 测试学期数据解析 ===');

// 模拟前端解析逻辑
function testSemesterParsing(response) {
  console.log('原始响应:', JSON.stringify(response, null, 2));
  
  let semesterData = [];
  
  // 使用修复后的解析逻辑
  if (response && response.result && response.result.result && Array.isArray(response.result.result)) {
    // 三层嵌套格式：result.result.result
    semesterData = response.result.result;
    console.log('✅ 使用三层嵌套格式 - result.result.result，数据条数:', semesterData.length);
  } else if (response && Array.isArray(response.result)) {
    // 二层嵌套格式：result
    semesterData = response.result;
    console.log('✅ 使用二层嵌套格式 - result，数据条数:', semesterData.length);
  } else if (response && Array.isArray(response)) {
    // 直接数组格式
    semesterData = response;
    console.log('✅ 使用直接数组格式，数据条数:', semesterData.length);
  } else {
    console.error('❌ 学期列表响应格式异常:', response);
    return [];
  }
  
  console.log('解析到的原始学期数据:', semesterData.slice(0, 3)); // 只显示前3条
  
  // 处理学期数据
  const processedSemesters = processSemesterData(semesterData);
  console.log('处理后的学期数据:', processedSemesters);
  
  return processedSemesters;
}

// 执行测试
const result = testSemesterParsing(mockBackendResponse);
console.log('=== 测试结果 ===');
console.log('学期数量:', result.length);
console.log('学期列表:', result.map(s => s.currentSemester));

// 验证学期选择器数据格式
console.log('=== 验证选择器格式 ===');
const pickerData = result.map(item => ({
  name: item.currentSemester,
  value: item.currentSemester
}));
console.log('选择器数据:', pickerData);

export { testSemesterParsing };
