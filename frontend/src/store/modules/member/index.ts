import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import type {
  Member,
  MemberCreateRequest,
  MemberUpdateRequest,
  MemberListParams,
  MemberStatistics
} from '@/typings/api/member';
import {
  fetchMemberList,
  fetchMemberDetail,
  createMember,
  updateMember,
  deleteMember,
  fetchMemberStatistics,
  searchMembers,
  activateMember,
  deactivateMember
} from '@/service/api/member';

/**
 * 成员管理 Store
 */
export const useMemberStore = defineStore('member-store', () => {
  // 状态
  const members = ref<Member[]>([]);
  const currentMember = ref<Member | null>(null);
  const statistics = ref<MemberStatistics | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  // 计算属性
  const activeMembers = computed(() => members.value.filter(m => m.status === 'active'));
  const adminMembers = computed(() => members.value.filter(m => m.role === 'admin'));
  const leaderMembers = computed(() => members.value.filter(m => m.role === 'leader'));

  // 成员按部门分组
  const membersByDepartment = computed(() => {
    const groups: Record<string, Member[]> = {};
    members.value.forEach(member => {
      const dept = member.department || '未分配';
      if (!groups[dept]) {
        groups[dept] = [];
      }
      groups[dept].push(member);
    });
    return groups;
  });

  /**
   * 加载成员列表
   */
  async function loadMembers(params?: MemberListParams) {
    loading.value = true;
    error.value = null;
    try {
      const { data, error: err } = await fetchMemberList(params);
      if (err) {
        error.value = err;
      } else if (data) {
        // 后端返回格式: { code: 0, message: "success", data: { list, total, page, pageSize } }
        const responseData = (data as any).data || data;

        if (Array.isArray(responseData)) {
          members.value = responseData;
        } else if (responseData && 'list' in responseData) {
          members.value = responseData.list || [];
        }
      }
    } finally {
      loading.value = false;
    }
  }

  /**
   * 加载成员详情
   */
  async function loadMemberDetail(id: number) {
    loading.value = true;
    error.value = null;
    try {
      const { data, error: err } = await fetchMemberDetail(id);
      if (err) {
        error.value = err;
        currentMember.value = null;
      } else {
        currentMember.value = data;
      }
    } finally {
      loading.value = false;
    }
  }

  /**
   * 创建新成员
   */
  async function createNewMember(memberData: MemberCreateRequest) {
    loading.value = true;
    error.value = null;
    try {
      const { data, error: err } = await createMember(memberData);
      if (err) {
        error.value = err;
        return null;
      }
      if (data) {
        members.value.push(data);
        return data;
      }
      return null;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 更新成员
   */
  async function updateMemberById(id: number, memberData: MemberUpdateRequest) {
    loading.value = true;
    error.value = null;
    try {
      const { data, error: err } = await updateMember(id, memberData);
      if (err) {
        error.value = err;
        return null;
      }
      if (data) {
        const index = members.value.findIndex(m => m.id === id);
        if (index !== -1) {
          members.value[index] = data;
        }
        if (currentMember.value?.id === id) {
          currentMember.value = data;
        }
        return data;
      }
      return null;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 删除成员
   */
  async function removeMember(id: number) {
    loading.value = true;
    error.value = null;
    try {
      const { error: err } = await deleteMember(id);
      if (err) {
        error.value = err;
        return false;
      }
      members.value = members.value.filter(m => m.id !== id);
      if (currentMember.value?.id === id) {
        currentMember.value = null;
      }
      return true;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 加载成员统计
   */
  async function loadStatistics() {
    try {
      const { data, error: err } = await fetchMemberStatistics();
      if (!err && data) {
        statistics.value = data;
      }
    } catch (e) {
      console.error('Failed to load member statistics:', e);
    }
  }

  /**
   * 搜索成员
   */
  async function search(keyword: string) {
    if (!keyword.trim()) {
      return [];
    }
    try {
      const { data, error: err } = await searchMembers(keyword);
      if (err) {
        return [];
      }
      return data || [];
    } catch (e) {
      console.error('Failed to search members:', e);
      return [];
    }
  }

  /**
   * 激活成员
   */
  async function activateMemberById(id: number) {
    const { data, error: err } = await activateMember(id);
    if (err) {
      return null;
    }
    if (data) {
      const index = members.value.findIndex(m => m.id === id);
      if (index !== -1) {
        members.value[index] = data;
      }
      return data;
    }
    return null;
  }

  /**
   * 停用成员
   */
  async function deactivateMemberById(id: number) {
    const { data, error: err } = await deactivateMember(id);
    if (err) {
      return null;
    }
    if (data) {
      const index = members.value.findIndex(m => m.id === id);
      if (index !== -1) {
        members.value[index] = data;
      }
      return data;
    }
    return null;
  }

  /**
   * 清空状态
   */
  function clearState() {
    members.value = [];
    currentMember.value = null;
    statistics.value = null;
    error.value = null;
  }

  return {
    // 状态
    members,
    currentMember,
    statistics,
    loading,
    error,

    // 计算属性
    activeMembers,
    adminMembers,
    leaderMembers,
    membersByDepartment,

    // Actions
    loadMembers,
    loadMemberDetail,
    createNewMember,
    updateMemberById,
    removeMember,
    loadStatistics,
    search,
    activateMemberById,
    deactivateMemberById,
    clearState
  };
});

export default useMemberStore;
