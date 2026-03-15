<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { NAvatar, NButton, NCard, NPopover, NSpace, NTag, NText } from 'naive-ui';
import type { Task } from '@/typings/api/task';

interface Props {
  task: Task;
}

const props = defineProps<Props>();
const router = useRouter();

// 优先级颜色映射
const priorityColorMap: Record<string, 'error' | 'warning' | 'success'> = {
  high: 'error',
  medium: 'warning',
  low: 'success'
};

// 优先级文字映射
const priorityTextMap: Record<string, string> = {
  high: '高',
  medium: '中',
  low: '低'
};

// 状态文字映射
const statusTextMap: Record<string, string> = {
  pending: '待处理',
  'in-progress': '进行中',
  done: '已完成',
  deferred: '已延期'
};

// 子任务进度
const subtaskProgress = computed(() => {
  if (!props.task.subtasks || props.task.subtasks.length === 0) {
    return { done: 0, total: 0 };
  }
  const done = props.task.subtasks.filter(st => st.status === 'done').length;
  return { done, total: props.task.subtasks.length };
});

// 是否逾期
const isOverdue = computed(() => {
  if (!props.task.dueDate || props.task.status === 'done') return false;
  const dueDate = new Date(props.task.dueDate);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  return dueDate < today;
});

// 剩余天数
const remainingDays = computed(() => {
  if (!props.task.dueDate) return null;
  const dueDate = new Date(props.task.dueDate);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const diff = dueDate.getTime() - today.getTime();
  return Math.ceil(diff / (1000 * 60 * 60 * 24));
});

// 查看详情
function viewDetail() {
  router.push(`/requirement/task-detail/${props.task.id}`);
}
</script>

<template>
  <NCard
    :title="task.titleTrans || task.title"
    size="small"
    class="task-kanban-card"
    :bordered="false"
    @click="viewDetail"
  >
    <template #header-extra>
      <NTag :type="priorityColorMap[task.priority]" size="small">
        {{ priorityTextMap[task.priority] }}
      </NTag>
    </template>

    <div class="card-content">
      <!-- 关联需求 -->
      <div v-if="task.requirementTitle" class="requirement-tag">
        <NText depth="3" style="font-size: 12px">
          <span class="i-mdi:link-variant"></span>
          {{ task.requirementTitle }}
        </NText>
      </div>

      <!-- 描述预览 -->
      <div v-if="task.description" class="description">
        <NText
          depth="3"
          style="
            font-size: 12px;
            display: -webkit-box;
            -webkit-line-clamp: 2;
            -webkit-box-orient: vertical;
            overflow: hidden;
          "
        >
          {{ task.description }}
        </NText>
      </div>

      <!-- 子任务进度 -->
      <div v-if="subtaskProgress.total > 0" class="subtask-progress">
        <NText depth="3" style="font-size: 12px">
          <span class="i-mdi:check-circle"></span>
          {{ subtaskProgress.done }}/{{ subtaskProgress.total }}
        </NText>
      </div>

      <!-- 截止日期 -->
      <div v-if="task.dueDate" class="due-date" :class="{ 'is-overdue': isOverdue }">
        <NText :type="isOverdue ? 'error' : '3'" style="font-size: 12px">
          <span class="i-mdi:calendar-clock"></span>
          {{ new Date(task.dueDate).toLocaleDateString('zh-CN') }}
          <span v-if="remainingDays !== null" class="days-left">
            {{ isOverdue ? `逾期${Math.abs(remainingDays)}天` : `剩余${remainingDays}天` }}
          </span>
        </NText>
      </div>
    </div>

    <template #footer>
      <div class="card-footer">
        <NSpace align="center" justify="space-between">
          <!-- 负责人 -->
          <div v-if="task.assignee" class="assignee">
            <NAvatar :size="20" :label="task.assignee.slice(0, 1)" color="blue" />
            <NText depth="3" style="font-size: 12px; margin-left: 4px">
              {{ task.assignee }}
            </NText>
          </div>
          <div v-else class="assignee-empty">
            <NText depth="3" style="font-size: 12px">
              <span class="i-mdi:account-off"></span>
              未分配
            </NText>
          </div>

          <!-- 操作按钮 -->
          <NButton size="tiny" text type="primary" @click.stop="viewDetail">详情</NButton>
        </NSpace>
      </div>
    </template>
  </NCard>
</template>

<style scoped lang="scss">
.task-kanban-card {
  cursor: pointer;
  transition: all 0.2s ease;
  margin-bottom: 12px;
  background-color: #fff;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    transform: translateY(-2px);
  }
}

.card-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.requirement-tag {
  display: flex;
  align-items: center;
  gap: 4px;
}

.description {
  line-height: 1.5;
}

.subtask-progress {
  display: flex;
  align-items: center;
  gap: 4px;
}

.due-date {
  display: flex;
  align-items: center;
  gap: 4px;

  &.is-overdue {
    color: #d03050;
  }

  .days-left {
    margin-left: 4px;
    font-weight: 500;
  }
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.assignee {
  display: flex;
  align-items: center;
}

.assignee-empty {
  display: flex;
  align-items: center;
  gap: 4px;
}
</style>
