<template>
  <el-empty class="table-empty" :image-size="68" aria-label="暂无数据">
    <template #description><span /></template>
  </el-empty>
</template>

<style scoped>
.table-empty :deep(.el-empty__description) {
  display: none;
}
</style>
