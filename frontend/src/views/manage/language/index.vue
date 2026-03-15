<script setup lang="ts">
import { h, onMounted, ref } from 'vue';
import {
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NPopconfirm,
  NSpace,
  NSwitch,
  useMessage
} from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import {
  type Language,
  type LanguageCreateRequest,
  createLanguage,
  deleteLanguage,
  fetchLanguageList,
  updateLanguage
} from '@/service/api/language';

const message = useMessage();
const loading = ref(false);
const languages = ref<Language[]>([]);
const showModal = ref(false);
const editingId = ref<number | null>(null);
const formData = ref<LanguageCreateRequest>({
  name: '',
  category: 'backend',
  displayName: '',
  framework: '',
  description: '',
  codeHints: '',
  remark: '',
  isActive: true,
  sortOrder: 0
});

// 表格列定义
const columns: DataTableColumns<Language> = [
  {
    title: 'ID',
    key: 'id',
    width: 60
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
    width: 150
  },
  {
    title: '描述',
    key: 'description',
    ellipsis: { tooltip: true }
  },
  {
    title: 'Claude备注',
    key: 'remark',
    ellipsis: { tooltip: true }
  },
  {
    title: '启用',
    key: 'isActive',
    width: 80,
    render(row) {
      return row.isActive ? '是' : '否';
    }
  },
  {
    title: '排序',
    key: 'sortOrder',
    width: 60
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    render(row) {
      return h(NSpace, {}, () => [
        h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            text: true,
            onClick: () => handleEdit(row)
          },
          () => '编辑'
        ),
        h(
          NPopconfirm,
          { onPositiveClick: () => handleDelete(row.id) },
          {
            trigger: () => h(NButton, { size: 'small', type: 'error', text: true }, () => '删除'),
            default: () => '确定删除该语言吗？'
          }
        )
      ]);
    }
  }
];

// 加载语言列表
async function loadLanguages() {
  loading.value = true;
  try {
    const { data } = await fetchLanguageList(true);
    if (data) {
      languages.value = data;
    }
  } catch {
    message.error('加载语言列表失败');
  } finally {
    loading.value = false;
  }
}

// 打开新增弹窗
function handleAdd() {
  editingId.value = null;
  formData.value = {
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
  showModal.value = true;
}

// 打开编辑弹窗
function handleEdit(row: Language) {
  editingId.value = row.id;
  formData.value = {
    name: row.name,
    displayName: row.displayName,
    framework: row.framework,
    description: row.description,
    codeHints: row.codeHints,
    remark: row.remark,
    isActive: row.isActive,
    sortOrder: row.sortOrder
  };
  showModal.value = true;
}

// 删除语言
async function handleDelete(id: number) {
  try {
    await deleteLanguage(id);
    message.success('删除成功');
    loadLanguages();
  } catch {
    message.error('删除失败');
  }
}

// 提交表单
async function handleSubmit() {
  if (!formData.value.name || !formData.value.displayName) {
    message.warning('请填写必填项');
    return;
  }

  try {
    if (editingId.value) {
      await updateLanguage(editingId.value, formData.value);
      message.success('更新成功');
    } else {
      await createLanguage(formData.value);
      message.success('创建成功');
    }
    showModal.value = false;
    loadLanguages();
  } catch {
    message.error('操作失败');
  }
}

onMounted(() => {
  loadLanguages();
});
</script>

<template>
  <div class="language-page">
    <NCard title="语言管理">
      <template #header-extra>
        <NButton type="primary" @click="handleAdd">新增语言</NButton>
      </template>

      <NDataTable
        :columns="columns"
        :data="languages"
        :loading="loading"
        :pagination="false"
        :row-key="(row: Language) => row.id"
      />
    </NCard>

    <NModal
      v-model:show="showModal"
      preset="dialog"
      :title="editingId ? '编辑语言' : '新增语言'"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleSubmit"
    >
      <NForm label-placement="left" label-width="80px">
        <NFormItem label="名称" required>
          <NInput v-model:value="formData.name" placeholder="如: go、java、python" />
        </NFormItem>
        <NFormItem label="显示名称" required>
          <NInput v-model:value="formData.displayName" placeholder="如: Go (Gin + GORM)" />
        </NFormItem>
        <NFormItem label="框架">
          <NInput v-model:value="formData.framework" placeholder="如: Gin + GORM" />
        </NFormItem>
        <NFormItem label="描述">
          <NInput v-model:value="formData.description" type="textarea" placeholder="语言/技术栈描述" :rows="2" />
        </NFormItem>
        <NFormItem label="代码提示">
          <NInput v-model:value="formData.codeHints" type="textarea" placeholder="代码实现提示模板" :rows="3" />
        </NFormItem>
        <NFormItem label="Claude备注">
          <NInput
            v-model:value="formData.remark"
            type="textarea"
            placeholder="Claude开发时的备注说明（如：禁止开启我已禁止的菜单）"
            :rows="2"
          />
        </NFormItem>
        <NFormItem label="启用">
          <NSwitch v-model:value="formData.isActive" />
        </NFormItem>
        <NFormItem label="排序">
          <NInputNumber v-model:value="formData.sortOrder" :min="0" />
        </NFormItem>
      </NForm>
    </NModal>
  </div>
</template>

<style scoped>
.language-page {
  padding: 16px;
}
</style>
