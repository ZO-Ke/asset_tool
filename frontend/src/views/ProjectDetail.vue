<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft, Upload, Search, Delete, DocumentCopy, Download,
} from '@element-plus/icons-vue'
import {
  ListAssets, DeleteAssets,
} from '../../wailsjs/go/main/App'
import type { model } from '../../wailsjs/go/models'
import ImportDialog from '../components/ImportDialog.vue'

const route = useRoute()
const router = useRouter()
const projectId = Number(route.params.id)
const projectName = (route.query.name as string) || `项目 #${projectId}`

const tab = ref<'ip' | 'domain'>('ip')
const ipAssets = ref<model.Asset[]>([])
const domainAssets = ref<model.Asset[]>([])
const loading = ref(false)

const search = ref('')
const filterMode = ref<'all' | 'alive' | 'non-http' | 'unprobed' | 'dead'>('all')

const selected = ref<model.Asset[]>([])
const importVisible = ref(false)

async function refresh() {
  loading.value = true
  try {
    ipAssets.value = (await ListAssets(projectId, 'ip', '')) || []
    domainAssets.value = (await ListAssets(projectId, 'domain', '')) || []
  } catch (e: any) {
    ElMessage.error('加载失败: ' + e)
  } finally {
    loading.value = false
  }
}

const currentList = computed(() => (tab.value === 'ip' ? ipAssets.value : domainAssets.value))

// 应用过滤 + 搜索（仅前端过滤，性能足够）
const httpPorts = new Set(['', '80', '443', '8080', '8443', '8000', '8888'])
const filteredList = computed(() => {
  const kw = search.value.trim().toLowerCase()
  return currentList.value.filter((a) => {
    // 下拉过滤
    switch (filterMode.value) {
      case 'alive':
        if (a.status !== 'alive') return false
        break
      case 'non-http':
        if (a.status !== 'dead') return false
        if (httpPorts.has(a.port || '')) return false
        break
      case 'unprobed':
        if (a.status) return false
        break
      case 'dead':
        if (a.status !== 'dead') return false
        break
    }
    // 关键字
    if (kw) {
      const blob = [
        a.host, a.port, a.status, a.status_code, a.title, a.server,
        (a.sources || []).join(','),
      ].map((x) => String(x || '').toLowerCase()).join(' ')
      if (!blob.includes(kw)) return false
    }
    return true
  })
})

const tabLabel = (kind: 'ip' | 'domain') => {
  const total = kind === 'ip' ? ipAssets.value.length : domainAssets.value.length
  if (filterMode.value === 'all' && !search.value.trim()) {
    return `${kind === 'ip' ? 'IP' : '域名'}列表 (${total})`
  }
  // 过滤后数量需要切换 tab 之前算
  return `${kind === 'ip' ? 'IP' : '域名'}列表 (${total})`
}

function statusType(s: string) {
  if (s === 'alive') return 'success'
  if (s === 'dead') return 'danger'
  return 'info'
}

function statusText(s: string) {
  if (s === 'alive') return 'alive'
  if (s === 'dead') return 'dead'
  return '未探活'
}

function fullURL(a: model.Asset) {
  if (a.host.startsWith('http://') || a.host.startsWith('https://')) return a.host
  const target = a.port ? `${a.host}:${a.port}` : a.host
  const scheme = ['443', '8443'].includes(a.port || '') ? 'https' : 'http'
  return `${scheme}://${target}`
}

function copyHost(a: model.Asset) {
  const text = a.port ? `${a.host}:${a.port}` : a.host
  navigator.clipboard.writeText(text)
  ElMessage.success(`已复制: ${text}`)
}

function openInBrowser(a: model.Asset) {
  // Wails 不直接支持 openExternal，用 window.open 或绑定 BrowserOpenURL
  // @ts-ignore
  if (window.runtime && window.runtime.BrowserOpenURL) {
    // @ts-ignore
    window.runtime.BrowserOpenURL(fullURL(a))
  } else {
    window.open(fullURL(a), '_blank')
  }
}

function copyVisible() {
  if (filteredList.value.length === 0) {
    ElMessage.warning('没有数据可复制')
    return
  }
  const lines = filteredList.value.map((a) =>
    a.port ? `${a.host}:${a.port}` : a.host
  )
  navigator.clipboard.writeText(lines.join('\n'))
  ElMessage.success(`已复制 ${lines.length} 条到剪贴板`)
}

async function deleteVisible() {
  if (filteredList.value.length === 0) return
  try {
    await ElMessageBox.confirm(
      `将删除当前可见的 ${filteredList.value.length} 条资产，建议先点📋复制备份`,
      '确认批量删除',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await DeleteAssets(filteredList.value.map((a) => a.id))
    ElMessage.success('已删除')
    refresh()
  } catch { /* cancel */ }
}

async function deleteSelected() {
  if (selected.value.length === 0) return
  try {
    await ElMessageBox.confirm(
      `删除选中的 ${selected.value.length} 条资产？`,
      '确认删除',
      { type: 'warning' }
    )
    await DeleteAssets(selected.value.map((a) => a.id))
    selected.value = []
    refresh()
  } catch { /* cancel */ }
}

function exportTXT() {
  if (filteredList.value.length === 0) {
    ElMessage.warning('没有数据可导出')
    return
  }
  const lines = filteredList.value.map((a) =>
    a.port ? `${a.host}:${a.port}` : a.host
  )
  download(`${projectName}_${tab.value}.txt`, lines.join('\n'))
}

function exportCSV() {
  if (filteredList.value.length === 0) {
    ElMessage.warning('没有数据可导出')
    return
  }
  const header = ['host', 'port', 'type', 'sources', 'status', 'status_code', 'title', 'server']
  const rows = filteredList.value.map((a) => [
    a.host, a.port, a.type, (a.sources || []).join(', '),
    a.status || '', a.status_code ?? '', (a.title || '').replace(/,/g, ' '), a.server || '',
  ])
  const csv = '﻿' + [header, ...rows].map((r) => r.join(',')).join('\n')
  download(`${projectName}_${tab.value}.csv`, csv)
}

function download(filename: string, content: string) {
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = filename
  a.click()
  URL.revokeObjectURL(a.href)
  ElMessage.success(`已导出 ${filename}`)
}

onMounted(refresh)
</script>

<template>
  <div class="page">
    <!-- 顶部导航 -->
    <div class="header">
      <el-button :icon="ArrowLeft" @click="router.back()">返回</el-button>
      <h2>{{ projectName }}</h2>
      <span class="muted">#{{ projectId }}</span>
      <div class="spacer" />
    </div>

    <!-- 操作栏 -->
    <div class="toolbar">
      <el-button :icon="Upload" @click="importVisible = true">导入资产</el-button>
      <div class="spacer" />
      <el-select v-model="filterMode" style="width: 200px">
        <el-option label="全部" value="all" />
        <el-option label="仅 HTTP 存活" value="alive" />
        <el-option label="仅非 HTTP 服务" value="non-http" />
        <el-option label="仅未探活" value="unprobed" />
        <el-option label="仅 dead" value="dead" />
      </el-select>
      <el-button :icon="DocumentCopy" @click="copyVisible">复制可见</el-button>
      <el-button :icon="Download" @click="exportTXT">导出 TXT</el-button>
      <el-button :icon="Download" @click="exportCSV">导出 CSV</el-button>
      <el-button :icon="Delete" type="danger" plain @click="deleteVisible">删除可见</el-button>
    </div>

    <!-- 搜索 -->
    <el-input
      v-model="search"
      :prefix-icon="Search"
      placeholder="搜索 host / 状态码 / title / server …"
      clearable
    />

    <!-- 选中后的批量操作条 -->
    <div v-if="selected.length > 0" class="batch-bar">
      已选中 {{ selected.length }} 条
      <el-button size="small" type="danger" :icon="Delete" @click="deleteSelected">删除选中</el-button>
      <el-button size="small" link @click="selected = []">取消选择</el-button>
    </div>

    <!-- Tabs + 表格 -->
    <el-tabs v-model="tab" class="tabs-area">
      <el-tab-pane :label="tabLabel('ip')" name="ip" />
      <el-tab-pane :label="tabLabel('domain')" name="domain" />
    </el-tabs>

    <el-table
      :data="filteredList"
      v-loading="loading"
      height="100%"
      class="asset-table"
      @selection-change="(rows: model.Asset[]) => (selected = rows)"
    >
      <el-table-column type="selection" width="44" />
      <el-table-column prop="host" label="host" min-width="240">
        <template #default="{ row }">
          <a class="host-link" :title="'单击复制 / 双击在浏览器打开'"
             @click="copyHost(row)" @dblclick="openInBrowser(row)">
            {{ row.host }}
          </a>
        </template>
      </el-table-column>
      <el-table-column prop="port" label="port" width="80" />
      <el-table-column label="来源" width="140">
        <template #default="{ row }">
          <el-tag v-for="s in row.sources" :key="s" size="small" style="margin-right:4px">
            {{ s }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90" sortable>
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">
            {{ statusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status_code" label="状态码" width="80" sortable />
      <el-table-column prop="title" label="title" min-width="180" show-overflow-tooltip />
      <el-table-column prop="server" label="server" width="140" show-overflow-tooltip />
      <el-table-column prop="probed_at" label="探活时间" width="160" />
    </el-table>

    <ImportDialog
      v-model:visible="importVisible"
      :project-id="projectId"
      @imported="refresh"
    />
  </div>
</template>

<style scoped>
.page {
  height: 100vh;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  box-sizing: border-box;
}
.header {
  display: flex;
  align-items: center;
  gap: 12px;
}
.header h2 {
  margin: 0;
  color: #fff;
  font-size: 18px;
}
.muted {
  color: #6c7080;
  font-size: 13px;
}
.spacer {
  flex: 1;
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.batch-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 6px 12px;
  background: #25282f;
  border-radius: 6px;
  color: #1890ff;
}
.tabs-area {
  margin-bottom: -8px;
}
:deep(.el-tabs__item) {
  color: #b8b8b8;
  font-size: 14px;
}
:deep(.el-tabs__item:hover) {
  color: #ffffff;
}
:deep(.el-tabs__item.is-active) {
  color: #1890ff;
  font-weight: 600;
}
:deep(.el-tabs__active-bar) {
  background-color: #1890ff;
}
:deep(.el-tabs__nav-wrap::after) {
  background-color: #3a3e4a;
}
.asset-table {
  flex: 1;
  background: transparent !important;
}
.host-link {
  color: #40a9ff;
  cursor: pointer;
  text-decoration: none;
}
.host-link:hover {
  text-decoration: underline;
}
/* ── 表格深色主题（覆盖 Element Plus 默认） ───────────────── */
:deep(.el-table),
:deep(.el-table__inner-wrapper),
:deep(.el-table__body-wrapper),
:deep(.el-table__header-wrapper),
:deep(.el-table tr),
:deep(.el-table th.el-table__cell),
:deep(.el-table td.el-table__cell) {
  background-color: #25282f !important;
  color: #e8e8e8;
}
:deep(.el-table) {
  --el-table-bg-color: #25282f;
  --el-table-tr-bg-color: #25282f;
  --el-table-header-bg-color: #2a2d36;
  --el-table-row-hover-bg-color: #2f323a;
  --el-table-border-color: #3a3e4a;
  --el-table-header-text-color: #b8b8b8;
  --el-table-text-color: #e8e8e8;
  --el-table-fixed-box-shadow: none;
  border: 1px solid #3a3e4a;
  border-radius: 6px;
}
:deep(.el-table th.el-table__cell) {
  background-color: #2a2d36 !important;
  color: #b8b8b8 !important;
  font-weight: 600;
  border-bottom: 1px solid #3a3e4a;
}
:deep(.el-table tr:hover > td.el-table__cell) {
  background-color: #2f323a !important;
}
:deep(.el-table--enable-row-hover .el-table__body tr:hover > td.el-table__cell) {
  background-color: #2f323a !important;
}
:deep(.el-table__empty-block),
:deep(.el-table__empty-text) {
  background-color: #25282f !important;
  color: #6c7080;
}
</style>
