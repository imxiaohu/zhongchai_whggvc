{{ if .Versions -}}
# Changelog

所有重要的项目更新都将记录在此文件中。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。

格式说明：提交信息请遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范。

```
类型(范围): 描述
# 示例
feat(community): 新增用户主页功能
fix(login): 修复登录超时问题
docs: 更新接口文档
```

{{ range .Versions -}}
<a name="{{ .Tag.Name }}"></a>
## [{{ .Tag.Name }}]{{ if .Tag.Date }} - {{ datetime "2006年1月2日" .Tag.Date }}{{ end }}

{{ range .CommitGroups -}}
### {{ .Title }}

{{ range .Commits -}}
{{ if .Scope }}**{{ .Scope }}**: {{ end }}{{ .Subject }}
{{ end }}
{{ end -}}

{{- if .NoteGroups -}}
{{ range .NoteGroups -}}
### {{ .Title }}

{{ range .Notes }}
{{ .String }}
{{ end }}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}

---

*以下为历史手动维护版本记录，保留以供参考。*

---END---
