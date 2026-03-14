<script setup lang="ts">
import { computed, ref } from 'vue'
import { NBadge, NIcon, NPopover, NList, NListItem, NEmpty, NSpin } from 'naive-ui'
import { useMessageStore } from '@/store/modules/message'
import { useRouter } from 'vue-router'
import type { Message } from '@/typings/api/message'

const messageStore = useMessageStore()
const router = useRouter()

const showPopover = ref(false)
const loading = ref(false)

// 未读消息数量
const unreadCount = computed(() => messageStore.unreadCount)

// 消息列表
const messages = computed(() => messageStore.messages)

// 跳转到消息列表
function goToMessageList() {
  showPopover.value = false
  router.push('/message/list')
}

// 查看消息详情
async function viewMessage(message: Message) {
  loading.value = true
  try {
    await messageStore.markAsRead(message.id)
    if (message.status === 'success') {
      // 可以跳转到任务详情
      router.push(`/requirement/task-detail/${message.taskId}`)
    }
  } finally {
    loading.value = false
  }
}

// 关闭弹窗
function handleClose() {
  showPopover.value = false
}

// 获取消息类型文本
function getMessageTypeText(type: string): string {
  const typeMap: Record<string, string> = {
    expand_task: '任务拆分',
    regenerate_subtask: '重写子任务'
  }
  return typeMap[type] || type
}

// 获取消息状态文本
function getMessageStatusText(status: string): string {
  const statusMap: Record<string, string> = {
    pending: '等待中',
    processing: '处理中',
    success: '成功',
    failed: '失败'
  }
  return statusMap[status] || status
}

// 获取消息状态颜色
function getMessageStatusColor(status: string): string {
  const colorMap: Record<string, string> = {
    pending: 'warning',
    processing: 'info',
    success: 'success',
    failed: 'error'
  }
  return colorMap[status] || 'default'
}
</script>

<template>
  <NPopover
    v-model:show="showPopover"
    trigger="click"
    placement="bottom-end"
    :show-arrow="false"
  >
    <template #trigger>
      <div class="message-bell" @click.stop.prevent>
        <span class="bell-icon" :class="showPopover ? 'active' : ''">
          <svg v-if="unreadCount > 0" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
            <path d="M21 19v1H3v-1l2-2v-6c0-3.1 2.03-5.83 5-6.71V4a2 2 0 0 1 4 0v.29c2.97.88 5 3.61 5 6.71v6l2 2zm-7 2a2 2 0 0 1-4 0"/>
          </svg>
          <svg v-else xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="20" height="20">
            <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
            <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
          </svg>
        </span>
        <NBadge
          v-if="unreadCount > 0"
          :value="unreadCount"
          :max="99"
          type="error"
          class="badge"
        />
      </div>
    </template>
    <div class="message-list-wrapper">
      <NSpin :show="loading">
        <div v-if="messages.length > 0" class="message-list">
          <NList hoverable clickable>
            <NListItem v-for="message in messages" :key="message.id" @click="viewMessage(message)">
              <div class="message-item">
                <div class="message-header">
                  <span class="message-type">{{ getMessageTypeText(message.type) }}</span>
                  <span class="message-title">{{ message.title }}</span>
                </div>
                <div class="message-content">
                  <span class="message-status" :class="`status-${message.status}`">
                    {{ getMessageStatusText(message.status) }}
                  </span>
                  <span class="message-time">{{ new Date(message.createdAt).toLocaleString() }}</span>
                </div>
              </div>
            </NListItem>
          </NList>
        </div>
        <NEmpty v-else description="暂无消息" />
      </NSpin>
    </div>
  </NPopover>
</template>

<style scoped>
.message-bell {
  cursor: pointer;
  position: relative;
  display: flex;
  align-items: center;
  padding: 8px;
}

.message-bell:hover {
  background-color: rgba(0, 0, 0, 0.05);
  border-radius: 4px;
}

.bell-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #666;
  transition: color 0.2s;
}

.bell-icon.active {
  color: #faad14;
}

.bell-icon svg {
  width: 20px;
  height: 20px;
}

.bell-icon svg[fill="currentColor"] {
  color: #faad14;
}

.badge {
  position: absolute;
  top: 2px;
  right: 2px;
}

.message-list-wrapper {
  max-height: 400px;
  overflow-y: auto;
}

.message-list {
  min-width: 300px;
}

.message-item {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.message-item:hover {
  background-color: #f5f5f5;
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.message-type {
  font-size: 12px;
  padding: 4px 8px;
  border-radius: 4px;
  font-weight: 500;
  background-color: #e8f4fe;
  color: #1890ff;
}

.message-title {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.message-content {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: #666;
}

.message-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.status-pending {
  background-color: #fff3e6;
  border: 1px solid #faad14;
  color: #faad14;
}

.status-processing {
  background-color: #e8f4fe;
  border: 1px solid #1890ff;
  color: #1890ff;
}

.status-success {
  background-color: #e8f5e2;
  border: 1px solid #18a058;
  color: #18a058;
}

.status-failed {
  background-color: #fff0f0;
  border: 1px solid #f0a020;
  color: #f0a020;
}

.message-time {
  font-size: 12px;
  color: #999;
}
</style>
