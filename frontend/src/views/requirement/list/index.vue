<script setup lang="ts">
import { onMounted, onActivated, ref, computed, h } from 'vue';
import { useRouter } from 'vue-router';
import { NButton, NCard, NDataTable, NGrid, NGi, NInput, NSelect, NSpace, NStatistic, NTag, NPopconfirm, NProgress, NSpin, NEmpty, NModal, NRadioGroup, NRadio } from 'naive-ui';
import type { DataTableColumns, SelectOption } from 'naive-ui';
import { useRequirementStore } from '@/store/modules/requirement';
import type { Requirement, RequirementStatus, RequirementPriority } from '@/typings/api/requirement';
import { splitRequirementToTasksAsync, type TaskType } from '@/service/api/requirement';

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
const currentSplittingRequirement = ref<Requirement | null>(null);

// 任务类型选项
const taskTypeOptions = [
  { label: '后端', value: 'backend' as TaskType },
  { label: '前端', value: 'frontend' as TaskType },
  { label: '前后端', value: 'fullstack' as TaskType }
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
  selectedTaskType.value = 'backend'; // 默认选择后端
  showTaskTypeModal.value = true;
}

// 确认拆分任务 - 异步版本
async function confirmSplitTasks() {
  if (!currentSplittingRequirement.value) return;

  const row = currentSplittingRequirement.value;
  const taskType = selectedTaskType.value;

  showTaskTypeModal.value = false;

  if (splittingTaskIds.value.has(row.id)) {
    window.$message?.info('需求正在拆分中，完成后会通知您');
    return;
  }

  splittingTaskIds.value.add(row.id);

  try {
    const { data, error } = await splitRequirementToTasksAsync(row.id, taskType);

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

// 加载数据
async function loadData() {
  await Promise.all([requirementStore.loadRequirements(), requirementStore.loadStatistics()]);
}

onMounted(() => {
  loadData();
});

// 页面激活时重新加载数据
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

    <!-- 搜索和筛选 -->
    <NCard class="mb-4">
      <NSpace justify="center">
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
        <NButton type="primary" @click="handleCreate">
          新建需求
        </NButton>
      </NSpace>
    </NCard>

    <!-- 数据表格 -->
    <NCard title="需求列表">
      <template #header-extra>
        <NButton @click="loadData">
          <template #icon>
            <span class="i-mdi:refresh"></span>
          </template>
          刷新
        </NButton>
      </template>
      <NSpin :show="requirementStore.loading">
        <NDataTable
          :columns="columns"
          :data="filteredRequirements"
          :pagination="{
            pageSize: 20
          }"
          :row-key="(row: Requirement) => row.id"
          striped
          :bordered="false"
        />
        <NEmpty v-if="filteredRequirements.length === 0 && !requirementStore.loading" description="暂无需求数据" />
      </NSpin>
    </NCard>

    <!-- 任务类型选择弹框 -->
    <NModal
      v-model:show="showTaskTypeModal"
      preset="dialog"
      title="选择任务类型"
      positive-text="确认拆分"
      negative-text="取消"
      @positive-click="confirmSplitTasks"
    >
      <div class="py-4">
        <p class="mb-4 text-gray-600">请选择需要生成的任务类型：</p>
        <NRadioGroup v-model:value="selectedTaskType" class="flex flex-col gap-3">
          <NRadio
            v-for="option in taskTypeOptions"
            :key="option.value"
            :value="option.value"
            :label="option.label"
          />
        </NRadioGroup>
      </div>
    </NModal>
  </div>
</template>

<style scoped>
.requirement-list-page {
  padding: 16px;
}

/* 统计卡片样式 - 更紧凑 */
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

/* 其他卡片保持原样 */
:deep(.n-card) {
  --n-padding-top: 16px;
  --n-padding-bottom: 16px;
}
</style>
