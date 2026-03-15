import { request } from '@/service/request';

export type LanguageCategory = 'frontend' | 'backend';

export interface Language {
  id: number;
  name: string;
  category: LanguageCategory;
  displayName: string;
  framework: string;
  description: string;
  codeHints: string;
  remark: string;
  isActive: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export interface LanguageCreateRequest {
  name: string;
  category: LanguageCategory;
  displayName: string;
  framework?: string;
  description?: string;
  codeHints?: string;
  remark?: string;
  isActive?: boolean;
  sortOrder?: number;
}

export type LanguageUpdateRequest = Partial<LanguageCreateRequest>;

const API_BASE = '/api';

/**
 * 获取语言列表
 */
export function fetchLanguageList(all: boolean = false, category?: LanguageCategory) {
  const params: Record<string, string> = {};
  if (all) params.all = 'true';
  if (category) params.category = category;

  return request<Language[]>({
    url: `${API_BASE}/languages`,
    method: 'GET',
    params
  });
}

/**
 * 获取语言详情
 */
export function fetchLanguageDetail(id: number) {
  return request<Language>({
    url: `${API_BASE}/languages/${id}`,
    method: 'GET'
  });
}

/**
 * 创建语言
 */
export function createLanguage(data: LanguageCreateRequest) {
  return request<Language>({
    url: `${API_BASE}/languages`,
    method: 'POST',
    data
  });
}

/**
 * 更新语言
 */
export function updateLanguage(id: number, data: LanguageUpdateRequest) {
  return request<Language>({
    url: `${API_BASE}/languages/${id}/update`,
    method: 'POST',
    data
  });
}

/**
 * 删除语言
 */
export function deleteLanguage(id: number) {
  return request<void>({
    url: `${API_BASE}/languages/${id}/delete`,
    method: 'POST'
  });
}
