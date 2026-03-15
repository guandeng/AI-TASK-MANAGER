<script setup lang="ts">
import { onMounted, onActivated, ref, computed, h, watch } from 'vue';
import { useRouter } from 'vue-router';
import { NButton, NCard, NDataTable, NGrid, NGi, NInput, NSelect, NSpace, NStatistic, NTag, NPopconfirm, NProgress, NSpin, NEmpty, NModal, NRadioGroup, NRadio, NDivider, NTabs, NTabPane, NForm, NFormItem, NInputNumber, NSwitch } from 'naive-ui';
import type { DataTableColumns, SelectOption } from 'naive-ui';
import { useRequirementStore } from '@/store/modules/requirement';
import type { Requirement, RequirementStatus, RequirementPriority } from '@/typings/api/requirement';
import { splitRequirementToTasksAsync, type TaskType } from '@/service/api/requirement';
import { fetchLanguageList, createLanguage, updateLanguage, deleteLanguage, type Language, type LanguageCategory } from '@/service/api/language';

defineOptions({
  name: 'RequirementList'
});

const router = useRouter();
const requirementStore = useRequirementStore();

// 搜索和筛选
const searchKeyword = ref('');
const filterStatus = ref<RequirementStatus | null>(null);
const filterPriority = ref<RequirementPriority | null>(null);

// 任务类型选择弹框
const showTaskTypeModal = ref(false);
const selectedTaskType = ref<TaskType>('backend');
const selectedLanguageId = ref<number | null>(null);
const currentSplittingRequirement = ref<Requirement | null>(null);

// 语言列表
const languages = ref<Language[]>([]);
const languageLoading = ref(false);

// 语言管理弹框
const showLanguageModal = ref(false);
const editingLanguage = ref<Language | null>(null);
const languageForm = ref({
  name: '',
  category: 'backend' as LanguageCategory,
  displayName: '',
  framework: '',
  description: '',
  codeHints: '',
  remark: '',
  isActive: true,
  sortOrder: 0
});

// 任务类型选项
const taskTypeOptions: SelectOption[] = [
  { label: '后端', value: 'backend' as TaskType },
  { label: '前端', value: 'frontend' as TaskType },
  { label: '前后端', value: 'fullstack' as TaskType }
];

// 语言分类选项
const categoryOptions: SelectOption[] = [
  { label: '后端', value: 'backend' },
  { label: '前端', value: 'frontend' }
];

// 状态选项
const statusOptions: SelectOption[] = [
  { label: '全部状态', value: null as unknown as string },
  { label: '草稿', value: 'draft' },
  { label: '进行中', value: 'active' },
  { label: '已完成', value: 'completed' },
  { label: '已归档', value: 'archived' }
];

// 优先级选项
const priorityOptions: SelectOption[] = [
  { label: '全部优先级', value: null as unknown as string },
  { label: '高', value: 'high' },
  { label: '中', value: 'medium' },
  { label: '低', value: 'low' }
];

// 状态标签颜色
const statusColors: Record<RequirementStatus, string> = {
  draft: 'default',
  active: 'info',
  completed: 'success',
  archived: 'warning'
};

// 状态文本
const statusText: Record<RequirementStatus, string> = {
  draft: '草稿',
  active: '进行中',
  completed: '已完成',
  archived: '已归档'
};

// 优先级标签颜色
const priorityColors: Record<RequirementPriority, string> = {
  high: 'error',
  medium: 'warning',
  low: 'default'
};

// 优先级文本
const priorityText: Record<RequirementPriority, string> = {
  high: '高',
  medium: '中',
  low: '低'
};

// 完成进度
const completionProgress = computed(() => {
  if (!requirementStore.statistics) return 0;
  const { total, completed } = requirementStore.statistics;
  if (total === 0) return 0;
  return Math.round((completed / total) * 100);
});

// 拆分任务加载状态
const splittingTaskIds = ref<Set<number>>(new Set());

// 过滤后的需求列表
const filteredRequirements = computed(() => {
  let list = [...requirementStore.requirements];

  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase();
    list = list.filter(
      req =>
        req.title.toLowerCase().includes(keyword) ||
        req.content.toLowerCase().includes(keyword)
    );
  }

  if (filterStatus.value) {
    list = list.filter(req => req.status === filterStatus.value);
  }

  if (filterPriority.value) {
    list = list.filter(req => req.priority === filterPriority.value);
  }

  return list.sort((a, b) => b.id - a.id);
});

// 根据任务类型筛选语言选项
const languageOptions = computed(() => {
  const list = (languages.value || []).filter(l => l.isActive);

  if (selectedTaskType.value === 'fullstack') {
    // 前后端都选，返回所有
    return list.map(l => ({ label: `${l.category === 'backend' ? '[后端]' : '[前端]'} ${l.displayName}`, value: l.id }));
  }

  const category: LanguageCategory = selectedTaskType.value === 'frontend' ? 'frontend' : 'backend';
  return list.filter(l => l.category === category).map(l => ({ label: l.displayName, value: l.id }));
});

// 语言表格列
const languageColumns: DataTableColumns<Language> = [
  {
    title: 'ID',
    key: 'id',
    width: 60,
    align: 'center'
  },
  {
    title: '分类',
    key: 'category',
    width: 80,
    align: 'center',
    render(row) {
      return h(NTag, { type: row.category === 'backend' ? 'info' : 'success', size: 'small' }, () => row.category === 'backend' ? '后端' : '前端');
    }
  },
  {
    title: '名称',
    key: 'name',
    width: 100
  },
  {
    title: '显示名称',
    key: 'displayName',
    width: 180
  },
  {
    title: '框架',
    key: 'framework',
    ellipsis: { tooltip: true }
  },
  {
    title: '启用',
    key: 'isActive',
    width: 70,
    align: 'center',
    render(row) {
      return h(NTag, { type: row.isActive ? 'success' : 'default', size: 'small' }, () => row.isActive ? '是' : '否');
    }
  },
  {
    title: '排序',
    key: 'sortOrder',
    width: 70,
    align: 'center'
  },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    align: 'center',
    render(row) {
      return h(NSpace, { justify: 'center' }, () => [
        h(NButton, { size: 'small', type: 'primary', text: true, onClick: () => openEditLanguage(row) }, () => '编辑'),
        h(NPopconfirm, { onPositiveClick: () => handleDeleteLanguage(row.id) }, {
          trigger: () => h(NButton, { size: 'small', type: 'error', text: true }, () => '删除'),
          default: () => '确定删除该语言吗？'
        })
      ]);
    }
  }
];

// 表格列定义
const columns: DataTableColumns<Requirement> = [
  {
    title: 'ID',
    key: 'id',
    width: 80,
    align: 'center'
  },
  {
    title: '标题',
    key: 'title',
    ellipsis: {
      tooltip: true
    },
    render(row) {
      return h(
        'a',
        {
          class: 'cursor-pointer hover:text-primary transition-colors',
          onClick: () => handleViewDetail(row.id)
        },
        row.title
      );
    }
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    align: 'center',
    render(row) {
      return h(NTag, { type: statusColors[row.status] as any, size: 'small' }, () => statusText[row.status]);
    }
  },
  {
    title: '优先级',
    key: 'priority',
    width: 80,
    align: 'center',
    render(row) {
      return h(NTag, { type: priorityColors[row.priority] as any, size: 'small' }, () => priorityText[row.priority]);
    }
  },
  {
    title: '负责人',
    key: 'assignee',
    width: 100,
    align: 'center',
    render(row) {
      return row.assignee || '-';
    }
  },
  {
    title: '创建时间',
    key: 'createdAt',
    width: 160,
    align: 'center',
    render(row) {
      return new Date(row.createdAt).toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      });
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    align: 'center',
    render(row) {
      const isSplitting = splittingTaskIds.value.has(row.id);

      return h(NSpace, { justify: 'center' }, () => [
        h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            text: true,
            onClick: () => handleViewDetail(row.id)
          },
          () => '查看'
        ),
        h(
          NButton,
          {
            size: 'small',
            type: 'info',
            text: true,
            onClick: () => handleViewTasks(row)
          },
          () => '查看任务'
        ),
        h(
          NButton,
          {
            size: 'small',
            type: 'warning',
            text: true,
            loading: isSplitting,
            disabled: isSplitting,
            onClick: () => openTaskTypeModal(row)
          },
          () => isSplitting ? '拆分中...' : '拆分任务'
        ),
        h(
          NPopconfirm,
          {
            onPositiveClick: () => handleDelete(row.id)
          },
          {
            trigger: () =>
              h(
                NButton,
                {
                  size: 'small',
                  type: 'error',
                  text: true
                },
                () => '删除'
              ),
            default: () => '确定要删除这个需求吗？'
          }
        )
      ]);
    }
  }
];

// 查看详情
function handleViewDetail(id: number) {
  router.push(`/requirement/detail/${id}`);
}

// 新建需求
function handleCreate() {
  router.push('/requirement/detail/new');
}

function handleViewTasks(row: Requirement) {
  router.push({
    path: '/requirement/task-list',
    query: {
      requirementId: String(row.id)
    }
  });
}

// 删除需求
async function handleDelete(id: number) {
  const { error } = await requirementStore.deleteRequirementById(id);
  if (!error) {
    window.$message?.success('删除成功');
  } else {
    window.$message?.error('删除失败');
  }
}

// 打开任务类型选择弹框
function openTaskTypeModal(row: Requirement) {
  currentSplittingRequirement.value = row;
  selectedTaskType.value = 'backend';
  selectedLanguageId.value = null;
  showTaskTypeModal.value = true;
}

// 监听任务类型变化，重置语言选择
watch(selectedTaskType, () => {
  const opts = languageOptions.value;
  selectedLanguageId.value = opts.length > 0 ? opts[0].value as number : null;
});

// 确认拆分任务
async function confirmSplitTasks() {
  if (!currentSplittingRequirement.value) return;

  const row = currentSplittingRequirement.value;
  const taskType = selectedTaskType.value;
  const languageId = selectedLanguageId.value;

  showTaskTypeModal.value = false;

  if (splittingTaskIds.value.has(row.id)) {
    window.$message?.info('需求正在拆分中，完成后会通知您');
    return;
  }

  splittingTaskIds.value.add(row.id);

  try {
    const { data, error } = await splitRequirementToTasksAsync(row.id, taskType, languageId || undefined);

    if (!error && data) {
      const responseData = (data as any)?.data || data;
      if (responseData?.messageId) {
        window.$message?.success('拆分任务已提交，正在后台处理，完成后会通知您');
      } else {
        window.$message?.error('提交拆分任务失败');
      }
    } else {
      const errorMessage = error?.message || '拆分任务失败';
      window.$message?.error(errorMessage);
    }
  } catch (err: any) {
    window.$message?.error(err.message || '拆分任务失败');
  } finally {
    splittingTaskIds.value.delete(row.id);
    currentSplittingRequirement.value = null;
  }
}

// 加载语言列表
async function loadLanguages() {
  languageLoading.value = true;
  const { data } = await fetchLanguageList(true);
  if (data) {
    languages.value = Array.isArray(data) ? data : ((data as any).data || []);
  }
  languageLoading.value = false;
}

// 打开新建语言弹框
function openCreateLanguage() {
  editingLanguage.value = null;
  languageForm.value = {
    name: '',
    category: 'backend',
    displayName: '',
    framework: '',
    description: '',
    codeHints: '',
    remark: '',
    isActive: true,
    sortOrder: 0
  };
  showLanguageModal.value = true;
}

// 打开编辑语言弹框
function openEditLanguage(row: Language) {
  editingLanguage.value = row;
  languageForm.value = {
    name: row.name,
    category: row.category || 'backend',
    displayName: row.displayName,
    framework: row.framework || '',
    description: row.description || '',
    codeHints: row.codeHints || '',
    remark: row.remark || '',
    isActive: row.isActive,
    sortOrder: row.sortOrder
  };
  showLanguageModal.value = true;
}

// 保存语言
async function saveLanguage() {
  if (!languageForm.value.name || !languageForm.value.displayName) {
    window.$message?.error('名称和显示名称不能为空');
    return;
  }

  if (editingLanguage.value) {
    const { error } = await updateLanguage(editingLanguage.value.id, languageForm.value);
    if (!error) {
      window.$message?.success('更新成功');
      showLanguageModal.value = false;
      loadLanguages();
    } else {
      window.$message?.error('更新失败');
    }
  } else {
    const { error } = await createLanguage(languageForm.value);
    if (!error) {
      window.$message?.success('创建成功');
      showLanguageModal.value = false;
      loadLanguages();
    } else {
      window.$message?.error('创建失败');
    }
  }
}

// 删除语言
async function handleDeleteLanguage(id: number) {
  const { error } = await deleteLanguage(id);
  if (!error) {
    window.$message?.success('删除成功');
    loadLanguages();
  } else {
    window.$message?.error('删除失败');
  }
}

// 加载数据
async function loadData() {
  await Promise.all([requirementStore.loadRequirements(), requirementStore.loadStatistics()]);
  loadLanguages();
}

onMounted(() => {
  loadData();
});

onActivated(() => {
  loadData();
});
</script>

<template>
  <div class="requirement-list-page">
    <!-- 统计卡片 -->
    <NGrid :cols="5" :x-gap="16" :y-gap="16" class="mb-4">
      <NGi>
        <NCard>
          <NStatistic label="总需求数" :value="requirementStore.statistics?.total || 0">
            <template #prefix>
              <span class="i-mdi:file-document-outline text-lg"></span>
            </template>
          </NStatistic>
        </NCard>
      </NGi>
      <NGi>
        <NCard>
          <NStatistic label="进行中" :value="requirementStore.statistics?.active || 0">
            <template #prefix>
              <span class="i-mdi:progress-clock text-lg text-blue-500"></span>
            </template>
          </NStatistic>
        </NCard>
      </NGi>
      <NGi>
        <NCard>
          <NStatistic label="已完成" :value="requirementStore.statistics?.completed || 0">
            <template #prefix>
              <span class="i-mdi:check-circle-outline text-lg text-green-500"></span>
            </template>
          </NStatistic>
        </NCard>
      </NGi>
      <NGi>
        <NCard>
          <NStatistic label="高优先级" :value="requirementStore.statistics?.highPriority || 0">
            <template #prefix>
              <span class="i-mdi:alert-circle-outline text-lg text-red-500"></span>
            </template>
          </NStatistic>
        </NCard>
      </NGi>
      <NGi>
        <NCard>
          <div class="flex flex-col">
            <span class="text-gray-500 text-sm mb-2">完成进度</span>
            <NProgress
              type="line"
              :percentage="completionProgress"
              :indicator-placement="'inside'"
              :processing="requirementStore.loading"
            />
          </div>
        </NCard>
      </NGi>
    </NGrid>

    <!-- 标签页 -->
    <NCard>
      <NTabs type="line">
        <NTabPane name="requirements" tab="需求列表">
          <!-- 搜索和筛选 -->
          <NSpace class="mb-4 mt-2" justify="space-between">
            <NSpace>
              <NInput
                v-model:value="searchKeyword"
                placeholder="搜索需求标题或内容"
                clearable
                class="w-60"
              />
              <NSelect
                v-model:value="filterStatus"
                :options="statusOptions"
                placeholder="状态筛选"
                clearable
                class="w-32"
              />
              <NSelect
                v-model:value="filterPriority"
                :options="priorityOptions"
                placeholder="优先级筛选"
                clearable
                class="w-32"
              />
            </NSpace>
            <NSpace>
              <NButton @click="loadData">刷新</NButton>
              <NButton type="primary" @click="handleCreate">新建需求</NButton>
            </NSpace>
          </NSpace>

          <!-- 需求表格 -->
          <NSpin :show="requirementStore.loading">
            <NDataTable
              :columns="columns"
              :data="filteredRequirements"
              :pagination="{ pageSize: 20 }"
              :row-key="(row: Requirement) => row.id"
              striped
              :bordered="false"
            />
            <NEmpty v-if="filteredRequirements.length === 0 && !requirementStore.loading" description="暂无需求数据" />
          </NSpin>
        </NTabPane>

        <NTabPane name="languages" tab="语言管理">
          <NSpace class="mb-4 mt-2" justify="end">
            <NButton @click="loadLanguages">刷新</NButton>
            <NButton type="primary" @click="openCreateLanguage">新增语言</NButton>
          </NSpace>

          <NSpin :show="languageLoading">
            <NDataTable
              :columns="languageColumns"
              :data="languages"
              :pagination="{ pageSize: 10 }"
              :row-key="(row: Language) => row.id"
              striped
              :bordered="false"
            />
            <NEmpty v-if="languages.length === 0 && !languageLoading" description="暂无语言数据" />
          </NSpin>
        </NTabPane>
      </NTabs>
    </NCard>

    <!-- 任务类型选择弹框 -->
    <NModal
      v-model:show="showTaskTypeModal"
      preset="dialog"
      title="拆分任务设置"
      positive-text="确认拆分"
      negative-text="取消"
      @positive-click="confirmSplitTasks"
    >
      <div class="py-4">
        <p class="mb-4 text-gray-600">请选择需要生成的任务类型：</p>
        <NSelect
          v-model:value="selectedTaskType"
          :options="taskTypeOptions"
          placeholder="选择任务类型"
          style="width: 100%;"
        />

        <NDivider style="margin: 16px 0;" />

        <p class="mb-4 text-gray-600">选择编程语言/技术栈：</p>
        <NSelect
          v-model:value="selectedLanguageId"
          :options="languageOptions"
          placeholder="请选择语言"
          style="width: 100%;"
        />
      </div>
    </NModal>

    <!-- 语言编辑弹框 -->
    <NModal
      v-model:show="showLanguageModal"
      preset="dialog"
      :title="editingLanguage ? '编辑语言' : '新增语言'"
      positive-text="保存"
      negative-text="取消"
      @positive-click="saveLanguage"
    >
      <NForm label-placement="left" label-width="80">
        <NFormItem label="分类" required>
          <NSelect v-model:value="languageForm.category" :options="categoryOptions" placeholder="选择分类" />
        </NFormItem>
        <NFormItem label="名称" required>
          <NInput v-model:value="languageForm.name" placeholder="如: go, java, vue, react" />
        </NFormItem>
        <NFormItem label="显示名" required>
          <NInput v-model:value="languageForm.displayName" placeholder="如: Go (Gin + GORM)" />
        </NFormItem>
        <NFormItem label="框架">
          <NInput v-model:value="languageForm.framework" placeholder="如: Gin + GORM" />
        </NFormItem>
        <NFormItem label="描述">
          <NInput v-model:value="languageForm.description" type="textarea" placeholder="语言描述" />
        </NFormItem>
        <NFormItem label="代码提示">
          <NInput v-model:value="languageForm.codeHints" type="textarea" :rows="4" placeholder="代码提示模板" />
        </NFormItem>
        <NFormItem label="备注">
          <NInput v-model:value="languageForm.remark" type="textarea" placeholder="Claude开发备注" />
        </NFormItem>
        <NFormItem label="启用">
          <NSwitch v-model:value="languageForm.isActive" />
        </NFormItem>
        <NFormItem label="排序">
          <NInputNumber v-model:value="languageForm.sortOrder" :min="0" />
        </NFormItem>
      </NForm>
    </NModal>
  </div>
</template>

<style scoped>
.requirement-list-page {
  padding: 16px;
}

:deep(.n-grid .n-card) {
  --n-padding-top: 12px;
  --n-padding-bottom: 12px;
  --n-padding-left: 16px;
  --n-padding-right: 16px;
}

:deep(.n-grid .n-statistic) {
  --n-label-font-size: 13px;
}

:deep(.n-grid .n-statistic .n-statistic-value) {
  font-size: 24px;
}

:deep(.n-grid .n-statistic .n-statistic-label) {
  margin-bottom: 4px;
}

:deep(.n-card) {
  --n-padding-top: 16px;
  --n-padding-bottom: 16px;
}
</style>
