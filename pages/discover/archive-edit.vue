<template>
	<view class="archive-edit-page">
		<!-- 自定义导航栏 -->
		<view class="nav-bar" :style="{paddingTop: statusBarHeight + 'px'}">
			<view class="nav-bar__content">
				<view class="nav-bar__back" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
				</view>
				<text class="nav-bar__title">完善档案</text>
				<view class="nav-bar__action" @tap="submitForm">
					<text class="nav-bar__save">保存</text>
				</view>
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
				<button class="state-btn" @tap="loadData">重试</button>
			</view>

			<!-- 表单 -->
			<view v-else class="form-content">

				<!-- 学业信息（只读展示） -->
				<view class="section-card">
					<view class="section-title">
						<l-icon name="school" style="font-size: 16px; color: #8b5cf6; margin-right: 8rpx;"></l-icon>
						<text>学业信息</text>
					</view>

					<view class="info-readonly-grid">
						<view class="readonly-item">
							<text class="readonly-item__label">校区</text>
							<text class="readonly-item__value">{{ academicInfo.campus || '--' }}</text>
						</view>
						<view class="readonly-item">
							<text class="readonly-item__label">学籍状态</text>
							<text class="readonly-item__value">{{ academicInfo.studentStatus || '--' }}</text>
						</view>
						<view class="readonly-item">
							<text class="readonly-item__label">院系名称</text>
							<text class="readonly-item__value">{{ academicInfo.department || '--' }}</text>
						</view>
						<view class="readonly-item">
							<text class="readonly-item__label">入学日期</text>
							<text class="readonly-item__value">{{ academicInfo.entranceDate || '--' }}</text>
						</view>
						<view class="readonly-item">
							<text class="readonly-item__label">专业名称</text>
							<text class="readonly-item__value">{{ academicInfo.major || '--' }}</text>
						</view>
						<view class="readonly-item">
							<text class="readonly-item__label">预计毕业日期</text>
							<text class="readonly-item__value">{{ academicInfo.expectedGraduationDate || '--' }}</text>
						</view>
						<view class="readonly-item">
							<text class="readonly-item__label">行政班</text>
							<text class="readonly-item__value">{{ academicInfo.className || '--' }}</text>
						</view>
						<view class="readonly-item">
							<text class="readonly-item__label">学习形式</text>
							<text class="readonly-item__value">{{ academicInfo.learningForm || '--' }}</text>
						</view>
						<view class="readonly-item">
							<text class="readonly-item__label">年级</text>
							<text class="readonly-item__value">{{ academicInfo.grade || '--' }}</text>
						</view>
						<view class="readonly-item">
							<text class="readonly-item__label">学制</text>
							<text class="readonly-item__value">{{ academicInfo.academicSystem || '--' }}</text>
						</view>
						<view class="readonly-item">
							<text class="readonly-item__label">录取类别</text>
							<text class="readonly-item__value">{{ academicInfo.admissionCategory || '--' }}</text>
						</view>
					</view>
				</view>

				<!-- 联系方式 -->
				<view class="section-card">
					<view class="section-title">
						<l-icon name="cellphone" style="font-size: 16px; color: #8b5cf6; margin-right: 8rpx;"></l-icon>
						<text>联系方式</text>
					</view>

					<view class="field-row">
						<view class="field-label">手机号码</view>
						<t-input
							class="field-tinput"
							v-model="form.phoneNumber"
							type="number"
							placeholder="请输入手机号码"
							:clearable="true"
						/>
					</view>

					<view class="field-row">
						<view class="field-label">联系地址</view>
						<t-input
							class="field-tinput"
							v-model="form.personalAddress"
							placeholder="请输入联系地址"
							:clearable="true"
						/>
					</view>

					<view class="field-row field-row--email">
						<view class="field-label field-label--email">EMAIL</view>
						<t-input
							class="field-tinput field-tinput--email"
							v-model="form.email"
							placeholder="请输入邮箱"
							:clearable="true"
						/>
					</view>

					<view class="field-row">
						<view class="field-label">QQ</view>
						<t-input
							class="field-tinput"
							v-model="form.qq"
							type="number"
							placeholder="请输入QQ号"
							:clearable="true"
						/>
					</view>
				</view>

				<!-- 户口情况 -->
				<view class="section-card">
					<view class="section-title">
						<l-icon name="home" style="font-size: 16px; color: #8b5cf6; margin-right: 8rpx;"></l-icon>
						<text>户口情况</text>
					</view>

					<view class="field-row">
						<view class="field-label">户口类型</view>
						<picker
							class="field-picker"
							:range="householdTypeOptions"
							range-key="label"
							@change="onHouseholdTypeChange"
						>
							<view class="picker-value">
								{{ getHouseholdTypeLabel(form.householdRegistrationType) }}
								<l-icon name="chevron-down" style="font-size: 16px; color: #94a3b8;"></l-icon>
							</view>
						</picker>
					</view>

					<view class="field-row">
						<view class="field-label">户口登记机关</view>
						<t-input
							class="field-tinput"
							v-model="form.householdRegistrationOffice"
							placeholder="请输入户口登记机关"
							:clearable="true"
						/>
					</view>

					<view class="field-row field-row--col">
						<view class="field-label">户籍地</view>
						<view class="field-row-group">
							<picker
								class="field-picker field-picker--half"
								:range="provinceOptions"
								range-key="name"
								@change="onProvinceChange"
							>
								<view class="picker-value">
									{{ form.householdProvince || '省' }}
									<l-icon name="chevron-down" style="font-size: 16px; color: #94a3b8;"></l-icon>
								</view>
							</picker>
							<picker
								class="field-picker field-picker--half"
								:range="cityOptions"
								range-key="name"
								@change="onCityChange"
							>
								<view class="picker-value">
									{{ form.householdCity || '市' }}
									<l-icon name="chevron-down" style="font-size: 16px; color: #94a3b8;"></l-icon>
								</view>
							</picker>
						</view>
					</view>

					<view class="field-row">
						<view class="field-label">户籍地址</view>
						<t-input
							class="field-tinput"
							v-model="form.householdRegistrationAddress"
							placeholder="请输入详细地址"
							:clearable="true"
						/>
					</view>

					<view class="field-row field-row--col">
						<view class="field-label">火车乘车区间</view>
						<view class="field-row-group">
							<t-input
								class="field-tinput field-tinput--half"
								v-model="form.trainStartStation"
								placeholder="出发站"
								:clearable="true"
							/>
							<text class="field-sep">往返</text>
							<t-input
								class="field-tinput field-tinput--half"
								v-model="form.trainStopStation"
								placeholder="到达站"
								:clearable="true"
							/>
						</view>
					</view>
				</view>

				<!-- 家庭资料 -->
				<view class="section-card">
					<view class="section-title">
						<l-icon name="account-group" style="font-size: 16px; color: #8b5cf6; margin-right: 8rpx;"></l-icon>
						<text>家庭资料</text>
					</view>

					<view class="field-row">
						<view class="field-label">家庭地址</view>
						<t-input
							class="field-tinput"
							v-model="form.familyAddress"
							placeholder="请输入家庭地址"
							:clearable="true"
						/>
					</view>

					<view class="field-row">
						<view class="field-label">家庭电话</view>
						<t-input
							class="field-tinput"
							v-model="form.familyPhone"
							placeholder="请输入家庭电话"
							:clearable="true"
						/>
					</view>

					<view class="field-row">
						<view class="field-label">邮政编码</view>
						<t-input
							class="field-tinput"
							v-model="form.familyPost"
							placeholder="请输入邮编"
							:clearable="true"
						/>
					</view>
				</view>

				<!-- 家庭成员 -->
				<view class="section-card">
					<view class="section-title-row">
						<view class="section-title">
							<l-icon name="account-multiple" style="font-size: 16px; color: #8b5cf6; margin-right: 8rpx;"></l-icon>
							<text>家庭成员</text>
						</view>
						<view class="add-btn" @tap="addFamilyMember">
							<l-icon name="plus" style="font-size: 14px; color: #8b5cf6;"></l-icon>
							<text class="add-btn__text">新增</text>
						</view>
					</view>

					<view v-if="form.familyMembers.length === 0" class="empty-tip">
						<text>暂无家庭成员，点击上方按钮添加</text>
					</view>

					<view
						v-for="(member, idx) in form.familyMembers"
						:key="idx"
						class="family-member-card"
					>
						<view class="member-header">
							<text class="member-index">成员 {{ idx + 1 }}</text>
							<view class="member-delete" @tap="removeFamilyMember(idx)">
								<l-icon name="close" style="font-size: 14px; color: #ef4444;"></l-icon>
							</view>
						</view>

						<view class="field-row">
							<view class="field-label">姓名</view>
							<t-input
								class="field-tinput"
								v-model="member.name"
								placeholder="姓名"
								:clearable="true"
							/>
						</view>
						<view class="field-row">
							<view class="field-label">关系</view>
							<t-input
								class="field-tinput"
								v-model="member.relationship"
								placeholder="如：父亲、母亲"
								:clearable="true"
							/>
						</view>
						<view class="field-row">
							<view class="field-label">联系电话</view>
							<t-input
								class="field-tinput"
								v-model="member.mobile"
								type="number"
								placeholder="手机号码"
								:clearable="true"
							/>
						</view>
						<view class="field-row">
							<view class="field-label">工作单位</view>
							<t-input
								class="field-tinput"
								v-model="member.workplace"
								placeholder="工作单位"
								:clearable="true"
							/>
						</view>
						<view class="field-row">
							<view class="field-label">政治面貌</view>
							<t-input
								class="field-tinput"
								v-model="member.politicsStatus"
								placeholder="如：群众、中共党员"
								:clearable="true"
							/>
						</view>
					</view>
				</view>

				<!-- 学校经历 -->
				<view class="section-card">
					<view class="section-title-row">
						<view class="section-title">
							<l-icon name="school" style="font-size: 16px; color: #8b5cf6; margin-right: 8rpx;"></l-icon>
							<text>学校经历</text>
						</view>
						<view class="add-btn" @tap="addSchoolExperience">
							<l-icon name="plus" style="font-size: 14px; color: #8b5cf6;"></l-icon>
							<text class="add-btn__text">新增</text>
						</view>
					</view>

					<view class="exp-tip">
						<text class="exp-tip__text">从初中填起</text>
					</view>

					<view v-if="form.schoolExperiences.length === 0" class="empty-tip">
						<text>暂无学校经历，点击上方按钮添加</text>
					</view>

					<view
						v-for="(exp, idx) in form.schoolExperiences"
						:key="idx"
						class="exp-card"
					>
						<view class="member-header">
							<text class="member-index">经历 {{ idx + 1 }}</text>
							<view class="member-delete" @tap="removeSchoolExperience(idx)">
								<l-icon name="close" style="font-size: 14px; color: #ef4444;"></l-icon>
							</view>
						</view>

						<view class="field-row">
							<view class="field-label">开始时间</view>
							<picker
								class="field-picker"
								mode="date"
								:value="exp.startDate"
								@change="(e) => onExpDateChange(idx, 'startDate', e)"
							>
								<view class="picker-value">
									{{ exp.startDate || '请选择' }}
									<l-icon name="chevron-down" style="font-size: 16px; color: #94a3b8;"></l-icon>
								</view>
							</picker>
						</view>
						<view class="field-row">
							<view class="field-label">结束时间</view>
							<picker
								class="field-picker"
								mode="date"
								:value="exp.endDate"
								@change="(e) => onExpDateChange(idx, 'endDate', e)"
							>
								<view class="picker-value">
									{{ exp.endDate || '请选择' }}
									<l-icon name="chevron-down" style="font-size: 16px; color: #94a3b8;"></l-icon>
								</view>
							</picker>
						</view>
						<view class="field-row">
							<view class="field-label">学校</view>
							<t-input
								class="field-tinput"
								v-model="exp.school"
								placeholder="学校名称"
								:clearable="true"
							/>
						</view>
						<view class="field-row">
							<view class="field-label">在校职务</view>
							<t-input
								class="field-tinput"
								v-model="exp.job"
								placeholder="如：班长、学习委员"
								:clearable="true"
							/>
						</view>
						<view class="field-row">
							<view class="field-label">证明人</view>
							<t-input
								class="field-tinput"
								v-model="exp.proveMan"
								placeholder="证明人姓名"
								:clearable="true"
							/>
						</view>
					</view>
				</view>

				<!-- 提交按钮 -->
				<view class="submit-section">
					<button
						class="submit-btn"
						:disabled="submitting"
						@tap="submitForm"
					>
						<view v-if="submitting" class="btn-spinner"></view>
						<text v-else>保存档案</text>
					</button>
				</view>

				<!-- 错误提示 -->
				<view v-if="errorMsg" class="error-tip">
					<l-icon name="alert-circle" style="font-size: 16px; color: #ef4444; margin-right: 8rpx;"></l-icon>
					<text class="error-tip__text">{{ errorMsg }}</text>
				</view>

				<view class="bottom-safe-area"></view>
			</view>
		</scroll-view>

		<!-- PC验证码弹窗 -->
		<PCCaptchaModal
			:visible="captchaModalVisible"
			:sessionId="captchaSessionId"
			:captchaImage="captchaImage"
			:tips="'登录已过期，请输入验证码'"
			@close="onCaptchaClose"
			@success="onCaptchaSuccess"
			@refresh-captcha="onCaptchaRefresh"
		/>
	</view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { pcSubmitArchiveEdit, pcGetStudentInfo } from '../../pages/api/discover.js'
import PCCaptchaModal from '@/components/PCCaptchaModal.vue'

const statusBarHeight = ref(20)
const loading = ref(true)
const error = ref('')
const submitting = ref(false)
const errorMsg = ref('')
const captchaModalVisible = ref(false)
const captchaSessionId = ref('')
const captchaImage = ref('')

const academicInfo = ref({
	campus: '',
	studentStatus: '',
	department: '',
	entranceDate: '',
	major: '',
	expectedGraduationDate: '',
	className: '',
	learningForm: '',
	grade: '',
	academicSystem: '',
	admissionCategory: ''
})

const householdTypeOptions = [
	{ label: '请选择', value: '' },
	{ label: '农村', value: '0' },
	{ label: '城市', value: '1' }
]

const form = ref({
	phoneNumber: '',
	personalAddress: '',
	email: '',
	qq: '',
	householdRegistrationType: '',
	householdRegistrationOffice: '',
	householdProvince: '',
	householdProvinceID: '',
	householdCity: '',
	householdCityID: '',
	householdRegistrationAddress: '',
	trainStartStation: '',
	trainStopStation: '',
	familyAddress: '',
	familyPhone: '',
	familyPost: '',
	familyMembers: [],
	schoolExperiences: []
})

const provinceOptions = ref([
	{ name: '北京市', code: '110000' },
	{ name: '天津市', code: '120000' },
	{ name: '河北省', code: '130000' },
	{ name: '山西省', code: '140000' },
	{ name: '内蒙古', code: '150000' },
	{ name: '辽宁省', code: '210000' },
	{ name: '吉林省', code: '220000' },
	{ name: '黑龙江省', code: '230000' },
	{ name: '上海市', code: '310000' },
	{ name: '江苏省', code: '320000' },
	{ name: '浙江省', code: '330000' },
	{ name: '安徽省', code: '340000' },
	{ name: '福建省', code: '350000' },
	{ name: '江西省', code: '360000' },
	{ name: '山东省', code: '370000' },
	{ name: '河南省', code: '410000' },
	{ name: '湖北省', code: '420000' },
	{ name: '湖南省', code: '430000' },
	{ name: '广东省', code: '440000' },
	{ name: '广西', code: '450000' },
	{ name: '海南省', code: '460000' },
	{ name: '重庆市', code: '500000' },
	{ name: '四川省', code: '510000' },
	{ name: '贵州省', code: '520000' },
	{ name: '云南省', code: '530000' },
	{ name: '西藏', code: '540000' },
	{ name: '陕西省', code: '610000' },
	{ name: '甘肃省', code: '620000' },
	{ name: '青海省', code: '630000' },
	{ name: '宁夏', code: '640000' },
	{ name: '新疆', code: '650000' }
])

const cityOptions = ref([])

function goBack() {
	uni.navigateBack()
}

function getHouseholdTypeLabel(val) {
	const item = householdTypeOptions.find(o => o.value === val)
	return item ? item.label : '请选择'
}

function onHouseholdTypeChange(e) {
	form.value.householdRegistrationType = householdTypeOptions[e.detail.value]?.value || ''
}

function onProvinceChange(e) {
	const selected = provinceOptions.value[e.detail.value]
	if (selected) {
		form.value.householdProvince = selected.name
		form.value.householdProvinceID = selected.code
	}
}

function onCityChange(e) {
	const selected = cityOptions.value[e.detail.value]
	if (selected) {
		form.value.householdCity = selected.name
		form.value.householdCityID = selected.code
	}
}

function onExpDateChange(idx, field, e) {
	form.value.schoolExperiences[idx][field] = e.detail.value
}

function addFamilyMember() {
	form.value.familyMembers.push({
		name: '',
		relationship: '',
		isGuardian: '1',
		credentialsType: '',
		credentialsNo: '',
		workplace: '',
		mobile: '',
		politicsStatus: ''
	})
}

function removeFamilyMember(idx) {
	form.value.familyMembers.splice(idx, 1)
}

function addSchoolExperience() {
	form.value.schoolExperiences.push({
		startDate: '',
		endDate: '',
		school: '',
		job: '',
		proveMan: ''
	})
}

function removeSchoolExperience(idx) {
	form.value.schoolExperiences.splice(idx, 1)
}

async function loadData() {
	loading.value = true
	error.value = ''
	try {
		const res = await pcGetStudentInfo()
		if (res && res.success && res.result) {
			const info = res.result

			if (info.needManual) {
				captchaModalVisible.value = true
				captchaSessionId.value = info.sessionId || ''
				captchaImage.value = info.captcha || ''
				return
			}

			// 学业信息（只读）
			if (info.campus) academicInfo.value.campus = info.campus
			if (info.studentStatus) academicInfo.value.studentStatus = info.studentStatus
			if (info.department) academicInfo.value.department = info.department
			if (info.entranceDate) academicInfo.value.entranceDate = info.entranceDate
			if (info.major) academicInfo.value.major = info.major
			if (info.expectedGraduationDate) academicInfo.value.expectedGraduationDate = info.expectedGraduationDate
			if (info.className) academicInfo.value.className = info.className
			if (info.learningForm) academicInfo.value.learningForm = info.learningForm
			if (info.grade) academicInfo.value.grade = info.grade
			if (info.academicSystem) academicInfo.value.academicSystem = info.academicSystem
			if (info.admissionCategory) academicInfo.value.admissionCategory = info.admissionCategory

			if (info.phoneNumber || info.phone) form.value.phoneNumber = info.phoneNumber || info.phone
			if (info.personalAddress) form.value.personalAddress = info.personalAddress
			if (info.email) form.value.email = info.email
			if (info.qq) form.value.qq = info.qq
			if (info.householdRegistrationType !== undefined) form.value.householdRegistrationType = String(info.householdRegistrationType)
			if (info.householdRegistrationOffice) form.value.householdRegistrationOffice = info.householdRegistrationOffice
			if (info.householdRegistrationAddress) form.value.householdRegistrationAddress = info.householdRegistrationAddress
			if (info.trainStartStation) form.value.trainStartStation = info.trainStartStation
			if (info.trainStopStation) form.value.trainStopStation = info.trainStopStation
			if (info.familyAddress) form.value.familyAddress = info.familyAddress
			if (info.familyPhone) form.value.familyPhone = info.familyPhone
			if (info.familyPost) form.value.familyPost = info.familyPost
		}
	} catch (e) {
		error.value = e?.message || '加载失败'
	} finally {
		loading.value = false
	}
}

function onCaptchaClose() {
	captchaModalVisible.value = false
}
async function onCaptchaSuccess() {
	captchaModalVisible.value = false
	await loadData()
}
function onCaptchaRefresh() {
	captchaModalVisible.value = false
	uni.navigateTo({ url: '/pages/discover/pc-login?redirect=/pages/discover/archive-edit' })
}

async function submitForm() {
	if (submitting.value) return
	submitting.value = true
	errorMsg.value = ''

	// 构建提交数据（只提交有值的字段）
	const submitData = {}

	if (form.value.phoneNumber) submitData.phoneNumber = form.value.phoneNumber
	if (form.value.personalAddress) submitData.personalAddress = form.value.personalAddress
	if (form.value.email) submitData.email = form.value.email
	if (form.value.qq) submitData.qq = form.value.qq
	if (form.value.householdRegistrationType) submitData.householdRegistrationType = form.value.householdRegistrationType
	if (form.value.householdRegistrationOffice) submitData.householdRegistrationOffice = form.value.householdRegistrationOffice
	if (form.value.householdProvinceID) submitData.householdRegistrationProvinceID = form.value.householdProvinceID
	if (form.value.householdProvince) submitData.householdProvince = form.value.householdProvince
	if (form.value.householdCityID) submitData.householdRegistrationCityID = form.value.householdCityID
	if (form.value.householdCity) submitData.householdCity = form.value.householdCity
	if (form.value.householdRegistrationAddress) submitData.householdRegistrationAddress = form.value.householdRegistrationAddress
	if (form.value.trainStartStation) submitData.trainStartStation = form.value.trainStartStation
	if (form.value.trainStopStation) submitData.trainStopStation = form.value.trainStopStation
	if (form.value.familyAddress) submitData.familyAddress = form.value.familyAddress
	if (form.value.familyPhone) submitData.familyPhone = form.value.familyPhone
	if (form.value.familyPost) submitData.familyPost = form.value.familyPost

	// 家庭成员
	if (form.value.familyMembers.length > 0) {
		submitData.family = form.value.familyMembers.filter(m => m.name && m.relationship)
	}

	// 学校经历
	if (form.value.schoolExperiences.length > 0) {
		submitData.notes = form.value.schoolExperiences.filter(e => e.school)
	}

	try {
		const res = await pcSubmitArchiveEdit(submitData)
		if (res && res.success && res.result) {
			const result = res.result
			if (result.needManual) {
				captchaModalVisible.value = true
				captchaSessionId.value = result.sessionId || ''
				captchaImage.value = result.captcha || ''
				return
			}
			uni.showToast({ title: '保存成功', icon: 'success' })
			setTimeout(() => uni.navigateBack(), 1000)
		} else {
			errorMsg.value = res?.message || '保存失败，请重试'
		}
	} catch (e) {
		errorMsg.value = e?.message || e?.errMsg || '保存失败，请重试'
	} finally {
		submitting.value = false
	}
}

onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 20
})

onLoad(() => {
	loadData()
})
</script>

<style lang="scss" scoped>
@import url('../../static/css/home-page.css');

.archive-edit-page {
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

.nav-bar__action {
	display: flex;
	align-items: center;
}

.nav-bar__save {
	font-size: 30rpx;
	font-weight: 700;
	color: #fff;
}

.nav-bar__title {
	font-size: 32rpx;
	font-weight: 700;
	color: #fff;
	text-align: center;
}

.content-scroll {
	flex: 1;
	height: 0;
}

.form-content {
	padding: 24rpx;
}

.bottom-safe-area {
	height: 120rpx;
}

/* 状态 */
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

/* 学业信息只读展示 */
.info-readonly-grid {
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: 0;
}

.readonly-item {
	display: flex;
	flex-direction: column;
	padding: 12rpx 0;
	border-bottom: 1px solid rgba(226, 232, 240, 0.5);
}

.readonly-item:nth-last-child(-n+2) {
	border-bottom: none;
}

.readonly-item__label {
	font-size: 24rpx;
	color: var(--text-tertiary);
	margin-bottom: 4rpx;
}

.readonly-item__value {
	font-size: 26rpx;
	color: var(--text-primary);
	font-weight: 600;
}

/* 卡片区块 */
.section-card {
	background: #fff;
	border-radius: 20rpx;
	padding: 24rpx;
	margin-bottom: 20rpx;
	box-shadow: 0 2rpx 12rpx rgba(139, 92, 246, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.section-title {
	display: flex;
	align-items: center;
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	margin-bottom: 24rpx;
}

.section-title-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 20rpx;
}

.add-btn {
	display: flex;
	align-items: center;
	gap: 6rpx;
	padding: 8rpx 20rpx;
	border-radius: 24rpx;
	background: rgba(139, 92, 246, 0.08);
	border: 1px solid rgba(139, 92, 246, 0.2);
}

.add-btn__text {
	font-size: 26rpx;
	font-weight: 600;
	color: #8b5cf6;
}

/* 表单字段 */
.field-row {
	display: flex;
	align-items: center;
	padding: 12rpx 0;
	border-bottom: 1px solid rgba(226, 232, 240, 0.5);
}

.field-row:last-child {
	border-bottom: none;
}

.field-row--col {
	flex-direction: column;
	align-items: flex-start;
	gap: 12rpx;
}

.field-row-group {
	display: flex;
	align-items: center;
	gap: 12rpx;
	width: 100%;
}

.field-label {
	font-size: 28rpx;
	color: var(--text-secondary);
	width: 180rpx;
	flex-shrink: 0;
}

.field-label--email {
	color: #3b82f6;
	font-weight: 700;
	letter-spacing: 0.5px;
}

.field-tinput {
	flex: 1;
}

.field-tinput--half {
	flex: 1;
}

.field-tinput--email {
	font-weight: 600;
	color: #2563eb;
}

.field-sep {
	font-size: 26rpx;
	color: var(--text-tertiary);
	flex-shrink: 0;
}

.field-row--email {
	background: rgba(59, 130, 246, 0.03);
	border-radius: 12rpx;
	padding: 16rpx 16rpx;
	margin: 0 -4rpx;
	border: 1px solid rgba(59, 130, 246, 0.1);
	border-bottom: none;
}

.field-picker {
	flex: 1;
}

.picker-value {
	display: flex;
	align-items: center;
	justify-content: space-between;
	font-size: 28rpx;
	color: var(--text-primary);
	font-weight: 600;
}

.picker-value--half {
	padding: 0 8rpx;
}

.field-picker--half {
	flex: 1;
}

/* 空提示 */
.empty-tip {
	padding: 24rpx 0;
	text-align: center;
	font-size: 26rpx;
	color: var(--text-tertiary);
}

/* 家庭成员/学校经历卡片 */
.family-member-card,
.exp-card {
	background: rgba(139, 92, 246, 0.04);
	border: 1px solid rgba(139, 92, 246, 0.12);
	border-radius: 16rpx;
	padding: 16rpx;
	margin-bottom: 16rpx;
}

.member-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 16rpx;
}

.member-index {
	font-size: 26rpx;
	font-weight: 700;
	color: #8b5cf6;
}

.member-delete {
	width: 44rpx;
	height: 44rpx;
	border-radius: 50%;
	background: rgba(239, 68, 68, 0.08);
	display: flex;
	align-items: center;
	justify-content: center;
}

.exp-tip {
	padding: 0 0 12rpx;
}

.exp-tip__text {
	font-size: 24rpx;
	color: var(--text-tertiary);
}

/* 提交按钮 */
.submit-section {
	padding: 0 0 24rpx;
}

.submit-btn {
	width: 100%;
	height: 88rpx;
	border-radius: 44rpx;
	background: linear-gradient(135deg, #8b5cf6, #a78bfa);
	color: #fff;
	font-size: 30rpx;
	font-weight: 700;
	border: none;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 6rpx 20rpx rgba(139, 92, 246, 0.3);
}

.submit-btn[disabled] {
	opacity: 0.6;
}

.btn-spinner {
	width: 36rpx;
	height: 36rpx;
	border: 4rpx solid rgba(255, 255, 255, 0.3);
	border-top-color: #fff;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

/* 错误提示 */
.error-tip {
	display: flex;
	align-items: center;
	background: rgba(239, 68, 68, 0.06);
	border: 1px solid rgba(239, 68, 68, 0.2);
	border-radius: 12rpx;
	padding: 16rpx 20rpx;
	margin-bottom: 24rpx;
}

.error-tip__text {
	font-size: 26rpx;
	color: #ef4444;
}

@keyframes spin {
	to { transform: rotate(360deg); }
}
</style>
