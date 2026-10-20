<template>
  <div class="page-container">
    <div class="query-context-bar">
      <div class="datasource-picker">
        <el-select
          v-model="selectedDS"
          class="datasource-select"
          aria-label="Prometheus 数据源"
          multiple
          filterable
          collapse-tags
          collapse-tags-tooltip
          :max-collapse-tags="1"
          placeholder="选择数据源"
        >
          <el-option v-for="ds in datasources" :key="ds.id" :label="ds.name" :value="ds.id" />
        </el-select>
      </div>

      <el-tooltip content="历史查询" placement="left">
        <el-button class="history-button" aria-label="历史查询" @click="openQueryHistory">
          <svg class="history-icon" viewBox="64 64 896 896" aria-hidden="true">
            <path d="M536.1 273H488c-4.4 0-8 3.6-8 8v275.3c0 2.6 1.2 5 3.3 6.5l165.3 120.7c3.6 2.6 8.6 1.9 11.2-1.7l28.6-39c2.7-3.7 1.9-8.7-1.7-11.2L544.1 528.5V281c0-4.4-3.6-8-8-8zm219.8 75.2l156.8 38.3c5 1.2 9.9-2.6 9.9-7.7l.8-161.5c0-6.7-7.7-10.5-12.9-6.3L752.9 334.1a8 8 0 003 14.1zm167.7 301.1l-56.7-19.5a8 8 0 00-10.1 4.8c-1.9 5.1-3.9 10.1-6 15.1-17.8 42.1-43.3 80-75.9 112.5a353 353 0 01-112.5 75.9 352.18 352.18 0 01-137.7 27.8c-47.8 0-94.1-9.3-137.7-27.8a353 353 0 01-112.5-75.9c-32.5-32.5-58-70.4-75.9-112.5A353.44 353.44 0 01171 512c0-47.8 9.3-94.2 27.8-137.8 17.8-42.1 43.3-80 75.9-112.5a353 353 0 01112.5-75.9C430.6 167.3 477 158 524.8 158s94.1 9.3 137.7 27.8A353 353 0 01775 261.7c10.2 10.3 19.8 21 28.6 32.3l59.8-46.8C784.7 146.6 662.2 81.9 524.6 82 285 82.1 92.6 276.7 95 516.4 97.4 751.9 288.9 942 524.8 942c185.5 0 343.5-117.6 403.7-282.3 1.5-4.2-.7-8.9-4.9-10.4z" />
          </svg>
        </el-button>
      </el-tooltip>
    </div>

    <div class="query-run-row">
      <div class="query-input-row">
        <div class="input-shell">
          <div class="input-leading">
            <el-icon><Monitor /></el-icon>
          </div>
          <textarea
            ref="inputRef"
            v-model="query"
            class="promql-input"
            :placeholder="inputPlaceholder"
            rows="1"
            @keydown="onKeydown"
            @input="onInput"
            spellcheck="false"
          />
        </div>

        <!-- 自动补全 -->
        <div v-if="showSuggestions" ref="suggestRef" class="autocomplete-dropdown" :style="dropdownStyle">
          <div v-if="suggestLoading" class="suggest-loading">加载中...</div>
          <template v-else>
            <div v-for="(item, idx) in suggestions" :key="item" class="suggest-item"
              :class="{ active: idx === selectedIdx }"
              @mousedown.prevent="selectSuggestion(item)" @mouseenter="selectedIdx = idx">
              <span class="suggest-type-icon">{{ suggestType === 'metric' ? 'M' : suggestType === 'label' ? 'L' : 'V' }}</span>
              <span class="suggest-text">
                <span
                  v-for="(part, partIdx) in highlightedSuggestionParts(item)"
                  :key="`${item}-${partIdx}`"
                  :class="{ matched: part.matched }"
                >{{ part.text }}</span>
              </span>
            </div>
            <div v-if="suggestions.length === 0 && !suggestLoading" class="suggest-empty">无匹配项</div>
          </template>
        </div>
      </div>
      <el-button type="primary" class="execute-button" @click="executeQuery" :loading="loading">Execute</el-button>
    </div>

    <div class="result-panel">
      <div class="result-tabs">
        <button class="tab" :class="{ active: activeTab === 'table' }" type="button" @click="switchToTable()">
          <el-icon><Tickets /></el-icon>
          <span>Table</span>
        </button>
        <button class="tab" :class="{ active: activeTab === 'graph' }" type="button" @click="switchToGraph()">
          <el-icon><TrendCharts /></el-icon>
          <span>Graph</span>
        </button>
        <div class="tab-meta">
          <span v-if="result" class="result-count">{{ result.data?.result?.length || 0 }} 条结果</span>
          <span v-if="execTime" class="exec-time">{{ execTime }}</span>
        </div>
      </div>

      <!-- Table -->
      <div v-show="activeTab === 'table'" class="result-content">
        <div class="panel-toolbar">
          <div class="eval-control">
            <el-button text :icon="ArrowLeft" title="Decrease time" aria-label="Decrease time" @click="shiftEvalTime(-1)" />
            <el-date-picker
              v-model="evalTime"
              type="datetime"
              placeholder="Evaluation time"
              format="YYYY-MM-DD HH:mm:ss"
              value-format="YYYY-MM-DDTHH:mm:ss"
              size="small"
              class="eval-picker"
              @change="onTimeChange"
            />
            <el-button text :icon="ArrowRight" title="Increase time" aria-label="Increase time" @click="shiftEvalTime(1)" />
          </div>
        </div>
        <div v-if="queryWarning" class="warning-box"><el-icon><WarningFilled /></el-icon>{{ queryWarning }}</div>
        <div v-if="error" class="error-box"><el-icon color="#F56C6C"><WarningFilled /></el-icon>{{ error }}</div>
        <el-table v-else-if="result" :data="tableRows" class="query-result-table" stripe size="small" max-height="55vh" style="width:100%">
          <el-table-column label="数据源" width="180" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="result-source">{{ row.datasourceName }}</span>
            </template>
          </el-table-column>
          <el-table-column label="标签" min-width="360" show-overflow-tooltip>
            <template #default="{ row }"><span class="metric-cell">{{ row.element }}</span></template>
          </el-table-column>
          <el-table-column label="值" width="180" align="right" header-align="right" show-overflow-tooltip>
            <template #default="{ row }"><span class="value-cell">{{ row.displayValue }}</span></template>
          </el-table-column>
          <template #empty><TableEmpty /></template>
        </el-table>
        <div v-else class="query-alert">No data queried yet</div>
      </div>

      <!-- Graph -->
      <div v-show="activeTab === 'graph'" class="result-content">
        <div class="graph-params">
          <div class="range-item">
            <span class="toolbar-label">Range</span>
            <el-select v-model="graphRange" size="small" style="width:100px" @change="onGraphRangeChange">
              <el-option v-for="r in rangeOptions" :key="r.value" :label="r.label" :value="r.value" />
            </el-select>
          </div>
          <div class="range-item">
            <span class="toolbar-label">End time</span>
            <el-date-picker v-model="graphEndTime" type="datetime" format="YYYY-MM-DD HH:mm:ss"
              value-format="YYYY-MM-DDTHH:mm:ss" size="small" style="width:195px" @change="onGraphParamChange" />
          </div>
          <div class="range-item">
            <span class="toolbar-label">Step</span>
            <el-select v-model="graphStep" size="small" style="width:110px" @change="onGraphParamChange">
              <el-option v-for="s in stepOptions" :key="s.value" :label="s.label" :value="s.value" :disabled="isGraphStepTooFine(s.value)" />
            </el-select>
          </div>
        </div>
        <div v-if="queryWarning" class="warning-box"><el-icon><WarningFilled /></el-icon>{{ queryWarning }}</div>
        <div v-if="error" class="error-box"><el-icon color="#F56C6C"><WarningFilled /></el-icon>{{ error }}</div>
        <div v-else-if="chartSeries.length === 0" class="empty-chart">
          当前结果无可渲染的图表数据（请切换到范围查询模式）
        </div>
        <template v-else>
          <v-chart :option="chartOption" :autoresize="true" style="height:380px" />
          <div class="metric-legend" aria-label="图表指标">
            <div class="metric-legend-header" aria-hidden="true">
              <span>数据源</span>
              <span>标签</span>
            </div>
            <button
              v-for="(series, index) in chartSeries"
              :key="`${series.name}-${index}`"
              type="button"
              class="metric-legend-item"
              :class="{ active: hasSeriesFilter && isMetricSelected(index), muted: hasSeriesFilter && !isMetricSelected(index) }"
              :aria-pressed="isMetricSelected(index)"
              title="点击仅显示此指标，Shift + 点击可多选"
              @click="toggleMetric(index, $event)"
            >
              <span class="metric-legend-line" :style="{ '--series-color': chartColors[index % chartColors.length] }" aria-hidden="true"></span>
              <span class="metric-legend-source" :title="series.datasourceName">{{ series.datasourceName }}</span>
              <span class="metric-legend-name">{{ series.metricName }}</span>
            </button>
          </div>
        </template>
      </div>
    </div>

    <el-drawer v-model="historyVisible" class="query-history-drawer" size="420px" append-to-body>
      <template #header>
        <div class="history-drawer-header">
          <strong>历史查询</strong>
          <el-button v-if="queryHistory.length" text type="danger" @click="clearQueryHistory">清空</el-button>
        </div>
      </template>
      <div v-if="queryHistory.length" class="history-list">
        <button v-for="item in queryHistory" :key="item.id" class="history-item" type="button" @click="restoreQuery(item)">
          <span class="history-query">{{ item.query }}</span>
          <span class="history-meta">
            <el-tag size="small" effect="plain">{{ item.mode === 'graph' ? 'Graph' : 'Table' }}</el-tag>
            <span>{{ item.datasource_name }}</span>
            <time :datetime="item.created_at">{{ formatHistoryTime(item.created_at) }}</time>
          </span>
        </button>
      </div>
      <el-empty v-else description="暂无查询记录" :image-size="72" />
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount, computed, nextTick, watch } from 'vue'
import { promInstantQuery, promRangeQuery, getPromDatasources, promMetadata, getPromQueryHistory, clearPromQueryHistory } from '../../api/prometheus'
import { ElMessage } from 'element-plus'
import { ArrowLeft, ArrowRight, Monitor, Tickets, TrendCharts, WarningFilled } from '@element-plus/icons-vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, DataZoomComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([LineChart, GridComponent, TooltipComponent, DataZoomComponent, CanvasRenderer])

const datasources = ref([])
const selectedDS = ref([])
const query = ref('')
const loading = ref(false)
const result = ref(null)
const error = ref('')
const queryWarning = ref('')
const activeTab = ref('table')
const execTime = ref('')
const evalTime = ref(null)
const inputRef = ref(null)
const resultMode = ref('')
const historyVisible = ref(false)
const queryHistory = ref([])

async function loadQueryHistory() {
  try {
    const res = await getPromQueryHistory()
    queryHistory.value = res.data || []
  } catch {}
}

async function openQueryHistory() {
  historyVisible.value = true
  await loadQueryHistory()
}

function restoreQuery(item) {
  const datasourceExists = datasources.value.some(ds => String(ds.id) === String(item.datasource_id))
  if (datasourceExists) selectedDS.value = [item.datasource_id]
  else ElMessage.warning('该历史记录的数据源已不可用')
  query.value = item.query
  activeTab.value = item.mode === 'graph' ? 'graph' : 'table'
  historyVisible.value = false
  nextTick(() => inputRef.value?.focus())
}

async function clearQueryHistory() {
  try {
    await clearPromQueryHistory()
    queryHistory.value = []
    ElMessage.success('历史记录已清空')
  } catch {}
}

function formatHistoryTime(value) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleString('zh-CN', { hour12: false })
}

// Graph params
const rangeOptions = [
  { label: '1h',  value: '1h' },
  { label: '6h',  value: '6h' },
  { label: '12h', value: '12h' },
  { label: '1d',  value: '1d' },
  { label: '2d',  value: '2d' },
  { label: '7d',  value: '7d' },
  { label: '30d', value: '30d' },
]
const stepOptions = [
  { label: '10s',  value: '10s' },
  { label: '30s',  value: '30s' },
  { label: '1m',   value: '1m' },
  { label: '5m',   value: '5m' },
  { label: '15m',  value: '15m' },
  { label: '1h',   value: '1h' },
]
const graphRange = ref('1h')
const graphEndTime = ref('')
const graphStep = ref('1m')
const maxPointsPerSeries = 11000

const inputPlaceholder = '输入 PromQL 表达式'

onMounted(async () => {
  try {
    const res = await getPromDatasources()
    datasources.value = res.data || []
    if (datasources.value.length > 0) selectedDS.value = [datasources.value[0].id]
  } catch {}
  await loadQueryHistory()
  initGraphTime()
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => document.removeEventListener('click', handleClickOutside))

function initGraphTime() {
  const now = new Date()
  graphEndTime.value = formatDateTime(now)
}

function formatDateTime(d) {
  const pad = n => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function toPromTimestamp(value) {
  if (!value) return undefined
  const date = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(date.getTime())) return undefined
  return date.toISOString()
}

// ── 自动补全 ──

const showSuggestions = ref(false)
const suggestions = ref([])
const selectedIdx = ref(0)
const suggestLoading = ref(false)
const suggestType = ref('metric')
const suggestToken = ref('')
const replaceRange = ref({ start: 0, end: 0 })
const dropdownStyle = ref({ top: '0px', left: '0px' })
let suggestTimer = null

function onInput() {
  const el = inputRef.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 150) + 'px'
  clearTimeout(suggestTimer)
  const text = query.value
  const pos = el.selectionStart
  suggestTimer = setTimeout(() => triggerSuggest(text, pos, el), 200)
}

function triggerSuggest(text, pos, el) {
  const { token, start, end, type, labelName } = parseTokenAt(text, pos)
  if (!token || selectedDS.value.length === 0) { showSuggestions.value = false; return }
  suggestType.value = type
  suggestToken.value = token
  replaceRange.value = { start, end }
  const posPx = cursorPixelPos(el, start)
  dropdownStyle.value = { top: (posPx.top + 22) + 'px', left: (posPx.left + 4) + 'px' }
  suggestLoading.value = true; showSuggestions.value = true; selectedIdx.value = 0; suggestions.value = []
  const fetchers = selectedDS.value.map(dsId => type === 'label'
    ? promMetadata({ datasource_id: dsId, target: 'labels' })
    : type === 'value'
      ? promMetadata({ datasource_id: dsId, target: 'label_values', label: labelName })
      : promMetadata({ datasource_id: dsId, target: 'metrics' }))
  Promise.allSettled(fetchers).then(results => {
    const all = [...new Set(results.flatMap(item => item.status === 'fulfilled' ? (item.value?.data || []) : []))]
    suggestions.value = fuzzySuggestions(all, token).slice(0, 30)
    if (results.every(item => item.status === 'rejected')) showSuggestions.value = false
  }).finally(() => { suggestLoading.value = false })
}

function fuzzySuggestions(values, token) {
  const keyword = String(token || '').trim()
  if (!keyword) return []
  return values
    .map(value => ({ value: String(value), score: suggestionScore(String(value), keyword) }))
    .filter(item => item.score < Number.POSITIVE_INFINITY)
    .sort((a, b) => a.score - b.score || a.value.length - b.value.length || a.value.localeCompare(b.value))
    .map(item => item.value)
}

function suggestionScore(value, token) {
  const lowerValue = value.toLowerCase()
  const lowerToken = token.toLowerCase()
  const includesAt = lowerValue.indexOf(lowerToken)
  if (includesAt >= 0) return includesAt * 2 + Math.max(0, value.length - token.length) * 0.01

  const positions = fuzzyMatchPositions(value, token)
  if (positions.length === token.length) {
    const span = positions[positions.length - 1] - positions[0]
    const separators = positions.filter(pos => '_:-.'.includes(value[pos - 1] || '')).length
    return 100 + span - separators * 4 + Math.max(0, value.length - token.length) * 0.01
  }

  const tokenParts = splitSuggestToken(token)
  const matchedParts = tokenParts.filter(part => lowerValue.includes(part.toLowerCase()))
  if (matchedParts.length > 0) {
    const firstPartAt = lowerValue.indexOf(matchedParts[0].toLowerCase())
    return 200 + firstPartAt + (tokenParts.length - matchedParts.length) * 20
  }

  return Number.POSITIVE_INFINITY
}

function splitSuggestToken(token) {
  const normalized = String(token || '').replace(/[_:.-]+/g, ' ')
  const explicitParts = normalized.split(/\s+/).filter(part => part.length >= 2)
  if (explicitParts.length > 1) return explicitParts

  const commonParts = ['cpu', 'memory', 'mem', 'total', 'seconds', 'bytes', 'http', 'request', 'requests', 'node', 'container', 'process', 'rate']
  const lowerToken = token.toLowerCase()
  const parts = commonParts.filter(part => lowerToken.includes(part))
  return parts.length > 0 ? parts : [token]
}

function fuzzyMatchPositions(value, token) {
  const lowerValue = value.toLowerCase()
  const lowerToken = token.toLowerCase()
  const positions = []
  let searchFrom = 0
  for (const char of lowerToken) {
    const pos = lowerValue.indexOf(char, searchFrom)
    if (pos < 0) return []
    positions.push(pos)
    searchFrom = pos + 1
  }
  return positions
}

function highlightedSuggestionParts(value) {
  const token = suggestToken.value
  if (!token) return [{ text: value, matched: false }]

  const lowerValue = value.toLowerCase()
  const lowerToken = token.toLowerCase()
  const includesAt = lowerValue.indexOf(lowerToken)
  if (includesAt >= 0) {
    return splitByMatchedPositions(value, Array.from({ length: token.length }, (_, idx) => includesAt + idx))
  }

  const positions = fuzzyMatchPositions(value, token)
  if (positions.length === token.length) {
    return splitByMatchedPositions(value, positions)
  }

  const matchedPositions = new Set()
  for (const part of splitSuggestToken(token)) {
    const start = lowerValue.indexOf(part.toLowerCase())
    if (start >= 0) {
      for (let idx = start; idx < start + part.length; idx += 1) matchedPositions.add(idx)
    }
  }
  return splitByMatchedPositions(value, [...matchedPositions].sort((a, b) => a - b))
}

function splitByMatchedPositions(value, positions) {
  if (!positions.length) return [{ text: value, matched: false }]
  const matched = new Set(positions)
  const parts = []
  let buffer = ''
  let currentMatched = matched.has(0)
  for (let idx = 0; idx < value.length; idx += 1) {
    const isMatched = matched.has(idx)
    if (idx > 0 && isMatched !== currentMatched) {
      parts.push({ text: buffer, matched: currentMatched })
      buffer = ''
      currentMatched = isMatched
    }
    buffer += value[idx]
  }
  if (buffer) parts.push({ text: buffer, matched: currentMatched })
  return parts
}

function parseTokenAt(text, pos) {
  const before = text.slice(0, pos)
  let braceIdx = -1, braceDepth = 0
  for (let i = before.length - 1; i >= 0; i--) {
    if (before[i] === '}') braceDepth++
    else if (before[i] === '{') { if (braceDepth === 0) { braceIdx = i; break } braceDepth-- }
  }
  if (braceIdx >= 0) {
    const inside = before.slice(braceIdx + 1)
    const eqMatch = inside.match(/(\w+)\s*=\s*(~?\s*")([^"]*)$/)
    if (eqMatch) return { token: eqMatch[3] || '', start: braceIdx + 1 + eqMatch.index + eqMatch[1].length + eqMatch[2].length, end: pos, type: 'value', labelName: eqMatch[1] }
    const commaMatch = inside.match(/(?:,\s*|^\s*)(\w*)$/)
    if (commaMatch) return { token: commaMatch[1], start: braceIdx + 1 + commaMatch.index + (commaMatch[0].length - commaMatch[1].length), end: pos, type: 'label' }
  }
  const tokenMatch = before.match(/([a-zA-Z_:][a-zA-Z0-9_:]*)$/)
  if (tokenMatch) return { token: tokenMatch[1], start: pos - tokenMatch[1].length, end: pos, type: 'metric' }
  return { token: '', start: 0, end: 0, type: 'metric', labelName: '' }
}

function cursorPixelPos(textarea, charIdx) {
  const style = getComputedStyle(textarea)
  const mirror = document.createElement('div')
  Object.assign(mirror.style, { position:'absolute', visibility:'hidden', whiteSpace:'pre-wrap', wordWrap:'break-word', overflowWrap:'break-word', font:style.font, fontFamily:style.fontFamily, fontSize:style.fontSize, lineHeight:style.lineHeight, padding:style.padding, border:style.border, width:style.width, boxSizing:style.boxSizing })
  document.body.appendChild(mirror)
  const before = textarea.value.slice(0, charIdx)
  mirror.textContent = before
  const marker = document.createElement('span'); marker.textContent = '​'; mirror.appendChild(marker)
  const taRect = textarea.getBoundingClientRect()
  const lineH = parseFloat(style.lineHeight) || 20
  const lines = before.split('\n')
  document.body.removeChild(mirror)
  return { top: taRect.top - textarea.scrollTop + (lines.length - 1) * lineH + lineH, left: taRect.left - textarea.scrollLeft + (marker.offsetLeft || lines[lines.length-1].length * 8) }
}

function onKeydown(e) {
  if (!showSuggestions.value || suggestions.value.length === 0) {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') { e.preventDefault(); executeQuery() }
    return
  }
  switch (e.key) {
    case 'ArrowDown': e.preventDefault(); selectedIdx.value = (selectedIdx.value + 1) % suggestions.value.length; break
    case 'ArrowUp': e.preventDefault(); selectedIdx.value = (selectedIdx.value - 1 + suggestions.value.length) % suggestions.value.length; break
    case 'Enter': case 'Tab': e.preventDefault(); selectSuggestion(suggestions.value[selectedIdx.value]); break
    case 'Escape': showSuggestions.value = false; break
  }
}

function selectSuggestion(item) {
  const text = query.value; const start = replaceRange.value.start; const end = replaceRange.value.end
  query.value = text.slice(0, start) + item + text.slice(end)
  showSuggestions.value = false
  nextTick(() => { const el = inputRef.value; if (el) { const np = start + item.length; el.setSelectionRange(np, np); el.focus(); onInput() } })
}

function handleClickOutside(e) {
  const wrap = document.querySelector('.query-input-row')
  if (wrap && !wrap.contains(e.target)) showSuggestions.value = false
}

// ── 查询执行 ──

let execTimer = null

function onTimeChange() {
  clearTimeout(execTimer)
  execTimer = setTimeout(() => executeQuery(), 100)
}

function shiftEvalTime(direction) {
  const base = evalTime.value ? new Date(evalTime.value) : new Date()
  if (Number.isNaN(base.getTime())) return
  base.setMinutes(base.getMinutes() + direction)
  evalTime.value = formatDateTime(base)
  onTimeChange()
}

function onGraphParamChange() {
  clearTimeout(execTimer)
  execTimer = setTimeout(() => switchToGraph(true), 100)
}

function onGraphRangeChange() {
  const adjustedStep = ensureValidGraphStep()
  if (adjustedStep) ElMessage.info(`${graphRange.value} 查询范围已自动将 Step 调整为 ${adjustedStep}`)
  onGraphParamChange()
}

function switchToTable(forceExec = false) {
  activeTab.value = 'table'
  if (forceExec || resultMode.value !== 'table') executeInstantQuery()
}

function switchToGraph(forceExec = false) {
  activeTab.value = 'graph'
  if (forceExec || resultMode.value !== 'graph') executeRangeQuery()
}

async function executeQuery() {
  if (!query.value.trim() || selectedDS.value.length === 0) return
  if (activeTab.value === 'graph') {
    await executeRangeQuery()
  } else {
    await executeInstantQuery()
  }
  await loadQueryHistory()
}

async function executeInstantQuery() {
  if (!query.value.trim() || selectedDS.value.length === 0) return
  showSuggestions.value = false
  loading.value = true; error.value = ''; queryWarning.value = ''; result.value = null
  const startTime = performance.now()
  try {
    result.value = await querySelectedDatasources(dsId => promInstantQuery({ datasource_id: dsId, query: query.value, time: toPromTimestamp(evalTime.value) }))
    resultMode.value = 'table'
  } catch (e) { error.value = e?.response?.data?.message || e?.message || '请求失败' }
  finally {
    execTime.value = ((performance.now() - startTime) / 1000).toFixed(2) + 's'
    loading.value = false
    if (result.value?.status === 'error') { error.value = result.value.error || 'Prometheus 返回错误'; result.value = null }
  }
}

async function executeRangeQuery() {
  if (!query.value.trim() || selectedDS.value.length === 0) return
  loading.value = true; error.value = ''; queryWarning.value = ''; result.value = null
  const startTime = performance.now()
  try {
    ensureValidGraphStep()
    const end = graphEndTime.value || formatDateTime(new Date())
    const start = calcRangeStart(end, graphRange.value)
    result.value = await querySelectedDatasources(dsId => promRangeQuery({
      datasource_id: dsId,
      query: query.value,
      start: toPromTimestamp(start),
      end: toPromTimestamp(end),
      step: graphStep.value,
    }))
    resultMode.value = 'graph'
    selectedSeriesIndexes.value = []
  } catch (e) { error.value = e?.response?.data?.message || e?.message || '请求失败' }
  finally {
    execTime.value = ((performance.now() - startTime) / 1000).toFixed(2) + 's'
    loading.value = false
    if (result.value?.status === 'error') { error.value = result.value.error || 'Prometheus 返回错误'; result.value = null }
  }
}

function rangeSeconds(range) {
  const amount = Number.parseInt(range)
  const unit = range.slice(-1)
  return amount * ({ s: 1, m: 60, h: 3600, d: 86400, w: 604800 }[unit] || 3600)
}

function stepSeconds(step) {
  const amount = Number.parseInt(step)
  const unit = step.slice(-1)
  return amount * ({ s: 1, m: 60, h: 3600 }[unit] || 1)
}

function isGraphStepTooFine(step) {
  return Math.floor(rangeSeconds(graphRange.value) / stepSeconds(step)) + 1 > maxPointsPerSeries
}

function ensureValidGraphStep() {
  if (!isGraphStepTooFine(graphStep.value)) return ''
  const nextStep = stepOptions.find(item => !isGraphStepTooFine(item.value))
  if (!nextStep) return ''
  graphStep.value = nextStep.value
  return nextStep.label
}

async function querySelectedDatasources(queryFn) {
  const datasourceIDs = [...selectedDS.value]
  const names = new Map(datasources.value.map(item => [String(item.id), item.name]))
  const outcomes = await Promise.all(datasourceIDs.map(async datasourceID => {
    try {
      return { datasourceID, datasourceName: names.get(String(datasourceID)) || `数据源 ${datasourceID}`, response: await queryFn(datasourceID) }
    } catch (requestError) {
      return { datasourceID, datasourceName: names.get(String(datasourceID)) || `数据源 ${datasourceID}`, error: requestError }
    }
  }))
  const successful = outcomes.filter(item => item.response)
  const failed = outcomes.filter(item => item.error)
  if (successful.length === 0) {
    throw new Error(failed.map(item => `${item.datasourceName}：${item.error?.message || '查询失败'}`).join('；') || '查询失败')
  }
  if (failed.length > 0) queryWarning.value = `${failed.map(item => item.datasourceName).join('、')} 查询失败，已展示其他数据源的结果`

  return {
    datasource_count: datasourceIDs.length,
    data: {
      resultType: successful[0].response?.data?.resultType,
      result: successful.flatMap(item => (item.response?.data?.result || []).map(series => ({
        ...series,
        datasource_id: item.datasourceID,
        datasource_name: item.datasourceName,
      }))),
    },
  }
}

function calcRangeStart(endStr, rangeStr) {
  const end = new Date(endStr)
  const num = parseInt(rangeStr); const unit = rangeStr.slice(-1)
  const mult = { s: 1000, m: 60000, h: 3600000, d: 86400000, w: 604800000 }
  return formatDateTime(new Date(end.getTime() - num * (mult[unit] || 3600000)))
}

// ── 表格 ──

const tableRows = computed(() => {
  if (!result.value?.data?.result) return []
  return result.value.data.result.map(r => {
    const metric = { ...r.metric }; const name = metric.__name__ || ''; delete metric.__name__
    const labelParts = Object.entries(metric).map(([k,v]) => `${k}="${v}"`)
    const element = labelParts.length > 0 ? `${name}{${labelParts.join(',')}}` : name
    const val = r.value?.[1]; const values = r.values
    return { datasourceName: r.datasource_name, element: values ? `${element}  (${values.length} 个点)` : element, value: val, displayValue: formatValue(val) }
  })
})
const hasMultiDatasourceResult = computed(() => (result.value?.datasource_count || 0) > 1)

function formatValue(v) {
  if (v === undefined || v === null) return '-'
  const num = Number(v); return isNaN(num) ? String(v) : num
}

// ── 图表 ──

const selectedSeriesIndexes = ref([])
const chartColors = ['#5470c6', '#91cc75', '#fac858', '#ee6666', '#73c0de', '#3ba272', '#fc8452', '#9a60b4', '#ea7ccc']

const chartSeries = computed(() => {
  if (!result.value?.data?.result) return []
  return result.value.data.result.map((r, index) => {
    const metric = { ...r.metric }; const name = metric.__name__ || Object.values(metric).join(',') || 'value'
    const labelStr = Object.entries(metric).filter(([k]) => k !== '__name__').map(([k,v]) => `${k}="${v}"`).join(',')
    const metricName = labelStr ? `${name}{${labelStr}}` : name
    const data = result.value.data.resultType === 'vector'
      ? [[Number(r.value[0]) * 1000, Number(r.value[1])]]
      : (r.values || []).map(([ts, val]) => [Number(ts) * 1000, Number(val)])
    const color = chartColors[index % chartColors.length]
    return {
      name: `${hasMultiDatasourceResult.value ? `[${r.datasource_name}] ` : ''}${metricName}`,
      datasourceName: r.datasource_name,
      metricName,
      type: 'line',
      symbol: 'none',
      data,
      lineStyle: { color },
      itemStyle: { color },
    }
  })
})

const hasSeriesFilter = computed(() => selectedSeriesIndexes.value.length > 0)
const visibleChartSeries = computed(() => {
  if (!hasSeriesFilter.value) return chartSeries.value
  return chartSeries.value.filter((_, index) => selectedSeriesIndexes.value.includes(index))
})

function isMetricSelected(index) {
  return !hasSeriesFilter.value || selectedSeriesIndexes.value.includes(index)
}

function toggleMetric(index, event) {
  const selected = selectedSeriesIndexes.value

  if (!event.shiftKey) {
    selectedSeriesIndexes.value = selected.length === 1 && selected[0] === index ? [] : [index]
    return
  }

  if (selected.length === 0) {
    selectedSeriesIndexes.value = [index]
    return
  }

  if (selected.includes(index)) {
    if (selected.length > 1) selectedSeriesIndexes.value = selected.filter(item => item !== index)
    return
  }

  selectedSeriesIndexes.value = [...selected, index].sort((a, b) => a - b)
}

const chartOption = computed(() => ({
  color: chartColors,
  tooltip: { trigger: 'axis', formatter(p) { let h = `<small>${new Date(p[0].data[0]).toLocaleString()}</small><br/>`; p.forEach(i => h += `${i.marker} ${i.seriesName}: <b>${i.data[1]}</b><br/>`); return h }},
  grid: { left: 60, right: 20, top: 10, bottom: 32 },
  xAxis: { type: 'time' },
  yAxis: { type: 'value', axisLabel: { formatter: v => formatValue(v) } },
  dataZoom: [{ type: 'inside' }],
  series: visibleChartSeries.value,
}))
</script>

<style scoped>
.page-container { display: flex; flex-direction: column; gap: 10px; }
.query-context-bar {
  display: flex;
  min-height: 36px;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}
.datasource-picker {
  display: flex;
  align-items: center;
  min-height: 36px;
  flex: 0 1 260px;
}
.datasource-select {
  width: 100%;
  min-width: 0;
}
.datasource-select :deep(.el-select__wrapper) {
  min-height: 36px;
  padding: 3px 11px;
  background: #fff;
  border-radius: 4px;
  box-shadow: 0 0 0 1px #dcdfe6 inset;
  transition: box-shadow .16s ease;
}
.datasource-select :deep(.el-select__wrapper.is-focused) {
  box-shadow: 0 0 0 1px #409eff inset;
}
.datasource-select :deep(.el-select__tags-text) {
  max-width: 190px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.history-button {
  width: 36px;
  height: 36px;
  margin-left: auto;
  padding: 0;
  color: #526176;
  background: #fff;
  border-color: #dcdfe6;
  border-radius: 6px;
}
.history-button:hover,
.history-button:focus-visible {
  color: #409eff;
  border-color: #a8cdf7;
  background: #f4f9ff;
}
.history-icon {
  width: 16px;
  height: 16px;
  fill: currentColor;
}
.toolbar-label { font-size: 13px; font-weight: 500; color: #606266; white-space: nowrap; }

/* ── 查询输入行 ── */
.query-run-row { display: flex; align-items: flex-start; gap: 8px; }
.query-input-row {
  position: relative;
  flex: 1;
  min-width: 0;
}
.input-shell {
  display: flex;
  align-items: stretch;
  min-height: 40px;
  overflow: hidden;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  transition: border-color .16s ease, box-shadow .16s ease;
}
.input-shell:focus-within {
  border-color: #409eff;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, .12);
}
.input-leading {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  flex: 0 0 42px;
  color: #6b7280;
  background: #fafafa;
  border-right: 1px solid #ebeef5;
}
.promql-input {
  flex: 1;
  min-width: 0;
  border: none;
  outline: none;
  padding: 9px 12px;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 14px;
  line-height: 1.5;
  color: #303133;
  resize: none;
  overflow: hidden;
  box-sizing: border-box;
  background: #fff;
}
.promql-input::placeholder { color: #c0c4cc; font-family: inherit; }
.execute-button { flex: 0 0 auto; min-height: 40px; }
/* legacy class retained for old DOM-free hot reloads */
.input-right {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px;
  background: #fafbfc;
  border-left: 1px solid #e4e7ed;
  border-radius: 0 8px 8px 0;
}

/* ── 自动补全 ── */
.autocomplete-dropdown {
  position: fixed; z-index: 9999; width: 360px; max-height: 240px; overflow-y: auto;
  background: #fff; border: 1px solid #e4e7ed; border-radius: 8px;
  box-shadow: 0 6px 20px rgba(0,0,0,0.12);
}
.suggest-loading, .suggest-empty { padding: 12px 16px; font-size: 13px; color: #909399; }
.suggest-item { display: flex; align-items: center; gap: 8px; padding: 7px 14px; font-size: 13px; font-family: 'SF Mono', monospace; cursor: pointer; transition: background 0.1s; }
.suggest-item:hover, .suggest-item.active { background: #ecf5ff; }
.suggest-type-icon { width: 18px; height: 18px; border-radius: 4px; display: flex; align-items: center; justify-content: center; font-size: 10px; font-weight: 700; flex-shrink: 0; color: #fff; background: #409EFF; }
.suggest-text { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.suggest-text .matched {
  color: #1677ff;
  font-weight: 700;
}

/* ── 结果 ── */
.result-panel { background: #fff; }
.result-tabs { display: flex; align-items: center; border-bottom: 1px solid #dcdfe6; }
.tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 38px;
  padding: 0 14px;
  color: #606266;
  cursor: pointer;
  background: transparent;
  border: 0;
  border-bottom: 2px solid transparent;
  font-size: 14px;
  transition: color .16s ease, border-color .16s ease;
}
.tab:hover { color: #409EFF; }
.tab.active { color: #409EFF; border-bottom-color: #409EFF; font-weight: 500; }
.tab-meta { margin-left: auto; padding-right: 16px; font-size: 12px; color: #909399; display: flex; gap: 16px; }
.result-content { padding-top: 12px; }

.panel-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  padding: 0 16px;
}
.eval-control {
  display: inline-flex;
  align-items: center;
  overflow: hidden;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  background: #fff;
}
.eval-control :deep(.el-button) {
  width: 36px;
  height: 32px;
  border-radius: 0;
}
.eval-picker {
  width: 220px;
}
.eval-picker :deep(.el-input__wrapper) {
  border-radius: 0;
  box-shadow: none;
  border-left: 1px solid #ebeef5;
  border-right: 1px solid #ebeef5;
}

.error-box, .warning-box { display: flex; align-items: center; gap: 8px; padding: 12px 16px; border-radius: 6px; }
.error-box { color: #F56C6C; background: #fef0f0; }
.warning-box { margin-bottom: 8px; color: #b88230; background: #fdf6ec; }
.query-alert {
  margin-top: 2px;
  padding: 14px 16px;
  color: #409eff;
  font-size: 14px;
  background: #f7fbff;
  border-radius: 6px;
}

.query-result-table :deep(.el-table__cell .cell) { white-space: nowrap; }
.metric-cell, .value-cell, .result-source { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.metric-cell { font-family: 'SF Mono', monospace; font-size: 13px; color: #303133; }
.result-source { max-width: 100%; color: #606266; font-size: 13px; line-height: 20px; }
.value-cell { font-family: 'SF Mono', monospace; font-size: 13px; font-weight: 500; color: #303133; text-align: right; }

.empty-chart { display: flex; align-items: center; justify-content: center; height: 300px; color: #909399; font-size: 13px; }

/* ── Graph params ── */
.graph-params {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: center;
  margin-bottom: 12px;
  padding: 0 16px;
}
.range-item { display: flex; align-items: center; gap: 6px; }

/* ── Metric legend ── */
.metric-legend {
  padding: 4px 16px 0;
  border-top: 1px solid #ebeef5;
}
.metric-legend-header {
  display: grid;
  grid-template-columns: 22px 180px minmax(0, 1fr);
  gap: 10px;
  min-height: 32px;
  align-items: center;
  color: #909399;
  border-bottom: 1px solid #ebeef5;
  font-size: 12px;
  font-weight: 500;
}
.metric-legend-header::before { content: ''; }
.metric-legend-item {
  display: flex;
  width: 100%;
  min-height: 30px;
  align-items: center;
  gap: 10px;
  padding: 0;
  color: inherit;
  text-align: left;
  cursor: pointer;
  background: transparent;
  border: 0;
  border-bottom: 1px solid #f2f3f5;
  transition: background-color .16s ease, opacity .16s ease;
}
.metric-legend-item:hover { background: #f7faff; }
.metric-legend-item.active { background: #eef5ff; }
.metric-legend-item.muted { opacity: .38; }
.metric-legend-item:focus-visible {
  position: relative;
  z-index: 1;
  outline: 2px solid #409eff;
  outline-offset: -2px;
}
.metric-legend-item:last-child { border-bottom: 0; }
.metric-legend-line {
  position: relative;
  width: 22px;
  height: 2px;
  flex: 0 0 22px;
  background: var(--series-color);
}
.metric-legend-line::after {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--series-color);
  content: '';
  transform: translate(-50%, -50%);
}
.metric-legend-source {
  width: 180px;
  flex: 0 0 180px;
  overflow: hidden;
  color: #606266;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.metric-legend-name {
  flex: 1;
  min-width: 0;
  padding: 6px 0;
  color: #606266;
  font-family: 'SF Mono', monospace;
  font-size: 12px;
  line-height: 18px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:global(.query-history-drawer) {
  width: min(420px, 92vw) !important;
  border-radius: 0 !important;
}
.history-drawer-header {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
}
.history-drawer-header strong {
  color: #1f2937;
  font-size: 16px;
}
.history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.history-item {
  display: flex;
  width: 100%;
  flex-direction: column;
  gap: 9px;
  padding: 12px 14px;
  color: #303133;
  text-align: left;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  cursor: pointer;
  transition: border-color .16s ease, background .16s ease, box-shadow .16s ease;
}
.history-item:hover,
.history-item:focus-visible {
  outline: none;
  border-color: #9fc8f7;
  background: #f7fbff;
  box-shadow: 0 4px 12px rgba(64, 158, 255, .08);
}
.history-query {
  overflow: hidden;
  font: 13px/1.5 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.history-meta {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 8px;
  color: #909399;
  font-size: 11px;
}
.history-meta > span:not(.el-tag) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.history-meta time {
  margin-left: auto;
  white-space: nowrap;
}

@media (max-width: 720px) {
  .query-run-row {
    align-items: stretch;
    flex-direction: column;
  }
  .datasource-picker {
    min-width: 0;
    flex: 1;
  }
  .execute-button {
    width: 100%;
  }
  .eval-picker {
    width: 100%;
  }
}
</style>
