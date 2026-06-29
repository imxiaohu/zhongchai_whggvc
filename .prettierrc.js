/**
 * Prettier 代码格式化配置
 * 与 ESLint 配合使用，ESLint 处理代码质量，Prettier 处理代码风格
 */

module.exports = {
	// 行宽
	printWidth: 100,

	// 使用单引号
	singleQuote: true,

	// 不使用分号
	semi: false,

	// 对象括号内不需要空格
	bracketSpacing: true,

	// 箭头函数始终包含括号
	arrowParens: 'always',

	// 不使用尾随逗号
	trailingComma: 'none',

	// 使用空格缩进（与 ESLint 保持一致）
	useTabs: true,

	// Tab 宽度
	tabWidth: 2,

	// 不自动格式化 markdown
	proseWrap: 'preserve',

	// 包裹多行 HTML/JSX/TSX 元素
	htmlWhitespaceSensitivity: 'css',

	// Vue 文件脚本和样式中的缩进
	vueIndentScriptAndStyle: false,

	// 组件的 props 排序
	componentPropsSort: true,

	// 末尾自动格式化
	endOfLine: 'lf',

	// 范围格式化的文件
	overrides: [
		{
			files: '*.json',
			options: {
				printWidth: 120,
				semi: false,
				singleQuote: false
			}
		},
		{
			files: '*.md',
			options: {
				proseWrap: 'preserve'
			}
		}
	]
}
