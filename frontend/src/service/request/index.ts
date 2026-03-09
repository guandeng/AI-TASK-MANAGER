import type { AxiosResponse } from 'axios';
import { BACKEND_ERROR_CODE, createFlatRequest } from '@sa/axios';
import { getServiceBaseURL } from '@/utils/service';

const isHttpProxy = import.meta.env.DEV && import.meta.env.VITE_HTTP_PROXY === 'Y';
const { baseURL } = getServiceBaseURL(import.meta.env, isHttpProxy);

/**
 * 任务管理器 API 请求
 * 返回 { data, error } 格式，便于错误处理
 */
export const request = createFlatRequest<
  any,
  any,
  Record<string, unknown>
>(
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
    isBackendSuccess() {
      // 任务管理器 API 不使用标准错误码，直接返回成功
      return true;
    },
    async onBackendFail() {
      // 不处理
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
