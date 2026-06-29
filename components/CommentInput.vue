<template>
	<t-popup v-model:visible="visible" placement="bottom" :overlay="true" :close-on-overlay-click="false" @visible-change="handlePopupChange">
		<view class="comment-input-container">
			<view class="input-header">
				<text class="popup-title">
					{{ replyTarget ? `回复 @${getReplyTargetName()}` : '添加评论' }}
				</text>
				<view class="close-btn" @tap="close">
					<l-icon name="close" size="20" color="var(--text-secondary)"></l-icon>
				</view>
			</view>

			<view v-if="replyTarget" class="reply-target-info">
				<view class="target-content">
					<text class="target-text">{{ replyTarget.content }}</text>
				</view>
				<view class="clear-target" @tap="clearReplyTarget">
					<l-icon name="close" size="16" color="var(--text-secondary)"></l-icon>
				</view>
			</view>

			<view class="input-content">
				<textarea
					class="input-area"
					v-model="commentText"
					placeholder="说点什么吧..."
					fixed="true"
					cursor-spacing="100"
					maxlength="500"
					focus="true"
					@input="updateCursorPosition"
					@blur="onInputBlur"
				></textarea>

				<view v-if="showMentionSuggestions" class="mention-suggestions">
					<view
						v-for="user in mentionSuggestions"
						:key="user.id"
						class="mention-item"
						@tap="selectMentionUser(user)"
					>
						<UserAvatar
						class="mention-avatar"
						:src="user.avatar"
						:name="getUserDisplayName(user)"
						:size="56"
					></UserAvatar>
						<text class="mention-name">{{ getUserDisplayName(user) }}</text>
					</view>
				</view>

				<view v-if="images.length > 0" class="images-preview">
					<view
						v-for="(image, index) in images"
						:key="index"
						class="image-preview-item"
					>
						<image
							:src="image"
							mode="aspectFill"
							class="preview-image"
							@tap="previewImage(image)"
						></image>
						<view class="remove-image" @tap="removeImage(index)">
							<l-icon name="close" size="16" color="#fff"></l-icon>
						</view>
					</view>
				</view>

				<view class="input-toolbar">
					<view class="toolbar-left">
						<view class="tool-btn" @tap="chooseImage" v-if="images.length < maxImages">
							<l-icon name="image" size="20" color="var(--primary-color)"></l-icon>
						</view>
						<view class="tool-btn" @tap="showEmojiPanel">
							<text class="emoji-icon">😊</text>
						</view>
					</view>
					<view class="toolbar-right">
						<text
							class="char-count"
							:class="{
								warning: commentText.length > maxLength * 0.8,
								danger: commentText.length > maxLength * 0.95
							}"
						>
							{{ commentText.length }}/{{ maxLength }}
						</text>
						<button
							class="send-btn"
							:disabled="!canSubmit"
							@tap="submit"
							:loading="submitting"
						>
							<l-icon v-if="!submitting" name="send-filled" size="16" color="#fff"></l-icon>
							<text class="send-text">{{ submitting ? '发送中...' : '发送' }}</text>
						</button>
					</view>
				</view>
			</view>
		</view>
	</t-popup>

	<EmojiPicker
		ref="emojiPicker"
		@select="handleEmojiSelect"
		@close="handleEmojiClose"
	/>
</template>

<script>
import UserAvatar from './UserAvatar.vue'
import { useCommentInput } from '@/composables/useCommentInput.js';
import { getUserDisplayName } from '@/utils/userMention.js';
import EmojiPicker from '@/components/EmojiPicker.vue';

export default {
	name: 'CommentInput',
	components: {
		UserAvatar,
		EmojiPicker
	},
	emits: ['success'],
	props: {
		postId: {
			type: [String, Number],
			default: null
		},
		maxLength: {
			type: Number,
			default: 2000
		},
		maxImages: {
			type: Number,
			default: 9
		}
	},
	setup(props, { emit }) {
		const {
			visible,
			commentText,
			images,
			showMentionSuggestions,
			mentionSuggestions,
			replyTarget,
			cursorPosition,
			canSubmit,
			show,
			close,
			handlePopupChange,
			reset,
			getReplyTargetName,
			clearReplyTarget,
			detectMention,
			searchUsers,
			selectMentionUser,
			addMentionedUser,
			chooseImage,
			handleImageUpload,
			removeImage,
			previewImage,
			showEmojiPanel,
			handleEmojiSelect,
			handleEmojiClose,
			insertEmojiAtCursor,
			updateCursorPosition,
			submit
		} = useCommentInput({ postId: props.postId });

		const onInputBlur = () => {
			setTimeout(() => {
				showMentionSuggestions.value = false;
			}, 200);
		};

		return {
			visible,
			commentText,
			images,
			showMentionSuggestions,
			mentionSuggestions,
			replyTarget,
			cursorPosition,
			canSubmit,
			show,
			close,
			handlePopupChange,
			reset,
			getReplyTargetName,
			clearReplyTarget,
			detectMention,
			searchUsers,
			selectMentionUser,
			addMentionedUser,
			chooseImage,
			handleImageUpload,
			removeImage,
			previewImage,
			showEmojiPanel,
			handleEmojiSelect,
			handleEmojiClose,
			insertEmojiAtCursor,
			updateCursorPosition,
			submit: () => submit(emit),
			getUserDisplayName,
			onInputBlur,
			maxLength: props.maxLength,
			maxImages: props.maxImages
		};
	}
};
</script>

<style lang="scss" scoped>
.comment-input-container {
	background: var(--bg-primary);
	border-radius: 32rpx 32rpx 0 0;
	max-height: 80vh;
	min-height: 400rpx;
	display: flex;
	flex-direction: column;
	box-shadow: 0 -12rpx 48rpx rgba(0, 0, 0, 0.15);
	backdrop-filter: blur(20px);
	position: relative;
	animation: slideUp 0.4s cubic-bezier(0.4, 0, 0.2, 1);

	&::before {
		content: '';
		position: absolute;
		top: 16rpx;
		left: 50%;
		transform: translateX(-50%);
		width: 80rpx;
		height: 6rpx;
		background: var(--border-color);
		border-radius: 3rpx;
		opacity: 0.6;
	}

	.input-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 40rpx 32rpx 24rpx;
		border-bottom: 1px solid var(--border-color-light);
		background: var(--bg-secondary);
		border-radius: 32rpx 32rpx 0 0;

		.popup-title {
			font-size: 34rpx;
			font-weight: 700;
			color: var(--text-primary);
			letter-spacing: 0.5rpx;
		}

		.close-btn {
			padding: 12rpx;
			border-radius: 50%;
			background: var(--bg-tertiary);
			transition: all 0.3s ease;

			&:hover {
				background: var(--bg-primary);
				transform: rotate(90deg);
			}

			&:active {
				transform: rotate(90deg) scale(0.95);
			}
		}
	}

	.reply-target-info {
		display: flex;
		align-items: flex-start;
		gap: 20rpx;
		padding: 24rpx 32rpx;
		background: linear-gradient(135deg, rgba(99, 102, 241, 0.05), rgba(139, 92, 246, 0.03));
		border-bottom: 1px solid var(--border-color-light);
		border-left: 4rpx solid var(--primary-color);
		margin: 0 32rpx;
		border-radius: 12rpx;
		position: relative;

		&::before {
			content: '回复';
			position: absolute;
			top: -12rpx;
			left: 16rpx;
			background: var(--primary-color);
			color: #fff;
			font-size: 20rpx;
			padding: 4rpx 12rpx;
			border-radius: 8rpx;
			font-weight: 600;
		}

		.target-content {
			flex: 1;
			padding-top: 8rpx;

			.target-text {
				font-size: 28rpx;
				color: var(--text-primary);
				line-height: 1.5;
				display: -webkit-box;
				-webkit-line-clamp: 3;
				line-clamp: 3;
				-webkit-box-orient: vertical;
				overflow: hidden;
				background: var(--bg-primary);
				padding: 16rpx;
				border-radius: 8rpx;
				border: 1rpx solid var(--border-color-light);
			}
		}

		.clear-target {
			padding: 12rpx;
			border-radius: 50%;
			background: rgba(255, 255, 255, 0.8);
			transition: all 0.3s ease;

			&:hover {
				background: #ff4757;
				color: #fff;
				transform: rotate(90deg);
			}
		}
	}

	.input-content {
		flex: 1;
		padding: 32rpx;
		display: flex;
		flex-direction: column;

		.input-area {
			min-height: 160rpx;
			max-height: 400rpx;
			font-size: 30rpx;
			line-height: 1.7;
			color: var(--text-primary);
			background: var(--bg-secondary);
			border: 2rpx solid var(--border-color);
			border-radius: 16rpx;
			padding: 24rpx;
			outline: none;
			resize: none;
			margin-bottom: 24rpx;
			transition: all 0.3s ease;
			letter-spacing: 0.3rpx;

			&:focus {
				border-color: var(--primary-color);
				background: var(--bg-primary);
				box-shadow: 0 0 0 4rpx rgba(99, 102, 241, 0.1);
			}

			&::placeholder {
				color: var(--text-disabled);
				font-size: 28rpx;
			}
		}

		.mention-suggestions {
			max-height: 240rpx;
			overflow-y: auto;
			border: 2rpx solid var(--primary-color);
			border-radius: 16rpx;
			background: var(--bg-primary);
			margin-bottom: 24rpx;
			box-shadow: 0 8rpx 24rpx rgba(99, 102, 241, 0.15);
			backdrop-filter: blur(10px);

			.mention-item {
				display: flex;
				align-items: center;
				gap: 20rpx;
				padding: 20rpx 24rpx;
				border-bottom: 1px solid var(--border-color-light);
				transition: all 0.3s ease;

				&:last-child {
					border-bottom: none;
				}

				&:hover {
					background: var(--bg-secondary);
					transform: translateX(8rpx);
				}

				&:active {
					background: var(--bg-tertiary);
					transform: translateX(4rpx);
				}

				.mention-avatar {
					width: 56rpx;
					height: 56rpx;
					border-radius: 50%;
					border: 2rpx solid var(--border-color-light);
				}

				.mention-name {
					font-size: 30rpx;
					color: var(--text-primary);
					font-weight: 500;
				}
			}
		}

		.images-preview {
			display: flex;
			flex-wrap: wrap;
			gap: 20rpx;
			margin-bottom: 24rpx;
			padding: 20rpx;
			background: var(--bg-secondary);
			border-radius: 16rpx;
			border: 2rpx dashed var(--border-color);

			.image-preview-item {
				position: relative;
				width: 140rpx;
				height: 140rpx;
				border-radius: 16rpx;
				overflow: hidden;
				transition: all 0.3s ease;

				&:hover {
					transform: scale(1.05);
					box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.15);
				}

				.preview-image {
					width: 100%;
					height: 100%;
					border-radius: 16rpx;
					object-fit: cover;
				}

				.remove-image {
					position: absolute;
					top: -8rpx;
					right: -8rpx;
					width: 40rpx;
					height: 40rpx;
					background: linear-gradient(135deg, #ff4757, #ff3742);
					border-radius: 50%;
					display: flex;
					align-items: center;
					justify-content: center;
					box-shadow: 0 4rpx 12rpx rgba(255, 71, 87, 0.3);
					transition: all 0.3s ease;

					&:hover {
						transform: scale(1.1) rotate(90deg);
						background: linear-gradient(135deg, #ff3742, #ff1744);
					}

					&:active {
						transform: scale(0.95) rotate(90deg);
					}
				}
			}
		}

		.input-toolbar {
			display: flex;
			justify-content: space-between;
			align-items: center;
			padding: 24rpx 0;
			border-top: 2rpx solid var(--border-color-light);
			margin-top: 20rpx;
			background: var(--bg-secondary);
			border-radius: 16rpx;
			padding: 24rpx;
			margin: 20rpx -32rpx 0;

			.toolbar-left {
				display: flex;
				gap: 20rpx;

				.tool-btn {
					padding: 20rpx;
					border-radius: 16rpx;
					background: var(--bg-primary);
					border: 2rpx solid var(--border-color);
					transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
					box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.05);

					&:hover {
						background: var(--primary-color);
						border-color: var(--primary-color);
						transform: translateY(-4rpx) scale(1.05);
						box-shadow: 0 8rpx 20rpx rgba(99, 102, 241, 0.3);

						uni-icons {
							color: #fff !important;
						}

						.emoji-icon {
							transform: scale(1.1);
						}
					}

					&:active {
						transform: translateY(-2rpx) scale(1.02);
					}

					.emoji-icon {
						font-size: 40rpx;
						line-height: 1;
						display: flex;
						align-items: center;
						justify-content: center;
						transition: all 0.3s ease;
					}
				}
			}

			.toolbar-right {
				display: flex;
				align-items: center;
				gap: 24rpx;

				.char-count {
					font-size: 24rpx;
					color: var(--text-secondary);
					font-weight: 600;
					padding: 12rpx 20rpx;
					background: linear-gradient(135deg, var(--bg-primary), var(--bg-secondary));
					border-radius: 16rpx;
					border: 2rpx solid var(--border-color);
					box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.05);
					transition: all 0.3s ease;

					&.warning {
						color: #f56565;
						border-color: #f56565;
						background: linear-gradient(135deg, rgba(245, 101, 101, 0.1), rgba(245, 101, 101, 0.05));
					}

					&.danger {
						color: #e53e3e;
						border-color: #e53e3e;
						background: linear-gradient(135deg, rgba(229, 62, 62, 0.15), rgba(229, 62, 62, 0.08));
						animation: shake 0.5s ease-in-out;
					}
				}

				.send-btn {
					display: flex;
					align-items: center;
					gap: 12rpx;
					padding: 20rpx 48rpx;
					background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
					color: #fff;
					border: none;
					border-radius: 32rpx;
					font-size: 30rpx;
					font-weight: 700;
					letter-spacing: 1rpx;
					box-shadow: 0 8rpx 24rpx rgba(102, 126, 234, 0.4);
					transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
					overflow: hidden;
					position: relative;
					min-width: 140rpx;
					justify-content: center;

					&::before {
						content: '';
						position: absolute;
						top: 0;
						left: -100%;
						width: 100%;
						height: 100%;
						background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.3), transparent);
						transition: left 0.6s ease;
					}

					&:hover {
						transform: translateY(-4rpx) scale(1.02);
						box-shadow: 0 12rpx 32rpx rgba(102, 126, 234, 0.5);
						background: linear-gradient(135deg, #5a67d8 0%, #6b46c1 100%);

						&::before {
							left: 100%;
						}
					}

					&:active {
						transform: translateY(-2rpx) scale(0.98);
						box-shadow: 0 6rpx 20rpx rgba(102, 126, 234, 0.4);
					}

					&[disabled] {
						background: linear-gradient(135deg, #e2e8f0 0%, #cbd5e0 100%);
						color: #a0aec0;
						box-shadow: none;
						cursor: not-allowed;
						transform: none;

						&::before {
							display: none;
						}
					}

					.send-text {
						font-size: 28rpx;
						font-weight: 700;
						letter-spacing: 0.5rpx;
					}
				}
			}
		}
	}
}

@keyframes slideUp {
	0% {
		transform: translateY(100%);
		opacity: 0;
	}
	100% {
		transform: translateY(0);
		opacity: 1;
	}
}

@keyframes shake {
	0%, 100% { transform: translateX(0); }
	25% { transform: translateX(-4rpx); }
	75% { transform: translateX(4rpx); }
}

/* 深色模式优化 */
.theme-dark {
	.comment-input-container {
		background: var(--bg-secondary);
		box-shadow: 0 -12rpx 48rpx rgba(0, 0, 0, 0.3);

		.reply-target-info {
			background: linear-gradient(135deg, rgba(99, 102, 241, 0.15), rgba(139, 92, 246, 0.1));
		}

		.input-content {
			.input-area {
				background: var(--bg-tertiary);
				border-color: var(--border-color-light);

				&:focus {
					background: var(--bg-secondary);
					border-color: var(--primary-color);
				}
			}

			.input-toolbar {
				background: var(--bg-tertiary);

				.tool-btn {
					background: var(--bg-secondary);
					border-color: var(--border-color-light);
				}

				.send-btn {
					background: linear-gradient(135deg, #4c51bf 0%, #553c9a 100%);

					&:hover {
						background: linear-gradient(135deg, #434190 0%, #4c1d95 100%);
					}
				}
			}
		}
	}
}

.mp-theme-dark {
	.comment-input-container {
		.reply-target-info {
			background: linear-gradient(135deg, rgba(99, 102, 241, 0.1), rgba(139, 92, 246, 0.06));
		}

		.input-content {
			.input-area {
				background: var(--bg-tertiary);

				&:focus {
					background: var(--bg-secondary);
				}
			}
		}
	}
}

/* 响应式设计 */
@media screen and (max-width: 480px) {
	.comment-input-container {
		max-height: 90vh;

		.input-header {
			padding: 32rpx 24rpx 20rpx;
		}

		.reply-target-info {
			margin: 0 24rpx;
			padding: 20rpx;
		}

		.input-content {
			padding: 24rpx;
			padding-bottom: 160rpx;

			.input-area {
				min-height: 120rpx;
				font-size: 28rpx;
				padding: 20rpx;
			}

			.images-preview {
				gap: 16rpx;
				padding: 16rpx;

				.image-preview-item {
					width: 120rpx;
					height: 120rpx;
				}
			}

			.input-toolbar {
				.toolbar-left .tool-btn {
					padding: 12rpx;
				}

				.toolbar-right {
					gap: 16rpx;

					.char-count {
						font-size: 22rpx;
						padding: 6rpx 12rpx;
					}

					.send-btn {
						padding: 16rpx 36rpx;
						font-size: 26rpx;
						border-radius: 28rpx;
					}
				}
			}
		}
	}
}

/* 安全区域适配 */
@supports (padding: env(safe-area-inset-bottom)) {
	.comment-input-container .input-content {
		padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
	}
}

/* #ifdef H5 */
.comment-input-container .input-content {
	padding-bottom: 140rpx;
}
/* #endif */

/* #ifdef MP-WEIXIN */
.comment-input-container .input-content {
	padding-bottom: 120rpx;
}
/* #endif */

/* #ifdef APP-PLUS */
.comment-input-container .input-content {
	padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
}
/* #endif */
</style>
