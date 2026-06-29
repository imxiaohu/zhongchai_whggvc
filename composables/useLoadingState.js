/**
 * 加载状态管理 Composable
 * 用于统一处理页面加载状态的清除
 */

const defaultLoadingProps = [
	'loading',
	'loadingNotices',
	'loadingData',
	'loadingList',
	'submitting',
	'refreshing',
	'fetching',
	'processing'
];

const defaultErrorProps = [
	'error',
	'errorMessage',
	'noticeError',
	'dataError',
	'listError',
	'submitError',
	'fetchError'
];

export function useLoadingState(options = {}) {
	const loadingStateProperties = options.loadingStateProperties || ['loading', 'loadingNotices', 'loadingData'];

	function clearAllLoadingStates(vm) {
		const target = vm || this;
		if (!target) return;

		loadingStateProperties.forEach(prop => {
			if (prop in target) {
				target[prop] = false;
			}
		});

		defaultLoadingProps.forEach(prop => {
			if (prop in target) {
				target[prop] = false;
			}
		});
	}

	function clearAllErrorStates(vm) {
		const target = vm || this;
		if (!target) return;

		defaultErrorProps.forEach(prop => {
			if (prop in target) {
				target[prop] = '';
			}
		});
	}

	function handleClearLoadingStates(data, vm) {
		const target = vm || this;
		if (!target) return;

		const name = target.$options?.name || '页面';
		console.log(`${name}收到清除加载状态事件:`, data);

		clearAllLoadingStates(target);
		clearAllErrorStates(target);
	}

	function manualClearLoadingStates(vm) {
		const target = vm || this;
		if (!target) return;

		handleClearLoadingStates({
			reason: 'manual',
			source: target.$options?.name || '页面'
		}, target);
	}

	function setupLoadingStateListener(vm) {
		const target = vm || this;
		if (!target) return;

		const handler = (data) => handleClearLoadingStates(data, target);
		uni.$on('clearAllLoadingStates', handler);

		return () => {
			uni.$off('clearAllLoadingStates', handler);
		};
	}

	return {
		clearAllLoadingStates,
		clearAllErrorStates,
		handleClearLoadingStates,
		manualClearLoadingStates,
		setupLoadingStateListener
	};
}
