<template>
  <div class="mcp-page">
    <section class="dashboard-intro" role="note" aria-label="MCP 服务说明">
      <el-icon class="intro-icon" aria-hidden="true"><InfoFilled /></el-icon>
      <span class="intro-message">通过标准 MCP 协议向 Codex、Claude、Cursor 等客户端开放平台工具。工具和参数由服务端统一维护，平台升级后客户端无需重新安装配置。</span>
    </section>

    <section class="service-grid" aria-label="MCP 服务概览">
      <el-card shadow="never" class="service-card endpoint-card">
        <div class="card-heading"><div><strong>MCP 服务地址</strong><span>Streamable HTTP</span></div><el-tag type="success" effect="light">运行中</el-tag></div>
        <div class="endpoint-line"><code>{{ service.endpoint || '-' }}</code><el-button :icon="DocumentCopy" @click="copyText(service.endpoint, '服务地址')">复制</el-button></div>
        <div class="service-meta">
          <span><small>认证方式</small>{{ service.authentication || '-' }}</span>
          <span><small>协议版本</small>{{ service.protocol_version || '-' }}</span>
          <span><small>服务版本</small>{{ service.server_version || '-' }}</span>
        </div>
      </el-card>

      <el-card shadow="never" class="service-card config-card">
        <div class="card-heading"><div><strong>客户端配置</strong><span>将 Token 替换为连接绑定的访问令牌</span></div><el-button link type="primary" :icon="DocumentCopy" @click="copyText(clientConfig, '配置')">复制配置</el-button></div>
        <pre><code>{{ clientConfig }}</code></pre>
      </el-card>
    </section>

    <el-card shadow="never" class="tools-card">
      <template #header><div class="section-heading"><div><strong>平台工具</strong><span>客户端连接后会自动发现当前可用工具，实际范围取决于访问令牌绑定用户的角色权限。</span></div><el-tag effect="light">{{ service.tools?.length || 0 }} 个工具</el-tag></div></template>
      <el-table :data="service.tools || []" stripe table-layout="fixed">
        <el-table-column prop="title" label="工具" min-width="180"><template #default="{ row }"><strong class="tool-title">{{ row.title }}</strong><code class="tool-name">{{ row.name }}</code></template></el-table-column>
        <el-table-column prop="description" label="用途" min-width="360" show-overflow-tooltip />
        <el-table-column prop="permission" label="所需权限" min-width="190"><template #default="{ row }"><code class="permission-code">{{ row.permission }}</code></template></el-table-column>
        <el-table-column label="类型" width="100" align="center"><template #default="{ row }"><el-tag :type="row.read_only ? 'info' : 'warning'" effect="light">{{ row.read_only ? '只读' : '执行' }}</el-tag></template></el-table-column>
        <template #empty><TableEmpty /></template>
      </el-table>
    </el-card>

    <div class="connections-heading"><div><strong>连接管理</strong><span>每个外部客户端使用独立访问令牌，便于权限隔离、停用和连接追踪。</span></div></div>
    <header class="task-toolbar">
      <div class="filter-bar" role="search" aria-label="MCP 连接筛选">
        <el-input v-model="filters.keyword" class="task-search" clearable :prefix-icon="Search" placeholder="搜索连接、用户或访问令牌" @input="searchLater" />
        <el-select v-model="filters.client_type" class="filter-select" clearable placeholder="客户端类型" @change="applyFilters"><el-option v-for="option in clientTypeOptions" :key="option.value" :label="option.label" :value="option.value" /></el-select>
        <el-select v-model="filters.compatibility" class="filter-select" clearable placeholder="兼容状态" @change="applyFilters"><el-option label="兼容" value="compatible" /><el-option label="需要更新" value="incompatible" /><el-option label="从未连接" value="never_connected" /></el-select>
        <el-select v-model="filters.status" class="status-filter" clearable placeholder="启用状态" @change="applyFilters"><el-option label="启用" :value="1" /><el-option label="停用" :value="0" /></el-select>
        <el-button :disabled="!hasFilters" @click="clearFilters">清空</el-button>
      </div>
      <div class="page-header-actions"><el-button :icon="Refresh" @click="load">刷新</el-button><el-button type="primary" :icon="Plus" @click="openCreate" v-permission="['mcp-service:edit']">创建连接</el-button></div>
    </header>

    <el-card shadow="never" class="task-card">
      <el-table v-loading="loading" :data="clients" stripe table-layout="fixed">
        <el-table-column prop="name" label="连接名称" min-width="200" show-overflow-tooltip><template #default="{ row }"><strong class="client-name">{{ row.name }}</strong><div class="client-id">{{ row.id }}</div></template></el-table-column>
        <el-table-column label="客户端" width="140"><template #default="{ row }"><el-tag effect="light" type="primary">{{ clientTypeLabel(row.client_type) }}</el-tag><span v-if="row.reported_name" class="secondary-text">{{ row.reported_name }}</span></template></el-table-column>
        <el-table-column label="访问身份" min-width="200" show-overflow-tooltip><template #default="{ row }"><div>{{ row.api_token_name || '-' }}</div><span class="secondary-text">{{ row.nickname || row.username || '-' }} · {{ row.token_prefix || '-' }}</span></template></el-table-column>
        <el-table-column label="版本" min-width="175"><template #default="{ row }"><div>{{ row.client_version || '-' }}</div><span class="secondary-text">{{ row.protocol_version || '尚未连接' }}</span></template></el-table-column>
        <el-table-column label="兼容状态" width="110" align="center"><template #default="{ row }"><el-tag :type="compatibilityMeta(row.compatibility).type" effect="light">{{ compatibilityMeta(row.compatibility).label }}</el-tag></template></el-table-column>
        <el-table-column label="最后连接" min-width="170"><template #default="{ row }"><span>{{ row.last_seen_at ? formatTime(row.last_seen_at) : '从未连接' }}</span><el-tooltip v-if="row.last_ip || row.last_user_agent" :content="connectionDetail(row)" placement="top-start"><div class="secondary-text connection-detail">{{ row.last_ip || '-' }}</div></el-tooltip></template></el-table-column>
        <el-table-column label="状态" width="80" align="center"><template #default="{ row }"><el-switch :model-value="row.status === 1" :aria-label="`${row.name}${row.status === 1 ? '停用' : '启用'}`" @change="changeStatus(row, $event)" v-permission="['mcp-service:edit']" /></template></el-table-column>
        <el-table-column label="操作" width="100" align="center" fixed="right"><template #default="{ row }"><div class="row-actions"><el-tooltip content="编辑"><el-button link type="primary" :icon="Edit" aria-label="编辑 MCP 连接" @click="openEdit(row)" v-permission="['mcp-service:edit']" /></el-tooltip><el-tooltip content="删除"><el-button link type="danger" :icon="Delete" aria-label="删除 MCP 连接" @click="remove(row)" v-permission="['mcp-service:edit']" /></el-tooltip></div></template></el-table-column>
        <template #empty><TableEmpty /></template>
      </el-table>
      <div class="task-pagination"><el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.page_size" :page-sizes="[10,20,50,100]" :total="pagination.total" layout="total, sizes, prev, pager, next" @current-change="loadClients" @size-change="pageSizeChanged" /></div>
    </el-card>

    <el-drawer v-model="drawerVisible" :title="editingID ? '编辑 MCP 连接' : '创建 MCP 连接'" size="min(620px, 94vw)" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <div class="form-grid"><el-form-item label="连接名称" prop="name"><el-input v-model="form.name" maxlength="100" placeholder="例如：张三的 Codex" /></el-form-item><el-form-item label="客户端类型" prop="client_type"><el-select v-model="form.client_type" placeholder="选择客户端类型"><el-option v-for="option in clientTypeOptions" :key="option.value" :label="option.label" :value="option.value" /></el-select></el-form-item></div>
        <el-form-item label="访问令牌" prop="api_token_id"><el-select v-model="form.api_token_id" filterable placeholder="选择尚未用于 MCP 连接的访问令牌"><el-option v-for="token in availableTokens" :key="token.id" :label="`${token.name} · ${token.nickname || token.username}`" :value="token.id"><span>{{ token.name }}</span><span class="token-option-meta">{{ token.nickname || token.username }} · {{ token.token_prefix }}</span></el-option></el-select></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.description" type="textarea" :rows="4" maxlength="500" show-word-limit placeholder="记录使用者、设备或接入用途" /></el-form-item>
        <div v-if="editingID" class="readonly-meta"><span>连接 ID</span><code>{{ editingID }}</code></div>
        <div class="status-line"><strong>启用连接</strong><el-switch v-model="form.enabled" /></div>
      </el-form>
      <template #footer><div class="drawer-footer"><el-button @click="drawerVisible=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">{{ editingID ? '保存' : '创建' }}</el-button></div></template>
    </el-drawer>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, DocumentCopy, Edit, InfoFilled, Plus, Refresh, Search } from '@element-plus/icons-vue'
import { createMCPClient, deleteMCPClient, getMCPClients, getMCPService, updateMCPClient, updateMCPClientStatus } from '../../api/mcpClient'
import { getAPITokens } from '../../api/apiToken'

const clientTypeOptions=[{label:'Codex',value:'codex'},{label:'Claude Desktop',value:'claude'},{label:'Cursor',value:'cursor'},{label:'自研客户端',value:'custom'}]
const loading=ref(false),saving=ref(false),drawerVisible=ref(false),editingID=ref(''),clients=ref([]),tokens=ref([]),formRef=ref(),searchTimer=ref()
const service=reactive({endpoint:'',authentication:'',protocol_version:'',server_version:'',tools:[]})
const filters=reactive({keyword:'',client_type:'',compatibility:'',status:''})
const pagination=reactive({page:1,page_size:10,total:0})
const blankForm=()=>({name:'',client_type:'codex',api_token_id:'',description:'',enabled:true})
const form=reactive(blankForm())
const rules={name:[{required:true,message:'请输入连接名称',trigger:'blur'}],client_type:[{required:true,message:'请选择客户端类型',trigger:'change'}],api_token_id:[{required:true,message:'请选择访问令牌',trigger:'change'}]}
const hasFilters=computed(()=>filters.keyword||filters.client_type||filters.compatibility||filters.status!=='')
const availableTokens=computed(()=>tokens.value.filter(item=>item.status===1&&!isExpired(item)&&(!item.mcp_client_id||item.mcp_client_id===editingID.value)))
const clientConfig=computed(()=>JSON.stringify({mcpServers:{'aiops-operations-center':{type:'streamable-http',url:service.endpoint||'https://your-aiops.example.com/api/mcp',headers:{Authorization:'Bearer <API_TOKEN>'}}}},null,2))

onMounted(load)
onBeforeUnmount(()=>clearTimeout(searchTimer.value))
async function load(){loading.value=true;try{const [serviceResult,tokenResult]=await Promise.all([getMCPService(),getAPITokens()]);Object.assign(service,serviceResult.data||{});service.endpoint=`${window.location.origin}/api/mcp`;tokens.value=tokenResult.data||[];await loadClients()}finally{loading.value=false}}
async function loadClients(){loading.value=true;try{const result=await getMCPClients({page:pagination.page,page_size:pagination.page_size,keyword:filters.keyword.trim()||undefined,client_type:filters.client_type||undefined,compatibility:filters.compatibility||undefined,status:filters.status===''?undefined:filters.status});clients.value=result.data?.list||[];pagination.total=result.data?.total||0}finally{loading.value=false}}
function searchLater(){clearTimeout(searchTimer.value);searchTimer.value=setTimeout(()=>{pagination.page=1;loadClients()},260)}
function applyFilters(){pagination.page=1;loadClients()}
function clearFilters(){Object.assign(filters,{keyword:'',client_type:'',compatibility:'',status:''});pagination.page=1;loadClients()}
function pageSizeChanged(){pagination.page=1;loadClients()}
function openCreate(){editingID.value='';Object.assign(form,blankForm());drawerVisible.value=true;formRef.value?.clearValidate()}
function openEdit(row){editingID.value=row.id;Object.assign(form,{name:row.name,client_type:row.client_type,api_token_id:row.api_token_id,description:row.description||'',enabled:row.status===1});drawerVisible.value=true;formRef.value?.clearValidate()}
async function save(){if(!await formRef.value.validate().catch(()=>false))return;saving.value=true;try{const wasEditing=Boolean(editingID.value);const payload={name:form.name.trim(),client_type:form.client_type,api_token_id:form.api_token_id,description:form.description.trim(),status:form.enabled?1:0};if(wasEditing)await updateMCPClient(editingID.value,payload);else await createMCPClient(payload);drawerVisible.value=false;await load();ElMessage.success(wasEditing?'MCP 连接已保存':'MCP 连接已创建')}finally{saving.value=false}}
async function changeStatus(row,value){await updateMCPClientStatus(row.id,value?1:0);await loadClients();ElMessage.success(value?'MCP 连接已启用':'MCP 连接已停用')}
async function remove(row){await ElMessageBox.confirm(`删除 MCP 连接「${row.name}」？访问令牌不会被删除。`,'删除 MCP 连接',{type:'warning'});await deleteMCPClient(row.id);await load();ElMessage.success('MCP 连接已删除')}
async function copyText(value,label){if(!value)return;if(navigator.clipboard?.writeText)await navigator.clipboard.writeText(value);else{const input=document.createElement('textarea');input.value=value;input.style.position='fixed';input.style.opacity='0';document.body.appendChild(input);input.select();document.execCommand('copy');input.remove()}ElMessage.success(`${label}已复制`)}
function clientTypeLabel(value){return clientTypeOptions.find(item=>item.value===value)?.label||value||'-'}
function compatibilityMeta(value){return {compatible:{label:'兼容',type:'success'},incompatible:{label:'需要更新',type:'danger'},never_connected:{label:'从未连接',type:'info'}}[value]||{label:'未知',type:'info'}}
function connectionDetail(row){return [`IP：${row.last_ip||'-'}`,`User-Agent：${row.last_user_agent||'-'}`].join('\n')}
function isExpired(row){return !row.permanent&&new Date(row.expires_at).getTime()<=Date.now()}
function formatTime(value){return new Intl.DateTimeFormat('zh-CN',{year:'numeric',month:'2-digit',day:'2-digit',hour:'2-digit',minute:'2-digit',second:'2-digit',hour12:false}).format(new Date(value))}
</script>

<style scoped>
.mcp-page{min-width:0}.dashboard-intro{display:flex;min-height:36px;box-sizing:border-box;align-items:center;gap:9px;margin-bottom:14px;padding:7px 12px;background:color-mix(in srgb,var(--ops-primary) 9%,var(--ops-surface));font-size:11px}.intro-icon{flex:none;color:var(--ops-primary);font-size:14px}.intro-message{min-width:0;color:var(--ops-muted)}.service-grid{display:grid;grid-template-columns:minmax(0,1fr) minmax(360px,.85fr);gap:14px;margin-bottom:14px}.service-card{min-width:0;border-color:var(--ops-border)}.service-card :deep(.el-card__body){padding:18px}.card-heading,.section-heading,.connections-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:16px}.card-heading>div,.section-heading>div,.connections-heading>div{display:flex;min-width:0;flex-direction:column;gap:4px}.card-heading strong,.section-heading strong,.connections-heading strong{color:var(--ops-text);font-size:14px}.card-heading span,.section-heading span,.connections-heading span{color:var(--ops-muted);font-size:11px}.endpoint-line{display:flex;align-items:center;gap:10px;margin:18px 0}.endpoint-line code{min-width:0;flex:1;overflow:hidden;padding:10px 12px;background:var(--ops-surface-soft);color:var(--ops-primary);font-size:12px;text-overflow:ellipsis;white-space:nowrap}.service-meta{display:grid;grid-template-columns:repeat(3,1fr);gap:12px}.service-meta span{display:flex;min-width:0;flex-direction:column;gap:4px;color:var(--ops-text);font-size:12px}.service-meta small{color:var(--ops-muted);font-size:10px}.config-card pre{height:126px;box-sizing:border-box;overflow:auto;margin:14px 0 0;padding:12px;background:#111827;color:#dbeafe;font:10px/1.55 ui-monospace,SFMono-Regular,Menlo,monospace}.tools-card,.task-card{min-width:0;border-color:var(--ops-border)}.tools-card{margin-bottom:18px}.tools-card :deep(.el-card__header){padding:14px 16px}.tools-card :deep(.el-card__body),.task-card :deep(.el-card__body){padding:0}.tool-title{display:block;color:var(--ops-text);font-size:12px}.tool-name,.permission-code{display:block;margin-top:3px;color:var(--ops-muted);font-size:10px}.connections-heading{margin-bottom:10px}.task-toolbar{display:flex;min-width:0;align-items:center;justify-content:space-between;gap:14px;margin-bottom:14px}.filter-bar,.page-header-actions{display:flex;min-width:0;align-items:center;gap:10px}.filter-bar{flex:1}.task-search{width:250px;flex:none}.filter-select{width:128px;flex:none}.status-filter{width:118px}.page-header-actions{flex:none}.client-name{color:var(--ops-text);font-size:13px}.client-id{margin-top:3px;overflow:hidden;color:var(--ops-muted);font:10px ui-monospace,SFMono-Regular,Menlo,monospace;text-overflow:ellipsis}.secondary-text{display:block;margin-top:3px;color:var(--ops-muted);font-size:11px}.connection-detail{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.row-actions{display:flex;align-items:center;justify-content:center;gap:5px}.task-pagination{display:flex;justify-content:flex-end;padding:14px 16px;border-top:1px solid var(--ops-border-soft)}.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:16px}.token-option-meta{float:right;margin-left:24px;color:var(--ops-muted);font-size:11px}.readonly-meta{display:flex;flex-direction:column;gap:7px;margin-bottom:18px;padding:12px 14px;background:var(--ops-surface-soft);border:1px solid var(--ops-border-soft)}.readonly-meta span{color:var(--ops-muted);font-size:11px}.readonly-meta code{overflow-wrap:anywhere;color:var(--ops-text);font-size:11px}.status-line{display:flex;align-items:center;justify-content:space-between;padding:14px 0;border-top:1px solid var(--ops-border-soft)}.status-line strong{font-size:13px}.drawer-footer{display:flex;justify-content:flex-end;gap:10px}@media(max-width:1180px){.service-grid{grid-template-columns:1fr}.task-toolbar{align-items:flex-start;flex-direction:column}.filter-bar{width:100%;flex-wrap:wrap}.page-header-actions{align-self:flex-end}}@media(max-width:720px){.task-search,.filter-select,.status-filter{width:100%}.form-grid,.service-meta{grid-template-columns:1fr}.page-header-actions{align-self:stretch;justify-content:flex-end}}
</style>
