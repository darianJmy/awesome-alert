<template>
  <el-select :model-value="modelValue" filterable aria-label="拨测类型" placeholder="搜索或选择拨测类型" popper-class="shared-probe-type-popper" @change="$emit('change', $event)">
    <el-option-group v-for="group in probeTypeGroups" :key="group.label" :label="group.label">
      <el-option v-for="item in group.options" :key="item.value" :label="`${item.label} · ${item.title}`" :value="item.value">
        <div class="probe-option"><span class="probe-option-label">{{ item.label }}</span><strong>{{ item.title }}</strong><HelpTooltip :content="item.description" :label="`${item.label}说明`" /></div>
      </el-option>
    </el-option-group>
  </el-select>
</template>
<script setup>
import { probeTypeGroups } from '../utils/probeTypes'
import HelpTooltip from './HelpTooltip.vue'
defineProps({ modelValue: { type: String, required: true } })
defineEmits(['change'])
</script>
<style scoped>
.el-select{width:100%}.probe-option{display:flex;align-items:center;gap:10px;height:100%}.probe-option-label{display:flex;min-width:64px;align-items:center;justify-content:center;padding:4px 6px;color:var(--el-color-primary);background:var(--el-color-primary-light-9);border-radius:4px;font-size:12px;font-weight:600}.probe-option strong{flex:1;color:var(--el-text-color-primary);font-size:13px;font-weight:500}:global(.shared-probe-type-popper .el-select-dropdown__item){height:42px;line-height:normal;padding:4px 12px}
</style>
