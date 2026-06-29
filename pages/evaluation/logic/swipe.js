import { ref } from 'vue';
import { onLoad as onUniLoad, onShow, onUnload as onUniUnload } from '@dcloudio/uni-app';
import { onMounted } from 'vue';
import { getEvaluationList, getEvaluationDetail } from '@/pages/api/evaluation.js';
import { useEvaluationForm } from '@/composables/useEvaluationForm.js';
import { EVALUATION_CONSTANTS, createDefaultEvaluationData, processEvaluationDetail } from '@/utils/evaluation.js';
import { submitEvaluation } from '@/pages/api/evaluation.js';

export default {
	setup() {
		let resizeTimer = null;
		let autoSaveTimer = null;

		const {
			currentIndex,
			loading,
			evaluationList,
			evaluationData,
			quickComments,
			CONSTANTS,
			hasUnevaluatedItems,
			statusBarHeight,
			swiperHeight,
			getCurrentItemId,
			getNormList,
			getComment,
			updateComment,
			selectQuickComment,
			selectScore,
			setMaxScore,
			isCompleted,
			cleanup,
			loadSingleEvaluationData,
			loadCachedData,
			loadFromCache,
			checkIsCompleted,
			saveAllEvaluations,
			saveCurrentEvaluation,
			calculateScorePercentage,
			formatTimeRange,
			getScoreClass
		} = useEvaluationForm();

		function showToast(title, icon = 'none') {
			uni.showToast({ title, icon, duration: 3000 });
		}

		function showLoadingDialog(title = '加载中...') {
			uni.showLoading({ title, mask: true });
		}

		function hideLoadingDialog() {
			uni.hideLoading();
		}

		function collectSubmissionData(evaluationDataArr) {
			return evaluationDataArr.map(item => ({
				id: item.id,
				normList: item.normList.map(norm => ({
					normId: norm.courseTeachingEvaluationNormID || norm.normId || norm.id || 0,
					score: norm.score ?? norm.selectedScore ?? 0
				})),
				comment: item.comment || ''
			}));
		}

		function validateSubmissionData(submissionData) {
			if (!Array.isArray(submissionData) || submissionData.length === 0) {
				return { valid: false, message: '提交数据为空' };
			}
			for (const item of submissionData) {
				if (!item.id) return { valid: false, message: '课程ID缺失' };
				if (!Array.isArray(item.normList) || item.normList.length === 0) {
					return { valid: false, message: '评教项为空' };
				}
				const unscored = item.normList.find(n => !n.score || n.score <= 0);
				if (unscored) return { valid: false, message: '存在未评分的项' };
			}
			return { valid: true };
		}

		function mergeUserScores(evaluationDataArr) {
			evaluationDataArr.forEach(item => {
				const userData = evaluationData.value[item.id];
				if (userData && Array.isArray(userData.normList)) {
					userData.normList.forEach((norm, idx) => {
						if (item.normList[idx] && item.normList[idx].score > 0) {
							norm.score = item.normList[idx].score;
							norm.selectedScore = item.normList[idx].score;
						}
					});
				}
			});
		}

		async function submitCurrentEvaluation() {
			const id = evaluationList.value[currentIndex.value]?.id;
			if (!id) {
				showToast('无法获取当前评教ID');
				return false;
			}
			const data = evaluationData.value[id];
			if (!data) {
				showToast('评教数据为空');
				return false;
			}
			const submissionData = collectSubmissionData([{ id, ...data }]);
			const validation = validateSubmissionData(submissionData);
			if (!validation.valid) {
				showToast(validation.message);
				return false;
			}
			try {
				showLoadingDialog('提交中...');
				await submitEvaluation(submissionData);
				hideLoadingDialog();

				const idx = evaluationList.value.findIndex(e => e.id == id);
				if (idx >= 0) evaluationList.value[idx].evaluated = true;

				uni.removeStorageSync(`evaluation_cache_${id}`);
				showToast('提交成功', 'success');
				return true;
			} catch (error) {
				hideLoadingDialog();
				console.error('提交评教失败:', error);
				showToast('提交失败，请重试');
				return false;
			}
		}

		function showQuickSubmitConfirmDialog(evaluationDataArr) {
			const count = evaluationDataArr.length;
			uni.showModal({
				title: '一键评教确认',
				content: `检测到 ${count} 项待评教，是否使用当前评分一键提交？`,
				confirmText: '一键提交',
				cancelText: '取消',
				success: async (res) => {
					if (!res.confirm) return;
					const submissionData = collectSubmissionData(evaluationDataArr);
					const validation = validateSubmissionData(submissionData);
					if (!validation.valid) {
						showToast(validation.message);
						return;
					}
					try {
						showLoadingDialog('一键提交中...');
						await submitEvaluation(submissionData);
						hideLoadingDialog();
						evaluationDataArr.forEach(item => {
							const idx = evaluationList.value.findIndex(e => e.id == item.id);
							if (idx >= 0) evaluationList.value[idx].evaluated = true;
							uni.removeStorageSync(`evaluation_cache_${item.id}`);
						});
						showToast('一键评教成功', 'success');
					} catch (error) {
						hideLoadingDialog();
						console.error('一键评教失败:', error);
						showToast('一键评教失败，请重试');
					}
				}
			});
		}

		const isValidEvaluationListResponse = (res) => {
			if (!res) return false;
			if (res.result && Array.isArray(res.result)) return true;
			if (res.data && Array.isArray(res.data)) return true;
			if (Array.isArray(res)) return true;
			return false;
		};

		const formatEvaluationList = (res) => {
			let dataArray = [];
			if (res.result && Array.isArray(res.result)) dataArray = res.result;
			else if (res.data && Array.isArray(res.data)) dataArray = res.data;
			else if (Array.isArray(res)) dataArray = res;

			if (dataArray.length === 0) {
				return [{ id: 'default_completed', name: '所有评教已完成', teacherName: '系统提示', startTime: '', endTime: '', evaluated: true }];
			}
			return dataArray.map(item => ({
				id: item.id || item.courseTeachingEvaluationStudentConfigID || item.courseTeachingEvaluationStudentConfigId,
				courseTeachingClassTaskID: item.courseTeachingClassTaskID,
				teacherID: item.teacherID,
				name: item.courseName || item.name,
				teacherName: item.teacherName,
				startTime: item.startTime,
				endTime: item.endTime,
				evaluated: item.evaluated || false,
				courseTeachingEvaluationQuestionsVOS: item.courseTeachingEvaluationQuestionsVOS
			}));
		};

		const loadEvaluationList = async () => {
			loading.value = true;
			try {
				const res = await getEvaluationList();
				if (!isValidEvaluationListResponse(res)) {
					throw new Error('获取评教列表失败');
				}
				evaluationList.value = formatEvaluationList(res);
			await loadAllEvaluationData();
			loadCachedData();
			} catch (error) {
				showToast('加载评教列表失败');
			} finally {
				loading.value = false;
			}
		};

		// 从详情接口加载单条评教数据（用于从列表页点击跳转来的场景）
		const loadSingleEvaluationDataFromDetail = async (item) => {
			loading.value = true;
			try {
				if (!item?.id || item.id === 'undefined' || item.id === 'null') {
					console.warn('[swipe] skip loadSingleEvaluationDataFromDetail: invalid item.id', item?.id);
					loading.value = false;
					return;
				}
				// 优先使用列表已携带的题目，避免额外 API 请求
				if (item && Array.isArray(item.courseTeachingEvaluationQuestionsVOS) && item.courseTeachingEvaluationQuestionsVOS.length > 0) {
					evaluationData.value[item.id] = {
						normList: item.courseTeachingEvaluationQuestionsVOS.map(norm => ({ ...norm, score: 0 })),
						comment: '',
						isCompleted: false
					};
				} else {
					const detail = await getEvaluationDetail(item.id);
					if (!detail) {
						evaluationData.value[item.id] = createDefaultEvaluationData();
					} else {
						evaluationData.value[item.id] = processEvaluationDetail(detail);
					}
				}
				// 尝试合并本地缓存的评分
				const cached = loadFromCache(item.id);
				if (cached && cached.normList && cached.normList.length > 0) {
					const data = evaluationData.value[item.id];
					if (data && data.normList && data.normList.length > 0) {
						cached.normList.forEach((cachedNorm, idx) => {
							if (data.normList[idx]) {
								data.normList[idx].score = cachedNorm.score ?? data.normList[idx].score;
							}
						});
					}
					evaluationData.value[item.id] = {
						...data,
						comment: cached.comment || data.comment || '',
						isCompleted: checkIsCompleted(data.normList)
					};
				}
			} catch (error) {
				console.error(`加载评教详情失败 ID:${item.id}`, error);
				evaluationData.value[item.id] = createDefaultEvaluationData();
			} finally {
				loading.value = false;
			}
		};

		const loadAllEvaluationData = async () => {
			const promises = evaluationList.value.map(item => loadSingleEvaluationData(item));
			await Promise.all(promises);
		};

		const getPendingEvaluationItems = () => evaluationList.value.filter(item => !item.evaluated);

		const prepareQuickEvaluationData = async (pendingItems) => {
			const evaluationDataArr = [];
			for (const item of pendingItems) {
				try {
					if (item && Array.isArray(item.courseTeachingEvaluationQuestionsVOS) && item.courseTeachingEvaluationQuestionsVOS.length) {
						const normList = item.courseTeachingEvaluationQuestionsVOS.map(q => ({
							normId: q.courseTeachingEvaluationNormID || q.normId || q.id,
							normName: q.normName || q.name || '',
							score: q.maxScore || 5,
							maxScore: q.maxScore || 5,
							selectedScore: q.maxScore || 5
						}));
						evaluationDataArr.push({ id: item.id, normList, comment: '' });
					} else {
						const detail = await getEvaluationDetail(item.id);
						if (detail) {
							const normList = (Array.isArray(detail) ? detail : (detail.result || detail.data || [])).map(q => ({
								normId: q.courseTeachingEvaluationNormID || q.normId || q.id,
								normName: q.normName || q.name || '',
								score: q.maxScore || 5,
								maxScore: q.maxScore || 5,
								selectedScore: q.maxScore || 5
							}));
							evaluationDataArr.push({ id: item.id, normList, comment: '' });
						}
					}
				} catch (error) {
					console.error(`准备评教数据失败(${item.id}):`, error);
				}
			}
			return evaluationDataArr;
		};

		const quickEvaluateAndSubmitAll = async () => {
			try {
				const pendingItems = getPendingEvaluationItems();
				if (pendingItems.length === 0) { showToast('没有待评教的课程'); return; }
				showLoadingDialog('数据准备中...');
				const evaluationDataArr = await prepareQuickEvaluationData(pendingItems);
				hideLoadingDialog();
				if (evaluationDataArr.length === 0) { showToast('获取评教数据失败'); return; }
				mergeUserScores(evaluationDataArr);
				showQuickSubmitConfirmDialog(evaluationDataArr);
			} catch (error) {
				hideLoadingDialog();
				showToast('准备一键评教数据失败');
			}
		};

		const onSwiperChange = (e) => {
			currentIndex.value = e.detail.current;
		};

		const calculateSwiperHeight = () => {
			try {
				const systemInfo = uni.getSystemInfoSync();
				const safeArea = systemInfo.safeAreaInsets || { bottom: 0 };
				const statusBarH = systemInfo.statusBarHeight || 0;
				const finalHeight = systemInfo.windowHeight - statusBarH - uni.upx2px(100) - uni.upx2px(120) - safeArea.bottom;
				swiperHeight.value = finalHeight > 0 ? finalHeight : 600;
			} catch (error) {
				swiperHeight.value = EVALUATION_CONSTANTS.DEFAULT_SWIPER_HEIGHT;
			}
		};

		const debouncedCalculateHeight = () => {
			if (resizeTimer) clearTimeout(resizeTimer);
			resizeTimer = setTimeout(calculateSwiperHeight, EVALUATION_CONSTANTS.DEBOUNCE_DELAY);
		};

		const findNextUnevaluatedIndex = () => {
			return evaluationList.value.findIndex((item, idx) => idx > currentIndex.value && !item.evaluated);
		};

		const findFirstUnevaluatedIndex = () => {
			return evaluationList.value.findIndex(item => !item.evaluated);
		};

		const goToNextUnevaluated = () => {
			const nextIndex = findNextUnevaluatedIndex();
			if (nextIndex >= 0) {
				currentIndex.value = nextIndex;
			} else {
				const firstIndex = findFirstUnevaluatedIndex();
				if (firstIndex >= 0) currentIndex.value = firstIndex;
			}
		};

		const navigateBack = () => {
			try {
				uni.navigateBack({ delta: 1, fail: () => uni.switchTab({ url: '/pages/index/index' }) });
			} catch {
				uni.switchTab({ url: '/pages/index/index' });
			}
		};

		const cleanupPage = () => {
			if (autoSaveTimer) { clearInterval(autoSaveTimer); autoSaveTimer = null; }
			if (resizeTimer) { clearTimeout(resizeTimer); resizeTimer = null; }
			saveAllEvaluations();
			cleanup();
		};

		onUniLoad(async (options) => {
			const token = uni.getStorageSync('token');
			if (!token) {
				uni.showToast({ title: '请先登录获取评教数据', icon: 'none', duration: 3000 });
				setTimeout(() => uni.navigateBack(), 1500);
				return;
			}
			const systemInfo = uni.getSystemInfoSync();
			statusBarHeight.value = systemInfo.statusBarHeight || 0;

			// 从 URL 参数读取单条评教数据（从列表页点击单个课程跳转而来）
			if (options && options.id && options.id !== 'undefined') {
				const item = {
					id: options.id,
					name: options.name || '',
					teacherName: options.teacherName || '',
					startTime: options.startTime || '',
					endTime: options.endTime || ''
				};
				evaluationList.value = [item];
				await loadSingleEvaluationDataFromDetail(item);
			} else {
				// 完整列表模式（直接访问 swipe 页面或一键评教入口）
				await loadEvaluationList();
			}

			calculateSwiperHeight();
			autoSaveTimer = setInterval(() => {
				saveAllEvaluations();
			}, EVALUATION_CONSTANTS.AUTO_SAVE_INTERVAL);
		});

		onMounted(() => {
			debouncedCalculateHeight();
		});

		onShow(() => {
			debouncedCalculateHeight();
		});

		onUniUnload(() => {
			cleanupPage();
		});

		return {
			statusBarHeight,
			swiperHeight,
			currentIndex,
			loading,
			evaluationList,
			evaluationData,
			quickComments,
			CONSTANTS,
			hasUnevaluatedItems,
			navigateBack,
			quickEvaluateAndSubmitAll,
			onSwiperChange,
			isCompleted,
			submitCurrentEvaluation,
			saveCurrentEvaluation,
			getCurrentItemId,
			getNormList,
			getComment,
			updateComment,
			selectQuickComment,
			selectScore,
			setMaxScore,
			calculateScorePercentage,
			formatTimeRange,
			getScoreClass
		};
	}
};
