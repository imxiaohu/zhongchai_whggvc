/**
 * 性能监控工具
 * 用于监控主题切换和语言切换的性能指标
 */

class PerformanceMonitor {
  constructor() {
    this.metrics = {
      themeSwitch: {
        count: 0,
        totalTime: 0,
        averageTime: 0,
        lastSwitchTime: 0,
        maxTime: 0,
        minTime: Infinity
      },
      languageSwitch: {
        count: 0,
        totalTime: 0,
        averageTime: 0,
        lastSwitchTime: 0,
        maxTime: 0,
        minTime: Infinity
      },
      domUpdates: {
        count: 0,
        totalTime: 0,
        averageTime: 0
      }
    };
    
    this.activeTimers = new Map();
    this.isEnabled = true; // 可以通过配置控制是否启用
  }

  /**
   * 开始计时
   * @param {string} operation - 操作类型 ('themeSwitch', 'languageSwitch', 'domUpdate')
   * @param {string} id - 操作ID（可选，用于区分同类型的不同操作）
   */
  startTimer(operation, id = 'default') {
    if (!this.isEnabled) return;
    
    const timerKey = `${operation}_${id}`;
    const startTime = performance.now ? performance.now() : Date.now();
    
    this.activeTimers.set(timerKey, {
      operation,
      startTime,
      id
    });
    
    console.log(`PerformanceMonitor: 开始计时 ${operation} (${id})`);
  }

  /**
   * 结束计时并记录性能指标
   * @param {string} operation - 操作类型
   * @param {string} id - 操作ID
   */
  endTimer(operation, id = 'default') {
    if (!this.isEnabled) return;
    
    const timerKey = `${operation}_${id}`;
    const timer = this.activeTimers.get(timerKey);
    
    if (!timer) {
      console.warn(`PerformanceMonitor: 未找到计时器 ${timerKey}`);
      return;
    }
    
    const endTime = performance.now ? performance.now() : Date.now();
    const duration = endTime - timer.startTime;
    
    this.recordMetric(operation, duration);
    this.activeTimers.delete(timerKey);
    
    console.log(`PerformanceMonitor: ${operation} 耗时 ${duration.toFixed(2)}ms`);
    
    return duration;
  }

  /**
   * 记录性能指标
   * @param {string} operation - 操作类型
   * @param {number} duration - 持续时间（毫秒）
   */
  recordMetric(operation, duration) {
    if (!this.metrics[operation]) {
      console.warn(`PerformanceMonitor: 未知操作类型 ${operation}`);
      return;
    }
    
    const metric = this.metrics[operation];
    
    metric.count++;
    metric.totalTime += duration;
    metric.averageTime = metric.totalTime / metric.count;
    metric.lastSwitchTime = duration;
    metric.maxTime = Math.max(metric.maxTime, duration);
    metric.minTime = Math.min(metric.minTime, duration);
  }

  /**
   * 获取性能报告
   * @param {string} operation - 操作类型（可选，不传则返回所有）
   */
  getReport(operation = null) {
    if (operation) {
      return this.metrics[operation] || null;
    }
    
    return {
      ...this.metrics,
      summary: this.generateSummary()
    };
  }

  /**
   * 生成性能摘要
   */
  generateSummary() {
    const summary = {
      totalOperations: 0,
      averageThemeSwitchTime: 0,
      averageLanguageSwitchTime: 0,
      performanceGrade: 'A'
    };
    
    // 计算总操作数
    Object.values(this.metrics).forEach(metric => {
      summary.totalOperations += metric.count;
    });
    
    // 计算平均时间
    summary.averageThemeSwitchTime = this.metrics.themeSwitch.averageTime;
    summary.averageLanguageSwitchTime = this.metrics.languageSwitch.averageTime;
    
    // 计算性能等级
    const avgThemeTime = summary.averageThemeSwitchTime;
    const avgLangTime = summary.averageLanguageSwitchTime;
    
    if (avgThemeTime < 50 && avgLangTime < 100) {
      summary.performanceGrade = 'A+';
    } else if (avgThemeTime < 100 && avgLangTime < 200) {
      summary.performanceGrade = 'A';
    } else if (avgThemeTime < 200 && avgLangTime < 400) {
      summary.performanceGrade = 'B';
    } else if (avgThemeTime < 500 && avgLangTime < 800) {
      summary.performanceGrade = 'C';
    } else {
      summary.performanceGrade = 'D';
    }
    
    return summary;
  }

  /**
   * 重置所有指标
   */
  reset() {
    this.metrics = {
      themeSwitch: {
        count: 0,
        totalTime: 0,
        averageTime: 0,
        lastSwitchTime: 0,
        maxTime: 0,
        minTime: Infinity
      },
      languageSwitch: {
        count: 0,
        totalTime: 0,
        averageTime: 0,
        lastSwitchTime: 0,
        maxTime: 0,
        minTime: Infinity
      },
      domUpdates: {
        count: 0,
        totalTime: 0,
        averageTime: 0
      }
    };
    
    this.activeTimers.clear();
    console.log('PerformanceMonitor: 已重置所有指标');
  }

  /**
   * 启用/禁用性能监控
   * @param {boolean} enabled - 是否启用
   */
  setEnabled(enabled) {
    this.isEnabled = enabled;
    console.log(`PerformanceMonitor: ${enabled ? '启用' : '禁用'}性能监控`);
  }

  /**
   * 导出性能数据
   */
  exportData() {
    const data = {
      timestamp: new Date().toISOString(),
      metrics: this.metrics,
      summary: this.generateSummary(),
      userAgent: typeof navigator !== 'undefined' ? navigator.userAgent : 'Unknown',
      platform: 'uni-app'
    };
    
    return JSON.stringify(data, null, 2);
  }

  /**
   * 打印性能报告到控制台
   */
  printReport() {
    const report = this.getReport();
    
    console.group('🚀 性能监控报告');
    console.log('📊 主题切换性能:', report.themeSwitch);
    console.log('🌐 语言切换性能:', report.languageSwitch);
    console.log('🔄 DOM更新性能:', report.domUpdates);
    console.log('📈 性能摘要:', report.summary);
    console.groupEnd();
  }

  /**
   * 监控函数执行时间的装饰器
   * @param {string} operation - 操作类型
   * @param {Function} func - 要监控的函数
   * @param {string} id - 操作ID
   */
  monitor(operation, func, id = 'default') {
    return (...args) => {
      this.startTimer(operation, id);
      
      try {
        const result = func.apply(this, args);
        
        // 如果返回Promise，等待完成后结束计时
        if (result && typeof result.then === 'function') {
          return result.finally(() => {
            this.endTimer(operation, id);
          });
        } else {
          this.endTimer(operation, id);
          return result;
        }
      } catch (error) {
        this.endTimer(operation, id);
        throw error;
      }
    };
  }

  /**
   * 获取当前活跃的计时器数量
   */
  getActiveTimersCount() {
    return this.activeTimers.size;
  }

  /**
   * 检查是否有性能问题
   */
  checkPerformanceIssues() {
    const issues = [];
    
    // 检查主题切换性能
    if (this.metrics.themeSwitch.averageTime > 200) {
      issues.push({
        type: 'themeSwitch',
        severity: 'warning',
        message: `主题切换平均耗时 ${this.metrics.themeSwitch.averageTime.toFixed(2)}ms，建议优化`
      });
    }
    
    // 检查语言切换性能
    if (this.metrics.languageSwitch.averageTime > 400) {
      issues.push({
        type: 'languageSwitch',
        severity: 'warning',
        message: `语言切换平均耗时 ${this.metrics.languageSwitch.averageTime.toFixed(2)}ms，建议优化`
      });
    }
    
    // 检查是否有未结束的计时器
    if (this.activeTimers.size > 0) {
      issues.push({
        type: 'memory',
        severity: 'error',
        message: `发现 ${this.activeTimers.size} 个未结束的计时器，可能存在内存泄漏`
      });
    }
    
    return issues;
  }
}

// 创建全局实例
const performanceMonitor = new PerformanceMonitor();

// 在开发环境下自动启用，生产环境下可以通过配置控制
// #ifdef APP-PLUS || H5
performanceMonitor.setEnabled(true);
// #endif

// #ifdef MP-WEIXIN
// 微信小程序环境下默认启用，但可以通过配置关闭
performanceMonitor.setEnabled(true);
// #endif

export default performanceMonitor;
