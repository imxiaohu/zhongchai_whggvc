<template>
	<view class="score-check-config-card">
		<view class="config-grid">
			<view class="config-cell">
				<text class="cell-label">查询频率</text>
				<picker
					:range="frequencyOptions"
					:range-key="'label'"
					:value="frequencyIndex"
					@change="(e) => $emit('frequency-change', e.detail.value)"
				>
					<view class="cell-value-box">
						<text class="cell-value">{{ frequencyLabel }}</text>
						<l-icon name="chevron-down" style="font-size: 14px; color: var(--text-tertiary);"></l-icon>
					</view>
				</picker>
			</view>
			<view class="config-cell">
				<text class="cell-label">查询时段</text>
				<picker
					mode="time"
					:value="time"
					@change="(e) => $emit('time-change', e.detail.value)"
				>
					<view class="cell-value-box">
						<text class="cell-value">{{ time }}</text>
						<l-icon name="time" style="font-size: 14px; color: var(--text-tertiary);"></l-icon>
					</view>
				</picker>
			</view>
			<view class="config-cell full-width">
				<text class="cell-label">查询学期</text>
				<picker
					:range="semesterOptions"
					:range-key="'label'"
					:value="semesterIndex"
					@change="(e) => $emit('semester-change', e.detail.value)"
					:disabled="loadingSemesters"
				>
					<view class="cell-value-box">
						<text class="cell-value">{{ loadingSemesters ? '加载中...' : semesterLabel }}</text>
						<l-icon name="calendar" style="font-size: 14px; color: var(--text-tertiary);"></l-icon>
					</view>
				</picker>
			</view>
		</view>
		
		<view class="config-actions">
			<view class="action-btn test" @tap="$emit('test')" :disabled="!enabled">
				<l-icon name="check-1" style="font-size: 16px; margin-right: 6px;"></l-icon>
				<text>立即测试</text>
			</view>
			<view class="action-btn logs" @tap="$emit('view-logs')">
				<l-icon name="view-list" style="font-size: 16px; margin-right: 6px;"></l-icon>
				<text>查看日志</text>
			</view>
		</view>
	</view>
</template>

<script>
export default {
	name: 'ScoreCheckConfig',
	props: {
		frequencyOptions: { type: Array, default: () => [] },
		frequencyIndex: { type: Number, default: 0 },
		frequencyLabel: { type: String, default: '' },
		time: { type: String, default: '09:00' },
		semesterOptions: { type: Array, default: () => [] },
		semesterIndex: { type: Number, default: 0 },
		semesterLabel: { type: String, default: '' },
		loadingSemesters: { type: Boolean, default: false },
		enabled: { type: Boolean, default: false }
	}
};
</script>

<style lang="scss" scoped>
.score-check-config-card {
	margin-top: 20rpx;
	background-color: var(--bg-secondary);
	border-radius: 16rpx;
	padding: 24rpx;
	border: 1rpx solid var(--border-light);
}

.config-grid {
	display: flex;
	flex-wrap: wrap;
	gap: 20rpx;
	margin-bottom: 24rpx;
}

.config-cell {
	flex: 1;
	min-width: calc(50% - 10rpx);
	display: flex;
	flex-direction: column;
	gap: 8rpx;

	&.full-width {
		min-width: 100%;
	}

	.cell-label {
		font-size: 22rpx;
		color: var(--text-tertiary);
	}

	.cell-value-box {
		height: 72rpx;
		background-color: var(--bg-card);
		border-radius: 10rpx;
		padding: 0 20rpx;
		display: flex;
		justify-content: space-between;
		align-items: center;
		border: 1rpx solid var(--border-primary);
	}

	.cell-value {
		font-size: 26rpx;
		color: var(--text-primary);
	}
}

.config-actions {
	display: flex;
	gap: 16rpx;

	.action-btn {
		flex: 1;
		height: 72rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 12rpx;
		font-size: 26rpx;
		transition: all 0.2s ease;

		&.test {
			background-color: var(--primary-color);
			color: #fff;
		}

		&.logs {
			background-color: var(--bg-card);
			color: var(--text-primary);
			border: 1rpx solid var(--border-primary);
		}

		&:active {
			transform: scale(0.98);
		}
	}
}
</style>
