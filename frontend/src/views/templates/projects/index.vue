<script setup lang="ts">
import { h, computed, onMounted, onActivated, ref } from 'vue';
import {
  NCard,
  NDataTable,
  NTag,
  NSpace,
  NButton,
  NInput,
  NSelect,
  NModal,
  NForm,
  NFormItem,
  NSwitch,
  NInputNumber,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NEmpty,
  NSpin,
  NPopconfirm,
  useMessage
} from 'naive-ui';
import type { DataTableColumns, FormInst, FormRules } from 'naive-ui';
import {
  fetchProjectTemplates,
  fetchProjectTemplate,
  createProjectTemplate,
  updateProjectTemplate,
  deleteProjectTemplate,
  instantiateProjectTemplate
} from '@/service/api/template';
import type { ProjectTemplate, CreateProjectTemplateRequest, TemplateTask } from '@/typings/api/template';
import {
  TEMPLATE_CATEGORY_OPTIONS,
  TEMPLATE_PRIORITY_LABELS,
  TEMPLATE_PRIORITY_COLORS
} from '@/typings/api/template';

const message = useMessage();

// 状态
const loading = ref(false);
const templates = ref<ProjectTemplate[]>([]);
const filterCategory = ref<string | null>(null);
const filterKeyword = ref('');

// 模态框状态
const showModal = ref(false);
const modalType = ref<'create' | 'edit' | 'view'>('create');
const editingTemplate = ref<ProjectTemplate | null>(null);
const formRef = ref<FormInst | null>(null);
const formData = ref<CreateProjectTemplateRequest>({
  name: '',
  description: '',
  category: 'other',
  isPublic: true,
  tags: [],
  tasks: []
});

// 实例化模态框
const showInstantiateModal = ref(false);
const instantiatingTemplate = ref<ProjectTemplate | null>(null);
const instantiateForm = ref({
  name: '',
  description: '',
  startDate: null as string | null,
  dueDate: null as string | null
});
const instantiateLoading = ref(false);

// 过滤后的模板列表
const filteredTemplates = computed(() => {
  return templates.value.filter(t => {
    if (filterCategory.value && t.category !== filterCategory.value) {
      return false;
    }
    if (filterKeyword.value) {
      const keyword = filterKeyword.value.toLowerCase();
      return (
        t.name.toLowerCase().includes(keyword) ||
        (t.description?.toLowerCase().includes(keyword))
      );
    }
    return true;
  });
});

// 表格列定义
const columns: DataTableColumns<ProjectTemplate> = [
  {
    title: 'ID',
    key: 'id',
    width: 60
  },
  {
    title: '模板名称',
    key: 'name',
    ellipsis: { tooltip: true }
  },
  {
    title: '分类',
    key: 'category',
    width: 100,
    render(row) {
      const option = TEMPLATE_CATEGORY_OPTIONS.find(o => o.value === row.category);
      return h(NTag, { size: 'small' }, () => option?.label || row.category);
    }
  },
  {
    title: '任务数',
    key: 'taskCount',
    width: 80,
    render(row) {
      return row.tasks?.length || 0;
    }
  },
  {
    title: '使用次数',
    key: 'usageCount',
    width: 100
  },
  {
    title: '公开',
    key: 'isPublic',
    width: 80,
    render(row) {
      return h(NTag, {
        type: row.isPublic ? 'success' : 'default',
        size: 'small'
      }, () => row.isPublic ? '是' : '否');
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 250,
    render(row) {
      return h(NSpace, {}, () => [
        h(NButton, {
          size: 'small',
          onClick: () => handleView(row)
        }, () => '查看'),
        h(NButton, {
          size: 'small',
          onClick: () => handleEdit(row)
        }, () => '编辑'),
        h(NButton, {
          size: 'small',
          type: 'primary',
          onClick: () => handleOpenInstantiate(row)
        }, () => '使用'),
        h(NPopconfirm, {
          onPositiveClick: () => handleDelete(row)
        }, {
          trigger: () => h(NButton, { size: 'small', type: 'error' }, () => '删除'),
          default: () => '确认删除此模板？'
        })
      ]);
    }
  }
];

// 表单验证规则
const formRules: FormRules = {
  name: { required: true, message: '请输入模板名称', trigger: 'blur' }
};

// 辅助函数：提取后端返回的 data 字段
// 后端返回格式: { code: 0, message: "success", data: {...} }
function extractData(responseData: any): any {
  if (!responseData) return null;
  if (responseData.data !== undefined) {
    return responseData.data;
  }
  return responseData;
}

// 加载模板列表
async function loadTemplates() {
  loading.value = true;
  try {
    const { data, error } = await fetchProjectTemplates();
    if (!error && data) {
      templates.value = extractData(data) || [];
    }
  } finally {
    loading.value = false;
  }
}

// 查看模板详情
async function handleView(template: ProjectTemplate) {
  const { data } = await fetchProjectTemplate(template.id);
  if (data) {
    editingTemplate.value = extractData(data);
    modalType.value = 'view';
    showModal.value = true;
  }
}

// 打开创建模态框
function handleCreate() {
  modalType.value = 'create';
  editingTemplate.value = null;
  formData.value = {
    name: '',
    description: '',
    category: 'other',
    isPublic: true,
    tags: [],
    tasks: []
  };
  showModal.value = true;
}

// 打开编辑模态框
async function handleEdit(template: ProjectTemplate) {
  const { data } = await fetchProjectTemplate(template.id);
  if (data) {
    const templateData = extractData(data);
    modalType.value = 'edit';
    editingTemplate.value = templateData;
    formData.value = {
      name: templateData.name,
      description: templateData.description || '',
      category: templateData.category || 'other',
      isPublic: templateData.isPublic,
      tags: templateData.tags || [],
      tasks: templateData.tasks?.map((t: any) => ({
        title: t.title,
        description: t.description || '',
        priority: t.priority,
        order: t.order,
        estimatedHours: t.estimatedHours,
        dependencies: t.dependencies,
        subtasks: t.subtasks?.map((s: any) => ({
          title: s.title,
          description: s.description || '',
          order: s.order,
          estimatedHours: s.estimatedHours
        }))
      })) || []
    };
    showModal.value = true;
  }
}

// 提交表单
async function handleSubmit() {
  try {
    await formRef.value?.validate();

    if (modalType.value === 'create') {
      const { error } = await createProjectTemplate(formData.value);
      if (error) {
        message.error('创建失败');
      } else {
        message.success('创建成功');
        showModal.value = false;
        await loadTemplates();
      }
    } else if (editingTemplate.value) {
      const { error } = await updateProjectTemplate(editingTemplate.value.id, formData.value);
      if (error) {
        message.error('更新失败');
      } else {
        message.success('更新成功');
        showModal.value = false;
        await loadTemplates();
      }
    }
  } catch (e) {
    // 验证失败
  }
}

// 删除模板
async function handleDelete(template: ProjectTemplate) {
  const { error } = await deleteProjectTemplate(template.id);
  if (error) {
    message.error('删除失败');
  } else {
    message.success('已删除');
    await loadTemplates();
  }
}

// 打开实例化模态框
function handleOpenInstantiate(template: ProjectTemplate) {
  instantiatingTemplate.value = template;
  instantiateForm.value = {
    name: `${template.name} - 副本`,
    description: template.description || '',
    startDate: null,
    dueDate: null
  };
  showInstantiateModal.value = true;
}

// 执行实例化
async function handleInstantiate() {
  if (!instantiatingTemplate.value) return;

  instantiateLoading.value = true;
  try {
    const { data, error } = await instantiateProjectTemplate(instantiatingTemplate.value.id, {
      name: instantiateForm.value.name,
      description: instantiateForm.value.description
    });

    if (error) {
      message.error('创建失败');
    } else {
      const result = extractData(data);
      message.success(`已创建项目，包含 ${result?.taskIds?.length || 0} 个任务`);
      showInstantiateModal.value = false;
      await loadTemplates();
    }
  } finally {
    instantiateLoading.value = false;
  }
}

// 生命周期
onMounted(() => {
  loadTemplates();
});

onActivated(() => {
  loadTemplates();
});
</script>

<template>
  <div class="template-page p-16px">
    <NCard title="项目模板">
      <template #header-extra>
        <NSpace>
          <NButton type="primary" @click="handleCreate">新建模板</NButton>
          <NButton @click="loadTemplates">刷新</NButton>
        </NSpace>
      </template>

      <!-- 筛选栏 -->
      <NSpace class="mb-16px">
        <NInput
          v-model:value="filterKeyword"
          placeholder="搜索模板名称"
          clearable
          style="width: 200px"
        />
        <NSelect
          v-model:value="filterCategory"
          :options="[{ label: '全部分类', value: '' }, ...TEMPLATE_CATEGORY_OPTIONS]"
          placeholder="选择分类"
          clearable
          style="width: 140px"
        />
      </NSpace>

      <!-- 数据表格 -->
      <NSpin :show="loading">
        <NDataTable
          :columns="columns"
          :data="filteredTemplates"
          :pagination="{ pageSize: 20 }"
          :row-key="(row: ProjectTemplate) => row.id"
        />
      </NSpin>
    </NCard>

    <!-- 创建/编辑/查看模态框 -->
    <NModal
      v-model:show="showModal"
      :title="modalType === 'create' ? '新建模板' : modalType === 'edit' ? '编辑模板' : '模板详情'"
      preset="card"
      style="width: 700px; max-height: 80vh; overflow-y: auto"
    >
      <!-- 查看模式 -->
      <template v-if="modalType === 'view' && editingTemplate">
        <NDescriptions label-placement="left" :column="2">
          <NDescriptionsItem label="模板名称">{{ editingTemplate.name }}</NDescriptionsItem>
          <NDescriptionsItem label="分类">
            {{ TEMPLATE_CATEGORY_OPTIONS.find(o => o.value === editingTemplate!.category)?.label }}
          </NDescriptionsItem>
          <NDescriptionsItem label="公开">
            <NTag :type="editingTemplate.isPublic ? 'success' : 'default'" size="small">
              {{ editingTemplate.isPublic ? '是' : '否' }}
            </NTag>
          </NDescriptionsItem>
          <NDescriptionsItem label="使用次数">{{ editingTemplate.usageCount }}</NDescriptionsItem>
          <NDescriptionsItem label="描述" :span="2">{{ editingTemplate.description || '-' }}</NDescriptionsItem>
        </NDescriptions>

        <NDivider>任务列表</NDivider>

        <div v-if="editingTemplate.tasks && editingTemplate.tasks.length > 0" class="task-list">
          <div v-for="task in editingTemplate.tasks" :key="task.id" class="task-item">
            <div class="task-header">
              <span class="task-title">{{ task.title }}</span>
              <NTag :type="TEMPLATE_PRIORITY_COLORS[task.priority]" size="small">
                {{ TEMPLATE_PRIORITY_LABELS[task.priority] }}
              </NTag>
            </div>
            <div v-if="task.description" class="task-desc">{{ task.description }}</div>
            <div v-if="task.estimatedHours" class="task-hours">预估: {{ task.estimatedHours }}h</div>
            <div v-if="task.subtasks && task.subtasks.length > 0" class="subtask-list">
              <div v-for="subtask in task.subtasks" :key="subtask.id" class="subtask-item">
                • {{ subtask.title }}
              </div>
            </div>
          </div>
        </div>
        <NEmpty v-else description="暂无任务" />
      </template>

      <!-- 创建/编辑模式 -->
      <template v-else>
        <NForm ref="formRef" :model="formData" :rules="formRules" label-placement="left" label-width="80">
          <NFormItem label="名称" path="name">
            <NInput v-model:value="formData.name" placeholder="请输入模板名称" />
          </NFormItem>
          <NFormItem label="描述" path="description">
            <NInput
              v-model:value="formData.description"
              type="textarea"
              placeholder="请输入描述"
              :autosize="{ minRows: 2, maxRows: 4 }"
            />
          </NFormItem>
          <NFormItem label="分类" path="category">
            <NSelect v-model:value="formData.category" :options="TEMPLATE_CATEGORY_OPTIONS" />
          </NFormItem>
          <NFormItem label="公开" path="isPublic">
            <NSwitch v-model:value="formData.isPublic" />
          </NFormItem>
        </NForm>

        <NDivider>任务配置（可选）</NDivider>

        <div class="tasks-config">
          <NEmpty description="任务配置功能开发中..." />
        </div>
      </template>

      <template v-if="modalType !== 'view'" #footer>
        <NSpace justify="end">
          <NButton @click="showModal = false">取消</NButton>
          <NButton type="primary" @click="handleSubmit">确定</NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- 实例化模态框 -->
    <NModal
      v-model:show="showInstantiateModal"
      title="使用模板创建项目"
      preset="card"
      style="width: 500px"
    >
      <NForm label-placement="left" label-width="80">
        <NFormItem label="项目名称">
          <NInput v-model:value="instantiateForm.name" placeholder="请输入项目名称" />
        </NFormItem>
        <NFormItem label="描述">
          <NInput
            v-model:value="instantiateForm.description"
            type="textarea"
            placeholder="请输入描述"
            :autosize="{ minRows: 2 }"
          />
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showInstantiateModal = false">取消</NButton>
          <NButton type="primary" :loading="instantiateLoading" @click="handleInstantiate">
            创建
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped lang="scss">
.template-page {
  height: 100%;
}

.mb-16px {
  margin-bottom: 16px;
}

.p-16px {
  padding: 16px;
}

.task-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-item {
  padding: 12px;
  background: #fafafa;
  border-radius: 6px;

  .task-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 4px;

    .task-title {
      font-weight: 500;
    }
  }

  .task-desc {
    font-size: 13px;
    color: #666;
    margin-bottom: 4px;
  }

  .task-hours {
    font-size: 12px;
    color: #999;
    margin-bottom: 4px;
  }

  .subtask-list {
    padding-left: 16px;
    margin-top: 8px;

    .subtask-item {
      font-size: 13px;
      color: #666;
      padding: 2px 0;
    }
  }
}

.tasks-config {
  padding: 16px;
  background: #fafafa;
  border-radius: 6px;
}
</style>
