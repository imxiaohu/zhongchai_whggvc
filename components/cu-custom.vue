<template>
	<view>
		<view class="cu-custom" :style="[{height: CustomBar + 'px'}]">
			<view class="cu-bar fixed" :style="barStyle" :class="[bgImage != '' ? 'none-bg text-white bg-img' : '', bgColor]">
				<view class="action" @tap="BackPage" v-if="isBack">
					<text class="cuIcon-back"></text>
					<slot name="backText"></slot>
				</view>
				<view class="content" :style="[{top: StatusBar + 'px'}]">
					<slot name="content"></slot>
				</view>
				<slot name="right"></slot>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

const props = defineProps({
	bgColor: {
		type: String,
		default: ''
	},
	isBack: {
		type: Boolean,
		default: false
	},
	bgImage: {
		type: String,
		default: ''
	}
})

const StatusBar = ref(0)
const CustomBar = ref(0)

const barStyle = computed(() => {
	let style = `height:${CustomBar.value}px;padding-top:${StatusBar.value}px;`
	if (props.bgImage) {
		style = `${style}background-image:url(${props.bgImage});`
	}
	return style
})

onMounted(() => {
	StatusBar.value = uni.getSystemInfoSync().statusBarHeight || 0
	CustomBar.value = StatusBar.value + 45
})

function BackPage() {
	uni.navigateBack({ delta: 1 })
}
</script>

<style scoped>
/* Import theme styles */

.cu-custom {
	display: block;
	position: relative;
}

.cu-bar {
	display: flex;
	position: relative;
	align-items: center;
	min-height: 100rpx;
	justify-content: space-between;
	background-color: var(--nav-bg-color, #ffffff);
	color: var(--nav-text-color, #333333);
	transition: background-color 0.3s ease, color 0.3s ease;
}

.cu-bar.fixed {
	position: fixed;
	width: 100%;
	top: 0;
	z-index: 1024;
	box-shadow: 0 1rpx 6rpx rgba(0, 0, 0, 0.1);
}

.cu-bar .action {
	display: flex;
	align-items: center;
	height: 100%;
	justify-content: center;
	max-width: 100%;
	margin: 0 30rpx;
}

.cu-bar .content {
	position: absolute;
	text-align: center;
	width: calc(100% - 340rpx);
	left: 0;
	right: 0;
	bottom: 0;
	top: 0;
	margin: auto;
	height: 60rpx;
	font-size: 32rpx;
	line-height: 60rpx;
	font-weight: 500;
	text-overflow: ellipsis;
	white-space: nowrap;
	overflow: hidden;
}

.cu-bar .action .cuIcon-back {
	font-size: 36rpx;
	margin-right: 10rpx;
}

.bg-gradual-blue {
	background-image: linear-gradient(45deg, #0081ff, #1cbbb4);
	color: #ffffff;
}

.bg-img {
	background-size: cover;
	background-position: center;
	background-repeat: no-repeat;
}

.text-white {
	color: #ffffff !important;
}

.none-bg {
	background: transparent !important;
}

/* 浅色主题 */
.theme-light .cu-bar {
	background-color: #ffffff;
	color: #333333;
}

.theme-light .cu-bar.fixed {
	box-shadow: 0 1rpx 6rpx rgba(0, 0, 0, 0.1);
}

.theme-light .cu-bar .content {
	color: #333333;
}

.theme-light .cu-bar .action .cuIcon-back {
	color: #333333;
}

/* 深色主题 */
.theme-dark .cu-bar {
	background-color: #1c1c1e;
	color: #ffffff;
}

.theme-dark .cu-bar.fixed {
	box-shadow: 0 1rpx 6rpx rgba(255, 255, 255, 0.1);
}

.theme-dark .cu-bar .content {
	color: #ffffff;
}

.theme-dark .cu-bar .action .cuIcon-back {
	color: #ffffff;
}
</style>
