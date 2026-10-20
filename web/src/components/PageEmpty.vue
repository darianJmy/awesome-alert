<template>
  <el-empty class="page-empty" :description="description" :image-size="120" />
</template>

<script setup>
defineProps({ description: { type: String, required: true } })
</script>

<style scoped>
.page-empty {
  width: 100%;
  min-height: 320px;
  margin: 0;
  padding: 40px 16px;
  align-items: center;
  justify-content: center;
}
</style>
