<template>
  <div class="page-container">
    <el-card class="probe-card" shadow="never">
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" class="probe-form">
        <el-form-item class="probe-type-section" label="拨测类型">
          <template #label><span class="heading-with-help">拨测类型<HelpTooltip content="从指定节点执行 HTTP、ICMP、TCP 或证书到期检测。本次参数仅用于即时执行，不会保存为拨测任务，也不会触发告警通知。" label="即时拨测说明" /></span></template>
          <ProbeTypeSelect :model-value="form.type" @change="changeProbeType" />
        </el-form-item>

        <section class="form-section">
            <div class="section-heading">
              <div class="heading-with-help"><h2>基本配置</h2><HelpTooltip :content="currentType.hint" label="基本配置说明" /></div>
            </div>
            <div class="form-grid two-columns" >
              <el-form-item
                v-if="form.type === 'http'"
                class="full-row http-method-item"
                label="请求方式"
                prop="method"
              >
                <el-radio-group v-model="form.method">
                  <el-radio-button v-for="method in httpMethods" :key="method" :value="method">{{ method }}</el-radio-button>
                </el-radio-group>
              </el-form-item>
              <el-form-item v-if="form.type === 'http'" class="full-row" label="请求 URL" prop="target">
                <el-input
                  v-model="form.target"
                  type="textarea"
                  :autosize="{ minRows: 1, maxRows: 4 }"
                  resize="vertical"
                  class="url-input"
                  placeholder="例如 https://api.example.com:8443/v1/health"
                />
              </el-form-item>
              <el-form-item v-if="form.type !== 'http'" :label="form.type === 'certificate_expiry' ? '网站域名' : '目标 Host'" prop="host">
                <el-input v-model="form.host" placeholder="被探测的 IP 或域名，如 api.example.com" />
              </el-form-item>
              <el-form-item v-if="['tcp', 'certificate_expiry'].includes(form.type)" label="端口" prop="port">
                <el-input-number v-model="form.port" :min="1" :max="65535" controls-position="right" />
              </el-form-item>
              <el-form-item v-if="form.type === 'certificate_expiry'" label="提前多少天告警" prop="expiry_alert_days">
                <template #label><span class="heading-with-help">提前多少天告警<HelpTooltip content="用于判断是否即将到期；即时检测只展示结果，不发送告警通知。" label="到期阈值说明" /></span></template>
                <el-input-number v-model="form.expiry_alert_days" :min="1" :max="365" :precision="0" controls-position="right" />
              </el-form-item>
              <el-form-item v-if="form.type === 'icmp'" label="发包数量">
                <el-input-number v-model="form.count" :min="1" :max="10" controls-position="right" />
              </el-form-item>
            </div>

            <el-form-item v-if="form.type === 'http' && requestBodyVisible" label="请求 Body">
              <el-input v-model="form.body" type="textarea" :rows="5" class="mono-input" placeholder='{"name":"value"}' />
            </el-form-item>
        </section>

        <section v-if="form.type === 'http'" class="form-section">
            <div class="section-heading">
              <div class="heading-with-help"><h2>请求配置</h2><HelpTooltip content="按需附加请求头与查询参数。" label="请求配置说明" /></div>
            </div>
            <el-form-item label="请求 Headers">
              <key-value-editor v-model="form.headers" key-placeholder="Header 名称" value-placeholder="Header 值" />
            </el-form-item>
            <el-form-item label="查询 Params">
              <key-value-editor v-model="form.params" key-placeholder="参数名称" value-placeholder="参数值" />
            </el-form-item>
        </section>

        <el-collapse v-model="activePanels" class="advanced-collapse">
            <el-collapse-item name="advanced">
              <template #title>
                <span class="advanced-title">高级设置</span>
                <HelpTooltip :content="form.type === 'http' ? '超时、认证与响应校验' : '设置本次拨测超时时间'" label="高级设置说明" />
              </template>
              <div class="form-grid two-columns advanced-grid" >
                <el-form-item label="超时时间（ms）" :class="{ 'full-row': form.type !== 'http' }">
                  <el-input-number v-model="form.timeout_ms" :min="100" :max="300000" :step="1000" controls-position="right" />
                </el-form-item>
                <el-form-item v-if="form.type === 'certificate_expiry'" label="SNI / Server Name"><el-input v-model="form.server_name" placeholder="留空时使用网站域名" /></el-form-item>
                <el-form-item v-if="form.type === 'http'" label="Content-Type">
                  <el-select v-model="form.content_type" filterable allow-create default-first-option>
                    <el-option label="application/json" value="application/json" />
                    <el-option label="application/x-www-form-urlencoded" value="application/x-www-form-urlencoded" />
                    <el-option label="text/plain" value="text/plain" />
                  </el-select>
                </el-form-item>
                <el-form-item v-if="form.type === 'http'" class="full-row auth-type-item" label="认证方式">
                  <el-radio-group v-model="form.auth_type">
                    <el-radio-button value="none">无认证</el-radio-button>
                    <el-radio-button value="basic">Basic Auth</el-radio-button>
                  </el-radio-group>
                </el-form-item>
                <template v-if="form.type === 'http' && form.auth_type === 'basic'">
                  <el-form-item label="用户名" prop="username">
                    <el-input v-model="form.username" />
                  </el-form-item>
                  <el-form-item label="密码">
                    <el-input v-model="form.password" type="password" show-password autocomplete="off" />
                  </el-form-item>
                </template>
                <el-form-item v-if="form.type === 'http'" label="成功状态码">
                  <el-input v-model="form.success_codes" placeholder="200-399 或 200,201,204" />
                </el-form-item>
                <el-form-item v-if="form.type === 'http'" label="响应关键字（包含）">
                  <el-input v-model="form.response_keyword" clearable placeholder='可选，例如 "status":200' />
                </el-form-item>

              </div>
            </el-collapse-item>
        </el-collapse>

        <section class="form-section execution-section">
          <div class="section-heading">
            <div class="heading-with-help"><h2>执行节点</h2><HelpTooltip content="默认由服务端直接发起；探测内网或指定网络区域时，可切换到在线的 SSH、Agent 或 Docker 节点。" label="执行节点说明" /></div>
          </div>
          <div class="execution-panel">
            <el-form-item class="execution-node-field" prop="node_value">
              <el-select v-model="form.node_value" filterable aria-label="执行节点" placeholder="请选择执行节点">
                <el-option label="服务端直连" value="direct" />
                <el-option-group v-if="sshNodes.length" label="SSH 节点">
                  <el-option
                    v-for="node in sshNodes"
                    :key="`ssh:${node.id}`"
                    :label="`${node.name}（${node.host}:${node.port || 22}）`"
                    :value="`ssh:${node.id}`"
                    :disabled="node.status !== 1"
                  />
                </el-option-group>
                <el-option-group v-if="hostAgentNodes.length" label="Agent 节点">
                  <el-option
                    v-for="node in hostAgentNodes"
                    :key="`agent:${node.id}`"
                    :label="`${node.name}（${node.status === 'online' ? '在线' : '离线'}）`"
                    :value="`agent:${node.id}`"
                    :disabled="node.status !== 'online'"
                  />
                </el-option-group>
                <el-option-group v-if="dockerAgentNodes.length" label="Docker 节点">
                  <el-option
                    v-for="node in dockerAgentNodes"
                    :key="`agent:${node.id}`"
                    :label="`${node.name}（${node.status === 'online' ? '在线' : '离线'}）`"
                    :value="`agent:${node.id}`"
                    :disabled="node.status !== 'online'"
                  />
                </el-option-group>
              </el-select>
            </el-form-item>
          </div>
        </section>

        <div class="form-actions">
          <el-button @click="resetForm">重置</el-button>
          <el-button type="primary" :loading="probing" @click="handleProbe">开始拨测</el-button>
        </div>
      </el-form>
    </el-card>

    <el-card v-if="result" class="result-card" shadow="never">
      <div class="result-header" :class="result.status === 'success' ? 'is-success' : 'is-failed'">
        <span class="result-status-mark" aria-hidden="true">{{ result.status === 'success' ? '✓' : '!' }}</span>
        <div class="result-heading">
          <div class="heading-with-help"><h2>{{ result.status === 'success' ? '拨测成功' : '拨测失败' }}</h2><HelpTooltip :content="result.message || result.error || '拨测已完成'" label="拨测结果说明" /></div>
        </div>
      </div>

      <div class="result-metrics">
        <div class="result-metric">
          <span>拨测类型</span>
          <strong>{{ currentType.label }}</strong>
        </div>
        <div class="result-metric">
          <span>响应状态</span>
          <strong>{{ result.status_code || (result.status === 'success' ? '已连通' : '未通过') }}</strong>
        </div>
        <div class="result-metric">
          <span>执行耗时</span>
          <strong>{{ result.duration_ms }} <small>ms</small></strong>
        </div>
        <div class="result-metric">
          <span>执行节点</span>
          <strong>{{ result.node_name || '服务端直连' }}</strong>
        </div>
      </div>

      <div v-if="result.probe_type === 'certificate_expiry'" class="result-details"><CertificateExpiryResult :result="result.result" /></div>
      <div v-else-if="result.url || result.host" class="result-details">
        <div v-if="result.url" class="result-detail-row">
          <span>请求地址</span>
          <code>{{ result.method }} {{ result.url }}</code>
        </div>
        <div v-if="result.host" class="result-detail-row">
          <span>探测目标</span>
          <code>{{ result.host }}{{ result.port ? `:${result.port}` : '' }}</code>
        </div>

      </div>

      <div class="result-output">
        <div class="result-output-heading">
          <strong>{{ form.type === 'http' ? '响应内容' : '执行输出' }}</strong>
          <span v-if="result.content_type">{{ result.content_type }}</span>
        </div>
        <pre class="result-box">{{ formatResult(result) }}</pre>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { ElButton, ElInput, ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import { instantProbe } from '../../api/probe'
import { getAgents, getCertificateExecutors } from '../../api/certificate'
import HelpTooltip from '../../components/HelpTooltip.vue'
import ProbeTypeSelect from '../../components/ProbeTypeSelect.vue'
import CertificateExpiryResult from '../../components/CertificateExpiryResult.vue'
import { probeTypes } from '../../utils/probeTypes'
import { isDockerAgent } from '../../utils/agentNode'

const KeyValueEditor = defineComponent({
  name: 'KeyValueEditor',
  props: {
    modelValue: { type: Array, required: true },
    keyPlaceholder: { type: String, default: '键' },
    valuePlaceholder: { type: String, default: '值' },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const update = (index, field, value) => {
      const rows = props.modelValue.map((row, rowIndex) => rowIndex === index ? { ...row, [field]: value } : row)
      emit('update:modelValue', rows)
    }
    const add = () => emit('update:modelValue', [...props.modelValue, { key: '', value: '' }])
    const remove = (index) => {
      const rows = props.modelValue.filter((_, rowIndex) => rowIndex !== index)
      emit('update:modelValue', rows.length ? rows : [{ key: '', value: '' }])
    }
    return () => h('div', { class: 'kv-editor' }, props.modelValue.map((row, index) => h('div', { class: 'kv-row', key: index }, [
      h(ElInput, { modelValue: row.key, placeholder: props.keyPlaceholder, 'onUpdate:modelValue': value => update(index, 'key', value) }),
      h(ElInput, { modelValue: row.value, placeholder: props.valuePlaceholder, 'onUpdate:modelValue': value => update(index, 'value', value) }),
      h(ElButton, { icon: index === props.modelValue.length - 1 ? Plus : Delete, circle: true, onClick: index === props.modelValue.length - 1 ? add : () => remove(index), 'aria-label': index === props.modelValue.length - 1 ? '新增一行' : '删除此行' }),
    ])))
  },
})

const formRef = ref()
const probing = ref(false)
const result = ref(null)
const activePanels = ref([])
const sshNodes = ref([])
const agentNodes = ref([])
const hostAgentNodes = computed(() => agentNodes.value.filter(item => !isDockerAgent(item)))
const dockerAgentNodes = computed(() => agentNodes.value.filter(isDockerAgent))
const httpMethods = ['GET', 'POST', 'PUT', 'DELETE']

const defaultForm = (type = 'http') => ({
  type,
  target: '',
  host: '',
  port: probeTypes.find(item => item.value === type)?.defaultPort ?? 80,
  method: 'GET',
  headers: [{ key: '', value: '' }],
  params: [{ key: '', value: '' }],
  body: '',
  timeout_ms: 10000,
  content_type: 'application/json',
  auth_type: 'none',
  username: '',
  password: '',
  success_codes: '200-399',
  response_keyword: '',
  expiry_alert_days: 30,
  count: 4,
  server_name: '',
  node_value: 'direct',
})

const form = reactive(defaultForm())
const requestBodyVisible = computed(() => ['POST', 'PUT'].includes(form.method))
const currentType = computed(() => probeTypes.find(item => item.value === form.type) || probeTypes[0])

const validateHost = (_rule, value, callback) => {
  if (form.type === 'http') return callback()
  if (String(value || '').trim()) callback()
  else callback(new Error('请输入目标 Host'))
}
const validateTarget = (_rule, value, callback) => {
  if (form.type !== 'http') return callback()
  if (/^https?:\/\/[^\s]+$/i.test(String(value || '').trim())) callback()
  else callback(new Error('请输入以 http:// 或 https:// 开头的完整 URL'))
}
const rules = {
  expiry_alert_days: [{ validator: (_rule, value, callback) => form.type !== 'certificate_expiry' || (Number.isInteger(value) && value >= 1 && value <= 365) ? callback() : callback(new Error('提前告警天数应为 1-365 天')), trigger: 'change' }],
  target: [{ validator: validateTarget, trigger: 'blur' }],
  host: [{ validator: validateHost, trigger: 'blur' }],
  username: [{ validator: (_rule, value, callback) => form.type !== 'http' || form.auth_type !== 'basic' || String(value || '').trim() ? callback() : callback(new Error('请输入认证用户名')), trigger: 'blur' }],
}

onMounted(async () => {
  const [sshResult, agentResult] = await Promise.allSettled([
    getCertificateExecutors(),
    getAgents(),
  ])
  if (sshResult.status === 'fulfilled') sshNodes.value = sshResult.value.data || []
  if (agentResult.status === 'fulfilled') agentNodes.value = agentResult.value.data || []
})

function changeProbeType(type) {
  if (form.type === type) return
  const selected = probeTypes.find(item => item.value === type)
  form.type = type
  form.port = selected?.defaultPort ?? 80
  result.value = null
  formRef.value?.clearValidate?.()
}

function resetForm() {
  Object.assign(form, defaultForm(form.type))
  result.value = null
  activePanels.value = []
  formRef.value?.clearValidate?.()
}


async function handleProbe() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  probing.value = true
  result.value = null
  try {
    const response = await instantProbe(buildProbePayload())
    const data = response.data
    let detail = {}; try { detail = JSON.parse(data.result || '{}') } catch {}
    result.value = { ...data, node_name: data.node_name || detail.node_name || (form.node_value === 'direct' ? '服务端直连' : [...sshNodes.value.map(node => ({ ...node, key: `ssh:${node.id}` })), ...agentNodes.value.map(node => ({ ...node, key: `agent:${node.id}` }))].find(node => node.key === form.node_value)?.name || '所选节点') }
    if (response.data.status === 'success') ElMessage.success('拨测成功')
    else ElMessage.warning(response.data.message || response.data.error || '拨测失败')
  } finally {
    probing.value = false
  }
}

function buildProbePayload() {
  const [nodeType, rawNodeID] = form.node_value.includes(':') ? form.node_value.split(':') : ['direct', '0']
  const payload = {
    type: form.type,
    timeout_ms: form.timeout_ms,
    count: form.count,
    server_name: form.server_name.trim(),
    node_type: nodeType,
    node_id: Number(rawNodeID) || 0,
  }
  if (form.type === 'http') Object.assign(payload, {
    target: form.target.trim(),
    method: form.method,
    headers: rowsToObject(form.headers),
    params: rowsToObject(form.params),
    body: form.body,
    content_type: form.content_type,
    auth_type: form.auth_type,
    username: form.auth_type === 'basic' ? form.username : '',
    password: form.auth_type === 'basic' ? form.password : '',
    success_codes: form.success_codes,
    response_keyword: form.response_keyword.trim(),
  })
  else Object.assign(payload, { host: form.host.trim(), port: form.port, ...(form.type === 'certificate_expiry' ? { expiry_alert_days: form.expiry_alert_days } : {}) })
  return payload
}

function rowsToObject(rows) {
  return rows.reduce((resultMap, row) => {
    const key = String(row.key || '').trim()
    if (key) resultMap[key] = row.value || ''
    return resultMap
  }, {})
}

function formatResult(value) {
  if (value.error) return value.error
  return value.body || value.output || value.result || '(无附加输出)'
}

function formatDate(value) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString('zh-CN', { hour12: false })
}
</script>

<style scoped>
.heading-with-help{display:inline-flex;align-items:center;gap:6px}.probe-type-section :deep(.el-form-item__label){display:flex;align-items:center}

.page-container {
  --probe-ink: #17233b;
  --probe-muted: #718096;
  --probe-border: #e2e8f2;
  --probe-soft: #f7f9fc;
  --probe-blue: #1769e0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}



.probe-card,
.result-card {
  border-color: var(--probe-border);
  border-radius: 10px;
  box-shadow: 0 8px 28px rgba(31, 50, 81, .045);
}

.probe-card { container: probe-card / inline-size; }
.probe-card :deep(.el-card__body) { padding: 0; }
.result-card { container: result-card / inline-size; overflow: hidden; }
.result-card :deep(.el-card__body) { padding: 0; }
.probe-form { max-width: 1040px; margin: 0 auto; }

.probe-type-section { margin: 20px 28px 0; }


.form-section { padding: 26px 28px 8px; }
.form-section + .form-section { border-top: 1px solid #edf1f6; }

.section-heading {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 20px;
}

.section-heading h2 {
  margin: 0;
  color: var(--probe-ink);
  font-size: 15px;
  font-weight: 720;
  line-height: 1.3;
}



.form-grid { display: grid; gap: 0 24px; }
.two-columns { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.full-row { grid-column: 1 / -1; }
.http-method-item :deep(.el-radio-group) {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  width: 100%;
}
.http-method-item :deep(.el-radio-button) { width: 100%; }
.http-method-item :deep(.el-radio-button__inner) {
  width: 100%;
  padding-right: 16px;
  padding-left: 16px;
}
.auth-type-item :deep(.el-radio-group) {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  width: 100%;
}
.auth-type-item :deep(.el-radio-button),
.auth-type-item :deep(.el-radio-button__inner) { width: 100%; }
.url-input :deep(.el-textarea__inner) {
  overflow-wrap: anywhere;
  line-height: 1.65;
}
:deep(.el-form-item) { margin-bottom: 20px; }
:deep(.el-form-item__label) { padding-bottom: 7px; color: #46536a; font-size: 13px; font-weight: 650; }
:deep(.el-select),
:deep(.el-input-number) { width: 100%; }
:deep(.el-input__wrapper),
:deep(.el-select__wrapper),
:deep(.el-input-number) { min-height: 38px; transition: box-shadow .18s ease, border-color .18s ease; }
:deep(.el-input__wrapper.is-focus),
:deep(.el-select__wrapper.is-focused) { box-shadow: 0 0 0 1px var(--probe-blue) inset, 0 0 0 3px rgba(23, 105, 224, .08); }

.advanced-collapse {
  margin: 14px 28px 0;
  overflow: hidden;
  background: var(--probe-soft);
  border: 1px solid #dfe7f1;
  border-radius: 9px;
}

:deep(.advanced-collapse .el-collapse-item__header) { height: 54px; padding: 0 16px; border-bottom: 0; background: var(--probe-soft); }
:deep(.advanced-collapse .el-collapse-item__header:hover) { background: #f1f6fc; }
:deep(.advanced-collapse .el-collapse-item__wrap) { border-bottom: 0; }
:deep(.advanced-collapse .el-collapse-item__content) { padding: 20px 20px 0; }
.advanced-title { color: #315781; font-weight: 700; }
.execution-section {
  margin-top: 18px;
  padding-top: 22px;
  padding-bottom: 8px;
  border-top: 1px solid #edf1f6;
}
.execution-section .section-heading { margin-bottom: 16px; }
.execution-panel { padding: 0; }

.execution-node-field {
  width: 100%;
  margin-bottom: 0;
}

.form-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding: 18px 28px 22px;
  background: #fbfcfe;
  border-top: 1px solid var(--probe-border);
}

.form-actions :deep(.el-button) { min-width: 98px; height: 38px; }
.form-actions :deep(.el-button + .el-button) { margin-left: 0; }
.form-actions :deep(.el-button--primary) { box-shadow: 0 6px 14px rgba(23, 105, 224, .18); }
.mono-input :deep(textarea),
.result-box { font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; }
.result-header {
  display: flex;
  align-items: center;
  gap: 13px;
  padding: 19px 24px;
  border-bottom: 1px solid;
}
.result-header.is-success { background: #f4fbf8; border-bottom-color: #dcefe7; }
.result-header.is-failed { background: #fff7f6; border-bottom-color: #f4dfdc; }
.result-status-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  flex: 0 0 36px;
  color: #fff;
  border-radius: 50%;
  font-size: 18px;
  font-weight: 800;
}
.is-success .result-status-mark { background: #18a779; box-shadow: 0 0 0 5px rgba(24, 167, 121, .09); }
.is-failed .result-status-mark { background: #df5b50; box-shadow: 0 0 0 5px rgba(223, 91, 80, .09); }
.result-heading { min-width: 0; }
.result-heading h2 { margin: 0; color: var(--probe-ink); font-size: 15px; font-weight: 750; line-height: 1.35; }
.result-heading p { margin: 3px 0 0; color: #748296; font-size: 12px; line-height: 1.45; }
.result-metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
  padding: 18px 24px;
}
.result-metric {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 6px;
  padding: 13px 14px;
  background: #f8fafc;
  border: 1px solid #e5ebf2;
  border-radius: 8px;
}
.result-metric span { color: #8793a5; font-size: 11px; }
.result-metric strong { overflow: hidden; color: #26364d; font-size: 14px; font-weight: 720; text-overflow: ellipsis; white-space: nowrap; }
.result-metric small { color: #7f8ca0; font-size: 10px; font-weight: 600; }
.result-details {
  margin: 0 24px 18px;
  overflow: hidden;
  border: 1px solid #e3e9f1;
  border-radius: 8px;
}
.result-detail-row {
  display: grid;
  grid-template-columns: 110px minmax(0, 1fr);
  gap: 14px;
  align-items: start;
  padding: 11px 14px;
  border-bottom: 1px solid #edf1f5;
}
.result-detail-row:last-child { border-bottom: 0; }
.result-detail-row > span { color: #7f8ca0; font-size: 12px; }
.result-detail-row strong,
.result-detail-row code { min-width: 0; color: #34445a; font-size: 12px; font-style: normal; line-height: 1.55; overflow-wrap: anywhere; }
.result-detail-row code { font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; }
.result-detail-row.split-value { grid-template-columns: 110px minmax(0, 1fr) auto; }
.result-detail-row em { padding: 2px 8px; color: #8a5c14; background: #fff7e6; border-radius: 999px; font-size: 11px; font-style: normal; white-space: nowrap; }
.result-output { padding: 0 24px 24px; }
.result-output-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.result-output-heading strong { color: var(--probe-ink); font-size: 13px; font-weight: 700; }
.result-output-heading span { color: #8a96a8; font-size: 11px; }
.result-box { margin: 10px 0 0; padding: 16px; max-height: 360px; overflow: auto; color: #d7e0ec; background: #111b2e; border: 1px solid #24324a; border-radius: 8px; font-size: 12px; line-height: 1.65; white-space: pre-wrap; overflow-wrap: anywhere; }
:deep(.kv-editor) { display: flex; flex-direction: column; gap: 9px; width: 100%; }
:deep(.kv-row) { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 36px; gap: 10px; align-items: center; }
:deep(.kv-row .el-button) { width: 34px; height: 34px; color: #54708f; background: #f7f9fc; border-color: #dce4ef; }

@media (prefers-reduced-motion: reduce) {
  :deep(.el-input__wrapper),
  :deep(.el-select__wrapper),
  :deep(.el-input-number) { transition: none; }
}

@container probe-card (max-width: 760px) {
  .execution-panel { grid-template-columns: 1fr; gap: 14px; }
}

@container result-card (max-width: 680px) {
  .result-metrics { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .result-detail-row,
  .result-detail-row.split-value { grid-template-columns: 1fr; gap: 5px; }
  .result-detail-row em { justify-self: start; }
}

@media (max-width: 760px) {
  .probe-type-section { grid-template-columns: 1fr; gap: 11px; margin: 16px 18px 0; padding: 15px; }
  .form-section { padding: 22px 18px 6px; }
  .two-columns { grid-template-columns: 1fr; }
  .advanced-collapse { margin-right: 18px; margin-left: 18px; }
  .advanced-hint { display: none; }
  .execution-panel { padding: 15px; }
  .form-actions { padding: 16px 18px 20px; }
  :deep(.kv-row) { grid-template-columns: 1fr 1fr 34px; }
  .result-header { padding: 17px 18px; }
  .result-metrics { padding: 15px 18px; }
  .result-details { margin-right: 18px; margin-left: 18px; }
  .result-output { padding-right: 18px; padding-bottom: 20px; padding-left: 18px; }
}
</style>
