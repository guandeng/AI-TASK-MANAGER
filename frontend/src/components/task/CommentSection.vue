<script setup lang="ts">
import { computed, onMounted, ref, watch, nextTick } from 'vue';
import {
  NCard,
  NSpace,
  NButton,
  NInput,
  NAvatar,
  NEmpty,
  NSpin,
  NPopconfirm,
  NDropdown,
  NTooltip,
  NEllipsis
} from 'naive-ui';
import { fetchTaskComments, createComment, updateComment, deleteComment, fetchCommentReplies } from '@/service/api/comment';
import type { Comment } from '@/typings/api/comment';
import { formatTimeAgo } from '@/utils/common';

const props = defineProps<{
  taskId: number;
  memberId?: number; // 当前用户成员ID
}>();

const emit = defineEmits<{
  (e: 'commented'): void;
  (e: 'deleted'): void;
}>();

// 状态
const loading = ref(false);
const comments = ref<Comment[]>([]);
const newCommentContent = ref('');
const submitLoading = ref(false);

// 回复状态
const replyTo = ref<Comment | null>(null);
const replyContent = ref('');
const replyLoading = ref(false);

// 编辑状态
const editingComment = ref<Comment | null>(null);
const editContent = ref('');

// 展开的评论ID（用于显示回复）
const expandedComments = ref<Set<number>>(new Set());
const repliesMap = ref<Map<number, Comment[]>>(new Map());
const loadingReplies = ref<Set<number>>(new Set());

// 主评论（非回复）
const rootComments = computed(() => {
  return comments.value.filter(c => !c.parentId);
});

// 加载评论列表
async function loadComments() {
  if (!props.taskId) return;
  loading.value = true;
  try {
    const { data, error } = await fetchTaskComments(props.taskId);
    if (!error && data) {
      comments.value = data;
    }
  } finally {
    loading.value = false;
  }
}

// 加载回复
async function loadReplies(commentId: number) {
  if (loadingReplies.value.has(commentId)) return;

  loadingReplies.value.add(commentId);
  try {
    const { data, error } = await fetchCommentReplies(props.taskId, commentId);
    if (!error && data) {
      repliesMap.value.set(commentId, data);
      expandedComments.value.add(commentId);
    }
  } finally {
    loadingReplies.value.delete(commentId);
  }
}

// 切换回复显示
async function toggleReplies(comment: Comment) {
  if (expandedComments.value.has(comment.id)) {
    expandedComments.value.delete(comment.id);
  } else {
    await loadReplies(comment.id);
  }
}

// 提交新评论
async function handleSubmitComment() {
  if (!newCommentContent.value.trim()) {
    window.$message?.warning('请输入评论内容');
    return;
  }
  if (!props.memberId) {
    window.$message?.warning('请先选择成员身份');
    return;
  }

  submitLoading.value = true;
  try {
    const { error } = await createComment(props.taskId, {
      content: newCommentContent.value.trim(),
      memberId: props.memberId
    });

    if (error) {
      window.$message?.error('评论失败');
    } else {
      window.$message?.success('评论成功');
      newCommentContent.value = '';
      await loadComments();
      emit('commented');
    }
  } finally {
    submitLoading.value = false;
  }
}

// 开始回复
function startReply(comment: Comment) {
  replyTo.value = comment;
  replyContent.value = '';
}

// 取消回复
function cancelReply() {
  replyTo.value = null;
  replyContent.value = '';
}

// 提交回复
async function handleSubmitReply() {
  if (!replyContent.value.trim()) {
    window.$message?.warning('请输入回复内容');
    return;
  }
  if (!props.memberId || !replyTo.value) return;

  replyLoading.value = true;
  try {
    const { error } = await createComment(props.taskId, {
      content: replyContent.value.trim(),
      memberId: props.memberId,
      parentId: replyTo.value.id
    });

    if (error) {
      window.$message?.error('回复失败');
    } else {
      window.$message?.success('回复成功');
      cancelReply();
      await loadComments();
      // 自动展开回复
      if (replyTo.value) {
        await loadReplies(replyTo.value.id);
      }
      emit('commented');
    }
  } finally {
    replyLoading.value = false;
  }
}

// 开始编辑
function startEdit(comment: Comment) {
  editingComment.value = comment;
  editContent.value = comment.content;
}

// 取消编辑
function cancelEdit() {
  editingComment.value = null;
  editContent.value = '';
}

// 保存编辑
async function handleSaveEdit(comment: Comment) {
  if (!editContent.value.trim()) {
    window.$message?.warning('评论内容不能为空');
    return;
  }

  try {
    const { error } = await updateComment(props.taskId, comment.id, {
      content: editContent.value.trim(),
      memberId: props.memberId
    });

    if (error) {
      window.$message?.error('编辑失败');
    } else {
      window.$message?.success('已更新');
      cancelEdit();
      await loadComments();
    }
  } catch (e) {
    window.$message?.error('编辑失败');
  }
}

// 删除评论
async function handleDelete(comment: Comment) {
  if (!props.memberId) return;

  try {
    const { error } = await deleteComment(props.taskId, comment.id, props.memberId);
    if (error) {
      window.$message?.error('删除失败');
    } else {
      window.$message?.success('已删除');
      await loadComments();
      emit('deleted');
    }
  } catch (e) {
    window.$message?.error('删除失败');
  }
}

// 格式化时间
function formatTime(dateStr: string) {
  return formatTimeAgo(dateStr);
}

// 获取回复列表
function getReplies(commentId: number): Comment[] {
  return repliesMap.value.get(commentId) || [];
}

// 监听 taskId 变化
watch(() => props.taskId, () => {
  loadComments();
}, { immediate: true });
</script>

<template>
  <NCard title="评论" size="small">
    <template #header-extra>
      <span class="comment-count">{{ comments.length }} 条</span>
    </template>

    <NSpin :show="loading">
      <div class="comment-section">
        <!-- 评论输入框 -->
        <div class="comment-input-area">
          <NInput
            v-model:value="newCommentContent"
            type="textarea"
            placeholder="写下你的评论..."
            :autosize="{ minRows: 2, maxRows: 4 }"
            :disabled="!memberId"
          />
          <div class="input-actions">
            <NButton
              type="primary"
              size="small"
              :loading="submitLoading"
              :disabled="!newCommentContent.trim() || !memberId"
              @click="handleSubmitComment"
            >
              发表评论
            </NButton>
          </div>
        </div>

        <!-- 评论列表 -->
        <div v-if="rootComments.length > 0" class="comment-list">
          <div
            v-for="comment in rootComments"
            :key="comment.id"
            class="comment-item"
          >
            <!-- 评论内容 -->
            <div class="comment-header">
              <NSpace align="center">
                <NAvatar
                  round
                  size="small"
                  :src="comment.member?.avatar"
                  :name="comment.member?.name"
                />
                <span class="author-name">{{ comment.member?.name || '未知成员' }}</span>
                <span class="comment-time">{{ formatTime(comment.createdAt) }}</span>
              </NSpace>
            </div>

            <!-- 编辑模式 -->
            <div v-if="editingComment?.id === comment.id" class="comment-content edit-mode">
              <NInput
                v-model:value="editContent"
                type="textarea"
                :autosize="{ minRows: 2 }"
              />
              <NSpace class="edit-actions">
                <NButton size="small" @click="cancelEdit">取消</NButton>
                <NButton type="primary" size="small" @click="handleSaveEdit(comment)">保存</NButton>
              </NSpace>
            </div>
            <div v-else class="comment-content">
              <NEllipsis :line-clamp="5" expand-trigger="click">
                {{ comment.content }}
              </NEllipsis>
            </div>

            <!-- 评论操作 -->
            <div class="comment-actions">
              <NButton text size="small" @click="startReply(comment)">回复</NButton>
              <template v-if="memberId === comment.memberId">
                <NButton text size="small" @click="startEdit(comment)">编辑</NButton>
                <NPopconfirm @positive-click="handleDelete(comment)">
                  <template #trigger>
                    <NButton text type="error" size="small">删除</NButton>
                  </template>
                  确认删除此评论？
                </NPopconfirm>
              </template>
              <!-- 回复数量提示 -->
              <NButton
                v-if="comment.replyCount && comment.replyCount > 0"
                text
                size="small"
                type="info"
                @click="toggleReplies(comment)"
              >
                {{ expandedComments.has(comment.id) ? '收起' : '' }}{{ comment.replyCount }} 条回复
              </NButton>
            </div>

            <!-- 回复输入框 -->
            <div v-if="replyTo?.id === comment.id" class="reply-input-area">
              <NInput
                v-model:value="replyContent"
                type="textarea"
                placeholder="写下你的回复..."
                :autosize="{ minRows: 2, maxRows: 3 }"
              />
              <NSpace class="reply-actions">
                <NButton size="small" @click="cancelReply">取消</NButton>
                <NButton
                  type="primary"
                  size="small"
                  :loading="replyLoading"
                  :disabled="!replyContent.trim()"
                  @click="handleSubmitReply"
                >
                  回复
                </NButton>
              </NSpace>
            </div>

            <!-- 回复列表 -->
            <div
              v-if="expandedComments.has(comment.id) && getReplies(comment.id).length > 0"
              class="replies-list"
            >
              <div
                v-for="reply in getReplies(comment.id)"
                :key="reply.id"
                class="reply-item"
              >
                <div class="reply-header">
                  <NSpace align="center">
                    <NAvatar
                      round
                      :size="20"
                      :src="reply.member?.avatar"
                      :name="reply.member?.name"
                    />
                    <span class="author-name">{{ reply.member?.name || '未知成员' }}</span>
                    <span class="comment-time">{{ formatTime(reply.createdAt) }}</span>
                  </NSpace>
                </div>
                <div class="reply-content">{{ reply.content }}</div>
                <div class="comment-actions">
                  <template v-if="memberId === reply.memberId">
                    <NPopconfirm @positive-click="handleDelete(reply)">
                      <template #trigger>
                        <NButton text type="error" size="tiny">删除</NButton>
                      </template>
                      确认删除此回复？
                    </NPopconfirm>
                  </template>
                </div>
              </div>
            </div>
          </div>
        </div>
        <NEmpty v-else description="暂无评论，来说点什么吧~" />
      </div>
    </NSpin>
  </NCard>
</template>

<style scoped lang="scss">
.comment-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.comment-count {
  font-size: 13px;
  color: #999;
}

.comment-input-area {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;

  .input-actions {
    display: flex;
    justify-content: flex-end;
  }
}

.comment-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.comment-item {
  padding: 12px;
  background: #fafafa;
  border-radius: 8px;

  &:hover {
    background: #f5f5f5;
  }
}

.comment-header {
  margin-bottom: 8px;

  .author-name {
    font-weight: 500;
    font-size: 14px;
  }

  .comment-time {
    font-size: 12px;
    color: #999;
  }
}

.comment-content {
  font-size: 14px;
  line-height: 1.6;
  color: #333;
  margin-bottom: 8px;

  &.edit-mode {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
}

.comment-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.reply-input-area {
  margin-top: 12px;
  padding: 12px;
  background: #fff;
  border-radius: 6px;
  border: 1px solid #e8e8e8;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.replies-list {
  margin-top: 12px;
  padding-left: 24px;
  border-left: 2px solid #e8e8e8;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.reply-item {
  padding: 8px 12px;
  background: #fff;
  border-radius: 6px;

  .reply-header {
    margin-bottom: 4px;

    .author-name {
      font-size: 13px;
      font-weight: 500;
    }

    .comment-time {
      font-size: 11px;
      color: #999;
    }
  }

  .reply-content {
    font-size: 13px;
    color: #666;
    line-height: 1.5;
  }
}
</style>
