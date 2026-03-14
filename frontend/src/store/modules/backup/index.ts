import { defineStore } from 'pinia';
import { ref } from 'vue';
import {
  fetchBackupList,
  createBackup,
  restoreBackup,
  deleteBackup,
  fetchBackupSchedule,
  updateBackupSchedule,
  disableBackupSchedule
} from '@/service/api/backup';
import type { BackupRecord, BackupSchedule } from '@/service/api/backup';

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
        const responseData = (data as any).data || data;
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
        schedule.value = data;
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
        window.$message?.success('备份创建成功');
        // 重新加载列表
        await loadBackups(requirementId);
        return true;
      } else {
        window.$message?.error('备份创建失败');
      }
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
        window.$message?.success('恢复成功');
        return true;
      } else {
        window.$message?.error('恢复失败');
      }
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
        window.$message?.success('删除成功');
        // 重新加载列表
        await loadBackups(requirementId);
        return true;
      } else {
        window.$message?.error('删除失败');
      }
    } catch (error: any) {
      console.error('Failed to delete backup:', error);
      window.$message?.error(error?.response?.data?.message || '删除失败');
    }
    return false;
  }

  async function updateScheduleAction(requirementId: number, data: Partial<BackupSchedule>) {
    try {
      const result = await updateBackupSchedule(requirementId, data);
      if (result && !result.error) {
        window.$message?.success('保存成功');
        // 重新加载计划
        await loadSchedule(requirementId);
        return true;
      } else {
        window.$message?.error('保存失败');
      }
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
        window.$message?.success('已禁用备份计划');
        // 重新加载计划
        await loadSchedule(requirementId);
        return true;
      } else {
        window.$message?.error('禁用失败');
      }
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
