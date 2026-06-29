// 用户设置页面的常量定义

// 主题定义
export const THEMES = [
	{ id: 'light', name: '浅色模式', icon: 'sun' },
	{ id: 'dark', name: '深色模式', icon: 'moon' },
	{ id: 'system', name: '跟随系统', icon: 'monitor' }
];

// 字体大小选项
export const FONT_SIZE_OPTIONS = [
	{ id: 'small', name: '小', value: 14 },
	{ id: 'medium', name: '中', value: 16 },
	{ id: 'large', name: '大', value: 18 },
	{ id: 'xlarge', name: '特大', value: 20 }
];

// 学院映射
export const FACULTY_MAP = {
	'1': '信息工程学院',
	'2': '机电工程学院',
	'3': '经济管理学院'
};

// 同步状态映射
export const SYNC_STATUS_MAP = {
	'idle': '空闲',
	'syncing': '同步中',
	'success': '同步成功',
	'failed': '同步失败'
};

// 关于页面信息
export const ABOUT_INFO = {
	appName: '众柴智慧校园',
	developer: 'xiaohu',
	description: '众柴智慧校园是一款专为学生设计的评教系统'
};

// 默认同步设置
export const DEFAULT_SYNC_SETTINGS = {
	enabled: false,
	frequency: 'daily',
	timeRange: '08:30-22:20',
	autoRetryEnabled: true,
	maxRetryCount: 3
};

// 默认同步状态
export const DEFAULT_SYNC_STATUS = {
	syncStatus: 'idle',
	lastSyncAt: null,
	nextSyncAt: null,
	lastSyncMessage: '',
	coursesCount: 0
};

// 默认设置
export const DEFAULT_SETTINGS = {
	theme: 'system',
	fontSize: 'medium',
	language: 'zh-CN',
	notifications: {
		enabled: true,
		scoreUpdate: true,
		evaluation: true
	}
};

// 应用平台信息
export const APP_PLATFORMS = {
	'app-plus': {
		appId: () => plus.runtime.appid,
		version: () => plus.runtime.version,
		versionCode: () => plus.runtime.versionCode,
		platform: () => plus.os.name
	},
	'h5': {
		appId: '__UNI__84DD641',
		version: '1.0.0',
		versionCode: '100',
		platform: 'H5'
	},
	'mp-weixin': {
		appId: 'YOUR_WX_APPID2_REMOVED',
		version: '1.0.0',
		versionCode: '100',
		platform: 'WeChat'
	}
};

// 获取应用信息
export function getAppInfo() {
	// #ifdef APP-PLUS
	return {
		appId: plus.runtime.appid,
		version: plus.runtime.version,
		versionCode: plus.runtime.versionCode,
		platform: plus.os.name
	};
	// #endif

	// #ifdef H5
	return {
		appId: '__UNI__84DD641',
		version: '1.0.0',
		versionCode: '100',
		platform: 'H5'
	};
	// #endif

	// #ifdef MP-WEIXIN
	return {
		appId: 'YOUR_WX_APPID2_REMOVED',
		version: '1.0.0',
		versionCode: '100',
		platform: 'WeChat'
	};
	// #endif

	return {
		appId: '__UNI__84DD641',
		version: '1.0.0',
		versionCode: '100',
		platform: 'Unknown'
	};
}

// 获取学院名称
export function getFacultyName(facultyId) {
	if (!facultyId) return '';
	return FACULTY_MAP[String(facultyId)] || `${facultyId}`;
}

// 获取同步状态文本
export function getSyncStatusText(syncStatus, syncSettings, isBound) {
	if (!isBound) {
		return '需要绑定账号';
	}

	if (!syncSettings.enabled) {
		return '已禁用';
	}

	const statusText = SYNC_STATUS_MAP[syncStatus.syncStatus] || '未知';

	if (syncStatus.syncStatus === 'failed' && syncStatus.lastSyncMessage) {
		const errorMsg = syncStatus.lastSyncMessage;
		if (errorMsg.includes('学校服务器连接失败') || errorMsg.includes('School server connection failed')) {
			return '同步失败 (服务器连接失败)';
		} else if (errorMsg.includes('未绑定学校账号') || errorMsg.includes('Need to bind school account')) {
			return '同步失败 (需要绑定账号)';
		} else if (errorMsg.includes('获取当前学期失败') || errorMsg.includes('Failed to get current semester')) {
			return '同步失败 (数据解析错误)';
		}
		return '同步失败 (错误详情)';
	}

	if (syncStatus.lastSyncAt) {
		const lastSync = new Date(syncStatus.lastSyncAt);
		const now = new Date();
		const diffHours = Math.floor((now - lastSync) / (1000 * 60 * 60));

		if (diffHours < 1) {
			return `${statusText} (刚刚)`;
		} else if (diffHours < 24) {
			return `${statusText} (${diffHours}小时前)`;
		} else {
			const diffDays = Math.floor(diffHours / 24);
			return `${statusText} (${diffDays}天前)`;
		}
	}

	return statusText;
}

// 计算缓存大小
export function calculateCacheSize() {
	try {
		const storageInfo = uni.getStorageInfoSync();
		return (storageInfo.currentSize / 1024).toFixed(1);
	} catch (error) {
		console.error('获取存储信息失败:', error);
		return '0.0';
	}
}

// 保留的数据键
export const PRESERVED_DATA_KEYS = [
	'saved_username',
	'saved_password',
	'remember_password',
	'clientId',
	'userInfo',
	'loginType',
	'hasBindSchoolAccount',
	'token',
	'accessToken',
	'refreshToken',
	'wechatOpenId',
	'wechatUnionId',
	'schoolInfo',
	'currentSemester'
];

// 序列化用户详情
export function serializeUserDetail(userDetail) {
	let content = '';
	if (userDetail.realname) {
		content += `${'真实姓名'}：${userDetail.realname}\n`;
	}
	if (userDetail.nickname) {
		content += `${'昵称'}：${userDetail.nickname}\n`;
	}
	if (userDetail.username) {
		content += `${'学号'}：${userDetail.username}\n`;
	}
	if (userDetail.college) {
		content += `${'学院'}：${userDetail.college}\n`;
	}
	if (userDetail.className) {
		content += `${'班级'}：${userDetail.className}\n`;
	}
	if (userDetail.currentSemester) {
		content += `${'当前学期'}：${userDetail.currentSemester}\n`;
	}
	if (userDetail.loginType) {
		content += `${'登录类型'}：${userDetail.loginType}`;
	}
	return content;
}

// 获取默认用户信息
export function getDefaultUserInfo() {
	return {
		username: '未登录',
		studentId: '未登录',
		isBound: false,
		userTags: [],
		userDetail: {}
	};
}

// 更新用户信息显示
export function updateUserInfoDisplay(userInfo, loginType, hasBindSchoolAccount) {
	const info = {};

	if (loginType === 'mobile' || loginType === 'school') {
		info.username = userInfo.realname || userInfo.name || '学校用户';
		info.studentId = userInfo.username || userInfo.studentId || '';
		info.isBound = true;

		const tags = [];
		const college = userInfo.college || userInfo.facultyName || '';
		if (college) tags.push(college);
		if (userInfo.className) tags.push(userInfo.className);
		info.userTags = tags;

		info.userDetail = {
			realname: userInfo.realname,
			username: userInfo.username,
			college: college,
			className: userInfo.className,
			currentSemester: userInfo.currentSemester,
			loginType: '学校用户'
		};
	} else if (loginType === 'wechat') {
		if (hasBindSchoolAccount && userInfo.realname) {
			info.username = userInfo.realname || userInfo.nickname || '微信用户';
			info.studentId = userInfo.username || '';
			info.isBound = true;

			const tags = ['微信绑定'];
			const college = getFacultyName(userInfo.facultyId) || userInfo.college;
			if (college) tags.push(college);
			if (userInfo.className) tags.push(userInfo.className);
			info.userTags = tags;

			info.userDetail = {
				realname: userInfo.realname,
				nickname: userInfo.nickname,
				username: userInfo.username,
				college: college,
				className: userInfo.className,
				currentSemester: userInfo.currentSemester,
				loginType: '微信和学校'
			};
		} else {
			info.username = userInfo.nickname || '微信用户';
			info.studentId = '未绑定学校账号';
			info.isBound = false;
			info.userTags = ['微信用户'];
			info.userDetail = {
				nickname: userInfo.nickname,
				loginType: '微信用户'
			};
		}
	} else {
		info.username = userInfo.realname || userInfo.name || userInfo.nickname || '用户';
		info.studentId = userInfo.username || userInfo.studentId || '';
		info.isBound = !!info.studentId && info.studentId !== '未绑定学校账号';
		info.userTags = ['普通用户'];
		info.userDetail = {
			name: info.username,
			studentId: info.studentId,
			loginType: '其他'
		};
	}

	return info;
}

// 获取用户名首字母
export function getUserInitial(username) {
	if (!username) return '游';
	return username.charAt(0);
}
