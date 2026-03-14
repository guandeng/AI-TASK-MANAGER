<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue';
import {
  NCard,
  NSpace,
  NButton,
  NInput,
  NTree,
  NModal,
  NForm,
  NFormItem,
  NInputNumber,
  NSwitch,
  NPopconfirm,
  NEmpty,
  NTooltip,
  NTag,
  NIcon,
  NSelect,
  useMessage,
  type TreeOption,
  type TreeDropInfo,
  type FormInst,
  type FormRules
} from 'naive-ui';
import { Icon } from '@iconify/vue';
import { useMenuStore } from '@/store/modules/menu';
import type { MenuItem, MenuFormData, MenuTreeNode } from '@/typings/api/menu';

defineOptions({
  name: 'MenuManage'
});

const message = useMessage();
const menuStore = useMenuStore();

// 搜索关键词
const searchKey = ref('');

// 选中的菜单 key
const selectedKeys = ref<string[]>([]);
const checkedKeys = ref<string[]>([]);

// 弹窗相关
const showModal = ref(false);
const modalType = ref<'create' | 'edit'>('create');
const formRef = ref<FormInst | null>(null);

// 表单数据
const formData = ref<MenuFormData>({
  key: '',
  label: '',
  icon: '',
  path: '',
  routeName: '',
  order: 0,
  hideInMenu: false,
  fixed: false,
  parentKey: '',
  i18nKey: '',
  href: '',
  newWindow: false,
  enabled: true
});

// 表单验证规则
const formRules: FormRules = {
  key: [{ required: true, message: '请输入菜单标识', trigger: 'blur' }],
  label: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
  order: [{ required: true, type: 'number', message: '请输入排序', trigger: 'blur' }]
};

// 父级菜单选项
const parentOptions = computed(() => {
  const buildOptions = (nodes: MenuTreeNode[], level = 0): { label: string; value: string }[] => {
    const options: { label: string; value: string }[] = [];
    nodes.forEach(node => {
      const prefix = '　'.repeat(level);
      options.push({
        label: `${prefix}${node.label}`,
        value: node.key
      });
      if (node.children && node.children.length > 0) {
        options.push(...buildOptions(node.children, level + 1));
      }
    });
    return options;
  };

  return [
    { label: '顶级菜单', value: '' },
    ...buildOptions(menuStore.menuTree)
  ];
});

// 过滤后的菜单树
const filteredMenuTree = computed(() => {
  if (!searchKey.value) {
    return menuStore.menuTree;
  }

  const filterNodes = (nodes: MenuTreeNode[]): MenuTreeNode[] => {
    return nodes.reduce((acc: MenuTreeNode[], node) => {
      const labelMatch = node.label.toLowerCase().includes(searchKey.value.toLowerCase());
      const keyMatch = node.key.toLowerCase().includes(searchKey.value.toLowerCase());
      const pathMatch = node.path?.toLowerCase().includes(searchKey.value.toLowerCase());

      if (labelMatch || keyMatch || pathMatch) {
        acc.push(node);
      } else if (node.children && node.children.length > 0) {
        const filteredChildren = filterNodes(node.children);
        if (filteredChildren.length > 0) {
          acc.push({ ...node, children: filteredChildren });
        }
      }

      return acc;
    }, []);
  };

  return filterNodes(menuStore.menuTree);
});

// 将 MenuTreeNode 转换为 TreeOption
function convertToTreeOption(node: MenuTreeNode): TreeOption {
  return {
    key: node.key,
    label: node.label,
    children: node.children?.map(convertToTreeOption),
    prefix: () =>
      h(NIcon, null, {
        default: () =>
          h(Icon, {
            icon: node.children && node.children.length > 0 ? 'mdi:folder-outline' : 'mdi:file-document-outline'
          })
      }),
    suffix: () =>
      h(NSpace, { size: 'small' }, () => [
        node.enabled
          ? h(NTag, { size: 'small', type: 'success' }, () => '启用')
          : h(NTag, { size: 'small', type: 'default' }, () => '禁用'),
        node.hideInMenu
          ? h(NTag, { size: 'small', type: 'warning' }, () => '隐藏')
          : null,
        h(NSpace, { size: 'small' }, () => [
          h(
            NTooltip,
            {},
            {
              trigger: () =>
                h(
                  NButton,
                  {
                    size: 'tiny',
                    text: true,
                    onClick: (e: Event) => {
                      e.stopPropagation();
                      handleAddChild(node);
                    }
                  },
                  { icon: () => h(NIcon, null, { default: () => h(Icon, { icon: 'mdi:plus' }) }) }
                ),
              default: () => '添加子菜单'
            }
          ),
          h(
            NTooltip,
            {},
            {
              trigger: () =>
                h(
                  NButton,
                  {
                    size: 'tiny',
                    text: true,
                    onClick: (e: Event) => {
                      e.stopPropagation();
                      handleEdit(node);
                    }
                  },
                  { icon: () => h(NIcon, null, { default: () => h(Icon, { icon: 'mdi:pencil-outline' }) }) }
                ),
              default: () => '编辑'
            }
          ),
          h(
            NTooltip,
            {},
            {
              trigger: () =>
                h(
                  NPopconfirm,
                    { onPositiveClick: () => handleDelete(node.key) },
                    {
                      trigger: () =>
                        h(
                          NButton,
                          {
                            size: 'tiny',
                            text: true,
                            type: 'error',
                            onClick: (e: Event) => e.stopPropagation()
                          },
                          { icon: () => h(NIcon, null, { default: () => h(Icon, { icon: 'mdi:delete-outline' }) }) }
                        ),
                      default: () => '确定删除此菜单吗？'
                    }
                ),
              default: () => '删除'
            }
          )
        ])
      ])
  };
}

// 树选项
const treeOptions = computed<TreeOption[]>(() => {
  return filteredMenuTree.value.map(convertToTreeOption);
});

// 打开新建弹窗
function handleCreate() {
  modalType.value = 'create';
  formData.value = {
    key: '',
    label: '',
    icon: '',
    path: '',
    routeName: '',
    order: 0,
    hideInMenu: false,
    fixed: false,
    parentKey: '',
    i18nKey: '',
    href: '',
    newWindow: false,
    enabled: true
  };
  showModal.value = true;
}

// 添加子菜单
function handleAddChild(node: MenuTreeNode) {
  modalType.value = 'create';
  formData.value = {
    key: '',
    label: '',
    icon: '',
    path: '',
    routeName: '',
    order: 0,
    hideInMenu: false,
    fixed: false,
    parentKey: node.key,
    i18nKey: '',
    href: '',
    newWindow: false,
    enabled: true
  };
  showModal.value = true;
}

// 打开编辑弹窗
function handleEdit(node: MenuItem) {
  modalType.value = 'edit';
  formData.value = {
    key: node.key,
    label: node.label,
    icon: node.icon || '',
    path: node.path || '',
    routeName: node.routeName || '',
    order: node.order,
    hideInMenu: node.hideInMenu || false,
    fixed: node.fixed || false,
    parentKey: node.parentKey || '',
    i18nKey: node.i18nKey || '',
    href: node.href || '',
    newWindow: node.newWindow || false,
    enabled: node.enabled
  };
  showModal.value = true;
}

// 删除菜单
async function handleDelete(key: string) {
  await menuStore.deleteMenu(key);
}

// 批量删除
async function handleBatchDelete() {
  if (checkedKeys.value.length === 0) {
    message.warning('请选择要删除的菜单');
    return;
  }

  const { successKeys, failedKeys } = await menuStore.batchDeleteMenus(checkedKeys.value as string[]);
  if (successKeys.length > 0) {
    message.success(`成功删除 ${successKeys.length} 个菜单`);
    checkedKeys.value = [];
  }
  if (failedKeys.length > 0) {
    message.error(`${failedKeys.length} 个菜单删除失败`);
  }
}

// 提交表单
async function handleSubmit() {
  await formRef.value?.validate();

  if (modalType.value === 'create') {
    const result = await menuStore.createMenu(formData.value);
    if (result) {
      showModal.value = false;
    }
  } else {
    const result = await menuStore.updateMenu(formData.value.key, formData.value);
    if (result) {
      showModal.value = false;
    }
  }
}

// 拖拽处理
async function handleDrop({ node, dragNode, dropPosition }: TreeDropInfo) {
  const dragKey = dragNode.key as string;
  const dropKey = node.key as string;

  let targetParentKey: string | null = null;

  if (dropPosition === 'inside') {
    // 放入节点内部
    targetParentKey = dropKey;
  } else {
    // 放在节点前后，目标父级为该节点的父级
    const dropNode = findMenuByKey(dropKey, menuStore.menuTree);
    targetParentKey = dropNode?.parentKey || null;
  }

  // 不能将节点拖拽到自己的子节点
  if (isChildOf(dragKey, dropKey, menuStore.menuTree)) {
    message.warning('不能将菜单移动到自己的子菜单下');
    return;
  }

  await menuStore.moveMenu(dragKey, targetParentKey);
}

// 查找菜单
function findMenuByKey(key: string, nodes: MenuTreeNode[]): MenuItem | null {
  for (const node of nodes) {
    if (node.key === key) {
      return node;
    }
    if (node.children) {
      const found = findMenuByKey(key, node.children);
      if (found) return found;
    }
  }
  return null;
}

// 判断是否为子节点
function isChildOf(parentKey: string, childKey: string, nodes: MenuTreeNode[]): boolean {
  for (const node of nodes) {
    if (node.key === parentKey && node.children) {
      if (node.children.some(child => child.key === childKey)) {
        return true;
      }
      if (isChildOf(parentKey, childKey, node.children)) {
        return true;
      }
    }
    if (node.children && isChildOf(parentKey, childKey, node.children)) {
      return true;
    }
  }
  return false;
}

// 切换启用状态
async function handleToggleEnabled(key: string, enabled: boolean) {
  await menuStore.toggleMenuEnabled(key, enabled);
}

// 加载数据
onMounted(() => {
  menuStore.loadMenus();
});
</script>

<template>
  <div class="menu-manage-page">
    <NCard title="菜单管理">
      <template #header-extra>
        <NSpace>
          <NButton type="primary" @click="handleCreate">
            新建菜单
          </NButton>
        </NSpace>
      </template>

      <div class="menu-content">
        <!-- 工具栏 -->
        <div class="toolbar">
          <NInput
            v-model:value="searchKey"
            placeholder="搜索菜单名称、标识或路径"
            clearable
            style="width: 300px"
          >
            <template #prefix>
              <NIcon>
                <Icon icon="mdi:magnify" />
              </NIcon>
            </template>
          </NInput>

          <NSpace>
            <NPopconfirm @positive-click="handleBatchDelete">
              <template #trigger>
                <NButton
                  type="error"
                  :disabled="checkedKeys.length === 0"
                >
                  批量删除 ({{ checkedKeys.length }})
                </NButton>
              </template>
              确定删除选中的 {{ checkedKeys.length }} 个菜单吗？
            </NPopconfirm>
          </NSpace>
        </div>

        <!-- 菜单树 -->
        <div class="menu-tree-container">
          <NTree
            v-model:selected-keys="selectedKeys"
            v-model:checked-keys="checkedKeys"
            :data="treeOptions"
            :block-line="true"
            :cascade="true"
            :checkable="true"
            :draggable="true"
            :expand-on-click="true"
            selectable
            drop-position="all"
            @drop="handleDrop"
          >
            <template #empty>
              <NEmpty description="暂无菜单数据" style="margin: 20px 0">
                <template #extra>
                  <NButton type="primary" size="small" @click="handleCreate">
                    新建菜单
                  </NButton>
                </template>
              </NEmpty>
            </template>
          </NTree>
        </div>
      </div>
    </NCard>

    <!-- 新建/编辑弹窗 -->
    <NModal
      v-model:show="showModal"
      preset="card"
      :title="modalType === 'create' ? '新建菜单' : '编辑菜单'"
      :style="{ width: '600px' }"
      :mask-closable="false"
    >
      <NForm
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-placement="left"
        label-width="100px"
      >
        <NFormItem label="父级菜单" path="parentKey">
          <NSelect
            v-model:value="formData.parentKey"
            :options="parentOptions"
            placeholder="选择父级菜单（可选）"
            clearable
          />
        </NFormItem>

        <NFormItem label="菜单标识" path="key">
          <NInput
            v-model:value="formData.key"
            :disabled="modalType === 'edit'"
            placeholder="唯一标识，如：manage_menu"
          />
        </NFormItem>

        <NFormItem label="菜单名称" path="label">
          <NInput v-model:value="formData.label" placeholder="菜单显示名称" />
        </NFormItem>

        <NFormItem label="菜单图标" path="icon">
          <NInput v-model:value="formData.icon" placeholder="图标名称，如：mdi:cog" />
        </NFormItem>

        <NFormItem label="路由路径" path="path">
          <NInput v-model:value="formData.path" placeholder="路由路径，如：/manage/menu" />
        </NFormItem>

        <NFormItem label="路由名称" path="routeName">
          <NInput v-model:value="formData.routeName" placeholder="路由名称，如：manage_menu" />
        </NFormItem>

        <NFormItem label="国际化Key" path="i18nKey">
          <NInput v-model:value="formData.i18nKey" placeholder="国际化标识，如：route.manage_menu" />
        </NFormItem>

        <NFormItem label="外部链接" path="href">
          <NInput v-model:value="formData.href" placeholder="外链地址（可选）" />
        </NFormItem>

        <NFormItem label="排序" path="order">
          <NInputNumber v-model:value="formData.order" :min="0" style="width: 100%" />
        </NFormItem>

        <NFormItem label="启用" path="enabled">
          <NSwitch v-model:value="formData.enabled" />
        </NFormItem>

        <NFormItem label="菜单中隐藏" path="hideInMenu">
          <NSwitch v-model:value="formData.hideInMenu" />
        </NFormItem>

        <NFormItem label="固定标签栏" path="fixed">
          <NSwitch v-model:value="formData.fixed" />
        </NFormItem>

        <NFormItem v-if="formData.href" label="新窗口打开" path="newWindow">
          <NSwitch v-model:value="formData.newWindow" />
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showModal = false">取消</NButton>
          <NButton type="primary" :loading="menuStore.loading" @click="handleSubmit">
            {{ modalType === 'create' ? '创建' : '保存' }}
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.menu-manage-page {
  padding: 16px;
  height: 100%;
}

.menu-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.menu-tree-container {
  border: 1px solid var(--n-border-color);
  border-radius: 4px;
  padding: 16px;
  max-height: calc(100vh - 300px);
  overflow-y: auto;
}
</style>
