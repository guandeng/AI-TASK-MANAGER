<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import {
  NAlert,
  NButton,
  NCard,
  NCascader,
  NDatePicker,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSelect,
  NSpace,
  NSpin,
  NTag
} from 'naive-ui';
import type { FormInst, FormRules, SelectOption } from 'naive-ui';
import { useTaskStore } from '@/store/modules/task';
import { useRequirementStore } from '@/store/modules/requirement';
import type { TaskPriority } from '@/typings/api/task';
import TagSelector from '@/components/task/TagSelector.vue';

const router = useRouter();
const taskStore = useTaskStore();
const requirementStore = useRequirementStore();

// 表单引用
const formRef = ref<FormInst | null>(null);

// 加载状态
const submitting = ref(false);

// 表单数据
const formData = ref({
  title: '',
  description: '',
  details: '',
  testStrategy: '',
  priority: 'medium' as TaskPriority,
  assignee: undefined as string | undefined,
  requirementId: undefined as number | undefined,
  dependencies: [] as number[],
  tags: [] as string[],
  startDate: null as number | null,
  dueDate: null as number | null,
  estimatedHours: undefined as number | undefined
});

// 优先级选项
const priorityOptions = [
  { label: '高', value: 'high' },
  { label: '中', value: 'medium' },
  { label: '低', value: 'low' }
];

// 负责人选项（可以从成员列表获取）
const assigneeOptions = [
  { label: '张三', value: 'zhangsan' },
  { label: '李四', value: 'lisi' }
];

// 需求选项
const requirementOptions = computed<SelectOption[]>(() => {
  const options: SelectOption[] = [{ label: '不关联需求', value: undefined as unknown as number }];
  requirementStore.requirements.forEach(req => {
    options.push({
      label: req.title,
      value: req.id
    });
  });
  return options;
});

// 依赖任务选项（排除当前正在创建的任务）
const dependencyOptions = computed<SelectOption[]>(() => {
  const options: SelectOption[] = [];
  taskStore.tasks.forEach(task => {
    options.push({
      label: task.title,
      value: task.id,
      disabled: Boolean(formData.value.requirementId && task.requirementId === formData.value.requirementId)
    });
  });
  return options;
});

// 表单验证规则
const rules: FormRules = {
  title: {
    required: true,
    message: '请输入任务标题',
    trigger: ['blur', 'input']
  }
};

// 提交处理
async function handleSubmit() {
  try {
    await formRef.value?.validate();
    submitting.value = true;

    // 准备提交数据
    const submitData = {
      title: formData.value.title,
      description: formData.value.description || undefined,
      details: formData.value.details || undefined,
      testStrategy: formData.value.testStrategy || undefined,
      priority: formData.value.priority,
      assignee: formData.value.assignee,
      requirementId: formData.value.requirementId,
      dependencies: formData.value.dependencies,
      startDate: formData.value.startDate ? new Date(formData.value.startDate).toISOString().split('T')[0] : undefined,
      dueDate: formData.value.dueDate ? new Date(formData.value.dueDate).toISOString().split('T')[0] : undefined,
      estimatedHours: formData.value.estimatedHours
    };

    const result = await taskStore.createTask(submitData);

    if (result.success) {
      // 创建成功，返回列表页
      router.push('/requirement/task-list');
    } else {
      window.$message?.error('创建任务失败');
    }
  } catch (error) {
    console.error('Form validation failed:', error);
  } finally {
    submitting.value = false;
  }
}

// 取消处理
function handleCancel() {
  router.back();
}

// 加载依赖数据
onMounted(async () => {
  await requirementStore.loadRequirements();
  await taskStore.loadTasks({ pageSize: 100 });
});
</script>

<template>
  <div class="task-create-page">
    <NCard title="创建新任务">
      <NSpin :show="submitting">
        <NForm ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="120px">
          <NFormItem label="任务标题" path="title">
            <NInput v-model:value="formData.title" placeholder="请输入任务标题" maxlength="200" show-count clearable />
          </NFormItem>

          <NFormItem label="关联需求" path="requirementId">
            <NSelect
              v-model:value="formData.requirementId"
              :options="requirementOptions"
              placeholder="选择关联的需求（可选）"
              clearable
              filterable
            />
          </NFormItem>

          <NFormItem label="优先级" path="priority">
            <NSelect
              v-model:value="formData.priority"
              :options="priorityOptions"
              placeholder="选择优先级"
              style="width: 200px"
            />
          </NFormItem>

          <NFormItem label="负责人" path="assignee">
            <NSelect
              v-model:value="formData.assignee"
              :options="assigneeOptions"
              placeholder="选择负责人（可选）"
              style="width: 200px"
              clearable
              filterable
            />
          </NFormItem>

          <NFormItem label="依赖任务" path="dependencies">
            <NSelect
              v-model:value="formData.dependencies"
              :options="dependencyOptions"
              placeholder="选择依赖的任务（可选，多选）"
              multiple
              filterable
              style="width: 100%"
            />
          </NFormItem>

          <NFormItem label="开始日期" path="startDate">
            <NDatePicker
              v-model:value="formData.startDate"
              type="date"
              placeholder="选择开始日期"
              style="width: 200px"
              clearable
            />
          </NFormItem>

          <NFormItem label="截止日期" path="dueDate">
            <NDatePicker
              v-model:value="formData.dueDate"
              type="date"
              placeholder="选择截止日期"
              style="width: 200px"
              clearable
            />
          </NFormItem>

          <NFormItem label="预估工时（小时）" path="estimatedHours">
            <NInputNumber
              v-model:value="formData.estimatedHours"
              placeholder="输入预估工时"
              style="width: 200px"
              :min="0"
              :step="0.5"
              clearable
            />
          </NFormItem>

          <NFormItem label="任务描述" path="description">
            <NInput
              v-model:value="formData.description"
              type="textarea"
              placeholder="请输入任务描述"
              :rows="3"
              maxlength="1000"
              show-count
            />
          </NFormItem>

          <NFormItem label="实现细节" path="details">
            <NInput
              v-model:value="formData.details"
              type="textarea"
              placeholder="请输入实现细节（支持 Markdown 语法）"
              :rows="5"
              maxlength="5000"
              show-count
            />
          </NFormItem>

          <NFormItem label="测试策略" path="testStrategy">
            <NInput
              v-model:value="formData.testStrategy"
              type="textarea"
              placeholder="请输入测试策略"
              :rows="3"
              maxlength="2000"
              show-count
            />
          </NFormItem>

          <NFormItem>
            <NSpace>
              <NButton type="primary" :loading="submitting" @click="handleSubmit">创建任务</NButton>
              <NButton @click="handleCancel">取消</NButton>
            </NSpace>
          </NFormItem>
        </NForm>
      </NSpin>
    </NCard>
  </div>
</template>

<style scoped lang="scss">
.task-create-page {
  padding: 16px;
}
</style>
