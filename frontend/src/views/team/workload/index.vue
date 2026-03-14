<script setup lang="ts">
import { computed, onMounted, onActivated, ref } from 'vue';
import {
  NCard,
  NGrid,
  NGi,
  NDataTable,
  NTag,
  NSpace,
  NProgress,
  NAvatar,
  NEmpty,
  NSpin,
  NSelect,
  NDatePicker,
  NStatistic,
  NNumberAnimation
} from 'naive-ui';
import type { DataTableColumns, SelectOption } from 'naive-ui';
import { useMemberStore } from '@/store/modules/member';
import { fetchMemberWorkload, fetchMemberAssignments } from '@/service/api/assignment';
import type { MemberWorkload, MemberAssignment } from '@/typings/api/assignment';
import type { Member } from '@/typings/api/member';

const memberStore = useMemberStore();

// 状态
const loading = ref(false);
const workloadMap = ref<Map<number, MemberWorkload>>(new Map());
const selectedMember = ref<number | null>(null);
const memberAssignments = ref<MemberAssignment[]>([]);
const dateRange = ref<[number, number] | null>(null);

// 成员选项
const memberOptions = computed<SelectOption[]>(() => {
  return memberStore.members
    .filter(m => m.status === 'active')
    .map(m => ({
      label: m.name,
      value: m.id
    }));
});

// 工作量表格数据
const workloadTableData = computed(() => {
  return memberStore.members
    .filter(m => m.status === 'active')
    .map(member => {
      const workload = workloadMap.value.get(member.id);
      return {
        ...member,
        ...workload,
        utilization: workload
          ? Math.round((workload.actualHours / Math.max(workload.estimatedHours || 1, 1)) * 100)
          : 0
      };
    });
});

// 工作量表格列定义
const workloadColumns: DataTableColumns<any> = [
  {
    title: '成员',
    key: 'name',
    render(row) {
      return h(NSpace, { align: 'center' }, () => [
        h(NAvatar, {
          round: true,
          size: 'small',
          src: row.avatar,
          name: row.name
        }),
        h('span', row.name)
      ]);
    }
  },
  {
    title: '部门',
    key: 'department',
    ellipsis: { tooltip: true }
  },
  {
    title: '任务数',
    key: 'taskCount',
    width: 100,
    render(row) {
      return row.taskCount || 0;
    }
  },
  {
    title: '预估工时',
    key: 'estimatedHours',
    width: 100,
    render(row) {
      return `${row.estimatedHours || 0}h`;
    }
  },
  {
    title: '实际工时',
    key: 'actualHours',
    width: 100,
    render(row) {
      return `${row.actualHours || 0}h`;
    }
  },
  {
    title: '利用率',
    key: 'utilization',
    width: 150,
    render(row) {
      const percent = row.utilization || 0;
      const status: 'success' | 'warning' | 'error' = percent > 100 ? 'error' : percent > 80 ? 'warning' : 'success';
      return h(NProgress, {
        type: 'line',
        percentage: Math.min(percent, 100),
        status,
        indicatorPlacement: 'inside',
        processing: percent > 100
      });
    }
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render(row) {
      const percent = row.utilization || 0;
      let type: 'success' | 'warning' | 'error' | 'info' = 'info';
      let text = '空闲';

      if (percent > 100) {
        type = 'error';
        text = '超负荷';
      } else if (percent > 80) {
        type = 'warning';
        text = '较忙';
      } else if (percent > 50) {
        type = 'success';
        text = '正常';
      }

      return h(NTag, { type, size: 'small' }, () => text);
    }
  }
];

// 成员任务表格列
const assignmentColumns: DataTableColumns<MemberAssignment> = [
  {
    title: '任务',
    key: 'taskTitle',
    ellipsis: { tooltip: true }
  },
  {
    title: '角色',
    key: 'role',
    width: 100,
    render(row) {
      const roleLabels: Record<string, string> = {
        assignee: '负责人',
        reviewer: '审核人',
        collaborator: '协作者'
      };
      const roleColors: Record<string, 'success' | 'warning' | 'info'> = {
        assignee: 'success',
        reviewer: 'warning',
        collaborator: 'info'
      };
      return h(NTag, { type: roleColors[row.role] || 'info', size: 'small' }, () => roleLabels[row.role] || row.role);
    }
  },
  {
    title: '状态',
    key: 'taskStatus',
    width: 100,
    render(row) {
      const statusColors: Record<string, 'default' | 'success' | 'warning' | 'error' | 'info'> = {
        pending: 'default',
        in_progress: 'info',
        completed: 'success',
        cancelled: 'error'
      };
      const statusLabels: Record<string, string> = {
        pending: '待处理',
        in_progress: '进行中',
        completed: '已完成',
        cancelled: '已取消'
      };
      return h(NTag, { type: statusColors[row.taskStatus] || 'default', size: 'small' }, () => statusLabels[row.taskStatus] || row.taskStatus);
    }
  },
  {
    title: '预估工时',
    key: 'estimatedHours',
    width: 100,
    render(row) {
      return row.estimatedHours ? `${row.estimatedHours}h` : '-';
    }
  },
  {
    title: '实际工时',
    key: 'actualHours',
    width: 100,
    render(row) {
      return row.actualHours ? `${row.actualHours}h` : '-';
    }
  }
];

// 统计数据
const statistics = computed(() => {
  let totalTasks = 0;
  let totalEstimated = 0;
  let totalActual = 0;
  let overworkCount = 0;

  workloadMap.value.forEach(w => {
    totalTasks += w.taskCount || 0;
    totalEstimated += w.estimatedHours || 0;
    totalActual += w.actualHours || 0;
    if (w.actualHours && w.estimatedHours && w.actualHours > w.estimatedHours) {
      overworkCount++;
    }
  });

  return {
    totalTasks,
    totalEstimated,
    totalActual,
    overworkCount,
    avgUtilization: totalEstimated > 0 ? Math.round((totalActual / totalEstimated) * 100) : 0
  };
});

// 加载工作量数据
async function loadWorkload() {
  loading.value = true;
  try {
    const promises = memberStore.members
      .filter(m => m.status === 'active')
      .map(async member => {
        const { data } = await fetchMemberWorkload(member.id);
        if (data) {
          workloadMap.value.set(member.id, data);
        }
      });

    await Promise.all(promises);
  } finally {
    loading.value = false;
  }
}

// 加载成员任务
async function loadMemberAssignments(memberId: number) {
  selectedMember.value = memberId;
  const { data } = await fetchMemberAssignments(memberId);
  if (data) {
    memberAssignments.value = data;
  }
}

// 加载所有数据
async function loadData() {
  await memberStore.loadMembers();
  await loadWorkload();
}

// 引入 h 函数
import { h } from 'vue';

onMounted(() => {
  loadData();
});

onActivated(() => {
  loadData();
});
</script>

<template>
  <div class="workload-page p-16px">
    <!-- 统计概览 -->
    <NCard class="mb-16px">
      <NGrid :cols="5" :x-gap="16">
        <NGi>
          <NStatistic label="总任务数" :value="statistics.totalTasks">
            <template #prefix>
              <span class="stat-icon">📋</span>
            </template>
          </NStatistic>
        </NGi>
        <NGi>
          <NStatistic label="总预估工时" :value="statistics.totalEstimated">
            <template #suffix>h</template>
          </NStatistic>
        </NGi>
        <NGi>
          <NStatistic label="总实际工时" :value="statistics.totalActual">
            <template #suffix>h</template>
          </NStatistic>
        </NGi>
        <NGi>
          <NStatistic label="平均利用率">
            <template #default>
              <NNumberAnimation :from="0" :to="statistics.avgUtilization" />%
            </template>
          </NStatistic>
        </NGi>
        <NGi>
          <NStatistic label="超负荷人数" :value="statistics.overworkCount">
            <template #prefix>
              <span class="stat-icon warning">⚠️</span>
            </template>
          </NStatistic>
        </NGi>
      </NGrid>
    </NCard>

    <!-- 工作量表格 -->
    <NCard title="成员工作量" class="mb-16px">
      <template #header-extra>
        <NButton size="small" @click="loadData">刷新</NButton>
      </template>

      <NSpin :show="loading">
        <NDataTable
          :columns="workloadColumns"
          :data="workloadTableData"
          :pagination="false"
          :row-key="(row: any) => row.id"
          @update:checked-row-keys="(keys: any) => keys.length && loadMemberAssignments(keys[0])"
        />
      </NSpin>
    </NCard>

    <!-- 成员任务详情 -->
    <NCard v-if="selectedMember" title="成员任务详情">
      <template #header-extra>
        <NSpace align="center">
          <span>成员：{{ memberStore.members.find(m => m.id === selectedMember)?.name }}</span>
          <NButton size="small" @click="selectedMember = null">关闭</NButton>
        </NSpace>
      </template>

      <NDataTable
        :columns="assignmentColumns"
        :data="memberAssignments"
        :pagination="{ pageSize: 10 }"
        :row-key="(row: MemberAssignment) => row.taskId"
      />
    </NCard>
  </div>
</template>

<style scoped lang="scss">
.workload-page {
  height: 100%;

  .stat-icon {
    font-size: 20px;
    margin-right: 4px;

    &.warning {
      color: #faad14;
    }
  }
}

.mb-16px {
  margin-bottom: 16px;
}

.p-16px {
  padding: 16px;
}
</style>
