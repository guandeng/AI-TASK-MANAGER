<script setup lang="ts">
import { computed, ref, watch, h } from 'vue';
import {
  NModal,
  NCard,
  NSpace,
  NButton,
  NForm,
  NFormItem,
  NSwitch,
  NRadioGroup,
  NRadio,
  NInputNumber,
  NDataTable,
  NTag,
  NPopconfirm,
  NSpin,
  NEmpty
} from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import { useBackupStore } from '@/store/modules/backup';
import dayjs from 'dayjs';

interface Props {
  show: boolean;
  requirementId: number;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  'update:show': [value: boolean];
}>();

const backupStore = useBackupStore();

// 本地状态
const scheduleForm = ref({
  enabled: false,
  intervalType: 'minute' as 'minute' | 'hour',
  intervalValue: 5,
  retentionCount: 10
});

// 加载状态
const submitLoading = ref(false);

// 备份列表列定义
const columns: DataTableColumns = [
  {
    title: '备份时间',
    key: 'createdAt',
    width: 180,
    render(row) {
      return dayjs(row.createdAt).format('YYYY-MM-DD HH:mm:ss');
    }
  },
  {
    title: '类型',
    key: 'type',
    width: 80,
    render(row) {
      const typeMap: Record<string, { label: string; type: 'info' | 'success' | 'warning' }> = {
        full: { label: '完整', type: 'info' }
      };
      const config = typeMap[row.type] || { label: row.type, type: 'info' };
      return h(NTag, { type: config.type }, () => config.label);
    }
  },
  {
    title: '任务数',
    key: 'taskCount',
    width: 80,
    align: 'center'
  },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render(row) {
      const statusMap: Record<string, { label: string; type: 'success' | 'error' | 'warning' }> = {
        success: { label: '成功', type: 'success' },
        failed: { label: '失败', type: 'error' }
      };
      const config = statusMap[row.status] || { label: row.status, type: 'warning' };
      return h(NTag, { type: config.type }, () => config.label);
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    fixed: 'right',
    render(row) {
      return h(NSpace, [
        h(NButton, {
          size: 'small',
          type: 'info',
          text: true,
          onClick: () => handleRestore(row.id)
        }, () => '恢复'),
        h(NPopconfirm, {
          onPositiveClick: () => handleDelete(row.id),
          defaultShow: false
        }, {
          trigger: () => h(NButton, {
            size: 'small',
            type: 'error',
            text: true
          }, () => '删除'),
          default: () => '确认删除此备份？'
        })
      ]);
    }
  }
];

// 表单验证
function validateForm() {
  if (scheduleForm.value.intervalValue < 1) {
    window.$message?.warning('间隔值不能小于 1');
    return false;
  }
  if (scheduleForm.value.retentionCount < 1) {
    window.$message?.warning('保留数量不能小于 1');
    return false;
  }
  return true;
}

// 保存配置
async function handleSaveSchedule() {
  if (!validateForm()) return;

  submitLoading.value = true;
  try {
    const success = await backupStore.updateScheduleAction(props.requirementId, {
      enabled: scheduleForm.value.enabled,
      intervalType: scheduleForm.value.intervalType,
      intervalValue: scheduleForm.value.intervalValue,
      retentionCount: scheduleForm.value.retentionCount
    });
    if (success) {
      emit('update:show', false);
    }
  } finally {
    submitLoading.value = false;
  }
}

// 立即备份
async function handleCreateBackup() {
  submitLoading.value = true;
  try {
    await backupStore.createBackupAction(props.requirementId);
  } finally {
    submitLoading.value = false;
  }
}

// 恢复备份
async function handleRestore(backupId: number) {
  if (!window.confirm('确认恢复到此备份？恢复将覆盖当前需求的所有任务和设置。')) {
    return;
  }
  await backupStore.restoreBackupAction(props.requirementId, backupId);
}

// 删除备份
async function handleDelete(backupId: number) {
  await backupStore.deleteBackupAction(props.requirementId, backupId);
}

// 弹框打开时加载数据
watch(
  () => props.show,
  async (newVal) => {
    if (newVal) {
      await Promise.all([
        backupStore.loadSchedule(props.requirementId),
        backupStore.loadBackups(props.requirementId, { pageSize: 50 })
      ]);
      // 填充表单
      if (backupStore.schedule) {
        const s = backupStore.schedule;
        scheduleForm.value = {
          enabled: s.enabled,
          intervalType: (s.intervalType as 'minute' | 'hour') || 'minute',
          intervalValue: s.intervalValue || 5,
          retentionCount: s.retentionCount || 10
        };
      } else {
        // 默认值
        scheduleForm.value = {
          enabled: false,
          intervalType: 'minute',
          intervalValue: 5,
          retentionCount: 10
        };
      }
    }
  },
  { immediate: true }
);
</script>

<template>
  <NModal
    :show="props.show"
    preset="card"
    title="定时备份"
    size="large"
    :closable="true"
    :close-on-esc="true"
    class="backup-modal"
    @update:show="(val) => emit('update:show', val)"
  >
    <div class="backup-content">
      <!-- 备份配置 -->
      <NCard title="备份配置" size="small" class="mb-16">
        <NForm :model="scheduleForm" label-placement="left" label-width="100">
          <NFormItem label="启用备份">
            <NSwitch v-model:value="scheduleForm.enabled" />
          </NFormItem>
          <NFormItem label="备份频率">
            <NRadioGroup v-model:value="scheduleForm.intervalType">
              <NSpace>
                <NRadio value="minute">每分钟</NRadio>
                <NRadio value="hour">每小时</NRadio>
              </NSpace>
            </NRadioGroup>
          </NFormItem>
          <NFormItem :label="`每 ${scheduleForm.intervalType === 'minute' ? '分钟' : '小时'}`">
            <NInputNumber
              v-model:value="scheduleForm.intervalValue"
              :min="1"
              :max="60"
              :disabled="!scheduleForm.enabled"
            />
          </NFormItem>
          <NFormItem label="保留数量">
            <NInputNumber
              v-model:value="scheduleForm.retentionCount"
              :min="1"
              :max="100"
              :disabled="!scheduleForm.enabled"
            />
            <span class="text-gray-400 text-sm ml-2">最近保留的备份数量</span>
          </NFormItem>
          <NFormItem label="上次备份">
            {{ backupStore.schedule?.lastBackupAt ? dayjs(backupStore.schedule.lastBackupAt).format('YYYY-MM-DD HH:mm:ss') : '-' }}
          </NFormItem>
          <NFormItem label=" ">
            <NSpace>
              <NButton
                type="primary"
                :loading="submitLoading"
                :disabled="!scheduleForm.enabled"
                @click="handleSaveSchedule"
              >
                保存配置
              </NButton>
              <NButton
                v-if="scheduleForm.enabled"
                type="warning"
                ghost
                @click="backupStore.disableScheduleAction(requirementId)"
              >
                禁用
              </NButton>
            </NSpace>
          </NFormItem>
        </NForm>
      </NCard>

      <!-- 备份历史 -->
      <NCard title="备份历史" size="small">
        <template #header-extra>
          <NButton type="success" :loading="submitLoading" @click="handleCreateBackup">
            立即备份
          </NButton>
        </template>
        <NSpin :show="backupStore.loading">
          <NDataTable
            :columns="columns"
            :data="backupStore.backups"
            :row-key="(row: any) => row.id"
            :bordered="false"
            :max-height="400"
          />
          <NEmpty v-if="!backupStore.backups.length && !backupStore.loading" description="暂无备份记录" />
        </NSpin>
      </NCard>
    </div>
  </NModal>
</template>

<style scoped lang="scss">
.backup-modal {
  width: 900px;
  max-width: 90vw;
}

.backup-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.mb-16 {
  margin-bottom: 16px;
}

.text-gray-400 {
  color: #9ca3af;
}

.text-sm {
  font-size: 12px;
}

.ml-2 {
  margin-left: 8px;
}
</style>
