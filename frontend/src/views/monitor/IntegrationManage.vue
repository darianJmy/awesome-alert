<template>
  <div class="integration-page">
    <el-tabs v-model="activeTab" class="integration-tabs" @tab-change="changeTab">
      <el-tab-pane label="数据源" name="sources">
        <DatasourceManage />
      </el-tab-pane>
      <el-tab-pane label="数据源分组" name="workspaces" lazy>
        <DatasourceGroupManage />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import DatasourceManage from './DatasourceManage.vue'
import DatasourceGroupManage from './DatasourceGroupManage.vue'

const route = useRoute()
const router = useRouter()
const tabFromRoute = () => route.query.tab === 'workspaces' ? 'workspaces' : 'sources'
const activeTab = ref(tabFromRoute())

watch(() => route.query.tab, () => { activeTab.value = tabFromRoute() })

function changeTab(tab) {
  const query = { ...route.query }
  if (tab === 'workspaces') query.tab = 'workspaces'
  else delete query.tab
  router.replace({ path: '/integrations/sources', query })
}
</script>

<style scoped>
.integration-page { --integration-border: #e0e7ef; overflow: hidden; border: 1px solid var(--integration-border); border-radius: 7px; background: #fff; }
.integration-tabs :deep(.el-tabs__header) { margin: 0; padding: 0 22px; border-bottom: 1px solid #edf0f4; }
.integration-tabs :deep(.el-tabs__nav-wrap::after) { display: none; }
.integration-tabs :deep(.el-tabs__item) { height: 52px; padding: 0 21px; color: #66758a; font-size: 13px; }
.integration-tabs :deep(.el-tabs__item.is-active) { color: #1769d2; font-weight: 600; }
.integration-tabs :deep(.el-tabs__active-bar) { height: 2px; background: #1769d2; }
.integration-tabs :deep(.el-tabs__content) { overflow: visible; padding: 20px 22px 24px; }
.integration-tabs :deep(.datasource-table-card) { border: 0; border-radius: 0; }
.integration-tabs :deep(.datasource-table-card .el-card__body) { padding: 0; }
.integration-tabs :deep(.group-toolbar) { padding: 0; border: 0; border-radius: 0; }
.integration-tabs :deep(.group-card) { padding: 0; border: 0; border-radius: 0; }
@media (max-width: 600px) {
  .integration-tabs :deep(.el-tabs__header) { overflow-x: auto; }
  .integration-tabs :deep(.el-tabs__content) { padding: 16px; }
}
</style>
