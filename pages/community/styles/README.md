# 社区页面样式优化文档

## 概述

本次优化对社区页面的样式进行了全面改进，提升了用户体验、视觉一致性和代码可维护性。

## 主要改进

### 1. 视觉层次优化

- **增强的卡片设计**: 使用更现代的圆角、阴影和边框
- **改进的颜色对比**: 优化文字和背景的对比度
- **统一的间距系统**: 使用标准化的间距变量
- **渐变效果**: 为按钮和图标添加精美的渐变背景

### 2. 交互体验提升

- **流畅的动画**: 使用 cubic-bezier 缓动函数
- **触觉反馈**: 点击时的缩放和位移效果
- **悬停状态**: 鼠标悬停时的微妙变化
- **加载状态**: 统一的加载动画

### 3. 响应式设计

- **移动优先**: 针对移动设备优化
- **自适应布局**: 支持不同屏幕尺寸
- **触摸友好**: 合适的触摸目标大小

### 4. 主题系统

- **深色模式支持**: 完整的深色主题
- **微信小程序适配**: 专门的小程序主题
- **动态切换**: 平滑的主题切换动画

## 文件结构

```
pages/community/styles/
├── theme.scss      # 主题配置和变量
├── common.scss     # 通用样式和混入
└── README.md       # 文档说明
```

## 样式变量

### 颜色系统

```scss
// 主色调
--community-primary: #6366f1;
--community-secondary: #8b5cf6;
--community-accent: #ec4899;

// 状态颜色
--status-online: #10b981;
--status-offline: #6b7280;
--status-busy: #f59e0b;
--status-away: #ef4444;

// 内容类型
--type-article: #3b82f6;
--type-announcement: #f59e0b;
--type-activity: #10b981;
--type-discussion: #8b5cf6;
```

### 间距系统

```scss
--space-xs: 8rpx;
--space-sm: 12rpx;
--space-md: 16rpx;
--space-lg: 24rpx;
--space-xl: 32rpx;
--space-2xl: 48rpx;
--space-3xl: 64rpx;
```

### 阴影系统

```scss
--shadow-sm: 0 2rpx 8rpx rgba(0, 0, 0, 0.04);
--shadow-md: 0 4rpx 16rpx rgba(0, 0, 0, 0.08);
--shadow-lg: 0 8rpx 24rpx rgba(0, 0, 0, 0.12);
--shadow-xl: 0 12rpx 32rpx rgba(0, 0, 0, 0.16);
```

## 混入 (Mixins)

### 卡片样式

```scss
@include card-base;
```

应用统一的卡片样式，包括背景、边框、阴影和交互效果。

### 按钮样式

```scss
@include button-primary;  // 主要按钮
@include button-secondary; // 次要按钮
```

### 头像样式

```scss
@include avatar(64rpx); // 指定大小的头像
```

### 标签样式

```scss
@include tag-primary;   // 主要标签
@include tag-success;   // 成功标签
@include tag-warning;   // 警告标签
```

## 动画效果

### 内置动画

- `fadeInUp`: 从下方淡入
- `slideInRight`: 从右侧滑入
- `pulse`: 脉冲效果

### 使用方法

```scss
.element {
  animation: fadeInUp 0.3s ease-out;
}

// 或使用工具类
<view class="fade-in-up">内容</view>
```

## 响应式断点

```scss
@include mobile {
  // 移动设备样式
}

@include tablet {
  // 平板设备样式
}

@include desktop {
  // 桌面设备样式
}
```

## 最佳实践

### 1. 使用语义化变量

```scss
// 好的做法
color: var(--text-primary);
padding: var(--space-lg);

// 避免硬编码
color: #1e293b;
padding: 24rpx;
```

### 2. 利用混入减少重复

```scss
// 好的做法
.card {
  @include card-base;
}

// 避免重复代码
.card {
  background: var(--bg-secondary);
  border-radius: 20rpx;
  // ... 大量重复样式
}
```

### 3. 保持一致的命名

- 使用 BEM 命名规范
- 保持组件内样式的一致性
- 使用有意义的类名

### 4. 性能优化

- 使用 CSS 变量而非 JavaScript 动态样式
- 避免过度嵌套选择器
- 合理使用 transform 而非改变布局属性

## 兼容性

- **uni-app**: 完全兼容
- **微信小程序**: 专门适配
- **H5**: 现代浏览器支持
- **App**: 原生渲染支持

## 维护指南

### 添加新颜色

1. 在 `theme.scss` 中定义变量
2. 考虑深色模式的适配
3. 更新相关混入

### 添加新组件样式

1. 使用现有混入
2. 遵循设计系统
3. 考虑响应式需求

### 主题定制

1. 修改 CSS 变量值
2. 保持颜色对比度
3. 测试深色模式

## 更新日志

### v1.0.0 (2024-06-16)

- 初始版本发布
- 完整的主题系统
- 通用样式混入
- 响应式设计支持
- 深色模式适配

## 贡献指南

1. 遵循现有的代码风格
2. 添加适当的注释
3. 考虑性能影响
4. 测试多种设备和主题

## 技术支持

如有问题或建议，请联系开发团队或提交 Issue。
