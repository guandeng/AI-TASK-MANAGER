<script setup lang="ts">
import { h, computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import {
  NCard, NTag, NSpace, NButton, NDescriptions, NDescriptionsItem,
  NDivider, NEmpty, NSpin, NAlert, NDrawer, NDrawerContent,
  NProgress, NAvatar, useMessage, NBreadcrumb, NBreadcrumbItem
} from 'naive-ui';
import { MdPreview } from 'md-editor-v3';
import 'md-editor-v3/lib/style.css';
import {
  fetchProjectTemplate,
  scoreProjectTemplate,
  instantiateProjectTemplate
} from '@/service/api/template';
import type { ProjectTemplate, TemplateScoreResult } from '@/typings/api/template';
import {
  TEMPLATE_CATEGORY_OPTIONS,
  TEMPLATE_PRIORITY_LABELS,
  TEMPLATE_PRIORITY_COLORS
} from '@/typings/api/template';
import SvgIcon from '@/components/custom/svg-icon.vue';

const router = useRouter();
const message = useMessage();

const props = defineProps<{
  id: string
}>();

const loading = ref(false);
const template = ref<ProjectTemplate | null>(null);
const scoreLoading = ref(false);
const showScoreDrawer = ref(false);
const scoreResult = ref<TemplateScoreResult | null>(null);
const scoreLevel = ref('');

// 获取模板 ID
const templateId = computed(() => {
  return props.id || '';
});

// 加载模板详情
async function loadTemplate() {
  loading.value = true;
  try {
    const { data, error } = await fetchProjectTemplate(parseInt(templateId.value));
    if (!error && data) {
      template.value = data.data || data;
      // 如果有评分，解析并显示
      if (template.value?.lastScore) {
        try {
          const scoreData = JSON.parse(template.value.lastScore);
          scoreResult.value = scoreData;
          scoreLevel.value = getScoreLevel(scoreData.totalScore || 0);
        } catch (e) {
          console.error('解析评分失败', e);
        }
      }
    }
  } finally {
    loading.value = false;
  }
}

// 评分模板
async function handleScore() {
  scoreLoading.value = true;
  try {
    const { data, error } = await scoreProjectTemplate({ id: parseInt(templateId.value) });
    if (!error && data) {
      const result = data.data || data;
      scoreResult.value = result.score;
      scoreLevel.value = getScoreLevel(result.totalScore || 0);
      message.success('评分完成');
      showScoreDrawer.value = true;
      // 重新加载模板以获取更新的评分
      loadTemplate();
    }
  } catch (e) {
    message.error('评分失败');
  } finally {
    scoreLoading.value = false;
  }
}

// 打开评分抽屉
function openScoreDrawer() {
  if (scoreResult.value) {
    showScoreDrawer.value = true;
  } else {
    handleScore();
  }
}

// 实例化模板
async function handleInstantiate() {
  if (!template.value) return;

  const { data, error } = await instantiateProjectTemplate(template.value.id, {
    name: template.value.name,
    description: template.value.description
  });

  if (error) {
    message.error('创建失败');
  } else {
    const result = data?.data || data;
    message.success(`已创建项目，包含 ${result?.taskIds?.length || 0} 个任务`);
    router.push(`/requirement/list?id=${result?.requirementId}`);
  }
}

// 获取评分等级
function getScoreLevel(score: number): string {
  if (score >= 90) return '优秀';
  if (score >= 70) return '良好';
  if (score >= 50) return '一般';
  if (score >= 30) return '较差';
  return '很差';
}

// 获取评分等级颜色
function getScoreLevelColor(score: number): string {
  if (score >= 90) return '#18a058';
  if (score >= 70) return '#2080f0';
  if (score >= 50) return '#f0a020';
  return '#d03050';
}

// 返回列表
function goBack() {
  router.back();
}

onMounted(() => {
  if (props.id) {
    loadTemplate();
  }
});
</script>

<template>
  <div class="template-detail-page">
    <div class="page-header">
      <NBreadcrumb>
        <NBreadcrumbItem @click="goBack">项目模板</NBreadcrumbItem>
        <NBreadcrumbItem>模板详情</NBreadcrumbItem>
      </NBreadcrumb>
    </div>

    <NSpin :show="loading">
      <div v-if="template" class="template-content">
        <!-- 基本信息卡片 -->
        <NCard class="mb-16">
          <div class="template-header">
            <div class="template-title-section">
              <h1 class="template-title">{{ template.name }}</h1>
              <NSpace class="mt-8">
                <NTag :bordered="false">
                  {{ TEMPLATE_CATEGORY_OPTIONS.find(o => o.value === template.category)?.label || template.category }}
                </NTag>
                <NTag :type="template.isPublic ? 'success' : 'default'" :bordered="false">
                  {{ template.isPublic ? '公开' : '私有' }}
                </NTag>
                <NTag :bordered="false">
                  使用 {{ template.usageCount }} 次
                </NTag>
              </NSpace>
            </div>
            <NSpace class="template-actions">
              <NButton @click="handleScore" :loading="scoreLoading">
                <SvgIcon icon="heroicons:star" class="mr-4" />
                {{ scoreResult ? '重新评分' : '评分' }}
              </NButton>
              <NButton v-if="scoreResult" @click="openScoreDrawer">
                查看评分结果
              </NButton>
              <NButton type="primary" @click="handleInstantiate">
                <SvgIcon icon="heroicons:folder-plus" class="mr-4" />
                使用此模板
              </NButton>
            </NSpace>
          </div>

          <!-- 评分概览 -->
          <div v-if="scoreResult" class="score-overview">
            <div class="score-main">
              <div class="score-number" :style="{ color: getScoreLevelColor(scoreResult.totalScore || 0) }">
                {{ (scoreResult.totalScore || 0).toFixed(1) }}
              </div>
              <div class="score-label">
                <span :style="{ color: getScoreLevelColor(scoreResult.totalScore || 0) }">
                  {{ scoreLevel }}
                </span>
              </div>
            </div>
            <div class="score-dimensions">
              <div class="score-dimension-item">
                <span class="dimension-label">清晰度</span>
                <NProgress
                  type="line"
                  :percentage="(scoreResult.scores?.clarity || 0) * 10"
                  :color="getScoreLevelColor(scoreResult.scores?.clarity * 10 || 0)"
                  :show-indicator="false"
                  :height="6"
                />
                <span class="dimension-value">{{ scoreResult.scores?.clarity }}/10</span>
              </div>
              <div class="score-dimension-item">
                <span class="dimension-label">完整性</span>
                <NProgress
                  type="line"
                  :percentage="(scoreResult.scores?.completeness || 0) * 10"
                  :color="getScoreLevelColor(scoreResult.scores?.completeness * 10 || 0)"
                  :show-indicator="false"
                  :height="6"
                />
                <span class="dimension-value">{{ scoreResult.scores?.completeness }}/10</span>
              </div>
              <div class="score-dimension-item">
                <span class="dimension-label">结构化</span>
                <NProgress
                  type="line"
                  :percentage="(scoreResult.scores?.structure || 0) * 10"
                  :color="getScoreLevelColor(scoreResult.scores?.structure * 10 || 0)"
                  :show-indicator="false"
                  :height="6"
                />
                <span class="dimension-value">{{ scoreResult.scores?.structure }}/10</span>
              </div>
              <div class="score-dimension-item">
                <span class="dimension-label">可执行性</span>
                <NProgress
                  type="line"
                  :percentage="(scoreResult.scores?.actionability || 0) * 10"
                  :color="getScoreLevelColor(scoreResult.scores?.actionability * 10 || 0)"
                  :show-indicator="false"
                  :height="6"
                />
                <span class="dimension-value">{{ scoreResult.scores?.actionability }}/10</span>
              </div>
              <div class="score-dimension-item">
                <span class="dimension-label">一致性</span>
                <NProgress
                  type="line"
                  :percentage="(scoreResult.scores?.consistency || 0) * 10"
                  :color="getScoreLevelColor(scoreResult.scores?.consistency * 10 || 0)"
                  :show-indicator="false"
                  :height="6"
                />
                <span class="dimension-value">{{ scoreResult.scores?.consistency }}/10</span>
              </div>
            </div>
          </div>

          <NDivider />

          <NDescriptions label-placement="left" :column="2">
            <NDescriptionsItem label="描述" :span="2">
              <MdPreview v-if="template.description" :modelValue="template.description" />
              <span v-else class="text-gray-400">-</span>
            </NDescriptionsItem>
            <NDescriptionsItem label="创建时间">{{ template.createdAt }}</NDescriptionsItem>
            <NDescriptionsItem label="更新时间">{{ template.updatedAt }}</NDescriptionsItem>
          </NDescriptions>
        </NCard>

        <!-- 任务列表卡片 -->
        <NCard title="任务列表" class="mb-16">
          <div v-if="template.tasks && template.tasks.length > 0" class="task-list">
            <div v-for="task in template.tasks" :key="task.id" class="task-item">
              <div class="task-header">
                <span class="task-title">{{ task.title }}</span>
                <NSpace>
                  <NTag :type="TEMPLATE_PRIORITY_COLORS[task.priority]" size="small">
                    {{ TEMPLATE_PRIORITY_LABELS[task.priority] }}
                  </NTag>
                  <NTag v-if="task.estimatedHours" type="info" size="small">
                    {{ task.estimatedHours }}h
                  </NTag>
                </NSpace>
              </div>
              <div v-if="task.description" class="task-desc">{{ task.description }}</div>
              <div v-if="task.subtasks && task.subtasks.length > 0" class="subtask-list">
                <div class="subtask-title">子任务：</div>
                <div v-for="subtask in task.subtasks" :key="subtask.id" class="subtask-item">
                  <SvgIcon icon="heroicons:check-circle" class="subtask-icon" />
                  {{ subtask.title }}
                </div>
              </div>
            </div>
          </div>
          <NEmpty v-else description="暂无任务" />
        </NCard>
      </div>
      <NEmpty v-else description="模板不存在" />
    </NSpin>

    <!-- 评分结果抽屉 -->
    <NDrawer v-model:show="showScoreDrawer" :width="600" placement="right">
      <NDrawerContent title="项目模板评分结果" :native-scrollbar="false" closable>
        <NSpin :show="scoreLoading">
          <div v-if="scoreResult" class="score-result">
            <!-- 总分 -->
            <div class="total-score">
              <div class="score-number" :style="{ color: getScoreLevelColor(scoreResult.totalScore || 0) }">
                {{ (scoreResult.totalScore || 0).toFixed(1) }}
              </div>
              <div class="score-label">总分 / 100</div>
              <div class="score-level" :style="{ color: getScoreLevelColor(scoreResult.totalScore || 0) }">
                {{ getScoreLevel(scoreResult.totalScore || 0) }}
              </div>
            </div>

            <!-- 维度评分 -->
            <div class="dimensions" v-if="scoreResult.scores">
              <div class="dimension-item">
                <span class="dimension-name">清晰度</span>
                <NTag :type="scoreResult.scores.clarity >= 7 ? 'success' : 'warning'">
                  {{ scoreResult.scores.clarity }} / 10
                </NTag>
              </div>
              <div class="dimension-item">
                <span class="dimension-name">完整性</span>
                <NTag :type="scoreResult.scores.completeness >= 7 ? 'success' : 'warning'">
                  {{ scoreResult.scores.completeness }} / 10
                </NTag>
              </div>
              <div class="dimension-item">
                <span class="dimension-name">结构化</span>
                <NTag :type="scoreResult.scores.structure >= 7 ? 'success' : 'warning'">
                  {{ scoreResult.scores.structure }} / 10
                </NTag>
              </div>
              <div class="dimension-item">
                <span class="dimension-name">可执行性</span>
                <NTag :type="scoreResult.scores.actionability >= 7 ? 'success' : 'warning'">
                  {{ scoreResult.scores.actionability }} / 10
                </NTag>
              </div>
              <div class="dimension-item">
                <span class="dimension-name">一致性</span>
                <NTag :type="scoreResult.scores.consistency >= 7 ? 'success' : 'warning'">
                  {{ scoreResult.scores.consistency }} / 10
                </NTag>
              </div>
            </div>

            <!-- 评价内容 -->
            <template v-if="scoreResult.strengths || scoreResult.weaknesses || scoreResult.suggestions">
              <NDivider>评价内容</NDivider>
              <NAlert type="success" title="优点" class="eval-section" v-if="scoreResult.strengths">
                <ul>
                  <li v-for="(item, i) in scoreResult.strengths" :key="i">{{ item }}</li>
                </ul>
              </NAlert>
              <NAlert type="warning" title="待改进" class="eval-section" v-if="scoreResult.weaknesses">
                <ul>
                  <li v-for="(item, i) in scoreResult.weaknesses" :key="i">{{ item }}</li>
                </ul>
              </NAlert>
              <NAlert type="info" title="改进建议" class="eval-section" v-if="scoreResult.suggestions">
                <div v-for="(item, i) in scoreResult.suggestions" :key="i" class="suggestion-item">
                  <strong>{{ item.issue }}</strong>
                  <p>{{ item.suggestion }}</p>
                </div>
              </NAlert>
            </template>

            <!-- 详细分析 -->
            <template v-if="scoreResult.analysis">
              <NDivider>详细分析</NDivider>
              <div class="analysis">
                <MdPreview :modelValue="scoreResult.analysis" />
              </div>
            </template>
          </div>
          <NEmpty v-else description="暂无评分结果" />
        </NSpin>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>

<style scoped lang="scss">
.template-detail-page {
  padding: 16px;
  min-height: 100%;
}

.page-header {
  margin-bottom: 16px;
}

.template-content {
  max-width: 1200px;
  margin: 0 auto;
}

.mb-16 {
  margin-bottom: 16px;
}

.template-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;

  .template-title-section {
    flex: 1;
  }

  .template-title {
    font-size: 24px;
    font-weight: 600;
    margin: 0 0 8px 0;
  }

  .mt-8 {
    margin-top: 8px;
  }

  .template-actions {
    flex-shrink: 0;
  }
}

.score-overview {
  background: var(--n-border-color-popover);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;

  .score-main {
    display: flex;
    align-items: baseline;
    gap: 12px;
    margin-bottom: 16px;

    .score-number {
      font-size: 48px;
      font-weight: bold;
      line-height: 1;
    }

    .score-label {
      font-size: 16px;
      font-weight: 500;
    }
  }

  .score-dimensions {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 12px;

    .score-dimension-item {
      display: flex;
      align-items: center;
      gap: 8px;

      .dimension-label {
        width: 60px;
        font-size: 13px;
        color: var(--n-text-color-2);
      }

      .dimension-value {
        width: 40px;
        font-size: 12px;
        color: var(--n-text-color-3);
        text-align: right;
      }
    }
  }
}

.task-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-item {
  padding: 12px;
  background: var(--n-border-color-popover);
  border-radius: 6px;

  .task-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;

    .task-title {
      font-weight: 500;
      font-size: 15px;
    }
  }

  .task-desc {
    font-size: 13px;
    color: var(--n-text-color-2);
    margin-bottom: 8px;
    line-height: 1.6;
  }

  .subtask-list {
    margin-top: 8px;
    padding-left: 8px;

    .subtask-title {
      font-size: 12px;
      color: var(--n-text-color-3);
      margin-bottom: 4px;
    }

    .subtask-item {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 13px;
      color: var(--n-text-color-2);
      padding: 4px 0;

      .subtask-icon {
        width: 14px;
        height: 14px;
        color: var(--n-color-target);
        flex-shrink: 0;
      }
    }
  }
}

.score-result {
  display: flex;
  flex-direction: column;
  gap: 16px;

  .total-score {
    text-align: center;
    padding: 20px 0;
    background: var(--n-border-color-popover);
    border-radius: 8px;

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
