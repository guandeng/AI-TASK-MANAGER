<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { NCard, NSpace, NButton, NAvatar, NTag, NSelect, NModal, NInputNumber, NEmpty, NSpin, NPopconfirm, NTooltip } from 'naive-ui';
import type { SelectOption } from 'naive-ui';
import { fetchTaskAssignments, assignTaskToMember, unassignTaskFromMember } from '@/service/api/assignment';
import { fetchMemberList } from '@/service/api/member';
import type { Assignment, AssignmentRole } from '@/typings/api/assignment';
import type { Member } from '@/typings/api/member';
import { ASSIGNMENT_ROLE_LABELS, ASSIGNMENT_ROLE_COLORS } from '@/typings/api/assignment';

const props = defineProps<{
  taskId: number;
}>();

const emit = defineEmits<{
  (e: 'assigned'): void;
  (e: 'unassigned'): void;
}>();

// 状态
const loading = ref(false);
const assignments = ref<Assignment[]>([]);
const members = ref<Member[]>([]);

// 模态框状态
const showAssignModal = ref(false);
const selectedMemberId = ref<number | null>(null);
const selectedRole = ref<AssignmentRole>('assignee');
const estimatedHours = ref<number | null>(null);
const assignLoading = ref(false);

// 成员选项
const memberOptions = computed<SelectOption[]>(() => {
  // 过滤掉已分配的成员
  const assignedIds = new Set(assignments.value.map(a => a.memberId));
  return members.value
    .filter(m => m.status === 'active' && !assignedIds.has(m.id))
    .map(m => ({
      label: m.name,
      value: m.id,
      disabled: assignedIds.has(m.id)
    }));
});

// 角色选项
const roleOptions: SelectOption[] = [
  { label: '负责人', value: 'assignee' },
  { label: '审核人', value: 'reviewer' },
  { label: '协作者', value: 'collaborator' }
];

// 按角色分组
const assignmentsByRole = computed(() => {
  const result: Record<AssignmentRole, Assignment[]> = {
    assignee: [],
    reviewer: [],
    collaborator: []
  };
  assignments.value.forEach(a => {
    result[a.role].push(a);
  });
  return result;
});

// 加载分配列表
async function loadAssignments() {
  if (!props.taskId) return;
  loading.value = true;
  try {
    const { data, error } = await fetchTaskAssignments(props.taskId);
    if (!error && data) {
      assignments.value = data;
    }
  } finally {
    loading.value = false;
  }
}

// 加载成员列表
async function loadMembers() {
  try {
    const { data, error } = await fetchMemberList({ status: 'active' });
    if (!error && data) {
      members.value = data;
    }
  } catch (e) {
    console.error('Failed to load members:', e);
  }
}

// 打开分配模态框
function openAssignModal() {
  selectedMemberId.value = null;
  selectedRole.value = 'assignee';
  estimatedHours.value = null;
  showAssignModal.value = true;
}

// 执行分配
async function handleAssign() {
  if (!selectedMemberId.value) {
    window.$message?.warning('请选择成员');
    return;
  }

  assignLoading.value = true;
  try {
    const { error } = await assignTaskToMember(props.taskId, {
      memberId: selectedMemberId.value,
      role: selectedRole.value,
      estimatedHours: estimatedHours.value || undefined
    });

    if (error) {
      window.$message?.error('分配失败');
    } else {
      window.$message?.success('分配成功');
      showAssignModal.value = false;
      await loadAssignments();
      emit('assigned');
    }
  } finally {
    assignLoading.value = false;
  }
}

// 移除分配
async function handleUnassign(assignment: Assignment) {
  try {
    const { error } = await unassignTaskFromMember(props.taskId, assignment.id);
    if (error) {
      window.$message?.error('移除失败');
    } else {
      window.$message?.success('已移除');
      await loadAssignments();
      emit('unassigned');
    }
  } catch (e) {
    window.$message?.error('移除失败');
  }
}

// 监听 taskId 变化
watch(() => props.taskId, () => {
  loadAssignments();
}, { immediate: true });

onMounted(() => {
  loadMembers();
});
</script>

<template>
  <NCard title="任务分配" size="small">
    <template #header-extra>
      <NButton type="primary" size="small" @click="openAssignModal">
        分配成员
      </NButton>
    </template>

    <NSpin :show="loading">
      <div v-if="assignments.length > 0" class="assignment-list">
        <!-- 负责人 -->
        <div v-if="assignmentsByRole.assignee.length > 0" class="role-group">
          <div class="role-label">负责人</div>
          <NSpace vertical :size="8">
            <div
              v-for="assignment in assignmentsByRole.assignee"
              :key="assignment.id"
              class="assignment-item"
            >
              <NSpace align="center">
                <NAvatar
                  round
                  size="small"
                  :src="assignment.member?.avatar"
                  :name="assignment.member?.name"
                />
                <span>{{ assignment.member?.name || '未知成员' }}</span>
                <NTag v-if="assignment.estimatedHours" size="small" type="info">
                  预估: {{ assignment.estimatedHours }}h
                </NTag>
              </NSpace>
              <NPopconfirm @positive-click="handleUnassign(assignment)">
                <template #trigger>
                  <NButton text type="error" size="small">
                    移除
                  </NButton>
                </template>
                确认移除此分配？
              </NPopconfirm>
            </div>
          </NSpace>
        </div>

        <!-- 审核人 -->
        <div v-if="assignmentsByRole.reviewer.length > 0" class="role-group">
          <div class="role-label">审核人</div>
          <NSpace vertical :size="8">
            <div
              v-for="assignment in assignmentsByRole.reviewer"
              :key="assignment.id"
              class="assignment-item"
            >
              <NSpace align="center">
                <NAvatar
                  round
                  size="small"
                  :src="assignment.member?.avatar"
                  :name="assignment.member?.name"
                />
                <span>{{ assignment.member?.name || '未知成员' }}</span>
              </NSpace>
              <NPopconfirm @positive-click="handleUnassign(assignment)">
                <template #trigger>
                  <NButton text type="error" size="small">
                    移除
                  </NButton>
                </template>
                确认移除此分配？
              </NPopconfirm>
            </div>
          </NSpace>
        </div>

        <!-- 协作者 -->
        <div v-if="assignmentsByRole.collaborator.length > 0" class="role-group">
          <div class="role-label">协作者</div>
          <NSpace vertical :size="8">
            <div
              v-for="assignment in assignmentsByRole.collaborator"
              :key="assignment.id"
              class="assignment-item"
            >
              <NSpace align="center">
                <NAvatar
                  round
                  size="small"
                  :src="assignment.member?.avatar"
                  :name="assignment.member?.name"
                />
                <span>{{ assignment.member?.name || '未知成员' }}</span>
              </NSpace>
              <NPopconfirm @positive-click="handleUnassign(assignment)">
                <template #trigger>
                  <NButton text type="error" size="small">
                    移除
                  </NButton>
                </template>
                确认移除此分配？
              </NPopconfirm>
            </div>
          </NSpace>
        </div>
      </div>
      <NEmpty v-else description="暂未分配成员" />
    </NSpin>

    <!-- 分配模态框 -->
    <NModal
      v-model:show="showAssignModal"
      preset="dialog"
      title="分配任务"
      positive-text="确认"
      negative-text="取消"
      :loading="assignLoading"
      @positive-click="handleAssign"
    >
      <NSpace vertical :size="16">
        <div>
          <div class="mb-8px">选择成员</div>
          <NSelect
            v-model:value="selectedMemberId"
            :options="memberOptions"
            placeholder="请选择成员"
            filterable
          />
        </div>
        <div>
          <div class="mb-8px">角色</div>
          <NSelect
            v-model:value="selectedRole"
            :options="roleOptions"
            placeholder="请选择角色"
          />
        </div>
        <div>
          <div class="mb-8px">预估工时（小时，可选）</div>
          <NInputNumber
            v-model:value="estimatedHours"
            placeholder="预估工时"
            :min="0"
            :max="1000"
            style="width: 100%"
          />
        </div>
      </NSpace>
    </NModal>
  </NCard>
</template>

<style scoped lang="scss">
.assignment-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.role-group {
  .role-label {
    font-weight: 500;
    margin-bottom: 8px;
    color: #666;
    font-size: 13px;
  }
}

.assignment-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: #f5f5f5;
  border-radius: 6px;

  &:hover {
    background: #eee;
  }
}

.mb-8px {
  margin-bottom: 8px;
}
</style>
