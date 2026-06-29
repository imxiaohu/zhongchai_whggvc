<template>
	<!-- 只有在visible为true时才渲染整个组件 -->
	<view v-if="visible" class="simple-report-modal">
		<!-- 遮罩层 -->
		<view class="modal-overlay" @click="closeModal"></view>

		<!-- 弹窗内容 -->
		<view class="modal-content">
			<view class="modal-header">
				<text class="modal-title">举报内容</text>
				<button class="close-btn" @click="closeModal">×</button>
			</view>

			<view class="modal-body">
				<view class="form-group">
					<text class="label">举报原因</text>
					<view class="reason-list">
						<view 
							v-for="reason in reportReasons" 
							:key="reason.value"
							class="reason-item"
							:class="{ 'selected': selectedReason === reason.value }"
							@click="selectReason(reason.value)"
						>
							<text class="reason-radio">
								{{ selectedReason === reason.value ? '●' : '○' }}
							</text>
							<text class="reason-text">{{ reason.label }}</text>
						</view>
					</view>
				</view>

				<view class="form-group">
					<text class="label">详细描述（可选）</text>
					<textarea 
						v-model="description"
						class="description-input"
						placeholder="请详细描述举报原因"
						:maxlength="500"
					></textarea>
					<text class="char-count">{{ description.length }}/500</text>
				</view>
			</view>

			<view class="modal-footer">
				<button class="cancel-btn" @click="closeModal">取消</button>
				<button 
					class="submit-btn" 
					:disabled="!selectedReason || loading"
					@click="submitReport"
				>
					{{ loading ? '提交中...' : '提交举报' }}
				</button>
			</view>
		</view>
	</view>
</template>

<script>
	export default {
		name: 'SimpleReportModal',
		props: {
			visible: {
				type: Boolean,
				default: false
			},
			targetType: {
				type: String,
				required: true // 'post', 'comment', 'user'
			},
			targetId: {
				type: [Number, String],
				required: true
			}
		},
		data() {
			return {
				selectedReason: '',
				description: '',
				loading: false,
				reportReasons: [
					{ value: 'spam', label: '垃圾信息' },
					{ value: 'inappropriate', label: '不当内容' },
					{ value: 'harassment', label: '骚扰他人' },
					{ value: 'fake_info', label: '虚假信息' },
					{ value: 'violence', label: '暴力内容' },
					{ value: 'other', label: '其他原因' }
				]
			}
		},
		methods: {
			// 选择举报原因
			selectReason(reason) {
				this.selectedReason = reason
			},

			// 关闭弹窗
			closeModal() {
				this.resetForm()
				this.$emit('close')
			},

			// 重置表单
			resetForm() {
				this.selectedReason = ''
				this.description = ''
				this.loading = false
			},

			// 提交举报
			async submitReport() {
				if (!this.selectedReason) {
					uni.showToast({
						title: '请选择举报原因',
						icon: 'none'
					})
					return
				}

				this.loading = true

				try {
					// 模拟API调用
					await new Promise(resolve => setTimeout(resolve, 1000))

					uni.showToast({
						title: '举报提交成功',
						icon: 'success'
					})

					// 触发事件通知父组件
					this.$emit('report-submitted', {
						targetType: this.targetType,
						targetId: this.targetId,
						reason: this.selectedReason,
						description: this.description
					})

					this.closeModal()
				} catch (error) {
					console.error('提交举报失败:', error)
					uni.showToast({
						title: '提交失败，请重试',
						icon: 'none'
					})
				} finally {
					this.loading = false
				}
			}
		}
	}
</script>

<style lang="scss" scoped>
	.simple-report-modal {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		z-index: 9999;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.modal-overlay {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background-color: rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(4px);
	}

	.modal-content {
		position: relative;
		width: 90%;
		max-width: 500rpx;
		max-height: 80vh;
		background-color: white;
		border-radius: 24rpx;
		box-shadow: 0 20rpx 60rpx rgba(0, 0, 0, 0.3);
		overflow: hidden;
		animation: modalShow 0.3s ease;
	}

	@keyframes modalShow {
		from {
			transform: scale(0.9);
			opacity: 0;
		}
		to {
			transform: scale(1);
			opacity: 1;
		}
	}

	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 32rpx;
		border-bottom: 2rpx solid #f0f0f0;
	}

	.modal-title {
		font-size: 36rpx;
		font-weight: 600;
		color: #333;
	}

	.close-btn {
		width: 60rpx;
		height: 60rpx;
		border-radius: 50%;
		background-color: #f5f5f5;
		border: none;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 40rpx;
		color: #666;
		line-height: 1;
		transition: background-color 0.3s ease;

		&:hover {
			background-color: #e0e0e0;
		}
	}

	.modal-body {
		padding: 32rpx;
		max-height: 60vh;
		overflow-y: auto;
	}

	.form-group {
		margin-bottom: 32rpx;

		&:last-child {
			margin-bottom: 0;
		}
	}

	.label {
		display: block;
		font-size: 32rpx;
		font-weight: 500;
		color: #333;
		margin-bottom: 16rpx;
	}

	.reason-list {
		display: flex;
		flex-direction: column;
		gap: 16rpx;
	}

	.reason-item {
		display: flex;
		align-items: center;
		padding: 20rpx;
		border: 2rpx solid #e0e0e0;
		border-radius: 16rpx;
		cursor: pointer;
		transition: all 0.3s ease;

		&.selected {
			border-color: #ff6b35;
			background-color: rgba(255, 107, 53, 0.1);
		}

		&:hover {
			border-color: #ff6b35;
		}
	}

	.reason-radio {
		margin-right: 16rpx;
		font-size: 32rpx;
		color: #ff6b35;
	}

	.reason-text {
		font-size: 30rpx;
		color: #333;
	}

	.description-input {
		width: 100%;
		min-height: 200rpx;
		padding: 20rpx;
		border: 2rpx solid #e0e0e0;
		border-radius: 16rpx;
		font-size: 30rpx;
		color: #333;
		background-color: #fafafa;
		resize: none;
		box-sizing: border-box;

		&:focus {
			border-color: #ff6b35;
			background-color: white;
		}
	}

	.char-count {
		display: block;
		text-align: right;
		font-size: 24rpx;
		color: #999;
		margin-top: 8rpx;
	}

	.modal-footer {
		display: flex;
		gap: 16rpx;
		padding: 32rpx;
		border-top: 2rpx solid #f0f0f0;
	}

	.cancel-btn,
	.submit-btn {
		flex: 1;
		height: 80rpx;
		border-radius: 20rpx;
		font-size: 32rpx;
		font-weight: 500;
		border: none;
		transition: all 0.3s ease;

		&:active {
			transform: scale(0.95);
		}
	}

	.cancel-btn {
		background-color: #f5f5f5;
		color: #666;

		&:hover {
			background-color: #e0e0e0;
		}
	}

	.submit-btn {
		background-color: #ff6b35;
		color: white;

		&:hover {
			background-color: #e55a2b;
		}

		&:disabled {
			background-color: #ccc;
			cursor: not-allowed;
		}
	}

	/* 响应式设计 */
	@media screen and (max-width: 480px) {
		.modal-content {
			width: 95%;
			max-width: none;
		}

		.modal-header,
		.modal-body,
		.modal-footer {
			padding: 24rpx;
		}

		.modal-title {
			font-size: 32rpx;
		}

		.reason-item {
			padding: 16rpx;
		}

		.reason-text {
			font-size: 28rpx;
		}

		.description-input {
			min-height: 160rpx;
			font-size: 28rpx;
		}
	}
</style>
