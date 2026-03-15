<script setup lang="ts">
import { ref, watch } from 'vue';
import { NButton, NCard, NForm, NFormItem, NInput, NModal, NSelect, NSpace, NTag } from 'naive-ui';
import { MdEditor } from 'md-editor-v3';
import 'md-editor-v3/lib/style.css';
import type { Task, TaskPriority } from '@/typings/api/task';

interface Props {
  visible: boolean;
  task: Task | null;
  priorityOptions?: Array<{ label: string; value: TaskPriority }>;
  assigneeOptions?: Array<{ label: string; value: string }>;
}

interface Emits {
  (e: 'update:visible', value: boolean): void;
  (e: 'saved', success: boolean): void;
}

const props = withDefaults(defineProps<Props>(), {
  priorityOptions: () => [
    { label: '高', value: 'high' },
    { label: '中', value: 'medium' },
    { label: '低', value: 'low' }
  ],
  assigneeOptions: () => []
});

const emit = defineEmits<Emits>();

// 表单数据
const formData = ref<Partial<Task>>({
  title: '',
  description: '',
  details: '',
  testStrategy: '',
  risk: '',
  acceptanceCriteria: '',
  priority: 'medium',
  assignee: undefined,
  startDate: undefined,
  dueDate: undefined,
  estimatedHours: undefined
});

// 加载状态
const saving = ref(false);

// 监听 visible 变化，初始化表单数据
watch(
  () => props.visible,
  newVal => {
    if (newVal && props.task) {
      formData.value = {
        title: props.task.title || '',
        description: props.task.description || '',
        details: props.task.details || '',
        testStrategy: props.task.testStrategy || '',
        risk: props.task.risk || '',
        acceptanceCriteria: props.task.acceptanceCriteria || '',
        priority: props.task.priority || 'medium',
        assignee: props.task.assignee,
        startDate: props.task.startDate,
        dueDate: props.task.dueDate,
        estimatedHours: props.task.estimatedHours
      };
    }
  },
  { immediate: true }
);

// 关闭弹框
function handleClose() {
  emit('update:visible', false);
}

// 保存
function handleSave() {
  emit('saved', true);
  emit('update:visible', false);
}
</script>

<template>
  <NModal
    :show="visible"
    preset="card"
    title="编辑任务"
    style="width: 800px; max-height: 90vh; overflow: auto"
    :close-on-esc="false"
    @update:show="value => emit('update:visible', value)"
  >
    <NForm :model="formData" label-placement="left" label-width="100px">
      <NFormItem label="任务标题">
        <NInput v-model:value="formData.title" placeholder="请输入任务标题" maxlength="200" show-count />
      </NFormItem>

      <NFormItem label="优先级">
        <NSelect v-model:value="formData.priority" :options="priorityOptions" style="width: 150px" />
      </NFormItem>

      <NFormItem label="负责人">
        <NSelect
          v-model:value="formData.assignee"
          :options="assigneeOptions"
          placeholder="选择负责人"
          style="width: 200px"
          clearable
          filterable
        />
      </NFormItem>

      <NFormItem label="开始日期">
        <NInput
          v-model:value="formData.startDate"
          type="date"
          placeholder="选择开始日期"
          style="width: 200px"
          clearable
        />
      </NFormItem>

      <NFormItem label="截止日期">
        <NInput
          v-model:value="formData.dueDate"
          type="date"
          placeholder="选择截止日期"
          style="width: 200px"
          clearable
        />
      </NFormItem>

      <NFormItem label="预估工时（小时）">
        <NInput
          v-model:value="formData.estimatedHours"
          type="number"
          placeholder="输入预估工时"
          style="width: 200px"
          :min="0"
          :step="0.5"
          clearable
        />
      </NFormItem>

      <NFormItem label="任务描述">
        <NInput
          v-model:value="formData.description"
          type="textarea"
          placeholder="请输入任务描述"
          :rows="3"
          maxlength="1000"
          show-count
        />
      </NFormItem>

      <NFormItem label="实现细节">
        <MdEditor
          v-model="formData.details"
          language="zh-CN"
          :toolbars-exclude="['save', 'catalog', 'pageFullscreen']"
          :preview="false"
          style="height: 300px"
        />
      </NFormItem>

      <NFormItem label="测试策略">
        <MdEditor
          v-model="formData.testStrategy"
          language="zh-CN"
          :toolbars-exclude="['save', 'catalog', 'pageFullscreen']"
          :preview="false"
          style="height: 200px"
        />
      </NFormItem>

      <NFormItem label="风险点">
        <NInput
          v-model:value="formData.risk"
          type="textarea"
          placeholder="请输入风险点"
          :rows="3"
          maxlength="2000"
          show-count
        />
      </NFormItem>

      <NFormItem label="验收标准">
        <NInput
          v-model:value="formData.acceptanceCriteria"
          type="textarea"
          placeholder="请输入验收标准（JSON 格式）"
          :rows="5"
          maxlength="5000"
          show-count
        />
      </NFormItem>
    </NForm>

    <template #footer>
      <NSpace justify="end">
        <NButton @click="handleClose">取消</NButton>
        <NButton type="primary" :loading="saving" @click="handleSave">保存</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped lang="scss">
:deep(.md-editor) {
  border: 1px solid #e0e0e0;
  border-radius: 4px;
}
</style>
