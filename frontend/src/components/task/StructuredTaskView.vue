<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import {
  NButton,
  NCard,
  NCheckbox,
  NCode,
  NCollapse,
  NCollapseItem,
  NDescriptions,
  NDescriptionsItem,
  NEmpty,
  NIcon,
  NInput,
  NPopconfirm,
  NSpace,
  NSpin,
  NTabPane,
  NTabs,
  NTag
} from 'naive-ui';
import type { AcceptanceCriteria, CodeInterface, Subtask, Task } from '@/typings/api/task';

interface Props {
  task: Task | null;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false
});

const emit = defineEmits<{
  (e: 'update:task', field: string, value: unknown): void;
  (e: 'update:subtask', subtaskId: number, field: string, value: unknown): void;
  (e: 'update:acceptanceCriteria', subtaskId: number, criteria: AcceptanceCriteria[]): void;
}>();

// 本地编辑状态
const editingCodeHints = ref(false);
const codeHintsValue = ref('');
const editingSubtaskCodeHints = ref<number | null>(null);
const subtaskCodeHintsValue = ref('');

// 开始编辑任务代码提示
function startEditCodeHints() {
  codeHintsValue.value = props.task?.codeHints || '';
  editingCodeHints.value = true;
}

// 保存任务代码提示
function saveCodeHints() {
  emit('update:task', 'codeHints', codeHintsValue.value);
  editingCodeHints.value = false;
}

// 开始编辑子任务代码提示
function startEditSubtaskCodeHints(subtaskId: number) {
  const subtask = props.task?.subtasks?.find(s => s.id === subtaskId);
  subtaskCodeHintsValue.value = subtask?.codeHints || '';
  editingSubtaskCodeHints.value = subtaskId;
}

// 保存子任务代码提示
function saveSubtaskCodeHints(subtaskId: number) {
  emit('update:subtask', subtaskId, 'codeHints', subtaskCodeHintsValue.value);
  editingSubtaskCodeHints.value = null;
}

// 更新验收标准状态
function toggleAcceptanceCriteria(subtaskId: number, criteriaId: number, completed: boolean) {
  const subtask = props.task?.subtasks?.find(s => s.id === subtaskId);
  if (!subtask?.acceptanceCriteria) return;

  const updatedCriteria = subtask.acceptanceCriteria.map(c => (c.id === criteriaId ? { ...c, completed } : c));
  emit('update:acceptanceCriteria', subtaskId, updatedCriteria);
}

// 计算子任务完成进度
const subtaskProgress = computed(() => {
  if (!props.task?.subtasks) return { completed: 0, total: 0 };
  const total = props.task.subtasks.length;
  const completed = props.task.subtasks.filter(s => s.status === 'done').length;
  return { completed, total };
});

// 获取状态图标
function getStatusIcon(status: string): string {
  switch (status) {
    case 'done':
      return '✅';
    case 'in-progress':
      return '🔄';
    case 'deferred':
      return '📌';
    default:
      return '⏱️';
  }
}
</script>

<template>
  <NSpin :show="loading">
    <NTabs type="line" animated>
      <!-- 结构化视图 -->
      <NTabPane name="structure" tab="结构化视图">
        <div class="structured-view">
          <!-- 任务层级结构 -->
          <NCard size="small" title="任务层级">
            <div class="tree-view">
              <div class="tree-node task-node">
                <div class="node-header">
                  <NIcon class="i-mdi:cube-outline" :size="18" />
                  <span class="node-title">{{ task?.titleTrans || task?.title }}</span>
                  <NTag
                    :type="task?.status === 'done' ? 'success' : task?.status === 'in-progress' ? 'info' : 'default'"
                    size="small"
                  >
                    {{ task?.status }}
                  </NTag>
                </div>

                <!-- 子任务树 -->
                <div v-if="task?.subtasks?.length" class="subtask-tree">
                  <div v-for="subtask in task.subtasks" :key="subtask.id" class="tree-node subtask-node">
                    <div class="node-header">
                      <span class="tree-line">├──</span>
                      <NIcon class="i-mdi:cube-send-outline" :size="16" />
                      <span class="node-title">
                        {{ task.id }}.{{ subtask.id }} {{ subtask.titleTrans || subtask.title }}
                      </span>
                      <span class="status-icon">{{ getStatusIcon(subtask.status) }}</span>
                    </div>

                    <!-- 关联文件 -->
                    <div v-if="subtask.relatedFiles?.length" class="node-files">
                      <div v-for="file in subtask.relatedFiles" :key="file" class="file-item">
                        <span class="tree-line">│ └──</span>
                        <NIcon class="i-mdi:file-code-outline" :size="14" />
                        <code class="file-path">{{ file }}</code>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 进度统计 -->
            <div class="progress-info">
              <span>子任务进度: {{ subtaskProgress.completed }}/{{ subtaskProgress.total }}</span>
              <div class="progress-bar">
                <div
                  class="progress-fill"
                  :style="{
                    width: `${subtaskProgress.total > 0 ? (subtaskProgress.completed / subtaskProgress.total) * 100 : 0}%`
                  }"
                />
              </div>
            </div>
          </NCard>

          <!-- 代码接口定义 -->
          <NCard size="small" title="代码接口" class="mt-16px">
            <div v-if="task?.subtasks?.some(s => s.codeInterface)" class="interfaces-list">
              <NCollapse>
                <NCollapseItem
                  v-for="subtask in task.subtasks.filter(s => s.codeInterface)"
                  :key="subtask.id"
                  :name="subtask.id"
                >
                  <template #header>
                    <NSpace align="center">
                      <code>{{ subtask.codeInterface?.name }}</code>
                      <NTag size="small" type="info">{{ task.id }}.{{ subtask.id }}</NTag>
                    </NSpace>
                  </template>
                  <NDescriptions bordered :column="1" size="small">
                    <NDescriptionsItem label="接口名称">
                      <code>{{ subtask.codeInterface?.name }}</code>
                    </NDescriptionsItem>
                    <NDescriptionsItem v-if="subtask.codeInterface?.inputs" label="输入">
                      <NCode :code="subtask.codeInterface.inputs" language="typescript" />
                    </NDescriptionsItem>
                    <NDescriptionsItem v-if="subtask.codeInterface?.outputs" label="输出">
                      <NCode :code="subtask.codeInterface.outputs" language="typescript" />
                    </NDescriptionsItem>
                    <NDescriptionsItem v-if="subtask.codeInterface?.example" label="示例">
                      <NCode :code="subtask.codeInterface.example" language="typescript" />
                    </NDescriptionsItem>
                  </NDescriptions>
                </NCollapseItem>
              </NCollapse>
            </div>
            <NEmpty v-else description="暂无接口定义" />
          </NCard>
        </div>
      </NTabPane>

      <!-- 代码提示 -->
      <NTabPane name="codehints" tab="代码提示">
        <div class="code-hints-view">
          <!-- 任务级别代码提示 -->
          <NCard size="small">
            <template #header>
              <NSpace align="center" justify="space-between">
                <span>任务代码提示</span>
                <NButton v-if="!editingCodeHints" size="small" @click="startEditCodeHints">编辑</NButton>
              </NSpace>
            </template>

            <div v-if="editingCodeHints" class="edit-section">
              <NInput
                v-model:value="codeHintsValue"
                type="textarea"
                placeholder="输入实现思路、代码片段、技术决策等..."
                :autosize="{ minRows: 5, maxRows: 20 }"
              />
              <NSpace class="mt-8px">
                <NButton type="primary" size="small" @click="saveCodeHints">保存</NButton>
                <NButton size="small" @click="editingCodeHints = false">取消</NButton>
              </NSpace>
            </div>
            <div v-else class="pre-wrap hints-content">
              {{ task?.codeHints || '暂无代码提示' }}
            </div>
          </NCard>

          <!-- 子任务代码提示 -->
          <NCard size="small" class="mt-16px">
            <template #header>子任务代码提示</template>

            <div v-if="task?.subtasks?.length" class="subtask-hints-list">
              <div v-for="subtask in task.subtasks" :key="subtask.id" class="subtask-hint-item">
                <div class="subtask-hint-header">
                  <span class="subtask-label">{{ task.id }}.{{ subtask.id }}</span>
                  <span class="subtask-title">{{ subtask.titleTrans || subtask.title }}</span>
                </div>

                <div v-if="editingSubtaskCodeHints === subtask.id" class="edit-section">
                  <NInput
                    v-model:value="subtaskCodeHintsValue"
                    type="textarea"
                    placeholder="输入子任务的实现细节..."
                    :autosize="{ minRows: 3, maxRows: 10 }"
                  />
                  <NSpace class="mt-8px">
                    <NButton type="primary" size="small" @click="saveSubtaskCodeHints(subtask.id)">保存</NButton>
                    <NButton size="small" @click="editingSubtaskCodeHints = null">取消</NButton>
                  </NSpace>
                </div>
                <div v-else class="subtask-hint-content">
                  <div class="pre-wrap">{{ subtask.codeHints || '暂无提示' }}</div>
                  <NButton text type="primary" size="small" @click="startEditSubtaskCodeHints(subtask.id)">
                    编辑
                  </NButton>
                </div>
              </div>
            </div>
            <NEmpty v-else description="暂无子任务" />
          </NCard>
        </div>
      </NTabPane>

      <!-- 验收清单 -->
      <NTabPane name="acceptance" tab="验收清单">
        <div class="acceptance-view">
          <NCard size="small">
            <div v-if="task?.subtasks?.length" class="acceptance-list">
              <div v-for="subtask in task.subtasks" :key="subtask.id" class="acceptance-item">
                <div class="acceptance-header">
                  <NSpace align="center">
                    <span class="status-icon">{{ getStatusIcon(subtask.status) }}</span>
                    <span class="subtask-label">{{ task.id }}.{{ subtask.id }}</span>
                    <span>{{ subtask.titleTrans || subtask.title }}</span>
                  </NSpace>
                </div>

                <div v-if="subtask.acceptanceCriteria?.length" class="criteria-list">
                  <div v-for="criteria in subtask.acceptanceCriteria" :key="criteria.id" class="criteria-item">
                    <NCheckbox
                      :checked="criteria.completed"
                      @update:checked="(checked: boolean) => toggleAcceptanceCriteria(subtask.id, criteria.id, checked)"
                    >
                      <span :class="{ 'criteria-completed': criteria.completed }">
                        {{ criteria.description }}
                      </span>
                    </NCheckbox>
                  </div>
                </div>
                <div v-else class="no-criteria">暂无验收标准</div>

                <!-- 验收进度 -->
                <div v-if="subtask.acceptanceCriteria?.length" class="criteria-progress">
                  <span class="text-12px text-gray-500">
                    {{ subtask.acceptanceCriteria.filter(c => c.completed).length }}/{{
                      subtask.acceptanceCriteria.length
                    }}
                    完成
                  </span>
                </div>
              </div>
            </div>
            <NEmpty v-else description="暂无子任务" />
          </NCard>
        </div>
      </NTabPane>

      <!-- 相关文件 -->
      <NTabPane name="files" tab="相关文件">
        <div class="files-view">
          <NCard size="small">
            <template #header>
              <NSpace align="center" justify="space-between">
                <span>关联源文件</span>
              </NSpace>
            </template>

            <!-- 任务级别文件 -->
            <div v-if="task?.relatedFiles?.length" class="file-section">
              <div class="section-title">任务文件</div>
              <div class="file-list">
                <div v-for="file in task.relatedFiles" :key="file" class="file-tag">
                  <NIcon class="i-mdi:file-code-outline" :size="14" />
                  <code>{{ file }}</code>
                </div>
              </div>
            </div>

            <!-- 子任务文件 -->
            <div v-if="task?.subtasks?.some(s => s.relatedFiles?.length)" class="file-section mt-16px">
              <div class="section-title">子任务文件</div>
              <div
                v-for="subtask in task.subtasks.filter(s => s.relatedFiles?.length)"
                :key="subtask.id"
                class="subtask-files"
              >
                <div class="subtask-file-header">
                  <span class="subtask-label">{{ task.id }}.{{ subtask.id }}</span>
                  <span>{{ subtask.titleTrans || subtask.title }}</span>
                </div>
                <div class="file-list">
                  <div v-for="file in subtask.relatedFiles" :key="file" class="file-tag">
                    <NIcon class="i-mdi:file-code-outline" :size="14" />
                    <code>{{ file }}</code>
                  </div>
                </div>
              </div>
            </div>

            <NEmpty
              v-if="!task?.relatedFiles?.length && !task?.subtasks?.some(s => s.relatedFiles?.length)"
              description="暂无关联文件"
            />
          </NCard>
        </div>
      </NTabPane>
    </NTabs>
  </NSpin>
</template>

<style scoped lang="scss">
.structured-view,
.code-hints-view,
.acceptance-view,
.files-view {
  padding: 16px 0;
}

.mt-16px {
  margin-top: 16px;
}

.mt-8px {
  margin-top: 8px;
}

.text-12px {
  font-size: 12px;
}

.text-gray-500 {
  color: #6b7280;
}

.pre-wrap {
  white-space: pre-wrap;
  word-break: break-word;
}

// 树形视图样式
.tree-view {
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 14px;
}

.tree-node {
  padding: 8px 0;
}

.node-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.node-title {
  font-weight: 500;
  flex: 1;
}

.tree-line {
  color: #9ca3af;
  margin-right: 8px;
}

.subtask-tree {
  margin-left: 20px;
  border-left: 1px dashed #e5e7eb;
  padding-left: 8px;
}

.subtask-node {
  .node-header {
    font-size: 13px;
  }
}

.node-files {
  margin-left: 44px;
  margin-top: 4px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  padding: 2px 0;
}

.file-path {
  background-color: #f3f4f6;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

// 进度条
.progress-info {
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
}

.progress-bar {
  flex: 1;
  height: 8px;
  background-color: #e5e7eb;
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background-color: #18a058;
  transition: width 0.3s;
}

// 代码提示编辑
.edit-section {
  width: 100%;
}

.hints-content {
  min-height: 60px;
  padding: 12px;
  background-color: #f9fafb;
  border-radius: 8px;
}

// 子任务代码提示
.subtask-hints-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.subtask-hint-item {
  padding: 12px;
  background-color: #fafafa;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
}

.subtask-hint-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.subtask-label {
  font-family: monospace;
  font-size: 12px;
  background-color: #e5e7eb;
  padding: 2px 6px;
  border-radius: 4px;
}

.subtask-title {
  font-weight: 500;
}

.subtask-hint-content {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

// 验收清单
.acceptance-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.acceptance-item {
  padding: 12px;
  background-color: #fafafa;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
}

.acceptance-header {
  margin-bottom: 8px;
}

.status-icon {
  font-size: 14px;
}

.criteria-list {
  margin-left: 24px;
  margin-top: 8px;
}

.criteria-item {
  padding: 4px 0;
}

.criteria-completed {
  text-decoration: line-through;
  color: #9ca3af;
}

.no-criteria {
  margin-left: 24px;
  color: #9ca3af;
  font-size: 13px;
}

.criteria-progress {
  margin-top: 8px;
  margin-left: 24px;
}

// 文件视图
.file-section {
  .section-title {
    font-weight: 500;
    margin-bottom: 8px;
  }
}

.file-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.file-tag {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background-color: #f3f4f6;
  border-radius: 6px;
  font-size: 13px;

  code {
    background: none;
    padding: 0;
  }
}

.subtask-files {
  margin-bottom: 12px;
}

.subtask-file-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  font-size: 13px;
}
</style>
