<template>
  <div class="home">
    <el-card class="parse-card">
      <template #header>
        <div class="card-header">
          <span>视频解析</span>
        </div>
      </template>
      <div class="input-area">
        <el-input
          v-model="urlInput"
          placeholder="输入 B站视频链接或 BV/AV 号，例如：BV1xx411c7mD"
          size="large"
          clearable
          @keyup.enter="handleParse"
        >
          <template #prepend>
            <el-icon><Link /></el-icon>
          </template>
        </el-input>
        <el-select v-model="selectedQN" size="large" style="width: 140px; margin-left: 12px;">
          <el-option v-for="q in qualityOptions" :key="q.value" :label="q.label" :value="q.value" />
        </el-select>
        <el-button type="primary" size="large" :loading="parsing" @click="handleParse">
          <el-icon><Search /></el-icon> 解析
        </el-button>
      </div>
      <el-divider />
      <div v-if="video" class="video-info">
        <div class="video-meta">
          <el-image :src="video.videoPreview" class="cover" fit="cover" />
          <div class="meta-text">
            <h3>{{ video.videoName }}</h3>
            <p class="author">UP主: {{ video.author }}</p>
            <p class="brief">{{ video.brief }}</p>
            <p class="count">共 {{ clipCount }} 个视频</p>
          </div>
        </div>

        <el-table :data="clipList" border style="margin-top: 20px;">
          <el-table-column type="index" label="#" width="60" />
          <el-table-column prop="title" label="标题" min-width="300" />
          <el-table-column label="清晰度" width="280">
            <template #default="{ row }">
              <el-select v-model="row.selectedQN" size="small" @change="onQNChange(row)">
                <el-option
                  v-for="(url, qn) in row.links"
                  :key="qn"
                  :label="qnLabel(qn)"
                  :value="qn"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button
                type="primary"
                size="small"
                :loading="row.downloading"
                @click="handleDownload(row)"
              >
                <el-icon><Download /></el-icon> 下载
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { parseVideo, startDownload } from '../api'

const urlInput = ref('')
const selectedQN = ref(80)
const parsing = ref(false)
const video = ref(null)

const qualityOptions = [
  { label: '4K (120)', value: 120 },
  { label: '1080P60 (116)', value: 116 },
  { label: '1080P+ (112)', value: 112 },
  { label: '1080P (80)', value: 80 },
  { label: '720P60 (74)', value: 74 },
  { label: '720P (64)', value: 64 },
  { label: '480P (32)', value: 32 },
  { label: '360P (16)', value: 16 },
]

const qnLabelMap = {
  127: '8K', 126: 'HDR', 120: '4K', 116: '1080P60',
  112: '1080P+', 80: '1080P', 74: '720P60', 64: '720P',
  32: '480P', 16: '360P',
}
const qnLabel = (qn) => qnLabelMap[qn] || `qn=${qn}`

const clipCount = computed(() => video.value ? Object.keys(video.value.clips || {}).length : 0)
const clipList = computed(() => {
  if (!video.value) return []
  return Object.values(video.value.clips).map(c => ({
    ...c,
    selectedQN: Object.keys(c.links || {})[0] ? Number(Object.keys(c.links)[0]) : 80,
    downloading: false
  }))
})

const handleParse = async () => {
  if (!urlInput.value.trim()) {
    ElMessage.warning('请输入视频链接')
    return
  }
  parsing.value = true
  try {
    const res = await parseVideo(urlInput.value.trim(), selectedQN.value, false)
    video.value = res.data
    ElMessage.success('解析成功')
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '解析失败')
  } finally {
    parsing.value = false
  }
}

const onQNChange = async (row) => {
  // Re-parse to get download link for selected quality
  try {
    const res = await parseVideo(urlInput.value.trim(), row.selectedQN, true)
    const updatedClips = res.data.clips
    const clip = updatedClips[row.cId]
    if (clip) {
      row.links = clip.links
    }
  } catch (e) {
    ElMessage.error('获取下载链接失败')
  }
}

const handleDownload = async (row) => {
  let url = row.links[row.selectedQN]
  if (!url) {
    // Fetch link first
    try {
      const res = await parseVideo(urlInput.value.trim(), row.selectedQN, true)
      const clip = res.data.clips[row.cId]
      if (clip) {
        url = clip.links[row.selectedQN]
        row.links[row.selectedQN] = url
      }
    } catch (e) {
      ElMessage.error('获取下载链接失败')
      return
    }
  }
  if (!url) {
    ElMessage.error('没有可用的下载链接')
    return
  }
  row.downloading = true
  try {
    await startDownload({
      avId: row.avId,
      cId: row.cId,
      title: row.title,
      url: url
    })
    ElMessage.success('已加入下载队列')
  } catch (e) {
    ElMessage.error('创建下载任务失败')
  } finally {
    row.downloading = false
  }
}
</script>

<style scoped>
.home { max-width: 1100px; margin: 0 auto; }
.parse-card { border-radius: 12px; }
.card-header { font-size: 16px; font-weight: 600; }
.input-area { display: flex; align-items: center; }
.input-area .el-input { flex: 1; }
.video-meta { display: flex; gap: 20px; }
.cover { width: 240px; height: 150px; border-radius: 8px; }
.meta-text { flex: 1; }
.meta-text h3 { font-size: 18px; margin-bottom: 8px; color: #333; }
.meta-text .author { color: #fb7299; margin-bottom: 8px; }
.meta-text .brief { color: #666; font-size: 13px; line-height: 1.6; margin-bottom: 8px; }
.meta-text .count { color: #999; font-size: 13px; }
</style>
