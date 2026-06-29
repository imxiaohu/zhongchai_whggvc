<template>
	<view class="today-class-container">
		<view v-if="!hideHeader" class="header">
			<view class="title-row">
				<text class="title">今日课程</text>
				<text class="date">{{ todayDateStr }}</text>
			</view>
			<view class="view-schedule" @tap="goToSchedule">
				<text>完整课表</text>
				<text class="arrow-icon">›</text>
			</view>
		</view>

		<view v-if="loading" class="state-view">
			<view class="loading-ring">
				<view class="ring-dot"></view>
				<view class="ring-dot"></view>
				<view class="ring-dot"></view>
			</view>
			<text>正在获取课程...</text>
		</view>

		<view v-else-if="filteredClasses.length === 0" class="state-view empty">
			<view class="empty-illustration">
				<view class="empty-moon"></view>
				<view class="empty-stars">
					<view class="star star-1"></view>
					<view class="star star-2"></view>
					<view class="star star-3"></view>
				</view>
			</view>
			<text class="empty-main-text">今天没有课程</text>
			<text class="empty-sub-text">去休息一下吧~</text>
		</view>

		<!-- 横向布局（原始滚动卡片） -->
		<scroll-view v-else-if="layout === 'horizontal'" scroll-x class="class-scroll" :show-scrollbar="false">
			<view v-if="classStatus !== 'empty' && classStatus !== 'finished'" class="class-status-bar" :class="`class-status-bar--${classStatus}`">
				<view class="status-dot" :class="`status-dot--${classStatus}`"></view>
				<text class="status-badge-text">{{ classStatusLabel }}</text>
				<text v-if="countdownText" class="status-divider">|</text>
				<text v-if="countdownText" class="status-countdown">{{ countdownText }}</text>
			</view>

			<view class="class-list">
				<view
					v-for="(item, index) in filteredClasses"
					:key="item.id || index"
					class="class-card"
					:class="{ 'is-active': isCurrentClass(item) }"
				>
					<view class="card-left" :style="{ background: getCardGradient(index) }">
						<text class="time-start">{{ item.startTime || '-' }}</text>
						<view class="time-divider"></view>
						<text class="time-end">{{ item.endTime || '-' }}</text>
					</view>
					<view class="card-right">
						<view class="card-right-top">
							<text class="course-name">{{ item.courseName }}</text>
							<view v-if="item.sectionStart" class="course-tag" :style="{ background: getTagBg(index), color: getTagColor(index) }">
								{{ item.sectionStart }}-{{ item.sectionEnd }}节
							</view>
						</view>
						<view class="course-info">
							<view class="info-item">
								<text class="info-icon">⌂</text>
								<text>{{ item.location }}</text>
							</view>
							<view class="info-item">
								<text class="info-icon">◈</text>
								<text>{{ item.teacherName }}</text>
							</view>
						</view>
					</view>
					<view v-if="isCurrentClass(item)" class="active-indicator"></view>
				</view>
			</view>
		</scroll-view>

		<!-- 竖向布局（按上午/下午分组） -->
		<view v-else class="class-list-vertical">
			<view v-if="classStatus !== 'empty' && classStatus !== 'finished'" class="class-status-bar" :class="`class-status-bar--${classStatus}`">
				<view class="status-dot" :class="`status-dot--${classStatus}`"></view>
				<text class="status-badge-text">{{ classStatusLabel }}</text>
				<text v-if="countdownText" class="status-divider">|</text>
				<text v-if="countdownText" class="status-countdown">{{ countdownText }}</text>
			</view>

			<view
				v-for="group in groupedClasses"
				:key="group.period"
				class="period-group"
			>
				<view class="period-header">
					<view class="period-label" :class="`period-label--${group.period}`">
						<text class="period-icon">{{ group.period === 'am' ? '☀' : group.period === 'pm' ? '☾' : '✦' }}</text>
						<text>{{ group.periodLabel }}</text>
					</view>
					<view class="period-meta">
						<text class="period-dot"></text>
						<text class="period-count">{{ group.items.length }}节课</text>
					</view>
				</view>

				<view class="period-courses">
					<view
						v-for="(item, index) in group.items"
						:key="item.id || index"
						class="class-card-vertical"
						:class="{ 'is-active': isCurrentClass(item) }"
					>
						<view class="card-vertical-time" :style="{ background: getCardGradient(item._origIndex) }">
							<text class="time-lesson-num">{{ getSectionLabel(item) }}</text>
							<view class="time-divider-v"></view>
							<view class="time-range-v">
								<text class="time-start-v">{{ item.startTime || '-' }}</text>
								<text class="time-sep">~</text>
								<text class="time-end-v">{{ item.endTime || '-' }}</text>
							</view>
						</view>

						<view class="card-vertical-body">
							<view class="card-vertical-header">
								<text class="course-name-v">{{ item.courseName }}</text>
								<view v-if="isCurrentClass(item)" class="live-badge">正在上课</view>
							</view>
							<view class="course-info-v">
								<view class="info-item-v">
									<text class="info-icon-v">⌂</text>
									<text>{{ item.location }}</text>
								</view>
								<view class="info-item-v">
									<text class="info-icon-v">◈</text>
									<text>{{ item.teacherName }}</text>
								</view>
							</view>
						</view>
					</view>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import { useCourseCache } from '../store/courseCache.js'

const props = defineProps({
	classes: {
		type: Array,
		default: () => []
	},
	loading: {
		type: Boolean,
		default: false
	},
	hideHeader: {
		type: Boolean,
		default: false
	},
	layout: {
		type: String,
		default: 'vertical'
	},
	hidePast: {
		type: Boolean,
		default: false
	}
})

const gradients = [
	'linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%)',
	'linear-gradient(135deg, #10b981 0%, #059669 100%)',
	'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)',
	'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)',
	'linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%)',
	'linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%)',
]

function getCardGradient(index) {
	return gradients[index % gradients.length]
}

function getTagBg(index) {
	const tagBgs = [
		'rgba(79, 70, 229, 0.12)',
		'rgba(16, 185, 129, 0.12)',
		'rgba(245, 158, 11, 0.12)',
		'rgba(239, 68, 68, 0.12)',
		'rgba(14, 165, 233, 0.12)',
		'rgba(139, 92, 246, 0.12)',
	]
	return tagBgs[index % tagBgs.length]
}

function getTagColor(index) {
	const tagColors = ['#4f46e5', '#10b981', '#f59e0b', '#ef4444', '#0ea5e9', '#8b5cf6']
	return tagColors[index % tagColors.length]
}

const todayDateStr = computed(() => {
	const now = new Date()
	return `${now.getMonth() + 1}月${now.getDate()}日 星期${['日', '一', '二', '三', '四', '五', '六'][now.getDay()]}`
})

const lessonTimesMap = computed(() => {
	try {
		const courseCacheStore = useCourseCache()
		const lessonTimes = courseCacheStore.lessonTime?.data
		if (!lessonTimes || !Array.isArray(lessonTimes)) return {}
		const map = {}
		lessonTimes.forEach(item => {
			if (item.name && item.startTime && item.endTime) {
				map[String(item.name)] = { startTime: item.startTime, endTime: item.endTime }
			}
		})
		return map
	} catch (e) {
		return {}
	}
})

function parseSection(section) {
	if (!section) return { sectionStart: null, sectionEnd: null }
	const str = String(section)
	const matches = [...str.matchAll(/#?(\d+)#?/g)]
	if (matches.length === 0) return { sectionStart: null, sectionEnd: null }
	const start = parseInt(matches[0][1], 10)
	const end = matches[1] ? parseInt(matches[1][1], 10) : start
	return { sectionStart: start, sectionEnd: end }
}

function getLessonTimes(sectionStart) {
	if (!sectionStart) return { startTime: '', endTime: '' }
	const times = lessonTimesMap.value[String(sectionStart)]
	if (times) {
		return {
			startTime: times.startTime.substring(0, 5),
			endTime: times.endTime.substring(0, 5)
		}
	}
	const DEFAULT_TIMES = [
		'08:30', '09:20', '10:20', '11:10', '14:00', '14:50',
		'15:50', '16:40', '19:00', '19:50', '20:40', '21:30'
	]
	const DEFAULT_ENDS = [
		'09:15', '10:05', '11:05', '11:55', '14:45', '15:35',
		'16:35', '17:25', '19:45', '20:35', '21:25', '22:15'
	]
	const idx = sectionStart - 1
	const start = idx >= 0 && idx < 12 ? DEFAULT_TIMES[idx] : ''
	const end = idx >= 0 && idx < 12 ? DEFAULT_ENDS[idx] : ''
	return { startTime: start, endTime: end }
}

function getPeriod(sectionStart) {
	if (!sectionStart) return 'am'
	if (sectionStart <= 4) return 'am'
	if (sectionStart <= 8) return 'pm'
	return 'evening'
}

function getSectionLabel(item) {
	if (!item.sectionStart) return '第-节'
	if (item.sectionEnd && item.sectionEnd !== item.sectionStart) {
		return `${item.sectionStart}-${item.sectionEnd}`
	}
	return `第${item.sectionStart}节`
}

const sortedClasses = computed(() => {
	const indexed = props.classes.map((item, i) => ({ ...item, _origIndex: i }))
	return indexed.sort((a, b) => {
		const aSection = parseSection(a.lessonScope || a.section || '').sectionStart || 99
		const bSection = parseSection(b.lessonScope || b.section || '').sectionStart || 99
		return aSection - bSection
	})
})

const normalizedClasses = computed(() => {
	return sortedClasses.value.map((item, mapIndex) => {
		const courseName = item.name || item.courseName || '未知课程'
		const teacherName = item.teacher || item.teacherNames || '未指定教师'
		const location = item.classroom || item.classroomName || '未知地点'
		const section = item.lessonScope || item.section || ''

		const { sectionStart, sectionEnd } = parseSection(section)
		const { startTime, endTime } = getLessonTimes(sectionStart)

		return {
			id: item.id || mapIndex,
			_origIndex: item._origIndex ?? mapIndex,
			courseName,
			teacherName,
			location,
			sectionStart,
			sectionEnd,
			startTime,
			endTime
		}
	})
})

const filteredClasses = computed(() => {
	if (!props.hidePast) return normalizedClasses.value
	const now = new Date()
	const currentMinutes = now.getHours() * 60 + now.getMinutes()
	return normalizedClasses.value.filter(item => {
		if (!item.endTime) return true
		const [h, m] = item.endTime.split(':').map(Number)
		if (isNaN(h)) return true
		const endMinutes = h * 60 + m
		return currentMinutes < endMinutes
	})
})

const groupedClasses = computed(() => {
	const groups = { am: [], pm: [], evening: [] }
	for (const item of filteredClasses.value) {
		const period = getPeriod(item.sectionStart)
		groups[period].push(item)
	}
	const result = []
	if (groups.am.length > 0) {
		result.push({ period: 'am', periodLabel: '上午', items: groups.am })
	}
	if (groups.pm.length > 0) {
		result.push({ period: 'pm', periodLabel: '下午', items: groups.pm })
	}
	if (groups.evening.length > 0) {
		result.push({ period: 'evening', periodLabel: '晚上', items: groups.evening })
	}
	return result
})

function goToSchedule() {
	uni.navigateTo({
		url: '/pages/schedule/index'
	})
}

function isCurrentClass(item) {
	if (!item.startTime || !item.endTime) return false
	const now = new Date()
	const currentTime = now.getHours() * 60 + now.getMinutes()
	const [startH, startM] = item.startTime.split(':').map(Number)
	const [endH, endM] = item.endTime.split(':').map(Number)
	if (isNaN(startH) || isNaN(endH)) return false
	const startTotal = startH * 60 + startM
	const endTotal = endH * 60 + endM
	return currentTime >= startTotal && currentTime <= endTotal
}

const now = ref(new Date())
const CLASS_STATUS_TEXTS = {
	inClass: '上课中',
	upcoming: '下节课',
	break: '休息中',
	finished: '今日课程已结束'
}

let clockTimer = null

function startClock() {
	stopClock()
	now.value = new Date()
	clockTimer = setInterval(() => {
		now.value = new Date()
	}, 10000)
}

function stopClock() {
	if (clockTimer) {
		clearInterval(clockTimer)
		clockTimer = null
	}
}

function timeToMinutes(time) {
	if (!time) return null
	const [h, m] = time.split(':').map(Number)
	if (isNaN(h) || isNaN(m)) return null
	return h * 60 + m
}

function formatCountdown(minutes) {
	if (minutes == null || !Number.isFinite(minutes)) return ''
	const rounded = Math.max(0, Math.round(minutes))
	if (rounded >= 60) {
		const h = Math.floor(rounded / 60)
		const m = rounded % 60
		return m > 0 ? `${h}小时${m}分钟后` : `${h}小时后`
	}
	return `${rounded}分钟后`
}

const classesWithOrder = computed(() =>
	normalizedClasses.value
		.filter(item => item.startTime && item.endTime)
		.slice()
)

const currentClass = computed(() => {
	const currentTotal = now.value.getHours() * 60 + now.value.getMinutes()
	return classesWithOrder.value.find(item => {
		const start = timeToMinutes(item.startTime)
		const end = timeToMinutes(item.endTime)
		return start != null && end != null && currentTotal >= start && currentTotal <= end
	}) || null
})

const nextClass = computed(() => {
	const currentTotal = now.value.getHours() * 60 + now.value.getMinutes()
	const upcoming = classesWithOrder.value.find(item => {
		const start = timeToMinutes(item.startTime)
		return start != null && start > currentTotal
	})
	return upcoming || null
})

const classStatus = computed(() => {
	const hasClasses = normalizedClasses.value.length > 0
	if (!hasClasses) return 'empty'
	if (currentClass.value) return 'inClass'
	if (nextClass.value) return 'upcoming'
	const lastEnd = classesWithOrder.value.reduce((max, item) => {
		const end = timeToMinutes(item.endTime)
		return end != null && end > max ? end : max
	}, -1)
	const currentTotal = now.value.getHours() * 60 + now.value.getMinutes()
	if (lastEnd >= 0 && currentTotal > lastEnd) return 'finished'
	return 'break'
})

const countdownText = computed(() => {
	if (classStatus.value === 'empty') return ''
	if (classStatus.value === 'finished') return CLASS_STATUS_TEXTS.finished
	if (classStatus.value === 'inClass' && currentClass.value) {
		const end = timeToMinutes(currentClass.value.endTime)
		if (end == null) return ''
		const diff = end - (now.value.getHours() * 60 + now.value.getMinutes())
		return `还有 ${formatCountdown(diff)}下课`
	}
	if (classStatus.value === 'upcoming' && nextClass.value) {
		const start = timeToMinutes(nextClass.value.startTime)
		if (start == null) return ''
		const diff = start - (now.value.getHours() * 60 + now.value.getMinutes())
		return `${CLASS_STATUS_TEXTS.upcoming}还有 ${formatCountdown(diff)}`
	}
	return ''
})

const classStatusLabel = computed(() => CLASS_STATUS_TEXTS[classStatus.value] || '')

onMounted(() => {
	startClock()
})

onBeforeUnmount(() => {
	stopClock()
})
</script>

<style lang="scss" scoped>
.today-class-container {
	// margin-bottom: var(--spacing-lg);
}

/* ========== Header ========== */
.header {
	display: flex;
	justify-content: space-between;
	align-items: flex-end;
	margin-bottom: var(--spacing-md);
	padding: var(--spacing-md);
	background: var(--bg-card);
	border-radius: var(--radius-xl);
	border: 1px solid var(--border-light);
	box-shadow: var(--shadow-xs);
}

.title-row {
	display: flex;
	flex-direction: column;
	gap: 4rpx;
}

.title {
	font-size: var(--font-size-xl);
	font-weight: 800;
	color: var(--text-primary);
	letter-spacing: -0.3px;
}

.date {
	font-size: var(--font-size-xs);
	color: var(--text-tertiary);
	font-weight: 500;
}

.view-schedule {
	display: flex;
	align-items: center;
	gap: 4rpx;
	font-size: var(--font-size-sm);
	color: var(--primary-color);
	font-weight: 600;

	.arrow-icon {
		font-size: 18px;
		font-weight: 300;
		line-height: 1;
		margin-top: -2rpx;
	}
}

/* ========== State Views ========== */
.state-view {
	background-color: var(--bg-card);
	border-radius: var(--radius-xl);
	// padding: var(--spacing-xl) var(--spacing-lg);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	border: 1px solid var(--border-light);
	box-shadow: var(--shadow-xs);
	color: var(--text-tertiary);
	font-size: var(--font-size-sm);
	min-height: 240rpx;
}

/* Loading */
.loading-ring {
	display: flex;
	gap: 8rpx;
	margin-bottom: var(--spacing-md);

	.ring-dot {
		width: 12rpx;
		height: 12rpx;
		border-radius: 50%;
		background: var(--primary-color);
		animation: pulse-dot 1.2s ease-in-out infinite;

		&:nth-child(2) { animation-delay: 0.2s; }
		&:nth-child(3) { animation-delay: 0.4s; }
	}
}

@keyframes pulse-dot {
	0%, 80%, 100% { transform: scale(0.6); opacity: 0.4; }
	40% { transform: scale(1); opacity: 1; }
}

/* Empty */
.empty-illustration {
	position: relative;
	width: 120rpx;
	height: 80rpx;
	margin-bottom: var(--spacing-lg);

	.empty-moon {
		width: 80rpx;
		height: 80rpx;
		border-radius: 50%;
		background: linear-gradient(135deg, #e0e7ff 0%, #c7d2fe 100%);
		position: absolute;
		top: 20rpx;
		left: 20rpx;
		box-shadow: 0 8rpx 24rpx rgba(99, 102, 241, 0.15);
	}
}

.empty-stars {
	position: absolute;
	width: 100%;
	height: 100%;

	.star {
		position: absolute;
		width: 8rpx;
		height: 8rpx;
		border-radius: 50%;
		background: var(--primary-color);
		opacity: 0.3;
		animation: twinkle 2s ease-in-out infinite;

		&-1 { top: 16rpx; right: 20rpx; animation-delay: 0s; }
		&-2 { top: 50rpx; right: 4rpx; animation-delay: 0.6s; width: 6rpx; height: 6rpx; }
		&-3 { top: 80rpx; right: 36rpx; animation-delay: 1.2s; width: 5rpx; height: 5rpx; }
	}
}

@keyframes twinkle {
	0%, 100% { opacity: 0.2; transform: scale(0.8); }
	50% { opacity: 0.6; transform: scale(1.2); }
}

.empty-main-text {
	font-size: var(--font-size-md);
	font-weight: 600;
	color: var(--text-secondary);
	margin-bottom: 4rpx;
}

.empty-sub-text {
	font-size: var(--font-size-sm);
	color: var(--text-tertiary);
}

/* ========== Status Bar ========== */
.class-status-bar {
	display: flex;
	align-items: center;
	padding: 16rpx 24rpx;
	border-radius: var(--radius-xl);
	background-color: var(--bg-card);
	border: 1px solid var(--border-light);
	gap: 12rpx;
	margin-bottom: var(--spacing-md);
	transition: all 0.3s var(--ease-out);

	&--inClass {
		border-color: rgba(37, 99, 235, 0.25);
		background: linear-gradient(135deg, rgba(37, 99, 235, 0.06), rgba(16, 185, 129, 0.04));
	}

	&--upcoming {
		border-color: rgba(245, 158, 11, 0.25);
		background: linear-gradient(135deg, rgba(245, 158, 11, 0.06), rgba(239, 68, 68, 0.03));
	}

	&--break {
		border-color: rgba(99, 102, 241, 0.2);
		background: rgba(99, 102, 241, 0.04);
	}

	&--finished,
	&--empty {
		border-color: var(--border-light);
	}
}

.status-dot {
	width: 14rpx;
	height: 14rpx;
	border-radius: 50%;
	flex-shrink: 0;

	&--inClass {
		background: linear-gradient(135deg, #4f46e5, #10b981);
		box-shadow: 0 0 8rpx rgba(79, 70, 229, 0.5);
		animation: glow-pulse 2s ease-in-out infinite;
	}

	&--upcoming {
		background: linear-gradient(135deg, #f59e0b, #ef4444);
		box-shadow: 0 0 8rpx rgba(245, 158, 11, 0.5);
	}

	&--break {
		background: #6366f1;
		box-shadow: 0 0 8rpx rgba(99, 102, 241, 0.4);
	}

	&--finished,
	&--empty {
		background: var(--text-tertiary);
	}
}

@keyframes glow-pulse {
	0%, 100% { box-shadow: 0 0 8rpx rgba(79, 70, 229, 0.4); }
	50% { box-shadow: 0 0 16rpx rgba(79, 70, 229, 0.7); }
}

.status-badge-text {
	font-size: 26rpx;
	font-weight: 700;
	color: var(--text-primary);
	flex-shrink: 0;
}

.status-divider {
	font-size: 24rpx;
	color: var(--border-primary);
	margin: 0 4rpx;
}

.status-countdown {
	font-size: 24rpx;
	font-weight: 600;
	color: var(--text-secondary);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

/* ========== Horizontal Layout ========== */
.class-scroll {
	width: 100%;
	white-space: nowrap;
}

.class-list {
	display: inline-flex;
	gap: var(--spacing-md);
	padding: 4rpx var(--spacing-md) var(--spacing-md);
	vertical-align: top;
}

.class-card {
	flex-shrink: 0;
	width: 480rpx;
	background-color: var(--bg-card);
	border-radius: var(--radius-xl);
	display: flex;
	overflow: hidden;
	box-shadow: var(--shadow-sm);
	border: 1px solid var(--border-light);
	transition: all 0.25s var(--ease-out);
	position: relative;

	&.is-active {
		border-color: var(--primary-color);
		box-shadow: 0 0 0 3rpx rgba(37, 99, 235, 0.15), var(--shadow-md);
		transform: translateY(-2rpx);
	}

	.card-left {
		width: 130rpx;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		color: #fff;
		font-size: 24rpx;
		font-weight: 700;
		flex-shrink: 0;
		position: relative;
		overflow: hidden;

		&::after {
			content: '';
			position: absolute;
			bottom: -20rpx;
			right: -20rpx;
			width: 60rpx;
			height: 60rpx;
			border-radius: 50%;
			background: rgba(255, 255, 255, 0.1);
		}

		.time-divider {
			width: 40rpx;
			height: 2rpx;
			background-color: rgba(255, 255, 255, 0.4);
			margin: 6rpx 0;
			border-radius: 1rpx;
		}
	}

	.card-right {
		flex: 1;
		padding: 20rpx 20rpx 16rpx;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		min-height: 140rpx;

		.card-right-top {
			display: flex;
			align-items: flex-start;
			justify-content: space-between;
			gap: 12rpx;
		}

		.course-name {
			font-size: 30rpx;
			font-weight: 700;
			color: var(--text-primary);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			flex: 1;
			line-height: 1.3;
		}

		.course-tag {
			font-size: 20rpx;
			font-weight: 700;
			padding: 4rpx 12rpx;
			border-radius: var(--radius-full);
			flex-shrink: 0;
			white-space: nowrap;
		}

		.course-info {
			display: flex;
			flex-direction: column;
			gap: 6rpx;
			margin-top: 4rpx;
		}

		.info-item {
			display: flex;
			align-items: center;
			gap: 6rpx;
			font-size: 23rpx;
			color: var(--text-secondary);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;

			.info-icon {
				font-size: 20rpx;
				opacity: 0.5;
				flex-shrink: 0;
			}
		}
	}

	.active-indicator {
		position: absolute;
		top: 0;
		right: 0;
		width: 0;
		height: 0;
		border-top: 28rpx solid var(--primary-color);
		border-left: 28rpx solid transparent;
	}
}

/* ========== Vertical Layout ========== */
.class-list-vertical {
	display: flex;
	flex-direction: column;
	gap: var(--spacing-md);
}

.period-group {
	display: flex;
	flex-direction: column;
	gap: 12rpx;
}

.period-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 4rpx;
}

.period-label {
	display: flex;
	align-items: center;
	gap: 8rpx;
	padding: 6rpx 20rpx;
	border-radius: var(--radius-full);
	font-size: 24rpx;
	font-weight: 800;
	color: #fff;
	box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);

	.period-icon {
		font-size: 20rpx;
		line-height: 1;
	}

	&--am { background: linear-gradient(135deg, #f59e0b, #fb923c); }
	&--pm { background: linear-gradient(135deg, #10b981, #34d399); }
	&--evening { background: linear-gradient(135deg, #6366f1, #8b5cf6); }
}

.period-meta {
	display: flex;
	align-items: center;
	gap: 6rpx;

	.period-dot {
		width: 4rpx;
		height: 4rpx;
		border-radius: 50%;
		background: var(--text-tertiary);
	}

	.period-count {
		font-size: 22rpx;
		color: var(--text-tertiary);
		font-weight: 500;
	}
}

.period-courses {
	display: flex;
	flex-direction: column;
	gap: 12rpx;
}

.class-card-vertical {
	display: flex;
	align-items: stretch;
	background-color: var(--bg-card);
	border-radius: var(--radius-xl);
	overflow: hidden;
	box-shadow: var(--shadow-sm);
	border: 1px solid var(--border-light);
	transition: all 0.25s var(--ease-out);
	min-height: 132rpx;
	position: relative;

	&.is-active {
		border-color: var(--primary-color);
		box-shadow: 0 0 0 3rpx rgba(37, 99, 235, 0.12), var(--shadow-md);
		transform: translateX(4rpx);

		&::before {
			content: '';
			position: absolute;
			left: 0;
			top: 0;
			bottom: 0;
			width: 4rpx;
			background: var(--primary-color);
			border-radius: 4rpx 0 0 4rpx;
		}
	}

	.card-vertical-time {
		width: 152rpx;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 16rpx 0;
		color: #fff;
		font-size: 22rpx;
		font-weight: 700;
		flex-shrink: 0;
		position: relative;
		overflow: hidden;

		&::after {
			content: '';
			position: absolute;
			top: -30rpx;
			right: -30rpx;
			width: 80rpx;
			height: 80rpx;
			border-radius: 50%;
			background: rgba(255, 255, 255, 0.08);
		}

		.time-lesson-num {
			font-size: 24rpx;
			font-weight: 800;
			line-height: 1.2;
			letter-spacing: 0.5px;
		}

		.time-range-v {
			display: flex;
			flex-direction: column;
			align-items: center;
			gap: 0;
			width: 100%;
		}

		.time-divider-v {
			width: 36rpx;
			height: 2rpx;
			background-color: rgba(255, 255, 255, 0.4);
			margin: 6rpx 0;
			border-radius: 1rpx;
		}

		.time-start-v,
		.time-end-v {
			font-size: 20rpx;
			font-weight: 600;
			line-height: 1.2;
		}

		.time-sep {
			font-size: 16rpx;
			opacity: 0.6;
			line-height: 1;
		}
	}

	.card-vertical-body {
		flex: 1;
		padding: 16rpx 20rpx;
		display: flex;
		flex-direction: column;
		justify-content: center;
		gap: 8rpx;
		min-height: 132rpx;

		.card-vertical-header {
			display: flex;
			align-items: center;
			gap: 12rpx;
		}

		.course-name-v {
			font-size: 31rpx;
			font-weight: 700;
			color: var(--text-primary);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			line-height: 1.3;
		}

		.live-badge {
			font-size: 18rpx;
			font-weight: 700;
			color: #fff;
			background: linear-gradient(135deg, var(--primary-color), #7c3aed);
			padding: 3rpx 12rpx;
			border-radius: var(--radius-full);
			flex-shrink: 0;
			animation: live-pulse 2s ease-in-out infinite;
		}

		.course-info-v {
			display: flex;
			gap: 24rpx;
			overflow: hidden;
		}

		.info-item-v {
			display: flex;
			align-items: center;
			gap: 6rpx;
			font-size: 23rpx;
			color: var(--text-secondary);
			white-space: nowrap;
			flex-shrink: 0;

			.info-icon-v {
				font-size: 20rpx;
				opacity: 0.45;
				flex-shrink: 0;
			}
		}
	}
}

@keyframes live-pulse {
	0%, 100% { opacity: 1; }
	50% { opacity: 0.7; }
}
</style>
