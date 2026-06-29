package config

// getScoreUpdateDingTalkTemplate 获取成绩更新钉钉模板
func getScoreUpdateDingTalkTemplate() string {
	return `## 📊 成绩更新通知

**学生：** {{.UserName}}

**更新时间：** {{.UpdateTime}}

### 成绩更新详情

{{range .ScoreUpdates}}
**{{.CourseName}}**
- 成绩类型：{{.ScoreType}}
{{if .OldScore}}- 成绩变化：{{.OldScore}} → **{{.NewScore}}**{{else}}- 成绩：**{{.NewScore}}**{{end}}
- 更新时间：{{.UpdateTime}}

{{end}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getScoreReminderDingTalkTemplate 获取成绩查询提醒钉钉模板
func getScoreReminderDingTalkTemplate() string {
	return `## 📋 成绩查询提醒

**学生：** {{.UserName}}

您有新的成绩可以查询，请及时登录系统查看详细信息。

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getGradePublishedDingTalkTemplate 获取成绩发布通知钉钉模板
func getGradePublishedDingTalkTemplate() string {
	return `## 🎓 成绩发布通知

**学生：** {{.UserName}}

**学期：** {{.Semester}}

**课程：** {{.CourseName}}

该课程成绩已发布，请及时查看。

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getSystemNoticeDingTalkTemplate 获取系统通知钉钉模板
func getSystemNoticeDingTalkTemplate() string {
	return `## 📢 {{.Subject}}

**接收人：** {{.UserName}}

### 通知内容

{{.Content}}

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getMaintenanceNoticeDingTalkTemplate 获取维护通知钉钉模板
func getMaintenanceNoticeDingTalkTemplate() string {
	return `## 🔧 系统维护通知

**维护时间：** {{.MaintenanceTime}}

**预计时长：** {{.Duration}}

维护期间系统服务可能中断，请合理安排使用时间。

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getSecurityAlertDingTalkTemplate 获取安全警告钉钉模板
func getSecurityAlertDingTalkTemplate() string {
	return `## 🔒 安全警告

**用户：** {{.UserName}}

检测到您的账户在 **{{.LoginTime}}** 于 **{{.Location}}** 有异常登录。

如非本人操作，请及时修改密码。

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getAccountBindingDingTalkTemplate 获取账户绑定钉钉模板
func getAccountBindingDingTalkTemplate() string {
	return `## ✅ 账户绑定成功

**用户：** {{.UserName}}

您的账户已成功绑定到 **{{.Username}}**。

如非本人操作，请联系管理员。

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getPasswordResetDingTalkTemplate 获取密码重置钉钉模板
func getPasswordResetDingTalkTemplate() string {
	return `## 🔑 密码重置通知

**用户：** {{.UserName}}

您的账户密码已重置，请使用新密码登录并及时修改。

如非本人操作，请联系管理员。

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getLoginAlertDingTalkTemplate 获取登录提醒钉钉模板
func getLoginAlertDingTalkTemplate() string {
	return `## 🔐 登录提醒

**用户：** {{.UserName}}

您的账户于 **{{.LoginTime}}** 在 **{{.Device}}** 设备登录。

如非本人操作，请及时修改密码。

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getCourseReminderDingTalkTemplate 获取课程提醒钉钉模板
func getCourseReminderDingTalkTemplate() string {
	return `## 📚 课程提醒

**学生：** {{.UserName}}

**课程时间：** {{.CourseTime}}

**课程名称：** {{.CourseName}}

**上课地点：** {{.Location}}

请准时参加课程。

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getExamNoticeDingTalkTemplate 获取考试通知钉钉模板
func getExamNoticeDingTalkTemplate() string {
	return `## 📝 考试通知

**学生：** {{.UserName}}

**考试课程：** {{.CourseName}}

**考试时间：** {{.ExamTime}}

**考试地点：** {{.Location}}

请携带相关证件准时参加考试。

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getAssignmentDueDingTalkTemplate 获取作业截止提醒钉钉模板
func getAssignmentDueDingTalkTemplate() string {
	return `## 📋 作业截止提醒

**学生：** {{.UserName}}

**课程：** {{.CourseName}}

**作业：** {{.AssignmentName}}

**截止时间：** {{.Deadline}}

请及时提交作业。

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getEvaluationReminderDingTalkTemplate 获取评教提醒钉钉模板
func getEvaluationReminderDingTalkTemplate() string {
	return `## ⭐ 评教提醒

**学生：** {{.UserName}}

**学期：** {{.Semester}}

**待评教课程数：** {{.CourseCount}} 门

课程评教已开始，请及时完成评教。

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getEvaluationDeadlineDingTalkTemplate 获取评教截止提醒钉钉模板
func getEvaluationDeadlineDingTalkTemplate() string {
	return `## ⏰ 评教截止提醒

**学生：** {{.UserName}}

**截止时间：** {{.Deadline}}

**剩余未评教课程：** {{.RemainingCount}} 门

课程评教即将截止，请抓紧时间完成。

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getCommunityReplyDingTalkTemplate 获取社区回复通知钉钉模板
func getCommunityReplyDingTalkTemplate() string {
	return `## 💬 社区回复通知

**用户：** {{.UserName}}

**回复者：** {{.ReplyUser}}

**帖子标题：** {{.PostTitle}}

### 回复内容

{{.ReplyContent}}

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getCommunityMentionDingTalkTemplate 获取社区@提醒钉钉模板
func getCommunityMentionDingTalkTemplate() string {
	return `## 📢 社区@提醒

**用户：** {{.UserName}}

**@您的用户：** {{.MentionUser}}

**帖子标题：** {{.PostTitle}}

### 内容

{{.Content}}

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getRechargeSuccessDingTalkTemplate 获取充值成功钉钉模板
func getRechargeSuccessDingTalkTemplate() string {
	return `## 💰 充值成功通知

**用户：** {{.UserName}}

**充值金额：** ¥{{.Amount}}

**当前余额：** ¥{{.Balance}}

充值成功，感谢您的支持！

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getBalanceLowDingTalkTemplate 获取余额不足提醒钉钉模板
func getBalanceLowDingTalkTemplate() string {
	return `## ⚠️ 余额不足提醒

**用户：** {{.UserName}}

**当前余额：** ¥{{.Balance}}

您的余额不足，为避免影响服务使用，请及时充值。

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getWeeklyReportDingTalkTemplate 获取周报钉钉模板
func getWeeklyReportDingTalkTemplate() string {
	return `## 📊 周报

**用户：** {{.UserName}}

**统计周期：** {{.WeekRange}}

### 本周数据

{{.ReportData}}

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}

// getMonthlyReportDingTalkTemplate 获取月报钉钉模板
func getMonthlyReportDingTalkTemplate() string {
	return `## 📈 月报

**用户：** {{.UserName}}

**统计月份：** {{.Month}}

### 本月数据

{{.ReportData}}

**发送时间：** {{.SendTime}}

---

*此消息由 {{.SystemName}} 自动发送*`
}
