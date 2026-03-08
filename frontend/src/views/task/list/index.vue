<script setup lang="ts">
import { h, computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { NCard, NDataTable, NTag, NSpace, NButton, NInput, NSelect, NStatistic, NGrid, NGi, NProgress, NEmpty, NSpin } from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import { useTaskStore } from '@/store/modules/task';
import type { Task, TaskStatus } from '@/typings/api/task';

const router = useRouter();
const taskStore = useTaskStore();

// 筛选状态
const filterStatus = ref<TaskStatus | 'all'>('all');
const searchText = ref('');

// 状态选项
const statusOptions = [
  { label: '全部', value: 'all' },
  { label: '待处理', value: 'pending' },
  { label: '进行中', value: 'in-progress' },
  { label: '已完成', value: 'done' },
  { label: '已延期', value: 'deferred' }
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
  done: 'success',
  deferred: 'warning'
};

// 状态文字映射
const statusTextMap: Record<string, string> = {
  pending: '待处理',
  'in-progress': '进行中',
  done: '已完成',
  deferred: '已延期'
};

// 优先级文字映射
const priorityTextMap: Record<string, string> = {
  high: '高',
  medium: '中',
  low: '低'
};

// 过滤后的任务
const filteredTasks = computed(() => {
  let result = taskStore.tasks;

  // 状态筛选
  if (filterStatus.value !== 'all') {
    result = result.filter(task => task.status === filterStatus.value);
  }

  // 搜索筛选
  if (searchText.value) {
    const keyword = searchText.value.toLowerCase();
    result = result.filter(task =>
      task.title.toLowerCase().includes(keyword) ||
      (task.titleTrans && task.titleTrans.toLowerCase().includes(keyword)) ||
      task.description.toLowerCase().includes(keyword) ||
      (task.descriptionTrans && task.descriptionTrans.toLowerCase().includes(keyword))
    );
  }

  return result;
});

// 表格列定义
const columns: DataTableColumns<Task> = [
  {
    title: 'ID',
    key: 'id',
    width: 80
  },
  {
    title: '需求',
    key: 'requirement',
    width: 150,
    render() {
      return '-';
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
    title: '子任务',
    key: 'subtasks',
    width: 100,
    render(row) {
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
    width: 100,
    render(row) {
      return row.assignee || '-';
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
  }
];

// 查看任务详情
function viewTaskDetail(id: number) {
  router.push(`/task/detail/${id}`);
}

// 处理状态变更
async function handleStatusChange(id: number, status: TaskStatus) {
  await taskStore.setTaskStatus(id, status);
}

// 完成率
const completionRate = computed(() => {
  if (taskStore.statistics.total === 0) return 0;
  return Math.round((taskStore.statistics.done / taskStore.statistics.total) * 100);
});

// 加载数据
onMounted(async () => {
  await taskStore.loadTasks();
});
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
          <NInput
            v-model:value="searchText"
            placeholder="搜索任务..."
            clearable
            style="width: 200px"
          />
          <NSelect
            v-model:value="filterStatus"
            :options="statusOptions"
            style="width: 120px"
          />
          <NButton @click="taskStore.loadTasks()">
            <template #icon>
              <span class="i-mdi:refresh"></span>
            </template>
            刷新
          </NButton>
        </NSpace>
      </template>

      <NSpin :show="taskStore.loading">
        <NDataTable
          :columns="columns"
          :data="filteredTasks"
          :row-key="(row: Task) => row.id"
          :bordered="false"
        />
        <NEmpty v-if="filteredTasks.length === 0 && !taskStore.loading" description="暂无任务数据" />
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
