<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import {
  NCard,
  NSpace,
  NButton,
  NTag,
  NPopover,
  NScrollbar,
  NEmpty,
  NSpin
} from 'naive-ui';
import { VueDraggable } from 'vue-draggable-plus';
import { useTaskStore } from '@/store/modules/task';
import type { Task, TaskStatus } from '@/typings/api/task';
import TaskKanbanCard from './TaskKanbanCard.vue';

const router = useRouter();
const taskStore = useTaskStore();

// 看板列定义
const kanbanColumns = [
  { key: 'pending', title: '待处理', color: '#d0d0d0' },
  { key: 'in-progress', title: '进行中', color: '#2080f0' },
  { key: 'done', title: '已完成', color: '#18a058' },
  { key: 'deferred', title: '已延期', color: '#f0a020' }
] as const;

// 按状态分组的任务
const tasksByStatus = computed(() => {
  return {
    pending: taskStore.tasksByStatus.pending,
    'in-progress': taskStore.tasksByStatus.inProgress,
    done: taskStore.tasksByStatus.done,
    deferred: taskStore.tasksByStatus.deferred
  };
});

// 拖拽结束处理 - 更新任务状态
async function onDragEnd(task: Task, newIndex: number, oldIndex: number, fromColumn: string) {
  // 如果从一列拖拽到另一列，需要更新状态
  // 注意：vue-draggable-plus 的跨列拖拽需要特殊处理
  // 这里简化实现，实际状态更新在卡片组件中处理
}

// 刷新数据
async function refresh() {
  await taskStore.loadTasks({ pageSize: 100 });
}

onMounted(async () => {
  await refresh();
});
</script>

<template>
  <div class="task-board-page">
    <NCard
      title="任务看板"
      :bordered="false"
    >
      <template #header-extra>
        <NSpace>
          <NButton size="small" @click="refresh">
            刷新
          </NButton>
          <NButton
            size="small"
            type="primary"
            @click="router.push('/requirement/task-create')"
          >
            新建任务
          </NButton>
        </NSpace>
      </template>

      <NSpin :show="taskStore.loading">
        <div class="kanban-board">
          <div
            v-for="column in kanbanColumns"
            :key="column.key"
            class="kanban-column"
          >
            <div class="column-header" :style="{ borderTopColor: column.color }">
              <div class="column-title">
                <span class="column-dot" :style="{ backgroundColor: column.color }"></span>
                {{ column.title }}
              </div>
              <NTag :bordered="false" size="small">
                {{ tasksByStatus[column.key as keyof typeof tasksByStatus].length }}
              </NTag>
            </div>

            <div class="column-content">
              <NScrollbar style="max-height: calc(100vh - 250px)">
                <VueDraggable
                  :list="tasksByStatus[column.key as keyof typeof tasksByStatus]"
                  :group="'kanban'"
                  :animation="150"
                  :ghost-class="'ghost-card'"
                  :drag-class="'dragging-card'"
                  item-key="id"
                  @end="(evt: any) => {
                    const task = tasksByStatus[column.key as keyof typeof tasksByStatus][evt.newIndex];
                    if (task && task.status !== column.key) {
                      taskStore.setTaskStatus(task.id, column.key as TaskStatus);
                    }
                  }"
                >
                  <template #item="{ element }">
                    <TaskKanbanCard :task="element" />
                  </template>
                </VueDraggable>

                <NEmpty
                  v-if="tasksByStatus[column.key as keyof typeof tasksByStatus].length === 0"
                  description="暂无任务"
                  :image-size="60"
                />
              </NScrollbar>
            </div>
          </div>
        </div>
      </NSpin>
    </NCard>
  </div>
</template>

<style scoped lang="scss">
.task-board-page {
  padding: 16px;
  height: calc(100vh - 64px);
  overflow: hidden;
}

.kanban-board {
  display: flex;
  gap: 16px;
  height: 100%;
  overflow-x: auto;
}

.kanban-column {
  flex: 1;
  min-width: 280px;
  max-width: 350px;
  background-color: #f5f7f9;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.column-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: #fff;
  border-top: 3px solid;
  border-bottom: 1px solid #e8e8e8;
}

.column-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  font-size: 14px;
}

.column-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.column-content {
  flex: 1;
  padding: 12px;
  overflow: auto;
}

.ghost-card {
  opacity: 0.5;
  background-color: #e0e0e0;
}

.dragging-card {
  opacity: 0.8;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: rotate(2deg);
}
</style>
