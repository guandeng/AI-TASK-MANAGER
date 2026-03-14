/**
 * 长耗时操作进度服务
 * 提供全局 LoadingBar + 阶段性消息提示
 */

export interface LoadingStep {
  message: string;
  duration?: number; // 预计耗时(毫秒)，用于进度估算
}

export interface LoadingOptions {
  /** 操作名称 */
  title?: string;
  /** 操作步骤 */
  steps: LoadingStep[];
  /** 完成后的消息 */
  successMessage?: string;
  /** 失败后的消息 */
  errorMessage?: string;
}

/**
 * 长耗时操作进度管理器
 */
export class LoadingService {
  private currentStep = 0;
  private options: LoadingOptions | null = null;
  private messageInstance: ReturnType<typeof window.$message.loading> | null = null;
  private startTime = 0;
  private isFinished = false;

  /**
   * 开始一个长耗时操作
   */
  start(options: LoadingOptions) {
    this.options = options;
    this.currentStep = 0;
    this.isFinished = false;
    this.startTime = Date.now();

    // 启动顶部进度条
    window.$loadingBar?.start();

    // 显示第一条消息
    if (options.steps.length > 0) {
      this.updateMessage(options.steps[0].message);
    }
  }

  /**
   * 进入下一步
   */
  nextStep() {
    if (!this.options || this.isFinished) return;

    this.currentStep++;

    if (this.currentStep < this.options.steps.length) {
      const step = this.options.steps[this.currentStep];
      this.updateMessage(step.message);

      // 更新进度条（根据步骤估算）
      const progress = (this.currentStep + 1) / this.options.steps.length;
      // loadingBar 没有直接的进度设置，但我们可以在消息中显示
    }
  }

  /**
   * 更新当前步骤的消息
   */
  updateMessage(message: string) {
    // 销毁之前的消息
    if (this.messageInstance) {
      this.messageInstance.destroy();
    }

    // 显示新的 loading 消息
    const stepInfo = this.options && this.options.steps.length > 1
      ? ` (${this.currentStep + 1}/${this.options.steps.length})`
      : '';

    this.messageInstance = window.$message?.loading(`${message}${stepInfo}`, {
      duration: 0, // 不自动关闭
      closable: false
    });
  }

  /**
   * 手动更新进度消息（用于更灵活的场景）
   */
  update(message: string) {
    this.updateMessage(message);
  }

  /**
   * 操作成功完成
   */
  success(message?: string) {
    if (this.isFinished) return;
    this.isFinished = true;

    // 销毁 loading 消息
    if (this.messageInstance) {
      this.messageInstance.destroy();
      this.messageInstance = null;
    }

    // 完成进度条
    window.$loadingBar?.finish();

    // 显示成功消息
    const elapsed = this.startTime ? ((Date.now() - this.startTime) / 1000).toFixed(1) : '';
    const successMsg = message || this.options?.successMessage || '操作完成';
    window.$message?.success(`${successMsg} (耗时 ${elapsed}s)`);
  }

  /**
   * 操作失败
   */
  error(message?: string) {
    if (this.isFinished) return;
    this.isFinished = true;

    // 销毁 loading 消息
    if (this.messageInstance) {
      this.messageInstance.destroy();
      this.messageInstance = null;
    }

    // 进度条显示错误状态
    window.$loadingBar?.error();

    // 显示错误消息
    window.$message?.error(message || this.options?.errorMessage || '操作失败');
  }

  /**
   * 取消操作
   */
  cancel() {
    if (this.isFinished) return;
    this.isFinished = true;

    if (this.messageInstance) {
      this.messageInstance.destroy();
      this.messageInstance = null;
    }

    window.$loadingBar?.finish();
    window.$message?.warning('操作已取消');
  }
}

/**
 * 预定义的加载步骤配置
 */
export const LOADING_PRESETS = {
  /** 拆分子任务 */
  expandTask: {
    title: '拆分子任务',
    steps: [
      { message: '正在分析任务内容...' },
      { message: '正在调用 AI 生成子任务...' },
      { message: '正在保存子任务...' }
    ],
    successMessage: '子任务拆分成功'
  },

  /** 重写子任务 */
  regenerateSubtask: {
    title: '重写子任务',
    steps: [
      { message: '正在分析原子任务...' },
      { message: '正在调用 AI 重新生成...' },
      { message: '正在保存...' }
    ],
    successMessage: '子任务重写成功'
  },

  /** 需求拆分任务 */
  splitRequirement: {
    title: '拆分需求为任务',
    steps: [
      { message: '正在解析需求文档...' },
      { message: '正在调用 AI 分析需求...' },
      { message: '正在生成任务列表...' },
      { message: '正在保存任务...' }
    ],
    successMessage: '需求拆分成功'
  },

  /** 批量删除 */
  batchDelete: {
    title: '批量删除',
    steps: [
      { message: '正在删除...' }
    ],
    successMessage: '批量删除成功'
  },

  /** 复制任务 */
  copyTask: {
    title: '复制任务',
    steps: [
      { message: '正在复制任务...' },
      { message: '正在复制子任务...' }
    ],
    successMessage: '任务复制成功'
  }
} as const;

/**
 * 创建一个新的 LoadingService 实例
 */
export function createLoadingService(): LoadingService {
  return new LoadingService();
}

/**
 * 快速执行带进度提示的异步操作
 */
export async function withLoading<T>(
  options: LoadingOptions,
  task: (loading: LoadingService) => Promise<T>
): Promise<T> {
  const loading = createLoadingService();
  loading.start(options);

  try {
    const result = await task(loading);
    loading.success();
    return result;
  } catch (error) {
    loading.error(error instanceof Error ? error.message : '操作失败');
    throw error;
  }
}
