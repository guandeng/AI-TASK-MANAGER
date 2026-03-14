<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import {
  NCard,
  NSpace,
  NTimeline,
  NTimelineItem,
  NTag,
  NEmpty,
  NSpin,
  NButton,
  NSelect,
  NPopover,
  NEllipsis
} from 'naive-ui';
import type { SelectOption } from 'naive-ui';
import { fetchTaskActivities, fetchGlobalActivities } from '@/service/api/activity';
import type { Activity, ActivityAction } from '@/typings/api/activity';
import { ACTIVITY_TYPE_LABELS } from '@/typings/api/activity';

const props = defineProps<{
  taskId?: number; // 可选，不传则显示全局活动
  limit?: number;
}>();

// 状态
const loading = ref(false);
const activities = ref<Activity[]>([]);
const filterAction = ref<ActivityAction | 'all'>('all');

// 操作类型选项
const actionOptions: SelectOption[] = [
  { label: '全部操作', value: 'all' },
  { label: '创建任务', value: 'task_created' },
  { label: '更新任务', value: 'task_updated' },
  { label: '状态变更', value: 'task_status_changed' },
  { label: '优先级变更', value: 'task_priority_changed' },
  { label: '分配任务', value: 'task_assigned' },
  { label: '评论', value: 'comment_added' },
  { label: '子任务', value: 'subtask_created' }
];

// 过滤后的活动
const filteredActivities = computed(() => {
  if (filterAction.value === 'all') {
    return activities.value;
  }
  return activities.value.filter(a => a.action === filterAction.value);
});

// 加载活动日志
async function loadActivities() {
  loading.value = true;
  try {
    let result;
    if (props.taskId) {
      result = await fetchTaskActivities(props.taskId, { limit: props.limit || 50 });
    } else {
      result = await fetchGlobalActivities({ limit: props.limit || 50 });
    }

    const { data, error } = result;
    if (!error && data) {
      activities.value = data;
    }
  } finally {
    loading.value = false;
  }
}

// 格式化时间
function formatTime(dateStr: string) {
  const date = new Date(dateStr);
  const now = new Date();
  const diff = now.getTime() - date.getTime();

  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);

  if (minutes < 1) return '刚刚';
  if (minutes < 60) return `${minutes}分钟前`;
  if (hours < 24) return `${hours}小时前`;
  if (days < 7) return `${days}天前`;

  return date.toLocaleDateString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  });
}

// 获取操作标签
function getActionLabel(action: ActivityAction): string {
  return ACTIVITY_TYPE_LABELS[action] || action;
}

// 获取操作颜色
function getActionColor(action: ActivityAction): 'default' | 'success' | 'info' | 'warning' | 'error' {
  const colorMap: Partial<Record<ActivityAction, 'default' | 'success' | 'info' | 'warning' | 'error'>> = {
    task_created: 'info',
    task_updated: 'default',
    task_deleted: 'error',
    task_status_changed: 'warning',
    task_priority_changed: 'warning',
    task_assigned: 'info',
    task_unassigned: 'default',
    subtask_created: 'info',
    subtask_completed: 'success',
    comment_added: 'default',
    time_logged: 'info'
  };
  return colorMap[action] || 'default';
}

// 获取操作图标类型
function getTimelineType(action: ActivityAction): 'default' | 'success' | 'info' | 'warning' | 'error' {
  return getActionColor(action);
}

// 生成活动描述
function getActivityDescription(activity: Activity): string {
  const memberName = activity.member?.name || '系统';
  const fieldName = activity.fieldName ? ` (${activity.fieldName})` : '';

  switch (activity.action) {
    case 'task_created':
      return `${memberName} 创建了任务`;
    case 'task_updated':
      return `${memberName} 更新了任务${fieldName}`;
    case 'task_deleted':
      return `${memberName} 删除了任务`;
    case 'task_status_changed':
      return `${memberName} 将状态从 "${activity.oldValue || '无'}" 改为 "${activity.newValue || '无'}"`;
    case 'task_priority_changed':
      return `${memberName} 将优先级从 "${activity.oldValue || '无'}" 改为 "${activity.newValue || '无'}"`;
    case 'task_assigned':
      return `${memberName} 将任务分配给了 ${activity.newValue || '成员'}`;
    case 'task_unassigned':
      return `${memberName} 取消了任务分配`;
    case 'comment_added':
      return `${memberName} 添加了评论`;
    case 'comment_updated':
      return `${memberName} 更新了评论`;
    case 'comment_deleted':
      return `${memberName} 删除了评论`;
    case 'subtask_created':
      return `${memberName} 创建了子任务`;
    case 'subtask_updated':
      return `${memberName} 更新了子任务`;
    case 'subtask_deleted':
      return `${memberName} 删除了子任务`;
    case 'subtask_status_changed':
      return `${memberName} 更新了子任务状态`;
    case 'subtask_assigned':
      return `${memberName} 分配了子任务`;
    case 'subtask_unassigned':
      return `${memberName} 取消了子任务分配`;
    case 'time_estimated':
      return `${memberName} 设置预估工时为 ${activity.newValue}小时`;
    case 'time_logged':
      return `${memberName} 记录了 ${activity.newValue}小时工时`;
    case 'due_date_changed':
      return `${memberName} 修改了截止日期`;
    default:
      return `${memberName} 执行了操作`;
  }
}

// 监听 taskId 变化
watch(() => props.taskId, () => {
  loadActivities();
}, { immediate: true });
</script>

<template>
  <NCard :title="taskId ? '活动记录' : '全局活动'" size="small">
    <template #header-extra>
      <NSpace align="center">
        <NSelect
          v-model:value="filterAction"
          :options="actionOptions"
          size="small"
          style="width: 120px"
        />
        <NButton size="small" @click="loadActivities">刷新</NButton>
      </NSpace>
    </template>

    <NSpin :show="loading">
      <div v-if="filteredActivities.length > 0" class="activity-timeline">
        <NTimeline>
          <NTimelineItem
            v-for="activity in filteredActivities"
            :key="activity.id"
            :type="getTimelineType(activity.action)"
            :title="getActionLabel(activity.action)"
          >
            <template #header>
              <NSpace align="center" :size="8">
                <NTag :type="getActionColor(activity.action)" size="small">
                  {{ getActionLabel(activity.action) }}
                </NTag>
                <span class="activity-time">{{ formatTime(activity.createdAt) }}</span>
              </NSpace>
            </template>

            <div class="activity-content">
              <div class="activity-description">
                {{ getActivityDescription(activity) }}
              </div>

              <!-- 显示详细变更 -->
              <NPopover
                v-if="activity.oldValue || activity.newValue"
                trigger="hover"
                placement="right"
              >
                <template #trigger>
                  <NButton text size="tiny" type="info">查看详情</NButton>
                </template>
                <div class="change-details">
                  <div v-if="activity.oldValue" class="change-item">
                    <span class="label">旧值：</span>
                    <span class="value old">{{ activity.oldValue }}</span>
                  </div>
                  <div v-if="activity.newValue" class="change-item">
                    <span class="label">新值：</span>
                    <span class="value new">{{ activity.newValue }}</span>
                  </div>
                </div>
              </NPopover>

              <!-- 任务链接（全局活动时显示） -->
              <div v-if="!taskId && activity.task" class="task-link">
                <NTag size="small" :bordered="false">
                  {{ activity.task.title }}
                </NTag>
              </div>
            </div>
          </NTimelineItem>
        </NTimeline>
      </div>
      <NEmpty v-else description="暂无活动记录" />
    </NSpin>
  </NCard>
</template>

<style scoped lang="scss">
.activity-timeline {
  max-height: 500px;
  overflow-y: auto;
}

.activity-time {
  font-size: 12px;
  color: #999;
}

.activity-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 4px;
}

.activity-description {
  font-size: 13px;
  color: #666;
  line-height: 1.5;
}

.change-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px 12px;
  background: #fafafa;
  border-radius: 4px;
  min-width: 200px;

  .change-item {
    display: flex;
    gap: 8px;
    font-size: 13px;

    .label {
      color: #999;
    }

    .value {
      flex: 1;
      word-break: break-all;

      &.old {
        color: #d03050;
        text-decoration: line-through;
      }

      &.new {
        color: #18a058;
      }
    }
  }
}

.task-link {
  margin-top: 4px;
}
</style>
