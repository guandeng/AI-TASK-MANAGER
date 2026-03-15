<script setup lang="ts">
import { computed } from 'vue';
import { NButton, NCard, NDescriptions, NDescriptionsItem, NInputNumber, NPopover, NSpace, NTag } from 'naive-ui';
import type { Task } from '@/typings/api/task';

interface Props {
  task: Task | null;
  editable?: boolean;
}

interface Emits {
  (e: 'update:task', data: Partial<Task>): void;
}

const props = withDefaults(defineProps<Props>(), {
  editable: false
});

const emit = defineEmits<Emits>();

// 格式化日期
function formatDate(dateStr?: string): string {
  if (!dateStr) return '-';
  try {
    const date = new Date(dateStr);
    return date.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit'
    });
  } catch {
    return dateStr;
  }
}

// 计算是否逾期
const isOverdue = computed(() => {
  if (!props.task?.dueDate) return false;
  if (props.task.status === 'done') return false;
  const dueDate = new Date(props.task.dueDate);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  return dueDate < today;
});

// 计算剩余天数
const remainingDays = computed(() => {
  if (!props.task?.dueDate) return null;
  const dueDate = new Date(props.task.dueDate);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const diff = dueDate.getTime() - today.getTime();
  return Math.ceil(diff / (1000 * 60 * 60 * 24));
});

// 计算工时进度
const hoursProgress = computed(() => {
  if (!props.task?.estimatedHours || !props.task.actualHours) return 0;
  return Math.min(100, Math.round((props.task.actualHours / props.task.estimatedHours) * 100));
});

// 工时状态标签
const hoursStatus = computed(() => {
  if (!props.task?.estimatedHours || !props.task.actualHours) return { text: '-', type: 'default' as const };
  const ratio = props.task.actualHours / props.task.estimatedHours;
  if (ratio >= 1) return { text: '已超支', type: 'error' as const };
  if (ratio >= 0.8) return { text: '接近预估', type: 'warning' as const };
  return { text: '正常', type: 'success' as const };
});
</script>

<template>
  <NCard title="时间信息" :bordered="false" size="small">
    <template #header-extra>
      <NSpace v-if="editable">
        <NButton size="small" text type="primary">编辑</NButton>
      </NSpace>
    </template>

    <NDescriptions bordered :column="2" size="small">
      <NDescriptionsItem label="开始日期">
        <NPopover v-if="task?.startDate" trigger="hover">
          <template #trigger>
            <span>{{ formatDate(task.startDate) }}</span>
          </template>
          <span>任务计划开始时间</span>
        </NPopover>
        <span v-else>-</span>
      </NDescriptionsItem>

      <NDescriptionsItem label="截止日期">
        <span :class="{ 'text-red': isOverdue }">
          {{ formatDate(task?.dueDate) }}
        </span>
        <NTag v-if="isOverdue" type="error" size="small" style="margin-left: 8px">已逾期</NTag>
        <NTag v-else-if="remainingDays && remainingDays <= 3" type="warning" size="small" style="margin-left: 8px">
          剩余 {{ remainingDays }} 天
        </NTag>
      </NDescriptionsItem>

      <NDescriptionsItem label="预估工时">
        <NInputNumber
          v-if="editable"
          :value="task?.estimatedHours"
          :min="0"
          :step="0.5"
          placeholder="小时"
          style="width: 120px"
          @update:value="val => emit('update:task', { estimatedHours: val || undefined })"
        />
        <span v-else>{{ task?.estimatedHours ? `${task.estimatedHours} 小时` : '-' }}</span>
      </NDescriptionsItem>

      <NDescriptionsItem label="实际工时">
        <NInputNumber
          v-if="editable"
          :value="task?.actualHours"
          :min="0"
          :step="0.5"
          placeholder="小时"
          style="width: 120px"
          @update:value="val => emit('update:task', { actualHours: val || undefined })"
        />
        <span v-else>
          {{ task?.actualHours ? `${task.actualHours} 小时` : '-' }}
          <NTag
            v-if="task?.estimatedHours && task.actualHours"
            :type="hoursStatus.type"
            size="small"
            style="margin-left: 8px"
          >
            {{ hoursStatus.text }}
          </NTag>
        </span>
      </NDescriptionsItem>

      <NDescriptionsItem label="完成时间" :span="2">
        {{ formatDate(task?.completedAt) }}
      </NDescriptionsItem>

      <NDescriptionsItem v-if="task?.estimatedHours" label="工时进度" :span="2">
        <div class="hours-progress-bar">
          <div
            class="hours-progress-fill"
            :style="{ width: `${hoursProgress}%` }"
            :class="{ 'hours-progress-overdue': hoursStatus.type === 'error' }"
          ></div>
          <span class="hours-progress-text">{{ hoursProgress }}%</span>
        </div>
      </NDescriptionsItem>
    </NDescriptions>
  </NCard>
</template>

<style scoped lang="scss">
.text-red {
  color: #d03050;
  font-weight: 500;
}

.hours-progress-bar {
  position: relative;
  height: 20px;
  background-color: #f0f0f0;
  border-radius: 4px;
  overflow: hidden;
}

.hours-progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #18a058 0%, #63e2b7 100%);
  transition: width 0.3s ease;
  border-radius: 4px;
}

.hours-progress-overdue {
  background: linear-gradient(90deg, #d03050 0%, #f080a0 100%);
}

.hours-progress-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 12px;
  color: #666;
  font-weight: 500;
}
</style>
