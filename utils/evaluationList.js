// 评教列表页面的常量定义
export const EVALUATION_CONSTANTS = {
	REFRESH_INTERVAL: 3000, // 3秒刷新间隔
	AUTO_REFRESH_INTERVAL: 60000, // 1分钟自动刷新
	DEFAULT_SCORE: 5, // 默认满分
	DEFAULT_NORMS: [
		{ id: 1, name: '教学态度', score: 0 },
		{ id: 2, name: '教学内容', score: 0 },
		{ id: 3, name: '教学方法', score: 0 },
		{ id: 4, name: '教学效果', score: 0 },
		{ id: 5, name: '教材使用', score: 0 }
	]
};

// 预览数据生成
export function generatePreviewEvaluationList() {
	return [
		{
			id: 'preview-1',
			name: '高等数学A',
			teacherName: '张教授',
			className: '计算机科学与技术2023-1班',
			semester: '2024-2025学年第二学期',
			startTime: '2024-09-01',
			endTime: '2025-01-15',
			evaluated: false,
			isPreview: true
		},
		{
			id: 'preview-2',
			name: '大学英语',
			teacherName: '李老师',
			className: '计算机科学与技术2023-1班',
			semester: '2024-2025学年第二学期',
			startTime: '2024-09-01',
			endTime: '2025-01-15',
			evaluated: false,
			isPreview: true
		},
		{
			id: 'preview-3',
			name: '数据结构与算法',
			teacherName: '王教授',
			className: '计算机科学与技术2023-1班',
			semester: '2024-2025学年第二学期',
			startTime: '2024-09-01',
			endTime: '2025-01-15',
			evaluated: true,
			isPreview: true
		},
		{
			id: 'preview-4',
			name: '计算机网络',
			teacherName: '陈老师',
			className: '计算机科学与技术2023-1班',
			semester: '2024-2025学年第二学期',
			startTime: '2024-09-01',
			endTime: '2025-01-15',
			evaluated: false,
			isPreview: true
		},
		{
			id: 'preview-5',
			name: '操作系统',
			teacherName: '刘教授',
			className: '计算机科学与技术2023-1班',
			semester: '2024-2025学年第二学期',
			startTime: '2024-09-01',
			endTime: '2025-01-15',
			evaluated: true,
			isPreview: true
		}
	];
}

// 时间范围格式化
export function formatTimeRange(startTime, endTime) {
	if (!startTime || !endTime) return '未设置';

	const formatDate = (dateStr) => {
		if (!dateStr) return '';
		return dateStr.split(' ')[0];
	};

	return `${formatDate(startTime)} ~ ${formatDate(endTime)}`;
}

// 统计信息计算
export function calculateStatistics(evaluationList) {
	const total = evaluationList.length;
	const completed = evaluationList.filter(item => item.evaluated).length;
	const pending = total - completed;

	return { total, completed, pending };
}

// 评教状态文本获取
export function getEvaluationStatusText(item, hasCache) {
	if (hasCache) return '草稿';
	return item.evaluated ? '已完成' : '待评教';
}

// 评教状态样式类获取
export function getEvaluationStatusClass(item, hasCache) {
	if (hasCache) return { tag: 'draft', dot: 'draft' };
	return item.evaluated 
		? { tag: 'completed', dot: 'completed' } 
		: { tag: 'pending', dot: 'pending' };
}

// 处理评教列表数据
export function processEvaluationListData(listData) {
	if (!listData || !Array.isArray(listData)) {
		console.warn('评教列表数据格式异常:', listData);
		return [];
	}

	return listData.map(item => {
		const mappedId = item.id || item.courseTeachingEvaluationStudentConfigID || item.courseTeachingClassTaskID;
		return {
			...item,
			id: mappedId,
			title: item.courseName || item.name || '',
			teacherName: item.teacherName || '',
			startTime: item.startTime,
			endTime: item.endTime,
			evaluated: item.evaluated || false
		};
	});
}

// 处理评教详情数据
export function processEvaluationDetail(detail) {
	if (Array.isArray(detail)) {
		return {
			normList: detail.length > 0 ? detail.map(norm => ({ ...norm, score: 0 })) : [...EVALUATION_CONSTANTS.DEFAULT_NORMS],
			comment: '',
			isCompleted: false
		};
	}

	if (detail && detail.success && Array.isArray(detail.result)) {
		return {
			normList: detail.result.length > 0 ? detail.result.map(norm => ({ ...norm, score: 0 })) : [...EVALUATION_CONSTANTS.DEFAULT_NORMS],
			comment: '',
			isCompleted: false
		};
	}

	if (detail && Array.isArray(detail.courseTeachingEvaluationQuestionsVOS)) {
		return {
			normList: detail.courseTeachingEvaluationQuestionsVOS.map(norm => ({ ...norm, score: 0 })),
			comment: '',
			isCompleted: false
		};
	}

	if (detail && detail.data && Array.isArray(detail.data.normList)) {
		return {
			normList: detail.data.normList.map(norm => ({ ...norm, score: 0 })),
			comment: '',
			isCompleted: false
		};
	}

	return {
		normList: [...EVALUATION_CONSTANTS.DEFAULT_NORMS],
		comment: '',
		isCompleted: false
	};
}

// 创建默认评教数据
export function createDefaultEvaluationData() {
	return {
		normList: [...EVALUATION_CONSTANTS.DEFAULT_NORMS],
		comment: '',
		isCompleted: false
	};
}

// 获取缓存的评教数据
export function getCachedEvaluationData(id) {
	const cacheKey = `evaluation_cache_${id}`;
	const cacheStr = uni.getStorageSync(cacheKey);

	if (cacheStr) {
		const cacheData = JSON.parse(cacheStr);
		return {
			normList: cacheData.normList || [],
			comment: cacheData.comment || '',
			isCompleted: cacheData.isCompleted || false
		};
	}

	return null;
}

// 获取缓存状态
export function getCacheStatus(evaluationCache, id) {
	return evaluationCache && evaluationCache[id] ? true : false;
}

// 加载评教缓存数据
export function loadCacheData() {
	try {
		const cacheStr = uni.getStorageSync('evaluation_list_cache');
		if (cacheStr) {
			return JSON.parse(cacheStr);
		}
		return {};
	} catch (e) {
		console.error('加载缓存失败:', e);
		return {};
	}
}

// 格式化评教提交数据
export function formatEvaluationData(evaluationData) {
	const formattedData = evaluationData.map(item => {
		const configId = item.courseTeachingEvaluationStudentConfigID || item.id || item.courseTeachingClassTaskID;
		const classTaskId = item.courseTeachingClassTaskID || item.id;
		const teacherID = item.teacherID || 0;
		const teacherName = item.teacherName || item.name || '';
		const otherContect = item.otherContect || '';

		const questions = (item.detail || item.courseTeachingEvaluationQuestionsVOS || []).map((norm, idx) => {
			const score = norm.score > 0 ? norm.score : EVALUATION_CONSTANTS.DEFAULT_SCORE;
			const getScore = norm.getScore || score;
			return {
				courseTeachingEvaluationStudentConfigID: configId,
				courseTeachingClassTaskID: classTaskId,
				otherContect: norm.otherContect || '',
				teacherID: norm.teacherID || teacherID,
				teacherName: norm.teacherName || teacherName,
				score,
				getScore,
				content: norm.content || '',
				courseTeachingEvaluationNormID: norm.courseTeachingEvaluationNormID || norm.id || norm.normId || 0,
				evaluationNumber: norm.evaluationNumber || String(idx + 1)
			};
		});

		return {
			courseTeachingEvaluationStudentConfigID: configId,
			courseTeachingClassTaskID: classTaskId,
			otherContect,
			teacherID,
			teacherName,
			courseTeachingEvaluationQuestionsDTOS: questions
		};
	});

	return formattedData;
}

// 获取待评教项目
export function getPendingEvaluationItems(evaluationList) {
	return evaluationList.filter(item => !item.evaluated);
}

// 验证待评教项目
export function validatePendingItems(pendingItems) {
	if (!pendingItems || pendingItems.length === 0) {
		return false;
	}
	return true;
}

// 应用用户评分到详情数据
export function applyUserScoresToDetail(evaluationData, itemId, detailArray) {
	const userNormList = evaluationData[itemId]?.normList || [];

	return detailArray.map((normItem, index) => {
		const userScore = userNormList[index]?.score;
		const score = userScore > 0 ? userScore : EVALUATION_CONSTANTS.DEFAULT_SCORE;

		return {
			...normItem,
			score
		};
	});
}

// 处理单个评教项数据
export function processEvaluationItemData(item, detail, evaluationData) {
	if (!detail) {
		console.warn(`获取评教详情失败(${item.id}): 返回数据为空`);
		return null;
	}

	let detailArray;
	let baseInfo;

	if (detail.detail && Array.isArray(detail.detail)) {
		detailArray = detail.detail;
		baseInfo = {
			teacherName: item.teacherName || detail.teacherName,
			name: item.name || detail.name
		};
	} else if (Array.isArray(detail)) {
		detailArray = detail;
		baseInfo = {
			teacherName: item.teacherName,
			name: item.name
		};
	} else {
		console.warn(`评教详情数据结构不符合预期(${item.id}):`, detail);
		return null;
	}

	const detailWithScores = applyUserScoresToDetail(evaluationData, item.id, detailArray);

	return {
		...baseInfo,
		startTime: item.startTime,
		id: item.id,
		endTime: item.endTime,
		detail: detailWithScores,
		otherContect: evaluationData[item.id]?.comment || ''
	};
}

// 获取评教详情（带缓存和预览支持）
export async function fetchEvaluationDetail(id, isPreviewMode, evaluationList, evaluationData) {
	// 无效 id 直接跳过
	if (!id || id === 'undefined' || id === 'null') {
		console.warn('[fetchEvaluationDetail] invalid id:', id);
		return null;
	}

	// 预览模式返回默认数据
	if (isPreviewMode || (id && id.toString().startsWith('preview-'))) {
		return EVALUATION_CONSTANTS.DEFAULT_NORMS;
	}

	// 优先使用列表已携带的题目
	const fromList = evaluationList.find(i => i.id == id);
	if (fromList && Array.isArray(fromList.courseTeachingEvaluationQuestionsVOS) && fromList.courseTeachingEvaluationQuestionsVOS.length) {
		return fromList.courseTeachingEvaluationQuestionsVOS;
	}

	// 使用缓存
	const cached = evaluationData[id]?.normList;
	if (Array.isArray(cached) && cached.length) {
		return cached;
	}

	const { getEvaluationDetail } = await import('../pages/api/evaluation.js');
	return await getEvaluationDetail(id);
}

// 清除评教缓存
export function clearEvaluationCache(id) {
	const cacheKey = `evaluation_cache_${id}`;
	uni.removeStorageSync(cacheKey);
}

// 更新列表缓存
export function updateListCache(evaluationCache) {
	try {
		uni.setStorageSync('evaluation_list_cache', JSON.stringify(evaluationCache));
	} catch (e) {
		console.error('更新列表缓存失败:', e);
	}
}
