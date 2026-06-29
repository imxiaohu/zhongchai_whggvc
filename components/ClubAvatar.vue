<template>
	<view class="club-avatar" :style="containerStyle">
		<image v-if="src && !imgError" class="avatar-img" :src="src" mode="aspectFill" @error="onImgError"></image>
		<text v-else class="avatar-letter" :style="letterStyle">{{ letter }}</text>
	</view>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
	src: { type: String, default: '' },
	name: { type: String, default: '' },
	size: { type: [Number, String], default: 100 },
	round: { type: Boolean, default: true },
	bgColor: { type: String, default: '' }
})

const imgError = ref(false)

function onImgError() {
	imgError.value = true
}

watch(() => props.src, () => { imgError.value = false })

const letter = computed(() => {
	if (!props.name) return '?'
	return props.name.trim()[0].toUpperCase()
})

const COLORS = [
	'#4A90E2', '#F5A623', '#D0021B', '#7ED321',
	'#9013FE', '#50E3C2', '#FF6B6B', '#6C5CE7'
]

const letterBgColor = computed(() => {
	if (props.bgColor) return props.bgColor
	const code = props.name
		? props.name.split('').reduce((acc, c) => acc + c.charCodeAt(0), 0)
		: 0
	return COLORS[code % COLORS.length]
})

const containerStyle = computed(() => ({
	width: typeof props.size === 'number' ? props.size + 'rpx' : props.size,
	height: typeof props.size === 'number' ? props.size + 'rpx' : props.size,
	borderRadius: props.round ? '50%' : '8rpx'
}))

const letterStyle = computed(() => ({
	backgroundColor: letterBgColor.value,
	fontSize: typeof props.size === 'number' ? Math.round(props.size * 0.45) + 'rpx' : '45rpx'
}))
</script>

<style scoped>
.club-avatar {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	overflow: hidden;
	flex-shrink: 0;
}

.avatar-img {
	width: 100%;
	height: 100%;
}

.avatar-letter {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 100%;
	height: 100%;
	color: #fff;
	font-weight: 600;
	font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
	letter-spacing: 0;
}
</style>
