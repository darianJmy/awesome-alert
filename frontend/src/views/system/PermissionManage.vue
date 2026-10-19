<template>
  <div class="page-container">
    <el-card>
      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="权限名称" min-width="240" />
        <el-table-column prop="code" label="权限编码" min-width="360" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <template #empty><TableEmpty /></template>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getPermissions } from '../../api/permission'

const loading = ref(false)
const tableData = ref([])

onMounted(async () => {
  loading.value = true
  try {
    const res = await getPermissions()
    tableData.value = res.data
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.page-container { display: flex; flex-direction: column; gap: 16px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: bold; }
</style>
