import type { AxiosResponse } from 'axios';
import { BACKEND_ERROR_CODE, createFlatRequest } from '@sa/axios';
import { getServiceBaseURL } from '@/utils/service';

const isHttpProxy = import.meta.env.DEV && import.meta.env.VITE_HTTP_PROXY === 'Y';
const { baseURL } = getServiceBaseURL(import.meta.env, isHttpProxy);

/**
 * 任务管理器 API 请求
 * 返回 { data, error } 格式，便于错误处理
 */
export const request = createFlatRequest<any, any, Record<string, unknown>>(
  {
    baseURL,
    timeout: 60000 // 60 秒超时
  },
  {
    transform(response: AxiosResponse) {
      // 直接返回响应数据
      return response.data;
    },
    async onRequest(config) {
      // 不需要认证
      return config;
    },
    isBackendSuccess(response) {
      // 检查后端返回的 code 字段，0 表示成功
      const data = response as any;
      // 如果有 code 字段，检查是否为 0
      if (data && typeof data.code !== 'undefined') {
        return data.code === 0;
      }
      // 如果没有 code 字段，默认成功
      return true;
    },
    async onBackendFail(_response) {
      // 不在这里处理错误，让调用方处理
    },
    onError(error) {
      // 显示错误信息
      let message = error.message;

      if (error.code === BACKEND_ERROR_CODE) {
        message = error.response?.data?.error || message;
      }

      window.$message?.error(message);
    }
  }
);
