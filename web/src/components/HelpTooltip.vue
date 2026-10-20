<template>
  <el-tooltip :content="content" placement="top" :show-after="150">
    <button type="button" class="help-tooltip" :aria-label="label || content" @click.stop @keydown.enter.stop @keydown.space.stop>
      <el-icon><QuestionFilled /></el-icon>
    </button>
  </el-tooltip>
</template>
<script setup>
import { QuestionFilled } from '@element-plus/icons-vue'
defineProps({ content: { type: String, required: true }, label: { type: String, default: '' } })
</script>
<style scoped>
.help-tooltip{display:inline-flex;align-items:center;justify-content:center;flex:none;width:24px;height:24px;padding:0;color:var(--el-text-color-secondary);background:transparent;border:0;border-radius:4px;cursor:help;vertical-align:middle}.help-tooltip:hover{color:var(--el-color-primary)}.help-tooltip:focus-visible{outline:2px solid var(--el-color-primary);outline-offset:2px}.help-tooltip .el-icon{font-size:14px}
</style>
