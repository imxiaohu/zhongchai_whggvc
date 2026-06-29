# 社区模块布局缺陷记录 (Layout Bugs)

| 缺陷 ID | 路径 | 现象描述 | 复现步骤 | 期望表现 | 修复状态 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| BUG-001 | `pages/community/components/PostList.vue` | 作者名称过长时，官方标识 (Official Badge) 会被挤出屏幕或发生错位。 | 1. 模拟一个名称超过 10 个汉字的作者。<br>2. 查看帖子列表。 | 作者名称应截断并显示省略号，官方标识应紧随其后。 | 已修复 |
| BUG-002 | `pages/community/components/ClubContent.vue` | 社团名称过长时，官方标签 (Official Tag) 会被挤出屏幕或导致标题行高度塌陷。 | 1. 模拟一个超长名称的社团。<br>2. 在社团发现列表中查看。 | 社团名称应截断，官方标签应保持在标题行右侧或紧随名称。 | 已修复 |
| BUG-003 | `pages/community/components/RecommendContent.vue` | 活跃社团卡片在宽屏（如 iPad ≥1200px）下显得过窄，留白过多，比例失调。 | 1. 在屏幕宽度 ≥1200px 的设备上打开页面。 | 响应式适配：宽屏下应适当增加卡片宽度或显示更多列。 | 已修复 |
| BUG-004 | `pages/community/index.vue` | 悬浮发布按钮 (Fab Button) 的位置在部分全面屏手机上可能与系统指示条或自定义 TabBar 重叠。 | 1. 在具有不同安全区域高度的模拟器上查看。 | 按钮位置应结合 `env(safe-area-inset-bottom)` 进行动态偏移。 | 已修复 |
| BUG-005 | `pages/community/components/ClubContent.vue` | 筛选栏 (Filter Bar) 在 H5 环境下滚动时，`sticky` 定位可能失效，导致筛选功能被滚出视口。 | 1. 在 H5 环境下向下滚动社团列表。 | 筛选栏应在滚动至顶部时吸顶，方便随时切换分类。 | 已修复 |
| BUG-006 | `pages/community/components/PostList.vue` | 帖子图片网格 (Grid-9) 在极小屏幕 (≤320px) 下，由于间距固定，图片可能显得极其局促甚至溢出。 | 1. 在屏幕宽度 ≤320px 的设备上查看 9 图帖子。 | 响应式适配：极小屏幕下应动态调整 gap 或缩减图片尺寸。 | 已修复 |
| BUG-007 | `pages/community/components/MemberList.vue` | 成员列表中姓名过长时会挤压角色标签，导致错位。 | 1. 模拟一个超长名称的成员。 | 姓名应自动截断并显示省略号。 | 已修复 |
| BUG-008 | `pages/community/post-detail.vue` | 详情页中作者名称和社团标签在窄屏下可能重叠或溢出。 | 1. 在窄屏手机上查看长名称作者和社团。 | 应使用 Flex 布局并设置最大宽度截断。 | 已修复 |
