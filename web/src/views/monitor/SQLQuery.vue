<template>
  <div class="sql-query">
    <div class="toolbar">
      <el-date-picker v-model="range" type="datetimerange" start-placeholder="开始时间" end-placeholder="结束时间" />
      <el-select v-model="limit" aria-label="返回行数"><el-option v-for="n in [100,500,1000,5000]" :key="n" :value="n" :label="`最多 ${n} 行`" /></el-select>
      <el-button type="primary" :loading="loading" @click="execute">执行 SQL</el-button>
    </div>
    <el-input v-model="query" type="textarea" :rows="8" :maxlength="65536" aria-label="SQL 查询" placeholder="SELECT * FROM default.your_table LIMIT 100" @keydown.ctrl.enter.prevent="execute" @keydown.meta.enter.prevent="execute" />
    <p class="hint">支持 SELECT / WITH 只读查询。时间范围仅通过 {from:Int64}、{to:Int64} 参数提供（Unix 秒），需在 SQL 中引用，不会自动过滤数据。</p>
    <el-button text @click="query = timeExample">插入时间过滤示例</el-button>
    <el-alert v-if="error" :title="error" type="error" :closable="false" />
    <el-alert v-if="result?.limit_reached" title="已达到行数上限，可调整条件或增加返回行数。" type="warning" :closable="false" />
    <p v-if="result" class="hint">{{ result.rows.length }} 行 · {{ elapsed }} ms</p>
    <el-table v-if="result" :data="result.rows" max-height="650" empty-text="查询没有返回数据">
      <el-table-column v-for="(column, index) in result.columns" :key="index" :label="column.name" min-width="180"><template #default="{ row }"><span class="cell">{{ cell(row[column.name]) }}</span></template></el-table-column>
    </el-table>
  </div>
</template>
<script setup>
import { onBeforeUnmount, ref } from 'vue'
import request from '../../api/request'
const props = defineProps({ source: { type: Object, required: true } })
const query = ref('SELECT 1 AS value'), range = ref([new Date(Date.now()-3600000),new Date()]), limit = ref(500), loading = ref(false), error = ref(''), result = ref(null), elapsed = ref(0)
const timeExample = 'SELECT *\nFROM default.your_table\nWHERE Timestamp >= toDateTime({from:Int64})\n  AND Timestamp < toDateTime({to:Int64})\nORDER BY Timestamp DESC'
const cell = value => value === null ? 'NULL' : typeof value === 'object' ? JSON.stringify(value) : String(value ?? '')
let controller
async function execute() {
  if (loading.value) return
  error.value = ''; result.value = null
  const [start,end] = range.value || []
  if (!query.value.trim() || !start || !end || end <= start) { error.value = '请输入 SQL 和有效时间范围'; return }
  controller = new AbortController(); loading.value = true; const now = performance.now()
  try { const response = await request({ url:'/query/sql',method:'post',timeout:35000,signal:controller.signal,data:{ datasource_id:props.source.id,query:query.value,start:start.toISOString(),end:end.toISOString(),limit:limit.value } }); result.value=response.data; elapsed.value=Math.round(performance.now()-now) }
  catch(e) { if(e.code!=='ERR_CANCELED') error.value=e.response?.data?.message || e.message }
  finally {loading.value=false}
}
onBeforeUnmount(()=>controller?.abort())
</script>
<style scoped>
.sql-query { display:flex; flex-direction:column; gap:14px; }.toolbar {display:flex;gap:12px;flex-wrap:wrap}.toolbar .el-select{width:160px}.hint{color:var(--ops-muted);font-size:13px;margin:0}.cell{white-space:pre-wrap;overflow-wrap:anywhere}
</style>
