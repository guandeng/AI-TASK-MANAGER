<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { NButton, NCard, NDescriptions, NDescriptionsItem, NEmpty, NSpace, NSpin, NTag, NTimeline, NTimelineItem } from 'naive-ui';
import { useTaskStore } from '@/store/modules/task';

const route = useRoute();
const router = useRouter();
const taskStore = useTaskStore();

const statusTextMap: Record<string, string> = {
  pending: '待处理',
  'in-progress': '进行中',
  done: '已完成',
  deferred: '已延期'
};

const statusColorMap: Record<string, 'default' | 'success' | 'warning' | 'error' | 'info'> = {
  pending: 'default',
  'in-progress': 'info',
  done: 'success',
  deferred: 'warning'
};

const priorityTextMap: Record<string, string> = {
  high: '高',
  medium: '中',
  low: '低'
};

const priorityColorMap: Record<string, 'error' | 'warning' | 'success'> = {
  high: 'error',
  medium: 'warning',
  low: 'success'
};

const taskId = computed(() => Number(route.params.id));
const task = computed(() => taskStore.currentTask);

async function handleExpandTask() {
  if (!taskId.value) {
    return;
  }

  await taskStore.expandTask(taskId.value);
}

async function handleClearSubtasks() {
  if (!taskId.value || !task.value?.subtasks?.length) {
    return;
  }

  if (!window.confirm('确认清空全部子任务吗？')) {
    return;
  }

  await taskStore.clearTaskSubtasks(taskId.value);
}

async function handleDeleteTask() {
  if (!taskId.value) {
    return;
  }

  if (!window.confirm(`确认删除任务 ${taskId.value} 吗？删除后会同时删除全部子任务。`)) {
    return;
  }

  const success = await taskStore.deleteTask(taskId.value);
  if (success) {
    await router.push('/requirement/task-list');
  }
}

async function handleDeleteSubtask(subtaskId: number) {
  if (!taskId.value) {
    return;
  }

  if (!window.confirm(`确认删除子任务 ${taskId.value}.${subtaskId} 吗？`)) {
    return;
  }

  await taskStore.deleteSubtask(taskId.value, subtaskId);
}

onMounted(async () => {
  if (taskId.value) {
    await taskStore.loadTaskDetail(taskId.value);
  }
});

onUnmounted(() => {
  taskStore.clearCurrentTask();
});
</script>

<template>
  <div class="task-detail-page">
    <NSpace vertical :size="16">
      <NSpace>
        <NButton secondary @click="router.back()">返回</NButton>
        <NButton type="primary" :loading="taskStore.loading" @click="handleExpandTask">
          拆分子任务
        </NButton>
        <NButton
          type="error"
          ghost
          :disabled="!task?.subtasks?.length"
          :loading="taskStore.loading"
          @click="handleClearSubtasks"
        >
          清空子任务
        </NButton>
        <NButton type="error" :loading="taskStore.loading" @click="handleDeleteTask">
          删除任务
        </NButton>
      </NSpace>

      <NSpin :show="taskStore.loading">
        <NCard v-if="task" :title="task.titleTrans || task.title">
          <NSpace vertical :size="16">
            <NSpace>
              <NTag :type="statusColorMap[task.status]">{{ statusTextMap[task.status] }}</NTag>
              <NTag :type="priorityColorMap[task.priority]">{{ priorityTextMap[task.priority] }}</NTag>
            </NSpace>

            <NDescriptions bordered :column="1" label-placement="left">
              <NDescriptionsItem label="ID">{{ task.id }}</NDescriptionsItem>
              <NDescriptionsItem label="描述">{{ task.descriptionTrans || task.description || '-' }}</NDescriptionsItem>
              <NDescriptionsItem label="依赖">
                {{ task.dependencies?.length ? task.dependencies.join(', ') : '-' }}
              </NDescriptionsItem>
              <NDescriptionsItem label="实现细节">
                <div class="pre-wrap">{{ task.detailsTrans || task.details || '-' }}</div>
              </NDescriptionsItem>
              <NDescriptionsItem label="测试策略">
                <div class="pre-wrap">{{ task.testStrategyTrans || task.testStrategy || '-' }}</div>
              </NDescriptionsItem>
            </NDescriptions>

            <NCard title="子任务" size="small">
              <NTimeline v-if="task.subtasks?.length">
                <NTimelineItem
                  v-for="subtask in task.subtasks"
                  :key="subtask.id"
                  :title="`${task.id}.${subtask.id} ${subtask.titleTrans || subtask.title}`"
                  :content="subtask.descriptionTrans || subtask.description || '-'"
                >
                  <template #footer>
                    <NSpace align="center" justify="space-between" style="width: 100%;">
                      <NTag :type="statusColorMap[subtask.status]">{{ statusTextMap[subtask.status] }}</NTag>
                      <NButton
                        type="error"
                        size="small"
                        ghost
                        :loading="taskStore.loading"
                        @click="handleDeleteSubtask(subtask.id)"
                      >
                        删除
                      </NButton>
                    </NSpace>
                  </template>
                </NTimelineItem>
              </NTimeline>
              <NEmpty v-else description="暂无子任务" />
            </NCard>
          </NSpace>
        </NCard>

        <NEmpty v-else-if="!taskStore.loading" description="任务不存在" />
      </NSpin>
    </NSpace>
  </div>
</template>

<style scoped lang="scss">
.task-detail-page {
  padding: 16px;
}

.pre-wrap {
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
