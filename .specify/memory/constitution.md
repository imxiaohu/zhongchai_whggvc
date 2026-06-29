# 众柴智慧校园系统 Constitution

<!--
## Sync Impact Report

**Version change**: 0.0.0 → 0.1.0 (MINOR — initial adoption, first full principle set + two new sections)

**Principles defined**: 7 core principles established covering:
- Layered Architecture (Go backend)
- API Abstraction (uni-app frontend)
- Cross-Platform Compatibility
- State Management (Pinia)
- Error Handling
- Security
- Performance & Quality

**Sections added**:
- Security Requirements
- Performance Standards

**Templates checked**: ✅ all templates aligned (no constitution-specific references to update)
- .specify/templates/plan-template.md — Constitution Check section references [Gates determined based on constitution file] placeholder, generic, no update needed
- .specify/templates/spec-template.md — no constitution references, no update needed
- .specify/templates/tasks-template.md — no constitution references, no update needed
- .specify/templates/checklist-template.md — no constitution references, no update needed
- .specify/templates/constitution-template.md — source template, no update needed

**Deferred items**:
- TODO(RATIFICATION_DATE): Original adoption date unknown. Set ratification date once confirmed by project maintainer.
-->

## Core Principles

### I. 分层架构 (Layered Architecture)

Go 后端必须严格遵循分层单向依赖：`handler → service → repository`，禁止反向导入。

- 数据库操作只允许出现在 `repository` 层，禁止在 handler/service 直接写原生 GORM
- 入参 DTO、出参 VO 转换统一在 service 层处理
- 所有 SQL 必须携带 `WithContext(ctx)`
- 统一使用 `pkg/response` 封装返回

**Rationale**: 分层架构是后端可维护性和可测试性的基础。单向依赖防止循环引用，repository 集中管理 SQL 便于审计和优化。

### II. API 抽象层 (API Abstraction Layer)

uni-app 前端禁止在页面中直接调用平台 API（`wx`、`plus`、`my` 等）。所有平台 API 必须收口到 `utils/` 或 `composables/`，对外暴露统一方法名。

- 支付统一封装 `utils/pay.js`
- 扫码统一封装 `utils/scan.js`
- 路由跳转统一封装 `utils/router.js`
- 网络请求统一走 `utils/request.js`，禁止直接调用 `uni.request`
- 缓存操作全部封装 `utils/storage.js`

**Rationale**: uni-app 跨三端（小程序/App/H5），平台 API 差异巨大。抽象层确保业务代码与平台解耦，换端无忧。

### III. 跨端兼容 (Cross-Platform Compatibility)

条件编译是 uni-app 跨端开发的唯一正确方式。

| 端 | 标记 |
|---|---|
| 微信小程序 | `#ifdef MP-WEIXIN` |
| App | `#ifdef APP-PLUS` |
| H5 | `#ifdef H5` |
| 所有小程序 | `#ifdef MP` |

- 端特有 API 必须抽离到 `utils/` 或 `composables/`，禁止直接写在页面
- 条件编译禁止嵌套超过 2 层，超过必须抽离独立文件
- 静态资源分端存放：`static/mp-weixin/`、`static/app/`、`static/h5/`

**禁止行为**：
- 使用 DOM 标签（`div`/`span`/`img`/`a`），统一使用 uni 规范标签（`view`/`text`/`image`/`navigator`）
- 样式穿透在小程序端使用 `:deep()`（必须用 `::v-deep`）

**Rationale**: 条件编译是 uni-app 的核心能力，在编译期消除平台差异，避免运行时兜满判断。

### IV. 状态管理 (State Management with Pinia)

Pinia store 是前端状态管理的唯一规范，禁止滥用 `globalData` 或在页面中直接操作。

- 按业务分仓库，禁止全部塞一个仓库
- 禁止组件内直接修改 store 原始变量，统一走 action
- 持久化仅在仓库内部处理，页面不操作 storage
- 敏感数据（token、用户信息）统一封装 storage 工具，App 优先使用原生加密存储
- 所有修改 state 的逻辑统一写在 actions

**Rationale**: Pinia Composition API 写法使状态逻辑可测试、可复用。集中管理避免状态散落导致难以追踪。

### V. 错误处理 (Error Handling)

每个 `async` 函数必须有 `try/catch/finally`，禁止丢弃关键错误。

- 通用编码规范：所有异步请求加 loading 状态和错误处理
- Go 后端：使用 `fmt.Errorf("操作描述: %w", err)` 包装错误，禁止直接抛原始数据库错误给前端
- uni-app 前端：`uni.showToast` 统一提示，`console.error` 记录详细日志
- 生产环境：移除所有 `console.log`

**禁止行为**：
- 用 `_` 丢弃关键错误（数据库、Redis、IO）
- 在模板内写复杂运算，逻辑抽成 computed 或 method

**Rationale**: 错误处理缺失是生产事故的主要来源。统一模式使问题可追踪、可复现。

### VI. 安全规范 (Security)

**前端安全**：
- API 域名、AppKey、Secret 禁止硬编码，统一放入 `.env` 文件
- 敏感信息在 App 端优先使用 `plus.nativeUI.encryptedStorage` 原生加密
- 小程序/H5 端手动 AES 加密敏感信息
- 小程序 `manifest.json` 必须填写 `mp-weixin.requiredPrivateInfos`

**后端安全**：
- 密码使用 bcrypt 加密，禁止明文存储
- JWT 设置过期时间，有刷新机制
- 生产关闭 Gin Debug 模式
- 日志禁止打印密码、token、密钥

**Rationale**: 安全问题一旦上线极难修复，必须从编码层面强制约束。

### VII. 性能与质量 (Performance & Quality)

**前端性能约束**：
- 长列表必须分页 + 虚拟列表，禁止一次性渲染 >1000 节点
- 图片全局懒加载 `lazy-load`，必须设置宽高
- 频繁切换 class 触发 setData 卡顿，优先写死 class
- App 端避免无限制新开 webview 窗口，高频页面复用

**后端性能约束**：
- 函数长度控制在 80 行以内，函数参数不超过 5 个
- 提前 return 减少嵌套深度，魔法数字全部提取为常量
- 禁止全局变量，仅单例客户端（DB、Redis）可用

**样式约束**：
- 布局单位统一使用 `rpx`，`px` 仅限原生组件（map/video/camera/canvas）
- 1px 细线统一用 `1rpx`
- 选择器嵌套 ≤3 层，禁止通配选择器 `*{ }`，禁止 `body` 选择器

**Rationale**: 性能问题在低端设备和弱网环境下被放大。前端 1000 节点渲染在小程序会直接导致白屏，后端长函数难以维护和测试。

## Security Requirements

1. 禁止在代码中硬编码任何密钥、Token、证书
2. `.env` 文件必须加入 `.gitignore`，禁止提交到仓库
3. GitHub Actions Secrets 用于 CI/CD 环境变量，禁止在工作流中明文暴露
4. 生产服务必须关闭 Debug 模式
5. 第三方教务系统代理凭证必须加密存储，禁止日志打印
6. 小程序隐私权限申请前必须弹窗告知用途，禁止静默申请

## Performance Standards

1. 页面数据刷新必须放在 `onShow`，定时器、监听必须在 `onUnload` 销毁
2. App 端 tab 页面切换不触发 `onUnload`，需额外监听 `onHide` 清理资源
3. H5 关键数据必须持久化到 localStorage 兜底
4. 大对象参数禁止拼 url 传递，统一存入 Pinia 临时 store 中转
5. 小程序 URL 参数过长会被截断，大数据必须 store 中转
6. 小程序 CSS 不支持本地背景图，改用 `<image>` 标签
7. Go 后端复杂查询必须设置 `context.WithTimeout`

## Development Workflow

- **Commit 规范**：遵循 Angular Convention `type(scope): description`
  - `feat`：新功能或新页面
  - `fix`：Bug 修复
  - `refactor`：代码重构（无行为变化）
  - `style`：格式化和空白调整
  - `docs`：文档或注释
  - `chore`：依赖和 CI/CD 相关
- **分支管理**：`main`（线上稳定）→ `dev`（开发测试）→ `feat/xxx` / `fix/xxx`（特性/bug 分支）
- **PR 规范**：所有 PR 必须验证合规性，复杂度必须说明理由
- **CI 流水线**：代码校验（ESLint/golangci-lint）→ 测试 → 多平台编译

## Governance

本宪章优先于其他所有开发实践。

** Amendment Procedure（修正程序）**：
1. 修正案必须由项目维护者评审通过
2. 重大变更（移除/重新定义原则，或添加新的强制规则）需附带迁移计划
3. 修正案在合并到主分支后生效

** Versioning Policy（版本策略）**：
- MAJOR：向后不兼容的原则移除或重新定义
- MINOR：新增原则或实质性扩展指导
- PATCH：澄清、用词、笔误修复

** Compliance Review（合规审查）**：
- 所有 PR 和代码评审必须验证对本宪章的合规性
- 使用 `rules/` 目录下的规则文件进行运行时开发指导
- 使用 `skills/` 目录下的技能文件执行特定任务

**运行时指导文件**：
- `.cursor/rules/uni-app-triple-platform.mdc` — 三端适配强制规范
- `.cursor/rules/uni-app-development-standards.mdc` — Vue3 + Pinia + uni-app JS 通用开发规范
- `.cursor/rules/core-standards.mdc` — 核心编码标准
- `.cursor/rules/uni-app-app-entry.mdc` — App.vue / 主入口规范
- `.cursor/rules/uni-app-js-api-basics.mdc` — JS API 基础规范
- `.cursor/rules/uni-app-env-testing.mdc` — 开发/生产环境判断与自动化测试规范
- `.cursor/skills/go-gin-backend/SKILL.md` — Go + Gin 后端开发规范
- `.cursor/skills/uniapp-js-dev/SKILL.md` — uni-app JS 开发规范
- `.cursor/skills/uniapp-cli-commands/SKILL.md` — uni-app CLI 脚手架命令
- `.cursor/skills/pages-json/SKILL.md` — pages.json 页面路由配置
- `.cursor/skills/improve-skills-rules-style/SKILL.md` — Skills 和 Rules 页面样式规范

**Version**: 0.1.0 | **Ratified**: TODO(RATIFICATION_DATE) | **Last Amended**: 2026-06-29
