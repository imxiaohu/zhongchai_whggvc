// 权限和角色相关的工具函数

// 用户角色定义
export const USER_ROLES = {
	VISITOR: 'visitor',
	USER: 'user',
	STUDENT: 'student',
	ADMIN: 'admin'
};

// 权限定义
export const PERMISSIONS = {
	// 评教相关
	CAN_EVALUATE: 'can_evaluate',
	CAN_BATCH_EVALUATE: 'can_batch_evaluate',
	CAN_VIEW_EVALUATION_RESULT: 'can_view_evaluation_result',
	
	// 社区相关
	CAN_CREATE_POST: 'can_create_post',
	CAN_CREATE_CLUB: 'can_create_club',
	CAN_MANAGE_CLUB: 'can_manage_club',
	CAN_REPORT: 'can_report',
	
	// 账号相关
	CAN_BIND_SCHOOL_ACCOUNT: 'can_bind_school_account',
	CAN_UNBIND_ACCOUNT: 'can_unbind_account',
	CAN_VIEW_SYNC_STATUS: 'can_view_sync_status',
	
	// 管理功能
	CAN_MANAGE_USERS: 'can_manage_users',
	CAN_VIEW_ANALYTICS: 'can_view_analytics',
	CAN_SYSTEM_SETTINGS: 'can_system_settings'
};

// 角色权限映射
export const ROLE_PERMISSIONS = {
	[USER_ROLES.VISITOR]: [
		PERMISSIONS.CAN_VIEW_EVALUATION_RESULT
	],
	[USER_ROLES.USER]: [
		PERMISSIONS.CAN_VIEW_EVALUATION_RESULT,
		PERMISSIONS.CAN_CREATE_POST,
		PERMISSIONS.CAN_REPORT
	],
	[USER_ROLES.STUDENT]: [
		PERMISSIONS.CAN_VIEW_EVALUATION_RESULT,
		PERMISSIONS.CAN_EVALUATE,
		PERMISSIONS.CAN_BATCH_EVALUATE,
		PERMISSIONS.CAN_CREATE_POST,
		PERMISSIONS.CAN_CREATE_CLUB,
		PERMISSIONS.CAN_REPORT,
		PERMISSIONS.CAN_BIND_SCHOOL_ACCOUNT,
		PERMISSIONS.CAN_VIEW_SYNC_STATUS
	],
	[USER_ROLES.ADMIN]: [
		PERMISSIONS.CAN_VIEW_EVALUATION_RESULT,
		PERMISSIONS.CAN_EVALUATE,
		PERMISSIONS.CAN_BATCH_EVALUATE,
		PERMISSIONS.CAN_CREATE_POST,
		PERMISSIONS.CAN_CREATE_CLUB,
		PERMISSIONS.CAN_MANAGE_CLUB,
		PERMISSIONS.CAN_REPORT,
		PERMISSIONS.CAN_BIND_SCHOOL_ACCOUNT,
		PERMISSIONS.CAN_UNBIND_ACCOUNT,
		PERMISSIONS.CAN_VIEW_SYNC_STATUS,
		PERMISSIONS.CAN_MANAGE_USERS,
		PERMISSIONS.CAN_VIEW_ANALYTICS,
		PERMISSIONS.CAN_SYSTEM_SETTINGS
	]
};

// 根据登录类型获取角色
export function getRoleByLoginType(loginType, hasSchoolAccount) {
	if (loginType === 'wechat') {
		return hasSchoolAccount ? USER_ROLES.STUDENT : USER_ROLES.USER;
	}
	if (loginType === 'mobile' || loginType === 'school') {
		return USER_ROLES.STUDENT;
	}
	return USER_ROLES.VISITOR;
}

// 检查用户是否具有某个权限
export function hasPermission(loginType, hasSchoolAccount, permission) {
	const role = getRoleByLoginType(loginType, hasSchoolAccount);
	const permissions = ROLE_PERMISSIONS[role] || [];
	return permissions.includes(permission);
}

// 检查是否已绑定学校账号
export function isSchoolAccountBound(hasBindSchoolAccount) {
	return !!hasBindSchoolAccount;
}

// 检查是否已登录
export function isLoggedIn(token) {
	return !!token;
}

// 检查是否是管理员
export function isAdmin(loginType) {
	return loginType === 'admin';
}

// 检查功能是否可用
export function isFeatureEnabled(feature, userRole) {
	const enabledFeatures = ROLE_PERMISSIONS[userRole] || [];
	return enabledFeatures.includes(feature);
}

// 获取用户功能列表
export function getUserFeatures(loginType, hasSchoolAccount) {
	const role = getRoleByLoginType(loginType, hasSchoolAccount);
	return ROLE_PERMISSIONS[role] || [];
}

// 检查是否可以执行某操作
export function canPerformAction(action, userInfo) {
	const loginType = uni.getStorageSync('loginType');
	const hasBindSchoolAccount = uni.getStorageSync('hasBindSchoolAccount');

	switch (action) {
		case 'evaluate':
			return hasPermission(loginType, hasBindSchoolAccount, PERMISSIONS.CAN_EVALUATE);
		
		case 'createPost':
			return hasPermission(loginType, hasBindSchoolAccount, PERMISSIONS.CAN_CREATE_POST);
		
		case 'createClub':
			return hasPermission(loginType, hasBindSchoolAccount, PERMISSIONS.CAN_CREATE_CLUB);
		
		case 'manageClub':
			return hasPermission(loginType, hasBindSchoolAccount, PERMISSIONS.CAN_MANAGE_CLUB);
		
		case 'report':
			return hasPermission(loginType, hasBindSchoolAccount, PERMISSIONS.CAN_REPORT);
		
		case 'bindSchool':
			return hasPermission(loginType, hasBindSchoolAccount, PERMISSIONS.CAN_BIND_SCHOOL_ACCOUNT);
		
		case 'unbindAccount':
			return hasPermission(loginType, hasBindSchoolAccount, PERMISSIONS.CAN_UNBIND_ACCOUNT);
		
		case 'viewSyncStatus':
			return hasPermission(loginType, hasBindSchoolAccount, PERMISSIONS.CAN_VIEW_SYNC_STATUS);
		
		case 'manageUsers':
			return hasPermission(loginType, hasBindSchoolAccount, PERMISSIONS.CAN_MANAGE_USERS);
		
		default:
			return false;
	}
}

// 获取权限检查的错误信息
export function getPermissionErrorMessage(action) {
	const messages = {
		evaluate: '请先绑定学校账号进行评教',
		createPost: '请登录后发布帖子',
		createClub: '请先绑定学校账号创建社团',
		manageClub: '您没有管理社团的权限',
		report: '请登录后使用举报功能',
		bindSchool: '您已经绑定了学校账号',
		unbindAccount: '您没有权限解除绑定',
		viewSyncStatus: '请先绑定学校账号查看同步状态',
		manageUsers: '您没有管理权限'
	};
	return messages[action] || '您没有执行此操作的权限';
}

// 批量权限检查
export function checkMultiplePermissions(permissions, loginType, hasSchoolAccount) {
	return permissions.map(permission => ({
		permission,
		has: hasPermission(loginType, hasSchoolAccount, permission)
	}));
}
