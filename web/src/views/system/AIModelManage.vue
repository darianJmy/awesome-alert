<template>
  <section class="model-settings">
    <header class="model-heading"><div><h2>AI 模型</h2><p>添加模型连接，聊天时可自由选择。保存后立即生效。</p></div><el-button v-permission="['system-setting:edit']" type="primary" @click="edit()">添加模型</el-button></header>
    <el-table v-loading="loading" :data="models" empty-text="暂无模型，请先添加模型连接">
      <el-table-column prop="name" label="显示名称" min-width="150" />
      <el-table-column prop="provider" label="Provider" width="130" />
      <el-table-column prop="model" label="模型标识" min-width="180" />
      <el-table-column prop="base_url" label="服务地址" min-width="240" show-overflow-tooltip />
      <el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag></template></el-table-column>
      <el-table-column label="操作" width="130"><template #default="{ row }"><el-button v-permission="['system-setting:edit']" link type="primary" @click="edit(row)">编辑</el-button><el-button v-permission="['system-setting:edit']" link type="danger" @click="remove(row)">删除</el-button></template></el-table-column>
    </el-table>
    <el-dialog v-model="visible" :title="form.id ? '编辑模型' : '添加模型'" width="560px" :close-on-click-modal="false">
      <el-form label-position="top" @submit.prevent="save">
        <el-form-item label="显示名称" required><el-input v-model="form.name" maxlength="128" placeholder="例如：日常运维助手" /></el-form-item>
        <el-form-item label="Provider" required><el-select v-model="form.provider"><el-option label="NewAPI" value="newapi" /><el-option label="OpenAI 兼容" value="openai" /><el-option label="Anthropic" value="anthropic" /></el-select></el-form-item>
        <el-form-item label="服务地址" required><el-input v-model="form.base_url" placeholder="https://your-model-service.example.com/v1" /></el-form-item>
        <el-form-item label="模型标识" required><el-input v-model="form.model" maxlength="200" placeholder="填写服务端实际启用的模型名称" /></el-form-item>
        <el-form-item label="API Key" :required="!form.id"><el-input v-model="form.api_key" type="password" show-password autocomplete="new-password" :placeholder="form.id ? '留空保留现有密钥，更换地址或 Provider 时需重填' : '输入 API Key'" /></el-form-item>
        <el-form-item label="启用"><el-switch v-model="form.enabled" /></el-form-item>
        <p class="model-note">请选择支持工具调用的模型。密钥加密保存，保存后不会回显。</p>
      </el-form>
      <template #footer><el-button :disabled="saving" @click="visible = false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </section>
</template>
<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAIModels, saveAIModel, deleteAIModel } from '../../api/aiModel'
const models = ref([]), loading = ref(false), saving = ref(false), visible = ref(false)
const blank = () => ({ id: null, name: '', provider: 'newapi', base_url: '', model: '', api_key: '', enabled: true })
const form = reactive(blank())
async function load() { loading.value = true; try { models.value = (await getAIModels()).data || [] } finally { loading.value = false } }
function edit(row) { Object.assign(form, blank(), row ? { ...row, api_key: '', enabled: row.status === 1 } : {}); visible.value = true }
async function save() {
  if (!form.name.trim() || !form.base_url.trim() || !form.model.trim() || (!form.id && !form.api_key.trim())) { ElMessage.warning('请填写完整的模型配置'); return }
  saving.value = true
  try { await saveAIModel(form.id, { name: form.name, provider: form.provider, base_url: form.base_url, model: form.model, api_key: form.api_key, status: form.enabled ? 1 : 0 }); form.api_key = ''; visible.value = false; ElMessage.success('模型已保存'); await load() } finally { saving.value = false }
}
async function remove(row) {
  try { await ElMessageBox.confirm(`删除“${row.name}”后，使用它的会话需要重新选择模型。`, '删除模型', { type: 'warning' }); await deleteAIModel(row.id); await load() } catch (error) { if (error !== 'cancel' && error !== 'close') ElMessage.error('删除失败') }
}
onMounted(load)
</script>
<style scoped>
.model-settings{max-width:1200px;margin:auto}.model-heading{display:flex;align-items:center;justify-content:space-between;margin-bottom:20px;gap:16px}.model-heading h2{margin:0;font-size:21px}.model-heading p,.model-note{color:var(--ops-muted);font-size:12px;line-height:1.6}.model-settings :deep(.el-select){width:100%}
</style>
