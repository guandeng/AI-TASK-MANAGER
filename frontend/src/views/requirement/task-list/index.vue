<script setup lang="ts">
import { h, computed, onMounted, onActivated, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { NCard, NDataTable, NTag, NSpace, NButton, NInput, NSelect, NStatistic, NGrid, NGi, NProgress, NEmpty, NSpin } from 'naive-ui';
import type { DataTableColumns, DataTableRowKey, SelectOption } from 'naive-ui';
import { useTaskStore } from '@/store/modules/task';
import { fetchLanguageList } from '@/service/api/language';
import type { Task, TaskStatus, TaskListParams, TaskCategory } from '@/typings/api/task';

const route = useRoute();
const router = useRouter();
const taskStore = useTaskStore();

function confirmAction(message: string) {
  return window.confirm(message);
}

// 筛选状态
const filterStatus = ref<TaskStatus | 'all'>('all');
const filterCategory = ref<TaskCategory | 'all'>('all');
const filterRequirementId = ref<number | 'all'>('all');
const filterLanguageId = ref<number | 'all'>('all');
const filterExpandStatus = ref<'all' | 'expanding' | 'expanded' | 'none'>('all');
const searchText = ref('');
const checkedRowKeys = ref<DataTableRowKey[]>([]);

// 分页
const currentPage = ref(1);
const pageSize = ref(20);
const total = ref(0);

// 语言选项
const languageOptions = ref<SelectOption[]>([{ label: '全部语言', value: 'all' }]);

// 加载语言列表
async function loadLanguageOptions() {
  const { data } = await fetchLanguageList();
  if (data) {
    const list = (data as any).data || data;
    if (Array.isArray(list)) {
      languageOptions.value = [
        { label: '全部语言', value: 'all' },
        ...list.map((lang: any) => ({
          label: lang.displayName || lang.name,
          value: lang.id
        }))
      ];
    }
  }
}

// 状态选项
const statusOptions = [
  { label: '全部', value: 'all' },
  { label: '待处理', value: 'pending' },
  { label: '进行中', value: 'in-progress' },
  { label: '已暂停', value: 'paused' },
  { label: '已完成', value: 'done' },
  { label: '已延期', value: 'deferred' }
];

// 分类选项
const categoryOptions = [
  { label: '全部分类', value: 'all' },
  { label: '前端', value: 'frontend' },
  { label: '后端', value: 'backend' }
];

// 拆分状态选项
const expandStatusOptions = [
  { label: '全部状态', value: 'all' },
  { label: '拆分中', value: 'expanding' },
  { label: '已拆分', value: 'expanded' },
  { label: '未拆分', value: 'none' }
];

// 负责人选项
const assigneeOptions = [
  { label: '张三', value: 'zhangsan' },
  { label: '李四', value: 'lisi' }
];

// 优先级颜色映射
const priorityColorMap: Record<string, 'error' | 'warning' | 'success'> = {
  high: 'error',
  medium: 'warning',
  low: 'success'
};

// 状态颜色映射
const statusColorMap: Record<string, 'default' | 'success' | 'warning' | 'error' | 'info'> = {
  pending: 'default',
  'in-progress': 'info',
  paused: 'warning',
  done: 'success',
  deferred: 'warning'
};

// 状态文字映射
const statusTextMap: Record<string, string> = {
  pending: '待处理',
  'in-progress': '进行中',
  paused: '已暂停',
  done: '已完成',
  deferred: '已延期'
};

// 优先级文字映射
const priorityTextMap: Record<string, string> = {
  high: '高',
  medium: '中',
  low: '低'
};

// 分类文字映射
const categoryTextMap: Record<string, string> = {
  '': '-',
  frontend: '前端',
  backend: '后端'
};

const requirementOptions = computed<SelectOption[]>(() => {
  const options: SelectOption[] = [{ label: '全部需求', value: 'all' }];
  const seen = new Set<number>();

  taskStore.tasks.forEach(task => {
    if (!task.requirementId) {
      return;
    }

    if (seen.has(task.requirementId)) {
      return;
    }

    seen.add(task.requirementId);
    options.push({
      label: task.requirementTitle || '(需求已删除)',
      value: task.requirementId
    });
  });

  return options;
});

// 表格列定义
const columns: DataTableColumns<Task> = [
  {
    type: 'selection'
  },
  {
    title: 'ID',
    key: 'id',
    width: 80
  },
  {
    title: '需求',
    key: 'requirementTitle',
    width: 150,
    render(row) {
      if (!row.requirementId) {
        return '-';
      }

      // 如果有关联需求ID但没有标题，说明需求可能已被删除
      const displayText = row.requirementTitle || '(需求已删除)';

      return h(
        'a',
        {
          class: 'cursor-pointer hover:text-primary transition-colors',
          onClick: () => {
            router.push(`/requirement/detail/${row.requirementId}`);
          }
        },
        displayText
      );
    }
  },
  {
    title: '任务',
    key: 'title',
    render(row) {
      return row.titleTrans || row.title;
    }
  },
  {
    title: '分类',
    key: 'category',
    width: 80,
    align: 'center',
    render(row) {
      if (!row.category) return '-';
      return h(NTag, {
        type: row.category === 'frontend' ? 'success' : 'info',
        size: 'small'
      }, { default: () => categoryTextMap[row.category] || row.category });
    }
  },
  {
    title: '语言',
    key: 'languageName',
    width: 100,
    render(row) {
      return row.languageName || '-';
    }
  },
  {
    title: '子任务',
    key: 'subtasks',
    width: 100,
    render(row) {
      // 优先使用后端返回的统计字段
      if (row.subtaskCount !== undefined) {
        if (row.subtaskCount === 0) return '-';
        return `${row.subtaskDoneCount || 0}/${row.subtaskCount}`;
      }
      // 兼容旧数据：从 subtasks 数组计算
      if (!row.subtasks || row.subtasks.length === 0) return '-';
      const doneCount = row.subtasks.filter(st => st.status === 'done').length;
      return `${doneCount}/${row.subtasks.length}`;
    }
  },
  {
    title: '优先级',
    key: 'priority',
    width: 80,
    render(row) {
      return h(NTag, { type: priorityColorMap[row.priority] }, { default: () => priorityTextMap[row.priority] });
    }
  },
  {
    title: '负责人',
    key: 'assignee',
    width: 120,
    render(row) {
      return h(NSelect, {
        value: row.assignee,
        options: assigneeOptions,
        size: 'small',
        style: 'width: 100px',
        placeholder: '选择',
        onUpdateValue: (value: string) => handleAssigneeChange(row.id, value)
      });
    }
  },
  {
    title: '状态',
    key: 'status',
    width: 120,
    render(row) {
      return h(NSelect, {
        value: row.status,
        options: statusOptions.slice(1), // 排除"全部"选项
        size: 'small',
        style: 'width: 100px',
        onUpdateValue: (value: TaskStatus) => handleStatusChange(row.id, value)
      });
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render(row) {
      return h(
        NButton,
        {
          size: 'small',
          type: 'primary',
          ghost: true,
          onClick: () => viewTaskDetail(row.id)
        },
        { default: () => '详情' }
      );
    }
  }
];

// 查看任务详情
function viewTaskDetail(id: number) {
  router.push(`/requirement/task-detail/${id}`);
}

async function loadTaskListData() {
  const params: TaskListParams = {
    page: currentPage.value,
    pageSize: pageSize.value
  };

  if (filterRequirementId.value !== 'all') {
    params.requirementId = filterRequirementId.value;
  }

  if (filterStatus.value !== 'all') {
    params.status = filterStatus.value;
  }

  if (filterCategory.value !== 'all') {
    params.category = filterCategory.value;
  }

  if (filterLanguageId.value !== 'all') {
    params.languageId = filterLanguageId.value;
  }

  if (filterExpandStatus.value !== 'all') {
    params.expandStatus = filterExpandStatus.value;
  }

  if (searchText.value) {
    params.keyword = searchText.value;
  }

  await taskStore.loadTasks(params);
  total.value = taskStore.total;
}

async function handleBatchDelete() {
  const taskIds = checkedRowKeys.value.map(key => Number(key)).filter(id => !Number.isNaN(id));

  if (!taskIds.length) {
    return;
  }

  if (!confirmAction(`确认批量删除选中的 ${taskIds.length} 个任务吗？`)) {
    return;
  }

  const { successIds, failedIds } = await taskStore.batchDeleteTasks(taskIds);

  checkedRowKeys.value = checkedRowKeys.value.filter(key => !successIds.includes(Number(key)));
  await loadTaskListData();

  if (successIds.length && !failedIds.length) {
    window.$message?.success(`已删除 ${successIds.length} 个任务`);
    return;
  }

  if (successIds.length && failedIds.length) {
    window.$message?.warning(`成功删除 ${successIds.length} 个，失败 ${failedIds.length} 个`);
    return;
  }

  window.$message?.error('批量删除失败');
}

function syncRequirementFilterFromRoute() {
  const rawRequirementId = Array.isArray(route.query.requirementId)
    ? route.query.requirementId[0]
    : route.query.requirementId;

  const parsedRequirementId = Number(rawRequirementId);

  if (rawRequirementId && !Number.isNaN(parsedRequirementId) && parsedRequirementId > 0) {
    filterRequirementId.value = parsedRequirementId;
  } else {
    filterRequirementId.value = 'all';
  }
}

// 处理状态变更
async function handleStatusChange(id: number, status: TaskStatus) {
  if (!confirmAction(`确认将任务 ${id} 状态改为“${statusTextMap[status]}”吗？`)) {
    return;
  }
  await taskStore.setTaskStatus(id, status);
}

// 处理负责人变更
async function handleAssigneeChange(id: number, assignee: string) {
  const assigneeLabel = assigneeOptions.find(item => item.value === assignee)?.label || assignee;
  if (!confirmAction(`确认将任务 ${id} 负责人改为“${assigneeLabel}”吗？`)) {
    return;
  }
  await taskStore.setTaskAssignee(id, assignee);
}

// 分页配置
const paginationConfig = computed(() => ({
  page: currentPage.value,
  pageSize: pageSize.value,
  itemCount: total.value,
  showSizePicker: true,
  showQuickJumper: true,
  pageSizes: [10, 20, 50, 100],
  onChange: (page: number, size?: number) => {
    currentPage.value = page;
    if (size && size !== pageSize.value) {
      pageSize.value = size;
    }
  },
  onUpdatePageSize: (size: number) => {
    pageSize.value = size;
    currentPage.value = 1;
  }
}));

// 监听分页变化
watch([currentPage, pageSize], () => {
  loadTaskListData();
});
// 完成率
const completionRate = computed(() => {
  if (taskStore.statistics.total === 0) return 0;
  return Math.round((taskStore.statistics.done / taskStore.statistics.total) * 100);
});

// 加载数据
onMounted(async () => {
  syncRequirementFilterFromRoute();
  await loadTaskListData();
});

// 页面激活时重新加载数据
onActivated(async () => {
  syncRequirementFilterFromRoute();
  await loadTaskListData();
});

watch(
  () => route.query.requirementId,
  async () => {
    syncRequirementFilterFromRoute();
    await loadTaskListData();
  }
);

watch(
  () => [filterRequirementId.value, filterStatus.value, filterCategory.value, searchText.value],
  () => {
    currentPage.value = 1;
    loadTaskListData();
  }
);
</script>

<template>
  <div class="task-list-page">
    <!-- 统计卡片 -->
    <NGrid :cols="4" :x-gap="16" :y-gap="16" class="mb-16px">
      <NGi>
        <NCard>
          <NStatistic label="总任务数" :value="taskStore.statistics.total">
            <template #prefix>
              <span class="i-mdi:clipboard-list text-24px"></span>
            </template>
          </NStatistic>
        </NCard>
      </NGi>
      <NGi>
        <NCard>
          <NStatistic label="已完成" :value="taskStore.statistics.done">
            <template #prefix>
              <span class="i-mdi:check-circle text-24px text-green"></span>
            </template>
          </NStatistic>
        </NCard>
      </NGi>
      <NGi>
        <NCard>
          <NStatistic label="待处理" :value="taskStore.statistics.pending">
            <template #prefix>
              <span class="i-mdi:clock-outline text-24px text-orange"></span>
            </template>
          </NStatistic>
        </NCard>
      </NGi>
      <NGi>
        <NCard>
          <div class="flex flex-col">
            <span class="text-gray-500 mb-8px">完成进度</span>
            <NProgress
              type="line"
              :percentage="completionRate"
              :show-indicator="true"
              :height="24"
              :border-radius="4"
            />
          </div>
        </NCard>
      </NGi>
    </NGrid>

    <!-- 任务列表 -->
    <NCard title="任务列表">
      <template #header-extra>
        <NSpace>
          <NButton type="primary" @click="$router.push('/requirement/task-create')">
            新建任务
          </NButton>
          <NButton @click="loadTaskListData">
            刷新
          </NButton>
          <NSelect
            v-model:value="filterRequirementId"
            :options="requirementOptions"
            style="width: 150px"
            placeholder="需求筛选"
          />
          <NSelect
            v-model:value="filterCategory"
            :options="categoryOptions"
            style="width: 100px"
            placeholder="分类"
          />
          <NSelect
            v-model:value="filterStatus"
            :options="statusOptions"
            style="width: 120px"
          />
          <NInput
            v-model:value="searchText"
            placeholder="搜索任务..."
            clearable
            style="width: 180px"
          />
        </NSpace>
      </template>

      <NSpin :show="taskStore.loading">
        <NDataTable
          :columns="columns"
          :data="taskStore.tasks"
          :row-key="(row: Task) => row.id"
          v-model:checked-row-keys="checkedRowKeys"
          :bordered="false"
          :pagination="paginationConfig"
        />
        <div class="mt-16px flex justify-start">
          <NButton type="error" ghost :disabled="checkedRowKeys.length === 0" @click="handleBatchDelete">
            批量删除
          </NButton>
        </div>
        <NEmpty v-if="taskStore.tasks.length === 0 && !taskStore.loading" description="暂无任务数据" />
      </NSpin>
    </NCard>
  </div>
</template>

<style scoped lang="scss">
.task-list-page {
  padding: 16px;
}

.mb-16px {
  margin-bottom: 16px;
}

.mb-8px {
  margin-bottom: 8px;
}

.text-24px {
  font-size: 24px;
}

.text-green {
  color: #18a058;
}

.text-orange {
  color: #f0a020;
}

.text-gray-500 {
  color: #6b7280;
}
</style>
