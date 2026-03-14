import { request } from '@/service/request';
import type { MenuItem, MenuFormData, MenuListResponse, MenuOperationResponse } from '@/typings/api/menu';

const API_BASE = '/api';

/** 获取菜单列表 */
export function fetchMenuList() {
  return request<MenuListResponse>({ url: `${API_BASE}/menus`, method: 'GET' });
}

/** 获取菜单树 */
export function fetchMenuTree() {
  return request<{ menus: MenuItem[] }>({ url: `${API_BASE}/menus/tree`, method: 'GET' });
}

/** 获取单个菜单 */
export function fetchMenu(key: string) {
  return request<{ menu: MenuItem }>({ url: `${API_BASE}/menus/${key}`, method: 'GET' });
}

/** 创建菜单 */
export function createMenu(data: MenuFormData) {
  return request<MenuOperationResponse>({ url: `${API_BASE}/menus`, method: 'POST', data });
}

/** 更新菜单 */
export function updateMenu(key: string, data: Partial<MenuFormData>) {
  return request<MenuOperationResponse>({ url: `${API_BASE}/menus/${key}/update`, method: 'POST', data });
}

/** 删除菜单 */
export function deleteMenu(key: string) {
  return request<MenuOperationResponse>({ url: `${API_BASE}/menus/${key}/delete`, method: 'POST' });
}

/** 批量删除菜单 */
export function batchDeleteMenus(keys: string[]) {
  return request<{ success: boolean; deletedKeys: string[] }>({
    url: `${API_BASE}/menus/batch-delete`,
    method: 'POST',
    data: { keys }
  });
}

/** 更新菜单排序 */
export function reorderMenus(data: { key: string; order: number }[]) {
  return request<{ success: boolean }>({ url: `${API_BASE}/menus/reorder`, method: 'POST', data });
}

/** 移动菜单 */
export function moveMenu(key: string, targetParentKey: string | null) {
  return request<MenuOperationResponse>({
    url: `${API_BASE}/menus/${key}/move`,
    method: 'POST',
    data: { targetParentKey }
  });
}

/** 切换菜单启用状态 */
export function toggleMenuEnabled(key: string, enabled: boolean) {
  return request<MenuOperationResponse>({
    url: `${API_BASE}/menus/${key}/toggle`,
    method: 'POST',
    data: { enabled }
  });
}
