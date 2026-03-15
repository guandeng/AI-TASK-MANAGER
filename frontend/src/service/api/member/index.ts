import { request } from '@/service/request';
import type {
  Member,
  MemberCreateRequest,
  MemberListParams,
  MemberListResponse,
  MemberStatistics,
  MemberUpdateRequest
} from '@/typings/api/member';

const API_BASE = '/api';

/**
 * 获取成员列表
 */
export function fetchMemberList(params?: MemberListParams) {
  return request<Member[]>({
    url: `${API_BASE}/members`,
    method: 'GET',
    params
  });
}

/**
 * 获取分页成员列表
 */
export function fetchMemberListWithPaging(params?: MemberListParams) {
  return request<MemberListResponse>({
    url: `${API_BASE}/members`,
    method: 'GET',
    params: { ...params, page: params?.page || 1, pageSize: params?.pageSize || 20 }
  });
}

/**
 * 获取成员详情
 */
export function fetchMemberDetail(id: number) {
  return request<Member>({
    url: `${API_BASE}/members/${id}`,
    method: 'GET'
  });
}

/**
 * 创建成员
 */
export function createMember(data: MemberCreateRequest) {
  return request<Member>({
    url: `${API_BASE}/members`,
    method: 'POST',
    data
  });
}

/**
 * 更新成员
 */
export function updateMember(id: number, data: MemberUpdateRequest) {
  return request<Member>({
    url: `${API_BASE}/members/${id}/update`,
    method: 'POST',
    data
  });
}

/**
 * 删除成员
 */
export function deleteMember(id: number) {
  return request<{ success: boolean; message: string }>({
    url: `${API_BASE}/members/${id}/delete`,
    method: 'POST'
  });
}

/**
 * 获取成员统计
 */
export function fetchMemberStatistics() {
  return request<MemberStatistics>({
    url: `${API_BASE}/members/statistics`,
    method: 'GET'
  });
}

/**
 * 获取部门列表
 */
export function fetchDepartments() {
  return request<string[]>({
    url: `${API_BASE}/members/departments`,
    method: 'GET'
  });
}

/**
 * 搜索成员
 */
export function searchMembers(keyword: string, limit: number = 10) {
  return request<Member[]>({
    url: `${API_BASE}/members/search`,
    method: 'GET',
    params: { keyword, limit }
  });
}

/**
 * 激活成员
 */
export function activateMember(id: number) {
  return request<Member>({
    url: `${API_BASE}/members/${id}/activate`,
    method: 'POST'
  });
}

/**
 * 停用成员
 */
export function deactivateMember(id: number) {
  return request<Member>({
    url: `${API_BASE}/members/${id}/deactivate`,
    method: 'POST'
  });
}
