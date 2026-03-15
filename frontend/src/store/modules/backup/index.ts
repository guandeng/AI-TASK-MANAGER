import { ref } from 'vue';
import { defineStore } from 'pinia';
import {
  createBackup,
  deleteBackup,
  disableBackupSchedule,
  fetchBackupList,
  fetchBackupSchedule,
  restoreBackup,
  updateBackupSchedule
} from '@/service/api/backup';
import type { BackupRecord, BackupSchedule } from '@/service/api/backup';

// 辅助函数：提取后端返回的 data 字段
// 后端返回格式: { code: 0, message: "success", data: {...} }
function extractData(responseData: any): any {
  if (!responseData) return null;
  if (responseData.data !== undefined) {
    return responseData.data;
  }
  return responseData;
}

export const useBackupStore = defineStore('backup-store', () => {
  // 状态
  const backups = ref<BackupRecord[]>([]);
  const schedule = ref<BackupSchedule | null>(null);
  const loading = ref(false);
  const total = ref(0);

  // Actions
  async function loadBackups(requirementId: number, params?: { page?: number; pageSize?: number }) {
    loading.value = true;
    try {
      const { data, error } = await fetchBackupList(requirementId, params);
      if (!error && data) {
        const responseData = extractData(data);
        if (responseData && 'list' in responseData) {
          backups.value = responseData.list || [];
          total.value = responseData.total || 0;
        } else if (Array.isArray(responseData)) {
          backups.value = responseData;
          total.value = responseData.length;
        }
      }
    } catch (error) {
      console.error('Failed to load backups:', error);
      window.$message?.error('加载备份列表失败');
    } finally {
      loading.value = false;
    }
  }

  async function loadSchedule(requirementId: number) {
    try {
      const { data, error } = await fetchBackupSchedule(requirementId);
      if (!error && data) {
        schedule.value = extractData(data);
      }
    } catch (error) {
      console.error('Failed to load schedule:', error);
    }
  }

  async function createBackupAction(requirementId: number) {
    loading.value = true;
    try {
      const { data, error } = await createBackup(requirementId);
      if (!error && data) {
        const response = data as any;
        if (response?.code === 0) {
          window.$message?.success('备份创建成功');
          // 重新加载列表
          await loadBackups(requirementId);
          return true;
        }
      }
      window.$message?.error('备份创建失败');
    } catch (error: any) {
      console.error('Failed to create backup:', error);
      window.$message?.error(error?.response?.data?.message || '备份创建失败');
    } finally {
      loading.value = false;
    }
    return false;
  }

  async function restoreBackupAction(requirementId: number, backupId: number) {
    try {
      const { data, error } = await restoreBackup(requirementId, backupId);
      if (!error && data) {
        const response = data as any;
        if (response?.code === 0) {
          window.$message?.success('恢复成功');
          return true;
        }
      }
      window.$message?.error('恢复失败');
    } catch (error: any) {
      console.error('Failed to restore backup:', error);
      window.$message?.error(error?.response?.data?.message || '恢复失败');
    }
    return false;
  }

  async function deleteBackupAction(requirementId: number, backupId: number) {
    try {
      const { data, error } = await deleteBackup(requirementId, backupId);
      if (!error && data) {
        const response = data as any;
        if (response?.code === 0) {
          window.$message?.success('删除成功');
          // 重新加载列表
          await loadBackups(requirementId);
          return true;
        }
      }
      window.$message?.error('删除失败');
    } catch (error: any) {
      console.error('Failed to delete backup:', error);
      window.$message?.error(error?.response?.data?.message || '删除失败');
    }
    return false;
  }

  async function updateScheduleAction(requirementId: number, scheduleData: Partial<BackupSchedule>) {
    try {
      const result = await updateBackupSchedule(requirementId, scheduleData);
      if (result && !result.error && result.data) {
        const response = result.data as any;
        if (response?.code === 0) {
          window.$message?.success('保存成功');
          // 重新加载计划
          await loadSchedule(requirementId);
          return true;
        }
      }
      window.$message?.error('保存失败');
    } catch (error: any) {
      console.error('Failed to update schedule:', error);
      window.$message?.error(error?.response?.data?.message || '保存失败');
    }
    return false;
  }

  async function disableScheduleAction(requirementId: number) {
    try {
      const { data, error } = await disableBackupSchedule(requirementId);
      if (!error && data) {
        const response = data as any;
        if (response?.code === 0) {
          window.$message?.success('已禁用备份计划');
          // 重新加载计划
          await loadSchedule(requirementId);
          return true;
        }
      }
      window.$message?.error('禁用失败');
    } catch (error: any) {
      console.error('Failed to disable schedule:', error);
      window.$message?.error(error?.response?.data?.message || '禁用失败');
    }
    return false;
  }

  // 清空备份列表
  function clearBackups() {
    backups.value = [];
    total.value = 0;
  }

  return {
    // 状态
    backups,
    schedule,
    loading,
    total,
    // Actions
    loadBackups,
    loadSchedule,
    createBackupAction,
    restoreBackupAction,
    deleteBackupAction,
    updateScheduleAction,
    disableScheduleAction,
    clearBackups
  };
});
