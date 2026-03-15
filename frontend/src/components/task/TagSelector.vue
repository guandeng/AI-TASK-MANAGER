<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { NButton, NInput, NSelect, NSpace, NTag } from 'naive-ui';
import type { SelectGroupOption, SelectOption } from 'naive-ui';

interface Props {
  modelValue?: string[];
  tags?: string[];
  placeholder?: string;
  creatable?: boolean;
}

interface Emits {
  (e: 'update:modelValue', value: string[]): void;
  (e: 'create', tag: string): void;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: () => [],
  tags: () => [],
  placeholder: '选择或创建标签',
  creatable: true
});

const emit = defineEmits<Emits>();

// 当前选中的标签
const selectedTags = ref<string[]>([...props.modelValue]);

// 输入值
const inputValue = ref('');

// 监听外部值变化
watch(
  () => props.modelValue,
  newVal => {
    selectedTags.value = [...newVal];
  }
);

// 监听选中变化
watch(
  selectedTags,
  newVal => {
    emit('update:modelValue', newVal);
  },
  { deep: true }
);

// 生成选项
const options = computed(() => {
  const opts: SelectOption[] = props.tags.map(tag => ({
    label: tag,
    value: tag
  }));
  return opts;
});

// 处理创建标签
function handleCreate(value: string) {
  if (!value.trim()) return;

  // 如果标签不存在，创建新标签
  if (!props.tags.includes(value)) {
    emit('create', value);
  }

  // 添加到选中列表
  if (!selectedTags.value.includes(value)) {
    selectedTags.value.push(value);
  }

  inputValue.value = '';
}

// 移除标签
function removeTag(tag: string) {
  selectedTags.value = selectedTags.value.filter(t => t !== tag);
}
</script>

<template>
  <div class="tag-selector">
    <!-- 已选标签展示 -->
    <NSpace v-if="selectedTags.length > 0" wrap class="selected-tags">
      <NTag v-for="tag in selectedTags" :key="tag" closable type="info" size="small" @close="removeTag(tag)">
        {{ tag }}
      </NTag>
    </NSpace>

    <!-- 标签选择/输入 -->
    <div class="tag-input-wrapper">
      <NInput
        v-model:value="inputValue"
        :placeholder="placeholder"
        size="small"
        class="tag-input"
        @keydown.enter.prevent="handleCreate(inputValue)"
      />
      <NButton v-if="creatable && inputValue.trim()" type="primary" size="small" @click="handleCreate(inputValue)">
        添加
      </NButton>
    </div>
  </div>
</template>

<style scoped lang="scss">
.tag-selector {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.selected-tags {
  min-height: 24px;
}

.tag-input-wrapper {
  display: flex;
  gap: 8px;
  align-items: center;
}

.tag-input {
  flex: 1;
}
</style>
