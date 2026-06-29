// request.js - 常量与配置

const PROD_BASE_URL = 'http://api.whggvc.imxiaohu.cn';
const LOCAL_BASE_URL = 'http://localhost:2333';

// 判断是否为本地开发环境
const isLocalDev = typeof window !== 'undefined' && 
	(window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1');

export const BASE_URL = isLocalDev ? LOCAL_BASE_URL : PROD_BASE_URL;

export const CONTENT_TYPES = {
	JSON: 'application/json',
	FORM: 'application/x-www-form-urlencoded',
	IMAGE: 'image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8',
	HTML: 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7'
};

export const HEADERS = {
	ACCEPT_LANGUAGE: 'zh-CN,zh;q=0.9',
	CACHE_CONTROL: 'no-cache',
	PRAGMA: 'no-cache'
};

export const HTTP_CACHE_PREFIX = 'http_cache_v1:';
export const DEFAULT_HTTP_CACHE_TTL_SEC = 600;

export const SIGN_MAP = {
	getCourseSchoolTimetable: 'E19D6243CB1945AB4F7202A1B00F77D5',
	getCourseTimeTableByDay: 'DF7B5CA48C599416CB7FAC6959E856E6',
	getCourseTimeTableByWeek: '69B6C6F65B34AED65E3EC616ECFFA935',
	mLogin: 'E19D6243CB1945AB4F7202A1B00F77D5',
	default: 'E19D6243CB1945AB4F7202A1B00F77D5'
};

export const NO_CACHE_PATHS = [
	'/scloud/init',
	'/scloud/login',
	'/scloud/validateCode',
	'/scloudoa/sys/mLogin',
	'/api/m/sys/mLogin',
	'/api/upload',
	'/api/payment'
];

export const MAX_RETRY_COUNT = 2;
export const RETRY_DELAY = 1000;
