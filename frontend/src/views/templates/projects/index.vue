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
  NAlert,
  NDrawer,
  NDrawerContent,
  NCard as NCardComp,
  useMessage
} from 'naive-ui';
import type { DataTableColumns, FormInst, FormRules } from 'naive-ui';
import { MdEditor, MdPreview } from 'md-editor-v3';
import 'md-editor-v3/lib/style.css';
import {
  fetchProjectTemplates,
  fetchProjectTemplate,
  createProjectTemplate,
  updateProjectTemplate,
  deleteProjectTemplate,
  instantiateProjectTemplate,
  scoreProjectTemplate,
  scoreProjectTemplateAsync
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

// 评分相关
const scoreLoading = ref(false);
const showScoreDrawer = ref(false);
const scoreResult = ref<any>(null);

// 编辑模式
const editingDescription = ref(false);
const tempDescription = ref('');

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
    width: 320,
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
          type: 'info',
          onClick: () => openScoreDrawer(row)
        }, () => '评分'),
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

// 评分模板（异步）
async function handleScore(template: ProjectTemplate) {
  scoreLoading.value = true;
  try {
    const { data, error } = await scoreProjectTemplateAsync(template.id);
    if (!error && data) {
      const result = data.data || data;
      message.info(result.message || '评分已开始，完成后会通知您');
      showScoreDrawer.value = false;
    }
  } catch (e) {
    message.error('启动评分失败');
  } finally {
    scoreLoading.value = false;
  }
}

// 打开评分抽屉
function openScoreDrawer(template: ProjectTemplate) {
  handleScore(template);
}

// 编辑描述
function startEditDescription(template: ProjectTemplate) {
  editingTemplate.value = template;
  tempDescription.value = template.description || '';
  editingDescription.value = true;
}

// 保存描述
async function saveDescription() {
  if (!editingTemplate.value) return;

  const { error } = await updateProjectTemplate(editingTemplate.value.id, {
    name: editingTemplate.value.name,
    description: tempDescription.value,
    category: editingTemplate.value.category,
    isPublic: editingTemplate.value.isPublic,
    tasks: editingTemplate.value.tasks || []
  });

  if (!error) {
    message.success('更新成功');
    editingDescription.value = false;
    editingTemplate.value.description = tempDescription.value;
    await loadTemplates();
  } else {
    message.error('更新失败');
  }
}

// 取消编辑
function cancelEditDescription() {
  editingDescription.value = false;
  editingTemplate.value = null;
}

// 获取评分等级
function getScoreLevel(score: number): string {
  if (score >= 90) return '优秀';
  if (score >= 70) return '良好';
  if (score >= 50) return '一般';
  if (score >= 30) return '较差';
  return '很差';
}

// 获取评分等级颜色
function getScoreLevelColor(score: number): string {
  if (score >= 90) return '#18a058';
  if (score >= 70) return '#2080f0';
  if (score >= 50) return '#f0a020';
  return '#d03050';
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
          <NDescriptionsItem label="描述" :span="2">
            <div v-if="!editingDescription" class="description-view">
              <MdPreview v-if="editingTemplate.description" :modelValue="editingTemplate.description" />
              <span v-else>-</span>
              <NButton size="small" text type="primary" @click="startEditDescription(editingTemplate)" class="edit-desc-btn">
                编辑
              </NButton>
            </div>
            <MdEditor v-else v-model="tempDescription" :language="zh_CN" :toolbars="['previewOnly', 'fullscreen']" />
          </NDescriptionsItem>
          <NDescriptionsItem v-if="editingDescription" :span="2">
            <NSpace>
              <NButton size="small" type="primary" @click="saveDescription">保存</NButton>
              <NButton size="small" @click="cancelEditDescription">取消</NButton>
            </NSpace>
          </NDescriptionsItem>
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

    <!-- 评分结果抽屉 -->
    <NDrawer v-model:show="showScoreDrawer" :width="600">
      <NDrawerContent title="项目模板评分结果" :native-scrollbar="false" closable>
        <NSpin :show="scoreLoading">
          <div v-if="scoreResult" class="score-result">
            <!-- 总分 -->
            <div class="total-score">
              <div class="score-number" :style="{ color: getScoreLevelColor(scoreResult.totalScore || 0) }">
                {{ (scoreResult.totalScore || 0).toFixed(1) }}
              </div>
              <div class="score-label">总分 / 100</div>
              <div class="score-level" :style="{ color: getScoreLevelColor(scoreResult.totalScore || 0) }">
                {{ scoreResult.level || getScoreLevel(scoreResult.totalScore || 0) }}
              </div>
            </div>

            <!-- 维度评分 -->
            <div class="dimensions" v-if="scoreResult.scores">
              <div class="dimension-item">
                <span class="dimension-name">清晰度</span>
                <NTag :type="scoreResult.scores.clarity >= 7 ? 'success' : 'warning'">
                  {{ scoreResult.scores.clarity }} / 10
                </NTag>
              </div>
              <div class="dimension-item">
                <span class="dimension-name">完整性</span>
                <NTag :type="scoreResult.scores.completeness >= 7 ? 'success' : 'warning'">
                  {{ scoreResult.scores.completeness }} / 10
                </NTag>
              </div>
              <div class="dimension-item">
                <span class="dimension-name">结构化</span>
                <NTag :type="scoreResult.scores.structure >= 7 ? 'success' : 'warning'">
                  {{ scoreResult.scores.structure }} / 10
                </NTag>
              </div>
              <div class="dimension-item">
                <span class="dimension-name">可执行性</span>
                <NTag :type="scoreResult.scores.actionability >= 7 ? 'success' : 'warning'">
                  {{ scoreResult.scores.actionability }} / 10
                </NTag>
              </div>
              <div class="dimension-item">
                <span class="dimension-name">一致性</span>
                <NTag :type="scoreResult.scores.consistency >= 7 ? 'success' : 'warning'">
                  {{ scoreResult.scores.consistency }} / 10
                </NTag>
              </div>
            </div>

            <!-- 评价内容 -->
            <template v-if="scoreResult.strengths || scoreResult.weaknesses || scoreResult.suggestions">
              <NDivider>评价内容</NDivider>
              <NAlert type="success" title="优点" class="eval-section" v-if="scoreResult.strengths">
                <ul>
                  <li v-for="(item, i) in scoreResult.strengths" :key="i">{{ item }}</li>
                </ul>
              </NAlert>
              <NAlert type="warning" title="待改进" class="eval-section" v-if="scoreResult.weaknesses">
                <ul>
                  <li v-for="(item, i) in scoreResult.weaknesses" :key="i">{{ item }}</li>
                </ul>
              </NAlert>
              <NAlert type="info" title="改进建议" class="eval-section" v-if="scoreResult.suggestions">
                <div v-for="(item, i) in scoreResult.suggestions" :key="i" class="suggestion-item">
                  <strong>{{ item.issue }}</strong>
                  <p>{{ item.suggestion }}</p>
                </div>
              </NAlert>
            </template>

            <!-- 详细分析 -->
            <template v-if="scoreResult.analysis">
              <NDivider>详细分析</NDivider>
              <div class="analysis">
                <MdPreview :modelValue="scoreResult.analysis" />
              </div>
            </template>
          </div>
          <NEmpty v-else description="暂无评分结果" />
        </NSpin>
      </NDrawerContent>
    </NDrawer>
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

.description-view {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  width: 100%;

  :deep(.md-editor-preview-wrapper) {
    flex: 1;
  }

  .edit-desc-btn {
    flex-shrink: 0;
  }
}

.score-result {
  display: flex;
  flex-direction: column;
  gap: 16px;

  .total-score {
    text-align: center;
    padding: 20px 0;
    background: var(--n-border-color-popover);
    border-radius: 8px;

    .score-number {
      font-size: 48px;
      font-weight: bold;
      line-height: 1;
    }

    .score-label {
      font-size: 12px;
      color: var(--n-text-color-3);
      margin-top: 4px;
    }

    .score-level {
      font-size: 18px;
      font-weight: 500;
      margin-top: 8px;
    }
  }

  .dimensions {
    display: flex;
    flex-direction: column;
    gap: 10px;

    .dimension-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 8px 12px;
      background: var(--n-border-color-popover);
      border-radius: 6px;

      .dimension-name {
        font-size: 14px;
        color: var(--n-text-color-2);
      }
    }
  }

  .eval-section {
    margin-bottom: 12px;

    ul {
      margin: 0;
      padding-left: 20px;
    }

    li {
      margin-bottom: 6px;
    }
  }

  .suggestion-item {
    margin-bottom: 12px;

    strong {
      display: block;
      margin-bottom: 4px;
      color: var(--n-text-color-2);
    }

    p {
      margin: 0;
      font-size: 13px;
      color: var(--n-text-color-3);
    }
  }

  .analysis {
    font-size: 14px;
    line-height: 1.8;
  }
}
</style>
