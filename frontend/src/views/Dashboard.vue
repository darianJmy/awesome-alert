<template>
  <div class="dashboard">
    <header class="page-header">
      <div class="page-heading">
        <h1>运维中枢</h1>
        <p>统一管理数据源、告警治理、拨测监控和自动化执行。</p>
      </div>
      <div class="hero-status page-header-meta">
        <span class="pulse"></span>
        平台运行中
      </div>
    </header>

    <el-row :gutter="14">
      <el-col :xs="24" :sm="12" :lg="6" v-for="card in cards" :key="card.title">
        <el-card class="metric-card" shadow="never">
          <div class="card-inner">
            <div class="card-icon" :class="card.tone">
              <el-icon :size="22"><component :is="card.icon" /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-title">{{ card.title }}</div>
              <div class="card-value">{{ card.value }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const cards = ref([
  { title: '用户数', value: '1', icon: 'User', tone: 'blue' },
  { title: '角色数', value: '1', icon: 'Avatar', tone: 'green' },
  { title: '权限数', value: '26', icon: 'Lock', tone: 'amber' },
  { title: '菜单数', value: '8', icon: 'Menu', tone: 'red' },
])
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.dashboard-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 24px;
  background: #fff;
  border: 1px solid var(--ops-border-soft);
  border-radius: 8px;
}

.dashboard-hero h1 {
  margin: 0;
  color: var(--ops-text);
  font-size: 26px;
  line-height: 1.25;
  font-weight: 800;
}

.dashboard-hero p {
  margin: 8px 0 0;
  color: var(--ops-muted);
  font-size: 14px;
}

.hero-status {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 34px;
  padding: 0 12px;
  color: var(--ops-success);
  background: #ecfdf3;
  border: 1px solid #abefc6;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 700;
  white-space: nowrap;
}

.pulse {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: var(--ops-success);
}

.metric-card {
  margin-bottom: 14px;
}

.card-inner {
  display: flex;
  align-items: center;
  gap: 14px;
}

.card-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 46px;
  height: 46px;
  border-radius: 8px;
}

.card-icon.blue {
  color: var(--ops-primary);
  background: var(--ops-primary-soft);
}

.card-icon.green {
  color: var(--ops-success);
  background: #ecfdf3;
}

.card-icon.amber {
  color: var(--ops-warning);
  background: #fffaeb;
}

.card-icon.red {
  color: var(--ops-danger);
  background: #fef3f2;
}

.card-info {
  min-width: 0;
}

.card-title {
  margin-bottom: 5px;
  color: var(--ops-muted);
  font-size: 13px;
  font-weight: 650;
}

.card-value {
  color: var(--ops-text);
  font-size: 26px;
  line-height: 1;
  font-weight: 800;
}

@media (max-width: 700px) {
  .dashboard-hero {
    align-items: flex-start;
    flex-direction: column;
    padding: 18px;
  }
}
</style>
