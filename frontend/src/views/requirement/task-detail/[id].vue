<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { NButton, NCard, NDescriptions, NDescriptionsItem, NEmpty, NInput, NSpace, NSpin, NTag, NTimeline, NTimelineItem, NModal, NSelect } from 'naive-ui';
import type { SelectOption } from 'naive-ui';
import { useTaskStore } from '@/store/modules/task';
import type { TaskStatus } from '@/typings/api/task';

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

// 子任务状态选项
const subtaskStatusOptions: SelectOption[] = [
  { label: '待处理', value: 'pending' },
  { label: '进行中', value: 'in-progress' },
  { label: '已完成', value: 'done' },
  { label: '已延期', value: 'deferred' }
];

// 获取子任务状态的 CSS 类
function getSubtaskStatusClass(status: TaskStatus): string {
  return `subtask-status-select subtask-status-${status}`;
}

const taskId = computed(() => Number(route.params.id));
const task = computed(() => taskStore.currentTask);

// 独立的 loading 状态
const expandLoading = ref(false);
const clearLoading = ref(false);
const deleteLoading = ref(false);
const saveLoading = ref(false);

// 编辑模式状态
type EditableTaskField = 'title' | 'details' | 'testStrategy';
type EditableSubtaskField = 'title' | 'description';
const editingType = ref<'task' | 'subtask' | null>(null);
const editingField = ref<EditableTaskField | EditableSubtaskField | null>(null);
const editingSubtaskId = ref<number | null>(null);
const editingValue = ref('');

// 开始编辑任务字段
function startEditTask(field: EditableTaskField) {
  if (!task.value) return;
  editingType.value = 'task';
  editingField.value = field;
  editingSubtaskId.value = null;
  if (field === 'title') {
    editingValue.value = task.value.titleTrans || task.value.title || '';
  } else if (field === 'details') {
    editingValue.value = task.value.detailsTrans || task.value.details || '';
  } else {
    editingValue.value = task.value.testStrategyTrans || task.value.testStrategy || '';
  }
}

// 开始编辑子任务字段
function startEditSubtask(subtaskId: number, field: EditableSubtaskField) {
  if (!task.value?.subtasks) return;
  const subtask = task.value.subtasks.find(st => st.id === subtaskId);
  if (!subtask) return;

  editingType.value = 'subtask';
  editingField.value = field;
  editingSubtaskId.value = subtaskId;
  if (field === 'title') {
    editingValue.value = subtask.titleTrans || subtask.title || '';
  } else {
    editingValue.value = subtask.descriptionTrans || subtask.description || '';
  }
}

// 取消编辑
function cancelEdit() {
  editingType.value = null;
  editingField.value = null;
  editingSubtaskId.value = null;
  editingValue.value = '';
}

// 保存编辑
async function saveEdit() {
  if (!taskId.value || !editingField.value) return;

  saveLoading.value = true;
  try {
    if (editingType.value === 'task') {
      // 更新任务
      const field = editingField.value as EditableTaskField;
      let updateData: Record<string, string>;
      if (field === 'title') {
        updateData = { title: editingValue.value };
      } else if (field === 'details') {
        updateData = { details: editingValue.value };
      } else {
        updateData = { testStrategy: editingValue.value };
      }
      const success = await taskStore.updateTask(taskId.value, updateData);
      if (success) {
        cancelEdit();
        window.$message?.success('保存成功');
      } else {
        window.$message?.error('保存失败');
      }
    } else if (editingType.value === 'subtask' && editingSubtaskId.value !== null) {
      // 更新子任务
      const field = editingField.value as EditableSubtaskField;
      let updateData: Record<string, string>;
      if (field === 'title') {
        updateData = { title: editingValue.value };
      } else {
        updateData = { description: editingValue.value };
      }
      const success = await taskStore.updateSubtask(taskId.value, editingSubtaskId.value, updateData);
      if (success) {
        cancelEdit();
        window.$message?.success('保存成功');
      } else {
        window.$message?.error('保存失败');
      }
    }
  } catch (error) {
    window.$message?.error('保存失败');
  } finally {
    saveLoading.value = false;
  }
}

// 重写子任务相关状态
const showRegenerateModal = ref(false);
const regeneratePrompt = ref('');
const regeneratingSubtaskId = ref<number | null>(null);

async function handleExpandTask() {
  if (!taskId.value) return;

  expandLoading.value = true;
  try {
    await taskStore.expandTask(taskId.value);
  } finally {
    expandLoading.value = false;
  }
}

async function handleClearSubtasks() {
  if (!taskId.value || !task.value?.subtasks?.length) return;

  if (!window.confirm('确认清空全部子任务吗？')) return;

  clearLoading.value = true;
  try {
    await taskStore.clearTaskSubtasks(taskId.value);
  } finally {
    clearLoading.value = false;
  }
}

async function handleDeleteTask() {
  if (!taskId.value) return;

  if (!window.confirm(`确认删除任务 ${taskId.value} 吗？删除后会同时删除全部子任务。`)) return;

  deleteLoading.value = true;
  try {
    const success = await taskStore.deleteTask(taskId.value);
    if (success) {
      await router.push('/requirement/task-list');
    }
  } finally {
    deleteLoading.value = false;
  }
}

// 处理主任务状态变更
async function handleTaskStatusChange(status: TaskStatus) {
  if (!taskId.value) return;
  if (!window.confirm(`确认将任务状态改为"${statusTextMap[status]}"吗？`)) return;
  await taskStore.setTaskStatus(taskId.value, status);
}

// 处理子任务状态变更
async function handleSubtaskStatusChange(subtaskId: number, status: TaskStatus) {
  if (!taskId.value) return;
  if (!window.confirm(`确认将子任务状态改为"${statusTextMap[status]}"吗？`)) return;
  await taskStore.setSubtaskStatus(taskId.value, subtaskId, status);
}

async function handleDeleteSubtask(subtaskId: number) {
  if (!taskId.value) return;

  if (!window.confirm(`确认删除子任务吗？`)) return;

  await taskStore.deleteSubtask(taskId.value, subtaskId);
}

function handleOpenRegenerateModal(subtaskId: number) {
  regeneratingSubtaskId.value = subtaskId;
  regeneratePrompt.value = '';
  showRegenerateModal.value = true;
}

function handleCloseRegenerateModal() {
  showRegenerateModal.value = false;
  regeneratePrompt.value = '';
  regeneratingSubtaskId.value = null;
}

async function handleConfirmRegenerate() {
  if (!taskId.value || regeneratingSubtaskId.value === null) return;

  await taskStore.regenerateSubtask(
    taskId.value,
    regeneratingSubtaskId.value,
    regeneratePrompt.value || undefined
  );

  handleCloseRegenerateModal();
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
        <NButton type="primary" :loading="expandLoading" @click="handleExpandTask">
          拆分子任务
        </NButton>
        <NButton
          type="error"
          ghost
          :disabled="!task?.subtasks?.length"
          :loading="clearLoading"
          @click="handleClearSubtasks"
        >
          清空子任务
        </NButton>
        <NButton type="error" :loading="deleteLoading" @click="handleDeleteTask">
          删除任务
        </NButton>
      </NSpace>

      <NSpin :show="taskStore.loading">
        <NCard v-if="task">
          <template #header>
            <div v-if="editingType === 'task' && editingField === 'title'" class="edit-field">
              <NInput
                v-model:value="editingValue"
                placeholder="请输入任务标题"
                style="width: 100%;"
              />
              <NSpace class="mt-8px">
                <NButton type="primary" size="small" :loading="saveLoading" @click="saveEdit">保存</NButton>
                <NButton size="small" @click="cancelEdit">取消</NButton>
              </NSpace>
            </div>
            <div v-else class="editable-title" @click="startEditTask('title')">
              {{ task.titleTrans || task.title }}
              <NButton text type="primary" size="small" class="edit-btn">
                <template #icon><span class="i-mdi:pencil text-14px"></span></template>
              </NButton>
            </div>
          </template>

          <NSpace vertical :size="16">
            <NSpace align="center">
              <NSelect
                :value="task.status"
                :options="subtaskStatusOptions"
                size="small"
                :class="getSubtaskStatusClass(task.status)"
                style="width: 120px"
                @update:value="(value: TaskStatus) => handleTaskStatusChange(value)"
              />
              <NTag :type="priorityColorMap[task.priority]">{{ priorityTextMap[task.priority] }}</NTag>
            </NSpace>

            <NDescriptions bordered :column="1" label-placement="left">
              <NDescriptionsItem label="ID">{{ task.id }}</NDescriptionsItem>
              <NDescriptionsItem label="描述">{{ task.descriptionTrans || task.description || '-' }}</NDescriptionsItem>
              <NDescriptionsItem label="依赖">
                {{ task.dependencies?.length ? task.dependencies.join(', ') : '-' }}
              </NDescriptionsItem>
              <NDescriptionsItem label="实现细节">
                <div v-if="editingType === 'task' && editingField === 'details'" class="edit-field">
                  <NInput
                    v-model:value="editingValue"
                    type="textarea"
                    :autosize="{ minRows: 3, maxRows: 10 }"
                    placeholder="请输入实现细节"
                  />
                  <NSpace class="mt-8px">
                    <NButton type="primary" size="small" :loading="saveLoading" @click="saveEdit">保存</NButton>
                    <NButton size="small" @click="cancelEdit">取消</NButton>
                  </NSpace>
                </div>
                <div v-else class="editable-field" @click="startEditTask('details')">
                  <div class="pre-wrap">{{ task.detailsTrans || task.details || '-' }}</div>
                  <NButton text type="primary" size="small" class="edit-btn">
                    <template #icon><span class="i-mdi:pencil text-14px"></span></template>
                  </NButton>
                </div>
              </NDescriptionsItem>
              <NDescriptionsItem label="测试策略">
                <div v-if="editingType === 'task' && editingField === 'testStrategy'" class="edit-field">
                  <NInput
                    v-model:value="editingValue"
                    type="textarea"
                    :autosize="{ minRows: 3, maxRows: 10 }"
                    placeholder="请输入测试策略"
                  />
                  <NSpace class="mt-8px">
                    <NButton type="primary" size="small" :loading="saveLoading" @click="saveEdit">保存</NButton>
                    <NButton size="small" @click="cancelEdit">取消</NButton>
                  </NSpace>
                </div>
                <div v-else class="editable-field" @click="startEditTask('testStrategy')">
                  <div class="pre-wrap">{{ task.testStrategyTrans || task.testStrategy || '-' }}</div>
                  <NButton text type="primary" size="small" class="edit-btn">
                    <template #icon><span class="i-mdi:pencil text-14px"></span></template>
                  </NButton>
                </div>
              </NDescriptionsItem>
            </NDescriptions>

            <NCard title="子任务" size="small">
              <NTimeline v-if="task.subtasks?.length">
                <NTimelineItem
                  v-for="subtask in task.subtasks"
                  :key="subtask.id"
                >
                  <template #header>
                    <div v-if="editingType === 'subtask' && editingSubtaskId === subtask.id && editingField === 'title'" class="edit-field-inline">
                      <NInput
                        v-model:value="editingValue"
                        placeholder="请输入子任务标题"
                        style="width: 300px;"
                      />
                      <NButton type="primary" size="tiny" :loading="saveLoading" @click="saveEdit">保存</NButton>
                      <NButton size="tiny" @click="cancelEdit">取消</NButton>
                    </div>
                    <div v-else class="editable-inline" @click="startEditSubtask(subtask.id, 'title')">
                      {{ task.id }}.{{ subtask.id }} {{ subtask.titleTrans || subtask.title }}
                      <NButton text type="primary" size="tiny" class="edit-btn-inline">
                        <span class="i-mdi:pencil text-12px"></span>
                      </NButton>
                    </div>
                  </template>
                  <template #default>
                    <div v-if="editingType === 'subtask' && editingSubtaskId === subtask.id && editingField === 'description'" class="edit-field">
                      <NInput
                        v-model:value="editingValue"
                        type="textarea"
                        :autosize="{ minRows: 2, maxRows: 6 }"
                        placeholder="请输入子任务描述"
                      />
                      <NSpace class="mt-8px">
                        <NButton type="primary" size="small" :loading="saveLoading" @click="saveEdit">保存</NButton>
                        <NButton size="small" @click="cancelEdit">取消</NButton>
                      </NSpace>
                    </div>
                    <div v-else class="editable-field" @click="startEditSubtask(subtask.id, 'description')">
                      <div class="pre-wrap">{{ subtask.descriptionTrans || subtask.description || '-' }}</div>
                      <NButton text type="primary" size="small" class="edit-btn">
                        <template #icon><span class="i-mdi:pencil text-14px"></span></template>
                      </NButton>
                    </div>
                  </template>
                  <template #footer>
                    <NSpace align="center" justify="space-between" style="width: 100%;">
                      <NSelect
                        :value="subtask.status"
                        :options="subtaskStatusOptions"
                        size="small"
                        :class="getSubtaskStatusClass(subtask.status)"
                        style="width: 120px"
                        @update:value="(value: TaskStatus) => handleSubtaskStatusChange(subtask.id, value)"
                      />
                      <NSpace>
                        <NButton
                          type="warning"
                          size="small"
                          ghost
                          :loading="taskStore.loading"
                          @click="handleOpenRegenerateModal(subtask.id)"
                        >
                          重写
                        </NButton>
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

    <!-- 重写子任务弹窗 -->
    <NModal
      v-model:show="showRegenerateModal"
      preset="dialog"
      title="重写子任务"
      positive-text="确认重写"
      negative-text="取消"
      :loading="taskStore.loading"
      @positive-click="handleConfirmRegenerate"
      @negative-click="handleCloseRegenerateModal"
    >
      <NSpace vertical :size="16">
        <p>可选：输入提示词来指导子任务的重新生成（留空则自动重新生成）</p>
        <NInput
          v-model:value="regeneratePrompt"
          type="textarea"
          placeholder="例如：请更关注性能优化方面..."
          :rows="4"
        />
      </NSpace>
    </NModal>
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

.mt-8px {
  margin-top: 8px;
}

.text-12px {
  font-size: 12px;
}

.text-14px {
  font-size: 14px;
}

// 可编辑标题样式
.editable-title {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  margin: -4px -8px;
  border-radius: 4px;
  transition: background-color 0.2s;

  &:hover {
    background-color: rgba(0, 0, 0, 0.04);

    .edit-btn {
      opacity: 1;
    }
  }

  .edit-btn {
    opacity: 0;
    transition: opacity 0.2s;
  }
}

// 可编辑字段样式
.editable-field {
  position: relative;
  cursor: pointer;
  padding: 4px 8px;
  margin: -4px -8px;
  border-radius: 4px;
  transition: background-color 0.2s;

  &:hover {
    background-color: rgba(0, 0, 0, 0.04);

    .edit-btn {
      opacity: 1;
    }
  }

  .edit-btn {
    position: absolute;
    top: 4px;
    right: 8px;
    opacity: 0;
    transition: opacity 0.2s;
  }
}

// 可编辑内联样式（子任务标题）
.editable-inline {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  padding: 2px 6px;
  margin: -2px -6px;
  border-radius: 4px;
  transition: background-color 0.2s;

  &:hover {
    background-color: rgba(0, 0, 0, 0.04);

    .edit-btn-inline {
      opacity: 1;
    }
  }

  .edit-btn-inline {
    opacity: 0;
    transition: opacity 0.2s;
  }
}

// 内联编辑字段
.edit-field-inline {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.edit-field {
  width: 100%;
}

// 子任务状态选择框颜色
.subtask-status-select {
  :deep(.n-base-selection) {
    border-radius: 4px;
  }

  :deep(.n-base-selection-label) {
    font-weight: 500;
  }
}

.subtask-status-pending {
  :deep(.n-base-selection) {
    background-color: rgba(0, 0, 0, 0.04);
    border-color: #e0e0e6;
  }
  :deep(.n-base-selection-label) {
    color: #6b7280;
  }
}

.subtask-status-in-progress {
  :deep(.n-base-selection) {
    background-color: rgba(24, 144, 255, 0.1);
    border-color: #1890ff;
  }
  :deep(.n-base-selection-label) {
    color: #1890ff;
  }
}

.subtask-status-done {
  :deep(.n-base-selection) {
    background-color: rgba(24, 160, 88, 0.1);
    border-color: #18a058;
  }
  :deep(.n-base-selection-label) {
    color: #18a058;
  }
}

.subtask-status-deferred {
  :deep(.n-base-selection) {
    background-color: rgba(240, 160, 32, 0.1);
    border-color: #f0a020;
  }
  :deep(.n-base-selection-label) {
    color: #f0a020;
  }
}
</style>
