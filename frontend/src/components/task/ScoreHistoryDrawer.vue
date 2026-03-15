<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { NAlert, NButton, NCard, NDrawer, NDrawerContent, NEmpty, NModal, NSpace, NSpin, NTag } from 'naive-ui';
import { MdPreview } from 'md-editor-v3';
import { useTaskStore } from '@/store/modules/task';
import type { EvalSuggestion, EvaluationData, TaskQualityScore } from '@/typings/api/task';
import 'md-editor-v3/lib/style.css';

interface Props {
  taskId: number | null;
  show: boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: 'update:show', value: boolean): void;
}>();

const taskStore = useTaskStore();

const localShow = computed({
  get: () => props.show,
  set: val => emit('update:show', val)
});

const selectedScore = ref<TaskQualityScore | null>(null);
const showRestoreModal = ref(false);
const showDeleteModal = ref(false);
const scoreToDelete = ref<number | null>(null);

// 加载评分历史
watch(
  () => props.taskId,
  async id => {
    if (id) {
      await taskStore.loadScoreHistory(id);
      if (taskStore.qualityScores.length > 0) {
        selectedScore.value = taskStore.qualityScores[0];
      }
    }
  },
  { immediate: true }
);

// 选择评分
function selectScore(score: TaskQualityScore) {
  selectedScore.value = score;
}

// 获取评价数据
function getEvaluationData(score: TaskQualityScore): EvaluationData | null {
  try {
    return JSON.parse(score.evaluation);
  } catch {
    return null;
  }
}

// 获取等级
function getScoreLevel(score: number): string {
  if (score >= 90) return '优秀';
  if (score >= 70) return '良好';
  if (score >= 50) return '一般';
  if (score >= 30) return '较差';
  return '很差';
}

// 获取等级颜色
function getScoreLevelColor(score: number): string {
  if (score >= 90) return '#18a058'; // green
  if (score >= 70) return '#2080f0'; // blue
  if (score >= 50) return '#f0a020'; // orange
  return '#d03050'; // red
}

// 确认删除
function confirmDelete(scoreId: number) {
  scoreToDelete.value = scoreId;
  showDeleteModal.value = true;
}

// 执行删除
async function handleDelete() {
  if (scoreToDelete.value && props.taskId) {
    await taskStore.deleteScore(props.taskId, scoreToDelete.value);
    showDeleteModal.value = false;
    scoreToDelete.value = null;
    if (selectedScore.value?.id === scoreToDelete.value) {
      selectedScore.value = taskStore.qualityScores[0] || null;
    }
  }
}

// 恢复版本
function confirmRestore(score: TaskQualityScore) {
  selectedScore.value = score;
  showRestoreModal.value = true;
}

// 执行恢复
async function handleRestore() {
  if (selectedScore.value && props.taskId) {
    const success = await taskStore.restoreScore(props.taskId, selectedScore.value.id);
    showRestoreModal.value = false;
    if (success) {
      // 刷新任务详情
      await taskStore.loadTaskDetail(props.taskId);
      window.$message?.success('任务已恢复到评分时的状态');
    }
  }
}

// 关闭抽屉
function closeDrawer() {
  localShow.value = false;
}
</script>

<template>
  <NDrawer v-model:show="localShow" :width="600" :on-update:show="closeDrawer">
    <NDrawerContent title="任务质量评分历史" :native-scrollbar="false" closable>
      <NSpin :show="taskStore.scoreHistoryLoading">
        <div class="score-history-container">
          <!-- 评分列表 -->
          <div v-if="taskStore.qualityScores.length > 0" class="score-list">
            <NCard
              v-for="score in taskStore.qualityScores"
              :key="score.id"
              :title="`版本 ${score.version}`"
              :bordered="false"
              size="small"
              :class="{ selected: selectedScore?.id === score.id }"
              class="score-item"
              @click="selectScore(score)"
            >
              <template #header-extra>
                <NTag :type="score.totalScore >= 70 ? 'success' : 'warning'">
                  {{ score.totalScore }}分 - {{ getScoreLevel(score.totalScore) }}
                </NTag>
              </template>
              <div class="score-meta">
                <span class="score-time">{{ new Date(score.createdAt).toLocaleString('zh-CN') }}</span>
                <span class="score-provider">{{ score.aiProvider }}</span>
              </div>
              <div class="score-dimensions">
                <NTag size="small" type="info" :bordered="false">清晰 {{ score.clarityScore }}</NTag>
                <NTag size="small" type="info" :bordered="false">完整 {{ score.completenessScore }}</NTag>
                <NTag size="small" type="info" :bordered="false">结构 {{ score.structureScore }}</NTag>
                <NTag size="small" type="info" :bordered="false">执行 {{ score.actionabilityScore }}</NTag>
                <NTag size="small" type="info" :bordered="false">一致 {{ score.consistencyScore }}</NTag>
              </div>
            </NCard>
          </div>
          <NEmpty v-else description="暂无评分记录" />

          <!-- 评分详情 -->
          <div v-if="selectedScore" class="score-detail">
            <NDivider />
            <NCard :title="`版本 ${selectedScore.version} 详情`" size="small">
              <template #header-extra>
                <NSpace>
                  <NButton size="small" quaternary @click.stop="confirmRestore(selectedScore)">
                    <template #icon>
                      <span class="i-mdi:backup-restore-outline" style="font-size: 18px"></span>
                    </template>
                    恢复此版本
                  </NButton>
                  <NButton size="small" quaternary @click.stop="confirmDelete(selectedScore.id)">
                    <template #icon><span class="i-mdi:delete-outline" style="font-size: 18px"></span></template>
                    删除
                  </NButton>
                </NSpace>
              </template>

              <!-- 总分 -->
              <div class="total-score">
                <div class="score-number" :style="{ color: getScoreLevelColor(selectedScore.totalScore) }">
                  {{ selectedScore.totalScore.toFixed(1) }}
                </div>
                <div class="score-label">总分 / 100</div>
                <div class="score-level" :style="{ color: getScoreLevelColor(selectedScore.totalScore) }">
                  {{ getScoreLevel(selectedScore.totalScore) }}
                </div>
              </div>

              <!-- 维度评分 -->
              <div class="dimensions">
                <div class="dimension-item">
                  <span class="dimension-name">清晰度</span>
                  <NTag :type="selectedScore.clarityScore >= 7 ? 'success' : 'warning'">
                    {{ selectedScore.clarityScore }} / 10
                  </NTag>
                </div>
                <div class="dimension-item">
                  <span class="dimension-name">完整性</span>
                  <NTag :type="selectedScore.completenessScore >= 7 ? 'success' : 'warning'">
                    {{ selectedScore.completenessScore }} / 10
                  </NTag>
                </div>
                <div class="dimension-item">
                  <span class="dimension-name">结构化</span>
                  <NTag :type="selectedScore.structureScore >= 7 ? 'success' : 'warning'">
                    {{ selectedScore.structureScore }} / 10
                  </NTag>
                </div>
                <div class="dimension-item">
                  <span class="dimension-name">可执行性</span>
                  <NTag :type="selectedScore.actionabilityScore >= 7 ? 'success' : 'warning'">
                    {{ selectedScore.actionabilityScore }} / 10
                  </NTag>
                </div>
                <div class="dimension-item">
                  <span class="dimension-name">一致性</span>
                  <NTag :type="selectedScore.consistencyScore >= 7 ? 'success' : 'warning'">
                    {{ selectedScore.consistencyScore }} / 10
                  </NTag>
                </div>
              </div>

              <!-- 评价内容 -->
              <template v-if="getEvaluationData(selectedScore)">
                <NDivider>评价内容</NDivider>
                <NAlert type="success" title="优点" class="eval-section">
                  <ul>
                    <li v-for="(item, i) in getEvaluationData(selectedScore)?.strengths" :key="i">{{ item }}</li>
                  </ul>
                </NAlert>
                <NAlert type="warning" title="待改进" class="eval-section">
                  <ul>
                    <li v-for="(item, i) in getEvaluationData(selectedScore)?.weaknesses" :key="i">{{ item }}</li>
                  </ul>
                </NAlert>
                <NAlert type="info" title="改进建议" class="eval-section">
                  <div
                    v-for="(item, i) in getEvaluationData(selectedScore)?.suggestions"
                    :key="i"
                    class="suggestion-item"
                  >
                    <strong>{{ item.issue }}</strong>
                    <p>{{ item.suggestion }}</p>
                  </div>
                </NAlert>
                <NDivider>详细分析</NDivider>
                <div class="analysis">
                  <MdPreview :model-value="getEvaluationData(selectedScore)?.analysis || ''" />
                </div>
              </template>
            </NCard>
          </div>
        </div>
      </NSpin>
    </NDrawerContent>
  </NDrawer>

  <!-- 删除确认弹窗 -->
  <NModal
    v-model:show="showDeleteModal"
    preset="dialog"
    title="确认删除"
    content="确定要删除这个评分记录吗？此操作不可恢复。"
    @positive-click="handleDelete"
  />

  <!-- 恢复确认弹窗 -->
  <NModal
    v-model:show="showRestoreModal"
    preset="dialog"
    title="确认恢复"
    :content="`确定要恢复到版本 ${selectedScore?.version} 吗？当前任务内容将被覆盖。`"
    @positive-click="handleRestore"
  />
</template>

<style scoped lang="scss">
.score-history-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.score-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.score-item {
  cursor: pointer;
  transition: all 0.2s;
  border: 2px solid transparent;

  &:hover {
    background-color: var(--n-card-color-hover);
  }

  &.selected {
    border-color: var(--n-color-target);
    background-color: var(--n-card-color-hover);
  }
}

.score-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--n-text-color-3);
  margin-top: 8px;
}

.score-dimensions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.score-detail {
  .total-score {
    text-align: center;
    padding: 20px 0;
    background: var(--n-border-color-popover);
    border-radius: 8px;
    margin-bottom: 16px;

    .score-number {
      font-size: 48px;
      font-weight: bold;
      line-height: 1;
    }

    .score-label {
      font-size: 12px;
      color: var(--n-text-color-3);
      margin-top: 4px;
    }

    .score-level {
      font-size: 18px;
      font-weight: 500;
      margin-top: 8px;
    }
  }

  .dimensions {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-bottom: 16px;

    .dimension-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 8px 12px;
      background: var(--n-border-color-popover);
      border-radius: 6px;

      .dimension-name {
        font-size: 14px;
        color: var(--n-text-color-2);
      }
    }
  }

  .eval-section {
    margin-bottom: 12px;

    ul {
      margin: 0;
      padding-left: 20px;
    }

    li {
      margin-bottom: 6px;
    }
  }

  .suggestion-item {
    margin-bottom: 12px;

    strong {
      display: block;
      margin-bottom: 4px;
      color: var(--n-text-color-2);
    }

    p {
      margin: 0;
      font-size: 13px;
      color: var(--n-text-color-3);
    }
  }

  .analysis {
    font-size: 14px;
    line-height: 1.8;
  }
}
</style>
