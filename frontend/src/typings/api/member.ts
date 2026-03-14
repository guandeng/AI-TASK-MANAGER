/** 成员角色 */
export type MemberRole = 'admin' | 'leader' | 'member';

/** 成员状态 */
export type MemberStatus = 'active' | 'inactive';

/** 成员信息 */
export interface Member {
  id: number;
  name: string;
  email?: string;
  avatar?: string;
  role: MemberRole;
  department?: string;
  skills: string[];
  status: MemberStatus;
  createdAt: string;
  updatedAt: string;
}

/** 成员创建请求 */
export interface MemberCreateRequest {
  name: string;
  email?: string;
  avatar?: string;
  role?: MemberRole;
  department?: string;
  skills?: string[];
  status?: MemberStatus;
}

/** 成员更新请求 */
export interface MemberUpdateRequest {
  name?: string;
  email?: string;
  avatar?: string;
  role?: MemberRole;
  department?: string;
  skills?: string[];
  status?: MemberStatus;
}

/** 成员列表筛选参数 */
export interface MemberListParams {
  status?: MemberStatus;
  role?: MemberRole;
  department?: string;
  keyword?: string;
  page?: number;
  pageSize?: number;
}

/** 分页成员列表响应 */
export interface MemberListResponse {
  members: Member[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

/** 成员统计 */
export interface MemberStatistics {
  total: number;
  active: number;
  inactive: number;
  admin: number;
  leader: number;
  member: number;
  departments: Record<string, number>;
}

/** 成员角色选项 */
export const MEMBER_ROLE_OPTIONS = [
  { label: '管理员', value: 'admin' },
  { label: '组长', value: 'leader' },
  { label: '成员', value: 'member' }
] as const;

/** 成员状态选项 */
export const MEMBER_STATUS_OPTIONS = [
  { label: '活跃', value: 'active' },
  { label: '停用', value: 'inactive' }
] as const;

/** 角色显示名称映射 */
export const MEMBER_ROLE_LABELS: Record<MemberRole, string> = {
  admin: '管理员',
  leader: '组长',
  member: '成员'
};

/** 状态显示名称映射 */
export const MEMBER_STATUS_LABELS: Record<MemberStatus, string> = {
  active: '活跃',
  inactive: '停用'
};
