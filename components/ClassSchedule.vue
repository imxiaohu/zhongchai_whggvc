<template>
	<view class="class-schedule">
		<!-- 周次选择器 - 毛玻璃卡片 -->
		<view class="week-selector">
			<view class="week-btn" @tap="prevWeek">
				<l-icon name="chevron-left" size="20" color="var(--text-primary)"></l-icon>
			</view>
			<view class="week-info" @tap="openWeekSelector">
				<text class="week-title">{{ displayWeek > currentWeek ? `第${displayWeek}周` : displayWeek === currentWeek ? '本周' : `第${displayWeek}周` }}</text>
				<text v-if="startDate" class="week-date">{{ startDate }} - {{ endDate }}</text>
				<text v-if="displayWeek !== currentWeek" class="dropdown-badge">点击回到本周</text>
			</view>
			<view class="week-btn" @tap="nextWeek">
				<l-icon name="chevron-right" size="20" color="var(--text-primary)"></l-icon>
			</view>
		</view>

		<!-- 课表容器 -->
		<view class="schedule-container">
			<!-- 星期头部 - 毛玻璃 -->
			<view class="weekday-header">
				<view class="corner-cell">
					<text class="month-text">{{ currentMonth }}月</text>
					<text class="unit-text">节</text>
				</view>
				<view
					v-for="(day, index) in weekdays"
					:key="index"
					class="day-header"
					:class="{ 'current-day': isCurrentDay(index) }"
				>
					<text class="day-name">{{ day.name }}</text>
					<view class="date-box">
						<text class="day-date">{{ formatDayOnly(day.date) }}</text>
					</view>
				</view>
			</view>

			<!-- 课表主体 -->
			<scroll-view class="timetable-scroll" scroll-y @touchstart="handleTouchStart" @touchmove="handleTouchMove" @touchend="handleTouchEnd">
				<view class="timetable">
					<!-- 背景网格 -->
					<view class="grid-background">
						<view v-for="i in 12" :key="i" class="grid-row"></view>
					</view>

					<!-- 时间列 -->
					<view class="time-column">
						<view v-for="i in 12" :key="i" class="time-cell">
							<text class="period-num">{{ i }}</text>
							<text v-if="timePeriods[i-1]" class="start-time">{{ timePeriods[i-1].startTime || timePeriods[i-1].start }}</text>
						</view>
					</view>

				<!-- 课程展示区域 -->
				<view class="courses-area">
					<!-- 当前日列高亮背景 -->
					<view
						v-if="currentDayIndex >= 0"
						class="current-day-column"
						:style="{
							left: `${(currentDayIndex * 100) / 7}%`,
							width: `${100 / 7}%`
						}"
					></view>

					<template v-for="(day, dayIndex) in weekdays">
						<view
							v-for="course in getCoursesForDay(dayIndex)"
							:key="`course-${course.id}-${dayIndex}`"
							class="course-card-wrapper"
							:class="{ 'current-day-card': currentDayIndex === dayIndex }"
							:style="getCourseStyle(course, dayIndex)"
							@click="showCourseDetail(course)"
						>
								<view
									class="course-card"
									:style="{ backgroundColor: getCourseColor(course.courseName) }"
									:class="{'pressed': isPressedCourse(course.id)}"
								>
									<view class="card-inner">
										<text class="course-title">{{course.courseName}}</text>
										<view class="course-meta">
											<view class="meta-item" v-if="course.classroomName">
												<l-icon name="location" size="10" color="rgba(255,255,255,0.8)"></l-icon>
												<text class="meta-item-text">{{formatLocation(course.classroomName)}}</text>
											</view>
											<view class="meta-item" v-if="course.teacherNames">
												<l-icon name="user" size="10" color="rgba(255,255,255,0.8)"></l-icon>
												<text class="meta-item-text">{{formatTeacher(course.teacherNames)}}</text>
											</view>
										</view>
									</view>
								</view>
							</view>
						</template>
					</view>
				</view>
			</scroll-view>
		</view>

		<!-- 课程详情弹窗 -->
		<view class="course-modal" v-if="showModal" @click="showModal = false">
			<view class="modal-content" @click.stop>
				<view class="modal-header" :style="{ backgroundColor: getCourseColor(selectedCourse.courseName) }">
					<view class="header-icon">
						<l-icon name="book-open" size="30" color="#fff"></l-icon>
					</view>
					<text class="modal-title">{{selectedCourse.courseName}}</text>
				</view>
				<view class="modal-body">
					<view class="detail-item">
						<view class="detail-icon"><l-icon name="user" size="18" color="var(--text-tertiary)"></l-icon></view>
						<view class="detail-content">
							<text class="detail-label">课程编号</text>
							<text class="detail-value">{{selectedCourse.courseCode || '-'}}</text>
						</view>
					</view>
					<view class="detail-item">
						<view class="detail-icon"><l-icon name="user" size="18" color="var(--text-tertiary)"></l-icon></view>
						<view class="detail-content">
							<text class="detail-label">授课教师</text>
							<text class="detail-value">{{selectedCourse.teacherNames || '-'}}</text>
						</view>
					</view>
					<view class="detail-item">
						<view class="detail-icon"><l-icon name="location" size="18" color="var(--text-tertiary)"></l-icon></view>
						<view class="detail-content">
							<text class="detail-label">上课地点</text>
							<text class="detail-value">{{selectedCourse.classroomName || '-'}}</text>
						</view>
					</view>
					<view class="detail-item">
						<view class="detail-icon"><l-icon name="time" size="18" color="var(--text-tertiary)"></l-icon></view>
						<view class="detail-content">
							<text class="detail-label">上课时间</text>
							<text class="detail-value">第 {{formatLessonScope(selectedCourse.lessonScope)}} 节</text>
						</view>
					</view>
				</view>
				<view class="modal-footer">
					<button class="modal-close" @click="showModal = false">我知道了</button>
				</view>
			</view>
		</view>

		<!-- 周数选择器弹窗 -->
		<week-selector
			:visible="showWeekSelector"
			:current-week="currentWeek"
			:display-week="displayWeek"
			:total-weeks="totalWeeks"
			:semester-info="semesterInfo"
			@close="showWeekSelector = false"
			@select="onWeekSelected"
		></week-selector>
	</view>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import WeekSelector from './WeekSelector.vue'

const props = defineProps({
	courses: { type: Array, default: () => [] },
	timePeriods: { type: Array, default: () => [] },
	currentWeek: { type: Number, default: 1 },
	displayWeek: { type: Number, default: 1 },
	totalWeeks: { type: Number, default: 20 },
	semesterInfo: { type: Object, default: () => ({}) }
})

const emit = defineEmits(['prev-week', 'next-week', 'week-change'])

const weekdays = reactive([
	{ name: '周一', date: null }, { name: '周二', date: null },
	{ name: '周三', date: null }, { name: '周四', date: null },
	{ name: '周五', date: null }, { name: '周六', date: null },
	{ name: '周日', date: null }
])

const selectedCourse = ref({})
const showModal = ref(false)
const showWeekSelector = ref(false)
const pressedCourseId = ref(null)

const currentDayIndex = computed(() => {
	const today = new Date()
	const d = weekdays[0].date
	if (!d || today.toDateString() !== d.toDateString() || props.displayWeek !== props.currentWeek) {
		return -1
	}
	return today.getDay() === 0 ? 6 : today.getDay() - 1
})

const colorPool = [
	'#4f46e5', '#10b981', '#f59e0b', '#ef4444',
	'#0ea5e9', '#8b5cf6', '#ec4899', '#06b6d4',
	'#f97316', '#84cc16'
]
const courseColorMap = reactive({})

let touchStartX = 0
let touchStartY = 0
const minSwipeDistance = 80
let isSwipeInProgress = false

const startDate = computed(() => {
	if (!weekdays[0].date) return ''
	const d = weekdays[0].date
	return `${d.getMonth() + 1}.${d.getDate()}`
})

const endDate = computed(() => {
	if (!weekdays[6].date) return ''
	const d = weekdays[6].date
	return `${d.getMonth() + 1}.${d.getDate()}`
})

const currentMonth = computed(() => {
	if (!weekdays[0].date) return ''
	return weekdays[0].date.getMonth() + 1
})

watch(() => props.displayWeek, () => { updateWeekDates() }, { immediate: true })
watch(() => props.courses, () => {
	assignCourseColors()
}, { immediate: true })

function updateWeekDates() {
	const now = new Date()
	const currentDay = now.getDay() || 7
	const mondayOffset = props.displayWeek - props.currentWeek
	const monday = new Date(now)
	monday.setDate(now.getDate() - currentDay + 1 + (mondayOffset * 7))

	weekdays.forEach((day, index) => {
		const date = new Date(monday)
		date.setDate(monday.getDate() + index)
		weekdays[index].date = date
	})
}

function assignCourseColors() {
	let colorIdx = 0
	props.courses.forEach(course => {
		if (!courseColorMap[course.courseName]) {
			courseColorMap[course.courseName] = colorPool[colorIdx % colorPool.length]
			colorIdx++
		}
	})
}

function getCourseColor(name) {
	return courseColorMap[name] || colorPool[0]
}

function getCoursesForDay(dayIndex) {
	const weekday = ['一', '二', '三', '四', '五', '六', '日'][dayIndex]
	return props.courses.filter(c => c.week === weekday)
}

function getCourseStyle(course, dayIndex) {
	const lessons = parseLessonScope(course.lessonScope)
	if (lessons.length === 0) return {}

	const start = Math.min(...lessons)
	const duration = Math.max(...lessons) - start + 1

	return {
		left: `${(dayIndex * 100) / 7}%`,
		top: `${(start - 1) * 120}rpx`,
		width: `${100 / 7}%`,
		height: `${duration * 120}rpx`
	}
}

function parseLessonScope(scope) {
	const matches = (scope || '').match(/\d+/g)
	return matches ? matches.map(Number) : []
}

function formatLessonScope(scope) {
	const nums = parseLessonScope(scope).sort((a, b) => a - b)
	if (nums.length <= 1) return nums[0] || ''
	return `${nums[0]}-${nums[nums.length - 1]}`
}

function formatDayOnly(date) {
	return date ? date.getDate() : ''
}

function formatLocation(loc) {
	if (!loc) return ''
	const parts = loc.split(' ')
	return parts[parts.length - 1]
}

function formatTeacher(name) {
	if (!name) return ''
	return name.split(',')[0]
}

function isCurrentDay(index) {
	const today = new Date()
	const d = weekdays[index].date
	return d && today.toDateString() === d.toDateString() && props.displayWeek === props.currentWeek
}

function showCourseDetail(course) {
	selectedCourse.value = course
	showModal.value = true
}

function prevWeek() {
	if (props.displayWeek > 1) emit('prev-week')
}

function nextWeek() {
	if (props.displayWeek < props.totalWeeks) emit('next-week')
}

function openWeekSelector() {
	showWeekSelector.value = true
}

function onWeekSelected(week) {
	emit('week-change', week)
}

function isPressedCourse(id) {
	return pressedCourseId.value === id
}

function handleTouchStart(e) {
	touchStartX = e.touches[0].clientX
	touchStartY = e.touches[0].clientY
	isSwipeInProgress = false
}

function handleTouchMove(e) {
	if (isSwipeInProgress) return
	const deltaX = e.touches[0].clientX - touchStartX
	const deltaY = e.touches[0].clientY - touchStartY
	if (Math.abs(deltaX) > minSwipeDistance && Math.abs(deltaX) > Math.abs(deltaY)) {
		isSwipeInProgress = true
		if (deltaX > 0) prevWeek()
		else nextWeek()
	}
}

function handleTouchEnd() {
	isSwipeInProgress = false
}
</script>

<style lang="scss" scoped>
.class-schedule {
	display: flex;
	flex-direction: column;
	height: 100%;
	background-color: var(--bg-secondary);
}

.week-selector {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: var(--spacing-md) var(--spacing-lg);
	background: rgba(255, 255, 255, 0.78);
	backdrop-filter: blur(20px) saturate(180%);
	-webkit-backdrop-filter: blur(20px) saturate(180%);
	// border-radius: var(--radius-xl);
	// margin: 0 var(--spacing-md);
	// margin-top: var(--spacing-md);
	box-shadow:
		0 4px 16px rgba(99, 102, 241, 0.06),
		0 8px 32px rgba(0, 0, 0, 0.04),
		inset 0 1px 0 rgba(255, 255, 255, 0.9);
	border: 1px solid rgba(255, 255, 255, 0.6);

	.week-btn {
		width: 36px;
		height: 36px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 50%;
		background: var(--bg-secondary);
		transition: all 0.15s var(--ease-out);

		&:active {
			background: var(--bg-muted);
			transform: scale(0.92);
		}
	}

	.week-info {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 0 var(--spacing-md);

		.week-title {
			font-size: 15px;
			font-weight: 700;
			color: var(--text-primary);
		}

		.week-date {
			font-size: 12px;
			color: var(--text-tertiary);
			margin-top: 2px;
		}

		.dropdown-badge {
			display: inline-block;
			background: var(--warning-soft);
			padding: 2px 8px;
			border-radius: 20px;
			font-size: 11px;
			color: var(--warning-color);
			font-weight: 600;
			margin-top: 4px;
		}
	}
}

.schedule-container {
	flex: 1;
	display: flex;
	flex-direction: column;
	overflow: hidden;
}

.weekday-header {
	display: flex;
	background: rgba(255, 255, 255, 0.78);
	backdrop-filter: blur(20px) saturate(180%);
	-webkit-backdrop-filter: blur(20px) saturate(180%);
	margin-bottom: 0;
	box-shadow:
		0 4px 16px rgba(99, 102, 241, 0.06),
		0 8px 32px rgba(0, 0, 0, 0.04),
		inset 0 1px 0 rgba(255, 255, 255, 0.9);
	border: 1px solid rgba(255, 255, 255, 0.6);
	overflow: hidden;

	.corner-cell {
		width: 60rpx;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		border-right: 1px solid var(--border-secondary);

		.month-text {
			font-size: 22rpx;
			font-weight: 700;
			color: var(--text-primary);
		}

		.unit-text {
			font-size: 16rpx;
			color: var(--text-tertiary);
		}
	}

	.day-header {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 4rpx;
		padding: 10rpx 0;
		transition: background-color 0.15s;

		&.current-day {
			background-color: var(--primary-soft);

			.day-name {
				color: var(--primary-600);
			}

			.date-box {
				background-color: var(--primary-600);

				.day-date {
					color: #ffffff;
				}
			}
		}

		.day-name {
			font-size: 22rpx;
			color: var(--text-secondary);
			font-weight: 600;
		}

		.date-box {
			width: 36rpx;
			height: 36rpx;
			display: flex;
			align-items: center;
			justify-content: center;
			border-radius: 50%;

			.day-date {
				font-size: 20rpx;
				color: var(--text-tertiary);
				font-weight: 600;
			}
		}
	}
}

.timetable-scroll {
	flex: 1;
	min-height: 600rpx;
}

.timetable {
	position: relative;
	display: flex;
	height: calc(120rpx * 12);
	min-height: calc(120rpx * 12);
}

.grid-background {
	position: absolute;
	top: 0; left: 60rpx; right: 0; bottom: 0;
	z-index: 0;

	.grid-row {
		height: 120rpx;
		border-bottom: 1px dashed var(--border-secondary);
	}
}

.time-column {
	width: 60rpx;
	z-index: 1;

	.time-cell {
		height: 120rpx;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		border-right: 1px solid var(--border-secondary);
		background-color: rgba(255, 255, 255, 0.78);
		backdrop-filter: blur(10px);

		.period-num {
			font-size: 22rpx;
			font-weight: 600;
			color: var(--text-secondary);
		}

		.start-time {
			font-size: 16rpx;
			color: var(--text-tertiary);
			margin-top: 2rpx;
		}
	}
}

.courses-area {
	flex: 1;
	position: relative;
	z-index: 2;
}

.current-day-column {
	position: absolute;
	top: 0;
	bottom: 0;
	background-color: var(--primary-soft);
	pointer-events: none;
	z-index: 1;
}

.course-card-wrapper {
	position: absolute;
	padding: 4rpx;
	box-sizing: border-box;

	&.current-day-card {
		z-index: 3;

		.course-card {
			box-shadow: 0 4rpx 16rpx rgba(37, 99, 235, 0.3);
		}
	}
}

.course-card {
	width: 100%;
	height: 100%;
	border-radius: 10rpx;
	padding: 8rpx;
	box-sizing: border-box;
	box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.12);
	transition: all 0.15s var(--ease-out);

	&.pressed {
		transform: scale(0.95);
		opacity: 0.8;
	}

	.card-inner {
		height: 100%;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
	}

	.course-title {
		font-size: 20rpx;
		font-weight: 700;
		color: #ffffff;
		line-height: 1.3;
		display: -webkit-box;
		-webkit-box-orient: vertical;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		overflow: hidden;
	}

	.course-meta {
		.meta-item {
			display: flex;
			align-items: center;
			gap: 4rpx;

			.meta-item-text {
				font-size: 16rpx;
				color: rgba(255, 255, 255, 0.9);
				white-space: nowrap;
				overflow: hidden;
				text-overflow: ellipsis;
			}
		}
	}
}

.course-modal {
	position: fixed;
	top: 0; left: 0; right: 0; bottom: 0;
	background: rgba(0, 0, 0, 0.45);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 999;
	backdrop-filter: blur(8px);

	.modal-content {
		width: 80%;
		max-width: 600rpx;
		background: rgba(255, 255, 255, 0.95);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border-radius: var(--radius-xl);
		overflow: hidden;
		box-shadow:
			0 24rpx 60rpx rgba(0, 0, 0, 0.18),
			inset 0 1px 0 rgba(255, 255, 255, 0.9);
		border: 1px solid rgba(255, 255, 255, 0.6);
	}

	.modal-header {
		padding: var(--spacing-xl) var(--spacing-lg);
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;

		.header-icon {
			margin-bottom: var(--spacing-md);
		}

		.modal-title {
			font-size: 18px;
			font-weight: 700;
			color: #ffffff;
		}
	}

	.modal-body {
		padding: var(--spacing-lg);

		.detail-item {
			display: flex;
			gap: var(--spacing-md);
			margin-bottom: var(--spacing-lg);

			&:last-child {
				margin-bottom: 0;
			}

			.detail-content {
				display: flex;
				flex-direction: column;

				.detail-label {
					font-size: 12px;
					color: var(--text-tertiary);
					margin-bottom: 4rpx;
				}

				.detail-value {
					font-size: 15px;
					color: var(--text-primary);
					font-weight: 500;
				}
			}
		}
	}

	.modal-footer {
		padding: var(--spacing-md) var(--spacing-lg);
		border-top: 1px solid var(--border-secondary);

		.modal-close {
			width: 100%;
			height: 44px;
			background: var(--bg-secondary);
			color: var(--text-secondary);
			border-radius: var(--radius-md);
			font-size: 14px;
			font-weight: 600;
			border: none;
			display: flex;
			align-items: center;
			justify-content: center;

			&::after {
				border: none;
			}

			&:active {
				background: var(--bg-muted);
			}
		}
	}
}
</style>
