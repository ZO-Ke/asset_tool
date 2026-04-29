<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Folder, FolderOpened, Edit, Delete } from '@element-plus/icons-vue'
import {
  ListProjects,
  CreateProject,
  RenameProject,
  DeleteProject,
} from '../../wailsjs/go/main/App'
import type { model } from '../../wailsjs/go/models'

const router = useRouter()
const search = ref('')
const projects = ref<model.Project[]>([])
const loading = ref(false)

async function refresh() {
  loading.value = true
  try {
    projects.value = (await ListProjects(search.value)) || []
  } catch (e: any) {
    ElMessage.error('加载项目失败: ' + e)
  } finally {
    loading.value = false
  }
}

async function createProject() {
  try {
    const { value } = await ElMessageBox.prompt('项目名称', '新建项目', {
      confirmButtonText: '创建',
      cancelButtonText: '取消',
      inputValidator: (v) => (v && v.trim() ? true : '名称不能为空'),
    })
    await CreateProject(value.trim())
    ElMessage.success('已创建')
    refresh()
  } catch {
    /* user cancelled */
  }
}

async function rename(p: model.Project) {
  try {
    const { value } = await ElMessageBox.prompt('新名称', '重命名', {
      inputValue: p.name,
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputValidator: (v) => (v && v.trim() ? true : '名称不能为空'),
    })
    await RenameProject(p.id, value.trim())
    refresh()
  } catch { /* cancel */ }
}

async function remove(p: model.Project) {
  try {
    await ElMessageBox.confirm(
      `删除项目「${p.name}」及其所有 ${p.asset_count} 条资产？`,
      '确认删除',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await DeleteProject(p.id)
    ElMessage.success('已删除')
    refresh()
  } catch { /* cancel */ }
}

function openProject(p: model.Project) {
  router.push({ name: 'ProjectDetail', params: { id: p.id }, query: { name: p.name } })
}

onMounted(refresh)
</script>

<template>
  <div class="page">
    <div class="header">
      <h1>资产测绘工具 <span class="by">by CrispSheep</span></h1>
    </div>

    <div class="toolbar">
      <el-input
        v-model="search"
        placeholder="搜索项目名…"
        :prefix-icon="Search"
        clearable
        @input="refresh"
        style="flex: 1"
      />
      <el-button type="primary" :icon="Plus" @click="createProject">
        新建项目
      </el-button>
    </div>

    <div class="list" v-loading="loading">
      <el-empty v-if="projects.length === 0" description="还没有项目，点右上角新建" />
      <div
        v-for="p in projects"
        :key="p.id"
        class="card"
        @dblclick="openProject(p)"
      >
        <el-icon class="folder-icon"><FolderOpened /></el-icon>
        <div class="info">
          <div class="name">{{ p.name }}</div>
          <div class="badges">
            <el-tag v-if="p.asset_count === 0" type="info" size="small">空</el-tag>
            <template v-else>
              <el-tag size="small">IP {{ p.ip_count }}</el-tag>
              <el-tag size="small" type="warning">域名 {{ p.domain_count }}</el-tag>
              <el-tag size="small" type="success">存活 {{ p.alive_count }}</el-tag>
            </template>
          </div>
        </div>
        <div class="meta">
          <span class="time">{{ p.created_at }}</span>
          <div class="actions">
            <el-button size="small" :icon="Folder" @click.stop="openProject(p)">打开</el-button>
            <el-button size="small" :icon="Edit" @click.stop="rename(p)">改名</el-button>
            <el-button size="small" type="danger" :icon="Delete" @click.stop="remove(p)">删除</el-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page {
  height: 100vh;
  padding: 24px 28px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  box-sizing: border-box;
}
.header h1 {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: #ffffff;
}
.header .by {
  font-size: 13px;
  color: #1890ff;
  font-weight: 500;
  margin-left: 8px;
}
.toolbar {
  display: flex;
  gap: 12px;
}
.list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 16px;
  background: #25282f;
  border: 1px solid #3a3e4a;
  border-radius: 8px;
  cursor: pointer;
  transition: border-color 0.15s, transform 0.15s;
}
.card:hover {
  border-color: #1890ff;
  transform: translateX(2px);
}
.folder-icon {
  font-size: 26px;
  color: #1890ff;
}
.info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}
.name {
  font-size: 15px;
  font-weight: 600;
  color: #ffffff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.badges {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
}
.time {
  font-size: 12px;
  color: #6c7080;
}
.actions {
  display: flex;
  gap: 4px;
}
</style>
