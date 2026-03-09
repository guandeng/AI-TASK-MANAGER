<script setup lang="ts">
defineOptions({
  name: 'GlobalFooter'
});

const currentYear = new Date().getFullYear();
</script>

<template>
  <DarkModeContainer class="h-full flex-center">
    <a href="https://github.com/soybeanjs/soybean-admin/blob/main/LICENSE" target="_blank" rel="noopener noreferrer">
      Copyright MIT © {{ currentYear }} 任务管理系统
    </a>
  </DarkModeContainer>
</template>

<style scoped></style>
