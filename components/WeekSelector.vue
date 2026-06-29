<template>
	<view class="week-selector-popup" v-if="visible" @tap="handleMaskClick">
		<view class="popup-content" @tap.stop>
			<!-- 弹窗头部 -->
			<view class="popup-header">
				<view class="header-info">
					<text class="semester-title">{{ semesterInfo.currentSemester || '当前学期' }}</text>
					<text class="semester-subtitle">选择周数</text>
				</view>
				<view class="close-btn" @tap="closePopup">
					<l-icon name="close" size="20" color="var(--text-secondary)"></l-icon>
				</view>
			</view>

			<!-- 周数网格 -->
			<view class="weeks-grid">
				<view 
					v-for="week in totalWeeks" 
					:key="week"
					class="week-item"
					:class="{
						'current': week === currentWeek,
						'selected': week === selectedWeek
					}"
					@tap="selectWeek(week)"
				>
					<view class="week-number">{{ week }}</view>
					<view v-if="week === currentWeek" class="current-label">本周</view>
				</view>
			</view>

			<!-- 底部操作按钮 -->
			<view class="popup-footer">
				<button class="cancel-btn" @tap="closePopup">取消</button>
				<button class="confirm-btn" @tap="confirmSelection" :disabled="!selectedWeek">确定</button>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref, computed, watch } from 'vue';

// Props
const props = defineProps({
	visible: {
		type: Boolean,
		default: false
	},
	currentWeek: {
		type: Number,
		default: 1
	},
	displayWeek: {
		type: Number,
		default: 1
	},
	totalWeeks: {
		type: Number,
		default: 20
	},
	semesterInfo: {
		type: Object,
		default: () => ({
			currentSemester: '',
			nowweek: 1,
			weekCount: 20,
			startDate: ''
		})
	}
});

// Emits
const emit = defineEmits(['close', 'select']);

// 响应式数据
const selectedWeek = ref(props.displayWeek || props.currentWeek);

// 计算属性
const weekList = computed(() => {
	return Array.from({ length: props.totalWeeks }, (_, i) => i + 1);
});

// 监听弹窗显示状态
watch(() => props.visible, (newVal) => {
	if (newVal) {
		// 优先使用当前显示的周数，如果没有则使用当前周
		selectedWeek.value = props.displayWeek || props.currentWeek;
	}
});

// 监听显示周变化，更新默认选中
watch(() => props.displayWeek, (newVal) => {
	if (newVal) {
		selectedWeek.value = newVal;
	}
});

// 方法
const handleMaskClick = () => {
	closePopup();
};

const closePopup = () => {
	selectedWeek.value = null;
	emit('close');
};

const selectWeek = (week) => {
	selectedWeek.value = week;
	// 添加触觉反馈
	uni.vibrateShort();
};

const confirmSelection = () => {
	if (selectedWeek.value) {
		emit('select', selectedWeek.value);
		closePopup();
	}
};
</script>

<style lang="scss" scoped>
.week-selector-popup {
	position: fixed;
	top: 0; left: 0; right: 0; bottom: 0;
	background: rgba(0, 0, 0, 0.45);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 9999;
	backdrop-filter: blur(8px);
	animation: fadeIn 0.25s ease;
}

.popup-content {
	background: rgba(255, 255, 255, 0.95);
	backdrop-filter: blur(20px);
	-webkit-backdrop-filter: blur(20px);
	border-radius: var(--radius-xl);
	margin: 40rpx;
	max-height: 80vh;
	width: calc(100% - 80rpx);
	max-width: 600rpx;
	box-shadow:
		0 24rpx 60rpx rgba(0, 0, 0, 0.18),
		inset 0 1px 0 rgba(255, 255, 255, 0.9);
	border: 1px solid rgba(255, 255, 255, 0.6);
	animation: slideUp 0.25s ease;
	overflow: hidden;
}

.popup-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: var(--spacing-lg) var(--spacing-lg) var(--spacing-md);
}

.header-info {
	flex: 1;

	.semester-title {
		display: block;
		font-size: 16px;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 4rpx;
	}

	.semester-subtitle {
		font-size: 13px;
		color: var(--text-tertiary);
	}
}

.close-btn {
	width: 32px;
	height: 32px;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 50%;
	background: var(--bg-secondary);
	transition: all 0.15s var(--ease-out);

	&:active {
		transform: scale(0.9);
		background: var(--bg-muted);
	}
}

.weeks-grid {
	padding: var(--spacing-md) var(--spacing-lg) var(--spacing-lg);
	display: grid;
	grid-template-columns: repeat(5, 1fr);
	gap: 16rpx;
	max-height: 400rpx;
	overflow-y: auto;
}

.week-item {
	aspect-ratio: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	border-radius: var(--radius-md);
	background: var(--bg-secondary);
	border: 2px solid transparent;
	transition: all 0.2s var(--ease-out);
	position: relative;

	&:active {
		transform: scale(0.94);
	}

	&.current {
		background: linear-gradient(135deg, #6366f1, #818cf8);
		box-shadow: 0 4rpx 14rpx rgba(99, 102, 241, 0.35);

		.week-number {
			color: #ffffff;
		}
	}

	&.selected {
		border-color: var(--primary-600);
		background: var(--primary-soft);

		.week-number {
			color: var(--primary-600);
			font-weight: 700;
		}
	}

	&.current.selected {
		border-color: rgba(255, 255, 255, 0.5);
		background: linear-gradient(135deg, #6366f1, #818cf8);
		box-shadow: 0 4rpx 14rpx rgba(99, 102, 241, 0.4);
	}
}

.week-number {
	font-size: 15px;
	font-weight: 600;
	color: var(--text-primary);
}

.current-label {
	font-size: 11px;
	color: rgba(255, 255, 255, 0.9);
	margin-top: 4rpx;
	font-weight: 500;
}

.popup-footer {
	display: flex;
	gap: 16rpx;
	padding: var(--spacing-md) var(--spacing-lg) var(--spacing-lg);
	border-top: 1px solid var(--border-secondary);
}

.cancel-btn,
.confirm-btn {
	flex: 1;
	height: 44px;
	border-radius: var(--radius-md);
	font-size: 15px;
	font-weight: 600;
	border: none;
	transition: all 0.15s var(--ease-out);

	&:active {
		transform: scale(0.97);
	}
}

.cancel-btn {
	background: var(--bg-secondary);
	color: var(--text-secondary);

	&:active {
		background: var(--bg-muted);
	}
}

.confirm-btn {
	background: linear-gradient(135deg, #6366f1, #818cf8);
	color: #ffffff;
	box-shadow: 0 4rpx 14rpx rgba(99, 102, 241, 0.3);

	&:disabled {
		background: var(--bg-muted);
		color: var(--text-tertiary);
		box-shadow: none;
		transform: none;
	}
}

/* 动画 */
@keyframes fadeIn {
	from { opacity: 0; }
	to { opacity: 1; }
}

@keyframes slideUp {
	from { transform: translateY(40rpx); opacity: 0; }
	to { transform: translateY(0); opacity: 1; }
}

/* #ifdef MP-WEIXIN */
.weeks-grid {
	max-height: 350rpx;
}

.week-item {
	min-height: 80rpx;
}
/* #endif */
</style>
