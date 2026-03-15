<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  NAlert,
  NButton,
  NCard,
  NCheckbox,
  NCode,
  NCollapse,
  NCollapseItem,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NDropdown,
  NEmpty,
  NIcon,
  NInput,
  NModal,
  NSelect,
  NSpace,
  NSpin,
  NTabPane,
  NTabs,
  NTag
} from 'naive-ui';
import type { DropdownOption, SelectOption } from 'naive-ui';
import { VueDraggable } from 'vue-draggable-plus';
import { MdEditor, MdPreview } from 'md-editor-v3';
import 'md-editor-v3/lib/style.css';
import { fetchTaskList } from '@/service/api/task';
import { useTaskStore } from '@/store/modules/task';
import { useRequirementStore } from '@/store/modules/requirement';
import { useMemberStore } from '@/store/modules/member';
import { useMessageStore } from '@/store/modules/message';
import type { Subtask, Task, TaskStatus } from '@/typings/api/task';
import CommentSection from '@/components/task/CommentSection.vue';
import AssignmentPanel from '@/components/task/AssignmentPanel.vue';
import ActivityTimeline from '@/components/task/ActivityTimeline.vue';
import DependencyGraph from '@/components/task/DependencyGraph.vue';
import ScoreHistoryDrawer from '@/components/task/ScoreHistoryDrawer.vue';
import SvgIcon from '@/components/custom/svg-icon.vue';

const route = useRoute();
const router = useRouter();
const taskStore = useTaskStore();
const requirementStore = useRequirementStore();
const memberStore = useMemberStore();
const messageStore = useMessageStore();

// 当前用户成员ID（用于评论功能）
const currentMemberId = computed(() => memberStore.currentMember?.id);

const statusTextMap: Record<string, string> = {
  pending: '待处理',
  'in-progress': '进行中',
  done: '已完成',
  deferred: '已延期',
  paused: '已暂停'
};

const statusColorMap: Record<string, 'default' | 'success' | 'warning' | 'error' | 'info'> = {
  pending: 'default',
  'in-progress': 'info',
  done: 'success',
  deferred: 'warning',
  paused: 'warning'
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
  { label: '已暂停', value: 'paused' },
  { label: '已完成', value: 'done' },
  { label: '已延期', value: 'deferred' }
];

// 获取子任务状态的 CSS 类
function getSubtaskStatusClass(status: TaskStatus): string {
  return `subtask-status-select subtask-status-${status}`;
}

const taskId = computed(() => Number(route.params.id));
const task = computed(() => taskStore.currentTask);

// 需求列表下拉框相关
const requirementTaskList = ref<Task[]>([]);
const requirementTaskLoading = ref(false);
const showTaskDropdown = ref(false);
const currentTaskIndex = computed(() => {
  if (!task.value || requirementTaskList.value.length === 0) return -1;
  return requirementTaskList.value.findIndex(t => t.id === task.value?.id);
});

// 当前任务在列表中的显示名称
const currentTaskLabel = computed(() => {
  if (!task.value) return '任务详情';
  const index = currentTaskIndex.value;
  if (index >= 0 && requirementTaskList.value.length > 0) {
    return `${index + 1} / ${requirementTaskList.value.length} - ${task.value.title?.slice(0, 20) || '任务'}`;
  }
  return task.value.title?.slice(0, 20) || '任务详情';
});

// 下拉选项
const taskDropdownOptions = computed<DropdownOption[]>(() => {
  return requirementTaskList.value.map((t, index) => ({
    label: `${index + 1}. ${t.title || '无标题'}`,
    key: t.id,
    disabled: t.id === task.value?.id
  }));
});

// 独立的 loading 状态
const expandLoading = ref(false);
const clearLoading = ref(false);
const deleteLoading = ref(false);
const saveLoading = ref(false);
const scoreLoading = ref(false);

// 评分历史抽屉
const showScoreDrawer = ref(false);

// 拆分超时检测
const EXPAND_TIMEOUT_MS = 5 * 60 * 1000; // 5分钟
const expandStartTime = ref<number | null>(null);

// 判断是否超时（超过5分钟）
const isExpandTimeout = computed(() => {
  if (!task.value?.isExpanding) return false;

  // 优先使用本地记录的开始时间
  if (expandStartTime.value) {
    return Date.now() - expandStartTime.value > EXPAND_TIMEOUT_MS;
  }

  // 使用后端返回的拆分开始时间
  if (task.value?.expandStartedAt) {
    const startedTime = new Date(task.value.expandStartedAt).getTime();
    return Date.now() - startedTime > EXPAND_TIMEOUT_MS;
  }

  return false;
});

// 实际的loading状态（考虑超时）
const actualExpandLoading = computed(() => {
  return (task.value?.isExpanding || expandLoading.value) && !isExpandTimeout.value;
});

// 实际的disabled状态（考虑超时）
const actualExpandDisabled = computed(() => {
  return task.value?.isExpanding && !isExpandTimeout.value;
});

// 编辑模式状态
type EditableTaskField = 'title' | 'details' | 'testStrategy';
type EditableSubtaskField =
  | 'title'
  | 'description'
  | 'details'
  | 'codeInterface'
  | 'acceptanceCriteria'
  | 'relatedFiles'
  | 'codeHints';
const editingType = ref<'task' | 'subtask' | null>(null);
const editingField = ref<EditableTaskField | EditableSubtaskField | null>(null);
const editingSubtaskId = ref<number | null>(null);
const editingValue = ref('');

// 预览模式（用于 Markdown 字段）
const previewFields = ref<Record<string, boolean>>({});

// 子任务展开状态
const expandedSubtaskIds = ref<Set<number>>(new Set());

// 解析 JSON 字段
function parseJsonField<T>(field: string | null | undefined): T | null {
  if (!field) return null;
  try {
    return JSON.parse(field) as T;
  } catch {
    return null;
  }
}

// 格式化字段名（蛇形/驼峰转中文或带空格的文本）
function formatFieldName(key: string): string {
  // 蛇形命名转空格：module归属 -> 模块归属
  const snakeToSpace = key.replace(/_/g, ' ');
  // 驼峰命名转空格：estimatedHours -> estimated Hours
  const camelToSpace = snakeToSpace.replace(/([a-z])([A-Z])/g, '$1 $2');
  // 首字母大写
  return camelToSpace.charAt(0).toUpperCase() + camelToSpace.slice(1);
}

// 切换子任务展开状态
function toggleSubtaskExpand(subtaskId: number) {
  if (expandedSubtaskIds.value.has(subtaskId)) {
    expandedSubtaskIds.value.delete(subtaskId);
  } else {
    expandedSubtaskIds.value.add(subtaskId);
  }
}

// 检查子任务是否展开
function isSubtaskExpanded(subtaskId: number): boolean {
  return expandedSubtaskIds.value.has(subtaskId);
}

// 获取预览状态 key
function getPreviewKey(type: 'task' | 'subtask', field: string, subtaskId?: number): string {
  return subtaskId !== undefined ? `${type}-${field}-${subtaskId}` : `${type}-${field}`;
}

// 切换预览模式
function togglePreview(type: 'task' | 'subtask', field: string, subtaskId?: number) {
  const key = getPreviewKey(type, field, subtaskId);
  previewFields.value[key] = !previewFields.value[key];
}

// 检查是否处于预览模式
function isPreviewMode(type: 'task' | 'subtask', field: string, subtaskId?: number): boolean {
  const key = getPreviewKey(type, field, subtaskId);
  return previewFields.value[key] ?? true; // 默认是预览模式
}

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
  // 切换到编辑模式
  if (field !== 'title') {
    togglePreview('task', field);
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
  } else if (field === 'description') {
    editingValue.value = subtask.descriptionTrans || subtask.description || '';
  } else if (field === 'details') {
    editingValue.value = subtask.detailsTrans || subtask.details || '';
  } else if (field === 'codeInterface') {
    editingValue.value = subtask.codeInterface || '';
  } else if (field === 'acceptanceCriteria') {
    editingValue.value = subtask.acceptanceCriteria || '';
  } else if (field === 'relatedFiles') {
    editingValue.value = subtask.relatedFiles || '';
  } else if (field === 'codeHints') {
    editingValue.value = subtask.codeHints || '';
  }
  // 切换到编辑模式
  if (field === 'description' || field === 'details') {
    togglePreview('subtask', field, subtaskId);
  }
}

// 取消编辑
function cancelEdit() {
  // 如果是 Markdown 字段，切回预览模式
  if (editingType.value && editingField.value) {
    const field = editingField.value;
    if (field === 'details' || field === 'testStrategy') {
      togglePreview('task', field);
    } else if ((field === 'description' || field === 'details') && editingSubtaskId.value !== null) {
      togglePreview('subtask', field as 'description' | 'details', editingSubtaskId.value);
    }
  }
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
      } else if (field === 'testStrategy') {
        updateData = { testStrategy: editingValue.value };
      } else {
        // 不应该到达这里
        return;
      }
      const success = await taskStore.updateTask(taskId.value, updateData);
      if (success) {
        // 先取消编辑状态，再切换预览模式
        cancelEdit();
        // 切回预览模式
        if (field === 'details' || field === 'testStrategy') {
          togglePreview('task', field);
        }
        window.$message?.success('保存成功');
      } else {
        window.$message?.error('保存失败');
      }
    } else if (editingType.value === 'subtask' && editingSubtaskId.value !== null) {
      // 更新子任务
      const field = editingField.value as EditableSubtaskField;
      console.log('[saveEdit] 更新子任务', { subtaskId: editingSubtaskId.value, field, value: editingValue.value });
      let updateData: Record<string, unknown>;
      if (field === 'title') {
        updateData = { title: editingValue.value };
      } else if (field === 'description') {
        updateData = { description: editingValue.value };
      } else if (field === 'details') {
        updateData = { details: editingValue.value };
      } else if (field === 'codeInterface') {
        try {
          updateData = { codeInterface: editingValue.value ? JSON.parse(editingValue.value) : null };
        } catch {
          window.$message?.error('代码接口 JSON 格式不正确');
          return;
        }
      } else if (field === 'acceptanceCriteria') {
        try {
          updateData = { acceptanceCriteria: editingValue.value ? JSON.parse(editingValue.value) : null };
        } catch {
          window.$message?.error('验收标准 JSON 格式不正确');
          return;
        }
      } else if (field === 'relatedFiles') {
        try {
          updateData = { relatedFiles: editingValue.value ? JSON.parse(editingValue.value) : null };
        } catch {
          window.$message?.error('关联文件 JSON 格式不正确');
          return;
        }
      } else if (field === 'codeHints') {
        updateData = { codeHints: editingValue.value };
      } else {
        return;
      }
      console.log('[saveEdit] 调用 updateSubtask API', {
        taskId: taskId.value,
        subtaskId: editingSubtaskId.value,
        updateData
      });
      const success = await taskStore.updateSubtask(taskId.value, editingSubtaskId.value, updateData);
      console.log('[saveEdit] updateSubtask 结果', success);
      if (success) {
        // 先取消编辑状态，再切换预览模式
        cancelEdit();
        // 切回预览模式
        if (field === 'description' || field === 'details') {
          togglePreview('subtask', field, editingSubtaskId.value);
        }
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
const regeneratingSubtaskIds = ref<Set<number>>(new Set());
const deletingSubtaskIds = ref<Set<number>>(new Set());

// 轮询相关状态
const pollingTimer = ref<ReturnType<typeof setInterval> | null>(null);
const pollingMessageId = ref<number | null>(null);

// 拆分子任务 - 异步版本
async function handleExpandTask() {
  if (!taskId.value) return;

  // 检查任务是否已经在拆分中（除非已超时）
  if (task.value?.isExpanding && !isExpandTimeout.value) {
    window.$message?.info('任务正在拆分中，完成后会通知您');
    return;
  }

  expandLoading.value = true;
  expandStartTime.value = Date.now(); // 记录开始时间
  const messageId = await taskStore.expandTaskAsync(taskId.value);
  if (messageId) {
    window.$message?.success('拆分任务已提交，正在后台处理，完成后会通知您');
    // 不再跳转到消息列表，保持当前页面
    pollingMessageId.value = messageId;
    startPolling();
  } else {
    window.$message?.error('拆分任务失败');
    expandLoading.value = false;
    expandStartTime.value = null;
  }
}

// 质量评分
async function handleScoreTask() {
  if (!taskId.value) return;

  scoreLoading.value = true;
  try {
    const score = await taskStore.scoreTask(taskId.value);
    if (score) {
      window.$message?.success(`评分完成！总分：${score.totalScore}`);
      // 打开评分历史抽屉
      showScoreDrawer.value = true;
    }
  } catch (err) {
    window.$message?.error('评分失败');
  } finally {
    scoreLoading.value = false;
  }
}

// 打开评分历史
function openScoreHistory() {
  showScoreDrawer.value = true;
}

// 开始轮询消息状态
function startPollingMessageStatus() {
  const interval = 5000; // 5秒

  pollingTimer.value = setInterval(async () => {
    if (!pollingMessageId.value) return;

    try {
      await messageStore.loadMessages({ taskId: taskId.value });
      const message = messageStore.messages.find(m => m.id === pollingMessageId.value);
      if (message) {
        if (message.status === 'success') {
          // 拆分成功，刷新任务
          await taskStore.loadTaskDetail(taskId.value);
          window.$message?.success('子任务拆分完成');
          stopPolling();
        } else if (message.status === 'failed') {
          // 拆分失败
          window.$message?.error(message.errorMessage || '拆分失败');
          stopPolling();
        }
      }
    } catch (error) {
      console.error('Failed to poll message status:', error);
    }
  }, interval);
}

// 别名
const startPolling = startPollingMessageStatus;

// 停止轮询
function stopPolling() {
  if (pollingTimer.value) {
    clearInterval(pollingTimer.value);
    pollingTimer.value = null;
    pollingMessageId.value = null;
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

  deletingSubtaskIds.value.add(subtaskId);
  try {
    await taskStore.deleteSubtask(taskId.value, subtaskId);
  } finally {
    deletingSubtaskIds.value.delete(subtaskId);
  }
}

// 返回列表页面
function goBackToList() {
  router.replace('/requirement/task-list');
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

  const subtaskId = regeneratingSubtaskId.value;
  regeneratingSubtaskIds.value.add(subtaskId);

  try {
    await taskStore.regenerateSubtask(taskId.value, subtaskId, regeneratePrompt.value || undefined);
  } finally {
    regeneratingSubtaskIds.value.delete(subtaskId);
  }

  handleCloseRegenerateModal();
}

// 子任务列表 - 使用 ref 配合智能同步
const localSubtasks = ref<Subtask[]>([]);

// 监听 task.subtasks 变化
watch(
  () => task.value?.subtasks,
  newSubtasks => {
    if (!newSubtasks) {
      localSubtasks.value = [];
      return;
    }
    // 直接替换整个数组，确保响应式更新
    localSubtasks.value = [...newSubtasks];
  },
  { immediate: true, deep: true }
);

async function handleDragEnd() {
  if (!taskId.value || localSubtasks.value.length === 0) return;

  const subtaskIds = localSubtasks.value.map(st => st.id);
  await taskStore.reorderSubtasks(taskId.value, subtaskIds);
}

// 加载需求下的所有任务
async function loadRequirementTasks() {
  if (!task.value?.requirementId) {
    requirementTaskList.value = [];
    return;
  }

  requirementTaskLoading.value = true;
  try {
    const { data } = await fetchTaskList({ requirementId: Number(task.value.requirementId) });
    if (data) {
      const responseData = data.data || data;
      if (responseData && Array.isArray(responseData.list || responseData)) {
        requirementTaskList.value = responseData.list || responseData;
      }
    }
  } catch (error) {
    console.error('Failed to load requirement tasks:', error);
  } finally {
    requirementTaskLoading.value = false;
  }
}

// 监听任务变化，加载对应的需求任务列表
watch(
  () => task.value?.requirementId,
  newRequirementId => {
    if (newRequirementId) {
      loadRequirementTasks();
    } else {
      requirementTaskList.value = [];
    }
  }
);

// 处理下拉选择
function handleTaskSelect(key: string | number) {
  router.push(`/requirement/task-detail/${key}`);
}

onMounted(async () => {
  if (taskId.value) {
    await taskStore.loadTaskDetail(taskId.value);
    await taskStore.loadTaskDependencies();
  }
  // 加载成员列表以便获取当前用户信息
  if (memberStore.members.length === 0) {
    await memberStore.loadMembers();
  }
});

onUnmounted(() => {
  taskStore.clearCurrentTask();
  stopPolling();
});
</script>

<template>
  <div class="task-detail-page">
    <!-- 顶部操作栏 -->
    <div class="task-toolbar">
      <NSpace>
        <NButton secondary @click="goBackToList">返回</NButton>
        <!-- 拆分子任务按钮 - 根据状态显示不同文案 -->
        <NButton
          type="primary"
          :loading="actualExpandLoading"
          :disabled="actualExpandDisabled"
          @click="handleExpandTask"
        >
          {{ isExpandTimeout ? '重新拆分' : task?.isExpanding ? '拆分中...' : '拆分子任务' }}
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
        <NButton type="error" :loading="deleteLoading" @click="handleDeleteTask">删除任务</NButton>
        <NDivider vertical />
        <NButton type="info" :loading="scoreLoading" @click="handleScoreTask">质量评分</NButton>
        <NButton @click="openScoreHistory">评分历史</NButton>
      </NSpace>
      <!-- 右侧：需求任务列表下拉框 -->
      <div v-if="requirementTaskList.length > 0" class="task-dropdown-wrapper">
        <NDropdown
          :options="taskDropdownOptions"
          :show="showTaskDropdown"
          :selected-keys="[task?.id]"
          placement="bottom-end"
          @select="handleTaskSelect"
          @update:show="showTaskDropdown = $event"
        >
          <NButton class="task-dropdown-btn">
            <template #icon>
              <SvgIcon icon="mdi:list-box-outline" :size="18" />
            </template>
            <span class="task-dropdown-label">{{ currentTaskLabel }}</span>
            <SvgIcon icon="mdi:chevron-down" :size="16" class="arrow-icon" />
          </NButton>
        </NDropdown>
      </div>
    </div>

    <!-- 主体内容区域：左右分栏布局 -->
    <div class="task-content-wrapper">
      <NSpin :show="taskStore.loading">
        <div v-if="task" class="task-layout">
          <!-- 左侧：任务详情（70%） -->
          <div class="task-main">
            <NCard>
              <template #header>
                <div v-if="editingType === 'task' && editingField === 'title'" class="edit-field">
                  <NInput v-model:value="editingValue" placeholder="请输入任务标题" style="width: 100%" />
                  <NSpace class="mt-8px">
                    <NButton type="primary" size="small" :loading="saveLoading" @click="saveEdit">保存</NButton>
                    <NButton size="small" @click="cancelEdit">取消</NButton>
                  </NSpace>
                </div>
                <div v-else class="editable-title" @click="startEditTask('title')">
                  {{ task.titleTrans || task.title }}
                  <NButton text type="primary" size="small" class="edit-btn">编辑</NButton>
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

                <NDescriptions bordered :column="1" label-placement="left" :label-style="{ width: '15%' }">
                  <NDescriptionsItem label="ID">{{ task.id }}</NDescriptionsItem>
                  <NDescriptionsItem label="描述">
                    {{ task.descriptionTrans || task.description || '-' }}
                  </NDescriptionsItem>
                  <NDescriptionsItem label="依赖">
                    <template v-if="task.dependencies?.length">
                      <NSpace wrap>
                        <template v-for="d in task.dependencies" :key="d.id">
                          <NButton
                            v-if="d.dependsOnTask"
                            text
                            type="primary"
                            @click="router.push(`/requirement/task-detail/${d.dependsOnTask.id}`)"
                          >
                            {{ d.dependsOnTask.title }} (ID: {{ d.dependsOnTask.id }})
                          </NButton>
                          <NTag v-else type="info">ID: {{ d.dependsOnTaskId }}</NTag>
                        </template>
                      </NSpace>
                    </template>
                    <span v-else class="text-gray-400">-</span>
                  </NDescriptionsItem>
                  <NDescriptionsItem label="模块归属">
                    <NTag v-if="task.module" type="info">{{ task.module }}</NTag>
                    <span v-else class="text-gray-400">-</span>
                  </NDescriptionsItem>
                  <NDescriptionsItem label="输入依赖">
                    {{ task.input || '-' }}
                  </NDescriptionsItem>
                  <NDescriptionsItem label="输出交付物">
                    {{ task.output || '-' }}
                  </NDescriptionsItem>
                  <NDescriptionsItem label="风险点">
                    <NAlert v-if="task.risk" type="warning" style="margin-top: 8px">
                      {{ task.risk }}
                    </NAlert>
                    <span v-else class="text-gray-400">-</span>
                  </NDescriptionsItem>
                  <NDescriptionsItem label="验收标准">
                    <div v-if="task.acceptanceCriteria">
                      <NCollapse>
                        <NCollapseItem
                          v-for="(item, index) in parseJsonField(task.acceptanceCriteria)"
                          :key="index"
                          :title="`验收点 ${index + 1}: ${item?.description || item || '-'}`"
                          :name="index"
                        >
                          {{ item?.description || item || '-' }}
                        </NCollapseItem>
                      </NCollapse>
                    </div>
                    <span v-else class="text-gray-400">-</span>
                  </NDescriptionsItem>
                  <NDescriptionsItem label="预估工时">
                    {{ task.estimatedHours ? `${task.estimatedHours} 小时` : '-' }}
                  </NDescriptionsItem>
                  <NDescriptionsItem label="实现细节">
                    <div class="markdown-field">
                      <!-- 编辑模式 -->
                      <div v-if="!isPreviewMode('task', 'details')" class="markdown-editor-wrapper">
                        <MdEditor
                          v-model="editingValue"
                          language="zh-CN"
                          placeholder="请输入实现细节，支持 Markdown 格式"
                          :style="{ height: '300px' }"
                        />
                        <NSpace class="mt-8px">
                          <NButton type="primary" size="small" :loading="saveLoading" @click="saveEdit">保存</NButton>
                          <NButton size="small" @click="cancelEdit">取消</NButton>
                        </NSpace>
                      </div>
                      <!-- 预览模式 -->
                      <div v-else class="markdown-preview-container editable-field" @click="startEditTask('details')">
                        <MdPreview
                          v-if="task.detailsTrans || task.details"
                          :model-value="task.detailsTrans || task.details || ''"
                          class="markdown-preview-wrapper"
                        />
                        <span v-else class="text-gray-400">点击编辑实现细节</span>
                        <NButton text type="primary" size="small" class="edit-btn">编辑</NButton>
                      </div>
                    </div>
                  </NDescriptionsItem>
                  <NDescriptionsItem label="测试策略">
                    <div class="markdown-field">
                      <!-- 编辑模式 -->
                      <div v-if="!isPreviewMode('task', 'testStrategy')" class="markdown-editor-wrapper">
                        <MdEditor
                          v-model="editingValue"
                          language="zh-CN"
                          placeholder="请输入测试策略，支持 Markdown 格式"
                          :style="{ height: '300px' }"
                        />
                        <NSpace class="mt-8px">
                          <NButton type="primary" size="small" :loading="saveLoading" @click="saveEdit">保存</NButton>
                          <NButton size="small" @click="cancelEdit">取消</NButton>
                        </NSpace>
                      </div>
                      <!-- 预览模式 -->
                      <div
                        v-else
                        class="markdown-preview-container editable-field"
                        @click="startEditTask('testStrategy')"
                      >
                        <MdPreview
                          v-if="task.testStrategyTrans || task.testStrategy"
                          :model-value="task.testStrategyTrans || task.testStrategy || ''"
                          class="markdown-preview-wrapper"
                        />
                        <span v-else class="text-gray-400">点击编辑测试策略</span>
                        <NButton text type="primary" size="small" class="edit-btn">编辑</NButton>
                      </div>
                    </div>
                  </NDescriptionsItem>
                  <NDescriptionsItem v-if="task.customFields" label="自定义字段">
                    <div v-if="parseJsonField<Record<string, any>>(task.customFields)" class="custom-fields">
                      <div
                        v-for="(value, key) in parseJsonField<Record<string, any>>(task.customFields)"
                        :key="key"
                        class="custom-field-row"
                      >
                        <span class="custom-field-label">{{ formatFieldName(key) }}:</span>
                        <span class="custom-field-value">
                          <template v-if="typeof value === 'object' && value !== null">
                            <NCode :code="JSON.stringify(value, null, 2)" language="json" />
                          </template>
                          <template v-else-if="Array.isArray(value)">
                            <NSpace>
                              <NTag v-for="(item, idx) in value" :key="idx" type="info" size="small">
                                {{ typeof item === 'object' ? JSON.stringify(item) : item }}
                              </NTag>
                            </NSpace>
                          </template>
                          <template v-else>
                            {{ value }}
                          </template>
                        </span>
                      </div>
                    </div>
                  </NDescriptionsItem>
                </NDescriptions>

                <!-- 子任务列表 -->
                <NCard title="子任务" size="small">
                  <VueDraggable
                    v-if="localSubtasks.length"
                    v-model="localSubtasks"
                    :animation="150"
                    handle=".drag-handle"
                    class="subtask-list"
                    @end="handleDragEnd"
                  >
                    <div v-for="subtask in localSubtasks" :key="subtask.id" class="subtask-item">
                      <!-- 子任务头部 -->
                      <div class="subtask-header">
                        <span class="drag-handle">
                          <SvgIcon icon="mdi:drag-vertical" class="text-16px" style="color: #9ca3af" />
                        </span>
                        <div
                          v-if="
                            editingType === 'subtask' && editingSubtaskId === subtask.id && editingField === 'title'
                          "
                          class="edit-field-inline"
                        >
                          <NInput v-model:value="editingValue" placeholder="请输入子任务标题" style="width: 300px" />
                          <NButton type="primary" size="tiny" :loading="saveLoading" @click="saveEdit">保存</NButton>
                          <NButton size="tiny" @click="cancelEdit">取消</NButton>
                        </div>
                        <div
                          v-else
                          class="editable-inline subtask-title-row"
                          @click="startEditSubtask(subtask.id, 'title')"
                        >
                          <NTag :type="priorityColorMap[subtask.priority] || 'default'" size="small">
                            {{ priorityTextMap[subtask.priority] || '中' }}
                          </NTag>
                          <span class="subtask-title-text">
                            {{ task.id }}.{{ subtask.id }} {{ subtask.titleTrans || subtask.title }}
                          </span>
                          <NButton text type="primary" size="tiny" class="edit-btn-inline">编辑</NButton>
                        </div>
                        <NButton text size="small" class="expand-btn" @click="toggleSubtaskExpand(subtask.id)">
                          <SvgIcon
                            :icon="isSubtaskExpanded(subtask.id) ? 'mdi:chevron-up' : 'mdi:chevron-down'"
                            :size="16"
                            :style="{ color: isSubtaskExpanded(subtask.id) ? '#1890ff' : '#9ca3af' }"
                          />
                        </NButton>
                      </div>

                      <!-- 子任务描述（基础信息） -->
                      <div class="subtask-content">
                        <div class="markdown-field">
                          <!-- 编辑模式 -->
                          <div
                            v-if="!isPreviewMode('subtask', 'description', subtask.id)"
                            class="markdown-editor-wrapper"
                          >
                            <MdEditor
                              v-model="editingValue"
                              language="zh-CN"
                              placeholder="请输入子任务描述，支持 Markdown 格式"
                              :style="{ height: '200px' }"
                            />
                            <NSpace class="mt-8px">
                              <NButton type="primary" size="small" :loading="saveLoading" @click="saveEdit">
                                保存
                              </NButton>
                              <NButton size="small" @click="cancelEdit">取消</NButton>
                            </NSpace>
                          </div>
                          <!-- 预览模式 -->
                          <div
                            v-else
                            class="markdown-preview-container editable-field"
                            @click="startEditSubtask(subtask.id, 'description')"
                          >
                            <MdPreview
                              v-if="subtask.descriptionTrans || subtask.description"
                              :model-value="subtask.descriptionTrans || subtask.description || ''"
                              class="markdown-preview-wrapper"
                            />
                            <span v-else class="text-gray-400">点击编辑子任务描述</span>
                            <NButton text type="primary" size="small" class="edit-btn">编辑</NButton>
                          </div>
                        </div>
                      </div>

                      <!-- 展开��详细信息 -->
                      <div v-if="isSubtaskExpanded(subtask.id)" class="subtask-details">
                        <NDivider style="margin: 8px 0 12px">详细信息</NDivider>

                        <!-- 实现细节 -->
                        <div class="detail-section">
                          <div class="detail-header">
                            <div class="detail-label">实现细节</div>
                            <NButton
                              v-if="
                                !(
                                  editingType === 'subtask' &&
                                  editingSubtaskId === subtask.id &&
                                  editingField === 'details'
                                )
                              "
                              text
                              type="primary"
                              size="tiny"
                              @click="startEditSubtask(subtask.id, 'details')"
                            >
                              编辑
                            </NButton>
                          </div>
                          <div
                            v-if="
                              editingType === 'subtask' && editingSubtaskId === subtask.id && editingField === 'details'
                            "
                            class="detail-content"
                          >
                            <MdEditor
                              v-model="editingValue"
                              language="zh-CN"
                              placeholder="请输入实现细节，支持 Markdown 格式"
                              :style="{ height: '200px' }"
                            />
                            <NSpace class="mt-8px">
                              <NButton type="primary" size="small" :loading="saveLoading" @click="saveEdit">
                                保存
                              </NButton>
                              <NButton size="small" @click="cancelEdit">取消</NButton>
                            </NSpace>
                          </div>
                          <div v-else-if="subtask.details || subtask.detailsTrans" class="detail-content">
                            <MdPreview
                              :model-value="subtask.detailsTrans || subtask.details || ''"
                              class="markdown-preview-wrapper"
                            />
                          </div>
                          <div v-else class="detail-content text-gray-400">点击编辑实现细节</div>
                        </div>

                        <!-- 代码接口 -->
                        <div class="detail-section">
                          <div class="detail-header">
                            <div class="detail-label">代码接口</div>
                            <NButton
                              v-if="
                                !(
                                  editingType === 'subtask' &&
                                  editingSubtaskId === subtask.id &&
                                  editingField === 'codeInterface'
                                )
                              "
                              text
                              type="primary"
                              size="tiny"
                              @click="startEditSubtask(subtask.id, 'codeInterface')"
                            >
                              编辑
                            </NButton>
                          </div>
                          <div
                            v-if="
                              editingType === 'subtask' &&
                              editingSubtaskId === subtask.id &&
                              editingField === 'codeInterface'
                            "
                            class="detail-content"
                          >
                            <NInput
                              v-model:value="editingValue"
                              type="textarea"
                              placeholder='JSON 格式，例如: {"name": "funcName", "inputs": "string", "outputs": "boolean"}'
                              :rows="6"
                            />
                            <NSpace class="mt-8px">
                              <NButton type="primary" size="small" :loading="saveLoading" @click="saveEdit">
                                保存
                              </NButton>
                              <NButton size="small" @click="cancelEdit">取消</NButton>
                            </NSpace>
                          </div>
                          <div v-else-if="subtask.codeInterface" class="detail-content code-interface">
                            <template
                              v-if="
                                parseJsonField<{ name: string; inputs: string; outputs: string; example: string }>(
                                  subtask.codeInterface
                                )
                              "
                            >
                              <div class="interface-name">
                                <NTag type="info" size="small">函数名</NTag>
                                <NCode
                                  :code="
                                    parseJsonField<{ name: string; inputs: string; outputs: string; example: string }>(
                                      subtask.codeInterface
                                    )!.name
                                  "
                                  language="typescript"
                                />
                              </div>
                              <div class="interface-row">
                                <span class="interface-label">输入:</span>
                                <NCode
                                  :code="
                                    parseJsonField<{ name: string; inputs: string; outputs: string; example: string }>(
                                      subtask.codeInterface
                                    )!.inputs || ''
                                  "
                                  language="typescript"
                                />
                              </div>
                              <div class="interface-row">
                                <span class="interface-label">输出:</span>
                                <NCode
                                  :code="
                                    parseJsonField<{ name: string; inputs: string; outputs: string; example: string }>(
                                      subtask.codeInterface
                                    )!.outputs || ''
                                  "
                                  language="typescript"
                                />
                              </div>
                              <div
                                v-if="
                                  parseJsonField<{ name: string; inputs: string; outputs: string; example: string }>(
                                    subtask.codeInterface
                                  )!.example
                                "
                                class="interface-row"
                              >
                                <span class="interface-label">示例:</span>
                                <NCode
                                  :code="
                                    parseJsonField<{ name: string; inputs: string; outputs: string; example: string }>(
                                      subtask.codeInterface
                                    )!.example
                                  "
                                  language="typescript"
                                />
                              </div>
                            </template>
                          </div>
                          <div v-else class="detail-content text-gray-400">点击编辑代码接口</div>
                        </div>

                        <!-- 验收标准 -->
                        <div class="detail-section">
                          <div class="detail-header">
                            <div class="detail-label">验收标准</div>
                            <NButton
                              v-if="
                                !(
                                  editingType === 'subtask' &&
                                  editingSubtaskId === subtask.id &&
                                  editingField === 'acceptanceCriteria'
                                )
                              "
                              text
                              type="primary"
                              size="tiny"
                              @click="startEditSubtask(subtask.id, 'acceptanceCriteria')"
                            >
                              编辑
                            </NButton>
                          </div>
                          <div
                            v-if="
                              editingType === 'subtask' &&
                              editingSubtaskId === subtask.id &&
                              editingField === 'acceptanceCriteria'
                            "
                            class="detail-content"
                          >
                            <NInput
                              v-model:value="editingValue"
                              type="textarea"
                              placeholder='JSON 数组格式，例如: [{"id": 1, "description": "功能正常", "completed": false}]'
                              :rows="6"
                            />
                            <NSpace class="mt-8px">
                              <NButton type="primary" size="small" :loading="saveLoading" @click="saveEdit">
                                保存
                              </NButton>
                              <NButton size="small" @click="cancelEdit">取消</NButton>
                            </NSpace>
                          </div>
                          <div v-else-if="subtask.acceptanceCriteria" class="detail-content acceptance-criteria">
                            <template
                              v-if="
                                parseJsonField<Array<{ id: number; description: string; completed: boolean }>>(
                                  subtask.acceptanceCriteria
                                )
                              "
                            >
                              <div
                                v-for="criteria in parseJsonField<
                                  Array<{ id: number; description: string; completed: boolean }>
                                >(subtask.acceptanceCriteria)"
                                :key="criteria.id"
                                class="criteria-item"
                              >
                                <NCheckbox :checked="criteria.completed" disabled>
                                  {{ criteria.description }}
                                </NCheckbox>
                              </div>
                            </template>
                          </div>
                          <div v-else class="detail-content text-gray-400">点击编辑验收标准</div>
                        </div>

                        <!-- 关联文件 -->
                        <div class="detail-section">
                          <div class="detail-header">
                            <div class="detail-label">关联文件</div>
                            <NButton
                              v-if="
                                !(
                                  editingType === 'subtask' &&
                                  editingSubtaskId === subtask.id &&
                                  editingField === 'relatedFiles'
                                )
                              "
                              text
                              type="primary"
                              size="tiny"
                              @click="startEditSubtask(subtask.id, 'relatedFiles')"
                            >
                              编辑
                            </NButton>
                          </div>
                          <div
                            v-if="
                              editingType === 'subtask' &&
                              editingSubtaskId === subtask.id &&
                              editingField === 'relatedFiles'
                            "
                            class="detail-content"
                          >
                            <NInput
                              v-model:value="editingValue"
                              type="textarea"
                              placeholder='JSON 数组格式，例如: ["src/main.ts", "src/utils.ts"]'
                              :rows="4"
                            />
                            <NSpace class="mt-8px">
                              <NButton type="primary" size="small" :loading="saveLoading" @click="saveEdit">
                                保存
                              </NButton>
                              <NButton size="small" @click="cancelEdit">取消</NButton>
                            </NSpace>
                          </div>
                          <div v-else-if="subtask.relatedFiles" class="detail-content">
                            <NSpace>
                              <NTag
                                v-for="(file, idx) in parseJsonField<string[]>(subtask.relatedFiles)"
                                :key="idx"
                                type="default"
                                size="small"
                                :bordered="false"
                              >
                                {{ file }}
                              </NTag>
                            </NSpace>
                          </div>
                          <div v-else class="detail-content text-gray-400">点击编辑关联文件</div>
                        </div>

                        <!-- 代码提示 -->
                        <div class="detail-section">
                          <div class="detail-header">
                            <div class="detail-label">代码提示</div>
                            <NButton
                              v-if="
                                !(
                                  editingType === 'subtask' &&
                                  editingSubtaskId === subtask.id &&
                                  editingField === 'codeHints'
                                )
                              "
                              text
                              type="primary"
                              size="tiny"
                              @click="startEditSubtask(subtask.id, 'codeHints')"
                            >
                              编辑
                            </NButton>
                          </div>
                          <div
                            v-if="
                              editingType === 'subtask' &&
                              editingSubtaskId === subtask.id &&
                              editingField === 'codeHints'
                            "
                            class="detail-content"
                          >
                            <NInput
                              v-model:value="editingValue"
                              type="textarea"
                              placeholder="请输入代码提示"
                              :rows="4"
                            />
                            <NSpace class="mt-8px">
                              <NButton type="primary" size="small" :loading="saveLoading" @click="saveEdit">
                                保存
                              </NButton>
                              <NButton size="small" @click="cancelEdit">取消</NButton>
                            </NSpace>
                          </div>
                          <div v-else-if="subtask.codeHints" class="detail-content code-hints">
                            {{ subtask.codeHints }}
                          </div>
                          <div v-else class="detail-content text-gray-400">点击编辑代码提示</div>
                        </div>

                        <!-- 自定义字段（来自项目模板） -->
                        <div class="detail-section">
                          <div class="detail-header">
                            <div class="detail-label">自定义字段</div>
                          </div>
                          <div v-if="subtask.customFields" class="detail-content custom-fields">
                            <template v-if="parseJsonField<Record<string, any>>(subtask.customFields)">
                              <div
                                v-for="(value, key) in parseJsonField<Record<string, any>>(subtask.customFields)"
                                :key="key"
                                class="custom-field-row"
                              >
                                <span class="custom-field-label">{{ formatFieldName(key) }}:</span>
                                <span class="custom-field-value">
                                  <template v-if="typeof value === 'object' && value !== null">
                                    <NCode :code="JSON.stringify(value, null, 2)" language="json" />
                                  </template>
                                  <template v-else-if="Array.isArray(value)">
                                    <NSpace>
                                      <NTag v-for="(item, idx) in value" :key="idx" type="info" size="small">
                                        {{ typeof item === 'object' ? JSON.stringify(item) : item }}
                                      </NTag>
                                    </NSpace>
                                  </template>
                                  <template v-else>
                                    {{ value }}
                                  </template>
                                </span>
                              </div>
                            </template>
                          </div>
                          <div v-else class="detail-content text-gray-400">无自定义字段</div>
                        </div>
                      </div>

                      <!-- 子任务底部操作栏 -->
                      <div class="subtask-footer">
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
                            :loading="regeneratingSubtaskIds.has(subtask.id)"
                            @click="handleOpenRegenerateModal(subtask.id)"
                          >
                            重写
                          </NButton>
                          <NButton
                            type="error"
                            size="small"
                            ghost
                            :loading="deletingSubtaskIds.has(subtask.id)"
                            @click="handleDeleteSubtask(subtask.id)"
                          >
                            删除
                          </NButton>
                        </NSpace>
                      </div>
                    </div>
                  </VueDraggable>
                  <NEmpty v-else description="暂无子任务" />
                </NCard>
              </NSpace>
            </NCard>
          </div>

          <!-- 右侧边栏 -->
          <div class="task-sidebar">
            <NCard class="sidebar-card">
              <NTabs type="line" justify-content="space-evenly">
                <NTabPane name="assignments" tab="分配">
                  <div class="sidebar-content">
                    <AssignmentPanel :task-id="taskId!" @assigned="() => {}" @unassigned="() => {}" />
                  </div>
                </NTabPane>
                <NTabPane name="dependencies" tab="依赖">
                  <div class="sidebar-content">
                    <DependencyGraph :task-id="taskId" :height="300" />
                  </div>
                </NTabPane>
                <NTabPane name="activities" tab="活动">
                  <div class="sidebar-content">
                    <ActivityTimeline :task-id="taskId!" />
                  </div>
                </NTabPane>
                <NTabPane name="comments" tab="评论">
                  <div class="sidebar-content">
                    <CommentSection
                      :task-id="taskId!"
                      :member-id="currentMemberId"
                      @commented="() => {}"
                      @deleted="() => {}"
                    />
                  </div>
                </NTabPane>
              </NTabs>
            </NCard>
          </div>
        </div>

        <!-- 任务不存在时显示 -->
        <NEmpty v-if="!task && !taskStore.loading" description="任务不存在" />
      </NSpin>

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

      <!-- 评分历史抽屉 -->
      <ScoreHistoryDrawer :task-id="taskId" :show="showScoreDrawer" @update:show="showScoreDrawer = $event" />
    </div>
  </div>
</template>

<style scoped lang="scss">
.task-detail-page {
  padding: 16px;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.task-toolbar {
  margin-bottom: 16px;
  flex-shrink: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.task-dropdown-wrapper {
  margin-left: auto;
}

.task-dropdown-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid #e5e7eb;
  background-color: #fff;
  border-radius: 6px;
  font-size: 13px;
  max-width: 300px;

  &:hover {
    background-color: #f9fafb;
    border-color: #d1d5db;
  }

  .task-dropdown-label {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: #374151;
  }

  .arrow-icon {
    color: #9ca3af;
    transition: transform 0.2s;
  }
}

.task-content-wrapper {
  flex: 1;
  overflow: auto;
}

.task-layout {
  display: flex;
  gap: 16px;
  min-height: 100%;
}

.task-main {
  flex: 0 0 70%;
  min-width: 0;
}

.task-sidebar {
  flex: 0 0 calc(30% - 16px);
  min-width: 300px;
}

.sidebar-card {
  height: 100%;

  :deep(.n-card__content) {
    padding: 0;
  }
}

.sidebar-content {
  padding: 16px;
  max-height: calc(100vh - 200px);
  overflow-y: auto;
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

.text-gray-400 {
  color: #9ca3af;
}

// Markdown 字段容器
.markdown-field {
  width: 100%;
}

.markdown-editor-wrapper {
  width: 100%;
}

.markdown-preview-container {
  min-height: 40px;
  position: relative;
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

// 子任务拖拽列表样式
.subtask-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.subtask-item {
  padding: 12px;
  border: 1px solid #e0e0e6;
  border-radius: 8px;
  background-color: #fff;
  transition:
    box-shadow 0.2s,
    transform 0.2s;

  &:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  }
}

// 子任务标题行
.subtask-title-row {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
}

.subtask-title-text {
  font-weight: 500;
}

// 展开按钮
.expand-btn {
  margin-left: auto;
}

// 子任务详细信息区域
.subtask-details {
  margin-top: 12px;
  padding: 12px;
  background-color: #fafafa;
  border-radius: 6px;
}

// 详细信息区块
.detail-section {
  margin-bottom: 16px;

  &:last-child {
    margin-bottom: 0;
  }
}

.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.detail-label {
  font-weight: 600;
  font-size: 13px;
  color: #666;
  padding-left: 8px;
  border-left: 3px solid #1890ff;
}

.detail-content {
  padding-left: 12px;
}

// 代码接口样式
.code-interface {
  font-family: 'Fira Code', 'Monaco', monospace;

  .interface-name {
    margin-bottom: 8px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .interface-row {
    margin-bottom: 6px;
    display: flex;
    align-items: flex-start;
    gap: 8px;

    .interface-label {
      flex-shrink: 0;
      font-weight: 500;
      color: #666;
      min-width: 50px;
    }
  }
}

// 验收标准样式
.acceptance-criteria {
  .criteria-item {
    padding: 6px 0;
    border-bottom: 1px dashed #eee;

    &:last-child {
      border-bottom: none;
    }
  }
}

// 代码提示样式
.code-hints {
  background-color: #fff8e6;
  border: 1px solid #ffe58f;
  border-radius: 4px;
  padding: 8px 12px;
  font-size: 13px;
  color: #8c6b00;
}

// 自定义字段样式
.custom-fields {
  .custom-field-row {
    margin-bottom: 12px;
    padding: 8px 12px;
    background-color: #f5f5f5;
    border-radius: 6px;
    border-left: 3px solid #722ed1;

    &:last-child {
      margin-bottom: 0;
    }

    .custom-field-label {
      font-weight: 600;
      font-size: 13px;
      color: #722ed1;
      margin-right: 8px;
      min-width: 100px;
      display: inline-block;
    }

    .custom-field-value {
      color: #333;
      font-size: 13px;

      :deep(.n-code) {
        background-color: #fff;
        border-radius: 4px;
        margin-top: 4px;
      }
    }
  }
}

.subtask-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.drag-handle {
  cursor: grab;
  color: #9ca3af;
  display: flex;
  align-items: center;
  padding: 4px;
  border-radius: 4px;
  transition:
    color 0.2s,
    background-color 0.2s;

  &:hover {
    color: #1890ff;
    background-color: rgba(24, 144, 255, 0.1);
  }

  &:active {
    cursor: grabbing;
  }

  .n-icon {
    color: inherit;
  }
}

.subtask-content {
  margin-bottom: 12px;
  padding-left: 24px;
}

.subtask-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-left: 24px;
}

/* Markdown 预览容器样式 */
.markdown-preview-wrapper {
  border: none !important;
  background: transparent !important;
}

/* Markdown 编辑器样式覆盖 */
:deep(.md-editor) {
  --md-bk-color: transparent;
}

:deep(.md-editor-toolbar-wrapper) {
  border-radius: 8px 8px 0 0;
}

:deep(.md-editor-content) {
  border-radius: 0 0 8px 8px;
}

/* 确保编辑器在卡片内显示正常 */
:deep(.md-editor) {
  border: 1px solid #e0e0e6;
  border-radius: 8px;
}
</style>
