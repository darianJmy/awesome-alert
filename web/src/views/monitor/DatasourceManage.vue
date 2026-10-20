<template>
  <div class="page-container">
    <div class="datasource-toolbar">
      <div class="search-bar datasource-filters" role="search" aria-label="接入筛选">
        <el-input
          v-model="keyword"
          class="filter-keyword"
          :prefix-icon="Search"
          clearable
          aria-label="搜索接入"
          placeholder="搜索名称、地址或标签"
          @input="handleKeywordInput"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <el-select v-model="typeFilter" class="filter-select filter-type" aria-label="按类型筛选" placeholder="全部类型" clearable filterable @change="handleSearch">
          <el-option v-for="t in typeOptions" :key="t.value" :label="t.label" :value="t.value" />
        </el-select>
        <el-select v-model="statusFilter" class="filter-select filter-status" aria-label="按状态筛选" placeholder="全部状态" clearable @change="handleSearch">
          <el-option label="启用" value="1" />
          <el-option label="禁用" value="0" />
        </el-select>
        <el-button @click="handleReset">清空</el-button>
      </div>

      <div class="datasource-actions">
        <el-dropdown trigger="click" :disabled="!selectedRows.length || !!batchAction" @command="handleBatchCommand">
          <el-button class="batch-operation-button" :disabled="!selectedRows.length || !!batchAction" :loading="!!batchAction">
            批量操作
            <el-icon class="batch-operation-arrow"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="enable" v-permission="['datasource:edit']">批量启用</el-dropdown-item>
              <el-dropdown-item command="disable" v-permission="['datasource:edit']">批量禁用</el-dropdown-item>
              <el-dropdown-item command="delete" divided v-permission="['datasource:delete']">批量删除</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button :icon="Refresh" :loading="loading" @click="fetchData">刷新</el-button>
        <el-button class="add-source-button" type="primary" :icon="Plus" @click="handleAdd" v-permission="['datasource:create']">新增</el-button>
      </div>
    </div>

    <el-card class="datasource-table-card" shadow="never">
      <el-table :data="tableData" v-loading="loading" row-key="id" border @selection-change="handleSelectionChange" @sort-change="handleSortChange">
        <el-table-column type="selection" width="48" />
        <el-table-column prop="name" label="接入名称" width="170" sortable="custom" />
        <el-table-column prop="category" label="类别" width="120" sortable="custom">
          <template #default="{ row }">
            <span class="category-text">{{ categoryLabel(row.category || categoryByType(row.type)) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="140" align="center" header-align="center" sortable="custom">
          <template #default="{ row }">
            <span v-if="row.type === 'prometheus'" class="type-logo-cell" title="Prometheus">
              <img :src="prometheusLogo" alt="Prometheus" class="type-logo" />
            </span>
            <span v-else-if="row.type === 'sls'" class="type-logo-cell type-logo-cell-wide" title="阿里云 SLS">
              <img :src="aliyunSlsLogo" alt="阿里云 SLS" class="type-logo type-logo-wide" />
            </span>
            <span v-else-if="row.type === 'cls'" class="type-logo-cell type-logo-cell-wide" title="腾讯云 CLS">
              <img :src="tencentClsLogo" alt="腾讯云 CLS" class="type-logo type-logo-wide" />
            </span>
            <span v-else-if="row.type === 'k8s'" class="type-logo-cell" title="Kubernetes">
              <img :src="kubernetesLogo" alt="Kubernetes" class="type-logo" />
            </span>
            <span v-else-if="row.type === 'nacos'" class="type-logo-cell type-logo-cell-wide" title="Nacos">
              <img :src="nacosLogo" alt="Nacos" class="type-logo type-logo-wide" />
            </span>
            <el-tag v-else :type="typeColor(row.type)">{{ typeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="标签" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="scope-tags">
              <span
                v-for="item in datasourceLabelEntries(row)"
                :key="item.key"
                class="scope-tag"
                :class="labelToneClass(item.key)"
              >{{ item.key }}:{{ item.value }}</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="150" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100" sortable="custom">
          <template #default="{ row }">
            <div class="status-container">
              <div class="status-dot" :class="row.status === 1 ? 'status-enabled' : 'status-disabled'"></div>
              <span>{{ row.status === 1 ? '启用' : '禁用' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="110" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" :icon="Edit" title="编辑" aria-label="编辑" @click="handleEdit(row)" v-permission="['datasource:edit']" />
            <el-button link type="danger" :icon="Delete" title="删除" aria-label="删除" @click="handleDelete(row)" v-permission="['datasource:delete']" />
          </template>
        </el-table-column>
        <template #empty><TableEmpty /></template>
      </el-table>

      <div class="pagination">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize"
          :total="total" layout="total, prev, pager, next, sizes" :page-sizes="[10,20,50]"
          @current-change="fetchData" @size-change="handleSizeChange" />
      </div>
    </el-card>

    <!-- 右侧抽屉 -->
    <el-drawer v-model="drawerVisible" class="datasource-drawer" direction="rtl" size="600px" :close-on-click-modal="false"
      @close="handleDrawerClose">
      <template #header="{ titleId, titleClass }">
        <div class="drawer-header">
          <el-button text :icon="ArrowLeft" class="back-button" @click="handleDrawerBack">返回</el-button>
          <span :id="titleId" :class="[titleClass, 'drawer-title']">
            {{ step === 1 ? '选择数据源类型' : (isEdit ? '编辑接入' : '新增接入') }}
          </span>
        </div>
      </template>

      <!-- 步骤 1: 选择接入类型 -->
      <div v-if="step === 1" class="type-groups">
        <section v-for="group in groupedTypeOptions" :key="group.value" class="type-group">
          <div class="type-group-title">{{ group.label }}</div>
          <div class="type-grid">
            <div v-for="t in group.items" :key="t.value" class="type-card" :class="{ active: hoverType === t.value }"
              @click="selectType(t.value)" @mouseenter="hoverType = t.value" @mouseleave="hoverType = null">
              <img v-if="t.value === 'prometheus'" :src="prometheusLogo" alt="Prometheus" class="type-card-logo" />
              <img v-else-if="t.value === 'sls'" :src="aliyunSlsLogo" alt="阿里云 SLS" class="type-card-logo type-card-logo-wide" />
              <img v-else-if="t.value === 'cls'" :src="tencentClsLogo" alt="腾讯云 CLS" class="type-card-logo type-card-logo-wide" />
              <img v-else-if="t.value === 'k8s'" :src="kubernetesLogo" alt="Kubernetes" class="type-card-logo" />
              <img v-else-if="t.value === 'nacos'" :src="nacosLogo" alt="Nacos" class="type-card-logo type-card-logo-wide" />
              <el-icon v-else :size="32" :color="t.color">
                <component :is="t.icon" />
              </el-icon>
              <div class="type-card-name">{{ t.label }}</div>
              <div class="type-card-desc">{{ t.desc }}</div>
            </div>
          </div>
        </section>
        <div class="type-note">接入中心负责统一保存连接、认证和类型配置；查询、巡检等功能会使用这些接入。</div>
      </div>

      <!-- 步骤 2: 配置表单 -->
      <div v-if="step === 2" class="step2-wrapper">
        <!-- 已选类型卡片 -->
        <div class="selected-type-banner">
          <div class="selected-type-icon" :class="{ 'selected-type-icon-wide': ['sls', 'cls', 'nacos'].includes(form.type) }" :style="{ background: selectedTypeInfo.bg }">
            <img v-if="form.type === 'prometheus'" :src="prometheusLogo" alt="Prometheus" class="selected-type-logo" />
            <img v-else-if="form.type === 'sls'" :src="aliyunSlsLogo" alt="阿里云 SLS" class="selected-type-logo selected-type-logo-wide" />
            <img v-else-if="form.type === 'cls'" :src="tencentClsLogo" alt="腾讯云 CLS" class="selected-type-logo selected-type-logo-wide" />
            <img v-else-if="form.type === 'k8s'" :src="kubernetesLogo" alt="Kubernetes" class="selected-type-logo" />
            <img v-else-if="form.type === 'nacos'" :src="nacosLogo" alt="Nacos" class="selected-type-logo selected-type-logo-wide" />
            <el-icon v-else :size="22" :color="selectedTypeInfo.color">
              <component :is="selectedTypeInfo.icon" />
            </el-icon>
          </div>
          <div class="selected-type-text">
            <div class="selected-type-label">{{ selectedTypeInfo.label }}</div>
            <div class="selected-type-desc">{{ categoryLabel(form.category) }} / {{ selectedTypeInfo.desc }}</div>
          </div>
          <el-tag :type="typeColor(form.type)" size="small" effect="plain">已选择</el-tag>
        </div>

        <el-form ref="formRef" :model="form" :rules="rules" label-width="0" class="config-form">
          <!-- ── 基本配置 ── -->
          <div class="section-card">
            <div class="section-body">
              <div class="form-label">接入名称 <span class="required">*</span></div>
              <el-form-item prop="name">
                <el-input v-model="form.name" size="large" />
              </el-form-item>

              <template v-if="form.type === 'sls'">
                <div class="form-label">Region <span class="required">*</span></div>
                <el-form-item>
                  <el-select
                    v-model="configFields.region"
                    filterable
                    allow-create
                    default-first-option
                    placeholder="选择或输入阿里云 Region，如 cn-hangzhou"
                    size="large"
                    style="width: 100%"
                    @change="updateSlsDerivedFields"
                  >
                    <el-option v-for="r in aliyunRegionOptions" :key="r" :label="r" :value="r" />
                  </el-select>
                </el-form-item>

                <div class="form-label">Project 名称 <span class="required">*</span></div>
                <el-form-item>
                  <el-input v-model="configFields.project" placeholder="SLS Project 名称" size="large" @input="updateSlsDerivedFields" />
                </el-form-item>

              </template>

              <template v-if="form.type === 'cls'">
                <div class="form-label">Region <span class="required">*</span></div>
                <el-form-item>
                  <el-select
                    v-model="configFields.region"
                    filterable
                    allow-create
                    default-first-option
                    placeholder="选择或输入腾讯云 Region，如 ap-shanghai / shanghai"
                    size="large"
                    style="width: 100%"
                    @change="updateClsDerivedFields"
                  >
                    <el-option v-for="r in tencentRegionOptions" :key="r" :label="r" :value="r" />
                  </el-select>
                </el-form-item>

                <div class="form-label">日志主题 ID <span class="required">*</span></div>
                <el-form-item>
                  <el-input v-model="configFields.topic_id" placeholder="CLS Topic ID，如 xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" size="large" @input="updateClsDerivedFields" />
                </el-form-item>
              </template>

              <div v-if="shouldShowURLField()" class="form-label">URL <span class="required">*</span></div>
              <el-form-item v-if="shouldShowURLField()" prop="url">
                <el-input v-model="form.url" size="large" />
              </el-form-item>

              <el-collapse v-if="supportsCustomHeaders" v-model="advancedSections" class="advanced-config">
                <el-collapse-item name="http-headers">
                  <template #title>
                    <span class="advanced-title"><el-icon><Setting /></el-icon>高级配置</span>
                  </template>
                  <div class="advanced-content">
                    <div class="advanced-heading">自定义 HTTP 头 <span>（可选）</span></div>
                    <el-alert title="可添加自定义 HTTP 头，用于 API 密钥认证、请求追踪等场景。" type="info" :closable="false" show-icon />
                    <div class="header-editor">
                      <el-button class="add-header-button" :icon="Plus" plain :disabled="configFields.headers.length >= 20" @click="addHeader">
                        添加 HTTP 头
                      </el-button>
                      <div v-for="(item, index) in configFields.headers" :key="index" class="header-row">
                        <el-input v-model="item.key" placeholder="Header 名称" maxlength="128" />
                        <el-input v-model="item.value" placeholder="Header 值" maxlength="1000" />
                        <el-button class="header-remove-button" :icon="Minus" circle plain aria-label="删除 HTTP 头" @click="removeHeader(index)" />
                      </div>
                    </div>
                  </div>
                </el-collapse-item>
              </el-collapse>

              <div class="form-label">标签</div>
              <div class="label-editor">
                <el-button class="add-label-button" :icon="Plus" plain :disabled="form.labels.length >= 20" @click="addLabel">
                  添加标签
                </el-button>
                <div v-for="(item, index) in form.labels" :key="index" class="label-row">
                  <el-input v-model="item.key" placeholder="Key" maxlength="64" />
                  <el-input v-model="item.value" placeholder="Value" maxlength="200" />
                  <el-button
                    class="label-icon-button"
                    :icon="Minus"
                    circle
                    plain
                    aria-label="删除标签"
                    @click="removeLabel(index)"
                  />
                </div>
              </div>
              <div class="field-help">标签用于筛选和智能分析，最多添加 20 个。</div>
            </div>
          </div>

          <!-- ── 认证配置 ── -->
          <div class="section-card">
            <div class="section-body">
              <div class="form-label">认证方式</div>
              <el-form-item>
                <div class="auth-segmented">
                  <button
                    v-for="opt in authTypeOptions"
                    :key="opt.value"
                    type="button"
                    class="auth-seg"
                    :class="{ active: form.auth_type === opt.value }"
                    @click="form.auth_type = opt.value"
                  >
                    <span>{{ opt.label }}</span>
                  </button>
                </div>
              </el-form-item>

              <transition name="el-fade-in-linear">
                <div v-if="form.auth_type === 'basic'" class="auth-fields">
                  <div class="form-label">用户名 <span class="required">*</span></div>
                  <el-form-item>
                    <el-input v-model="authFields.username" placeholder="输入用户名" />
                  </el-form-item>
                  <div class="form-label">密码 <span class="required">*</span></div>
                  <el-form-item>
                    <el-input v-model="authFields.password" type="password" show-password placeholder="输入密码" />
                  </el-form-item>
                </div>
              </transition>

              <transition name="el-fade-in-linear">
                <div v-if="form.auth_type === 'bearer'" class="auth-fields">
                  <div class="form-label">Token</div>
                  <el-form-item>
                    <el-input v-model="authFields.token" type="password" show-password :placeholder="isEdit ? '已保存 Token，重新填写则更新' : '输入 Bearer Token'" />
                  </el-form-item>
                </div>
              </transition>

              <transition name="el-fade-in-linear">
                <div v-if="form.auth_type === 'ak_sk'" class="auth-fields">
                  <div class="form-label">AccessKey ID</div>
                  <el-form-item>
                    <el-input v-model="authFields.access_key" :placeholder="isEdit ? '留空则不修改 AccessKey ID' : 'Access Key ID / SecretId'" />
                  </el-form-item>
                  <div class="form-label">AccessKey Secret</div>
                  <el-form-item>
                    <el-input v-model="authFields.secret_key" type="password" show-password :placeholder="isEdit ? '留空则不修改 AccessKey Secret' : 'Access Key Secret / SecretKey'" />
                  </el-form-item>
                </div>
              </transition>

              <transition name="el-fade-in-linear">
                <div v-if="form.auth_type === 'password'" class="auth-fields">
                  <div class="form-label">密码</div>
                  <el-form-item>
                    <el-input v-model="authFields.password" type="password" show-password placeholder="输入密码" />
                  </el-form-item>
                </div>
              </transition>

              <transition name="el-fade-in-linear">
                <div v-if="form.auth_type === 'kubeconfig'" class="auth-fields">
                  <div class="form-label">Kubeconfig</div>
                  <el-form-item>
                    <el-input v-model="authFields.kubeconfig" type="textarea" :rows="8" :placeholder="isEdit ? '已保存 kubeconfig，重新粘贴则更新' : '粘贴 kubeconfig 内容'" @input="updateKubeconfigDerivedFields" />
                  </el-form-item>
                </div>
              </transition>

              <transition name="el-fade-in-linear">
                <div v-if="form.type === 'grpc'" class="auth-fields">
                  <div class="form-label">Service / Method</div>
                  <el-form-item>
                    <el-input v-model="configFields.method" placeholder="package.Service/Method" />
                  </el-form-item>
                </div>
              </transition>

            </div>
          </div>

          <!-- ── 其他 ── -->
          <div class="extra-row">
            <div class="extra-item">
              <span class="extra-label">备注</span>
              <el-input v-model="form.remark" type="textarea" :rows="2" />
            </div>
            <div class="extra-item inline">
              <span class="extra-label">启用状态</span>
              <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
            </div>
          </div>
        </el-form>

      </div>
      <template v-if="step === 2" #footer>
        <div class="form-footer">
          <el-button class="btn-test" :loading="testing" @click="testFromForm">接入测试</el-button>
          <el-button @click="drawerVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">保存</el-button>
        </div>
      </template>
    </el-drawer>

    <!-- 测试连接结果 -->
    <el-dialog v-model="testVisible" title="测试连接" width="400px">
      <div v-loading="testing">
        <el-result v-if="testResult !== null" :icon="testResult.ok ? 'success' : 'error'"
          :title="testResult.ok ? '连接成功' : '连接失败'" :sub-title="testResult.msg" />
      </div>
      <template #footer>
        <el-button @click="testVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onBeforeUnmount, onMounted, computed } from 'vue'
import { getDatasources, createDatasource, updateDatasource, deleteDatasource, batchUpdateDatasourceStatus, batchDeleteDatasources, testConnection, testSavedConnection } from '../../api/datasource'
import { ElMessage, ElMessageBox } from 'element-plus'
import { TrendCharts, Search, Timer, Coin, Link, Notebook, Connection, Tickets, DataBoard, ArrowLeft, ArrowDown, Setting, Cloudy, Platform, Cpu, MessageBox, Monitor, Plus, Minus, Edit, Delete, Refresh } from '@element-plus/icons-vue'
import prometheusLogo from '../../assets/prometheus-logo.7aa022e5.svg'
import aliyunSlsLogo from '../../assets/aliyun_sls.svg'
import tencentClsLogo from '../../assets/tencent_cls.svg'
import kubernetesLogo from '../../assets/kubernetes.svg'
import nacosLogo from '../../assets/nacos.svg'

// ── 类型选项定义 ──

const categoryOptions = [
  { label: '可观测', value: 'observability' },
  { label: '数据库', value: 'database' },
  { label: '云资源', value: 'cloud' },
  { label: 'Kubernetes', value: 'kubernetes' },
  { label: '网络探测', value: 'network' },
  { label: '中间件', value: 'middleware' },
  { label: '其他', value: 'other' },
]

const typeOptions = [
  { label: 'Prometheus', value: 'prometheus', category: 'observability', icon: TrendCharts, color: '#E6A23C', desc: '指标监控' },
  { label: 'Loki', value: 'loki', category: 'observability', icon: Tickets, color: '#E6A23C', desc: '日志聚合' },
  { label: 'Elasticsearch', value: 'elasticsearch', category: 'observability', icon: Search, color: '#909399', desc: '全文检索' },
  { label: 'Jaeger', value: 'jaeger', category: 'observability', icon: Connection, color: '#409EFF', desc: '分布式追踪' },
  { label: 'OpenTelemetry', value: 'opentelemetry', category: 'observability', icon: DataBoard, color: '#409EFF', desc: '链路采集' },
  { label: '阿里云 SLS', value: 'sls', category: 'observability', icon: Notebook, color: '#F56C6C', desc: '日志服务' },
  { label: '腾讯云 CLS', value: 'cls', category: 'observability', icon: DataBoard, color: '#67C23A', desc: '日志服务' },
  { label: 'MySQL', value: 'mysql', category: 'database', icon: Coin, color: '#409EFF', desc: '关系型数据库' },
  { label: 'ClickHouse', value: 'clickhouse', category: 'database', icon: Coin, color: '#67C23A', desc: '分析型数据库' },
  { label: 'PostgreSQL', value: 'postgresql', category: 'database', icon: Coin, color: '#409EFF', desc: '关系型数据库' },
  { label: 'Redis', value: 'redis', category: 'database', icon: Cpu, color: '#F56C6C', desc: '缓存数据库' },
  { label: 'InfluxDB', value: 'influxdb', category: 'database', icon: Timer, color: '#67C23A', desc: '时序数据库' },
  { label: '阿里云 ECS', value: 'aliyun_ecs', category: 'cloud', icon: Cloudy, color: '#409EFF', desc: '云服务器' },
  { label: '阿里云 CDN', value: 'aliyun_cdn', category: 'cloud', icon: Cloudy, color: '#E6A23C', desc: '内容分发' },
  { label: '阿里云 DCDN', value: 'aliyun_dcdn', category: 'cloud', icon: Cloudy, color: '#E6A23C', desc: '全站加速' },
  { label: '阿里云 SLB', value: 'aliyun_slb', category: 'cloud', icon: Cloudy, color: '#67C23A', desc: '负载均衡' },
  { label: 'Kubernetes', value: 'k8s', category: 'kubernetes', icon: Platform, color: '#409EFF', desc: '容器集群' },
  { label: 'HTTP', value: 'http', category: 'network', icon: Link, color: '#909399', desc: 'HTTP 探测' },
  { label: 'gRPC', value: 'grpc', category: 'network', icon: Connection, color: '#409EFF', desc: 'gRPC 探测' },
  { label: 'TCP', value: 'tcp', category: 'network', icon: Connection, color: '#67C23A', desc: '端口连通性' },
  { label: 'ICMP', value: 'icmp', category: 'network', icon: Monitor, color: '#909399', desc: 'Ping 探测' },
  { label: 'DNS', value: 'dns', category: 'network', icon: Search, color: '#409EFF', desc: '域名解析' },
  { label: 'Kafka', value: 'kafka', category: 'middleware', icon: MessageBox, color: '#303133', desc: '消息队列' },
  { label: 'RabbitMQ', value: 'rabbitmq', category: 'middleware', icon: MessageBox, color: '#E6A23C', desc: '消息队列' },
  { label: 'Nacos', value: 'nacos', category: 'middleware', icon: DataBoard, color: '#409EFF', desc: '注册配置' },
  { label: 'Consul', value: 'consul', category: 'middleware', icon: DataBoard, color: '#67C23A', desc: '服务发现' },
]

const groupedTypeOptions = computed(() =>
  categoryOptions
    .map(category => ({ ...category, items: typeOptions.filter(item => item.category === category.value) }))
    .filter(group => group.items.length > 0)
)

function typeLabel(type) {
  const found = typeOptions.find(t => t.value === type)
  return found ? found.label : type
}

function categoryLabel(category) {
  return categoryOptions.find(c => c.value === category)?.label || category || '其他'
}

function categoryByType(type) {
  return typeOptions.find(t => t.value === type)?.category || 'other'
}

function authLabel(type) {
  return { none: '无', basic: 'Basic', bearer: 'Bearer', ak_sk: 'AK/SK', password: '密码', kubeconfig: 'Kubeconfig', tls: 'TLS' }[type] || type
}

function typeColor(type) {
  const map = { prometheus: 'warning', elasticsearch: '', influxdb: 'success', mysql: 'primary', clickhouse: 'success', redis: 'danger', http: 'info', grpc: 'primary', tcp: 'success', icmp: 'info', dns: 'primary', sls: 'danger', jaeger: '', loki: 'warning', cls: 'success', k8s: 'primary' }
  return map[type] || ''
}

// ── 列表 ──

const loading = ref(false)
const tableData = ref([])
const selectedRows = ref([])
const batchAction = ref('')
const keyword = ref('')
const typeFilter = ref('')
const statusFilter = ref('')
const sortBy = ref('id')
const sortDesc = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
let keywordSearchTimer

onMounted(async () => { await fetchData() })
onBeforeUnmount(() => clearTimeout(keywordSearchTimer))

async function fetchData() {
  loading.value = true
  try {
    const res = await getDatasources({
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value,
      type: typeFilter.value,
      status: statusFilter.value,
      sort_by: sortBy.value,
      sort_desc: sortDesc.value,
    })
    tableData.value = res.data.list
    total.value = res.data.total
    selectedRows.value = []
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  clearTimeout(keywordSearchTimer)
  page.value = 1
  fetchData()
}

function handleKeywordInput() {
  clearTimeout(keywordSearchTimer)
  keywordSearchTimer = setTimeout(handleSearch, 300)
}

function handleReset() {
  clearTimeout(keywordSearchTimer)
  keyword.value = ''
  typeFilter.value = ''
  statusFilter.value = ''
  sortBy.value = 'id'
  sortDesc.value = false
  page.value = 1
  fetchData()
}

function handleSortChange({ prop, order }) {
  sortBy.value = prop || 'id'
  sortDesc.value = order === 'descending'
  page.value = 1
  fetchData()
}

function handleSizeChange() {
  page.value = 1
  fetchData()
}

function handleSelectionChange(rows) {
  selectedRows.value = rows
}

// ── 抽屉 & 步骤 ──

const drawerVisible = ref(false)
const step = ref(1)
const isEdit = ref(false)
const hoverType = ref(null)

const selectedTypeInfo = computed(() => {
  const found = typeOptions.find(t => t.value === form.type)
  const bgMap = { prometheus: '#fef0e6', sls: '#fef0f0', jaeger: '#ecf5ff', loki: '#fef0e6', cls: '#f0f9eb', elasticsearch: '#f4f4f5', influxdb: '#f0f9eb', mysql: '#ecf5ff', http: '#f4f4f5' }
  return found
    ? { ...found, bg: bgMap[found.value] || '#ecf5ff' }
    : { label: form.type, icon: Link, color: '#909399', desc: '', bg: '#f4f4f5' }
})

const AUTH_BY_TYPE = {
  prometheus:  ['none', 'basic'],
  sls:         ['ak_sk'],
  jaeger:      ['none', 'basic'],
  loki:        ['none', 'basic'],
  cls:         ['ak_sk'],
  elasticsearch: ['none', 'basic'],
  influxdb:    ['none', 'basic'],
  opentelemetry: ['none', 'basic'],
  mysql:       ['basic'],
  clickhouse:  ['none', 'basic'],
  postgresql:  ['basic'],
  redis:       ['none', 'basic'],
  aliyun_ecs:  ['ak_sk'],
  aliyun_cdn:  ['ak_sk'],
  aliyun_dcdn: ['ak_sk'],
  aliyun_slb:  ['ak_sk'],
  k8s:         ['kubeconfig', 'bearer'],
  http:        ['none', 'basic'],
  grpc:        ['none', 'basic'],
  tcp:         ['none'],
  icmp:        ['none'],
  dns:         ['none'],
  kafka:       ['none', 'basic'],
  rabbitmq:    ['none', 'basic'],
  nacos:       ['none', 'basic'],
  consul:      ['none', 'basic'],
}

const AUTH_LABEL = { none: '无认证', basic: 'Basic Auth', bearer: 'Bearer Token', ak_sk: 'AccessKey', password: '密码', kubeconfig: 'Kubeconfig', tls: 'TLS' }

const authTypeOptions = computed(() =>
  (AUTH_BY_TYPE[form.type] || ['none']).map(v => ({
    label: AUTH_LABEL[v], value: v,
  }))
)

function ensureAuthType(type) {
  const allowed = AUTH_BY_TYPE[type] || ['none']
  if (!allowed.includes(form.auth_type)) {
    form.auth_type = allowed[0]
  }
}

// ── 表单 ──

const formRef = ref()
const form = reactive({
  id: null, name: '', category: '', type: '', url: '', auth_type: 'none',
  region: '', env: '', domain: '', capabilities: '', priority: 100,
  labels: [], remark: '', status: 1,
})
const authFields = reactive({ username: '', password: '', token: '', access_key: '', secret_key: '', kubeconfig: '' })
const configFields = reactive({ project: '', logstore: '', topic_id: '', region: '', method: '', headers: [] })
const advancedSections = ref([])
const SECRET_MASK = '******'

const aliyunRegionOptions = [
  'cn-hangzhou',
  'cn-shanghai',
  'cn-beijing',
  'cn-shenzhen',
  'cn-guangzhou',
  'cn-qingdao',
  'cn-zhangjiakou',
  'cn-huhehaote',
  'cn-chengdu',
  'cn-hongkong',
  'ap-southeast-1',
  'ap-southeast-5',
  'ap-northeast-1',
  'us-west-1',
  'us-east-1',
  'eu-central-1',
]

const tencentRegionOptions = [
  'ap-shanghai',
  'ap-guangzhou',
  'ap-beijing',
  'ap-chengdu',
  'ap-chongqing',
  'ap-nanjing',
  'ap-hongkong',
  'ap-singapore',
  'ap-bangkok',
  'ap-seoul',
  'ap-tokyo',
  'na-siliconvalley',
  'na-ashburn',
  'eu-frankfurt',
]

function isDerivedEndpointType(type) {
  return ['sls', 'cls'].includes(type)
}

function isKubeconfigMode() {
  return form.type === 'k8s' && form.auth_type === 'kubeconfig'
}

function shouldShowURLField() {
  return !isDerivedEndpointType(form.type) && !isKubeconfigMode()
}

const supportsCustomHeaders = computed(() => isHTTPURLValue(form.url) || [
  'prometheus', 'jaeger', 'loki', 'elasticsearch', 'influxdb', 'opentelemetry',
  'clickhouse', 'http', 'rabbitmq', 'nacos', 'consul',
].includes(form.type))

function isHTTPURLValue(value) {
  return /^https?:\/\//i.test(String(value || '').trim())
}

const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
}

function parseAuthConfig(configStr) {
  if (!configStr) return { username: '', password: '', token: '', access_key: '', secret_key: '', kubeconfig: '' }
  try {
    const parsed = JSON.parse(configStr)
    return {
      username: parsed.username || '',
      password: parsed.password || '',
      token: parsed.token || parsed.security_token || '',
      access_key: parsed.access_key || parsed.access_key_id || parsed.secret_id || '',
      secret_key: parsed.secret_key || parsed.access_key_secret || '',
      kubeconfig: parsed.kubeconfig || '',
    }
  } catch {
    return { username: '', password: '', token: '', access_key: '', secret_key: '', kubeconfig: '' }
  }
}

function parseConfig(configStr) {
  const defaults = { project: '', logstore: '', topic_id: '', region: '', method: '', headers: [] }
  if (!configStr) return defaults
  try {
    const parsed = JSON.parse(configStr)
    return { ...defaults, ...parsed, headers: objectToRows(parsed.custom_headers || parsed.headers || {}) }
  } catch {
    return defaults
  }
}

function parseObject(value) {
  if (typeof value === 'object' && value) return value
  try { return JSON.parse(value || '{}') } catch { return {} }
}

function objectToRows(value) {
  return Object.entries(value || {}).map(([key, entry]) => ({ key, value: String(entry ?? '') }))
}

function rowsToObject(rows) {
  return rows.reduce((result, item) => {
    const key = String(item.key || '').trim()
    if (key) result[key] = String(item.value || '').trim()
    return result
  }, {})
}

function storedDatasourceLabels(row) {
  return parseObject(row.labels)
}

function datasourceLabelEntries(row) {
  return Object.entries(storedDatasourceLabels(row)).map(([key, value]) => ({ key, value }))
}

function labelToneClass(key) {
  const normalized = String(key || '').trim().toLowerCase()
  const preferred = { env: 'yellow', owner: 'purple', service: 'blue' }
  if (preferred[normalized]) return `scope-tag-${preferred[normalized]}`
  const tones = ['blue', 'yellow', 'purple', 'cyan', 'green']
  const hash = [...normalized].reduce((total, char) => total + char.charCodeAt(0), 0)
  return `scope-tag-${tones[hash % tones.length]}`
}

function addLabel() {
  if (form.labels.length >= 20) return ElMessage.warning('接入标签最多 20 个')
  form.labels.push({ key: '', value: '' })
}

function removeLabel(index) {
  form.labels.splice(index, 1)
}

function addHeader() {
  if (configFields.headers.length >= 20) return ElMessage.warning('自定义 HTTP 头最多 20 个')
  configFields.headers.push({ key: '', value: '' })
}

function removeHeader(index) {
  configFields.headers.splice(index, 1)
}

function validateHeaders() {
  const filled = configFields.headers.filter(item => String(item.key || '').trim() || String(item.value || '').trim())
  const keys = new Set()
  for (const item of filled) {
    const key = String(item.key || '').trim()
    if (!key) return '请填写 Header 名称'
    if (!/^[!#$%&'*+.^_`|~0-9A-Za-z-]+$/.test(key)) return `Header 名称 ${key} 格式不正确`
    const normalized = key.toLowerCase()
    if (keys.has(normalized)) return `Header 名称 ${key} 重复`
    keys.add(normalized)
  }
  return ''
}

function validateLabels() {
  const filled = form.labels.filter(item => String(item.key || '').trim() || String(item.value || '').trim())
  if (filled.length > 20) return '接入标签最多 20 个'
  const keys = new Set()
  for (const item of filled) {
    const key = String(item.key || '').trim()
    if (!key) return '请填写标签 Key'
    if (!/^[A-Za-z0-9_.-]+$/.test(key)) return '标签 Key 只能包含字母、数字、点、下划线和短横线'
    const normalized = key.toLowerCase()
    if (keys.has(normalized)) return `标签 Key ${key} 重复`
    keys.add(normalized)
  }
  return ''
}

function buildAuthConfig() {
  if (form.auth_type === 'basic') {
    return JSON.stringify({ username: authFields.username, password: authFields.password })
  }
  if (form.auth_type === 'bearer') {
    if (isEdit.value && isMaskedSecret(authFields.token)) {
      return ''
    }
    return JSON.stringify({ token: authFields.token })
  }
  if (form.auth_type === 'ak_sk') {
    if (isEdit.value && !hasNewAKSK()) {
      return ''
    }
    return JSON.stringify({
      access_key_id: authFields.access_key,
      access_key_secret: authFields.secret_key,
      secret_id: authFields.access_key,
      secret_key: authFields.secret_key,
    })
  }
  if (form.auth_type === 'password') {
    return JSON.stringify({ password: authFields.password })
  }
  if (form.auth_type === 'kubeconfig') {
    if (isEdit.value && isMaskedSecret(authFields.kubeconfig)) {
      return ''
    }
    updateKubeconfigDerivedFields()
    return JSON.stringify({ kubeconfig: authFields.kubeconfig })
  }
  if (form.auth_type === 'tls') {
    return JSON.stringify({ tls: true })
  }
  return ''
}

function buildConfig() {
  const config = {}
  if (form.type === 'sls') {
    Object.assign(config, { project: configFields.project, region: configFields.region })
  }
  if (form.type === 'cls') {
    Object.assign(config, { topic_id: configFields.topic_id, region: configFields.region })
  }
  if (form.type === 'grpc') {
    config.method = configFields.method
  }
  const customHeaders = rowsToObject(configFields.headers)
  if (Object.keys(customHeaders).length) config.custom_headers = customHeaders
  return Object.keys(config).length ? JSON.stringify(config) : ''
}

// ── 操作 ──

function handleAdd() {
  isEdit.value = false
  step.value = 1
  resetForm('')
  drawerVisible.value = true
}

function selectType(type) {
  resetForm(type)
  form.category = categoryByType(type)
  ensureAuthType(type)
  if (type === 'sls') {
    configFields.region = 'cn-hangzhou'
    updateSlsDerivedFields()
  }
  if (type === 'cls') {
    configFields.region = 'ap-shanghai'
    updateClsDerivedFields()
  }
  if (type === 'k8s') {
    form.auth_type = 'kubeconfig'
  }
  step.value = 2
}

function resetForm(type) {
  Object.assign(form, {
    id: null, name: '', category: categoryByType(type), type, url: '', auth_type: 'none',
    region: '', env: '', domain: '', capabilities: '', priority: 100,
    labels: [], remark: '', status: 1,
  })
  Object.assign(authFields, { username: '', password: '', token: '', access_key: '', secret_key: '', kubeconfig: '' })
  Object.assign(configFields, { project: '', logstore: '', topic_id: '', region: '', method: '', headers: [] })
  advancedSections.value = []
}

function handleEdit(row) {
  isEdit.value = true
  step.value = 2
  Object.assign(form, {
    id: row.id, name: row.name, category: row.category || categoryByType(row.type), type: row.type, url: row.url,
    auth_type: row.auth_type,
    region: row.region || '', env: row.env || '', domain: row.domain || '',
    capabilities: row.capabilities || '', priority: row.priority || 100,
    labels: objectToRows(storedDatasourceLabels(row)),
    remark: row.remark, status: row.status,
  })
  Object.assign(authFields, emptyAuthFields())
  if (row.auth_type === 'basic' && row.auth_config) {
    Object.assign(authFields, parseAuthConfig(row.auth_config))
  }
  if (row.auth_type === 'ak_sk' && row.auth_config) {
    authFields.access_key = SECRET_MASK
    authFields.secret_key = SECRET_MASK
  }
  if (row.auth_type === 'bearer' && row.auth_config) {
    authFields.token = SECRET_MASK
  }
  if (row.auth_type === 'kubeconfig' && row.auth_config) {
    authFields.kubeconfig = SECRET_MASK
  }
  Object.assign(configFields, parseConfig(row.config || ''))
  if (form.type === 'sls') updateSlsDerivedFields(false)
  if (form.type === 'cls') updateClsDerivedFields(false)
  if (isKubeconfigMode()) updateKubeconfigDerivedFields()
  drawerVisible.value = true
}

function emptyAuthFields() {
  return { username: '', password: '', token: '', access_key: '', secret_key: '', kubeconfig: '' }
}

function isMaskedSecret(value) {
  return value === SECRET_MASK
}

function hasNewAKSK() {
  return Boolean(
    authFields.access_key &&
    authFields.secret_key &&
    !isMaskedSecret(authFields.access_key) &&
    !isMaskedSecret(authFields.secret_key)
  )
}

function hasNewBearerToken() {
  return Boolean(authFields.token && !isMaskedSecret(authFields.token))
}

function hasNewKubeconfig() {
  return Boolean(authFields.kubeconfig && !isMaskedSecret(authFields.kubeconfig))
}

function handleDrawerClose() {
  formRef.value?.resetFields()
}

function handleDrawerBack() {
  if (!isEdit.value && step.value === 2) {
    step.value = 1
    return
  }
  drawerVisible.value = false
}

async function handleSubmit() {
  const labelError = validateLabels()
  if (labelError) return ElMessage.warning(labelError)
  if (!prepareDatasourcePayload()) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  const payload = { ...form, labels: JSON.stringify(rowsToObject(form.labels)), category: form.category || categoryByType(form.type), auth_config: buildAuthConfig(), config: buildConfig() }
  try {
    if (isEdit.value) {
      await updateDatasource(form.id, payload)
      ElMessage.success('更新成功')
    } else {
      await createDatasource(payload)
      ElMessage.success('创建成功')
    }
    drawerVisible.value = false
    fetchData()
  } catch {}
}

async function handleDelete(row) {
  await ElMessageBox.confirm(`确定删除数据源「${row.name}」吗？`, '提示', { type: 'warning' })
  await deleteDatasource(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

function handleBatchCommand(command) {
  if (command === 'enable') return handleBatchStatus(1)
  if (command === 'disable') return handleBatchStatus(0)
  if (command === 'delete') return handleBatchDelete()
}

async function handleBatchStatus(status) {
  if (!selectedRows.value.length) return
  batchAction.value = status === 1 ? 'enable' : 'disable'
  try {
    await batchUpdateDatasourceStatus(selectedRows.value.map(row => row.id), status)
    ElMessage.success(`已批量${status === 1 ? '启用' : '禁用'} ${selectedRows.value.length} 个接入`)
    await fetchData()
  } finally {
    batchAction.value = ''
  }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  const count = selectedRows.value.length
  const confirmed = await ElMessageBox.confirm(
    `确定删除选中的 ${count} 个接入吗？删除后无法恢复。`,
    '批量删除',
    { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
  ).then(() => true).catch(() => false)
  if (!confirmed) return

  batchAction.value = 'delete'
  try {
    await batchDeleteDatasources(selectedRows.value.map(row => row.id))
    ElMessage.success(`已删除 ${count} 个接入`)
    await fetchData()
  } finally {
    batchAction.value = ''
  }
}

// ── 测试连接 ──

const testing = ref(false)
const testVisible = ref(false)
const testResult = ref(null)

async function testFromForm() {
  if (!prepareDatasourcePayload()) return
  testing.value = true
  testResult.value = null
  testVisible.value = true
  try {
    const res = await runFormConnectionTest()
    testResult.value = { ok: res.code === 0, msg: res.message }
  } catch {
    testResult.value = { ok: false, msg: '测试请求失败' }
  } finally {
    testing.value = false
  }
}

async function runFormConnectionTest() {
  if (isEdit.value && form.id && shouldUseSavedAuthForTest()) {
    return testSavedConnection(form.id, {
      type: form.type,
      url: form.url,
      auth_type: form.auth_type,
      auth_config: buildAuthConfig(),
      config: buildConfig(),
    })
  }
  return testConnection({
    type: form.type,
    url: form.url,
    auth_type: form.auth_type,
    auth_config: buildAuthConfig(),
    config: buildConfig(),
  })
}

function shouldUseSavedAuthForTest() {
  if (form.auth_type === 'ak_sk') return !hasNewAKSK()
  if (form.auth_type === 'bearer') return !hasNewBearerToken()
  if (form.auth_type === 'kubeconfig') return !hasNewKubeconfig()
  return false
}

function slsEndpoint() {
  if (!configFields.project || !configFields.region) return ''
  return `https://${configFields.project}.${configFields.region}.log.aliyuncs.com`
}

function updateSlsDerivedFields(shouldFillName = true) {
  if (form.type !== 'sls') return
  const endpoint = slsEndpoint()
  if (endpoint) form.url = endpoint
  if (shouldFillName && !form.name && configFields.project) {
    form.name = `${configFields.project} SLS`
  }
}

function clsEndpoint() {
  return 'https://cls.tencentcloudapi.com'
}

function updateClsDerivedFields(shouldFillName = true) {
  if (form.type !== 'cls') return
  configFields.region = normalizeTencentRegion(configFields.region)
  form.url = clsEndpoint()
  if (shouldFillName && !form.name && configFields.topic_id) {
    form.name = `${configFields.topic_id} CLS`
  }
}

function updateKubeconfigDerivedFields() {
  if (!isKubeconfigMode()) return
  if (isMaskedSecret(authFields.kubeconfig)) return
  const server = extractKubeconfigServer(authFields.kubeconfig)
  if (server) {
    form.url = server
  }
}

function extractKubeconfigServer(kubeconfig) {
  const text = String(kubeconfig || '')
  const match = text.match(/^\s*server:\s*['"]?([^'"\s]+)['"]?\s*$/m)
  return match ? match[1].trim() : ''
}

function normalizeTencentRegion(region) {
  const value = String(region || '').trim()
  const map = {
    shanghai: 'ap-shanghai',
    guangzhou: 'ap-guangzhou',
    beijing: 'ap-beijing',
    chengdu: 'ap-chengdu',
    chongqing: 'ap-chongqing',
    nanjing: 'ap-nanjing',
    hongkong: 'ap-hongkong',
    singapore: 'ap-singapore',
    bangkok: 'ap-bangkok',
    seoul: 'ap-seoul',
    tokyo: 'ap-tokyo',
    上海: 'ap-shanghai',
    广州: 'ap-guangzhou',
    北京: 'ap-beijing',
    成都: 'ap-chengdu',
    重庆: 'ap-chongqing',
    南京: 'ap-nanjing',
    香港: 'ap-hongkong',
  }
  return map[value.toLowerCase()] || map[value] || value
}

function prepareDatasourcePayload() {
  const headerError = validateHeaders()
  if (headerError) {
    ElMessage.warning(headerError)
    return false
  }
  if (form.auth_type === 'basic') {
    if (!String(authFields.username || '').trim()) {
      ElMessage.warning('请输入 Basic Auth 用户名')
      return false
    }
    if (!authFields.password) {
      ElMessage.warning('请输入 Basic Auth 密码')
      return false
    }
  }
  if (form.type === 'sls') {
    updateSlsDerivedFields(false)
    if (!configFields.region) {
      ElMessage.warning('请选择或输入阿里云 SLS Region')
      return false
    }
    if (!configFields.project) {
      ElMessage.warning('请输入阿里云 SLS Project 名称')
      return false
    }
    form.url = slsEndpoint()
    if (!form.url) {
      ElMessage.warning('无法生成 SLS 接入地址，请检查 Region 和 Project')
      return false
    }
    if (!isEdit.value && (!authFields.access_key || !authFields.secret_key)) {
      ElMessage.warning('请输入 AccessKey ID 和 AccessKey Secret')
      return false
    }
    const akChanged = authFields.access_key && !isMaskedSecret(authFields.access_key)
    const skChanged = authFields.secret_key && !isMaskedSecret(authFields.secret_key)
    if (isEdit.value && (akChanged !== skChanged)) {
      ElMessage.warning('如需修改 AK/SK，请同时填写 AccessKey ID 和 AccessKey Secret')
      return false
    }
  }
  if (form.type === 'cls') {
    updateClsDerivedFields(false)
    if (!configFields.region) {
      ElMessage.warning('请选择或输入腾讯云 CLS Region')
      return false
    }
    if (!configFields.topic_id) {
      ElMessage.warning('请输入腾讯云 CLS 日志主题 ID')
      return false
    }
    form.url = clsEndpoint()
    if (!isEdit.value && (!authFields.access_key || !authFields.secret_key)) {
      ElMessage.warning('请输入 SecretId 和 SecretKey')
      return false
    }
    const akChanged = authFields.access_key && !isMaskedSecret(authFields.access_key)
    const skChanged = authFields.secret_key && !isMaskedSecret(authFields.secret_key)
    if (isEdit.value && (akChanged !== skChanged)) {
      ElMessage.warning('如需修改 SecretId/SecretKey，请同时填写 SecretId 和 SecretKey')
      return false
    }
  }
  if (form.type === 'k8s') {
    if (form.auth_type === 'kubeconfig') {
      updateKubeconfigDerivedFields()
      if (!authFields.kubeconfig) {
        ElMessage.warning('请粘贴 kubeconfig 内容')
        return false
      }
      if (!isEdit.value && !form.url) {
        ElMessage.warning('kubeconfig 中未找到 clusters[].cluster.server')
        return false
      }
    }
    if (form.auth_type === 'bearer') {
      if (!form.url) {
        ElMessage.warning('Bearer Token 模式请填写 Kubernetes API Server 地址')
        return false
      }
      if (!authFields.token) {
        ElMessage.warning('Bearer Token 模式请填写 Token')
        return false
      }
    }
  }
  return true
}
</script>

<style scoped>
:global(.datasource-drawer.el-drawer) { border-radius: 0 !important; }
:global(.datasource-drawer .el-drawer__footer) {
  padding: 12px 20px;
  border-top: 1px solid var(--ops-border);
}
.drawer-header {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 12px;
}
.back-button { height: 32px; padding: 0 8px; color: #3b6ea8; font-size: 14px; }
.drawer-title { color: #303133; font-size: 16px; font-weight: 600; }
:global(.datasource-drawer .el-drawer__header) {
  margin-bottom: 0;
  padding: 18px 20px;
  border-bottom: 1px solid #e5e9f2;
}
.page-container { display: flex; flex-direction: column; gap: 16px; }
.datasource-toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}
.datasource-filters {
  display: flex;
  min-width: 0;
  flex: 1 1 auto;
  align-items: center;
  gap: 8px;
  margin: 0;
  padding: 0;
}
.filter-type { width: 160px; flex: 0 0 160px; }
.filter-status { width: 120px; flex: 0 0 120px; }
.filter-keyword { width: 220px; min-width: 0; flex: 0 0 220px; }
.datasource-filters :deep(.el-button + .el-button) { margin-left: 0; }
.datasource-actions { display: flex; flex: 0 0 auto; align-items: center; gap: 8px; }
.datasource-actions :deep(.el-button + .el-button) { margin-left: 0; }
.batch-operation-arrow { margin-left: 6px; vertical-align: middle; }
.add-source-button { flex: 0 0 auto; }
.datasource-table-card { overflow: hidden; }
.pagination { display: flex; justify-content: flex-end; margin-top: 16px; }
.scope-tags {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.category-text { color: #606266; }
.scope-tag {
  display: inline-flex;
  height: 24px;
  align-items: center;
  padding: 0 8px;
  border: 1px solid transparent;
  border-radius: 4px;
  font-size: 12px;
  line-height: 22px;
  white-space: nowrap;
}
.scope-tag-yellow { color: #ad6800; background: #fffbe6; border-color: #ffe58f; }
.scope-tag-purple { color: #531dab; background: #f9f0ff; border-color: #d3adf7; }
.scope-tag-blue { color: #1d39c4; background: #f0f5ff; border-color: #adc6ff; }
.scope-tag-cyan { color: #006d75; background: #e6fffb; border-color: #87e8de; }
.scope-tag-green { color: #237804; background: #f6ffed; border-color: #b7eb8f; }
.status-container {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: #606266;
}
.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex: 0 0 auto;
}
.status-enabled {
  background: #67c23a;
  animation: status-pulse 2.2s ease-out infinite;
}
.status-disabled { background: #c0c4cc; }
@keyframes status-pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(103, 194, 58, 0); }
  45% { box-shadow: 0 0 0 4px rgba(103, 194, 58, 0.18); }
}
@media (prefers-reduced-motion: reduce) {
  .status-enabled { animation: none; }
}
.type-logo-cell {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 24px;
  vertical-align: middle;
}
.type-logo-cell-wide {
  width: 54px;
}
.type-logo {
  width: 18px;
  height: 18px;
  display: block;
}
.type-logo-wide {
  width: 48px;
  height: 14px;
  object-fit: contain;
}

/* ── 步骤 1：类型选择卡片 ── */
.type-groups {
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.type-group-title {
  font-size: 14px;
  font-weight: 700;
  color: #303133;
  margin-bottom: 10px;
}
.type-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}
.type-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px 12px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
  min-height: 110px;
}
.type-card:hover,
.type-card.active {
  border-color: #409EFF;
  background-color: #ecf5ff;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.15);
}
.type-card-name {
  font-size: 14px;
  font-weight: 500;
  margin-top: 8px;
}
.type-card-desc {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
.type-card-logo {
  width: 32px;
  height: 32px;
  display: block;
}
.type-card-logo-wide {
  width: 76px;
  height: 26px;
  object-fit: contain;
}
.type-note {
  padding: 10px 12px;
  background: #f4f8ff;
  border: 1px solid #d9ecff;
  border-radius: 8px;
  color: #606266;
  font-size: 13px;
  line-height: 1.5;
}

/* ── 步骤 2：表单 ── */
.step2-wrapper {
  animation: fadeInUp 0.3s ease;
}
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(8px); }
  to   { opacity: 1; transform: translateY(0); }
}

/* 已选类型横幅 */
.selected-type-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  margin-bottom: 16px;
  background: #f8fafc;
  border: 1px solid #e4e7ed;
  border-radius: 10px;
}
.selected-type-icon {
  width: 42px;
  height: 42px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  flex-shrink: 0;
}
.selected-type-icon-wide {
  width: 86px;
}
.selected-type-logo {
  width: 24px;
  height: 24px;
  display: block;
}
.selected-type-logo-wide {
  width: 76px;
  height: 18px;
  object-fit: contain;
}
.selected-type-label {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}
.selected-type-desc {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

/* 表单容器 */
.config-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* 扁平表单区块 */
.section-card {
  background: transparent;
}
.section-body {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.label-editor { display: flex; flex-direction: column; gap: 8px; }
.add-label-button { width: 100%; border-style: dashed; }
.label-row { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 28px; align-items: center; gap: 8px; }
.label-row .label-icon-button {
  width: 26px;
  height: 26px;
  color: #8a94a3;
  background: #fff;
  border-color: #dcdfe6;
}
.label-row .label-icon-button :deep(.el-icon) { font-size: 14px; }
.label-row .label-icon-button:hover,
.label-row .label-icon-button:focus-visible {
  color: #606b7b;
  background: #f5f7fa;
  border-color: #c8cdd5;
}
.field-help { margin-top: 7px; color: #909399; font-size: 12px; line-height: 1.5; }

/* 表单标签（放在 el-form-item 上方） */
.form-label {
  font-size: 13px;
  font-weight: 500;
  color: #606266;
  margin-bottom: 6px;
  margin-top: 4px;
}
.form-label .required {
  color: #F56C6C;
  margin-left: 2px;
}

/* 认证方式 —— 分段选择器 */
.auth-segmented {
  display: flex;
  width: 100%;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  overflow: hidden;
  background: #f5f7fa;
}
.auth-seg {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  padding: 8px 14px;
  border: none;
  background: transparent;
  font-size: 13px;
  color: #606266;
  cursor: pointer;
  transition: all 0.2s;
  outline: none;
  white-space: nowrap;
  font-family: inherit;
}
.auth-seg + .auth-seg {
  border-left: 1px solid #dcdfe6;
}
.auth-seg:hover {
  color: #409EFF;
  background: rgba(64,158,255,0.04);
}
.auth-seg.active {
  background: #409EFF;
  color: #fff;
}

/* 认证字段过渡 */
.auth-fields {
  padding-top: 4px;
}

.advanced-config {
  margin-top: 4px;
  overflow: hidden;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
}
.advanced-config :deep(.el-collapse-item__header) {
  height: 48px;
  padding: 0 14px;
  color: #4b5565;
  border-bottom: 0;
}
.advanced-config :deep(.el-collapse-item__wrap) { border-bottom: 0; }
.advanced-config :deep(.el-collapse-item__content) { padding: 0 14px 16px; }
.advanced-title { display: inline-flex; align-items: center; gap: 8px; font-weight: 600; }
.advanced-content { display: flex; flex-direction: column; gap: 14px; }
.advanced-heading { color: #303133; font-size: 13px; font-weight: 600; }
.advanced-heading span { color: #909399; font-weight: 400; }
.header-editor { display: flex; flex-direction: column; gap: 8px; }
.header-row { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 28px; align-items: center; gap: 8px; }
.header-remove-button {
  width: 26px;
  height: 26px;
  color: #8a94a3;
  background: #fff;
  border-color: #dcdfe6;
}
.header-remove-button:hover,
.header-remove-button:focus-visible {
  color: #606b7b;
  background: #f5f7fa;
  border-color: #c8cdd5;
}
.add-header-button { width: 100%; border-style: dashed; }

/* 其他行 */
.extra-row {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.extra-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.extra-item.inline {
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
}
.extra-label {
  font-size: 13px;
  font-weight: 500;
  color: #606266;
}

/* 表单底部 */
.form-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}
.btn-test {
  margin-right: auto;
  border-color: #67C23A;
  color: #67C23A;
  background: #f0f9eb;
}
.btn-test:hover {
  border-color: #67C23A;
  color: #fff;
  background: #67C23A;
}

@media (max-width: 900px) {
  .datasource-filters { flex-basis: 100%; }
  .datasource-actions { margin-left: auto; }
}

@media (max-width: 680px) {
  .filter-select { flex: 1 1 calc(50% - 4px); }
  .filter-keyword { min-width: 100%; flex-basis: 100%; }
}

@media (max-width: 460px) {
  .filter-select { flex-basis: 100%; }
}
</style>
