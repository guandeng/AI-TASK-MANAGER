<script setup lang="tsx">
import { onMounted, ref, computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  NButton,
  NCard,
  NSpace,
  NInput,
  NSelect,
  NUpload,
  NUploadDragger,
  NIcon,
  NText,
  NDataTable,
  NTag,
  NPopconfirm,
  NSpin,
  NGrid,
  NGi,
  NFormItem,
  NForm,
  NTabs,
  NTabPane,
  NDivider,
  NEmpty
} from 'naive-ui';
import type { UploadFileInfo } from 'naive-ui';
import { MdEditor, MdPreview } from 'md-editor-v3';
import 'md-editor-v3/lib/style.css';
import { useRequirementStore } from '@/store/modules/requirement';
import type { RequirementStatus, RequirementPriority, RequirementDocument } from '@/typings/api/requirement';
import { getDocumentDownloadUrl } from '@/service/api/requirement';

defineOptions({
  name: 'RequirementDetail'
});

const route = useRoute();
const router = useRouter();
const requirementStore = useRequirementStore();

// 判断是否为新建模式
const isNew = computed(() => route.params.id === 'new');
const requirementId = computed(() => Number(route.params.id));

// 表单数据
const formData = ref({
  title: '',
  content: '',
  status: 'draft' as RequirementStatus,
  priority: 'medium' as RequirementPriority,
  assignee: ''
});

// 预览模式
const previewMode = ref(false);

// Markdown 编辑器工具栏配置
const editorToolbars = [
  'bold',
  'underline',
  'italic',
  '-',
  'strikeThrough',
  'title',
  'sub',
  'sup',
  'quote',
  'unorderedList',
  'orderedList',
  'task',
  '-',
  'codeRow',
  'code',
  'link',
  'image',
  'table',
  'mermaid',
  '-',
  'revoke',
  'next',
  '=',
  'pageFullscreen',
  'fullscreen',
  'preview',
  'catalog'
];

// 处理编辑器图片上传
async function handleUploadImg(files: File[], callback: (urls: string[]) => void) {
  if (isNew.value) {
    window.$message?.warning('请先保存需求后再上传图片');
    callback([]);
    return;
  }

  try {
    const urls: string[] = [];
    for (const file of files) {
      const { data, error } = await requirementStore.uploadRequirementDocument(requirementId.value, file);
      if (!error && data) {
        // 返回图片访问 URL
        urls.push(getDocumentDownloadUrl(requirementId.value, data.id));
      }
    }
    callback(urls);
  } catch {
    window.$message?.error('图片上传失败');
    callback([]);
  }
}

// 状态选项
const statusOptions = [
  { label: '草稿', value: 'draft' },
  { label: '进行中', value: 'active' },
  { label: '已完成', value: 'completed' },
  { label: '已归档', value: 'archived' }
];

// 优先级选项
const priorityOptions = [
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

// 文档表格列
const documentColumns = [
  {
    title: '文件名',
    key: 'name',
    ellipsis: { tooltip: true }
  },
  {
    title: '大小',
    key: 'size',
    width: 120,
    render(row: RequirementDocument) {
      return formatFileSize(row.size);
    }
  },
  {
    title: '上传时间',
    key: 'createdAt',
    width: 180,
    render(row: RequirementDocument) {
      return new Date(row.createdAt).toLocaleString('zh-CN');
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    render(row: RequirementDocument) {
      return (
        <NSpace>
          <NButton size="small" type="primary" text onClick={() => handleDownloadDocument(row)}>
            下载
          </NButton>
          <NPopconfirm onPositiveClick={() => handleDeleteDocument(row.id)}>
            {{
              trigger: () => <NButton size="small" type="error" text>删除</NButton>,
              default: () => '确定要删除这个文档吗？'
            }}
          </NPopconfirm>
        </NSpace>
      );
    }
  }
];

// 格式化文件大小
function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// 加载需求详情
async function loadDetail() {
  if (!isNew.value && requirementId.value) {
    const { data, error } = await requirementStore.loadRequirementDetail(requirementId.value);
    if (!error && data) {
      formData.value = {
        title: data.title,
        content: data.content,
        status: data.status,
        priority: data.priority,
        assignee: data.assignee || ''
      };
    }
  }
}

// 保存需求
async function handleSave() {
  if (!formData.value.title.trim()) {
    window.$message?.warning('请输入需求标题');
    return;
  }

  if (isNew.value) {
    const { data, error } = await requirementStore.createNewRequirement(formData.value);
    if (!error && data) {
      window.$message?.success('创建成功');
      // 创建成功后返回列表页
      router.push('/requirement/list');
    } else {
      window.$message?.error('创建失败');
    }
  } else {
    const { error } = await requirementStore.updateRequirementById(requirementId.value, formData.value);
    if (!error) {
      window.$message?.success('保存成功');
    } else {
      window.$message?.error('保存失败');
    }
  }
}

// 返回列表
function handleBack() {
  router.push('/requirement/list');
}

// 切换预览模式
function togglePreview() {
  previewMode.value = !previewMode.value;
}

// 文件上传
async function handleUpload({ file }: { file: UploadFileInfo }) {
  if (isNew.value) {
    window.$message?.warning('请先保存需求后再上传文档');
    return false;
  }

  if (!file.file) return false;

  const { error } = await requirementStore.uploadRequirementDocument(requirementId.value, file.file);
  if (!error) {
    window.$message?.success('上传成功');
  } else {
    window.$message?.error('上传失败');
  }
  return false; // 阻止默认上传行为
}

// 下载文档
function handleDownloadDocument(doc: RequirementDocument) {
  const url = getDocumentDownloadUrl(requirementId.value, doc.id);
  window.open(url, '_blank');
}

// 删除文档
async function handleDeleteDocument(docId: number) {
  const { error } = await requirementStore.deleteRequirementDocument(requirementId.value, docId);
  if (!error) {
    window.$message?.success('删除成功');
  } else {
    window.$message?.error('删除失败');
  }
}

// 监听路由变化
watch(
  () => route.params.id,
  () => {
    if (route.params.id) {
      if (isNew.value) {
        formData.value = {
          title: '',
          content: '',
          status: 'draft',
          priority: 'medium',
          assignee: ''
        };
        requirementStore.clearCurrentRequirement();
      } else {
        loadDetail();
      }
    }
  },
  { immediate: true }
);

onMounted(() => {
  if (!isNew.value) {
    loadDetail();
  }
});
</script>

<template>
  <div class="h-full overflow-auto p-4">
    <NSpin :show="requirementStore.detailLoading">
      <!-- 主内容区 -->
      <NGrid :cols="24" :x-gap="16">
        <!-- 左侧：编辑/预览 -->
        <NGi :span="18">
          <NCard>
            <!-- 标题栏 -->
            <template #header>
              <div class="flex items-center gap-3">
                <NButton text @click="handleBack">
                  <template #icon>
                    <span class="i-mdi:arrow-left text-lg"></span>
                  </template>
                </NButton>
                <span class="text-lg font-medium">
                  {{ isNew ? '新建需求' : '编辑需求' }}
                </span>
                <NTag v-if="!isNew && requirementStore.currentRequirement" :type="statusColors[requirementStore.currentRequirement.status] as any">
                  {{ statusText[requirementStore.currentRequirement.status] }}
                </NTag>
              </div>
            </template>

            <!-- 预览模式 -->
            <div v-if="previewMode" class="min-h-96">
              <MdPreview v-if="formData.content" :model-value="formData.content" class="markdown-preview-wrapper" />
              <NEmpty v-else description="暂无内容" />
            </div>

            <!-- 编辑模式 -->
            <div v-else class="flex flex-col gap-4">
              <NFormItem label="标题" required>
                <NInput v-model:value="formData.title" placeholder="请输入需求标题" />
              </NFormItem>

              <NFormItem label="内容">
                <MdEditor
                  v-model="formData.content"
                  language="zh-CN"
                  :toolbars="editorToolbars"
                  placeholder="请输入需求内容，支持 Markdown 格式"
                  :style="{ height: '500px' }"
                  @onUploadImg="handleUploadImg"
                />
              </NFormItem>
            </div>
          </NCard>
        </NGi>

        <!-- 右侧：操作按钮和属性 -->
        <NGi :span="6">
          <NSpace vertical :size="16">
            <!-- 操作按钮 -->
            <NCard title="操作" size="small">
              <NSpace vertical>
                <NButton type="primary" @click="handleSave" block>
                  <template #icon>
                    <span class="i-mdi:content-save"></span>
                  </template>
                  保存
                </NButton>
                <NButton v-if="!isNew" @click="togglePreview" block>
                  <template #icon>
                    <span :class="previewMode ? 'i-mdi:pencil' : 'i-mdi:eye'"></span>
                  </template>
                  {{ previewMode ? '编辑' : '预览' }}
                </NButton>
                <NButton @click="handleBack" block>
                  <template #icon>
                    <span class="i-mdi:arrow-left"></span>
                  </template>
                  返回列表
                </NButton>
              </NSpace>
            </NCard>

            <!-- 基本信息 -->
            <NCard title="基本信息" size="small">
              <NSpace vertical>
                <NFormItem label="状态">
                  <NSelect v-model:value="formData.status" :options="statusOptions" />
                </NFormItem>
                <NFormItem label="优先级">
                  <NSelect v-model:value="formData.priority" :options="priorityOptions" />
                </NFormItem>
                <NFormItem label="负责人">
                  <NInput v-model:value="formData.assignee" placeholder="请输入负责人" />
                </NFormItem>
              </NSpace>
            </NCard>

            <!-- 文档上传 -->
            <NCard title="相关文档" size="small">
              <template v-if="!isNew">
                <NUpload
                  :custom-request="handleUpload as any"
                  :show-file-list="false"
                  accept=".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.md,.png,.jpg,.jpeg"
                  multiple
                >
                  <NUploadDragger>
                    <div class="py-4">
                      <div class="text-2xl text-gray-400 mb-2">
                        <span class="i-mdi:cloud-upload"></span>
                      </div>
                      <NText class="text-sm">
                        点击或拖拽文件到此处上传
                      </NText>
                      <p class="text-xs text-gray-400 mt-1">
                        支持 PDF, Word, Excel, PPT, TXT, MD, 图片等格式
                      </p>
                    </div>
                  </NUploadDragger>
                </NUpload>

                <NDivider v-if="requirementStore.currentRequirement?.documents?.length" />

                <!-- 文档列表 -->
                <NDataTable
                  v-if="requirementStore.currentRequirement?.documents?.length"
                  :columns="documentColumns"
                  :data="requirementStore.currentRequirement?.documents || []"
                  size="small"
                  :pagination="false"
                />

                <NEmpty v-else-if="!requirementStore.currentRequirement?.documents?.length" description="暂无文档" size="small" class="mt-4" />
              </template>
              <template v-else>
                <NText depth="3" class="text-sm">
                  请先保存需求后再上传文档
                </NText>
              </template>
            </NCard>

            <!-- 时间信息 -->
            <NCard v-if="!isNew && requirementStore.currentRequirement" title="时间信息" size="small">
              <NSpace vertical :size="8">
                <div class="flex justify-between">
                  <span class="text-gray-500">创建时间</span>
                  <span>{{ new Date(requirementStore.currentRequirement.createdAt).toLocaleString('zh-CN') }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-500">更新时间</span>
                  <span>{{ new Date(requirementStore.currentRequirement.updatedAt).toLocaleString('zh-CN') }}</span>
                </div>
              </NSpace>
            </NCard>
          </NSpace>
        </NGi>
      </NGrid>
    </NSpin>
  </div>
</template>

<style scoped>
/* Markdown 预览容器样式 */
.markdown-preview-wrapper {
  border: none !important;
  background: transparent !important;
}

/* Markdown 编辑器样式覆盖 */
:deep(.md-editor) {
  --md-bk-color: transparent;
}

:deep(.md-editor-toolbar-wrapper) {
  border-radius: 8px 8px 0 0;
}

:deep(.md-editor-content) {
  border-radius: 0 0 8px 8px;
}

/* 确保编辑器在卡片内显示正常 */
:deep(.md-editor) {
  border: 1px solid #e0e0e6;
  border-radius: 8px;
}
</style>
