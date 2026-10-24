<template>
  <div class="trace-query">
    <div class="toolbar"><el-input v-model="traceID" placeholder="输入 16 或 32 位 Trace ID" aria-label="Trace ID" @keyup.enter="execute" /><el-button type="primary" :loading="loading" @click="execute">查询链路</el-button></div>
    <el-alert v-if="error" :title="error" type="error" :closable="false" />
    <el-empty v-if="executed && !traces.length" description="未找到链路" />
    <section v-for="trace in traces" :key="trace.traceID">
      <p>{{ trace.traceID }} · {{ trace.spans.length }} 个 Span · {{ (duration(trace)/1000).toFixed(2) }} ms</p>
      <el-table :data="sortedSpans(trace)" max-height="650" @row-click="showSpan">
        <el-table-column label="服务 / 操作" min-width="240"><template #default="{ row }"><span :style="{ paddingLeft: `${depth(trace,row)*16}px` }">{{ trace.processes?.[row.processID]?.serviceName }} / {{ row.operationName }}</span></template></el-table-column>
        <el-table-column label="耗时" width="130"><template #default="{ row }">{{ (row.duration / 1000).toFixed(2) }} ms</template></el-table-column>
        <el-table-column label="调用时间线" min-width="280"><template #default="{ row }"><div class="timeline"><div class="bar" :style="{ marginLeft: `${(row.startTime-start(trace))/duration(trace)*100}%`, width: `${Math.max(0.5,row.duration/duration(trace)*100)}%` }" /></div></template></el-table-column>
      </el-table>
    </section>
    <el-drawer v-model="details" title="Span 详情" size="min(700px,95vw)"><pre>{{ JSON.stringify(span,null,2) }}</pre></el-drawer>
  </div>
</template>
<script setup>
import { ref,onBeforeUnmount } from 'vue'
import request from '../../api/request'
const props=defineProps({source:{type:Object,required:true}})
const traceID=ref(''),loading=ref(false),error=ref(''),traces=ref([]),executed=ref(false),details=ref(false),span=ref(null)
const sortedSpans=trace=>[...trace.spans].sort((a,b)=>a.startTime-b.startTime)
function depth(trace,row){
 const spans=new Map(trace.spans.map(item=>[item.spanID,item])), seen=new Set([row.spanID]);let level=0,current=row
 while(level<20){const parent=current.references?.find(item=>item.refType==='CHILD_OF')?.spanID;if(!parent||seen.has(parent)||!spans.has(parent))break;seen.add(parent);current=spans.get(parent);level++}
 return level
}
const start=trace=>Math.min(...trace.spans.map(item=>item.startTime))
const duration=trace=>trace.spans.length ? Math.max(1,Math.max(...trace.spans.map(item=>item.startTime+item.duration))-start(trace)) : 1
function showSpan(row){span.value=row;details.value=true}
let controller
async function execute(){
 if(loading.value)return
 error.value='';traces.value=[];executed.value=false
 if(!/^(?:[a-f\d]{16}|[a-f\d]{32})$/i.test(traceID.value.trim())){error.value='请输入有效的 Trace ID';return}
 loading.value=true;controller=new AbortController()
 try{const response=await request({url:'/query/traces',method:'post',timeout:35000,signal:controller.signal,data:{datasource_id:props.source.id,trace_id:traceID.value.trim()}});traces.value=response.data;executed.value=true}
 catch(e){if(e.code!=='ERR_CANCELED')error.value=e.response?.data?.message||e.message}
 finally{loading.value=false}
}
onBeforeUnmount(()=>controller?.abort())
</script>
<style scoped>
.trace-query{display:flex;flex-direction:column;gap:16px}.toolbar{display:flex;gap:12px}.timeline{height:16px;background:var(--ops-surface-soft);overflow:hidden}.bar{height:100%;background:var(--ops-primary);border-radius:3px}pre{white-space:pre-wrap;overflow-wrap:anywhere}
</style>
