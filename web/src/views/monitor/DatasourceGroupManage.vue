<template>
  <div class="group-page">
    <div class="group-toolbar">
      <div class="search-area" role="search" aria-label="数据源分组筛选">
        <el-input v-model="keyword" :prefix-icon="Search" clearable placeholder="搜索数据源分组名称" @keyup.enter="handleSearch" @clear="handleSearch" />
        <el-button :icon="Search" @click="handleSearch">查询</el-button>
      </div>
      <div class="toolbar-actions">
        <el-button :icon="Refresh" :loading="loading" @click="loadGroups">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openDrawer()">创建</el-button>
      </div>
    </div>

    <section class="group-card">
      <el-table v-loading="loading" :data="pagedGroups" row-key="id" border>
        <el-table-column prop="name" label="分组名称" min-width="180" />
        <el-table-column label="数据源" min-width="320">
          <template #default="{ row }">
            <div class="source-list">
              <el-tag v-for="name in row.datasource_names" :key="name" size="small" effect="plain">{{ name }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="描述" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">{{ row.description || '' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" :icon="Edit" aria-label="编辑数据源分组" title="编辑" @click="openDrawer(row)" />
            <el-button link type="danger" :icon="Delete" aria-label="删除数据源分组" title="删除" @click="removeGroup(row)" />
          </template>
        </el-table-column>
        <template #empty><TableEmpty /></template>
      </el-table>
      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :page-sizes="[10, 20, 50]" :total="filteredGroups.length" layout="total, sizes, prev, pager, next" />
      </div>
    </section>

    <el-drawer v-model="drawerVisible" class="datasource-group-drawer" size="min(460px, 94vw)" :close-on-click-modal="false" destroy-on-close>
      <template #header>
        <div class="drawer-header"><DrawerBackButton @click="drawerVisible = false" /><strong>{{ form.id ? '编辑数据源分组' : '新增数据源分组' }}</strong></div>
      </template>
      <el-form ref="formRef" :model="form" :rules="formRules" label-position="top">
        <el-form-item label="分组名称" prop="name"><el-input v-model="form.name" maxlength="128" show-word-limit /></el-form-item>
        <el-form-item label="数据源" prop="datasource_ids">
          <el-select v-model="form.datasource_ids" multiple filterable collapse-tags :max-collapse-tags="5" style="width: 100%">
            <el-option v-for="item in datasources" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="5" maxlength="500" show-word-limit /></el-form-item>
      </el-form>
      <template #footer><el-button @click="drawerVisible = false">取消</el-button><el-button type="primary" :loading="saving" @click="saveGroup">保存</el-button></template>
    </el-drawer>
  </div>
</template>

<script setup>
import DrawerBackButton from '../../components/DrawerBackButton.vue'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Delete, Edit, Plus, Refresh, Search } from '@element-plus/icons-vue'
import { createDatasourceGroup, deleteDatasourceGroup, getDatasourceGroups, updateDatasourceGroup } from '../../api/datasourceGroup'
import { getPromDatasources } from '../../api/prometheus'

const groups = ref([]), datasources = ref([]), keyword = ref(''), appliedKeyword = ref('')
const loading = ref(false), saving = ref(false), drawerVisible = ref(false), formRef = ref()
const page = ref(1), pageSize = ref(10)
const emptyForm = () => ({ id: null, name: '', description: '', datasource_ids: [] })
const form = reactive(emptyForm())
const formRules = {
  name: [{ required: true, message: '请输入数据源分组名称', trigger: 'blur' }],
  datasource_ids: [{ type: 'array', required: true, min: 1, message: '请至少选择一个数据源', trigger: 'change' }],
}
const filteredGroups = computed(() => {
  const value = appliedKeyword.value.toLowerCase()
  if (!value) return groups.value
  return groups.value.filter(item => `${item.name} ${item.description || ''}`.toLowerCase().includes(value))
})
const pagedGroups = computed(() => filteredGroups.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value))
watch(pageSize, () => { page.value = 1 })

onMounted(() => Promise.all([loadGroups(), loadDatasources()]))

async function loadGroups() {
  loading.value = true
  try { groups.value = (await getDatasourceGroups()).data || [] }
  finally { loading.value = false }
}
async function loadDatasources() { datasources.value = (await getPromDatasources()).data || [] }
function handleSearch() { appliedKeyword.value = keyword.value.trim(); page.value = 1 }
function openDrawer(row) {
  Object.assign(form, row ? { id: row.id, name: row.name, description: row.description || '', datasource_ids: [...(row.datasource_ids || [])] } : emptyForm())
  drawerVisible.value = true
}
async function saveGroup() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    const payload = { name: form.name.trim(), description: form.description.trim(), datasource_ids: form.datasource_ids }
    if (form.id) await updateDatasourceGroup(form.id, payload)
    else await createDatasourceGroup(payload)
    ElMessage.success(form.id ? '数据源分组已更新' : '数据源分组已创建')
    drawerVisible.value = false
    await loadGroups()
  } finally { saving.value = false }
}
async function removeGroup(row) {
  if (row.rule_count > 0) { ElMessage.warning(`该数据源分组仍关联 ${row.rule_count} 条告警规则，请先解除关联。`); return }
  await ElMessageBox.confirm(`确认删除数据源分组“${row.name}”？`, '删除数据源分组', { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消' })
  await deleteDatasourceGroup(row.id)
  ElMessage.success('数据源分组已删除')
  await loadGroups()
}
</script>

<style scoped>
.group-page { display: flex; flex-direction: column; gap: 16px; }.group-toolbar, .search-area, .toolbar-actions, .source-list, .drawer-header { display: flex; align-items: center; gap: 8px; }.group-toolbar { justify-content: space-between; padding: 16px; border: 1px solid #e5e9f0; border-radius: 8px; background: #fff; }.toolbar-actions :deep(.el-button + .el-button) { margin-left: 0; }.search-area .el-input { width: 280px; }.group-card { padding: 16px; border: 1px solid #e5e9f0; border-radius: 8px; background: #fff; }.source-list { flex-wrap: wrap; }.source-list :deep(.el-tag) { border-radius: 0; }.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; }.drawer-header { gap: 12px; }.drawer-header .el-button { margin-left: -10px; }
:global(.datasource-group-drawer.el-drawer) { border-radius: 0 !important; }:global(.datasource-group-drawer .el-drawer__header) { margin-bottom: 0; padding: 16px 20px; border-bottom: 1px solid #e5e9f0; }:global(.datasource-group-drawer .el-drawer__body) { padding: 22px 24px; }:global(.datasource-group-drawer .el-drawer__footer) { padding: 12px 20px; border-top: 1px solid #e5e9f0; }
@media (max-width: 720px) { .group-toolbar { align-items: stretch; flex-direction: column; }.search-area { flex-wrap: wrap; }.search-area .el-input { width: min(100%, 280px); }.toolbar-actions { justify-content: flex-end; }.pagination-wrap { overflow-x: auto; justify-content: flex-start; } }
</style>
