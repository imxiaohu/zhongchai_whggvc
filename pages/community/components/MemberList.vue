<template>
	<view class="member-list">
		<!-- 成员管理弹窗 -->
		<t-popup v-model:visible="menuVisible" placement="bottom" :overlay="true" :close-on-overlay-click="true">
			<view class="member-menu" v-if="currentMember">
				<view class="menu-header">
					<text class="menu-title">成员管理</text>
					<view class="close-btn" @tap="closeMenu">
						<l-icon name="close" size="20" color="var(--text-secondary)"></l-icon>
					</view>
				</view>
				<view class="menu-content">
					<view
						v-if="currentMember.role !== 'admin'"
						class="menu-item"
						@tap="updateRole('admin')"
					>
						<l-icon name="user-add" size="20" color="var(--primary-color)"></l-icon>
						<text class="menu-text">设为管理员</text>
					</view>
					<view
						v-if="currentMember.role === 'admin'"
						class="menu-item"
						@tap="updateRole('member')"
					>
						<l-icon name="user" size="20" color="var(--warning-color)"></l-icon>
						<text class="menu-text">撤销管理员</text>
					</view>
					<view class="menu-item danger" @tap="removeMember">
						<l-icon name="delete" size="20" color="var(--error-color)"></l-icon>
						<text class="menu-text">移除成员</text>
					</view>
				</view>
			</view>
		</t-popup>

		<!-- Grid 布局 -->
		<view v-if="layout === 'grid'" class="member-grid">
			<view
				v-for="member in members"
				:key="member.id"
				class="member-card"
				@tap="$emit('memberClick', member)"
			>
				<view class="card-avatar-wrap">
					<UserAvatar
						:src="member.user?.avatar"
						:name="member.user?.realname || member.user?.nickname"
						:size="120"
						class="card-avatar"
					></UserAvatar>
					<view v-if="member.role !== 'member'" class="card-role-dot" :class="getRoleClass(member.role)"></view>
					<view v-if="isAdmin && member.role !== 'creator'" class="card-more" @tap.stop="showMemberMenu(member)">
						<l-icon name="more" size="14" color="var(--text-secondary)"></l-icon>
					</view>
				</view>
				<text class="card-name">{{ member.user?.realname || member.user?.nickname || '未知' }}</text>
				<text v-if="member.role !== 'member'" class="card-role-tag" :class="getRoleClass(member.role)">
					{{ getRoleText(member.role) }}
				</text>
			</view>
		</view>

		<!-- List 布局 -->
		<view v-else class="member-list-layout">
			<view
				v-for="member in members"
				:key="member.id"
				class="member-item"
				@tap="$emit('memberClick', member)"
			>
				<UserAvatar
					:src="member.user?.avatar"
					:name="member.user?.realname || member.user?.nickname"
					:size="96"
					class="member-avatar"
				></UserAvatar>
				<view class="member-info">
					<view class="member-name-row">
						<text class="member-name">{{ member.user?.realname || member.user?.nickname }}</text>
						<view class="role-badge" :class="getRoleClass(member.role)">
							<text class="role-text">{{ getRoleText(member.role) }}</text>
						</view>
					</view>
					<view class="member-meta">
						<text class="join-time">加入时间: {{ formatTime(member.createdAt) }}</text>
					</view>
					<view v-if="member.user?.signature" class="member-signature">
						<text class="signature-text">{{ member.user.signature }}</text>
					</view>
				</view>
				<view v-if="isAdmin && member.role !== 'creator'" class="member-actions">
					<view class="action-btn" @tap.stop="showMemberMenu(member)">
						<l-icon name="more" size="20" color="var(--text-secondary)"></l-icon>
					</view>
				</view>
			</view>
		</view>

		<view v-if="loading" class="loading-container">
			<t-loading theme="circular" size="24rpx" color="var(--primary-color)"></t-loading>
		</view>

		<view v-if="!loading && members.length === 0" class="empty-state">
			<image class="empty-image" src="/static/images/empty-members.png" mode="aspectFit"></image>
			<text class="empty-text">暂无成员</text>
		</view>
	</view>
</template>

<script setup>
import UserAvatar from '@/components/UserAvatar.vue'
import { ref } from 'vue'
import { showModal } from '../../../pages/api/page.js'
import { useTimeFormat } from '@/composables/useTimeFormat.js'

const props = defineProps({
	loading: { type: Boolean, default: false },
	members: { type: Array, default: () => [] },
	isAdmin: { type: Boolean, default: false },
	layout: { type: String, default: 'grid' }
})

const emit = defineEmits(['memberClick', 'updateRole', 'removeMember'])

const { formatTime } = useTimeFormat()

const menuVisible = ref(false)
const currentMember = ref(null)

function getRoleText(role) {
	const roleMap = {
		'admin': '管理员',
		'member': '成员',
		'creator': '创建者'
	}
	return roleMap[role] || '成员'
}

function getRoleClass(role) {
	return `role-${role}`
}

function showMemberMenu(member) {
	currentMember.value = member
	menuVisible.value = true
}

function closeMenu() {
	menuVisible.value = false
}

async function updateRole(newRole) {
	if (!currentMember.value) return
	const actionText = newRole === 'admin' ? '设为管理员' : '撤销管理员'
	const confirmed = await showModal({
		title: '确定',
		content: `确定${actionText}?`
	})
	if (confirmed?.confirm) {
		emit('updateRole', currentMember.value, newRole)
	}
	closeMenu()
}

function removeMember() {
	if (!currentMember.value) return
	emit('removeMember', currentMember.value)
	closeMenu()
}
</script>

<style lang="scss" scoped>
.member-list {
	padding: var(--spacing-sm) var(--spacing-md);
}

/* ===== Grid Layout ===== */
.member-grid {
	display: grid;
	grid-template-columns: repeat(5, 1fr);
	gap: var(--spacing-sm);
	padding: var(--spacing-xs);
}

.member-card {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: var(--spacing-sm) var(--spacing-xs);
	background: var(--bg-card);
	border-radius: var(--radius-lg);
	border: 1px solid var(--border-primary);
	transition: transform 0.15s ease, box-shadow 0.15s ease;
	position: relative;

	&:active {
		transform: scale(0.96);
	}
}

.card-avatar-wrap {
	position: relative;
	margin-bottom: 6rpx;

	.card-avatar {
		display: block;
	}

	.card-role-dot {
		position: absolute;
		bottom: 0;
		right: 0;
		width: 22rpx;
		height: 22rpx;
		border-radius: 50%;
		border: 3rpx solid var(--bg-card);

		&.role-admin { background: #6366f1; }
		&.role-creator { background: #f59e0b; }
	}

	.card-more {
		position: absolute;
		top: -8rpx;
		right: -8rpx;
		width: 40rpx;
		height: 40rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(255, 255, 255, 0.9);
		border-radius: 50%;
		backdrop-filter: blur(8rpx);
		box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.08);
	}
}

.card-name {
	font-size: 24rpx;
	font-weight: 500;
	color: var(--text-primary);
	max-width: 100%;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
	text-align: center;
	margin-bottom: 2rpx;
}

.card-role-tag {
	font-size: 18rpx;
	font-weight: 600;
	padding: 1rpx 10rpx;
	border-radius: 20rpx;

	&.role-admin { background: rgba(99, 102, 241, 0.12); color: #6366f1; }
	&.role-creator { background: rgba(245, 158, 11, 0.12); color: #f59e0b; }
	&.role-member { display: none; }
}

/* ===== List Layout ===== */
.member-list-layout {
	display: flex;
	flex-direction: column;
	gap: var(--spacing-sm);
}

.member-item {
	display: flex;
	align-items: center;
	padding: var(--spacing-md);
	background: var(--bg-card);
	border-radius: var(--radius-lg);
	border: 1px solid var(--border-primary);

	.member-avatar {
		margin-right: var(--spacing-sm);
		flex-shrink: 0;
	}

	&:active {
		opacity: 0.85;
	}
}

.member-info {
	flex: 1;
	min-width: 0;
}

.member-name-row {
	display: flex;
	align-items: center;
	gap: var(--spacing-xs);
	margin-bottom: 4px;
}

.member-name {
	font-size: var(--font-size-sm);
	font-weight: 600;
	color: var(--text-primary);
}

.role-badge {
	padding: 2px 8px;
	border-radius: var(--radius-sm);
	font-size: 10px;
	font-weight: 600;

	&.role-admin { background: rgba(99, 102, 241, 0.1); color: #6366f1; }
	&.role-creator { background: rgba(245, 158, 11, 0.1); color: #f59e0b; }
	&.role-member { background: var(--bg-secondary); color: var(--text-secondary); }
}

.member-meta {
	margin-bottom: 4px;
}

.join-time {
	font-size: var(--font-size-xs);
	color: var(--text-tertiary);
}

.member-signature {
	.signature-text {
		font-size: var(--font-size-xs);
		color: var(--text-secondary);
		display: -webkit-box;
		-webkit-box-orient: vertical;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		overflow: hidden;
	}
}

.member-actions {
	display: flex;
	align-items: center;
	gap: var(--spacing-xs);
}

.action-btn {
	padding: var(--spacing-xs);
}

/* ===== Common ===== */
.member-menu {
	background: var(--bg-card);
	border-radius: var(--radius-xl) var(--radius-xl) 0 0;
	overflow: hidden;
}

.menu-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: var(--spacing-md);
	border-bottom: 1px solid var(--border-primary);

	.menu-title {
		font-size: var(--font-size-md);
		font-weight: 600;
		color: var(--text-primary);
	}

	.close-btn {
		padding: 8px;
		border-radius: 50%;
		background: var(--bg-secondary);
	}
}

.menu-content {
	padding: var(--spacing-sm) 0;
}

.menu-item {
	display: flex;
	align-items: center;
	gap: var(--spacing-sm);
	padding: var(--spacing-md);

	.menu-text {
		font-size: var(--font-size-sm);
		color: var(--text-primary);
	}

	&.danger .menu-text {
		color: var(--error-color);
	}
}

.loading-container,
.empty-state {
	padding: var(--spacing-md);
}

.empty-state {
	display: flex;
	flex-direction: column;
	align-items: center;
}

.empty-image {
	width: 80px;
	height: 80px;
	opacity: 0.5;
	margin-bottom: var(--spacing-sm);
}

.empty-text {
	font-size: var(--font-size-sm);
	color: var(--text-tertiary);
}
</style>
