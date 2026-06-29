// Package config provides HTML and text email templates.
// Part 1 of 2: functions getScoreUpdate through getMaintenanceNotice.
package config

// getScoreUpdateHTMLTemplate 获取成绩更新HTML模板
func getScoreUpdateHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>成绩更新通知</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #4CAF50, #45a049); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .score-item { background-color: #f9f9f9; margin: 15px 0; padding: 20px; border-left: 4px solid #4CAF50; border-radius: 4px; }
        .score-item h3 { margin: 0 0 10px 0; color: #4CAF50; font-size: 18px; }
        .score-detail { margin: 8px 0; }
        .score-detail strong { color: #333; }
        .highlight { color: #4CAF50; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
        .btn { display: inline-block; padding: 12px 24px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .btn:hover { background-color: #45a049; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 成绩更新通知</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <p>您有新的成绩更新，详情如下：</p>
            
            {{range .ScoreUpdates}}
            <div class="score-item">
                <h3>{{.CourseName}}</h3>
                <div class="score-detail"><strong>成绩类型：</strong>{{.ScoreType}}</div>
                {{if .OldScore}}
                <div class="score-detail"><strong>成绩变化：</strong>{{.OldScore}} → <span class="highlight">{{.NewScore}}</span></div>
                {{else}}
                <div class="score-detail"><strong>成绩：</strong><span class="highlight">{{.NewScore}}</span></div>
                {{end}}
                <div class="score-detail"><strong>更新时间：</strong>{{.UpdateTime}}</div>
            </div>
            {{end}}
            
            <p>请及时登录系统查看详细信息。</p>
            <a href="#" class="btn">查看详情</a>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.UpdateTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getScoreUpdateTextTemplate 获取成绩更新纯文本模板
func getScoreUpdateTextTemplate() string {
	return `成绩更新通知

亲爱的 {{.UserName}} 同学，您好！

您有新的成绩更新，详情如下：

{{range .ScoreUpdates}}
课程：{{.CourseName}}
成绩类型：{{.ScoreType}}
{{if .OldScore}}成绩变化：{{.OldScore}} → {{.NewScore}}{{else}}成绩：{{.NewScore}}{{end}}
更新时间：{{.UpdateTime}}

{{end}}
请及时登录系统查看详细信息。

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.UpdateTime}}`
}

// getScoreReminderHTMLTemplate 获取成绩查询提醒HTML模板
func getScoreReminderHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>成绩查询提醒</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #2196F3, #1976D2); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .highlight { color: #2196F3; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
        .btn { display: inline-block; padding: 12px 24px; background-color: #2196F3; color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .btn:hover { background-color: #1976D2; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📋 成绩查询提醒</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <p>您有新的成绩可以查询，请及时登录系统查看详细信息。</p>
            <a href="#" class="btn">立即查看</a>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getScoreReminderTextTemplate 获取成绩查询提醒纯文本模板
func getScoreReminderTextTemplate() string {
	return `成绩查询提醒

亲爱的 {{.UserName}} 同学，您好！

您有新的成绩可以查询，请及时登录系统查看详细信息。

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getGradePublishedHTMLTemplate 获取成绩发布通知HTML模板
func getGradePublishedHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>成绩发布通知</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #FF9800, #F57C00); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .course-info { background-color: #fff3e0; padding: 20px; border-left: 4px solid #FF9800; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #FF9800; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
        .btn { display: inline-block; padding: 12px 24px; background-color: #FF9800; color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .btn:hover { background-color: #F57C00; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎓 成绩发布通知</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="course-info">
                <p><strong>学期：</strong>{{.Semester}}</p>
                <p><strong>课程：</strong><span class="highlight">{{.CourseName}}</span></p>
                <p>该课程成绩已发布，请及时查看。</p>
            </div>
            <a href="#" class="btn">查看成绩</a>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getGradePublishedTextTemplate 获取成绩发布通知纯文本模板
func getGradePublishedTextTemplate() string {
	return `成绩发布通知

亲爱的 {{.UserName}} 同学，您好！

学期：{{.Semester}}
课程：{{.CourseName}}

该课程成绩已发布，请及时查看。

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getSystemNoticeHTMLTemplate 获取系统通知HTML模板
func getSystemNoticeHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Subject}}</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #2196F3, #1976D2); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .notice-content { background-color: #e3f2fd; padding: 20px; border-left: 4px solid #2196F3; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #2196F3; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📢 {{.Subject}}</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="notice-content">
                {{.Content}}
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

// getSystemNoticeTextTemplate 获取系统通知纯文本模板
func getSystemNoticeTextTemplate() string {
	return `{{.Subject}}

亲爱的 {{.UserName}} 同学，您好！

{{.Content}}

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getMaintenanceNoticeHTMLTemplate 获取维护通知HTML模板
func getMaintenanceNoticeHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>系统维护通知</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #FF5722, #D84315); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .maintenance-info { background-color: #ffebee; padding: 20px; border-left: 4px solid #FF5722; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #FF5722; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔧 系统维护通知</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="maintenance-info">
                <p><strong>维护时间：</strong><span class="highlight">{{.MaintenanceTime}}</span></p>
                <p><strong>预计时长：</strong>{{.Duration}}</p>
                <p>维护期间系统服务可能中断，请合理安排使用时间。</p>
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

// getMaintenanceNoticeTextTemplate 获取维护通知纯文本模板
func getMaintenanceNoticeTextTemplate() string {
	return `系统维护通知

亲爱的 {{.UserName}} 同学，您好！

维护时间：{{.MaintenanceTime}}
预计时长：{{.Duration}}

维护期间系统服务可能中断，请合理安排使用时间。

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getSecurityAlertHTMLTemplate 获取安全警告HTML模板
func getSecurityAlertHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>安全警告</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #f44336, #d32f2f); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .alert-info { background-color: #ffebee; padding: 20px; border-left: 4px solid #f44336; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #f44336; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
        .btn { display: inline-block; padding: 12px 24px; background-color: #f44336; color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .btn:hover { background-color: #d32f2f; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔒 安全警告</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="alert-info">
                <p>检测到您的账户在 <strong>{{.LoginTime}}</strong> 于 <strong>{{.Location}}</strong> 有异常登录。</p>
                <p>如非本人操作，请及时修改密码。</p>
            </div>
            <a href="#" class="btn">立即修改密码</a>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getSecurityAlertTextTemplate 获取安全警告纯文本模板
func getSecurityAlertTextTemplate() string {
	return `安全警告

亲爱的 {{.UserName}} 同学，您好！

检测到您的账户在 {{.LoginTime}} 于 {{.Location}} 有异常登录。

如非本人操作，请及时修改密码。

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getAccountBindingHTMLTemplate 获取账户绑定HTML模板
func getAccountBindingHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>账户绑定成功</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #4CAF50, #45a049); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .binding-info { background-color: #e8f5e8; padding: 20px; border-left: 4px solid #4CAF50; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #4CAF50; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ 账户绑定成功</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="binding-info">
                <p>您的账户已成功绑定到 <strong>{{.Username}}</strong>。</p>
                <p>如非本人操作，请联系管理员。</p>
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

// getAccountBindingTextTemplate 获取账户绑定纯文本模板
func getAccountBindingTextTemplate() string {
	return `账户绑定成功

亲爱的 {{.UserName}} 同学，您好！

您的账户已成功绑定到 {{.Username}}。

如非本人操作，请联系管理员。

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getPasswordResetHTMLTemplate 获取密码重置HTML模板
func getPasswordResetHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>密码重置通知</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #FF9800, #F57C00); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .reset-info { background-color: #fff3e0; padding: 20px; border-left: 4px solid #FF9800; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #FF9800; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
        .btn { display: inline-block; padding: 12px 24px; background-color: #FF9800; color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .btn:hover { background-color: #F57C00; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔑 密码重置通知</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="reset-info">
                <p>您的账户密码已重置，请使用新密码登录并及时修改。</p>
                <p>如非本人操作，请联系管理员。</p>
            </div>
            <a href="#" class="btn">立即登录</a>
        </div>
        <div class="footer">
            <p>此邮件由 {{.SystemName}} 自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>`
}

// getPasswordResetTextTemplate 获取密码重置纯文本模板
func getPasswordResetTextTemplate() string {
	return `密码重置通知

亲爱的 {{.UserName}} 同学，您好！

您的账户密码已重置，请使用新密码登录并及时修改。

如非本人操作，请联系管理员。

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getLoginAlertHTMLTemplate 获取登录提醒HTML模板
func getLoginAlertHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>登录提醒</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #2196F3, #1976D2); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .login-info { background-color: #e3f2fd; padding: 20px; border-left: 4px solid #2196F3; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #2196F3; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 登录提醒</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="login-info">
                <p>您的账户于 <strong>{{.LoginTime}}</strong> 在 <strong>{{.Device}}</strong> 设备登录。</p>
                <p>如非本人操作，请及时修改密码。</p>
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

// getLoginAlertTextTemplate 获取登录提醒纯文本模板
func getLoginAlertTextTemplate() string {
	return `登录提醒

亲爱的 {{.UserName}} 同学，您好！

您的账户于 {{.LoginTime}} 在 {{.Device}} 设备登录。

如非本人操作，请及时修改密码。

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getCourseReminderHTMLTemplate 获取课程提醒HTML模板
func getCourseReminderHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>课程提醒</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #9C27B0, #7B1FA2); color: white; padding: 30px 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; font-weight: normal; }
        .content { padding: 30px 20px; }
        .greeting { font-size: 16px; margin-bottom: 20px; }
        .course-info { background-color: #f3e5f5; padding: 20px; border-left: 4px solid #9C27B0; border-radius: 4px; margin: 20px 0; }
        .highlight { color: #9C27B0; font-weight: bold; }
        .footer { background-color: #f8f8f8; padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📚 课程提醒</h1>
        </div>
        <div class="content">
            <div class="greeting">
                亲爱的 <span class="highlight">{{.UserName}}</span> 同学，您好！
            </div>
            <div class="course-info">
                <p><strong>课程时间：</strong><span class="highlight">{{.CourseTime}}</span></p>
                <p><strong>课程名称：</strong>{{.CourseName}}</p>
                <p><strong>上课地点：</strong>{{.Location}}</p>
                <p>请准时参加课程。</p>
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

// getCourseReminderTextTemplate 获取课程提醒纯文本模板
func getCourseReminderTextTemplate() string {
	return `课程提醒

亲爱的 {{.UserName}} 同学，您好！

课程时间：{{.CourseTime}}
课程名称：{{.CourseName}}
上课地点：{{.Location}}

请准时参加课程。

此邮件由 {{.SystemName}} 自动发送，请勿回复。
发送时间：{{.SendTime}}`
}

// getExamNoticeHTMLTemplate 获取考试通知HTML模板
