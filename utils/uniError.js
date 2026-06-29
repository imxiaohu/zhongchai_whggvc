/**
 * Uni Error 规范前端实现
 * 兼容 uni-app 错误标准
 * 参考: https://uniapp.dcloud.net.cn/tutorial/errors.html
 */

// 项目统一错误主题
export const ERR_SUBJECT = 'pingjiao'

// 错误码常量 (7位码: 第1-2位=一级类目, 第3-4位=二级类目, 第5-7位=具体错误类型)
// 一级类目: 10=认证/授权  20=业务逻辑  30=数据访问  40=学校服务器代理
//           50=参数校验    60=系统/网络  70=第三方服务
// 平台专有码: 跨端=6xx  Android=7xx  iOS=8xx  Web=9xx  Harmony=5xx
export const ErrCode = {
	SUCCESS: 0,

	// 认证/授权 (10xxx)
	UNAUTHORIZED: 10001,
	FORBIDDEN: 10002,
	SCHOOL_NOT_BOUND: 10003,
	SCHOOL_LOGIN_EXPIRY: 10004,
	TOKEN_EXPIRED: 10005,
	INVALID_TOKEN: 10006,

	// 业务逻辑 (20xxx)
	NOT_FOUND: 20001,
	OPERATION_FORBIDDEN: 20002,
	CONFLICT: 20003,
	STATE_NOT_ALLOWED: 20004,

	// 数据访问 (30xxx)
	DB_ERROR: 30001,
	PARSE_ERROR: 30002,

	// 学校服务器代理 (40xxx)
	SCHOOL_NO_RESPONSE: 40001,
	SCHOOL_ERROR: 40002,
	SCHOOL_LOGIN_FAILED: 40003,
	SCHOOL_MAINTENANCE: 40004,

	// 参数校验 (50xxx)
	INVALID_PARAMS: 50001,
	MISSING_PARAMS: 50002,

	// 系统/网络 (60xxx)
	INTERNAL_ERROR: 60001,
	SERVICE_UNAVAILABLE: 60002,
	TIMEOUT: 60003,
	NETWORK_ERROR: 60004
}

// 错误码 -> 友好提示映射
const ERR_CODE_MESSAGES = {
	[ErrCode.SUCCESS]: '操作成功',
	[ErrCode.UNAUTHORIZED]: '未授权，请重新登录',
	[ErrCode.FORBIDDEN]: '权限不足',
	[ErrCode.SCHOOL_NOT_BOUND]: '请先绑定学校账号',
	[ErrCode.SCHOOL_LOGIN_EXPIRY]: '学校账号会话已失效，请重新绑定',
	[ErrCode.TOKEN_EXPIRED]: '登录已过期，请重新登录',
	[ErrCode.INVALID_TOKEN]: '无效的登录凭证',

	[ErrCode.NOT_FOUND]: '资源不存在',
	[ErrCode.OPERATION_FORBIDDEN]: '该操作被禁止',
	[ErrCode.CONFLICT]: '数据已存在',
	[ErrCode.STATE_NOT_ALLOWED]: '当前状态不允许此操作',

	[ErrCode.DB_ERROR]: '数据库操作失败',
	[ErrCode.PARSE_ERROR]: '数据解析失败',

	[ErrCode.SCHOOL_NO_RESPONSE]: '学校服务器暂时无法访问，请稍后再试',
	[ErrCode.SCHOOL_ERROR]: '学校服务器返回错误',
	[ErrCode.SCHOOL_LOGIN_FAILED]: '学校账号登录失败',
	[ErrCode.SCHOOL_MAINTENANCE]: '学校服务器正在维护中，请稍后再试',

	[ErrCode.INVALID_PARAMS]: '参数错误',
	[ErrCode.MISSING_PARAMS]: '必填参数缺失',

	[ErrCode.INTERNAL_ERROR]: '服务器内部错误',
	[ErrCode.SERVICE_UNAVAILABLE]: '服务暂不可用',
	[ErrCode.TIMEOUT]: '请求超时，请检查网络连接',
	[ErrCode.NETWORK_ERROR]: '网络连接失败'
}

/**
 * SourceError - 源错误信息
 * 保存引起错误的底层原因（如三方SDK错误、网络错误等）
 */
export class SourceError {
	/**
	 * @param {string} message - 源错误描述信息
	 * @param {string} [subject] - 源错误模块名称
	 * @param {number} [code] - 源错误原始错误码
	 */
	constructor(message, subject = '', code = 0) {
		this.subject = subject
		this.code = code
		this.message = message
		this.cause = null
	}

	toJSON() {
		return {
			...(this.subject && { subject: this.subject }),
			...(this.code && { code: this.code }),
			message: this.message,
			...(this.cause && { cause: this.cause })
		}
	}
}

/**
 * UniAggregateError - 聚合源错误
 * 包含多个源错误，如多个三方SDK错误
 */
export class UniAggregateError {
	/**
	 * @param {SourceError[]} errors - 源错误数组
	 */
	constructor(errors = []) {
		this.errors = errors
	}

	toJSON() {
		return {
			errors: this.errors.map((e) => (e && typeof e.toJSON === 'function' ? e.toJSON() : e))
		}
	}
}

/**
 * UniError - Uni统一错误信息
 * 统一各平台（端）错误信息格式
 */
export class UniError {
	/**
	 * @param {string} [errSubject] - 统一错误主题（模块）名称
	 * @param {number} [errCode] - 统一错误码
	 * @param {string} [errMsg] - 统一错误描述信息
	 */
	constructor(errSubject = ERR_SUBJECT, errCode = ErrCode.INTERNAL_ERROR, errMsg = '未知错误') {
		this.errSubject = errSubject
		this.errCode = errCode
		this.errMsg = errMsg
		this.data = null
		this.cause = null
	}

	/**
	 * 设置附加数据
	 * @param {*} data - 错误时返回的附加数据
	 * @returns {UniError} this
	 */
	withData(data) {
		this.data = data
		return this
	}

	/**
	 * 设置错误原因
	 * @param {SourceError|UniAggregateError} cause - 源错误
	 * @returns {UniError} this
	 */
	withCause(cause) {
		this.cause = cause
		return this
	}

	/**
	 * 实现 Error 接口
	 */
	toString() {
		return `[${this.errSubject}] ${this.errMsg} (code=${this.errCode})`
	}

	toJSON() {
		return {
			errSubject: this.errSubject,
			errCode: this.errCode,
			errMsg: this.errMsg,
			...(this.data !== null && { data: this.data }),
			...(this.cause !== null && { cause: this.cause })
		}
	}
}

/**
 * 从后端响应构建 UniError
 * @param {Object} response - 后端响应对象
 * @returns {UniError|null}
 */
export function fromResponse(response) {
	if (!response) return null

	const errCode = response.errCode ?? response.code ?? 0
	const errMsg = response.errMsg ?? response.message ?? '未知错误'
	const errSubject = response.errSubject || ERR_SUBJECT

	if (response.success === true || errCode === 0) return null

	const error = new UniError(errSubject, errCode, errMsg)

	if (response.data !== undefined) {
		error.withData(response.data)
	}

	if (response.cause) {
		const cause = response.cause
		if (cause.errors) {
			const aggErrors = cause.errors.map((e) => new SourceError(e.message, e.subject, e.code))
			error.withCause(new UniAggregateError(aggErrors))
		} else {
			error.withCause(new SourceError(cause.message, cause.subject, cause.code))
		}
	}

	return error
}

/**
 * 获取错误码对应的友好提示
 * @param {number} errCode - 错误码
 * @param {string} [fallbackMsg] - 自定义提示（优先使用）
 * @returns {string}
 */
export function getErrorMessage(errCode, fallbackMsg) {
	if (fallbackMsg) return fallbackMsg
	return ERR_CODE_MESSAGES[errCode] || '请求失败，请稍后再试'
}

/**
 * 判断是否为认证/授权类错误
 * @param {number} errCode - 错误码
 * @returns {boolean}
 */
export function isAuthError(errCode) {
	return (
		errCode === ErrCode.UNAUTHORIZED ||
		errCode === ErrCode.TOKEN_EXPIRED ||
		errCode === ErrCode.INVALID_TOKEN ||
		errCode === ErrCode.SCHOOL_NOT_BOUND ||
		errCode === ErrCode.SCHOOL_LOGIN_EXPIRY
	)
}

/**
 * 判断是否为学校服务器错误
 * @param {number} errCode - 错误码
 * @returns {boolean}
 */
export function isSchoolServerError(errCode) {
	return (
		errCode === ErrCode.SCHOOL_NO_RESPONSE ||
		errCode === ErrCode.SCHOOL_ERROR ||
		errCode === ErrCode.SCHOOL_LOGIN_FAILED ||
		errCode === ErrCode.SCHOOL_MAINTENANCE
	)
}

/**
 * 判断是否为网络/超时错误
 * @param {number} errCode - 错误码
 * @returns {boolean}
 */
export function isNetworkError(errCode) {
	return (
		errCode === ErrCode.TIMEOUT ||
		errCode === ErrCode.NETWORK_ERROR ||
		errCode === ErrCode.SERVICE_UNAVAILABLE
	)
}

/**
 * 判断是否需要跳转登录
 * @param {number} errCode - 错误码
 * @returns {boolean}
 */
export function shouldRedirectToLogin(errCode) {
	return errCode === ErrCode.UNAUTHORIZED || errCode === ErrCode.TOKEN_EXPIRED || errCode === ErrCode.INVALID_TOKEN
}

/**
 * 创建标准错误对象（用于在业务代码中抛出）
 * @param {number} errCode - 错误码
 * @param {string} [message] - 自定义消息
 * @param {*} [data] - 附加数据
 * @returns {UniError}
 */
export function createError(errCode, message, data) {
	const errMsg = getErrorMessage(errCode, message)
	const error = new UniError(ERR_SUBJECT, errCode, errMsg)
	if (data !== undefined) {
		error.withData(data)
	}
	return error
}

export default {
	ERR_SUBJECT,
	ErrCode,
	SourceError,
	UniAggregateError,
	UniError,
	fromResponse,
	getErrorMessage,
	isAuthError,
	isSchoolServerError,
	isNetworkError,
	shouldRedirectToLogin,
	createError
}
