/**
 * 评教工具函数
 */

// Constants for evaluation
export const EVALUATION_CONSTANTS = {
	AUTO_SAVE_INTERVAL: 10000,
	CACHE_PREFIX: 'evaluation_cache_',
	LIST_CACHE_KEY: 'evaluation_list_cache',
	DEFAULT_SWIPER_HEIGHT: 500,
	DEBOUNCE_DELAY: 300,
	MAX_SCORE: 5,
	COMMENT_MAX_LENGTH: 200,
	DEFAULT_NORM_LIST: [
		{ id: 1, name: '教学态度', score: 0 },
		{ id: 2, name: '教学内容', score: 0 },
		{ id: 3, name: '教学方法', score: 0 },
		{ id: 4, name: '教学效果', score: 0 },
		{ id: 5, name: '教材使用', score: 0 }
	],
	QUICK_COMMENTS: [
		'老师教学认真负责，课程内容丰富有趣，受益匪浅！',
		'老师讲解清晰，举例生动，课堂氛围活跃。',
		'老师备课充分，教学方法灵活多样，能够调动学生积极性。',
		'老师专业知识扎实，教学经验丰富，对学生有耐心。',
		'课程内容实用性强，理论与实践结合紧密。'
	]
};

// Cache helpers
export function saveEvaluationCache(courseId, data) {
	try {
		const cacheKey = `${EVALUATION_CONSTANTS.CACHE_PREFIX}${courseId}`;
		const cacheData = {
			...data,
			timestamp: Date.now()
		};
		uni.setStorageSync(cacheKey, JSON.stringify(cacheData));
		return true;
	} catch (e) {
		console.error('保存评教缓存失败:', e);
		return false;
	}
}

export function loadEvaluationCache(courseId) {
	try {
		const cacheKey = `${EVALUATION_CONSTANTS.CACHE_PREFIX}${courseId}`;
		const cacheStr = uni.getStorageSync(cacheKey);
		
		if (cacheStr) {
			return JSON.parse(cacheStr);
		}
	} catch (error) {
		console.error('加载评教缓存失败:', error);
	}
	return null;
}

export function removeEvaluationCache(courseId) {
	try {
		const cacheKey = `${EVALUATION_CONSTANTS.CACHE_PREFIX}${courseId}`;
		uni.removeStorageSync(cacheKey);
		return true;
	} catch (e) {
		console.error('删除评教缓存失败:', e);
		return false;
	}
}

export function clearAllEvaluationCache(evaluationList) {
	try {
		if (evaluationList && evaluationList.length) {
			evaluationList.forEach(item => {
				const cacheKey = `${EVALUATION_CONSTANTS.CACHE_PREFIX}${item.id}`;
				uni.removeStorageSync(cacheKey);
			});
		}
		return true;
	} catch (e) {
		console.error('清除所有评教缓存失败:', e);
		return false;
	}
}

export function getEvaluationListCache() {
	try {
		let listCache = uni.getStorageSync(EVALUATION_CONSTANTS.LIST_CACHE_KEY);
		
		if (!listCache) {
			return {};
		}
		
		if (typeof listCache === 'string') {
			listCache = JSON.parse(listCache);
		}
		
		return listCache || {};
	} catch (error) {
		console.error('获取评教列表缓存失败:', error);
		return {};
	}
}

export function setEvaluationListCache(data) {
	try {
		uni.setStorageSync(EVALUATION_CONSTANTS.LIST_CACHE_KEY, JSON.stringify(data));
		return true;
	} catch (e) {
		console.error('设置评教列表缓存失败:', e);
		return false;
	}
}

export function clearEvaluationListCache() {
	try {
		uni.removeStorageSync(EVALUATION_CONSTANTS.LIST_CACHE_KEY);
		return true;
	} catch (e) {
		console.error('清除评教列表缓存失败:', e);
		return false;
	}
}

// Calculate score helpers
export function calculateTotalScore(normScores, maxScore = EVALUATION_CONSTANTS.MAX_SCORE) {
	if (!normScores || normScores.length === 0) return 0;
	return normScores.length * maxScore;
}

export function calculateCurrentScore(normScores) {
	if (!normScores || normScores.length === 0) return 0;
	return normScores.reduce((sum, norm) => {
		return sum + Math.max(0, norm.score || 0);
	}, 0);
}

export function calculateTotalActualScore(normScores) {
	if (!normScores || normScores.length === 0) return 0;
	return normScores.reduce((sum, norm) => {
		const score = norm.score || norm.selectedScore || 0;
		const max = norm.maxScore || EVALUATION_CONSTANTS.MAX_SCORE;
		return sum + Math.max(0, score) * (100 / max);
	}, 0);
}

export function checkAllNormsScored(normScores) {
	if (!normScores || normScores.length === 0) {
		return false;
	}
	return normScores.every(norm => norm.score > 0);
}

export function validateEvaluationData(evaluationData, normList) {
	if (!evaluationData || !normList || normList.length === 0) {
		return { valid: false, message: '评教数据不完整' };
	}
	
	const unscoredNorms = normList.filter(norm => !norm.score || norm.score === 0);
	if (unscoredNorms.length > 0) {
		return { valid: false, message: `还有 ${unscoredNorms.length} 项未评分` };
	}
	
	return { valid: true };
}

// Format helpers
export function formatScoreColor(score) {
	const scoreNum = Number(score) || 0;
	if (scoreNum === 0) return 'var(--text-tertiary)';
	if (scoreNum < 60) return 'var(--warning-color)';
	if (scoreNum < 80) return 'var(--primary-color)';
	return 'var(--success-color)';
}

export function getScoreLabel(score) {
	const scoreNum = Number(score) || 0;
	if (scoreNum === 0) return '未评分';
	if (scoreNum < 60) return '较差';
	if (scoreNum < 80) return '良好';
	if (scoreNum < 100) return '优秀';
	return '满分';
}

export function getScoreClass(percentage) {
	const score = Number(percentage) || 0;
	
	if (score === 0) return 'score-none';
	if (score < 60) return 'score-low';
	if (score < 80) return 'score-medium';
	return 'score-high';
}

// Format time range
export function formatTimeRange(startTime, endTime) {
	if (!startTime || !endTime) {
		return '未设置';
	}
	
	const formattedStart = formatSingleDate(startTime);
	const formattedEnd = formatSingleDate(endTime);
	
	return `${formattedStart} ~ ${formattedEnd}`;
}

export function formatSingleDate(dateStr) {
	if (!dateStr) return '';
	return dateStr.split(' ')[0];
}

// Calculate percentage
export function calculateScorePercentage(normScores, maxScore = EVALUATION_CONSTANTS.MAX_SCORE) {
	const totalScore = calculateTotalScore(normScores, maxScore);
	if (totalScore === 0) return 0;
	
	const currentScore = calculateCurrentScore(normScores);
	return Math.round((currentScore / totalScore) * 100);
}

// Process evaluation detail from API
export function processEvaluationDetail(detail) {
	let normList = [];

	if (!detail) {
		normList = [...EVALUATION_CONSTANTS.DEFAULT_NORM_LIST];
	} else if (Array.isArray(detail)) {
		normList = detail.length > 0 ? detail.map(norm => ({ ...norm, score: 0 })) : [...EVALUATION_CONSTANTS.DEFAULT_NORM_LIST];
	} else if (detail.success && Array.isArray(detail.result)) {
		if (detail.result.length > 0) {
			normList = detail.result.map(norm => ({ ...norm, score: 0 }));
		} else {
			normList = [...EVALUATION_CONSTANTS.DEFAULT_NORM_LIST];
		}
	} else if (detail.success && detail.result && Array.isArray(detail.result.normList)) {
		normList = detail.result.normList.map(norm => ({ ...norm, score: 0 }));
	} else if (detail.data && Array.isArray(detail.data.normList)) {
		normList = detail.data.normList.map(norm => ({ ...norm, score: 0 }));
	} else {
		normList = [...EVALUATION_CONSTANTS.DEFAULT_NORM_LIST];
	}

	return {
		normList,
		comment: '',
		isCompleted: false
	};
}

// Create default evaluation data
export function createDefaultEvaluationData() {
	return {
		normList: [...EVALUATION_CONSTANTS.DEFAULT_NORM_LIST],
		comment: '',
		isCompleted: false
	};
}
