<template>
	<view class="sc-card" :class="{ 'sc-card--failed': isFailed }">
		<!-- 左侧分数色条 -->
		<view class="sc-score-bar" :class="scoreBarClass"></view>

		<!-- 卡片主体 -->
		<view class="sc-card-body">
			<!-- 头部：课程名 + 分数 -->
			<view class="sc-header">
				<view class="sc-course-info">
					<text class="sc-course-name">{{ course.courseName }}</text>
					<view class="sc-tags">
						<text class="sc-tag sc-tag--property">{{ course.courseProperty || course.examType || '必修' }}</text>
						<text class="sc-tag sc-tag--credit">{{ course.credit }} 学分</text>
					</view>
				</view>
				<view class="sc-score-display">
					<text class="sc-score-value" :class="scoreStatusClass">{{ course.finalScore }}</text>
					<text class="sc-score-label">总分</text>
				</view>
			</view>

			<!-- 详细分项 -->
			<view class="sc-detail-row">
				<view class="sc-detail-item" v-if="course.dailyScore !== undefined">
					<text class="sc-detail-label">平时</text>
					<text class="sc-detail-value">{{ course.dailyScore }}</text>
				</view>
				<view class="sc-detail-item" v-if="course.courseScore !== undefined">
					<text class="sc-detail-label">期末</text>
					<text class="sc-detail-value">{{ course.courseScore }}</text>
				</view>
				<view class="sc-detail-item">
					<text class="sc-detail-label">绩点</text>
					<text class="sc-detail-value sc-detail-value--primary">{{ course.getPoint || course.gpa || '0.0' }}</text>
				</view>
			</view>

			<!-- 底部：教师 + 考试日期 -->
			<view class="sc-footer" v-if="course.teacherNames || course.examDate">
				<view class="sc-footer-item" v-if="course.teacherNames">
					<l-icon name="user-circle-filled" style="font-size: 14px; color: var(--text-tertiary);"></l-icon>
					<text class="sc-footer-text">{{ formatTeacherName(course.teacherNames) }}</text>
				</view>
				<view class="sc-footer-item" v-if="course.examDate">
					<l-icon name="calendar" style="font-size: 14px; color: var(--text-tertiary);"></l-icon>
					<text class="sc-footer-text">{{ course.examDate }}</text>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
	course: {
		type: Object,
		required: true
	}
})

const isFailed = computed(() => {
	const score = parseFloat(props.course.finalScore)
	return !isNaN(score) && score < 60
})

const scoreStatusClass = computed(() => {
	if (isFailed.value) return 'sc-score--failed'
	const score = parseFloat(props.course.finalScore)
	if (score >= 90) return 'sc-score--excellent'
	if (score >= 80) return 'sc-score--good'
	return 'sc-score--normal'
})

const scoreBarClass = computed(() => {
	if (isFailed.value) return 'sc-score-bar--failed'
	const score = parseFloat(props.course.finalScore)
	if (score >= 90) return 'sc-score-bar--excellent'
	if (score >= 80) return 'sc-score-bar--good'
	return 'sc-score-bar--normal'
})

function formatTeacherName(names) {
	if (!names) return ''
	return names.replace(/,/g, ' ')
}
</script>

<style lang="scss" scoped>
.sc-card {
	display: flex;
	background: #fff;
	border-radius: 24rpx;
	overflow: hidden;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	transition: all 0.2s var(--ease-out);
	margin-bottom: 16rpx;

	&:active {
		transform: scale(0.99);
		background: #f8fafc;
	}
}

.sc-score-bar {
	width: 8rpx;
	flex-shrink: 0;

	&--failed { background: linear-gradient(180deg, var(--error-color), #dc2626); }
	&--excellent { background: linear-gradient(180deg, var(--success-color), #059669); }
	&--good { background: linear-gradient(180deg, #3b82f6, #2563eb); }
	&--normal { background: linear-gradient(180deg, var(--primary-500), var(--primary-600)); }
}

.sc-card-body {
	flex: 1;
	padding: 24rpx;
}

/* Header */
.sc-header {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
	margin-bottom: 20rpx;
}

.sc-course-info {
	flex: 1;
	margin-right: 20rpx;
}

.sc-course-name {
	display: block;
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	line-height: 1.4;
	margin-bottom: 10rpx;
}

.sc-tags {
	display: flex;
	flex-wrap: wrap;
	gap: 8rpx;
}

.sc-tag {
	padding: 4rpx 14rpx;
	border-radius: 100rpx;
	font-size: 20rpx;
	font-weight: 600;

	&--property {
		background: var(--primary-soft);
		color: var(--primary-600);
	}
	&--credit {
		background: var(--bg-muted);
		color: var(--text-secondary);
	}
}

/* Score Display */
.sc-score-display {
	display: flex;
	flex-direction: column;
	align-items: center;
	min-width: 80rpx;
}

.sc-score-value {
	font-size: 44rpx;
	font-weight: 800;
	line-height: 1;

	&.sc-score--failed { color: var(--error-color); }
	&.sc-score--excellent { color: var(--success-color); }
	&.sc-score--good { color: var(--info-color); }
	&.sc-score--normal { color: var(--primary-600); }
}

.sc-score-label {
	font-size: 18rpx;
	color: var(--text-tertiary);
	margin-top: 4rpx;
	font-weight: 500;
}

/* Detail Row */
.sc-detail-row {
	display: flex;
	background: var(--bg-muted);
	border-radius: 16rpx;
	padding: 12rpx 8rpx;
	margin-bottom: 16rpx;
	gap: 0;
}

.sc-detail-item {
	flex: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	border-right: 1px solid rgba(148, 163, 184, 0.2);

	&:last-child { border-right: none; }
}

.sc-detail-label {
	font-size: 18rpx;
	color: var(--text-tertiary);
	font-weight: 500;
	margin-bottom: 4rpx;
}

.sc-detail-value {
	font-size: 26rpx;
	font-weight: 700;
	color: var(--text-secondary);

	&.sc-detail-value--primary { color: var(--primary-600); }
}

/* Footer */
.sc-footer {
	display: flex;
	flex-wrap: wrap;
	gap: 20rpx;
	padding-top: 12rpx;
	border-top: 1px dashed rgba(226, 232, 240, 0.8);
}

.sc-footer-item {
	display: flex;
	align-items: center;
	gap: 6rpx;
}

.sc-footer-text {
	font-size: 22rpx;
	color: var(--text-tertiary);
}
</style>
