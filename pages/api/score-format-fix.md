# 成绩页面学期选择修复 - 解决三层嵌套数据格式问题

## 问题描述

前端成绩页面 `/pages/user/score` 的学期选择器没有正常显示学期选择，检查后发现后端返回的数据格式是三层嵌套的 `result.result.result`，而前端只处理了二层嵌套 `result`。

## 问题根源分析

### 1. 后端返回的数据格式

**实际返回格式：**
```json
{
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
        "currentSemester": "2017-2018学年第二学期",
        "courseName": null,
        "courseNumber": null,
        // ... 其他字段
      },
      {
        "currentSemester": "2018-2019学年第一学期",
        // ... 其他字段
      }
      // ... 更多学期数据
    ],
    "success": true,
    "timestamp": 1750376848628
  },
  "success": true,
  "timestamp": "2025-06-20 16:31:38"
}
```

**数据路径：** `response.result.result` (三层嵌套)

### 2. 前端原始解析逻辑

**原始代码问题：**
```javascript
// 只处理了二层嵌套
if (res && Array.isArray(res.result)) {
  semesterData = res.result; // ❌ 错误：应该是 res.result.result
  console.log('getSemesterList: 使用标准格式 - result，数据条数:', semesterData.length);
} else if (res && Array.isArray(res)) {
  semesterData = res;
  console.log('getSemesterList: 使用兼容格式 - 直接数组，数据条数:', semesterData.length);
} else {
  console.warn('学期列表响应格式异常:', res);
  return [];
}
```

**问题分析：**
- 前端期望 `res.result` 是数组
- 实际上 `res.result` 是对象，真正的数组在 `res.result.result`
- 导致 `semesterData` 为空数组，学期选择器无数据

## 解决方案

### 1. 修复 getSemesterList 函数

**修复后的解析逻辑：**
```javascript
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
```

### 2. 修复 getCourseScores 函数

**修复后的解析逻辑：**
```javascript
if (res && res.success && res.result) {
  // 处理后端返回的嵌套格式，支持多种数据结构
  if (res.result.result && res.result.result.records && Array.isArray(res.result.result.records)) {
    // 三层嵌套格式：result.result.records
    console.log('getCourseScores: 使用三层嵌套格式 - result.result.records，数据条数:', res.result.result.records.length);
    return res.result.result.records;
  } else if (res.result.result && Array.isArray(res.result.result)) {
    // 三层嵌套格式：result.result
    console.log('getCourseScores: 使用三层嵌套格式 - result.result，数据条数:', res.result.result.length);
    return res.result.result;
  } else if (res.result.records && Array.isArray(res.result.records)) {
    // 二层嵌套格式：result.records
    console.log('getCourseScores: 使用二层嵌套格式 - result.records，数据条数:', res.result.records.length);
    return res.result.records;
  } else if (Array.isArray(res.result)) {
    // 二层嵌套格式：result
    console.log('getCourseScores: 使用二层嵌套格式 - result数组，数据条数:', res.result.length);
    return res.result;
  } else {
    console.warn('课程成绩数据格式异常:', res.result);
    console.warn('期望的格式: result.result.records 或 result.result 或 result.records 或 result');
    return [];
  }
}
```

### 3. 修复其他相关函数

**统一修复了以下函数：**
- `getSemesterStatistics` - 学期统计信息
- `getLearningData` - 学习数据

**修复模式：**
```javascript
if (res && res.success && res.result) {
  // 处理后端返回的嵌套格式，支持多种数据结构
  if (res.result.result) {
    // 三层嵌套格式：result.result
    console.log('函数名: 使用三层嵌套格式 - result.result');
    return res.result.result;
  } else {
    // 二层嵌套格式：result
    console.log('函数名: 使用二层嵌套格式 - result');
    return res.result;
  }
}
```

## 兼容性保证

### 1. 多格式支持

修复后的代码支持以下数据格式：

1. **三层嵌套格式** (当前后端返回)：
   ```json
   {
     "result": {
       "result": [...]
     }
   }
   ```

2. **二层嵌套格式** (标准格式)：
   ```json
   {
     "result": [...]
   }
   ```

3. **直接数组格式** (简化格式)：
   ```json
   [...]
   ```

### 2. 向后兼容

- 保持原有的二层嵌套格式支持
- 添加三层嵌套格式支持
- 不影响现有的直接数组格式

### 3. 错误处理

```javascript
// 详细的错误日志
console.warn('学期列表响应格式异常:', res);
console.warn('期望的格式: result.result.result 或 result 或直接数组');
```

## 数据处理流程

### 1. 学期数据提取

**原始数据：**
```json
[
  {
    "currentSemester": "2017-2018学年第二学期",
    "courseName": null,
    "courseNumber": null,
    // ... 其他字段都是null
  }
]
```

**提取学期信息：**
```javascript
const processSemesterData = (rawSemesterData) => {
  // 提取 currentSemester 字段
  const semesterName = item.currentSemester ||
                      item.semesterName ||
                      item.semester ||
                      item.name ||
                      '';
  
  // 去重和排序
  // ...
};
```

### 2. 选择器数据格式

**最终格式：**
```javascript
[
  {
    name: "全部学期",
    value: "全部"
  },
  {
    name: "2024-2025学年第二学期",
    value: "2024-2025学年第二学期"
  },
  {
    name: "2024-2025学年第一学期", 
    value: "2024-2025学年第一学期"
  }
  // ...
]
```

## 测试验证

### 1. 数据解析测试

创建了测试文件 `score-test.js` 来验证数据解析逻辑：

```javascript
// 模拟后端响应
const mockResponse = { /* 实际的后端响应数据 */ };

// 测试解析
const result = testSemesterParsing(mockResponse);
console.log('学期数量:', result.length);
console.log('学期列表:', result.map(s => s.currentSemester));
```

### 2. 功能测试

**测试场景：**
1. ✅ 学期选择器正常显示学期列表
2. ✅ 选择学期后正常加载对应成绩
3. ✅ 快速查询按钮正常工作
4. ✅ 统计信息正确计算

## 影响范围

### 1. 修复的文件

- `pages/api/score.js` - 成绩API函数
- 新增 `pages/api/score-test.js` - 测试文件

### 2. 影响的功能

- ✅ 学期选择器显示
- ✅ 成绩数据加载
- ✅ 学期统计信息
- ✅ 快速查询功能

### 3. 用户体验改善

- **修复前**：学期选择器显示"Loading..."，无法选择学期
- **修复后**：正常显示所有学期，可以正常选择和查询成绩

## 总结

通过修复前端API解析逻辑，成功解决了学期选择器无法显示的问题：

1. **🔍 问题定位**：准确识别了三层嵌套数据格式问题
2. **🔧 精准修复**：添加了对三层嵌套格式的支持
3. **🛡️ 兼容保证**：保持了对其他格式的兼容性
4. **📊 全面覆盖**：修复了所有相关的API函数
5. **✅ 功能恢复**：学期选择器现在可以正常工作

这次修复确保了成绩页面的完整功能，用户现在可以正常选择学期并查看对应的成绩信息。
