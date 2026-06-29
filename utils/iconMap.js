/**
 * TDesign 图标映射工具
 * 将 MDI 图标名称和 uni-icons 图标名称映射到 TDesign 图标名称
 *
 * 注意: TDesign 图标使用 CSS font-face 方式，通过 class="t-icon-{name}" 渲染
 *       组件使用 <t-icon name="xxx"> 或 <TIcon name="xxx">
 *
 * 图标库在线预览: https://tdesign.tencent.com/icons
 */

// MDI 图标名称 → TDesign 图标名称
const mdiToTDesignMap = {
  // === 导航与操作 ===
  'magnify': 'search',
  'chevron-left': 'chevron-left',
  'chevron-right': 'chevron-right',
  'chevron-down': 'chevron-down',
  'chevron-up': 'chevron-up',
  'arrow-up-bold-circle': 'arrow-up-circle-filled',
  'dots-horizontal': 'more',
  'more-vertical': 'more',
  'close': 'close',
  'plus': 'add',
  'minus': 'minus',

  // === 用户与账户 ===
  'account': 'user',
  'account-outline': 'user',
  'account-circle-outline': 'user-circle-filled',
  'account-group': 'usergroup',
  'account-group-outline': 'usergroup',
  'account-tie': 'user',
  'account-voice': 'user',
  'account-cog-outline': 'user-setting',
  'login': 'login',
  'logout': 'logout',

  // === 教育相关 ===
  'school': 'education',
  'school-outline': 'education',
  'book-open-variant': 'book-open',

  // === 文档与内容 ===
  'file-document': 'book',
  'file-document-edit': 'book-open',
  'text-box-outline': 'book',

  // === 编辑 ===
  'pencil': 'edit',
  'pencil-outline': 'edit',
  'square-edit-outline': 'edit',
  'pencil-plus': 'edit-1',

  // === 媒体 ===
  'camera': 'photo',
  'image': 'image',

  // === 可视化 ===
  'eye': 'browse',
  'eye-outline': 'browse-filled',
  'eye-off-outline': 'browse-off',

  // === 状态与反馈 ===
  'check': 'check',
  'check-circle': 'check-circle-filled',
  'check-circle-outline': 'check-circle',
  'check-all': 'check-double',
  'check-decagram': 'check-circle-filled',
  'alert-circle': 'error-circle-filled',
  'alert-circle-outline': 'error-circle',
  'alert-outline': 'info-circle',
  'information': 'info-circle-filled',
  'information-outline': 'info-circle',

  // === 收藏与书签 ===
  'star': 'star',
  'star-outline': 'star',
  'star-filled': 'star-filled',
  'star-circle': 'star-filled',
  'bookmark': 'bookmark',
  'bookmark-outline': 'bookmark',
  'bookmark-off-outline': 'bookmark-minus',
  'heart': 'heart',
  'heart-outline': 'heart',
  'heart-filled': 'heart-filled',

  // === 时间与日期 ===
  'calendar': 'calendar',
  'calendar-check': 'calendar-filled',
  'calendar-outline': 'calendar',
  'calendar-clock': 'calendar-2',
  'clock-outline': 'time',
  'history': 'history',

  // === 设置与工具 ===
  'cog': 'setting',
  'cog-outline': 'setting',
  'wrench-outline': 'tools',
  'trash-can-outline': 'delete',
  'refresh': 'refresh',
  'sync': 'refresh',

  // === 网络与通信 ===
  'bell': 'notification',
  'bell-outline': 'notification',
  'bell-off-outline': 'notification-circle',
  'chat': 'chat',
  'chat-processing': 'chat-double',
  'chat-text': 'chat-message',
  'comment-outline': 'chat',
  'comment-text': 'chat-message',
  'email': 'mail',
  'email-outline': 'mail',
  'phone-outline': 'call',
  'send': 'send',
  'send-outline': 'send',
  'message-text': 'mail',
  'message-text-outline': 'mail',
  'wechat': 'logo-wechat-stroke',
  'link': 'link',
  'link-variant': 'link',

  // === 位置 ===
  'map-marker': 'location',
  'map-marker-radius': 'location',

  // === 教育指标 ===
  'trending-up': 'trending-up',
  'shield-check': 'lock-on',
  'shield-check-outline': 'lock-on',
  'shield-lock-outline': 'lock-on',
  'shield-outline': 'lock-on',

  // === 其他 ===
  'lock': 'lock-on',
  'lock-outline': 'lock-on',
  'bullhorn-variant': 'sound',
  'run': 'activity',
  'lifebuoy': 'help-circle',
  'identifier': 'user-search',
  'wallet-plus-outline': 'wallet',
  'send-check-outline': 'send',
  'chart-box-outline': 'chart-bar',
  'clipboard-text-outline': 'view-list',
  'clipboard-off-outline': 'view-list-off',
  'flask-outline': 'check-1',
  'format-list-bulleted': 'view-list',
  'content-save-outline': 'save',
  'thumb-up': 'thumb-up',
  'magnify-expand': 'search',
  'transmission-tower': 'wifi',
  'server': 'server',
  'inbox-outline': 'folder',
  'cloud-download-outline': 'cloud-download',
  'pin': 'pin',
};

// uni-icons 图标名称 → TDesign 图标名称
const uniToTDesignMap = {
  'search': 'search',
  'close': 'close',
  'clear': 'close',
  'eye': 'browse',
  'eye-filled': 'browse-filled',
  'eye-slash': 'browse-off',
  'eye-slash-filled': 'browse-off',
  'camera': 'photo',
  'image': 'image',
  'star': 'star',
  'star-filled': 'star-filled',
  'person': 'user',
  'person-add': 'user-add',
  'gear': 'setting',
  'settings-filled': 'setting-filled',
  'chat': 'chat',
  'chatboxes': 'chat',
  'chatboxes-filled': 'chat-filled',
  'fire': 'local',
  'fire-filled': 'local-filled',
  'trash': 'delete',
  'trash-filled': 'delete-filled',
  'up': 'chevron-up',
  'down': 'chevron-down',
  'top': 'jump',
  'bottom': 'jump-down',
  'left': 'chevron-left',
  'right': 'chevron-right',
  'arrow-right': 'arrow-right',
  'arrow-up': 'arrow-up',
  'arrow-down': 'arrow-down',
  'arrowthinleft': 'chevron-left',
  'arrowthinup': 'chevron-up',
  'arrowthindown': 'chevron-down',
  'arrowthinright': 'chevron-right',
  'arrowdown': 'arrow-down',
  'arrowup': 'arrow-up',
  'plus': 'add',
  'plusempty': 'add',
  'plus-filled': 'add-filled',
  'minus-filled': 'minus-filled',
  'more-filled': 'more',
  'spinner-cycle': 'refresh',
  'calendar': 'calendar',
  'calendar-filled': 'calendar-filled',
  'info': 'info-circle',
  'info-filled': 'info-circle-filled',
  'reload': 'refresh',
  'refreshempty': 'refresh',
  'locked': 'lock-on',
  'compose': 'edit',
  'checkbox-filled': 'check-rectangle-filled',
  'circle': 'circle',
  'staff': 'usergroup',
  'staff-filled': 'usergroup-filled',
  'sound': 'sound',
  'paperplane-filled': 'send-filled',
  'copy': 'copy',
  'download': 'download',
  'download-filled': 'download-filled',
  'link': 'link',
  'clock': 'time',
};

/**
 * 将 MDI 图标名映射为 TDesign 图标名
 * @param {string} mdiIconName - MDI 图标名称
 * @returns {string} TDesign 图标名称
 */
export function mapMdiToTDesign(mdiIconName) {
  return mdiToTDesignMap[mdiIconName] || mdiIconName;
}

/**
 * 将 uni-icons 图标名映射为 TDesign 图标名
 * @param {string} uniIconName - uni-icons 图标名称
 * @returns {string} TDesign 图标名称
 */
export function mapUniToTDesign(uniIconName) {
  return uniToTDesignMap[uniIconName] || uniIconName;
}

/**
 * 直接返回 TDesign 图标名（如果已经是 TDesign 格式）
 * @param {string} name - 图标名称
 * @returns {string} TDesign 图标名称
 */
export function toTDesignIcon(name) {
  return name;
}

export default {
  mapMdiToTDesign,
  mapUniToTDesign,
  toTDesignIcon,
};
