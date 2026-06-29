<template>
	<t-popup v-model:visible="visible" placement="bottom" :overlay="true" :close-on-overlay-click="true" @visible-change="handlePopupChange">
		<view class="emoji-picker-container">
			<!-- 表情选择器头部 -->
			<view class="emoji-header">
				<text class="emoji-title">选择表情</text>
				<view class="close-btn" @tap="close">
					<t-icon name="close" size="20" color="var(--text-secondary)"></t-icon>
				</view>
			</view>

			<!-- 表情分类标签 -->
			<view class="emoji-tabs">
				<view 
					v-for="(category, index) in emojiCategories" 
					:key="index"
					class="emoji-tab"
					:class="{ active: activeTab === index }"
					@tap="switchTab(index)"
				>
					<text class="tab-emoji">{{ category.icon }}</text>
					<text class="tab-name">{{ category.name }}</text>
				</view>
			</view>

			<!-- 表情网格 -->
			<scroll-view 
				class="emoji-grid-container" 
				scroll-y 
				:scroll-top="scrollTop"
				@scroll="handleScroll"
			>
				<view class="emoji-grid">
					<view 
						v-for="emoji in currentEmojis" 
						:key="emoji.code"
						class="emoji-item"
						@tap="selectEmoji(emoji)"
					>
						<text class="emoji-char">{{ emoji.char }}</text>
					</view>
				</view>
			</scroll-view>

			<!-- 最近使用的表情 -->
			<view v-if="recentEmojis.length > 0" class="recent-emojis">
				<view class="recent-title">
					<t-icon name="time" size="16" color="var(--text-secondary)"></t-icon>
					<text>最近使用</text>
				</view>
				<view class="recent-grid">
					<view 
						v-for="emoji in recentEmojis" 
						:key="emoji.code"
						class="emoji-item recent-item"
						@tap="selectEmoji(emoji)"
					>
						<text class="emoji-char">{{ emoji.char }}</text>
					</view>
				</view>
			</view>
		</view>
	</t-popup>
</template>

<script>
export default {
	name: 'EmojiPicker',
	data() {
		return {
			visible: false,
			activeTab: 0,
			scrollTop: 0,
			recentEmojis: [],
			emojiCategories: [
				{
					name: '常用',
					icon: '😊',
					emojis: [
						{ char: '😊', code: 'blush', name: '微笑' },
						{ char: '😂', code: 'joy', name: '笑哭' },
						{ char: '🤣', code: 'rofl', name: '大笑' },
						{ char: '😍', code: 'heart_eyes', name: '爱心眼' },
						{ char: '🥰', code: 'smiling_face_with_hearts', name: '花痴' },
						{ char: '😘', code: 'kissing_heart', name: '飞吻' },
						{ char: '😭', code: 'sob', name: '大哭' },
						{ char: '😅', code: 'sweat_smile', name: '汗笑' },
						{ char: '😓', code: 'sweat', name: '冷汗' },
						{ char: '🤔', code: 'thinking', name: '思考' },
						{ char: '😴', code: 'sleeping', name: '睡觉' },
						{ char: '🙄', code: 'eye_roll', name: '翻白眼' },
						{ char: '😤', code: 'huffing', name: '生气' },
						{ char: '😡', code: 'rage', name: '愤怒' },
						{ char: '🤯', code: 'exploding_head', name: '爆炸' },
						{ char: '😱', code: 'scream', name: '尖叫' },
						{ char: '🥺', code: 'pleading_face', name: '可怜' },
						{ char: '😢', code: 'cry', name: '哭泣' },
						{ char: '👍', code: 'thumbsup', name: '点赞' },
						{ char: '👎', code: 'thumbsdown', name: '踩' },
						{ char: '👌', code: 'ok_hand', name: 'OK' },
						{ char: '🙏', code: 'pray', name: '祈祷' },
						{ char: '💪', code: 'muscle', name: '肌肉' },
						{ char: '👏', code: 'clap', name: '鼓掌' }
					]
				},
				{
					name: '笑脸',
					icon: '😀',
					emojis: [
						{ char: '😀', code: 'grinning', name: '露齿笑' },
						{ char: '😃', code: 'smiley', name: '笑脸' },
						{ char: '😄', code: 'smile', name: '大笑' },
						{ char: '😁', code: 'grin', name: '咧嘴笑' },
						{ char: '😆', code: 'laughing', name: '大笑' },
						{ char: '😅', code: 'sweat_smile', name: '汗笑' },
						{ char: '🤣', code: 'rofl', name: '笑哭' },
						{ char: '😂', code: 'joy', name: '喜极而泣' },
						{ char: '🙂', code: 'slightly_smiling_face', name: '微笑' },
						{ char: '🙃', code: 'upside_down_face', name: '倒脸' },
						{ char: '😉', code: 'wink', name: '眨眼' },
						{ char: '😊', code: 'blush', name: '害羞' },
						{ char: '😇', code: 'innocent', name: '天使' },
						{ char: '🥰', code: 'smiling_face_with_hearts', name: '爱心眼' },
						{ char: '😍', code: 'heart_eyes', name: '花痴' },
						{ char: '🤩', code: 'star_struck', name: '星星眼' },
						{ char: '😘', code: 'kissing_heart', name: '飞吻' },
						{ char: '😗', code: 'kissing', name: '亲吻' },
						{ char: '☺️', code: 'relaxed', name: '放松' },
						{ char: '😚', code: 'kissing_closed_eyes', name: '闭眼亲吻' },
						{ char: '😙', code: 'kissing_smiling_eyes', name: '微笑亲吻' },
						{ char: '🥲', code: 'smiling_face_with_tear', name: '含泪微笑' }
					]
				},
				{
					name: '情绪',
					icon: '😭',
					emojis: [
						{ char: '😭', code: 'sob', name: '大哭' },
						{ char: '😢', code: 'cry', name: '哭泣' },
						{ char: '🥺', code: 'pleading_face', name: '可怜' },
						{ char: '😔', code: 'pensive', name: '沉思' },
						{ char: '😞', code: 'disappointed', name: '失望' },
						{ char: '😟', code: 'worried', name: '担心' },
						{ char: '😕', code: 'confused', name: '困惑' },
						{ char: '🙁', code: 'slightly_frowning_face', name: '微皱眉' },
						{ char: '☹️', code: 'frowning_face', name: '皱眉' },
						{ char: '😣', code: 'persevere', name: '坚持' },
						{ char: '😖', code: 'confounded', name: '困扰' },
						{ char: '😫', code: 'tired_face', name: '疲惫' },
						{ char: '😩', code: 'weary', name: '厌倦' },
						{ char: '🥱', code: 'yawning_face', name: '打哈欠' },
						{ char: '😤', code: 'huffing', name: '生气' },
						{ char: '😠', code: 'angry', name: '愤怒' },
						{ char: '😡', code: 'rage', name: '暴怒' },
						{ char: '🤬', code: 'swearing', name: '骂人' },
						{ char: '🤯', code: 'exploding_head', name: '爆炸' },
						{ char: '😳', code: 'flushed', name: '脸红' },
						{ char: '🥵', code: 'hot_face', name: '热脸' },
						{ char: '🥶', code: 'cold_face', name: '冷脸' },
						{ char: '😱', code: 'scream', name: '尖叫' },
						{ char: '😨', code: 'fearful', name: '恐惧' }
					]
				},
				{
					name: '手势',
					icon: '👍',
					emojis: [
						{ char: '👍', code: 'thumbsup', name: '点赞' },
						{ char: '👎', code: 'thumbsdown', name: '踩' },
						{ char: '👌', code: 'ok_hand', name: 'OK' },
						{ char: '✌️', code: 'v', name: '胜利' },
						{ char: '🤞', code: 'crossed_fingers', name: '祈祷' },
						{ char: '🤟', code: 'love_you_gesture', name: '爱你' },
						{ char: '🤘', code: 'metal', name: '摇滚' },
						{ char: '🤙', code: 'call_me_hand', name: '打电话' },
						{ char: '👈', code: 'point_left', name: '左指' },
						{ char: '👉', code: 'point_right', name: '右指' },
						{ char: '👆', code: 'point_up_2', name: '上指' },
						{ char: '🖕', code: 'middle_finger', name: '中指' },
						{ char: '👇', code: 'point_down', name: '下指' },
						{ char: '☝️', code: 'point_up', name: '食指向上' },
						{ char: '👋', code: 'wave', name: '挥手' },
						{ char: '🤚', code: 'raised_back_of_hand', name: '手背' },
						{ char: '🖐️', code: 'raised_hand_with_fingers_splayed', name: '张开手' },
						{ char: '✋', code: 'hand', name: '举手' },
						{ char: '🖖', code: 'vulcan_salute', name: '瓦肯礼' },
						{ char: '👏', code: 'clap', name: '鼓掌' },
						{ char: '🙌', code: 'raised_hands', name: '举双手' },
						{ char: '🤝', code: 'handshake', name: '握手' }
					]
				},
				{
					name: '爱心',
					icon: '❤️',
					emojis: [
						{ char: '❤️', code: 'heart', name: '红心' },
						{ char: '🧡', code: 'orange_heart', name: '橙心' },
						{ char: '💛', code: 'yellow_heart', name: '黄心' },
						{ char: '💚', code: 'green_heart', name: '绿心' },
						{ char: '💙', code: 'blue_heart', name: '蓝心' },
						{ char: '💜', code: 'purple_heart', name: '紫心' },
						{ char: '🖤', code: 'black_heart', name: '黑心' },
						{ char: '🤍', code: 'white_heart', name: '白心' },
						{ char: '🤎', code: 'brown_heart', name: '棕心' },
						{ char: '💔', code: 'broken_heart', name: '心碎' },
						{ char: '❣️', code: 'heavy_heart_exclamation', name: '心叹号' },
						{ char: '💕', code: 'two_hearts', name: '双心' },
						{ char: '💞', code: 'revolving_hearts', name: '旋转心' },
						{ char: '💓', code: 'heartbeat', name: '心跳' },
						{ char: '💗', code: 'heartpulse', name: '心脉' },
						{ char: '💖', code: 'sparkling_heart', name: '闪亮心' },
						{ char: '💘', code: 'cupid', name: '丘比特' },
						{ char: '💝', code: 'gift_heart', name: '礼物心' },
						{ char: '💟', code: 'heart_decoration', name: '心装饰' }
					]
				},
				{
					name: '动物',
					icon: '🐶',
					emojis: [
						{ char: '🐶', code: 'dog', name: '狗' },
						{ char: '🐱', code: 'cat', name: '猫' },
						{ char: '🐭', code: 'mouse', name: '老鼠' },
						{ char: '🐹', code: 'hamster', name: '仓鼠' },
						{ char: '🐰', code: 'rabbit', name: '兔子' },
						{ char: '🦊', code: 'fox_face', name: '狐狸' },
						{ char: '🐻', code: 'bear', name: '熊' },
						{ char: '🐼', code: 'panda_face', name: '熊猫' },
						{ char: '🐨', code: 'koala', name: '考拉' },
						{ char: '🐯', code: 'tiger', name: '老虎' },
						{ char: '🦁', code: 'lion', name: '狮子' },
						{ char: '🐮', code: 'cow', name: '牛' },
						{ char: '🐷', code: 'pig', name: '猪' },
						{ char: '🐸', code: 'frog', name: '青蛙' },
						{ char: '🐵', code: 'monkey_face', name: '猴子' },
						{ char: '🙈', code: 'see_no_evil', name: '非礼勿视' },
						{ char: '🙉', code: 'hear_no_evil', name: '非礼勿听' },
						{ char: '🙊', code: 'speak_no_evil', name: '非礼勿言' },
						{ char: '🐒', code: 'monkey', name: '猴' },
						{ char: '🐔', code: 'chicken', name: '鸡' },
						{ char: '🐧', code: 'penguin', name: '企鹅' },
						{ char: '🐦', code: 'bird', name: '鸟' }
					]
				},
				{
					name: '食物',
					icon: '🍎',
					emojis: [
						{ char: '🍎', code: 'apple', name: '苹果' },
						{ char: '🍊', code: 'tangerine', name: '橘子' },
						{ char: '🍋', code: 'lemon', name: '柠檬' },
						{ char: '🍌', code: 'banana', name: '香蕉' },
						{ char: '🍉', code: 'watermelon', name: '西瓜' },
						{ char: '🍇', code: 'grapes', name: '葡萄' },
						{ char: '🍓', code: 'strawberry', name: '草莓' },
						{ char: '🫐', code: 'blueberries', name: '蓝莓' },
						{ char: '🍈', code: 'melon', name: '甜瓜' },
						{ char: '🍒', code: 'cherries', name: '樱桃' },
						{ char: '🍑', code: 'peach', name: '桃子' },
						{ char: '🥭', code: 'mango', name: '芒果' },
						{ char: '🍍', code: 'pineapple', name: '菠萝' },
						{ char: '🥥', code: 'coconut', name: '椰子' },
						{ char: '🥝', code: 'kiwi_fruit', name: '猕猴桃' },
						{ char: '🍅', code: 'tomato', name: '番茄' },
						{ char: '🍆', code: 'eggplant', name: '茄子' },
						{ char: '🥑', code: 'avocado', name: '牛油果' },
						{ char: '🥦', code: 'broccoli', name: '西兰花' },
						{ char: '🥬', code: 'leafy_greens', name: '绿叶菜' },
						{ char: '🥒', code: 'cucumber', name: '黄瓜' },
						{ char: '🌶️', code: 'hot_pepper', name: '辣椒' }
					]
				}
			]
		}
	},
	computed: {
		currentEmojis() {
			return this.emojiCategories[this.activeTab]?.emojis || [];
		}
	},
	mounted() {
		this.loadRecentEmojis();
	},
	methods: {
		// 显示表情选择器
		show() {
			this.visible = true;
		},

		// 关闭表情选择器
		close() {
			this.visible = false;
		},

		// 处理弹窗状态变化
		handlePopupChange(e) {
			if (!e.visible) {
				this.$emit('close');
			}
		},

		// 切换表情分类
		switchTab(index) {
			this.activeTab = index;
			this.scrollTop = 0;
		},

		// 选择表情
		selectEmoji(emoji) {
			// 添加到最近使用
			this.addToRecent(emoji);
			
			// 触发选择事件
			this.$emit('select', emoji);
			
			// 震动反馈
			uni.vibrateShort({
				type: 'light'
			});
		},

		// 添加到最近使用
		addToRecent(emoji) {
			// 移除已存在的相同表情
			this.recentEmojis = this.recentEmojis.filter(item => item.code !== emoji.code);
			
			// 添加到开头
			this.recentEmojis.unshift(emoji);
			
			// 限制最多20个
			if (this.recentEmojis.length > 20) {
				this.recentEmojis = this.recentEmojis.slice(0, 20);
			}
			
			// 保存到本地存储
			this.saveRecentEmojis();
		},

		// 保存最近使用的表情
		saveRecentEmojis() {
			try {
				uni.setStorageSync('recent_emojis', JSON.stringify(this.recentEmojis));
			} catch (error) {
				console.error('保存最近表情失败:', error);
			}
		},

		// 加载最近使用的表情
		loadRecentEmojis() {
			try {
				const recent = uni.getStorageSync('recent_emojis');
				if (recent) {
					this.recentEmojis = JSON.parse(recent);
				}
			} catch (error) {
				console.error('加载最近表情失败:', error);
				this.recentEmojis = [];
			}
		},

		// 处理滚动
		handleScroll(e) {
			this.scrollTop = e.detail.scrollTop;
		}
	}
}
</script>

<style lang="scss" scoped>
.emoji-picker-container {
	background: var(--bg-primary);
	border-radius: 32rpx 32rpx 0 0;
	max-height: 80vh;
	min-height: 600rpx;
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

	.emoji-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 40rpx 32rpx 24rpx;
		border-bottom: 1px solid var(--border-color-light);
		background: var(--bg-secondary);
		border-radius: 32rpx 32rpx 0 0;

		.emoji-title {
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

			&:active {
				transform: rotate(90deg) scale(0.95);
			}
		}
	}

	.emoji-tabs {
		display: flex;
		padding: 20rpx 32rpx;
		background: var(--bg-secondary);
		border-bottom: 1px solid var(--border-color-light);
		overflow-x: auto;
		gap: 8rpx;

		.emoji-tab {
			display: flex;
			flex-direction: column;
			align-items: center;
			padding: 16rpx 20rpx;
			border-radius: 16rpx;
			transition: all 0.3s ease;
			cursor: pointer;
			min-width: 100rpx;
			flex-shrink: 0;

			&.active {
				background: var(--primary-color);
				transform: scale(1.05);

				.tab-emoji {
					transform: scale(1.2);
				}

				.tab-name {
					color: #fff;
					font-weight: 600;
				}
			}

			&:active {
				transform: scale(0.95);
			}

			.tab-emoji {
				font-size: 40rpx;
				margin-bottom: 8rpx;
				transition: all 0.3s ease;
			}

			.tab-name {
				font-size: 22rpx;
				color: var(--text-secondary);
				transition: all 0.3s ease;
			}
		}
	}

	.emoji-grid-container {
		flex: 1;
		padding: 24rpx 32rpx;

		.emoji-grid {
			display: grid;
			grid-template-columns: repeat(8, 1fr);
			gap: 16rpx;

			.emoji-item {
				display: flex;
				align-items: center;
				justify-content: center;
				width: 80rpx;
				height: 80rpx;
				border-radius: 16rpx;
				background: var(--bg-secondary);
				transition: all 0.3s ease;
				cursor: pointer;

				&:hover {
					background: var(--bg-tertiary);
					transform: scale(1.1);
				}

				&:active {
					transform: scale(0.9);
					background: var(--primary-color);
				}

				.emoji-char {
					font-size: 48rpx;
					line-height: 1;
				}
			}
		}
	}

	.recent-emojis {
		padding: 24rpx 32rpx;
		border-top: 1px solid var(--border-color-light);
		background: var(--bg-secondary);

		.recent-title {
			display: flex;
			align-items: center;
			gap: 12rpx;
			margin-bottom: 20rpx;
			font-size: 28rpx;
			color: var(--text-secondary);
			font-weight: 600;
		}

		.recent-grid {
			display: flex;
			gap: 16rpx;
			overflow-x: auto;
			padding-bottom: 8rpx;

			.recent-item {
				flex-shrink: 0;
				width: 72rpx;
				height: 72rpx;
				border: 2rpx solid var(--primary-color);

				.emoji-char {
					font-size: 40rpx;
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

/* 响应式设计 */
@media screen and (max-width: 480px) {
	.emoji-picker-container {
		.emoji-grid-container .emoji-grid {
			grid-template-columns: repeat(6, 1fr);
			gap: 12rpx;

			.emoji-item {
				width: 72rpx;
				height: 72rpx;

				.emoji-char {
					font-size: 40rpx;
				}
			}
		}
	}
}
</style>
