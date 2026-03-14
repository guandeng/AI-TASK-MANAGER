import { request } from '@/service/request';

const API_BASE = '/api';

// 配置类型定义
export interface AIProviderConfig {
  enabled: boolean;
  apiKey: string;
  model: string;
  baseUrl?: string;
}

export interface AIConfig {
  provider: string;
  providers: {
    qwen: AIProviderConfig;
    gemini: AIProviderConfig;
    perplexity: AIProviderConfig;
  };
  parameters: {
    maxTokens: number;
    temperature: number;
  };
}

export interface GeneralConfig {
  debug: boolean;
  logLevel: string;
  defaultSubtasks: number;
  defaultPriority: string;
  projectName: string;
  useChinese: boolean;
}

export interface DatabaseConfig {
  connection: string;
  host: string;
  hostSlave?: string;
  port: number;
  database: string;
  username: string;
  password: string;
  charset: string;
  collation: string;
}

export interface StorageConfig {
  type: string;
  database: DatabaseConfig;
}

export interface AppConfig {
  version: string;
  ai: AIConfig;
  general: GeneralConfig;
  storage: StorageConfig;
}

export interface ConfigResponse {
  config: AppConfig;
  configPath: string | null;
}

// 获取当前配置
export function fetchConfig() {
  return request<ConfigResponse>({ url: `${API_BASE}/config`, method: 'GET' });
}

// 更新配置
export function updateConfig(data: Partial<AppConfig>) {
  return request({ url: `${API_BASE}/config/update`, method: 'POST', data });
}

// 切换 AI 提供商
export function switchAIProvider(provider: string) {
  return request({
    url: `${API_BASE}/config/ai-provider`,
    method: 'POST',
    data: { provider }
  });
}

// 更新提供商配置
export function updateProviderConfig(provider: string, data: Partial<AIProviderConfig>) {
  return request({
    url: `${API_BASE}/config/ai-provider/${provider}`,
    method: 'POST',
    data
  });
}

// 重置配置
export function resetConfig() {
  return request({ url: `${API_BASE}/config/reset`, method: 'POST' });
}
