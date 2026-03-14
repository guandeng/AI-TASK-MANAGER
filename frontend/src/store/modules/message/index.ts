import { defineStore } from 'pinia';
import { ref } from 'vue';
import {
  fetchMessages,
  fetchUnreadCount,
  markMessageRead,
  deleteMessage
} from '@/service/api/message';
import type { Message, MessageListParams } from '@/typings/api/message';

export const useMessageStore = defineStore('message-store', () => {
  // 状态
  const messages = ref<Message[]>([]);
  const unreadCount = ref(0);
  const loading = ref(false);

  // 轮询定时器
  const pollingTimer = ref<NodeJS.Timeout | null>(null);
  // 轮询间隔（毫秒）
  const pollingInterval = 5000;

  // Actions
  async function loadMessages(params?: MessageListParams) {
    loading.value = true;
    try {
      const { data, error } = await fetchMessages(params);
      if (!error && data) {
        // 后端返回格式: { code: 0, message: "success", data: { list, total, page, pageSize } }
        const responseData = (data as any).data || data;

        if (Array.isArray(responseData)) {
          messages.value = responseData;
        } else if (responseData && 'list' in responseData) {
          messages.value = responseData.list || [];
          if (typeof responseData.total === 'number') {
            unreadCount.value = responseData.total;
          }
        } else if (responseData && 'messages' in responseData) {
          messages.value = responseData.messages || [];
          // 更新未读数量（如果返回了总数）
          if (typeof responseData.total === 'number') {
            unreadCount.value = responseData.total;
          } else {
            // 计算本地未读数量
            unreadCount.value = messages.value.filter(m => !m.isRead).length;
          }
        }
      }
    } catch (error) {
      console.error('Failed to load messages:', error);
      window.$message?.error('加载消息列表失败');
    } finally {
      loading.value = false;
    }
  }

  async function loadUnreadCount() {
    try {
      const { data, error } = await fetchUnreadCount();
      if (!error && data) {
        unreadCount.value = data.count;
      }
    } catch (error) {
      console.error('Failed to load unread count:', error);
    }
  }

  async function markAsRead(id: number) {
    try {
      await markMessageRead(id);
      // 更新本地状态
      const message = messages.value.find(m => m.id === id);
      if (message) {
        message.isRead = true;
        unreadCount.value = Math.max(0, unreadCount.value - 1);
      }
    } catch (error) {
      console.error('Failed to mark message as read:', error);
      window.$message?.error('标记消息已读失败');
    }
  }

  async function deleteMessageById(id: number) {
    try {
      await deleteMessage(id);
      messages.value = messages.value.filter(m => m.id !== id);
      unreadCount.value = Math.max(0, unreadCount.value - 1);
    } catch (error) {
      console.error('Failed to delete message:', error);
      window.$message?.error('删除消息失败');
    }
  }

  /**
   * 启动轮询检查消息状态
   * @param key 轮询的 key（taskId）
   * @param interval 轮询间隔（毫秒），默认 5000
   */
  function startPollingMessageStatus(key: string, interval: number = 5000) {
    if (pollingTimer.value) {
      stopPollingMessageStatus();
    }

    const poll = async () => {
      const { data, error } = await fetchMessages({ taskId: parseInt(key, 10) });

      if (!error && data) {
        messages.value = data.messages;
        unreadCount.value = data.total || 0;

        const message = data.messages[0];

        // 检查是否有状态变化
        if (message && (message.status === 'success' || message.status === 'failed')) {
          // 完成/失败时停止轮询
          stopPollingMessageStatus();

          // 通知UI
          if (message.status === 'success') {
            window.$message?.success(message.resultSummary || '任务拆分完成');
          } else if (message.status === 'failed') {
            window.$message?.error(`任务拆分失败: ${message.errorMessage}`);
          }
        }
      }
    };

    // 立即执行一次
    poll();

    // 启动定时轮询
    pollingTimer.value = setInterval(poll, interval);
  }

  /**
   * 停止轮询检查消息状态
   */
  function stopPollingMessageStatus() {
    if (pollingTimer.value) {
      clearInterval(pollingTimer.value);
      pollingTimer.value = null;
    }
  }

  return {
    messages,
    unreadCount,
    loading,
    loadMessages,
    loadUnreadCount,
    markAsRead,
    deleteMessageById,
    startPollingMessageStatus,
    stopPollingMessageStatus
  };
});
