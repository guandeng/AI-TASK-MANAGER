<script setup lang="ts">
import { computed, ref, onMounted, nextTick } from 'vue';
import { NCard, NEmpty, NSpin, NTag, NButton, NSpace, NTooltip } from 'naive-ui';
import type { Task, TaskDependency } from '@/typings/api/task';
import { useTaskStore } from '@/store/modules/task';

interface Props {
  taskId?: number;
  height?: number;
}

const props = withDefaults(defineProps<Props>(), {
  taskId: undefined,
  height: 400
});

const taskStore = useTaskStore();

// 加载状态
const loading = ref(false);

// 依赖关系数据
const dependencies = ref<TaskDependency[]>([]);

// 节点位置信息
interface NodePosition {
  id: number;
  x: number;
  y: number;
  task: Task;
  isHighlighted: boolean;
}

// 计算后的节点和连线
const nodes = computed<NodePosition[]>(() => {
  const tasks = taskStore.tasks;
  const taskMap = new Map<number, Task>();
  tasks.forEach(t => taskMap.set(t.id, t));

  // 如果有指定 taskId，只显示相关任务
  let relevantTaskIds = new Set<number>();
  if (props.taskId) {
    relevantTaskIds.add(props.taskId);
    // 找出所有相关联的任务
    dependencies.value.forEach(dep => {
      if (dep.taskId === props.taskId) {
        relevantTaskIds.add(dep.dependsOnTaskId);
      }
      if (dep.dependsOnTaskId === props.taskId) {
        relevantTaskIds.add(dep.taskId);
      }
    });
  } else {
    // 否则显示所有有依赖关系的任务
    dependencies.value.forEach(dep => {
      relevantTaskIds.add(dep.taskId);
      relevantTaskIds.add(dep.dependsOnTaskId);
    });
  }

  // 布局计算 - 简单的层级布局
  const nodePositions: NodePosition[] = [];
  const processed = new Set<number>();
  const levels = new Map<number, number>();

  // 计算每个任务的层级（没有依赖的为 0 层）
  function getLevel(taskId: number, visited = new Set<number>()): number {
    if (levels.has(taskId)) return levels.get(taskId)!;
    if (visited.has(taskId)) return 0; // 循环依赖
    visited.add(taskId);

    const deps = dependencies.value.filter(d => d.taskId === taskId);
    if (deps.length === 0) {
      levels.set(taskId, 0);
      return 0;
    }

    const maxLevel = Math.max(...deps.map(d => getLevel(d.dependsOnTaskId, visited)));
    levels.set(taskId, maxLevel + 1);
    return maxLevel + 1;
  }

  relevantTaskIds.forEach(id => getLevel(id));

  // 按层级分组
  const levelGroups = new Map<number, number[]>();
  relevantTaskIds.forEach(id => {
    const level = levels.get(id) || 0;
    if (!levelGroups.has(level)) levelGroups.set(level, []);
    levelGroups.get(level)!.push(id);
  });

  // 计算位置
  const maxLevel = Math.max(...Array.from(levelGroups.keys()), 0);
  const levelHeight = (props.height - 80) / Math.max(maxLevel + 1, 1);
  const nodeWidth = 160;
  const nodeHeight = 60;

  levelGroups.forEach((taskIds, level) => {
    const colWidth = 1000 / Math.max(taskIds.length, 1);
    taskIds.forEach((taskId, index) => {
      const task = taskMap.get(taskId);
      if (task) {
        nodePositions.push({
          id: taskId,
          x: colWidth * index + colWidth / 2,
          y: levelHeight * level + 40,
          task,
          isHighlighted: props.taskId === taskId
        });
      }
    });
  });

  return nodePositions;
});

// 连线数据
const links = computed(() => {
  return dependencies.value.map(dep => {
    const fromNode = nodes.value.find(n => n.id === dep.dependsOnTaskId);
    const toNode = nodes.value.find(n => n.id === dep.taskId);
    return { from: fromNode, to: toNode, dependency: dep };
  }).filter(l => l.from && l.to);
});

// 生成 SVG 路径（贝塞尔曲线）
function getLinkPath(from: NodePosition, to: NodePosition): string {
  const dx = to.x - from.x;
  const dy = to.y - from.y;
  const controlOffset = Math.min(Math.abs(dy) * 0.5, 50);

  return `M ${from.x} ${from.y + 30}
          C ${from.x} ${from.y + 30 + controlOffset},
            ${to.x} ${to.y - 30 - controlOffset},
            ${to.x} ${to.y - 30}`;
}

// 获取依赖状态颜色
function getDependencyColor(dep: TaskDependency): string {
  const task = taskStore.tasks.find(t => t.id === dep.taskId);
  const dependsOnTask = taskStore.tasks.find(t => t.id === dep.dependsOnTaskId);

  if (!dependsOnTask) return '#d0d0d0';
  if (dependsOnTask.status === 'done') return '#18a058';
  if (task?.status === 'pending') return '#f0a020';
  return '#2080f0';
}

// 悬停的任务 ID
const hoveredNodeId = ref<number | null>(null);

// 加载依赖数据
async function loadDependencies() {
  loading.value = true;
  try {
    const result = await taskStore.validateDependencies();
    if (result && !result.error) {
      // 这里简化处理，实际应该调用专门的依赖 API
      dependencies.value = taskStore.taskDependencies;
    }
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  loadDependencies();
});
</script>

<template>
  <NCard
    title="依赖关系图"
    :bordered="false"
    :style="{ height: `${height}px` }"
  >
    <NSpin :show="loading">
      <div v-if="nodes.length === 0" class="empty-state">
        <NEmpty description="暂无依赖关系" :image-size="60" />
      </div>

      <div v-else class="dependency-graph" :style="{ height: `${height - 50}px` }">
        <svg class="links-layer" :style="{ height: `${height - 50}px` }">
          <defs>
            <marker
              id="arrowhead"
              markerWidth="10"
              markerHeight="7"
              refX="9"
              refY="3.5"
              orient="auto"
            >
              <polygon points="0 0, 10 3.5, 0 7" fill="#666" />
            </marker>
            <marker
              id="arrowhead-highlighted"
              markerWidth="10"
              markerHeight="7"
              refX="9"
              refY="3.5"
              orient="auto"
            >
              <polygon points="0 0, 10 3.5, 0 7" fill="#2080f0" />
            </marker>
          </defs>

          <g v-for="(link, index) in links" :key="index">
            <path
              :d="getLinkPath(link.from!, link.to!)"
              :stroke="getDependencyColor(link.dependency)"
              :stroke-width="hoveredNodeId && (hoveredNodeId === link.from!.id || hoveredNodeId === link.to!.id) ? 3 : 2"
              fill="none"
              :marker-end="hoveredNodeId && (hoveredNodeId === link.from!.id || hoveredNodeId === link.to!.id) ? 'url(#arrowhead-highlighted)' : 'url(#arrowhead)'"
              class="dependency-link"
            />
          </g>
        </svg>

        <div
          v-for="node in nodes"
          :key="node.id"
          class="task-node"
          :class="{ 'is-highlighted': node.isHighlighted, 'is-hovered': hoveredNodeId === node.id }"
          :style="{
            left: `${node.x - 80}px`,
            top: `${node.y - 30}px`
          }"
          @mouseenter="hoveredNodeId = node.id"
          @mouseleave="hoveredNodeId = null"
        >
          <NTooltip>
            <template #trigger>
              <div class="node-content">
                <NTag
                  :type="node.task.priority === 'high' ? 'error' : node.task.priority === 'medium' ? 'warning' : 'success'"
                  size="small"
                  class="priority-tag"
                >
                  {{ node.task.priority === 'high' ? '高' : node.task.priority === 'medium' ? '中' : '低' }}
                </NTag>
                <span class="node-title">{{ node.task.title }}</span>
                <NTag
                  :type="node.task.status === 'done' ? 'success' : node.task.status === 'in-progress' ? 'info' : node.task.status === 'deferred' ? 'warning' : 'default'"
                  size="small"
                  class="status-tag"
                >
                  {{ node.task.status === 'done' ? '已完成' : node.task.status === 'in-progress' ? '进行中' : node.task.status === 'deferred' ? '已延期' : '待处理' }}
                </NTag>
              </div>
            </template>
            <div>
              <div><strong>ID:</strong> {{ node.task.id }}</div>
              <div><strong>标题:</strong> {{ node.task.title }}</div>
              <div><strong>优先级:</strong> {{ node.task.priority }}</div>
              <div><strong>状态:</strong> {{ node.task.status }}</div>
              <div v-if="node.task.dueDate"><strong>截止日期:</strong> {{ node.task.dueDate }}</div>
            </div>
          </NTooltip>
        </div>
      </div>
    </NSpin>
  </NCard>
</template>

<style scoped lang="scss">
.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
}

.dependency-graph {
  position: relative;
  overflow: hidden;
}

.links-layer {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  pointer-events: none;
  z-index: 1;
}

.dependency-link {
  transition: stroke-width 0.2s ease;
}

.task-node {
  position: absolute;
  z-index: 2;
  transition: all 0.2s ease;
  cursor: pointer;

  .node-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 8px 12px;
    background: #fff;
    border: 2px solid #e0e0e0;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    min-width: 120px;
    max-width: 200px;
  }

  .priority-tag {
    font-size: 11px;
  }

  .node-title {
    font-size: 12px;
    font-weight: 500;
    text-align: center;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 100%;
  }

  .status-tag {
    font-size: 11px;
  }

  &.is-highlighted {
    .node-content {
      border-color: #2080f0;
      box-shadow: 0 4px 12px rgba(32, 128, 240, 0.3);
    }
  }

  &.is-hovered {
    transform: scale(1.05);

    .node-content {
      border-color: #2080f0;
    }
  }
}
</style>
