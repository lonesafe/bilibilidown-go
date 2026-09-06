<template>
  <div class="downloads">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>下载任务</span>
          <el-button size="small" @click="refreshTasks">
            <el-icon><Refresh /></el-icon> 刷新
          </el-button>
        </div>
      </template>

      <el-table :data="tasks" border style="width: 100%">
        <el-table-column prop="title" label="视频标题" min-width="280" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="进度" width="260">
          <template #default="{ row }">
            <el-progress
              :percentage="progress(row)"
              :status="row.status === 'success' ? 'success' : row.status === 'fail' ? 'exception' : null"
              :stroke-width="14"
            />
          </template>
        </el-table-column>
        <el-table-column label="速度" width="100">
          <template #default="{ row }">
            {{ formatSpeed(row.speed) }}
          </template>
        </el-table-column>
        <el-table-column label="大小" width="140">
          <template #default="{ row }">
            {{ formatSize(row.downloaded) }} / {{ formatSize(row.totalSize) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'downloading' || row.status === 'merging'"
              type="danger"
              size="small"
              @click="handleStop(row)"
            >
              停止
            </el-button>
            <el-button
              v-else-if="row.status === 'success'"
              type="success"
              size="small"
              disabled
            >
              完成
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="tasks.length === 0" description="暂无下载任务" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTasks, stopTask } from '../api'

const tasks = ref([])
let ws = null
let pollTimer = null

const statusText = (s) => ({
  waiting: '等待中',
  downloading: '下载中',
  merging: '合并中',
  success: '已完成',
  fail: '失败',
  stopped: '已停止'
}[s] || s)

const statusType = (s) => ({
  waiting: 'info',
  downloading: 'primary',
  merging: 'warning',
  success: 'success',
  fail: 'danger',
  stopped: 'info'
}[s] || 'info')

const progress = (row) => {
  if (row.totalSize > 0) {
    return Math.min(100, Math.round((row.downloaded / row.totalSize) * 100))
  }
  return row.status === 'success' ? 100 : 0
}

const formatSize = (bytes) => {
  if (!bytes || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let val = bytes
  while (val >= 1024 && i < units.length - 1) {
    val /= 1024
    i++
  }
  return val.toFixed(1) + ' ' + units[i]
}

const formatSpeed = (bytes) => {
  if (!bytes || bytes <= 0) return '-'
  return formatSize(bytes) + '/s'
}

const refreshTasks = async () => {
  try {
    const res = await getTasks()
    tasks.value = res.data
  } catch (e) {
    ElMessage.error('获取任务列表失败')
  }
}

const connectWS = () => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/api/ws`
  try {
    ws = new WebSocket(wsUrl)
    ws.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data)
        if (msg.type === 'tasks') {
          tasks.value = msg.data
        }
      } catch (err) { /* ignore */ }
    }
    ws.onclose = () => {
      // Reconnect after 3s
      setTimeout(connectWS, 3000)
    }
  } catch (e) {
    // fallback to polling
    pollTimer = setInterval(refreshTasks, 2000)
  }
}

const handleStop = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要停止下载 "${row.title}" 吗？`, '提示', {
      type: 'warning'
    })
    await stopTask(row.id)
    ElMessage.success('已停止')
    refreshTasks()
  } catch (e) { /* cancelled */ }
}

onMounted(() => {
  refreshTasks()
  connectWS()
})

onUnmounted(() => {
  if (ws) ws.close()
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
.downloads { max-width: 1200px; margin: 0 auto; }
.card-header { display: flex; justify-content: space-between; align-items: center; font-size: 16px; font-weight: 600; }
</style>
