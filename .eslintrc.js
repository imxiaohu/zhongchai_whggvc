module.exports = {
  root: true,
  env: {
    browser: true,
    node: true,
    es2022: true,
  },
  extends: [
    'eslint:recommended',
  ],
  parserOptions: {
    ecmaVersion: 2022,
    sourceType: 'module',
  },
  globals: {
    uni: 'readonly',
    wx: 'readonly',
    plus: 'readonly',
    getApp: 'readonly',
    defineAsyncComponent: 'readonly',
    uniCloud: 'readonly',
    plusAndroid: 'readonly',
    plusIOS: 'readonly',
  },
  ignorePatterns: [
    'node_modules/',
    'dist/',
    'dist-ssr/',
    'unpackage/',
    '*.min.js',
    '*.config.js',
    'vite.config.*',
    'uni_modules/',
  ],
  rules: {
    'no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
    'no-console': ['warn', { allow: ['warn', 'error'] }],
    'no-debugger': 'error',
    'no-undef': 'off',
  },
};
