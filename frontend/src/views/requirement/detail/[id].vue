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

// 简单的 Markdown 渲染（基础版）
function renderMarkdown(content: string): string {
  if (!content) return '';

  let html = content
    // 转义 HTML
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    // 标题
    .replace(/^### (.*$)/gm, '<h3>$1</h3>')
    .replace(/^## (.*$)/gm, '<h2>$1</h2>')
    .replace(/^# (.*$)/gm, '<h1>$1</h1>')
    // 粗体和斜体
    .replace(/\*\*\*(.*?)\*\*\*/g, '<strong><em>$1</em></strong>')
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    // 代码块
    .replace(/```([\s\S]*?)```/g, '<pre><code>$1</code></pre>')
    // 行内代码
    .replace(/`(.*?)`/g, '<code>$1</code>')
    // 链接
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank">$1</a>')
    // 列表
    .replace(/^\- (.*$)/gm, '<li>$1</li>')
    .replace(/^\d+\. (.*$)/gm, '<li>$1</li>')
    // 段落
    .replace(/\n\n/g, '</p><p>')
    .replace(/\n/g, '<br>');

  return `<div class="markdown-content"><p>${html}</p></div>`;
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
      <!-- 顶部操作栏 -->
      <NCard class="mb-4">
        <NSpace justify="space-between" align="center">
          <NSpace align="center">
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
          </NSpace>
          <NSpace>
            <NButton v-if="!isNew" @click="togglePreview">
              <template #icon>
                <span :class="previewMode ? 'i-mdi:pencil' : 'i-mdi:eye'"></span>
              </template>
              {{ previewMode ? '编辑' : '预览' }}
            </NButton>
            <NButton type="primary" @click="handleSave">
              <template #icon>
                <span class="i-mdi:content-save"></span>
              </template>
              保存
            </NButton>
          </NSpace>
        </NSpace>
      </NCard>

      <!-- 主内容区 -->
      <NGrid :cols="24" :x-gap="16">
        <!-- 左侧：编辑/预览 -->
        <NGi :span="18">
          <NCard title="需求内容">
            <!-- 预览模式 -->
            <div v-if="previewMode" class="min-h-96">
              <div v-if="formData.content" class="prose max-w-none" v-html="renderMarkdown(formData.content)"></div>
              <NEmpty v-else description="暂无内容" />
            </div>

            <!-- 编辑模式 -->
            <div v-else>
              <NFormItem label="标题" required>
                <NInput v-model:value="formData.title" placeholder="请输入需求标题" />
              </NFormItem>

              <NFormItem label="内容（支持 Markdown 格式）">
                <NInput
                  v-model:value="formData.content"
                  type="textarea"
                  placeholder="请输入需求内容，支持 Markdown 格式"
                  :autosize="{
                    minRows: 20,
                    maxRows: 40
                  }"
                />
              </NFormItem>
            </div>
          </NCard>
        </NGi>

        <!-- 右侧：属性和文档 -->
        <NGi :span="6">
          <NSpace vertical :size="16">
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
.markdown-content {
  line-height: 1.8;
}

.markdown-content :deep(h1) {
  font-size: 1.75rem;
  font-weight: 600;
  margin: 1rem 0;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.markdown-content :deep(h2) {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0.875rem 0;
}

.markdown-content :deep(h3) {
  font-size: 1.25rem;
  font-weight: 500;
  margin: 0.75rem 0;
}

.markdown-content :deep(code) {
  background-color: #f3f4f6;
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.875rem;
}

.markdown-content :deep(pre) {
  background-color: #1f2937;
  color: #e5e7eb;
  padding: 1rem;
  border-radius: 0.5rem;
  overflow-x: auto;
  margin: 1rem 0;
}

.markdown-content :deep(pre code) {
  background-color: transparent;
  padding: 0;
  color: inherit;
}

.markdown-content :deep(a) {
  color: #3b82f6;
  text-decoration: none;
}

.markdown-content :deep(a:hover) {
  text-decoration: underline;
}

.markdown-content :deep(li) {
  margin-left: 1.5rem;
  list-style-type: disc;
}

.prose {
  max-width: none;
}
</style>
