<template>
	<view class="survey-detail-page">
		<!-- 自定义导航栏 -->
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">{{ pageTitle }}</text>
				<view class="nav-bar__placeholder"></view>
			</view>
		</view>

		<!-- 内容区 -->
		<scroll-view class="content-scroll" scroll-y>
			<!-- 加载中 -->
			<view v-if="loading" class="state-container">
				<view class="state-spinner"></view>
				<text class="state-text">加载中...</text>
			</view>

			<!-- 错误状态 -->
			<view v-else-if="error" class="state-container state-container--error">
				<view class="state-icon-circle state-icon-circle--error">!</view>
				<text class="state-text">{{ error }}</text>
				<button class="state-btn" @tap="fetchQuestions">重试</button>
			</view>

			<!-- 问卷内容 -->
			<view v-else class="survey-content">
				<!-- 问卷描述 -->
				<view class="survey-intro">
					<text class="survey-intro__title">{{ pageTitle }}</text>
				</view>

				<!-- 题目列表 -->
				<view class="question-list">
					<view
						v-for="(question, qIndex) in questions"
						:key="question.id"
						class="question-card"
					>
						<view class="question-card__header">
							<text class="question-card__index">{{ qIndex + 1 }}.</text>
							<text class="question-card__title">{{ getQuestionTitle(question) }}</text>
							<text class="question-card__required" v-if="question.isrequired === 1">*</text>
						</view>

						<!-- 题目描述 -->
						<view class="question-card__desc" v-if="getQuestionDesc(question)">
							<text class="question-card__desc-text">{{ getQuestionDesc(question) }}</text>
						</view>

						<!-- 单选题 -->
						<radio-group
							v-if="question.type === 0"
							class="question-card__options"
							@change="onSingleChoiceChange(qIndex, $event.detail.value)"
						>
							<label
								v-for="option in question.toaSurveyQuestionOption"
								:key="option.id"
								class="option-item"
								:class="{'option-item--selected': answers[qIndex] === option.id}"
							>
								<radio :value="String(option.id)" :checked="answers[qIndex] === option.id" />
								<text class="option-item__text">{{ option.name }}</text>
							</label>
						</radio-group>

						<!-- 多选题 -->
						<checkbox-group
							v-else-if="question.type === 1"
							class="question-card__options"
							@change="onMultiChoiceChange(qIndex, $event.detail.value)"
						>
							<label
								v-for="option in question.toaSurveyQuestionOption"
								:key="option.id"
								class="option-item"
								:class="{'option-item--selected': (answers[qIndex] || []).includes(option.id)}"
							>
								<checkbox :value="String(option.id)" :checked="(answers[qIndex] || []).includes(option.id)" />
								<text class="option-item__text">{{ option.name }}</text>
							</label>
						</checkbox-group>

						<!-- 填空题 -->
						<view v-else-if="question.type === 2" class="question-card__input">
							<textarea
								class="input-field"
								v-model="answers[qIndex]"
								:placeholder="'请输入您的答案'"
								maxlength="500"
							/>
						</view>
					</view>
				</view>

				<!-- 提交按钮 -->
				<view class="submit-area">
					<button
						class="submit-btn"
						:disabled="submitting"
						@tap="submitSurvey"
					>
						<view v-if="submitting" class="submit-btn__spinner"></view>
						<text>{{ submitting ? '提交中...' : '提交问卷' }}</text>
					</button>
				</view>

				<view class="bottom-safe-area"></view>
			</view>
		</scroll-view>
	</view>
</template>

<script>
import { getSurveyQuestions } from '../../pages/api/discover.js';

export default {
	data() {
		return {
			statusBarHeight: 20,
			pageTitle: '问卷详情',
			surveyId: '',
			questions: [],
			answers: {},
			loading: false,
			submitting: false,
			error: ''
		}
	},
	onLoad(options) {
		const systemInfo = uni.getSystemInfoSync();
		this.statusBarHeight = systemInfo.statusBarHeight || 20;
		if (options.id) {
			this.surveyId = options.id;
		}
		if (options.title) {
			this.pageTitle = decodeURIComponent(options.title);
		}
		this.fetchQuestions();
	},
	methods: {
		goBack() {
			uni.navigateBack();
		},
		async fetchQuestions() {
			if (!this.surveyId) return;
			this.loading = true;
			this.error = '';
			try {
				const res = await getSurveyQuestions(this.surveyId);
				if (Array.isArray(res)) {
					this.questions = res;
				} else if (res && res.result) {
					this.questions = Array.isArray(res.result) ? res.result : [];
				} else {
					this.questions = [];
				}
			} catch (e) {
				console.error('获取问卷问题失败', e);
				this.error = e.message || '获取问卷问题失败';
			} finally {
				this.loading = false;
			}
		},
		getQuestionTitle(question) {
			if (question.title && question.title.trim()) {
				return question.title;
			}
			return '';
		},
		getQuestionDesc(question) {
			if (question.description && question.description.trim()) {
				return question.description.trim();
			}
			return '';
		},
		onSingleChoiceChange(qIndex, value) {
			this.$set(this.answers, qIndex, Number(value));
		},
		onMultiChoiceChange(qIndex, values) {
			this.$set(this.answers, qIndex, values.map(v => Number(v)));
		},
		async submitSurvey() {
			uni.showToast({ title: '问卷提交功能开发中', icon: 'none' });
		}
	}
}
</script>

<style lang="scss" scoped>
@import url('../../static/css/home-page.css');

.survey-detail-page {
	width: 100%;
	height: 100vh;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
}

.nav-bar {
	background: linear-gradient(135deg, #8b5cf6, #a78bfa);
	flex-shrink: 0;
}

.nav-bar__content {
	display: flex;
	align-items: center;
	justify-content: space-between;
	height: 88rpx;
	padding: 0 24rpx;
}

.nav-bar__back,
.nav-bar__placeholder {
	width: 60rpx;
	height: 60rpx;
	display: flex;
	align-items: center;
}

.nav-bar__title {
	font-size: 32rpx;
	font-weight: 700;
	color: #fff;
	text-align: center;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
	max-width: 400rpx;
}

.content-scroll {
	flex: 1;
	height: 0;
}

.state-container {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 120rpx 0;
	gap: 16rpx;
}

.state-spinner {
	width: 48rpx;
	height: 48rpx;
	border: 4rpx solid rgba(139, 92, 246, 0.15);
	border-top-color: #8b5cf6;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

@keyframes spin {
	to { transform: rotate(360deg); }
}

.state-text {
	font-size: 28rpx;
	color: var(--text-secondary);
	font-weight: 600;
}

.state-container--error {
	gap: 12rpx;
}

.state-icon-circle {
	width: 72rpx;
	height: 72rpx;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 36rpx;
	font-weight: 800;
}

.state-icon-circle--error {
	background: rgba(239, 68, 68, 0.1);
	color: #ef4444;
}

.state-btn {
	margin-top: 8rpx;
	padding: 12rpx 40rpx;
	background: linear-gradient(135deg, #8b5cf6, #a78bfa);
	color: #fff;
	font-size: 26rpx;
	font-weight: 600;
	border-radius: 32rpx;
	border: none;
}

.survey-content {
	padding: 24rpx;
}

.survey-intro {
	background: #fff;
	border-radius: 24rpx;
	padding: 32rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 2rpx 12rpx rgba(139, 92, 246, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.survey-intro__title {
	font-size: 34rpx;
	font-weight: 800;
	color: var(--text-primary);
	line-height: 1.4;
}

.question-list {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
	margin-bottom: 24rpx;
}

.question-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 28rpx;
	box-shadow: 0 2rpx 12rpx rgba(139, 92, 246, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.question-card__header {
	display: flex;
	align-items: flex-start;
	margin-bottom: 8rpx;
}

.question-card__index {
	font-size: 30rpx;
	font-weight: 700;
	color: #8b5cf6;
	margin-right: 8rpx;
	flex-shrink: 0;
}

.question-card__title {
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	flex: 1;
	line-height: 1.4;
}

.question-card__required {
	font-size: 30rpx;
	font-weight: 700;
	color: #ef4444;
	flex-shrink: 0;
}

.question-card__desc {
	margin-bottom: 16rpx;
	padding: 12rpx 16rpx;
	background: var(--bg-secondary);
	border-radius: 12rpx;
}

.question-card__desc-text {
	font-size: 24rpx;
	color: var(--text-secondary);
	line-height: 1.5;
}

.question-card__options {
	display: flex;
	flex-direction: column;
	gap: 8rpx;
}

.option-item {
	display: flex;
	align-items: center;
	padding: 16rpx 20rpx;
	border-radius: 16rpx;
	background: var(--bg-secondary);
	transition: all 0.15s ease;
}

.option-item--selected {
	background: rgba(139, 92, 246, 0.1);
}

.option-item__text {
	margin-left: 12rpx;
	font-size: 28rpx;
	color: var(--text-primary);
}

.question-card__input {
	margin-top: 8rpx;
}

.input-field {
	width: 100%;
	min-height: 160rpx;
	padding: 20rpx;
	background: var(--bg-secondary);
	border-radius: 16rpx;
	font-size: 28rpx;
	color: var(--text-primary);
	box-sizing: border-box;
}

.submit-area {
	padding: 16rpx 0;
}

.submit-btn {
	width: 100%;
	height: 88rpx;
	background: linear-gradient(135deg, #8b5cf6, #a78bfa);
	color: #fff;
	font-size: 30rpx;
	font-weight: 700;
	border-radius: 44rpx;
	border: none;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 6rpx 24rpx rgba(139, 92, 246, 0.3);
}

.submit-btn[disabled] {
	opacity: 0.7;
}

.submit-btn__spinner {
	width: 32rpx;
	height: 32rpx;
	border: 3rpx solid rgba(255, 255, 255, 0.3);
	border-top-color: #fff;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
	margin-right: 12rpx;
}

.bottom-safe-area {
	height: env(safe-area-inset-bottom);
}
</style>
