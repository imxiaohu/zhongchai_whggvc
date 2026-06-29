// Package config provides HTML and text email templates.
// Part 2 of 2: functions getSecurityAlert through getMonthlyReport.
package config

func getExamNoticeHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>考试通知</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #FF5722, #D84315); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .exam-info { background-color: #ffebee; padding: 20px; border-left: 4px solid #FF5722; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #FF5722; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📝 考试通知</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="exam-info">
                <p><strong>考试课程：</strong><span class="highlight">{{.CourseName}}</span></p>
                <p><strong>考试时间：</strong>{{.ExamTime}}</p>
                <p><strong>考试地点：</strong>{{.Location}}</p>
                <p>请携带相关证件准时参加考试。</p>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getExamNoticeTextTemplate 获取考试通知纯文本模板
func getExamNoticeTextTemplate() string {
	return `考试通知

亲爱的 {{.UserName}} 同学，您好！

考试课程：{{.CourseName}}
考试时间：{{.ExamTime}}
考试地点：{{.Location}}

请携带相关证件准时参加考试。

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getAssignmentDueHTMLTemplate 获取作业截止提醒HTML模板
func getAssignmentDueHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>作业截止提醒</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #FF9800, #F57C00); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .assignment-info { background-color: #fff3e0; padding: 20px; border-left: 4px solid #FF9800; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #FF9800; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📋 作业截止提醒</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="assignment-info">
                <p><strong>课程：</strong>{{.CourseName}}</p>
                <p><strong>作业：</strong><span class="highlight">{{.AssignmentName}}</span></p>
                <p><strong>截止时间：</strong>{{.Deadline}}</p>
                <p>请及时提交作业。</p>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getAssignmentDueTextTemplate 获取作业截止提醒纯文本模板
func getAssignmentDueTextTemplate() string {
	return `作业截止提醒

亲爱的 {{.UserName}} 同学，您好！

课程：{{.CourseName}}
作业：{{.AssignmentName}}
截止时间：{{.Deadline}}

请及时提交作业。

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getEvaluationReminderHTMLTemplate 获取评教提醒HTML模板
func getEvaluationReminderHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>评教提醒</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #FFC107, #FF8F00); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .evaluation-info { background-color: #fffbf0; padding: 20px; border-left: 4px solid #FFC107; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #FFC107; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⭐ 评教提醒</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="evaluation-info">
                <p><strong>学期：</strong>{{.Semester}}</p>
                <p><strong>待评教课程数：</strong><span class="highlight">{{.CourseCount}} 门</span></p>
                <p>课程评教已开始，请及时完成评教。</p>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getEvaluationReminderTextTemplate 获取评教提醒纯文本模板
func getEvaluationReminderTextTemplate() string {
	return `评教提醒

亲爱的 {{.UserName}} 同学，您好！

学期：{{.Semester}}
待评教课程数：{{.CourseCount}} 门

课程评教已开始，请及时完成评教。

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getEvaluationDeadlineHTMLTemplate 获取评教截止提醒HTML模板
func getEvaluationDeadlineHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>评教截止提醒</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #f44336, #d32f2f); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .deadline-info { background-color: #ffebee; padding: 20px; border-left: 4px solid #f44336; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #f44336; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⏰ 评教截止提醒</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="deadline-info">
                <p><strong>截止时间：</strong><span class="highlight">{{.Deadline}}</span></p>
                <p><strong>剩余未评教课程：</strong>{{.RemainingCount}} 门</p>
                <p>课程评教即将截止，请抓紧时间完成。</p>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getEvaluationDeadlineTextTemplate 获取评教截止提醒纯文本模板
func getEvaluationDeadlineTextTemplate() string {
	return `评教截止提醒

亲爱的 {{.UserName}} 同学，您好！

截止时间：{{.Deadline}}
剩余未评教课程：{{.RemainingCount}} 门

课程评教即将截止，请抓紧时间完成。

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getCommunityReplyHTMLTemplate 获取社区回复通知HTML模板
func getCommunityReplyHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>社区回复通知</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #4CAF50, #45a049); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .reply-info { background-color: #e8f5e8; padding: 20px; border-left: 4px solid #4CAF50; border-radius: 4px; margin: 20px 0; }
        .reply-content { background-color: #f9f9f9; padding: 15px; border-radius: 4px; margin: 10px 0; }
        .highlight { color: #4CAF50; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>💬 社区回复通知</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="reply-info">
                <p><strong>回复者：</strong><span class="highlight">{{.ReplyUser}}</span></p>
                <p><strong>帖子标题：</strong>{{.PostTitle}}</p>
                <div class="reply-content">
                    <strong>回复内容：</strong><br>
                    {{.ReplyContent}}
                </div>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getCommunityReplyTextTemplate 获取社区回复通知纯文本模板
func getCommunityReplyTextTemplate() string {
	return `社区回复通知

亲爱的 {{.UserName}} 同学，您好！

回复者：{{.ReplyUser}}
帖子标题：{{.PostTitle}}

回复内容：
{{.ReplyContent}}

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getCommunityMentionHTMLTemplate 获取社区@提醒HTML模板
func getCommunityMentionHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>社区@提醒</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #2196F3, #1976D2); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .mention-info { background-color: #e3f2fd; padding: 20px; border-left: 4px solid #2196F3; border-radius: 4px; margin: 20px 0; }
        .mention-content { background-color: #f9f9f9; padding: 15px; border-radius: 4px; margin: 10px 0; }
        .highlight { color: #2196F3; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📢 社区@提醒</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="mention-info">
                <p><strong>@您的用户：</strong><span class="highlight">{{.MentionUser}}</span></p>
                <p><strong>帖子标题：</strong>{{.PostTitle}}</p>
                <div class="mention-content">
                    <strong>内容：</strong><br>
                    {{.Content}}
                </div>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getCommunityMentionTextTemplate 获取社区@提醒纯文本模板
func getCommunityMentionTextTemplate() string {
	return `社区@提醒

亲爱的 {{.UserName}} 同学，您好！

@您的用户：{{.MentionUser}}
帖子标题：{{.PostTitle}}

内容：
{{.Content}}

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getRechargeSuccessHTMLTemplate 获取充值成功HTML模板
func getRechargeSuccessHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>充值成功通知</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #4CAF50, #45a049); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .recharge-info { background-color: #e8f5e8; padding: 20px; border-left: 4px solid #4CAF50; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #4CAF50; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>💰 充值成功通知</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="recharge-info">
                <p><strong>充值金额：</strong><span class="highlight">¥{{.Amount}}</span></p>
                <p><strong>当前余额：</strong>¥{{.Balance}}</p>
                <p>充值成功，感谢您的支持！</p>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getRechargeSuccessTextTemplate 获取充值成功纯文本模板
func getRechargeSuccessTextTemplate() string {
	return `充值成功通知

亲爱的 {{.UserName}} 同学，您好！

充值金额：¥{{.Amount}}
当前余额：¥{{.Balance}}

充值成功，感谢您的支持！

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getBalanceLowHTMLTemplate 获取余额不足提醒HTML模板
func getBalanceLowHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>余额不足提醒</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #FF9800, #F57C00); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .balance-info { background-color: #fff3e0; padding: 20px; border-left: 4px solid #FF9800; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #FF9800; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⚠️ 余额不足提醒</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="balance-info">
                <p><strong>当前余额：</strong><span class="highlight">¥{{.Balance}}</span></p>
                <p>您的余额不足，为避免影响服务使用，请及时充值。</p>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getBalanceLowTextTemplate 获取余额不足提醒纯文本模板
func getBalanceLowTextTemplate() string {
	return `余额不足提醒

亲爱的 {{.UserName}} 同学，您好！

当前余额：¥{{.Balance}}

您的余额不足，为避免影响服务使用，请及时充值。

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getVerificationCodeHTMLTemplate 获取验证码HTML模板
func getVerificationCodeHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>验证码</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #2196F3, #1976D2); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .code-info { background-color: #e3f2fd; padding: 20px; border-left: 4px solid #2196F3; border-radius: 4px; margin: 20px 0; text-align: center; }
        .code { font-size: 32px; font-weight: bold; color: #2196F3; letter-spacing: 8px; margin: 20px 0; }
        .highlight { color: #2196F3; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 验证码</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="code-info">
                <p>您的验证码是：</p>
                <div class="code">{{.Code}}</div>
                <p>验证码 <strong>{{.Minutes}} 分钟</strong> 内有效，请勿泄露给他人。</p>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getVerificationCodeTextTemplate 获取验证码纯文本模板
func getVerificationCodeTextTemplate() string {
	return `验证码

亲爱的 {{.UserName}} 同学，您好！

您的验证码是：{{.Code}}

验证码 {{.Minutes}} 分钟内有效，请勿泄露给他人。

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getWeeklyReportHTMLTemplate 获取周报HTML模板
func getWeeklyReportHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>周报</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #9C27B0, #7B1FA2); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .report-info { background-color: #f3e5f5; padding: 20px; border-left: 4px solid #9C27B0; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #9C27B0; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 周报</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="report-info">
                <p><strong>统计周期：</strong>{{.WeekRange}}</p>
                <div>
                    <strong>本周数据：</strong><br>
                    {{.ReportData}}
                </div>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getWeeklyReportTextTemplate 获取周报纯文本模板
func getWeeklyReportTextTemplate() string {
	return `周报

亲爱的 {{.UserName}} 同学，您好！

统计周期：{{.WeekRange}}

本周数据：
{{.ReportData}}

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getMonthlyReportHTMLTemplate 获取月报HTML模板
func getMonthlyReportHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>月报</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #607D8B, #455A64); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .report-info { background-color: #eceff1; padding: 20px; border-left: 4px solid #607D8B; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #607D8B; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📈 月报</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="report-info">
                <p><strong>统计月份：</strong>{{.Month}}</p>
                <div>
                    <strong>本月数据：</strong><br>
                    {{.ReportData}}
                </div>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getMonthlyReportTextTemplate 获取月报纯文本模板
func getMonthlyReportTextTemplate() string {
	return `月报

亲爱的 {{.UserName}} 同学，您好！

统计月份：{{.Month}}

本月数据：
{{.ReportData}}

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}
