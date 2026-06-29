import { request } from '../../utils/request.js';

/**
 * 获取考勤记录列表
 * @param {Object} params 查询参数
 * @param {number} params.pageNo 页码
 * @param {number} params.pageSize 每页条数
 * @param {string} params.nowWeek 当前周次，如"第17周"
 * @param {string} params.currentSemester 当前学期，如"2025-2026学年第二学期"
 * @param {string} params.courseName 课程名（搜索用）
 * @returns {Promise<Object>} 返回考勤记录列表
 */
export const getAttendanceList = (params = {}) => {
	const defaultParams = {
		pageNo: 1,
		pageSize: 15,
		nowWeek: '',
		currentSemester: '',
		courseName: ''
	};
	return request({
		url: '/scloudoa/classroomAttendance/list',
		method: 'GET',
		data: { ...defaultParams, ...params }
	});
};

/**
 * 获取选课记录列表
 * @param {Object} params 查询参数
 * @param {number} params.current 页码
 * @param {number} params.size 每页条数
 * @param {string} params.currentSemester 当前学期
 * @returns {Promise<Object>} 返回选课记录列表
 */
export const getOptionalCourses = (params = {}) => {
	const defaultParams = {
		current: 1,
		size: 15,
		currentSemester: ''
	};
	return request({
		url: '/scloudoa/course/tCourseOptionalStudent/getOptionalTeachingClass',
		method: 'GET',
		data: { ...defaultParams, ...params }
	});
};

/**
 * 获取学期可选教学班列表
 * @param {Object} params 查询参数
 * @param {number} params.current 页码
 * @param {number} params.size 每页条数，-1表示全部
 * @param {string} params.currentSemester 当前学期
 * @returns {Promise<Object>} 返回可选教学班列表
 */
export const getSemesterOptionalTeachingClass = (params = {}) => {
	const defaultParams = {
		current: 1,
		size: -1,
		currentSemester: ''
	};
	return request({
		url: '/scloudoa/course/tCourseOptionalStudent/getSemesterOptionalTeachingClass',
		method: 'GET',
		data: { ...defaultParams, ...params }
	});
};

/**
 * 获取请假记录列表
 * @param {Object} params 查询参数
 * @param {number} params.pageNo 页码
 * @param {number} params.pageSize 每页条数
 * @returns {Promise<Object>} 返回请假记录列表
 */
export const getLeaveList = (params = {}) => {
	const defaultParams = {
		pageNo: 1,
		pageSize: 10
	};
	return request({
		url: '/scloudoa/scs/leave/tStudentLeave/getTStudentLeave',
		method: 'GET',
		data: { ...defaultParams, ...params }
	});
};

/**
 * 获取问卷列表
 * @param {Object} params 查询参数
 * @param {number} params.current 页码
 * @param {number} params.size 每页条数
 * @returns {Promise<Object>} 返回问卷列表
 */
export const getSurveyList = (params = {}) => {
	const defaultParams = {
		current: 1,
		size: 10
	};
	return request({
		url: '/scloudoa/scs/survey/tOaSurvey/getTOaSurvey',
		method: 'GET',
		data: { ...defaultParams, ...params }
	});
};

/**
 * 获取问卷问题列表
 * @param {number} oaSurveyID 问卷ID
 * @returns {Promise<Array>} 返回问卷问题列表
 */
export const getSurveyQuestions = (oaSurveyID) => {
	return request({
		url: '/scloudoa/scs/survey/tOaSurveyQuestion/getTOaSurveyQuestion',
		method: 'GET',
		data: { oaSurveyID }
	});
};

/**
 * 获取问卷答案列表
 * @param {number} oaSurveyID 问卷ID
 * @returns {Promise<Array>} 返回问卷答案列表
 */
export const getSurveyAnswers = (oaSurveyID) => {
	return request({
		url: '/scloudoa/scs/survey/tOaSurveyQuestionAnswer/getTOaSurveyQuestionAnswer',
		method: 'GET',
		data: { oaSurveyID }
	});
};

/**
 * 获取用户详细信息
 * @param {string} id 用户ID
 * @returns {Promise<Object>} 返回用户详细信息
 */
export const getUserProfile = (id) => {
	return request({
		url: '/scloudoa/sys/user/queryById',
		method: 'GET',
		data: { id }
	});
};

/**
 * 获取课程计划列表
 * @param {Object} params 查询参数
 * @param {number} params.current 页码
 * @param {number} params.size 每页条数
 * @param {string} params.currentSemester 当前学期
 * @returns {Promise<Object>} 返回课程计划列表
 */
export const getCoursePlan = (params = {}) => {
	const defaultParams = {
		current: 1,
		size: 15,
		currentSemester: ''
	};
	return request({
		url: '/scloudoa/scs/course/tCourseScore/getCoursePlan',
		method: 'GET',
		data: { ...defaultParams, ...params }
	});
};

/**
 * 获取学期列表（用于课程计划筛选）
 * @returns {Promise<Array>} 返回学期列表
 */
export const getSemesterList = async () => {
	const res = await request({
		url: '/api/m/scs/course/tCourseScore/getSemester',
		method: 'GET'
	});

	let semesterData = [];
	if (res && res.result && res.result.result && Array.isArray(res.result.result)) {
		semesterData = res.result.result;
	} else if (res && Array.isArray(res.result)) {
		semesterData = res.result;
	}

	return semesterData.map(item => {
		const name = item.currentSemester || item.semesterName || item.semester || item.name || ''
		return { name, value: name, currentSemester: name }
	})
};

/**
 * 获取当前学期信息
 * @returns {Promise<string>} 返回当前学期名称
 */
export const getCurrentSemester = async () => {
	try {
		const res = await request({
			url: '/scloudoa/userQuery/tSysUser/getCourseSchoolTimetable',
			method: 'GET'
		})
		return res?.result?.currentSemester || ''
	} catch (e) {
		return ''
	}
};

/**
 * 获取学生完整档案信息（合并本地用户数据 + 学校扩展字段）
 * @returns {Promise<Object>} 返回学生完整档案
 */
export const getUserArchive = () => {
	return request({
		url: '/api/user/archive',
		method: 'GET'
	});
};

// ========== 报修管理 ==========

/**
 * 获取报修记录列表
 * @param {Object} params 查询参数
 * @param {number} params.pageNo 页码
 * @param {number} params.pageSize 每页条数
 * @returns {Promise<Object>} 返回报修记录列表
 */
export const getRepairList = (params = {}) => {
	const defaultParams = {
		pageNo: 1,
		pageSize: 15
	};
	return request({
		url: '/scloudoa/repairReport/tLogisticsMaintenanceOrder/list',
		method: 'GET',
		data: { ...defaultParams, ...params }
	});
};

/**
 * 获取报修单详情
 * @param {number} id 报修单ID
 * @returns {Promise<Object>} 返回报修单详情
 */
export const getRepairDetail = (id) => {
	return request({
		url: '/scloudoa/repairReport/tLogisticsMaintenanceOrder/getStudent',
		method: 'GET',
		data: { id }
	});
};

/**
 * 获取报修类型（楼栋列表）
 * @returns {Promise<Array>} 返回报修类型列表
 */
export const getRepairTypes = () => {
	return request({
		url: '/scloudoa/repairReport/tLogisticsMaintenanceOrder/getLogisticsMaintenanceOrderType',
		method: 'GET'
	});
};

/**
 * 获取文件服务器地址
 * @returns {Promise<string>} 返回文件服务器地址
 */
export const getFileServerUrl = () => {
	return request({
		url: '/scloudoa/sys/common/getFileServerUrl',
		method: 'GET'
	});
};

/**
 * 提交报修单
 * @param {Object} data 报修单数据
 * @returns {Promise<Object>} 返回提交结果
 */
export const submitRepair = (data) => {
	return request({
		url: '/scloudoa/repairReport/tLogisticsMaintenanceOrder/add',
		method: 'POST',
		data
	});
};

/**
 * 获取报修单详情（通过ID）
 * @param {number} id 报修单ID
 * @returns {Promise<Object>} 返回报修单详情
 */
export const getRepairOrderDetail = (id) => {
	return request({
		url: '/scloudoa/repairReport/tLogisticsMaintenanceOrder/queryById',
		method: 'GET',
		data: { id }
	});
};

// ========== 银行卡 ==========

/**
 * 获取我的银行卡信息（同时返回学生档案信息用于补充个人资料）
 * @returns {Promise<Object>} 返回银行卡和学生档案信息
 */
export const getStudentBank = () => {
	return request({
		url: '/scloudoa/studentBank/getStudentBank',
		method: 'GET'
	});
};

/**
 * 编辑银行卡信息
 * @param {Object} data 银行卡数据
 * @param {string} data.bankcardnumber 银行卡号
 * @param {string} data.phoneNumber 手机号
 * @param {string} data.bankname 开户银行
 * @param {string} data.bankbranch 开户省/市
 * @param {string} data.bankprovincecity 开户市/区
 * @param {string} data.banksubbranch 开户支行
 * @returns {Promise<Object>} 返回编辑结果
 */
export const editStudentBank = (data) => {
	return request({
		url: '/scloudoa/studentBank/edit',
		method: 'PUT',
		data
	});
};

// ========== PC端专属接口（由后端透明处理授权，前端仅调用数据接口）==========

/**
 * 获取我的老师列表（后端自动处理PC会话授权）
 * @param {Object} params 查询参数
 * @param {number} params.pageSize 每页条数
 * @param {number} params.pageNum 页码
 * @param {string} params.gradeName 年级筛选
 * @returns {Promise<Object>}
 */
export const pcGetTeachers = (params = {}) => {
	return request({
		url: '/api/pc/teachers',
		method: 'GET',
		data: {
			pageSize: 15,
			pageNum: 1,
			gradeName: '',
			...params
		},
		timeout: 30000
	});
};

/**
 * 获取学生信息（后端自动处理PC会话授权）
 * @returns {Promise<Object>}
 */
export const pcGetStudentInfo = () => {
	return request({
		url: '/api/pc/student-info',
		method: 'GET',
		timeout: 30000
	});
};

/**
 * 提交档案编辑（后端自动处理PC会话授权）
 * @param {Object} data 编辑数据
 * @returns {Promise<Object>}
 */
export const pcSubmitArchiveEdit = (data) => {
	return request({
		url: '/api/pc/archive-edit',
		method: 'POST',
		data,
		timeout: 30000
	});
};

/**
 * 获取亲属关系类型列表（档案编辑用）
 * @returns {Promise<Array>}
 */
export const pcGetRelationTypes = () => {
	return request({
		url: '/scloud/student/base/relation',
		method: 'GET',
		timeout: 30000
	});
};

/**
 * 获取省份列表（档案编辑用）
 * @returns {Promise<Array>}
 */
export const pcGetProvinceList = () => {
	return request({
		url: '/scloud/student/base/getProvince',
		method: 'GET',
		timeout: 30000
	});
};

// ========== PC端专属业务接口（缺课统计/补考查询/违纪查询）==========

/**
 * 获取缺课统计列表（后端自动处理PC会话授权）
 * @param {Object} params 查询参数
 * @param {number} params.pageSize 每页条数
 * @param {number} params.pageNum 页码
 * @param {string} params.gradeName 年级筛选
 * @param {string} params.studySystem 学制
 * @param {string} params.semester 学期
 * @returns {Promise<Object>}
 */
export const pcGetMissClassList = (params = {}) => {
	return request({
		url: '/api/pc/miss-class',
		method: 'GET',
		data: {
			pageSize: 15,
			pageNum: 1,
			gradeName: '',
			studySystem: '',
			semester: '',
			...params
		},
		timeout: 30000
	});
};

/**
 * 获取补考查询列表（后端自动处理PC会话授权）
 * @param {Object} params 查询参数
 * @param {number} params.pageSize 每页条数
 * @param {number} params.pageNum 页码
 * @param {string} params.semester 学期
 * @returns {Promise<Object>}
 */
export const pcGetMakeupExamList = (params = {}) => {
	return request({
		url: '/api/pc/makeup-exam',
		method: 'GET',
		data: {
			pageSize: 15,
			pageNum: 1,
			semester: '',
			...params
		},
		timeout: 30000
	});
};

/**
 * 获取违纪查询列表（后端自动处理PC会话授权）
 * @param {Object} params 查询参数
 * @param {number} params.pageSize 每页条数
 * @param {number} params.pageNum 页码
 * @param {string} params.learnerdisciplinarytypeid 违纪类型ID
 * @param {string} params.learnerdisciplinarylevelid 违纪级别ID
 * @param {string} params.createTimeRange 时间范围
 * @returns {Promise<Object>}
 */
export const pcGetDisciplinaryList = (params = {}) => {
	return request({
		url: '/api/pc/disciplinary',
		method: 'GET',
		data: {
			pageSize: 15,
			pageNum: 1,
			learnerdisciplinarytypeid: '',
			learnerdisciplinarylevelid: '',
			createTimeRange: '',
			...params
		},
		timeout: 30000
	});
};

/**
 * 获取违纪类型列表（后端自动处理PC会话授权）
 * @returns {Promise<Array>}
 */
export const pcGetDisciplinaryTypes = () => {
	return request({
		url: '/api/pc/disciplinary/types',
		method: 'GET',
		timeout: 30000
	});
};

/**
 * 获取违纪级别列表（后端自动处理PC会话授权）
 * @returns {Promise<Array>}
 */
export const pcGetDisciplinaryLevels = () => {
	return request({
		url: '/api/pc/disciplinary/levels',
		method: 'GET',
		timeout: 30000
	});
};

// ========== 实习相关 ==========

/**
 * 获取实习要求列表
 * @param {Object} params 查询参数
 * @param {number} params.pageNo 页码
 * @param {number} params.pageSize 每页条数
 * @returns {Promise<Object>} 返回实习要求列表
 */
export const getInternshipRequirementsList = (params = {}) => {
	const defaultParams = {
		pageNo: 1,
		pageSize: 15
	};
	return request({
		url: '/scloudoa/jobInternship/internshipRequirementsInquiry/list',
		method: 'GET',
		data: { ...defaultParams, ...params }
	});
};

/**
 * 获取岗位职业分类
 * @returns {Promise<Array>} 返回岗位职业分类列表
 */
export const getInternshipJobClassification = () => {
	return request({
		url: '/scloudoa/jobInternship/internshipPositionInquiry/getJobClassification',
		method: 'GET'
	});
};

/**
 * 获取行业分类
 * @returns {Promise<Array>} 返回行业分类列表
 */
export const getInternshipIndustryClassification = () => {
	return request({
		url: '/scloudoa/jobInternship/internshipPositionInquiry/getIndustryClassification',
		method: 'GET'
	});
};

/**
 * 获取实习岗位列表
 * @param {Object} params 查询参数
 * @param {number} params.pageNo 页码
 * @param {number} params.pageSize 每页条数
 * @returns {Promise<Object>} 返回实习岗位列表
 */
export const getInternshipPositionList = (params = {}) => {
	const defaultParams = {
		pageNo: 1,
		pageSize: 15
	};
	return request({
		url: '/scloudoa/jobInternship/internshipPositionInquiry/list',
		method: 'GET',
		data: { ...defaultParams, ...params }
	});
};

/**
 * 获取实习申请列表
 * @param {Object} params 查询参数
 * @param {number} params.pageNo 页码
 * @param {number} params.pageSize 每页条数
 * @returns {Promise<Object>} 返回实习申请列表
 */
export const getInternshipApplicationList = (params = {}) => {
	const defaultParams = {
		pageNo: 1,
		pageSize: 15
	};
	return request({
		url: '/scloudoa/jobInternship/internshipPositionApplication/list',
		method: 'GET',
		data: { ...defaultParams, ...params }
	});
};

/**
 * 获取实习签到计划
 * @returns {Promise<Object>} 返回实习签到计划
 */
export const getInternshipSignInPlan = () => {
	return request({
		url: '/scloudoa/jobInternship/internshipSignIn/getPlan',
		method: 'GET'
	});
};

/**
 * 获取我的实习计划列表
 * @param {Object} params 查询参数
 * @param {number} params.pageNo 页码
 * @param {number} params.pageSize 每页条数
 * @returns {Promise<Object>} 返回我的实习计划列表
 */
export const getMyInternshipPlanList = (params = {}) => {
	const defaultParams = {
		pageNo: 1,
		pageSize: 15
	};
	return request({
		url: '/scloudoa/jobInternship/myInternshipPlan/list',
		method: 'GET',
		data: { ...defaultParams, ...params }
	});
};

/**
 * 获取我的实习总结列表
 * @param {Object} params 查询参数
 * @param {number} params.pageNo 页码
 * @param {number} params.pageSize 每页条数
 * @returns {Promise<Object>} 返回我的实习总结列表
 */
export const getMyInternshipSummaryList = (params = {}) => {
	const defaultParams = {
		pageNo: 1,
		pageSize: 15
	};
	return request({
		url: '/scloudoa/jobInternship/myInternshipSummary/list',
		method: 'GET',
		data: { ...defaultParams, ...params }
	});
};

/**
 * 获取实习鉴定表列表
 * @param {Object} params 查询参数
 * @param {number} params.pageNo 页码
 * @param {number} params.pageSize 每页条数
 * @returns {Promise<Object>} 返回实习鉴定表列表
 */
export const getInternshipAppraisalFormList = (params = {}) => {
	const defaultParams = {
		pageNo: 1,
		pageSize: 15
	};
	return request({
		url: '/scloudoa/jobInternship/internshipAppraisalForm/list',
		method: 'GET',
		data: { ...defaultParams, ...params }
	});
};

/**
 * 获取实习成绩查询列表
 * @param {Object} params 查询参数
 * @param {number} params.pageNo 页码
 * @param {number} params.pageSize 每页条数
 * @returns {Promise<Object>} 返回实习成绩查询列表
 */
export const getInternshipScoreInquiryList = (params = {}) => {
	const defaultParams = {
		pageNo: 1,
		pageSize: 15
	};
	return request({
		url: '/scloudoa/jobInternship/internshipScoreInquiry/list',
		method: 'GET',
		data: { ...defaultParams, ...params }
	});
};

// ============================================================
// PC 端登录授权 API
// ============================================================

/**
 * 全自动 PC 端登录（后端自动完成 OCR 识别 + 提交）
 * @returns {Promise<Object>}
 */
export const pcAutoLogin = () => {
	return request({
		url: '/api/pc/auto-login',
		method: 'POST',
		timeout: 60000
	});
};

/**
 * 提交手动输入的验证码
 * @param {string} captcha - 用户输入的4位验证码
 * @returns {Promise<Object>}
 */
export const pcLoginSubmit = (captcha) => {
	return request({
		url: '/api/pc/login/submit',
		method: 'POST',
		data: { captcha },
		timeout: 30000
	});
};

/**
 * 查询 PC 会话状态
 * @returns {Promise<Object>}
 */
export const pcGetSessionStatus = () => {
	return request({
		url: '/api/pc/session-status',
		method: 'GET',
		timeout: 10000
	});
};

/**
 * 登出 PC 端会话
 * @returns {Promise<Object>}
 */
export const pcLogout = () => {
	return request({
		url: '/api/pc/logout',
		method: 'POST',
		timeout: 10000
	});
};

/**
 * 获取存储的会话凭证（用于恢复会话）
 * @returns {Promise<Object>}
 */
export const pcGetSessionCredentials = () => {
	return request({
		url: '/api/pc/session-credentials',
		method: 'GET',
		timeout: 10000
	});
};

/**
 * 将前端存储的会话凭证同步到后端
 * @param {Object} credentials - { sessionId, expireAt }
 * @returns {Promise<Object>}
 */
export const pcSetSessionCredentials = (credentials) => {
	return request({
		url: '/api/pc/session-credentials',
		method: 'POST',
		data: credentials,
		timeout: 10000
	});
};
