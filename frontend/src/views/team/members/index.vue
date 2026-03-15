<script setup lang="ts">
import { computed, h, onActivated, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import {
  NAvatar,
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NSelect,
  NSpace,
  NTag,
  useMessage
} from 'naive-ui';
import type { DataTableColumns, FormInst } from 'naive-ui';
import { useMemberStore } from '@/store/modules/member';
import type { Member, MemberRole, MemberStatus } from '@/typings/api/member';
import {
  MEMBER_ROLE_LABELS,
  MEMBER_ROLE_OPTIONS,
  MEMBER_STATUS_LABELS,
  MEMBER_STATUS_OPTIONS
} from '@/typings/api/member';

const router = useRouter();
const message = useMessage();
const memberStore = useMemberStore();

// 筛选状态
const filterStatus = ref<MemberStatus | 'all'>('all');
const filterRole = ref<MemberRole | 'all'>('all');
const filterKeyword = ref('');

// 模态框状态
const showModal = ref(false);
const modalType = ref<'create' | 'edit'>('create');
const editingMember = ref<Member | null>(null);
const formRef = ref<FormInst | null>(null);
const formData = ref({
  name: '',
  email: '',
  role: 'member' as MemberRole,
  department: '',
  skills: [] as string[]
});

// 计算过滤后的成员列表
const filteredMembers = computed(() => {
  return memberStore.members.filter(member => {
    if (filterStatus.value !== 'all' && member.status !== filterStatus.value) {
      return false;
    }
    if (filterRole.value !== 'all' && member.role !== filterRole.value) {
      return false;
    }
    if (filterKeyword.value) {
      const keyword = filterKeyword.value.toLowerCase();
      return (
        member.name.toLowerCase().includes(keyword) ||
        member.email?.toLowerCase().includes(keyword) ||
        member.department?.toLowerCase().includes(keyword)
      );
    }
    return true;
  });
});

// 表格列定义
const columns: DataTableColumns<Member> = [
  {
    title: 'ID',
    key: 'id',
    width: 80
  },
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
    title: '邮箱',
    key: 'email',
    ellipsis: { tooltip: true }
  },
  {
    title: '角色',
    key: 'role',
    width: 100,
    render(row) {
      const typeMap: Record<MemberRole, 'success' | 'warning' | 'info'> = {
        admin: 'success',
        leader: 'warning',
        member: 'info'
      };
      return h(NTag, { type: typeMap[row.role], size: 'small' }, () => MEMBER_ROLE_LABELS[row.role]);
    }
  },
  {
    title: '部门',
    key: 'department',
    ellipsis: { tooltip: true }
  },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render(row) {
      return h(
        NTag,
        {
          type: row.status === 'active' ? 'success' : 'default',
          size: 'small'
        },
        () => MEMBER_STATUS_LABELS[row.status]
      );
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render(row) {
      return h(NSpace, {}, () => [
        h(
          NButton,
          {
            size: 'small',
            onClick: () => handleEdit(row)
          },
          () => '编辑'
        ),
        h(
          NButton,
          {
            size: 'small',
            type: row.status === 'active' ? 'warning' : 'success',
            onClick: () => handleToggleStatus(row)
          },
          () => (row.status === 'active' ? '停用' : '激活')
        ),
        h(
          NButton,
          {
            size: 'small',
            type: 'error',
            onClick: () => handleDelete(row)
          },
          () => '删除'
        )
      ]);
    }
  }
];

// 加载数据
async function loadData() {
  await memberStore.loadMembers();
}

// 打开创建模态框
function handleCreate() {
  modalType.value = 'create';
  editingMember.value = null;
  formData.value = {
    name: '',
    email: '',
    role: 'member',
    department: '',
    skills: []
  };
  showModal.value = true;
}

// 打开编辑模态框
function handleEdit(member: Member) {
  modalType.value = 'edit';
  editingMember.value = member;
  formData.value = {
    name: member.name,
    email: member.email || '',
    role: member.role,
    department: member.department || '',
    skills: member.skills || []
  };
  showModal.value = true;
}

// 提交表单
async function handleSubmit() {
  try {
    await formRef.value?.validate();

    if (modalType.value === 'create') {
      const result = await memberStore.createNewMember(formData.value);
      if (result) {
        message.success('成员创建成功');
        showModal.value = false;
      } else {
        message.error(memberStore.error || '创建失败');
      }
    } else if (editingMember.value) {
      const result = await memberStore.updateMemberById(editingMember.value.id, formData.value);
      if (result) {
        message.success('成员更新成功');
        showModal.value = false;
      } else {
        message.error(memberStore.error || '更新失败');
      }
    }
  } catch (e) {
    // 表单验证失败
  }
}

// 切换状态
async function handleToggleStatus(member: Member) {
  const action = member.status === 'active' ? '停用' : '激活';
  const result =
    member.status === 'active'
      ? await memberStore.deactivateMemberById(member.id)
      : await memberStore.activateMemberById(member.id);

  if (result) {
    message.success(`成员已${action}`);
  } else {
    message.error(`${action}失败`);
  }
}

// 删除成员
async function handleDelete(member: Member) {
  const confirmed = window.confirm(`确定要删除成员 "${member.name}" 吗？`);
  if (!confirmed) return;

  const success = await memberStore.removeMember(member.id);
  if (success) {
    message.success('成员已删除');
  } else {
    message.error(memberStore.error || '删除失败');
  }
}

// 生命周期
onMounted(() => {
  loadData();
});

onActivated(() => {
  loadData();
});
</script>

<template>
  <div class="member-list-page p-16px">
    <NCard title="团队成员">
      <template #header-extra>
        <NSpace>
          <NButton type="primary" @click="handleCreate">添加成员</NButton>
          <NButton @click="loadData">刷新</NButton>
        </NSpace>
      </template>

      <!-- 筛选栏 -->
      <NSpace class="mb-16px">
        <NInput v-model:value="filterKeyword" placeholder="搜索成员姓名/邮箱/部门" clearable style="width: 200px" />
        <NSelect
          v-model:value="filterStatus"
          :options="[{ label: '全部状态', value: 'all' }, ...MEMBER_STATUS_OPTIONS]"
          style="width: 120px"
        />
        <NSelect
          v-model:value="filterRole"
          :options="[{ label: '全部角色', value: 'all' }, ...MEMBER_ROLE_OPTIONS]"
          style="width: 120px"
        />
      </NSpace>

      <!-- 数据表格 -->
      <NDataTable
        :columns="columns"
        :data="filteredMembers"
        :loading="memberStore.loading"
        :pagination="{ pageSize: 20 }"
        :row-key="(row: Member) => row.id"
      />
    </NCard>

    <!-- 创建/编辑模态框 -->
    <NModal
      v-model:show="showModal"
      :title="modalType === 'create' ? '添加成员' : '编辑成员'"
      preset="card"
      style="width: 500px"
    >
      <NForm ref="formRef" :model="formData" label-placement="left" label-width="80">
        <NFormItem label="姓名" path="name" :rule="{ required: true, message: '请输入姓名' }">
          <NInput v-model:value="formData.name" placeholder="请输入成员姓名" />
        </NFormItem>
        <NFormItem label="邮箱" path="email">
          <NInput v-model:value="formData.email" placeholder="请输入邮箱地址" />
        </NFormItem>
        <NFormItem label="角色" path="role">
          <NSelect v-model:value="formData.role" :options="MEMBER_ROLE_OPTIONS" />
        </NFormItem>
        <NFormItem label="部门" path="department">
          <NInput v-model:value="formData.department" placeholder="请输入部门名称" />
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showModal = false">取消</NButton>
          <NButton type="primary" @click="handleSubmit">确定</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped lang="scss">
.member-list-page {
  height: 100%;
}
</style>
