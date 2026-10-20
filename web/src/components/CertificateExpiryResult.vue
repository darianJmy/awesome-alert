<template>
  <div class="expiry-result">
    <el-tag :type="meta.type" size="small">{{ meta.label }}</el-tag>
    <span v-if="hasExpiry">到期时间：{{ new Date(data.not_after).toLocaleString('zh-CN', { hour12: false }) }}</span>
    <span v-if="hasExpiry && data.certificate_status !== 'expired'">剩余 {{ data.days_remaining ?? 0 }} 天 · 提前 {{ data.expiry_alert_days }} 天告警</span>
    <span v-if="data.error" class="expiry-error">{{ data.error }}</span>
  </div>
</template>
<script setup>
import { computed } from 'vue'
const props = defineProps({ result: { type: [String, Object], default: '' } })
const data = computed(() => { try { return typeof props.result === 'string' ? JSON.parse(props.result) : props.result || {} } catch { return {} } })
const meta = computed(() => ({ normal: { label: '正常', type: 'success' }, expiring: { label: '即将到期', type: 'warning' }, expired: { label: '已过期', type: 'danger' }, check_failed: { label: '检测失败', type: 'danger' } }[data.value.certificate_status] || { label: '检测失败', type: 'danger' }))
const hasExpiry = computed(() => data.value.not_after && !data.value.not_after.startsWith('0001-') && !Number.isNaN(new Date(data.value.not_after).getTime()))
</script>
<style scoped>
.expiry-result{display:flex;flex-direction:column;align-items:flex-start;gap:6px;font-size:12px;line-height:1.6}.expiry-error{color:var(--el-color-danger);overflow-wrap:anywhere}
</style>
