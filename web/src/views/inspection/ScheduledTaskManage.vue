<template>
  <div class="schedule-page">
    <header class="page-header">
      <div class="page-heading"><h1>定时任务</h1><p>在指定时间触发已有脚本或工作流，执行内容仍由脚本库和工作流统一维护</p></div>
      <div class="page-header-actions"><el-button type="primary" :icon="Plus" @click="openCreate" v-permission="['task:create']">创建定时任务</el-button></div>
    </header>

    <el-card class="schedule-card" shadow="never">
      <div class="table-heading"><div><strong>调度列表</strong><span>共 {{ tasks.length }} 个计划</span></div><span class="scheduler-state"><i></i>调度器运行中</span></div>
      <el-table :data="tasks" v-loading="loading" stripe>
        <el-table-column label="任务名称" min-width="190"><template #default="{ row }"><div class="task-name"><strong>{{ row.name }}</strong><small>{{ row.description || targetName(row) }}</small></div></template></el-table-column>
        <el-table-column label="执行类型" width="110"><template #default="{ row }"><el-tag :type="row.type === 'workflow' ? 'primary' : row.type === 'script' ? 'success' : 'info'" effect="plain" size="small">{{ targetTypeLabel(row.type) }}</el-tag></template></el-table-column>
        <el-table-column label="执行对象" min-width="170"><template #default="{ row }">{{ targetName(row) }}</template></el-table-column>
        <el-table-column label="执行计划" width="170"><template #default="{ row }"><div class="cron-cell"><strong>{{ cronLabel(row.cron_expr) }}</strong><code>{{ row.cron_expr }}</code></div></template></el-table-column>
        <el-table-column label="上次执行" width="150"><template #default="{ row }">{{ row.last_run_at ? formatTime(row.last_run_at) : '尚未执行' }}</template></el-table-column>
        <el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '已启用' : '已停用' }}</el-tag></template></el-table-column>
        <el-table-column label="操作" width="230" fixed="right"><template #default="{ row }"><el-button link type="success" :loading="runId === row.id" @click="runOne(row)" v-permission="['task:run']">立即执行</el-button><el-button link type="primary" @click="openEdit(row)" v-permission="['task:edit']">编辑</el-button><el-button link type="info" @click="showHistory(row)">记录</el-button><el-button link type="danger" @click="removeOne(row)" v-permission="['task:delete']">删除</el-button></template></el-table-column>
        <template #empty><TableEmpty /></template>
      </el-table>
    </el-card>

    <el-drawer v-model="drawerVisible" class="scheduled-task-drawer" :title="form.id ? '编辑定时任务' : '创建定时任务'" size="min(720px, 94vw)" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <section class="drawer-section">
          <div class="section-title"><strong>基本信息</strong><span>定时任务只负责何时触发，不保存执行内容</span></div>
          <el-form-item label="任务名称" prop="name"><el-input v-model="form.name" maxlength="100" placeholder="例如：每天执行发布后巡检" /></el-form-item>
          <el-form-item label="执行类型"><el-segmented v-model="form.type" :options="typeOptions" @change="form.target_id = null" /></el-form-item>
          <el-form-item :label="form.type === 'workflow' ? '选择工作流' : '选择脚本'" prop="target_id">
            <el-select v-model="form.target_id" filterable :placeholder="form.type === 'workflow' ? '选择一个已启用的工作流' : '选择一个已启用的脚本'">
              <el-option v-for="item in selectableTargets" :key="item.id" :label="item.name" :value="item.id"><span>{{ item.name }}</span><small v-if="form.type === 'script'" class="target-purpose">{{ item.type === 'python' ? 'Python' : 'Shell' }}</small></el-option>
            </el-select>
          </el-form-item>
          <div class="reference-note"><el-icon><Link /></el-icon><span>任务执行时读取最新版本的{{ form.type === 'workflow' ? '工作流' : '脚本' }}，后续修改内容不需要重新配置定时任务。</span></div>
        </section>

        <section class="drawer-section">
          <div class="section-title"><strong>执行计划</strong><span>选择常用时间，或使用自定义 Cron</span></div>
          <el-form-item label="什么时候执行"><el-select v-model="scheduleValue" @change="applySchedule"><el-option v-for="option in scheduleOptions" :key="option.value" :label="option.label" :value="option.value" /><el-option label="自定义 Cron" value="custom" /></el-select></el-form-item>
          <el-form-item v-if="scheduleValue === 'custom'" label="Cron 表达式" prop="cron_expr"><el-input v-model="form.cron_expr" placeholder="*/5 * * * *" /></el-form-item>
          <div class="next-run-preview"><span>当前计划</span><strong>{{ cronLabel(form.cron_expr) }}</strong><code>{{ form.cron_expr }}</code></div>
        </section>

        <section class="drawer-section">
          <div class="section-title"><strong>故障接入</strong><span>执行失败进入故障中心，并复用分组、静默、抑制与通知路由</span></div>
          <div class="alert-grid">
            <el-form-item label="故障中心" prop="fault_center_id"><el-select v-model="form.fault_center_id" filterable placeholder="选择故障中心"><el-option v-for="center in activeCenters" :key="center.id" :label="center.name" :value="center.id" /></el-select></el-form-item>
            <el-form-item label="告警等级"><el-select v-model="form.severity"><el-option label="P0" value="P0" /><el-option label="P1" value="P1" /><el-option label="P2" value="P2" /></el-select></el-form-item>
            <el-form-item label="连续失败次数"><el-input-number v-model="form.failure_threshold" :min="1" :max="100" controls-position="right" /><span class="field-help">达到阈值后进入告警中</span></el-form-item>
          </div>
          <div class="labels-panel">
            <div class="labels-copy"><strong>额外标签</strong><span>与告警规则、拨测任务使用同一命名空间</span></div>
            <key-value-editor v-model="form.external_labels" key-placeholder="标签键，如 service" value-placeholder="标签值，如 order" />
            <p>通知模板使用 <code>${external_labels.service}</code>；运行标签使用 <code>${labels.task_name}</code>。</p>
          </div>
          <div class="reference-note"><el-icon><Link /></el-icon><span>首次失败创建待触发告警，达到阈值后通知；下一次执行成功时自动恢复并进入历史告警。</span></div>
        </section>

        <section class="drawer-section">
          <div class="section-title"><strong>其他设置</strong><span>补充说明并控制是否启用</span></div>
          <el-form-item label="说明"><el-input v-model="form.description" type="textarea" :rows="3" maxlength="500" placeholder="可选，记录这个计划的用途" /></el-form-item>
          <div class="status-row"><div><strong>任务状态</strong><span>{{ form.status === 1 ? '保存后按计划自动执行' : '保存但暂不自动执行' }}</span></div><el-switch v-model="form.status" :active-value="1" :inactive-value="0" /></div>
        </section>
      </el-form>
      <template #footer><div class="drawer-footer"><el-button @click="drawerVisible = false">取消</el-button><el-button type="primary" :loading="saving" @click="saveTask">{{ form.id ? '保存修改' : '创建定时任务' }}</el-button></div></template>
    </el-drawer>

    <el-dialog v-model="runVisible" title="执行结果" width="650px">
      <div v-if="runResult" class="run-result"><div class="result-state" :class="runResult.status === 'success' ? 'success' : 'failed'"><strong>{{ runResult.status === 'success' ? '执行成功' : '执行失败' }}</strong><span>耗时 {{ runResult.duration_ms }}ms</span></div><pre>{{ runResult.output || '(无输出)' }}</pre></div>
      <template #footer><el-button @click="runVisible = false">关闭</el-button></template>
    </el-dialog>

    <el-dialog v-model="historyVisible" :title="`执行记录 · ${historyName}`" width="780px">
      <el-table :data="historyData" v-loading="historyLoading" stripe max-height="430"><el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">{{ row.status === 'success' ? '成功' : '失败' }}</el-tag></template></el-table-column><el-table-column prop="duration_ms" label="耗时" width="90"><template #default="{ row }">{{ row.duration_ms }}ms</template></el-table-column><el-table-column prop="output" label="输出" min-width="360" show-overflow-tooltip /><el-table-column prop="created_at" label="执行时间" width="170" /><template #empty><TableEmpty /></template></el-table>
      <template #footer><el-button @click="historyVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { Delete, Link, Plus } from '@element-plus/icons-vue'
import { ElButton, ElInput, ElMessage, ElMessageBox } from 'element-plus'
import { createScheduledTask, deleteScheduledTask, getScheduledTasks, getTaskExecutions, runScheduledTask, updateScheduledTask } from '../../api/scheduledTask'
import { getScripts } from '../../api/script'
import { getWorkflows } from '../../api/workflow'
import { getFaultCenters } from '../../api/alert'

const KeyValueEditor = defineComponent({ name: 'KeyValueEditor', props: { modelValue: { type: Array, required: true }, keyPlaceholder: { type: String, default: '键' }, valuePlaceholder: { type: String, default: '值' } }, emits: ['update:modelValue'], setup(props, { emit }) { const update = (index, field, value) => emit('update:modelValue', props.modelValue.map((row, i) => i === index ? { ...row, [field]: value } : row)); const add = () => emit('update:modelValue', [...props.modelValue, { key: '', value: '' }]); const remove = index => { const rows = props.modelValue.filter((_, i) => i !== index); emit('update:modelValue', rows.length ? rows : [{ key: '', value: '' }]) }; return () => h('div', { class: 'kv-editor' }, props.modelValue.map((row, index) => h('div', { class: 'kv-row', key: index }, [h(ElInput, { modelValue: row.key, placeholder: props.keyPlaceholder, 'onUpdate:modelValue': value => update(index, 'key', value) }), h(ElInput, { modelValue: row.value, placeholder: props.valuePlaceholder, 'onUpdate:modelValue': value => update(index, 'value', value) }), h(ElButton, { icon: index === props.modelValue.length - 1 ? Plus : Delete, circle: true, onClick: index === props.modelValue.length - 1 ? add : () => remove(index) })]))) } })

const typeOptions = [{ label: '执行脚本', value: 'script' }, { label: '执行工作流', value: 'workflow' }]
const scheduleOptions = [{ label: '每 1 分钟', value: '*/1 * * * *' }, { label: '每 5 分钟', value: '*/5 * * * *' }, { label: '每 30 分钟', value: '*/30 * * * *' }, { label: '每小时整点', value: '0 * * * *' }, { label: '每 2 小时', value: '0 */2 * * *' }, { label: '每天 00:00', value: '0 0 * * *' }, { label: '每天 09:00', value: '0 9 * * *' }, { label: '工作日 09:00', value: '0 9 * * 1-5' }]
const tasks = ref([]), scripts = ref([]), workflows = ref([]), faultCenters = ref([]), loading = ref(false), saving = ref(false)
const drawerVisible = ref(false), formRef = ref(), scheduleValue = ref('*/5 * * * *')
const form = reactive({ id: null, name: '', type: 'script', target_id: null, cron_expr: '*/5 * * * *', description: '', external_labels: [{ key: '', value: '' }], severity: 'P2', failure_threshold: 1, fault_center_id: null, status: 1 })
const rules = { name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }], target_id: [{ required: true, message: '请选择执行对象', trigger: 'change' }], cron_expr: [{ required: true, message: '请输入 Cron 表达式', trigger: 'blur' }], fault_center_id: [{ required: true, message: '请选择故障中心', trigger: 'change' }] }
const selectableTargets = computed(() => (form.type === 'workflow' ? workflows.value : scripts.value).filter(item => item.status === 1))
const activeCenters = computed(() => faultCenters.value.filter(item => item.status === 1))
const runId = ref(null), runVisible = ref(false), runResult = ref(null)
const historyVisible = ref(false), historyName = ref(''), historyData = ref([]), historyLoading = ref(false)

onMounted(loadData)
async function loadData() { loading.value = true; try { const [taskResponse, scriptResponse, workflowResponse, centerResponse] = await Promise.all([getScheduledTasks(), getScripts(), getWorkflows(), getFaultCenters()]); tasks.value = taskResponse.data || []; scripts.value = scriptResponse.data || []; workflows.value = workflowResponse.data || []; faultCenters.value = centerResponse.data || [] } finally { loading.value = false } }
function openCreate() { Object.assign(form, { id: null, name: '', type: 'script', target_id: null, cron_expr: '*/5 * * * *', description: '', external_labels: [{ key: '', value: '' }], severity: 'P2', failure_threshold: 1, fault_center_id: activeCenters.value.length === 1 ? activeCenters.value[0].id : null, status: 1 }); scheduleValue.value = form.cron_expr; drawerVisible.value = true }
function openEdit(row) { Object.assign(form, { id: row.id, name: row.name, type: row.type, target_id: row.target_id, cron_expr: row.cron_expr, description: row.description || '', external_labels: objectToRows(parseObject(row.external_labels)), severity: row.severity || 'P2', failure_threshold: row.failure_threshold || 1, fault_center_id: row.fault_center_id || null, status: row.status }); scheduleValue.value = scheduleOptions.some(option => option.value === row.cron_expr) ? row.cron_expr : 'custom'; drawerVisible.value = true }
function applySchedule(value) { if (value !== 'custom') form.cron_expr = value }
async function saveTask() { const labelError = validateLabels(form.external_labels); if (labelError) return ElMessage.warning(labelError); if (!await formRef.value.validate().catch(() => false)) return; saving.value = true; try { const payload = { ...form, external_labels: JSON.stringify(rowsToObject(form.external_labels)) }; form.id ? await updateScheduledTask(form.id, payload) : await createScheduledTask(payload); ElMessage.success(form.id ? '定时任务已更新' : '定时任务已创建'); drawerVisible.value = false; await loadData() } finally { saving.value = false } }
async function removeOne(row) { await ElMessageBox.confirm(`确定删除定时任务「${row.name}」吗？`, '提示', { type: 'warning' }); await deleteScheduledTask(row.id); ElMessage.success('已删除'); loadData() }
async function runOne(row) { runId.value = row.id; try { const response = await runScheduledTask(row.id); runResult.value = response.data; runVisible.value = true; await loadData() } finally { runId.value = null } }
async function showHistory(row) { historyName.value = row.name; historyVisible.value = true; historyLoading.value = true; try { const response = await getTaskExecutions({ task_id: row.id }); historyData.value = response.data.list || [] } finally { historyLoading.value = false } }
function targetTypeLabel(type) { return type === 'workflow' ? '工作流' : '脚本' }
function targetName(row) { const list = row.type === 'workflow' ? workflows.value : scripts.value; return list.find(item => item.id === row.target_id)?.name || '执行对象已删除' }
function cronLabel(expression) { return scheduleOptions.find(option => option.value === expression)?.label || (expression ? '自定义计划' : '尚未配置') }
function formatTime(value) { const date = new Date(value); return Number.isNaN(date.getTime()) ? value : date.toLocaleString('zh-CN', { hour12: false }) }
function rowsToObject(rows) { return rows.reduce((result, row) => { const key = String(row.key || '').trim(); if (key) result[key] = String(row.value || '').trim(); return result }, {}) }
function objectToRows(value) { const rows = Object.entries(value || {}).map(([key, entry]) => ({ key, value: entry })); return rows.length ? rows : [{ key: '', value: '' }] }
function parseObject(value) { if (typeof value === 'object' && value) return value; try { return JSON.parse(value || '{}') } catch { return {} } }
function validateLabels(rows) { const filled = rows.filter(item => String(item.key || '').trim() || String(item.value || '').trim()); if (filled.length > 20) return '额外标签最多 20 个'; const keys = new Set(); for (const row of filled) { const key = String(row.key || '').trim(); if (!key) return '请填写标签键'; if (!/^[A-Za-z0-9_.-]+$/.test(key)) return '标签键只能包含字母、数字、点、下划线和短横线'; const normalized = key.toLowerCase(); if (keys.has(normalized)) return `标签键 ${key} 重复`; keys.add(normalized) } return '' }
</script>

<style scoped>
.schedule-page{--ink:#17233b;--muted:#758399;--border:#e1e8f1;display:flex;flex-direction:column;gap:16px}.schedule-hero{display:flex;align-items:center;gap:16px;min-height:92px;padding:20px 24px;background:linear-gradient(115deg,#fff,#f6f9fe 60%,#edf5ff);border:1px solid #dce7f6;border-radius:10px;box-shadow:0 8px 28px rgba(35,78,134,.06)}.hero-icon{display:inline-flex;align-items:center;justify-content:center;width:48px;height:48px;flex:0 0 48px;color:#fff;background:linear-gradient(145deg,#2478e8,#0f5bc6);border-radius:12px;box-shadow:0 8px 18px rgba(23,105,224,.22);font-size:24px}.hero-copy{margin-right:auto}.hero-copy h1{margin:0;color:var(--ink);font-size:22px;font-weight:750}.hero-copy p{margin:6px 0 0;color:var(--muted);font-size:13px}.schedule-card{border-color:var(--border);border-radius:10px;box-shadow:0 8px 28px rgba(31,50,81,.045)}.schedule-card :deep(.el-card__body){padding:0}.table-heading{display:flex;align-items:center;justify-content:space-between;padding:17px 20px;border-bottom:1px solid var(--border)}.table-heading>div{display:flex;align-items:baseline;gap:9px}.table-heading strong{color:var(--ink);font-size:14px}.table-heading span{color:#8b97a8;font-size:12px}.scheduler-state{display:inline-flex;align-items:center;gap:7px}.scheduler-state i{width:7px;height:7px;background:#17a67a;border-radius:50%;box-shadow:0 0 0 4px rgba(23,166,122,.1)}.task-name,.cron-cell{display:flex;flex-direction:column;gap:3px}.task-name strong,.cron-cell strong{color:#28384f;font-size:13px}.task-name small{overflow:hidden;color:#8a96a8;font-size:11px;text-overflow:ellipsis;white-space:nowrap}.cron-cell code{color:#8290a3;font-size:10px}.run-result .result-state{display:flex;align-items:center;justify-content:space-between;padding:13px 15px;border-radius:8px}.result-state.success{color:#18755b;background:#f1faf6;border:1px solid #d5eee4}.result-state.failed{color:#a1423a;background:#fff5f4;border:1px solid #f2d9d6}.run-result pre{max-height:300px;margin:12px 0 0;padding:14px;overflow:auto;color:#d7e0ec;background:#111b2e;border-radius:8px;font:12px/1.6 'SF Mono',Monaco,Menlo,Consolas,monospace;white-space:pre-wrap;overflow-wrap:anywhere}@media(max-width:700px){.schedule-hero{align-items:flex-start;flex-wrap:wrap;padding:18px}.hero-copy{width:calc(100% - 64px)}}
</style>

<style>
.scheduled-task-drawer .drawer-section{margin-bottom:18px;padding:20px;background:#fff;border:1px solid #e1e8f1;border-radius:10px}.scheduled-task-drawer .section-title{display:flex;align-items:baseline;gap:10px;margin-bottom:18px}.scheduled-task-drawer .section-title strong{color:#17233b;font-size:15px}.scheduled-task-drawer .section-title span{color:#8a96a8;font-size:12px}.scheduled-task-drawer .el-form-item{margin-bottom:18px}.scheduled-task-drawer .el-form-item__label{padding-bottom:7px;color:#46536a;font-size:13px;font-weight:650}.scheduled-task-drawer .el-select{width:100%}.scheduled-task-drawer .target-purpose{float:right;color:#96a0af}.scheduled-task-drawer .reference-note{display:flex;align-items:flex-start;gap:8px;padding:11px 13px;color:#58718e;background:#f4f8fd;border:1px solid #dce8f5;border-radius:7px;font-size:12px;line-height:1.55}.scheduled-task-drawer .reference-note .el-icon{margin-top:2px;flex:0 0 auto}.scheduled-task-drawer .next-run-preview{display:grid;grid-template-columns:90px minmax(0,1fr) auto;gap:12px;align-items:center;padding:13px 14px;background:#f8fafc;border:1px solid #e2e9f1;border-radius:8px}.scheduled-task-drawer .next-run-preview span{color:#8491a4;font-size:11px}.scheduled-task-drawer .next-run-preview strong{color:#32445d;font-size:13px}.scheduled-task-drawer .next-run-preview code{padding:3px 7px;color:#54708f;background:#edf3fa;border-radius:5px;font-size:11px}.scheduled-task-drawer .status-row{display:flex;align-items:center;justify-content:space-between}.scheduled-task-drawer .status-row>div{display:flex;flex-direction:column;gap:4px}.scheduled-task-drawer .status-row strong{color:#26364d;font-size:13px}.scheduled-task-drawer .status-row span{color:#8a96a8;font-size:11px}.scheduled-task-drawer .drawer-footer{display:flex;justify-content:flex-end;gap:10px}.scheduled-task-drawer .alert-grid{display:grid;grid-template-columns:1.2fr .7fr .8fr;gap:14px}.scheduled-task-drawer .alert-grid .el-input-number{width:100%}.scheduled-task-drawer .field-help{display:block;margin-top:5px;color:#8a96a8;font-size:11px}.scheduled-task-drawer .labels-panel{margin:0 0 14px;padding:14px;background:#f8fafc;border:1px solid #e1e8f1;border-radius:8px}.scheduled-task-drawer .labels-copy{display:flex;align-items:baseline;gap:9px;margin-bottom:11px}.scheduled-task-drawer .labels-copy strong{color:#34465d;font-size:12px}.scheduled-task-drawer .labels-copy span,.scheduled-task-drawer .labels-panel p{color:#8491a3;font-size:11px}.scheduled-task-drawer .labels-panel p{margin:10px 0 0}.scheduled-task-drawer .labels-panel code{padding:2px 5px;color:#285e9d;background:#edf4fc;border-radius:4px}.scheduled-task-drawer .kv-editor{display:flex;flex-direction:column;gap:8px}.scheduled-task-drawer .kv-row{display:grid;grid-template-columns:1fr 1fr 34px;gap:8px}@media(max-width:700px){.scheduled-task-drawer .alert-grid{grid-template-columns:1fr}}@media(max-width:600px){.scheduled-task-drawer .next-run-preview{grid-template-columns:1fr}.scheduled-task-drawer .next-run-preview code{justify-self:start}}
</style>
