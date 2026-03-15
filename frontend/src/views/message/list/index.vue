<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { NButton, NCard, NEmpty, NList, NListItem, NPopconfirm, NProgress, NSpace, NSpin, NTag } from 'naive-ui';
import { useMessageStore } from '@/store/modules/message';
import type { Message } from '@/typings/api/message';

defineOptions({
  name: 'MessageList'
});

const messageStore = useMessageStore();
const router = useRouter();

const loading = ref(false);
const pollingTimer = ref<NodeJS.Timeout | null>(null);
const pollingInterval = 5000; // 5秒轮询一次

// 获取处理中的消息
const processingMessages = computed(() => {
  return messageStore.messages.filter(m => m.status === 'pending' || m.status === 'processing');
});

// 是否有处理中的消息
const hasProcessingMessages = computed(() => processingMessages.value.length > 0);

// 加载消息列表
async function loadMessages() {
  loading.value = true;
  try {
    await messageStore.loadMessages();
  } finally {
    loading.value = false;
  }
}

// 查看消息详情
async function viewMessage(message: Message) {
  // 如果消息还在处理中，不跳转
  if (message.status === 'pending' || message.status === 'processing') {
    window.$message?.info('消息正在处理中，请稍后');
    return;
  }

  await messageStore.markAsRead(message.id);
  router.push(`/requirement/task-detail/${message.taskId}`);
}

// 删除消息
async function handleDelete(messageId: number) {
  await messageStore.deleteMessageById(messageId);
}

// 获取消息类型文本
function getMessageTypeText(type: string): string {
  const typeMap: Record<string, string> = {
    expand_task: '任务拆分',
    regenerate_subtask: '重写子任务'
  };
  return typeMap[type] || type;
}

// 获取消息状态文本
function getMessageStatusText(status: string): string {
  const statusMap: Record<string, string> = {
    pending: '等待中',
    processing: '处理中',
    success: '已完成',
    failed: '失败'
  };
  return statusMap[status] || status;
}

// 获取消息状态类型
function getMessageStatusType(status: string): 'default' | 'warning' | 'info' | 'success' | 'error' {
  const typeMap: Record<string, 'default' | 'warning' | 'info' | 'success' | 'error'> = {
    pending: 'warning',
    processing: 'info',
    success: 'success',
    failed: 'error'
  };
  return typeMap[status] || 'default';
}

// 记录已通知的消息ID，避免重复通知
const notifiedMessageIds = ref<Set<number>>(new Set());

// 轮询检查消息状态
function startPolling() {
  if (pollingTimer.value) return;

  const poll = async () => {
    await messageStore.loadMessages();

    // 检查是否有状态变化
    for (const message of messageStore.messages) {
      // 只处理未通知过的消息
      if (notifiedMessageIds.value.has(message.id)) continue;

      if (message.status === 'success') {
        notifiedMessageIds.value.add(message.id);
        window.$message?.success(message.resultSummary || '任务处理完成', { duration: 5000 });
      } else if (message.status === 'failed') {
        notifiedMessageIds.value.add(message.id);
        window.$message?.error(`任务处理失败: ${message.errorMessage || '未知错误'}`, { duration: 5000 });
      }
    }

    // 如果没有处理中的消息，停止轮询
    if (!hasProcessingMessages.value) {
      stopPolling();
    }
  };

  // 启动定时轮询
  pollingTimer.value = setInterval(poll, pollingInterval);
}

// 停止轮询
function stopPolling() {
  if (pollingTimer.value) {
    clearInterval(pollingTimer.value);
    pollingTimer.value = null;
  }
}

onMounted(async () => {
  await loadMessages();

  // 如果有处理中的消息，启动轮询
  if (hasProcessingMessages.value) {
    startPolling();
  }
});

onUnmounted(() => {
  stopPolling();
});
</script>

<template>
  <div class="message-list-page p-4">
    <NCard title="消息列表">
      <template #header-extra>
        <NSpace>
          <NButton v-if="hasProcessingMessages" @click="startPolling">
            {{ pollingTimer ? '轮询中...' : '开始轮询' }}
          </NButton>
          <NButton @click="loadMessages">刷新</NButton>
        </NSpace>
      </template>

      <!-- 处理中的消息提示 -->
      <div v-if="hasProcessingMessages" class="processing-tip">
        <NSpace align="center">
          <NSpin size="small" />
          <span>有 {{ processingMessages.length }} 条消息正在处理中，页面会自动刷新状态...</span>
        </NSpace>
      </div>

      <NSpin :show="loading">
        <div v-if="messageStore.messages.length > 0" class="message-list-container">
          <NList hoverable clickable>
            <NListItem v-for="message in messageStore.messages" :key="message.id" @click="viewMessage(message)">
              <template #prefix>
                <!-- 处理中显示进度指示 -->
                <NSpin v-if="message.status === 'processing' || message.status === 'pending'" size="small" />
              </template>
              <div class="message-content">
                <div class="message-header">
                  <NSpace align="center" :size="8">
                    <span class="message-id">#{{ message.id }}</span>
                    <NTag type="info" size="small">{{ getMessageTypeText(message.type) }}</NTag>
                    <NTag :type="getMessageStatusType(message.status)" size="small">
                      {{ getMessageStatusText(message.status) }}
                    </NTag>
                    <span
                      v-if="!message.isRead && message.status !== 'processing' && message.status !== 'pending'"
                      class="unread-dot"
                    ></span>
                  </NSpace>
                </div>
                <div class="message-title">{{ message.title }}</div>
                <div class="message-footer">
                  <span class="message-time">{{ new Date(message.createdAt).toLocaleString() }}</span>
                  <span v-if="message.errorMessage" class="message-error">{{ message.errorMessage }}</span>
                  <span v-else-if="message.resultSummary" class="message-result">{{ message.resultSummary }}</span>
                </div>
              </div>
              <template #suffix>
                <NPopconfirm @positive-click="handleDelete(message.id)">
                  <template #trigger>
                    <NButton text type="error" size="small" @click.stop>删除</NButton>
                  </template>
                  确定要删除这条消息吗？
                </NPopconfirm>
              </template>
            </NListItem>
          </NList>
        </div>
        <NEmpty v-else description="暂无消息" />
      </NSpin>
    </NCard>
  </div>
</template>

<style scoped>
.message-list-page {
  max-width: 1200px;
  margin: 0 auto;
}

.processing-tip {
  padding: 12px 16px;
  margin-bottom: 16px;
  background-color: #e6f7ff;
  border-radius: 4px;
  border: 1px solid #91d5ff;
}

.message-list-container {
  border-radius: 4px;
  overflow: hidden;
}

.message-content {
  flex: 1;
  min-width: 0;
}

.message-header {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.unread-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #f0a020;
}

.message-id {
  font-size: 12px;
  color: #999;
  font-family: monospace;
}

.message-title {
  font-size: 15px;
  font-weight: 500;
  color: #333;
  margin-bottom: 6px;
  word-wrap: break-word;
}

.message-footer {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #999;
}

.message-error {
  color: #f0a020;
}

.message-result {
  color: #18a058;
}
</style>
